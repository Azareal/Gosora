package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Prstore ProfileReplyStore

type ProfileReplyStore interface {
	Get(id int) (*ProfileReply, error)
	Exists(id int) bool
	ClearIPs() error
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

	clearIPs *sql.Stmt
}

func NewSQLProfileReplyStore(acc *qgen.Accumulator) (*SQLProfileReplyStore, error) {
	ur := "users_replies"
	return &SQLProfileReplyStore{
		get:    acc.Select(ur).Columns("uid,content,createdBy,createdAt,lastEdit,lastEditBy,ip").Where("rid=?").Stmt(),
		exists: acc.Exists(ur, "rid").Prepare(),
		create: acc.Insert(ur).Columns("uid,content,parsed_content,createdAt,createdBy,ip").Fields("?,?,?,UTC_TIMESTAMP(),?,?").Prepare(),
		count:  acc.Count(ur).Stmt(),

		clearIPs: acc.Update(ur).Set("ip=''").Where("ip!=''").Stmt(),
	}, acc.FirstError()
}

func (s *SQLProfileReplyStore) Get(id int) (*ProfileReply, error) {
	r := ProfileReply{ID: id}
	e := s.get.QueryRow(id).Scan(&r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IP)
	return &r, e
}

func (s *SQLProfileReplyStore) Exists(id int) bool {
	e := s.exists.QueryRow(id).Scan(&id)
	if e != nil && e != ErrNoRows {
		LogError(e)
	}
	return e != ErrNoRows
}

func (s *SQLProfileReplyStore) ClearIPs() error {
	_, e := s.clearIPs.Exec()
	return e
}

func (s *SQLProfileReplyStore) Create(profileID int, content string, createdBy int, ip string) (id int, e error) {
	if Config.DisablePostIP {
		ip = ""
	}
	res, e := s.create.Exec(profileID, content, ParseMessage(content, 0, "", nil, nil), createdBy, ip)
	if e != nil {
		return 0, e
	}
	lastID, e := res.LastInsertId()
	if e != nil {
		return 0, e
	}
	// Should we reload the user?
	return int(lastID), e
}

// TODO: Write a test for this
// Count returns the total number of topic replies on these forums
func (s *SQLProfileReplyStore) Count() (count int) {
	return Count(s.count)
}
