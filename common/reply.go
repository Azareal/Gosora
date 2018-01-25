/*
*
* Reply Resources File
* Copyright Azareal 2016 - 2018
*
 */
package common

import (
	"database/sql"
	"errors"
	"html"
	"time"

	"../query_gen/lib"
)

type ReplyUser struct {
	ID                int
	ParentID          int
	Content           string
	ContentHtml       string
	CreatedBy         int
	UserLink          string
	CreatedByName     string
	Group             int
	CreatedAt         time.Time
	RelativeCreatedAt string
	LastEdit          int
	LastEditBy        int
	Avatar            string
	ClassName         string
	ContentLines      int
	Tag               string
	URL               string
	URLPrefix         string
	URLName           string
	Level             int
	IPAddress         string
	Liked             bool
	LikeCount         int
	ActionType        string
	ActionIcon        string
}

type Reply struct {
	ID                int
	ParentID          int
	Content           string
	CreatedBy         int
	Group             int
	CreatedAt         time.Time
	RelativeCreatedAt string
	LastEdit          int
	LastEditBy        int
	ContentLines      int
	IPAddress         string
	Liked             bool
	LikeCount         int
}

var ErrAlreadyLiked = errors.New("You already liked this!")
var replyStmts ReplyStmts

type ReplyStmts struct {
	isLiked                *sql.Stmt
	createLike             *sql.Stmt
	edit                   *sql.Stmt
	setPoll                *sql.Stmt
	delete                 *sql.Stmt
	addLikesToReply        *sql.Stmt
	removeRepliesFromTopic *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		replyStmts = ReplyStmts{
			isLiked:                acc.Select("likes").Columns("targetItem").Where("sentBy = ? and targetItem = ? and targetType = 'replies'").Prepare(),
			createLike:             acc.Insert("likes").Columns("weight, targetItem, targetType, sentBy").Fields("?,?,?,?").Prepare(),
			edit:                   acc.Update("replies").Set("content = ?, parsed_content = ?").Where("rid = ? AND poll = 0").Prepare(),
			setPoll:                acc.Update("replies").Set("content = '', parsed_content = '', poll = ?").Where("rid = ? AND poll = 0").Prepare(),
			delete:                 acc.Delete("replies").Where("rid = ?").Prepare(),
			addLikesToReply:        acc.Update("replies").Set("likeCount = likeCount + ?").Where("rid = ?").Prepare(),
			removeRepliesFromTopic: acc.Update("topics").Set("postCount = postCount - ?").Where("tid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Write tests for this
// TODO: Wrap these queries in a transaction to make sure the state is consistent
func (reply *Reply) Like(uid int) (err error) {
	var rid int // unused, just here to avoid mutating reply.ID
	err = replyStmts.isLiked.QueryRow(uid, reply.ID).Scan(&rid)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	score := 1
	_, err = replyStmts.createLike.Exec(score, reply.ID, "replies", uid)
	if err != nil {
		return err
	}
	_, err = replyStmts.addLikesToReply.Exec(1, reply.ID)
	return err
}

// TODO: Write tests for this
func (reply *Reply) Delete() error {
	_, err := replyStmts.delete.Exec(reply.ID)
	if err != nil {
		return err
	}
	// TODO: Move this bit to *Topic
	_, err = replyStmts.removeRepliesFromTopic.Exec(1, reply.ParentID)
	tcache := Topics.GetCache()
	if tcache != nil {
		tcache.Remove(reply.ParentID)
	}
	return err
}

func (reply *Reply) SetPost(content string) error {
	topic, err := reply.Topic()
	if err != nil {
		return err
	}
	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, topic.ParentID, "forums")
	_, err = replyStmts.edit.Exec(content, parsedContent, reply.ID) // TODO: Sniff if this changed anything to see if we hit an existing poll
	return err
}

func (reply *Reply) SetPoll(pollID int) error {
	_, err := replyStmts.setPoll.Exec(pollID, reply.ID) // TODO: Sniff if this changed anything to see if we hit a poll
	return err
}

func (reply *Reply) Topic() (*Topic, error) {
	return Topics.Get(reply.ParentID)
}

// Copy gives you a non-pointer concurrency safe copy of the reply
func (reply *Reply) Copy() Reply {
	return *reply
}
