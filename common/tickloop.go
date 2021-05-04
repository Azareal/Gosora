package common

import (
	"log"
	"strconv"
	"sync/atomic"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

type TickLoop struct {
	HalfSec    *time.Ticker
	Sec        *time.Ticker
	FifteenMin *time.Ticker
	Hour       *time.Ticker
	Day        *time.Ticker

	HalfSecf    func() error
	Secf        func() error
	FifteenMinf func() error
	Hourf       func() error
	Dayf        func() error
}

func NewTickLoop() *TickLoop {
	return &TickLoop{
		// TODO: Write tests for these
		// Run this goroutine once every half second
		HalfSec:    time.NewTicker(time.Second / 2),
		Sec:        time.NewTicker(time.Second),
		FifteenMin: time.NewTicker(15 * time.Minute),
		Hour:       time.NewTicker(time.Hour),
		Day:        time.NewTicker(time.Hour * 24),
	}
}

func (l *TickLoop) Loop() {
	r := func(e error) {
		if e != nil {
			LogError(e)
		}
	}
	for {
		select {
		case <-l.HalfSec.C:
			r(l.HalfSecf())
		case <-l.Sec.C:
			r(l.Secf())
		case <-l.FifteenMin.C:
			r(l.FifteenMinf())
		case <-l.Hour.C:
			r(l.Hourf())
		// TODO: Handle the instance going down a lot better
		case <-l.Day.C:
			r(l.Dayf())
		}
	}
}

var ErrDBDown = errors.New("The database is down")

func StartTick() (abort bool) {
	db := qgen.Builder.GetConn()
	isDBDown := atomic.LoadInt32(&IsDBDown)
	if e := db.Ping(); e != nil {
		// TODO: There's a bit of a race here, but it doesn't matter if this error appears multiple times in the logs as it's capped at three times, we just want to cut it down 99% of the time
		if isDBDown == 0 {
			db.SetConnMaxLifetime(time.Second / 2) // Drop all the connections and start over
			LogWarning(e, ErrDBDown.Error())
		}
		atomic.StoreInt32(&IsDBDown, 1)
		return true
	}
	if isDBDown == 1 {
		log.Print("The database is back")
	}
	//db.SetConnMaxLifetime(time.Second * 60 * 5) // Make this infinite as the temporary lifetime change will purge the stale connections?
	db.SetConnMaxLifetime(-1)
	atomic.StoreInt32(&IsDBDown, 0)
	return false
}

// TODO: Move these into DailyTick() methods?
func asmMatches() error {
	// TODO: Find a more efficient way of doing this
	return qgen.NewAcc().Select("activity_stream").Cols("asid").EachInt(func(asid int) error {
		if ActivityMatches.CountAsid(asid) > 0 {
			return nil
		}
		return Activity.Delete(asid)
	})
}

// TODO: Name the tasks so we can figure out which one it was when something goes wrong? Or maybe toss it up WithStack down there?
func RunTasks(tasks []func() error) error {
	for _, task := range tasks {
		if e := task(); e != nil {
			return e
		}
	}
	return nil
}

/*func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		replyStmts = ReplyStmts{
			isLiked:                acc.Select("likes").Columns("targetItem").Where("sentBy=? and targetItem=? and targetType='replies'").Prepare(),
		}
		return acc.FirstError()
	})
}*/

func StartupTasks() (e error) {
	r := func(ee error) {
		if e == nil {
			e = ee
		}
	}
	if Config.DisableRegLog {
		r(RegLogs.Purge())
	}
	if Config.DisableLoginLog {
		r(LoginLogs.Purge())
	}
	if Config.DisablePostIP {
		// TODO: Clear the caches?
		r(Topics.ClearIPs())
		r(Rstore.ClearIPs())
		r(Prstore.ClearIPs())
	}
	if Config.DisablePollIP {
		r(Polls.ClearIPs())
	}
	if Config.DisableLastIP {
		r(Users.ClearLastIPs())
	}
	return e
}

func Dailies() (e error) {
	if e = asmMatches(); e != nil {
		return e
	}
	newAcc := func() *qgen.Accumulator {
		return qgen.NewAcc()
	}
	exec := func(ac qgen.AccExec) {
		if e != nil {
			return
		}
		_, ee := ac.Exec()
		e = ee
	}
	r := func(ee error) {
		if e == nil {
			e = ee
		}
	}

	if Config.LogPruneCutoff > -1 {
		// TODO: Clear the caches?
		if !Config.DisableLoginLog {
			r(LoginLogs.DeleteOlderThanDays(Config.LogPruneCutoff))
		}
		if !Config.DisableRegLog {
			r(RegLogs.DeleteOlderThanDays(Config.LogPruneCutoff))
		}
	}

	if !Config.DisablePostIP && Config.PostIPCutoff > -1 {
		// TODO: Use unixtime to remove this MySQLesque logic?
		f := func(tbl string) {
			exec(newAcc().Update(tbl).Set("ip=''").DateOlderThan("createdAt", Config.PostIPCutoff, "day").Where("ip!=''"))
		}
		f("topics")
		f("replies")
		f("users_replies")
	}

	if !Config.DisablePollIP && Config.PollIPCutoff > -1 {
		// TODO: Use unixtime to remove this MySQLesque logic?
		exec(newAcc().Update("polls_votes").Set("ip=''").DateOlderThan("castAt", Config.PollIPCutoff, "day").Where("ip!=''"))

		// TODO: Find some way of purging the ip data in polls_votes without breaking any anti-cheat measures which might be running... maybe hash it instead?
	}

	// TODO: lastActiveAt isn't currently set, so we can't rely on this to purge last_ips of users who haven't been on in a while
	if !Config.DisableLastIP && Config.LastIPCutoff > 0 {
		//exec(newAcc().Update("users").Set("last_ip='0'").DateOlderThan("lastActiveAt",c.Config.PostIPCutoff,"day").Where("last_ip!='0'"))
		mon := time.Now().Month()
		exec(newAcc().Update("users").Set("last_ip=''").Where("last_ip!='' AND last_ip NOT LIKE '" + strconv.Itoa(int(mon)) + "-%'"))
	}

	if e != nil {
		return e
	}
	if e = Tasks.Day.Run(); e != nil {
		return e
	}
	e = ForumActionStore.DailyTick()
	if e != nil {
		return e
	}

	{
		e := Meta.SetInt64("lastDaily", time.Now().Unix())
		if e != nil {
			return e
		}
	}

	return nil
}