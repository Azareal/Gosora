package main

import (
	"database/sql"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

// TODO: Name the tasks so we can figure out which one it was when something goes wrong? Or maybe toss it up WithStack down there?
func runTasks(tasks []func() error) {
	for _, task := range tasks {
		if err := task(); err != nil {
			c.LogError(err)
		}
	}
}

func startTick() (abort bool) {
	isDBDown := atomic.LoadInt32(&c.IsDBDown)
	if err := db.Ping(); err != nil {
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
	if err := c.RunTaskHook(name); err != nil {
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
	low := time.Now().Unix() - (60 * 60 * 24)
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
			if err := c.HandleExpiredScheduledGroups(); err != nil {
				c.LogError(err)
			}

			// TODO: Handle delayed moderation tasks

			// Sync with the database, if there are any changes
			if err = c.HandleServerSync(); err != nil {
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

func asmMatches() {
	// TODO: Find a more efficient way of doing this
	acc := qgen.NewAcc()
	countStmt := acc.Count("activity_stream_matches").Where("asid=?").Prepare()
	if err := acc.FirstError(); err != nil {
		c.LogError(err)
		return
	}

	err := acc.Select("activity_stream").Cols("asid").EachInt(func(asid int) error {
		var count int
		err := countStmt.QueryRow(asid).Scan(&count)
		if err != sql.ErrNoRows {
			return err
		}
		if count > 0 {
			return nil
		}
		_, err = qgen.NewAcc().Delete("activity_stream").Where("asid=?").Run(asid)
		return err
	})
	if err != nil && err != sql.ErrNoRows {
		c.LogError(err)
	}
}

func dailies() {
	asmMatches()

	if c.Config.DisableRegLog {
		_, err := qgen.NewAcc().Purge("registration_logs").Exec()
		if err != nil {
			c.LogError(err)
		}
	}
	if c.Config.LogPruneCutoff > -1 {
		f := func(tbl string) {
			_, err := qgen.NewAcc().Delete(tbl).DateOlderThan("doneAt", c.Config.LogPruneCutoff, "day").Run()
			if err != nil {
				c.LogError(err)
			}
		}
		f("login_logs")
		f("registration_logs")
	}

	if c.Config.DisablePostIP {
		f := func(tbl string) {
			_, err := qgen.NewAcc().Update(tbl).Set("ip='0'").Where("ip!='0'").Exec()
			if err != nil {
				c.LogError(err)
			}
		}
		f("topics")
		f("replies")
		f("users_replies")
	} else if c.Config.PostIPCutoff > -1 {
		// TODO: Use unixtime to remove this MySQLesque logic?
		f := func(tbl string) {
			_, err := qgen.NewAcc().Update(tbl).Set("ip='0'").DateOlderThan("createdAt", c.Config.PostIPCutoff, "day").Where("ip!='0'").Exec()
			if err != nil {
				c.LogError(err)
			}
		}
		f("topics")
		f("replies")
		f("users_replies")
	}

	if c.Config.DisablePollIP {
		_, err := qgen.NewAcc().Update("polls_votes").Set("ip='0'").Where("ip!='0'").Exec()
		if err != nil {
			c.LogError(err)
		}
	} else if c.Config.PollIPCutoff > -1 {
		// TODO: Use unixtime to remove this MySQLesque logic?
		_, err := qgen.NewAcc().Update("polls_votes").Set("ip='0'").DateOlderThan("castAt", c.Config.PollIPCutoff, "day").Where("ip!='0'").Exec()
		if err != nil {
			c.LogError(err)
		}

		// TODO: Find some way of purging the ip data in polls_votes without breaking any anti-cheat measures which might be running... maybe hash it instead?
	}

	// TODO: lastActiveAt isn't currently set, so we can't rely on this to purge last_ips of users who haven't been on in a while
	if c.Config.DisableLastIP {
		_, err := qgen.NewAcc().Update("users").Set("last_ip=0").Where("last_ip!=0").Exec()
		if err != nil {
			c.LogError(err)
		}
	} else if c.Config.LastIPCutoff > 0 {
		/*_, err = qgen.NewAcc().Update("users").Set("last_ip='0'").DateOlderThan("lastActiveAt",c.Config.PostIPCutoff,"day").Where("last_ip!='0'").Exec()
		if err != nil {
			c.LogError(err)
		}*/
		mon := time.Now().Month()
		_, err := qgen.NewAcc().Update("users").Set("last_ip=0").Where("last_ip!='0' AND last_ip NOT LIKE '" + strconv.Itoa(int(mon)) + "-%'").Exec()
		if err != nil {
			c.LogError(err)
		}
	}

	{
		err := c.Meta.Set("lastDaily", strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			c.LogError(err)
		}
	}
}

func sched() error {
	schedStr, err := c.Meta.Get("sched")
	// TODO: Report this error back correctly...
	if err != nil && err != sql.ErrNoRows {
		return errors.WithStack(err)
	}

	if schedStr == "recalc" {
		log.Print("Cleaning up orphaned data.")

		count, err := c.Recalc.Replies()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Printf("Deleted %d orphaned replies.", count)

		count, err = c.Recalc.Subscriptions()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Printf("Deleted %d orphaned subscriptions.", count)

		count, err = c.Recalc.ActivityStream()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Printf("Deleted %d orphaned activity stream items.", count)

		err = c.Recalc.Users()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Print("Recalculated user post stats.")

		count, err = c.Recalc.Attachments()
		if err != nil {
			return errors.WithStack(err)
		}
		log.Printf("Deleted %d orphaned attachments.", count)
	}

	return nil
}
