/*
*
* Reply Resources File
* Copyright Azareal 2016 - 2018
*
 */
package common

import (
	"errors"
	"time"
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

// TODO: Write tests for this
// TODO: Wrap these queries in a transaction to make sure the state is consistent
func (reply *Reply) Like(uid int) (err error) {
	var rid int // unused, just here to avoid mutating reply.ID
	err = stmts.hasLikedReply.QueryRow(uid, reply.ID).Scan(&rid)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	score := 1
	_, err = stmts.createLike.Exec(score, reply.ID, "replies", uid)
	if err != nil {
		return err
	}
	_, err = stmts.addLikesToReply.Exec(1, reply.ID)
	return err
}

// TODO: Write tests for this
func (reply *Reply) Delete() error {
	_, err := stmts.deleteReply.Exec(reply.ID)
	if err != nil {
		return err
	}
	_, err = stmts.removeRepliesFromTopic.Exec(1, reply.ParentID)
	tcache, ok := topics.(TopicCache)
	if ok {
		tcache.CacheRemove(reply.ParentID)
	}
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the reply
func (reply *Reply) Copy() Reply {
	return *reply
}
