/*
*
*	Gosora Task System
*	Copyright Azareal 2017 - 2019
*
 */
package common

import (
	"database/sql"
	"log"
	"time"

	"github.com/Azareal/Gosora/query_gen"
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
// TODO: Be a little more granular with the synchronisation
// TODO: Synchronise more things
// TODO: Does this even work?
func HandleServerSync() error {
	// We don't want to run any unnecessary queries when there is nothing to synchronise
	if Config.ServerCount == 1 {
		return nil
	}

	var lastUpdate time.Time
	err := taskStmts.getSync.QueryRow().Scan(&lastUpdate)
	if err != nil {
		return err
	}

	if lastUpdate.After(lastSync) {
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
		err = WordFilters.ReloadAll()
		if err != nil {
			log.Print("Unable to reload the word filters")
			return err
		}
	}
	return nil
}
