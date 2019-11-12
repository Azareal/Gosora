package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*ProfileReply, error)
	Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error)
	Count() (count int)
}

// TODO: Refactor this to stop using the global stmt store
// TODO: Add more methods to this like Create()
type SQLProfileReplyStore struct {
	get    *sql.Stmt
	create *sql.Stmt
	count  *sql.Stmt
}

func NewSQLProfileReplyStore(acc *qgen.Accumulator) (*SQLProfileReplyStore, error) {
	ur := "users_replies"
	return &SQLProfileReplyStore{
		get:    acc.Select(ur).Columns("uid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress").Where("rid = ?").Prepare(),
		create: acc.Insert(ur).Columns("uid, content, parsed_content, createdAt, createdBy, ipaddress").Fields("?,?,?,UTC_TIMESTAMP(),?,?").Prepare(),
		count:  acc.Count(ur).Prepare(),
	}, acc.FirstError()
}

func (s *SQLProfileReplyStore) Get(id int) (*ProfileReply, error) {
	r := ProfileReply{ID: id}
	err := s.get.QueryRow(id).Scan(&r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IP)
	return &r, err
}

func (s *SQLProfileReplyStore) Create(profileID int, content string, createdBy int, ipaddress string) (id int, err error) {
	res, err := s.create.Exec(profileID, content, ParseMessage(content, 0, ""), createdBy, ipaddress)
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

// TODO: Write a test for this
// Count returns the total number of topic replies on these forums
func (s *SQLProfileReplyStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
