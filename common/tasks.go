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

var Tasks *ScheduledTasks

type TaskSet interface {
	Add(func() error)
	GetList() []func() error
	Run() error
	Count() int
}

type DefaultTaskSet struct {
	Tasks []func() error
}

func (s *DefaultTaskSet) Add(task func() error) {
	s.Tasks = append(s.Tasks, task)
}

func (s *DefaultTaskSet) GetList() []func() error {
	return s.Tasks
}

func (s *DefaultTaskSet) Run() error {
	for _, task := range s.Tasks {
		if e := task(); e != nil {
			return e
		}
	}
	return nil
}

func (s *DefaultTaskSet) Count() int {
	return len(s.Tasks)
}

type ScheduledTasks struct {
	HalfSec    TaskSet
	Sec        TaskSet
	FifteenMin TaskSet
	Hour       TaskSet
	Day        TaskSet
	Shutdown   TaskSet
}

func NewScheduledTasks() *ScheduledTasks {
	return &ScheduledTasks{
		HalfSec:    &DefaultTaskSet{},
		Sec:        &DefaultTaskSet{},
		FifteenMin: &DefaultTaskSet{},
		Hour:       &DefaultTaskSet{},
		Day:        &DefaultTaskSet{},
		Shutdown:   &DefaultTaskSet{},
	}
}

/*var ScheduledHalfSecondTasks []func() error
var ScheduledSecondTasks []func() error
var ScheduledFifteenMinuteTasks []func() error
var ScheduledHourTasks []func() error
var ScheduledDayTasks []func() error
var ShutdownTasks []func() error*/
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
/*func AddScheduledHalfSecondTask(task func() error) {
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

// AddScheduledDayTask is not concurrency safe
func AddScheduledDayTask(task func() error) {
	ScheduledDayTasks = append(ScheduledDayTasks, task)
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

// ScheduledDayTaskCount is not concurrency safe
func ScheduledDayTaskCount() int {
	return len(ScheduledDayTasks)
}

// ShutdownTaskCount is not concurrency safe
func ShutdownTaskCount() int {
	return len(ShutdownTasks)
}*/

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
		if e = user.RevertGroupUpdate(); e != nil {
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
		if e = Forums.LoadForums(); e != nil {
			log.Print("Unable to reload the forums")
			return e
		}
		// TODO: Resync the groups
		// TODO: Resync the permissions
		if e = LoadSettings(); e != nil {
			log.Print("Unable to reload the settings")
			return e
		}
		if e = WordFilters.ReloadAll(); e != nil {
			log.Print("Unable to reload the word filters")
			return e
		}
	}
	return nil
}
