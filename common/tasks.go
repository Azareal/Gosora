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

var ScheduledSecondTasks []func() error
var ScheduledFifteenMinuteTasks []func() error
var taskStmts TaskStmts
var lastSync time.Time

func init() {
	lastSync = time.Now()
	DbInits.Add(func(acc *qgen.Accumulator) error {
		taskStmts = TaskStmts{
			getExpiredScheduledGroups: acc.Select("users_groups_scheduler").Columns("uid").Where("UTC_TIMESTAMP() > revert_at AND temporary = 1").Prepare(),
			getSync:                   acc.Select("sync").Columns("last_update").Prepare(),
		}
		return acc.FirstError()
	})
}

// AddScheduledSecondTask is not concurrency safe
func AddScheduledSecondTask(task func() error) {
	ScheduledSecondTasks = append(ScheduledSecondTasks, task)
}

// AddScheduledFifteenMinuteTask is not concurrency safe
func AddScheduledFifteenMinuteTask(task func() error) {
	ScheduledFifteenMinuteTasks = append(ScheduledFifteenMinuteTasks, task)
}

// TODO: Use AddScheduledSecondTask
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

// TODO: Use AddScheduledSecondTask
func HandleServerSync() error {
	var lastUpdate time.Time
	err := taskStmts.getSync.QueryRow().Scan(&lastUpdate)
	if err != nil {
		return err
	}

	if lastUpdate.After(lastSync) {
		// TODO: A more granular sync
		err = Forums.LoadForums()
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
