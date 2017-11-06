package main

import "database/sql"
import "./query_gen/lib"

var rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Create(tid int, content string, ipaddress string, fid int, uid int) (id int, err error)
}

type SQLReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLReplyStore() (*SQLReplyStore, error) {
	getReplyStmt, err := qgen.Builder.SimpleSelect("replies", "tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount", "rid = ?", "", "")
	if err != nil {
		return nil, err
	}
	createReplyStmt, err := qgen.Builder.SimpleInsert("replies", "tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy", "?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?")
	if err != nil {
		return nil, err
	}
	return &SQLReplyStore{
		get:    getReplyStmt,
		create: createReplyStmt,
	}, nil
}

func (store *SQLReplyStore) Get(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress, &reply.LikeCount)
	return &reply, err
}

// TODO: Write a test for this
func (store *SQLReplyStore) Create(tid int, content string, ipaddress string, fid int, uid int) (id int, err error) {
	wcount := wordCount(content)
	res, err := store.create.Exec(tid, content, parseMessage(content, fid, "forums"), ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = stmts.addRepliesToTopic.Exec(1, uid, tid)
	if err != nil {
		return int(lastID), err
	}
	tcache, ok := topics.(TopicCache)
	if ok {
		tcache.CacheRemove(tid)
	}
	return int(lastID), err
}
