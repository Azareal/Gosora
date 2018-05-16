package common

import (
	"database/sql"

	"../query_gen/lib"
)

var ModLogs LogStore
var AdminLogs LogStore

type LogItem struct {
	Action      string
	ElementID   int
	ElementType string
	IPAddress   string
	ActorID     int
	DoneAt      string
}

type LogStore interface {
	Create(action string, elementID int, elementType string, ipaddress string, actorID int) (err error)
	GlobalCount() int
	GetOffset(offset int, perPage int) (logs []LogItem, err error)
}

type SQLModLogStore struct {
	create    *sql.Stmt
	count     *sql.Stmt
	getOffset *sql.Stmt
}

func NewModLogStore(acc *qgen.Accumulator) (*SQLModLogStore, error) {
	return &SQLModLogStore{
		create:    acc.Insert("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		count:     acc.Count("moderation_logs").Prepare(),
		getOffset: acc.Select("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

// TODO: Make a store for this?
func (store *SQLModLogStore) Create(action string, elementID int, elementType string, ipaddress string, actorID int) (err error) {
	_, err = store.create.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}

func (store *SQLModLogStore) GlobalCount() (logCount int) {
	err := store.count.QueryRow().Scan(&logCount)
	if err != nil {
		LogError(err)
	}
	return logCount
}

func buildLogList(rows *sql.Rows) (logs []LogItem, err error) {
	for rows.Next() {
		var log LogItem
		err := rows.Scan(&log.Action, &log.ElementID, &log.ElementType, &log.IPAddress, &log.ActorID, &log.DoneAt)
		if err != nil {
			return logs, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (store *SQLModLogStore) GetOffset(offset int, perPage int) (logs []LogItem, err error) {
	rows, err := store.getOffset.Query(offset, perPage)
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
	return &SQLAdminLogStore{
		create:    acc.Insert("administration_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		count:     acc.Count("administration_logs").Prepare(),
		getOffset: acc.Select("administration_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

// TODO: Make a store for this?
func (store *SQLAdminLogStore) Create(action string, elementID int, elementType string, ipaddress string, actorID int) (err error) {
	_, err = store.create.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}

func (store *SQLAdminLogStore) GlobalCount() (logCount int) {
	err := store.count.QueryRow().Scan(&logCount)
	if err != nil {
		LogError(err)
	}
	return logCount
}

func (store *SQLAdminLogStore) GetOffset(offset int, perPage int) (logs []LogItem, err error) {
	rows, err := store.getOffset.Query(offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()
	return buildLogList(rows)
}
