package common

import (
	"database/sql"

	"../query_gen/lib"
)

var Prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*Reply, error)
	Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error)
}

// TODO: Refactor this to stop using the global stmt store
// TODO: Add more methods to this like Create()
type SQLProfileReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLProfileReplyStore() (*SQLProfileReplyStore, error) {
	acc := qgen.Builder.Accumulator()
	return &SQLProfileReplyStore{
		get:    acc.SimpleSelect("users_replies", "uid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress", "rid = ?", "", ""),
		create: acc.SimpleInsert("users_replies", "uid, content, parsed_content, createdAt, createdBy, ipaddress", "?,?,?,UTC_TIMESTAMP(),?,?"),
	}, acc.FirstError()
}

func (store *SQLProfileReplyStore) Get(id int) (*Reply, error) {
	reply := Reply{ID: id}
	err := store.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress)
	return &reply, err
}

func (store *SQLProfileReplyStore) Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error) {
	res, err := store.create.Exec(profileID, content, ParseMessage(content, 0, ""), createdBy, ipaddress)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Should we reload the user?
	return int(lastID), err
}
