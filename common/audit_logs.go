package common

import (
	"database/sql"

	"../query_gen/lib"
)

type LogStmts struct {
	addModLogEntry   *sql.Stmt
	addAdminLogEntry *sql.Stmt
}

var logStmts LogStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		logStmts = LogStmts{
			addModLogEntry:   acc.Insert("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
			addAdminLogEntry: acc.Insert("administration_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Make a store for this?
func AddModLog(action string, elementID int, elementType string, ipaddress string, actorID int) (err error) {
	_, err = logStmts.addModLogEntry.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}

// TODO: Make a store for this?
func AddAdminLog(action string, elementID string, elementType int, ipaddress string, actorID int) (err error) {
	_, err = logStmts.addAdminLogEntry.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}
