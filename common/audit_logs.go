package common

import (
	"database/sql"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var ModLogs LogStore
var AdminLogs LogStore

type LogItem struct {
	Action      string
	ElementID   int
	ElementType string
	IP          string
	ActorID     int
	DoneAt      string
	Extra       string
}

type LogStore interface {
	Create(action string, elementID int, elementType string, ip string, actorID int) (err error)
	CreateExtra(action string, elementID int, elementType string, ip string, actorID int, extra string) (err error)
	Count() int
	GetOffset(offset, perPage int) (logs []LogItem, err error)
}

type SQLModLogStore struct {
	create    *sql.Stmt
	count     *sql.Stmt
	getOffset *sql.Stmt
}

func NewModLogStore(acc *qgen.Accumulator) (*SQLModLogStore, error) {
	ml := "moderation_logs"
	return &SQLModLogStore{
		create:    acc.Insert(ml).Columns("action, elementID, elementType, ipaddress, actorID, doneAt, extra").Fields("?,?,?,?,?,UTC_TIMESTAMP(),?").Prepare(),
		count:     acc.Count(ml).Prepare(),
		getOffset: acc.Select(ml).Columns("action, elementID, elementType, ipaddress, actorID, doneAt, extra").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

// TODO: Make a store for this?
func (s *SQLModLogStore) Create(action string, elementID int, elementType string, ip string, actorID int) (err error) {
	return s.CreateExtra(action, elementID, elementType, ip, actorID, "")
}

func (s *SQLModLogStore) CreateExtra(action string, elementID int, elementType string, ip string, actorID int, extra string) (err error) {
	_, err = s.create.Exec(action, elementID, elementType, ip, actorID, extra)
	return err
}

func (s *SQLModLogStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func buildLogList(rows *sql.Rows) (logs []LogItem, err error) {
	for rows.Next() {
		var l LogItem
		var doneAt time.Time
		err := rows.Scan(&l.Action, &l.ElementID, &l.ElementType, &l.IP, &l.ActorID, &doneAt, &l.Extra)
		if err != nil {
			return logs, err
		}
		l.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SQLModLogStore) GetOffset(offset, perPage int) (logs []LogItem, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()
	return buildLogList(rows)
}

type SQLAdminLogStore struct {
	create    *sql.Stmt
	count     *sql.Stmt
	getOffset *sql.Stmt
}

func NewAdminLogStore(acc *qgen.Accumulator) (*SQLAdminLogStore, error) {
	al := "administration_logs"
	return &SQLAdminLogStore{
		create:    acc.Insert(al).Columns("action, elementID, elementType, ipaddress, actorID, doneAt, extra").Fields("?,?,?,?,?,UTC_TIMESTAMP(),?").Prepare(),
		count:     acc.Count(al).Prepare(),
		getOffset: acc.Select(al).Columns("action, elementID, elementType, ipaddress, actorID, doneAt, extra").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

// TODO: Make a store for this?
func (s *SQLAdminLogStore) Create(action string, elementID int, elementType string, ip string, actorID int) (err error) {
	return s.CreateExtra(action, elementID, elementType, ip, actorID, "")
}

func (s *SQLAdminLogStore) CreateExtra(action string, elementID int, elementType string, ip string, actorID int, extra string) (err error) {
	_, err = s.create.Exec(action, elementID, elementType, ip, actorID, extra)
	return err
}

func (s *SQLAdminLogStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *SQLAdminLogStore) GetOffset(offset, perPage int) (logs []LogItem, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()
	return buildLogList(rows)
}
