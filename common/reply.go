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
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type ReplyUser struct {
	Reply

	ContentHtml   string
	UserLink      string
	CreatedByName string
	Avatar        string
	MicroAvatar   string
	ClassName     string
	Tag           string
	URL           string
	//URLPrefix string
	//URLName   string
	Level      int
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
	deleteLikesForReply    *sql.Stmt
	deleteActivity         *sql.Stmt
	deleteActivitySubs     *sql.Stmt

	updateTopicReplies  *sql.Stmt
	updateTopicReplies2 *sql.Stmt

	getAidsOfReply *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		re := "replies"
		replyStmts = ReplyStmts{
			isLiked:                acc.Select("likes").Columns("targetItem").Where("sentBy=? and targetItem=? and targetType='replies'").Prepare(),
			createLike:             acc.Insert("likes").Columns("weight, targetItem, targetType, sentBy").Fields("?,?,?,?").Prepare(),
			edit:                   acc.Update(re).Set("content=?,parsed_content=?").Where("rid=? AND poll=0").Prepare(),
			setPoll:                acc.Update(re).Set("poll=?").Where("rid=? AND poll=0").Prepare(),
			delete:                 acc.Delete(re).Where("rid=?").Prepare(),
			addLikesToReply:        acc.Update(re).Set("likeCount=likeCount+?").Where("rid=?").Prepare(),
			removeRepliesFromTopic: acc.Update("topics").Set("postCount=postCount-?").Where("tid=?").Prepare(),
			deleteLikesForReply:    acc.Delete("likes").Where("targetItem=? AND targetType='replies'").Prepare(),
			deleteActivity:         acc.Delete("activity_stream").Where("elementID=? AND elementType='post'").Prepare(),
			deleteActivitySubs:     acc.Delete("activity_subscriptions").Where("targetID=? AND targetType='post'").Prepare(),

			// TODO: Optimise this to avoid firing an update if it's not the last reply in a topic. We will need to set lastReplyID properly in other places and in the patcher first so we can use it here.
			updateTopicReplies:  acc.RawPrepare("UPDATE topics t INNER JOIN replies r ON t.tid = r.tid SET t.lastReplyBy = r.createdBy, t.lastReplyAt = r.createdAt, t.lastReplyID = r.rid WHERE t.tid = ?"),
			updateTopicReplies2: acc.Update("topics").Set("lastReplyAt=createdAt,lastReplyBy=createdBy,lastReplyID=0").Where("postCount=1 AND tid=?").Prepare(),

			getAidsOfReply:  acc.Select("attachments").Columns("attachID").Where("originID=? AND originTable='replies'").Prepare(),
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
	_, err = userStmts.incLiked.Exec(1, uid)
	_ = Rstore.GetCache().Remove(r.ID)
	return err
}

// TODO: Refresh topic list?
func (r *Reply) Delete() error {
	creator, err := Users.Get(r.CreatedBy)
	if err == nil {
		err = creator.DecreasePostStats(WordCount(r.Content), false)
		if err != nil {
			return err
		}
	} else if err != ErrNoRows {
		return err
	}

	_, err = replyStmts.delete.Exec(r.ID)
	if err != nil {
		return err
	}
	// TODO: Move this bit to *Topic
	_, err = replyStmts.removeRepliesFromTopic.Exec(1, r.ParentID)
	if err != nil {
		return err
	}
	_, err = replyStmts.updateTopicReplies.Exec(r.ParentID)
	if err != nil {
		return err
	}
	_, err = replyStmts.updateTopicReplies2.Exec(r.ParentID)
	tc := Topics.GetCache()
	if tc != nil {
		tc.Remove(r.ParentID)
	}
	_ = Rstore.GetCache().Remove(r.ID)
	if err != nil {
		return err
	}
	_, err = replyStmts.deleteLikesForReply.Exec(r.ID)
	if err != nil {
		return err
	}
	err = handleReplyAttachments(r.ID)
	if err != nil {
		return err
	}
	err = Activity.DeleteByParamsExtra("reply",r.ParentID,"topic",strconv.Itoa(r.ID))
	if err != nil {
		return err
	}
	_, err = replyStmts.deleteActivitySubs.Exec(r.ID)
	if err != nil {
		return err
	}
	_, err = replyStmts.deleteActivity.Exec(r.ID)
	return err
}

func (r *Reply) SetPost(content string) error {
	topic, err := r.Topic()
	if err != nil {
		return err
	}
	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, topic.ParentID, "forums", nil)
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
