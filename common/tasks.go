/*
*
*	Gosora Task System
*	Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"database/sql"
	"log"
	"time"

	"../query_gen/lib"
)

type TaskStmts struct {
	getExpiredScheduledGroups *sql.Stmt
	getSync                   *sql.Stmt
}

var taskStmts TaskStmts
var lastSync time.Time

func init() {
	lastSync = time.Now()
	DbInits.Add(func() error {
		acc := qgen.Builder.Accumulator()
		taskStmts = TaskStmts{
			getExpiredScheduledGroups: acc.SimpleSelect("users_groups_scheduler", "uid", "UTC_TIMESTAMP() > revert_at AND temporary = 1", "", ""),
			getSync:                   acc.SimpleSelect("sync", "last_update", "", "", ""),
		}
		return acc.FirstError()
	})
}

func HandleExpiredScheduledGroups() error {
	rows, err := taskStmts.getExpiredScheduledGroups.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uid int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}

		// Sneaky way of initialising a *User, please use the methods on the UserStore instead
		user := BlankUser()
		user.ID = uid
		err = user.RevertGroupUpdate()
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func HandleServerSync() error {
	var lastUpdate time.Time
	err := taskStmts.getSync.QueryRow().Scan(&lastUpdate)
	if err != nil {
		return err
	}

	if lastUpdate.After(lastSync) {
		// TODO: A more granular sync
		err = Fstore.LoadForums()
		if err != nil {
			log.Print("Unable to reload the forums")
			return err
		}
		// TODO: Resync the groups
		// TODO: Resync the permissions
		err = LoadSettings()
		if err != nil {
			log.Print("Unable to reload the settings")
			return err
		}
		err = LoadWordFilters()
		if err != nil {
			log.Print("Unable to reload the word filters")
			return err
		}
	}
	return nil
}
