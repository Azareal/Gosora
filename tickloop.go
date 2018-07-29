package main

import (
	"errors"
	"log"
	"sync/atomic"
	"time"

	"./common"
)

// TODO: Name the tasks so we can figure out which one it was when something goes wrong? Or maybe toss it up WithStack down there?
func runTasks(tasks []func() error) {
	for _, task := range tasks {
		err := task()
		if err != nil {
			common.LogError(err)
		}
	}
}

func startTick() (abort bool) {
	var isDBDown = atomic.LoadInt32(&common.IsDBDown)
	err := db.Ping()
	if err != nil {
		// TODO: There's a bit of a race here, but it doesn't matter if this error appears multiple times in the logs as it's capped at three times, we just want to cut it down 99% of the time
		if isDBDown == 0 {
			common.LogWarning(err)
			common.LogWarning(errors.New("The database is down"))
		}
		atomic.StoreInt32(&common.IsDBDown, 1)
		return true
	}
	if isDBDown == 1 {
		log.Print("The database is back")
	}
	atomic.StoreInt32(&common.IsDBDown, 0)
	return false
}

func runHook(name string) {
	err := common.RunTaskHook(name)
	if err != nil {
		common.LogError(err, "Failed at task '"+name+"'")
	}
}

func tickLoop(thumbChan chan bool, halfSecondTicker *time.Ticker, secondTicker *time.Ticker, fifteenMinuteTicker *time.Ticker, hourTicker *time.Ticker) {
	for {
		select {
		case <-halfSecondTicker.C:
			if startTick() {
				continue
			}
			runHook("before_half_second_tick")
			runTasks(common.ScheduledHalfSecondTasks)
			runHook("after_half_second_tick")
		case <-secondTicker.C:
			if startTick() {
				continue
			}
			runHook("before_second_tick")
			go func() { thumbChan <- true }()
			runTasks(common.ScheduledSecondTasks)

			// TODO: Stop hard-coding this
			err := common.HandleExpiredScheduledGroups()
			if err != nil {
				common.LogError(err)
			}

			// TODO: Handle delayed moderation tasks

			// Sync with the database, if there are any changes
			err = common.HandleServerSync()
			if err != nil {
				common.LogError(err)
			}

			// TODO: Manage the TopicStore, UserStore, and ForumStore
			// TODO: Alert the admin, if CPU usage, RAM usage, or the number of posts in the past second are too high
			// TODO: Clean-up alerts with no unread matches which are over two weeks old. Move this to a 24 hour task?
			// TODO: Rescan the static files for changes
			runHook("after_second_tick")
		case <-fifteenMinuteTicker.C:
			if startTick() {
				continue
			}
			runHook("before_fifteen_minute_tick")
			runTasks(common.ScheduledFifteenMinuteTasks)

			// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
			// TODO: Publish scheduled posts.
			runHook("after_fifteen_minute_tick")
		case <-hourTicker.C:
			if startTick() {
				continue
			}
			runHook("before_hour_tick")

			jsToken, err := common.GenerateSafeString(80)
			if err != nil {
				common.LogError(err)
			}
			common.JSTokenBox.Store(jsToken)

			common.OldSessionSigningKeyBox.Store(common.SessionSigningKeyBox.Load().(string)) // TODO: We probably don't need this type conversion
			sessionSigningKey, err := common.GenerateSafeString(80)
			if err != nil {
				common.LogError(err)
			}
			common.SessionSigningKeyBox.Store(sessionSigningKey)

			runTasks(common.ScheduledHourTasks)
			runHook("after_hour_tick")
		}

		// TODO: Handle the daily clean-up.
	}
}
