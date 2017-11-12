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
	delete                 *sql.Stmt
	addLikesToReply        *sql.Stmt
	removeRepliesFromTopic *sql.Stmt
	getParent              *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		replyStmts = ReplyStmts{
			isLiked:                acc.Select("likes").Columns("targetItem").Where("sentBy = ? and targetItem = ? and targetType = 'replies'").Prepare(),
			createLike:             acc.Insert("likes").Columns("weight, targetItem, targetType, sentBy").Fields("?,?,?,?").Prepare(),
			delete:                 acc.Delete("replies").Where("rid = ?").Prepare(),
			addLikesToReply:        acc.Update("replies").Set("likeCount = likeCount + ?").Where("rid = ?").Prepare(),
			removeRepliesFromTopic: acc.Update("topics").Set("postCount = postCount - ?").Where("tid = ?").Prepare(),
			getParent:              acc.SimpleLeftJoin("replies", "topics", "topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, topics.data", "replies.tid = topics.tid", "rid = ?", "", ""),
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
	tcache, ok := Topics.(TopicCache)
	if ok {
		tcache.CacheRemove(reply.ParentID)
	}
	return err
}

func (reply *Reply) Topic() (*Topic, error) {
	topic := Topic{ID: 0}
	err := replyStmts.getParent.QueryRow(reply.ID).Scan(&topic.ID, &topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = BuildTopicURL(NameToSlug(topic.Title), topic.ID)
	return &topic, err
}

// Copy gives you a non-pointer concurrency safe copy of the reply
func (reply *Reply) Copy() Reply {
	return *reply
}

func BlankReply() *Reply {
	return &Reply{ID: 0}
}
