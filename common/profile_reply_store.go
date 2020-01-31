package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*ProfileReply, error)
	Exists(id int) bool
	Create(profileID int, content string, createdBy int, ip string) (id int, err error)
	Count() (count int)
}

// TODO: Refactor this to stop using the global stmt store
// TODO: Add more methods to this like Create()
type SQLProfileReplyStore struct {
	get    *sql.Stmt
	exists *sql.Stmt
	create *sql.Stmt
	count  *sql.Stmt
}

func NewSQLProfileReplyStore(acc *qgen.Accumulator) (*SQLProfileReplyStore, error) {
	ur := "users_replies"
	return &SQLProfileReplyStore{
		get:    acc.Select(ur).Columns("uid, content, createdBy, createdAt, lastEdit, lastEditBy, ip").Where("rid=?").Prepare(),
		exists: acc.Exists(ur, "rid").Prepare(),
		create: acc.Insert(ur).Columns("uid, content, parsed_content, createdAt, createdBy, ip").Fields("?,?,?,UTC_TIMESTAMP(),?,?").Prepare(),
		count:  acc.Count(ur).Prepare(),
	}, acc.FirstError()
}

func (s *SQLProfileReplyStore) Get(id int) (*ProfileReply, error) {
	r := ProfileReply{ID: id}
	err := s.get.QueryRow(id).Scan(&r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IP)
	return &r, err
}

func (s *SQLProfileReplyStore) Exists(id int) bool {
	err := s.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

func (s *SQLProfileReplyStore) Create(profileID int, content string, createdBy int, ip string) (id int, err error) {
	if Config.DisablePostIP {
		ip = "0"
	}
	res, err := s.create.Exec(profileID, content, ParseMessage(content, 0, "", nil), createdBy, ip)
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
