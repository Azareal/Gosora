package main

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/uutils"
	"github.com/pkg/errors"
)

var TickLoop *c.TickLoop

func runHook(name string) error {
	if e := c.RunTaskHook(name); e != nil {
		return errors.Wrap(e, "Failed at task '"+name+"'")
	}
	return nil
}

func deferredDailies() error {
	lastDailyStr, e := c.Meta.Get("lastDaily")
	// TODO: Report this error back correctly...
	if e != nil && e != sql.ErrNoRows {
		return e
	}
	lastDaily, _ := strconv.ParseInt(lastDailyStr, 10, 64)
	low := time.Now().Unix() - (60 * 60 * 24)
	if lastDaily < low {
		if e := c.Dailies(); e != nil {
			return e
		}
	}
	return nil
}

func handleLogLongTick(name string, cn int64) {
	if !c.Dev.LogLongTick {
		return
	}
	dur := time.Duration(uutils.Nanotime() - cn)
	if dur.Seconds() > 5 {
		log.Print("tick " + name + " completed in " + dur.String())
	}
}

func tickLoop(thumbChan chan bool) error {
	tl := c.NewTickLoop()
	TickLoop = tl
	if e := deferredDailies(); e != nil {
		return e
	}
	if e := c.StartupTasks(); e != nil {
		return e
	}

	tick := func(name string, tasks c.TaskSet) error {
		if c.StartTick() {
			return nil
		}
		if e := runHook("before_" + name + "_tick"); e != nil {
			return e
		}
		cn := uutils.Nanotime()
		if e := tasks.Run(); e != nil {
			return e
		}
		handleLogLongTick(name, cn)
		return runHook("after_" + name + "_tick")
	}

	tl.HalfSecf = func() error {
		return tick("half_second", c.Tasks.HalfSec)
	}
	// TODO: Automatically lock topics, if they're really old, and the associated setting is enabled.
	// TODO: Publish scheduled posts.
	tl.FifteenMinf = func() error {
		return tick("fifteen_minute", c.Tasks.FifteenMin)
	}
	// TODO: Handle the instance going down a lot better
	// TODO: Handle the daily clean-up.
	tl.Dayf = func() error {
		if c.StartTick() {
			return nil
		}
		cn := uutils.Nanotime()
		if e := c.Dailies(); e != nil {
			return e
		}
		handleLogLongTick("day", cn)
		return nil
	}

	tl.Secf = func() (e error) {
		if c.StartTick() {
			return nil
		}
		if e = runHook("before_second_tick"); e != nil {
			return e
		}
		cn := uutils.Nanotime()
		go func() { thumbChan <- true }()

		if e = c.Tasks.Sec.Run(); e != nil {
			return e
		}

		// TODO: Stop hard-coding this
		if e = c.HandleExpiredScheduledGroups(); e != nil {
			return e
		}

		// TODO: Handle delayed moderation tasks

		// Sync with the database, if there are any changes
		if e = c.HandleServerSync(); e != nil {
			return e
		}
		handleLogLongTick("second", cn)

		// TODO: Manage the TopicStore, UserStore, and ForumStore
		// TODO: Alert the admin, if CPU usage, RAM usage, or the number of posts in the past second are too high
		// TODO: Clean-up alerts with no unread matches which are over two weeks old. Move this to a 24 hour task?
		// TODO: Rescan the static files for changes
		return runHook("after_second_tick")
	}

	tl.Hourf = func() error {
		if c.StartTick() {
			return nil
		}
		if e := runHook("before_hour_tick"); e != nil {
			return e
		}
		cn := uutils.Nanotime()

		jsToken, e := c.GenerateSafeString(80)
		if e != nil {
			return e
		}
		c.JSTokenBox.Store(jsToken)

		c.OldSessionSigningKeyBox.Store(c.SessionSigningKeyBox.Load().(string)) // TODO: We probably don't need this type conversion
		sessionSigningKey, e := c.GenerateSafeString(80)
		if e != nil {
			return e
		}
		c.SessionSigningKeyBox.Store(sessionSigningKey)

		if e = c.Tasks.Hour.Run(); e != nil {
			return e
		}
		handleLogLongTick("hour", cn)
		return runHook("after_hour_tick")
	}

	go tl.Loop()

	return nil
}

func sched() error {
	ws := errors.WithStack
	schedStr, err := c.Meta.Get("sched")
	// TODO: Report this error back correctly...
	if err != nil && err != sql.ErrNoRows {
		return ws(err)
	}

	if schedStr == "recalc" {
		log.Print("Cleaning up orphaned data.")

		count, err := c.Recalc.Replies()
		if err != nil {
			return ws(err)
		}
		log.Printf("Deleted %d orphaned replies.", count)

		count, err = c.Recalc.Forums()
		if err != nil {
			return ws(err)
		}
		log.Printf("Recalculated %d forum topic counts.", count)

		count, err = c.Recalc.Subscriptions()
		if err != nil {
			return ws(err)
		}
		log.Printf("Deleted %d orphaned subscriptions.", count)

		count, err = c.Recalc.ActivityStream()
		if err != nil {
			return ws(err)
		}
		log.Printf("Deleted %d orphaned activity stream items.", count)

		err = c.Recalc.Users()
		if err != nil {
			return ws(err)
		}
		log.Print("Recalculated user post stats.")

		count, err = c.Recalc.Attachments()
		if err != nil {
			return ws(err)
		}
		log.Printf("Deleted %d orphaned attachments.", count)
	}

	return nil
}
