package main

import (
	"database/sql"

	"./query_gen/lib"
)

var prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*Reply, error)
}

// TODO: Refactor this to stop using the global stmt store
// TODO: Add more methods to this like Create()
type SQLProfileReplyStore struct {
	get *sql.Stmt
}

func NewSQLProfileReplyStore() (*SQLProfileReplyStore, error) {
	getUserReplyStmt, err := qgen.Builder.SimpleSelect("users_replies", "uid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress", "rid = ?", "", "")
	if err != nil {
		return nil, err
	}
	return &SQLProfileReplyStore{
		get: getUserReplyStmt,
	}, nil
}

func (store *SQLProfileReplyStore) Get(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress)
	return &reply, err
}
