package common

import (
	"database/sql"

	"../query_gen/lib"
)

var ModLogs LogStore
var AdminLogs LogStore

type LogStore interface {
	Create(action string, elementID int, elementType string, ipaddress string, actorID int) (err error)
	GlobalCount() int
}

type SQLModLogStore struct {
	create *sql.Stmt
	count  *sql.Stmt
}

func NewModLogStore() (*SQLModLogStore, error) {
	acc := qgen.Builder.Accumulator()
	return &SQLModLogStore{
		create: acc.Insert("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		count:  acc.Count("moderation_logs").Prepare(),
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

type SQLAdminLogStore struct {
	create *sql.Stmt
	count  *sql.Stmt
}

func NewAdminLogStore() (*SQLAdminLogStore, error) {
	acc := qgen.Builder.Accumulator()
	return &SQLAdminLogStore{
		create: acc.Insert("administration_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		count:  acc.Count("administration_logs").Prepare(),
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
