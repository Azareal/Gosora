/*
*
*	Gosora Task System
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"log"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type TaskStmts struct {
	getExpiredScheduledGroups *sql.Stmt
	getSync                   *sql.Stmt
}

var ScheduledHalfSecondTasks []func() error
var ScheduledSecondTasks []func() error
var ScheduledFifteenMinuteTasks []func() error
var ScheduledHourTasks []func() error
var ShutdownTasks []func() error
var taskStmts TaskStmts
var lastSync time.Time

// TODO: Add a TaskInits.Add
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

// AddScheduledHalfSecondTask is not concurrency safe
func AddScheduledHalfSecondTask(task func() error) {
	ScheduledHalfSecondTasks = append(ScheduledHalfSecondTasks, task)
}

// AddScheduledSecondTask is not concurrency safe
func AddScheduledSecondTask(task func() error) {
	ScheduledSecondTasks = append(ScheduledSecondTasks, task)
}

// AddScheduledFifteenMinuteTask is not concurrency safe
func AddScheduledFifteenMinuteTask(task func() error) {
	ScheduledFifteenMinuteTasks = append(ScheduledFifteenMinuteTasks, task)
}

// AddScheduledHourTask is not concurrency safe
func AddScheduledHourTask(task func() error) {
	ScheduledHourTasks = append(ScheduledHourTasks, task)
}

// AddShutdownTask is not concurrency safe
func AddShutdownTask(task func() error) {
	ShutdownTasks = append(ShutdownTasks, task)
}

// ScheduledHalfSecondTaskCount is not concurrency safe
func ScheduledHalfSecondTaskCount() int {
	return len(ScheduledHalfSecondTasks)
}

// ScheduledSecondTaskCount is not concurrency safe
func ScheduledSecondTaskCount() int {
	return len(ScheduledSecondTasks)
}

// ScheduledFifteenMinuteTaskCount is not concurrency safe
func ScheduledFifteenMinuteTaskCount() int {
	return len(ScheduledFifteenMinuteTasks)
}

// ScheduledHourTaskCount is not concurrency safe
func ScheduledHourTaskCount() int {
	return len(ScheduledHourTasks)
}

// ShutdownTaskCount is not concurrency safe
func ShutdownTaskCount() int {
	return len(ShutdownTasks)
}

// TODO: Use AddScheduledSecondTask
func HandleExpiredScheduledGroups() error {
	rows, e := taskStmts.getExpiredScheduledGroups.Query()
	if e != nil {
		return e
	}
	defer rows.Close()

	var uid int
	for rows.Next() {
		if e := rows.Scan(&uid); e != nil {
			return e
		}
		// Sneaky way of initialising a *User, please use the methods on the UserStore instead
		user := BlankUser()
		user.ID = uid
		e = user.RevertGroupUpdate()
		if e != nil {
			return e
		}
	}
	return rows.Err()
}

// TODO: Use AddScheduledSecondTask
// TODO: Be a little more granular with the synchronisation
// TODO: Synchronise more things
// TODO: Does this even work?
func HandleServerSync() error {
	// We don't want to run any unnecessary queries when there is nothing to synchronise
	if Config.ServerCount == 1 {
		return nil
	}

	var lastUpdate time.Time
	e := taskStmts.getSync.QueryRow().Scan(&lastUpdate)
	if e != nil {
		return e
	}

	if lastUpdate.After(lastSync) {
		e = Forums.LoadForums()
		if e != nil {
			log.Print("Unable to reload the forums")
			return e
		}
		// TODO: Resync the groups
		// TODO: Resync the permissions
		e = LoadSettings()
		if e != nil {
			log.Print("Unable to reload the settings")
			return e
		}
		e = WordFilters.ReloadAll()
		if e != nil {
			log.Print("Unable to reload the word filters")
			return e
		}
	}
	return nil
}
