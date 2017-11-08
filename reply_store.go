package main

import "database/sql"
import "./query_gen/lib"

var rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error)
}

type SQLReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLReplyStore() (*SQLReplyStore, error) {
	acc := qgen.Builder.Accumulator()
	return &SQLReplyStore{
		get:    acc.SimpleSelect("replies", "tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount", "rid = ?", "", ""),
		create: acc.SimpleInsert("replies", "tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy", "?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?"),
	}, acc.FirstError()
}

func (store *SQLReplyStore) Get(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress, &reply.LikeCount)
	return &reply, err
}

// TODO: Write a test for this
func (store *SQLReplyStore) Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error) {
	wcount := wordCount(content)
	res, err := store.create.Exec(topic.ID, content, parseMessage(content, topic.ParentID, "forums"), ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), topic.AddReply(uid)
}
