package common

import "database/sql"
import "../query_gen/lib"

var Rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error)
}

type SQLReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLReplyStore(acc *qgen.Accumulator) (*SQLReplyStore, error) {
	return &SQLReplyStore{
		get:    acc.Select("replies").Columns("tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount").Where("rid = ?").Prepare(),
		create: acc.Insert("replies").Columns("tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy").Fields("?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?").Prepare(),
	}, acc.FirstError()
}

func (store *SQLReplyStore) Get(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress, &reply.LikeCount)
	return &reply, err
}

// TODO: Write a test for this
func (store *SQLReplyStore) Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error) {
	wcount := WordCount(content)
	res, err := store.create.Exec(topic.ID, content, ParseMessage(content, topic.ParentID, "forums"), ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), topic.AddReply(uid)
}
