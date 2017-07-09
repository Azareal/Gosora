/* Copyright Azareal 2016 - 2017 */
package main
import "html/template"

// Should we add a reply store to centralise all the reply logic? Would this cover profile replies too or would that be seperate?

type Reply struct /* Should probably rename this to ReplyUser and rename ReplyShort to Reply */
{
	ID int
	ParentID int
	Content string
	ContentHtml string
	CreatedBy int
	UserSlug string
	CreatedByName string
	Group int
	CreatedAt string
	LastEdit int
	LastEditBy int
	Avatar string
	Css template.CSS
	ContentLines int
	Tag string
	URL string
	URLPrefix string
	URLName string
	Level int
	IpAddress string
	Liked bool
	LikeCount int
	ActionType string
	ActionIcon string
}

type ReplyShort struct
{
	ID int
	ParentID int
	Content string
	CreatedBy int
	Group int
	CreatedAt string
	LastEdit int
	LastEditBy int
	ContentLines int
	IpAddress string
	Liked bool
	LikeCount int
}

func get_reply(id int) (*ReplyShort, error) {
	reply := ReplyShort{ID:id}
	err := get_reply_stmt.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IpAddress, &reply.LikeCount)
	return &reply, err
}

func get_user_reply(id int) (*ReplyShort, error) {
	reply := ReplyShort{ID:id}
	err := get_user_reply_stmt.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IpAddress)
	return &reply, err
}
