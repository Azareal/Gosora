/*
*
* Reply Resources File
* Copyright Azareal 2016 - 2020
*
 */
package common

import (
	"database/sql"
	"errors"
	"html"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type ReplyUser struct {
	Reply
	//ID            int
	//ParentID      int
	//Content       string
	ContentHtml string
	//CreatedBy     int
	UserLink      string
	CreatedByName string
	//Group         int
	//CreatedAt     time.Time
	//LastEdit      int
	//LastEditBy    int
	Avatar      string
	MicroAvatar string
	ClassName   string
	//ContentLines  int
	Tag       string
	URL       string
	URLPrefix string
	URLName   string
	Level     int
	//IP     string
	//Liked         bool
	//LikeCount     int
	//AttachCount int
	//ActionType  string
	ActionIcon string

	Attachments []*MiniAttachment
	Deletable   bool
}

type Reply struct {
	ID           int
	ParentID     int
	Content      string
	CreatedBy    int
	Group        int
	CreatedAt    time.Time
	LastEdit     int
	LastEditBy   int
	ContentLines int
	IP           string
	Liked        bool
	LikeCount    int
	AttachCount  int
	ActionType   string
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
			setPoll:                acc.Update("replies").Set("poll = ?").Where("rid = ? AND poll = 0").Prepare(),
			delete:                 acc.Delete("replies").Where("rid = ?").Prepare(),
			addLikesToReply:        acc.Update("replies").Set("likeCount = likeCount + ?").Where("rid = ?").Prepare(),
			removeRepliesFromTopic: acc.Update("topics").Set("postCount = postCount - ?").Where("tid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Write tests for this
// TODO: Wrap these queries in a transaction to make sure the state is consistent
func (r *Reply) Like(uid int) (err error) {
	var rid int // unused, just here to avoid mutating reply.ID
	err = replyStmts.isLiked.QueryRow(uid, r.ID).Scan(&rid)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	score := 1
	_, err = replyStmts.createLike.Exec(score, r.ID, "replies", uid)
	if err != nil {
		return err
	}
	_, err = replyStmts.addLikesToReply.Exec(1, r.ID)
	if err != nil {
		return err
	}
	_, err = userStmts.incrementLiked.Exec(1, uid)
	_ = Rstore.GetCache().Remove(r.ID)
	return err
}

func (r *Reply) Delete() error {
	_, err := replyStmts.delete.Exec(r.ID)
	if err != nil {
		return err
	}
	// TODO: Move this bit to *Topic
	_, err = replyStmts.removeRepliesFromTopic.Exec(1, r.ParentID)
	tcache := Topics.GetCache()
	if tcache != nil {
		tcache.Remove(r.ParentID)
	}
	_ = Rstore.GetCache().Remove(r.ID)
	return err
}

func (r *Reply) SetPost(content string) error {
	topic, err := r.Topic()
	if err != nil {
		return err
	}
	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, topic.ParentID, "forums")
	_, err = replyStmts.edit.Exec(content, parsedContent, r.ID) // TODO: Sniff if this changed anything to see if we hit an existing poll
	_ = Rstore.GetCache().Remove(r.ID)
	return err
}

// TODO: Write tests for this
func (r *Reply) SetPoll(pollID int) error {
	_, err := replyStmts.setPoll.Exec(pollID, r.ID) // TODO: Sniff if this changed anything to see if we hit a poll
	_ = Rstore.GetCache().Remove(r.ID)
	return err
}

func (r *Reply) Topic() (*Topic, error) {
	return Topics.Get(r.ParentID)
}

func (r *Reply) GetID() int {
	return r.ID
}

func (r *Reply) GetTable() string {
	return "replies"
}

// Copy gives you a non-pointer concurrency safe copy of the reply
func (r *Reply) Copy() Reply {
	return *r
}
