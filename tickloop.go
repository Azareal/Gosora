package main

import (
	"errors"
	"log"
	"time"
	"strconv"
	"sync/atomic"
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

// TODO: Name the tasks so we can figure out which one it was when something goes wrong? Or maybe toss it up WithStack down there?
func runTasks(tasks []func() error) {
	for _, task := range tasks {
		err := task()
		if err != nil {
			c.LogError(err)
		}
	}
}

func startTick() (abort bool) {
	var isDBDown = atomic.LoadInt32(&c.IsDBDown)
	err := db.Ping()
	if err != nil {
		// TODO: There's a bit of a race here, but it doesn't matter if this error appears multiple times in the logs as it's capped at three times, we just want to cut it down 99% of the time
		if isDBDown == 0 {
			db.SetConnMaxLifetime(time.Second) // Drop all the connections and start over
			c.LogWarning(err)
			c.LogWarning(errors.New("The database is down"))
		}
		atomic.StoreInt32(&c.IsDBDown, 1)
		return true
	}
	if isDBDown == 1 {
		log.Print("The database is back")
	}
	//db.SetConnMaxLifetime(time.Second * 60 * 5) // Make this infinite as the temporary lifetime change will purge the stale connections?
	db.SetConnMaxLifetime(-1)
	atomic.StoreInt32(&c.IsDBDown, 0)
	return false
}

func runHook(name string) {
	err := c.RunTaskHook(name)
	if err != nil {
		c.LogError(err, "Failed at task '"+name+"'")
	}
}

func tickLoop(thumbChan chan bool) {
	lastDailyStr, err := c.Meta.Get("lastDaily")
	// TODO: Report this error back correctly...
	if err != nil && err != sql.ErrNoRows {
		c.LogError(err)
	}
	lastDaily, _ := strconv.ParseInt(lastDailyStr, 10, 64)
	now := time.Now().Unix()
	low := now - (60 * 60 * 24)
	if lastDaily < low {
		dailies()
	}
	
	// TODO: Write tests for these
	// Run this goroutine once every half second
	halfSecondTicker := time.NewTicker(time.Second / 2)
	secondTicker := time.NewTicker(time.Second)
	fifteenMinuteTicker := time.NewTicker(15 * time.Minute)
	hourTicker := time.NewTicker(time.Hour)
	dailyTicker := time.NewTicker(time.Hour * 24)
	for {
		select {
		case <-halfSecondTicker.C:
			if startTick() {
				continue
			}
			runHook("before_half_second_tick")
			runTasks(c.ScheduledHalfSecondTasks)
			runHook("after_half_second_tick")
		case <-secondTicker.C:
			if startTick() {
				continue
			}
			runHook("before_second_tick")
			go func() { thumbChan <- true }()
			runTasks(c.ScheduledSecondTasks)

			// TODO: Stop hard-coding this
			err := c.HandleExpiredScheduledGroups()
			if err != nil {
				c.LogError(err)
			}

			// TODO: Handle delayed moderation tasks

			// Sync with the database, if there are any changes
			err = c.HandleServerSync()
			if err != nil {
				c.LogError(err)
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
			runTasks(c.ScheduledFifteenMinuteTasks)

			// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
			// TODO: Publish scheduled posts.
			runHook("after_fifteen_minute_tick")
		case <-hourTicker.C:
			if startTick() {
				continue
			}
			runHook("before_hour_tick")

			jsToken, err := c.GenerateSafeString(80)
			if err != nil {
				c.LogError(err)
			}
			c.JSTokenBox.Store(jsToken)

			c.OldSessionSigningKeyBox.Store(c.SessionSigningKeyBox.Load().(string)) // TODO: We probably don't need this type conversion
			sessionSigningKey, err := c.GenerateSafeString(80)
			if err != nil {
				c.LogError(err)
			}
			c.SessionSigningKeyBox.Store(sessionSigningKey)

			runTasks(c.ScheduledHourTasks)
			runHook("after_hour_tick")
		// TODO: Handle the instance going down a lot better
		case <-dailyTicker.C:
			dailies()
		}

		// TODO: Handle the daily clean-up.
	}
}

func dailies() {
	// TODO: Find a more efficient way of doing this
	err := qgen.NewAcc().Select("activity_stream").Cols("asid").EachInt(func(asid int) error {
		count, err := qgen.NewAcc().Count("activity_stream_matches").Where("asid = " + strconv.Itoa(asid)).Total()
		if err != sql.ErrNoRows {
			return err
		}
		if count > 0 {
			return nil
		}
		_, err = qgen.NewAcc().Delete("activity_stream").Where("asid = ?").Run(asid)
		return err
	})
	if err != nil && err != sql.ErrNoRows {
		c.LogError(err)
	}

	if c.Config.PostIPCutoff > -1 {
		// TODO: Use unixtime to remove this MySQLesque logic?
		_, err := qgen.NewAcc().Update("topics").Set("ipaddress = '0'").DateOlderThan("createdAt",c.Config.PostIPCutoff,"day").Where("ipaddress != '0'").Exec()
		if err != nil {
			c.LogError(err)
		}

		_, err = qgen.NewAcc().Update("replies").Set("ipaddress = '0'").DateOlderThan("createdAt",c.Config.PostIPCutoff,"day").Where("ipaddress != '0'").Exec()
		if err != nil {
			c.LogError(err)
		}
		
		// TODO: Find some way of purging the ip data in polls_votes without breaking any anti-cheat measures which might be running... maybe hash it instead?

		_, err = qgen.NewAcc().Update("users_replies").Set("ipaddress = '0'").DateOlderThan("createdAt",c.Config.PostIPCutoff,"day").Where("ipaddress != '0'").Exec()
		if err != nil {
			c.LogError(err)
		}

		// TODO: lastActiveAt isn't currently set, so we can't rely on this to purge last_ips of users who haven't been on in a while
		/*_, err = qgen.NewAcc().Update("users").Set("last_ip = '0'").DateOlderThan("lastActiveAt",c.Config.PostIPCutoff,"day").Where("last_ip != '0'").Exec()
		if err != nil {
			c.LogError(err)
		}*/
	}

	err = c.Meta.Set("lastDaily", strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		c.LogError(err)
	}
}