/*
*
* Reply Resources File
* Copyright Azareal 2016 - 2018
*
 */
package main

// ? - Should we add a reply store to centralise all the reply logic? Would this cover profile replies too or would that be seperate?

type ReplyUser struct {
	ID            int
	ParentID      int
	Content       string
	ContentHtml   string
	CreatedBy     int
	UserLink      string
	CreatedByName string
	Group         int
	CreatedAt     string
	LastEdit      int
	LastEditBy    int
	Avatar        string
	ClassName     string
	ContentLines  int
	Tag           string
	URL           string
	URLPrefix     string
	URLName       string
	Level         int
	IPAddress     string
	Liked         bool
	LikeCount     int
	ActionType    string
	ActionIcon    string
}

type Reply struct {
	ID           int
	ParentID     int
	Content      string
	CreatedBy    int
	Group        int
	CreatedAt    string
	LastEdit     int
	LastEditBy   int
	ContentLines int
	IPAddress    string
	Liked        bool
	LikeCount    int
}

func (reply *Reply) Copy() Reply {
	return *reply
}

func getReply(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := getReplyStmt.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress, &reply.LikeCount)
	return &reply, err
}

func getUserReply(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := getUserReplyStmt.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress)
	return &reply, err
}
