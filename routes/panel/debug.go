package panel

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
)

func Debug(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "debug", "debug")
	if ferr != nil {
		return ferr
	}

	goVersion := runtime.Version()
	dbVersion := qgen.Builder.DbVersion()
	upDur := time.Since(c.StartTime)
	hours := int(upDur.Hours())
	mins := int(upDur.Minutes())
	secs := int(upDur.Seconds())
	var uptime string
	if hours > 24 {
		days := hours / 24
		hours -= days * 24
		uptime += strconv.Itoa(days) + "d"
		uptime += strconv.Itoa(hours) + "h"
	} else if hours >= 1 {
		mins -= hours * 60
		uptime += strconv.Itoa(hours) + "h"
		uptime += strconv.Itoa(mins) + "m"
	} else if mins >= 1 {
		secs -= mins * 60
		uptime += strconv.Itoa(mins) + "m"
		uptime += strconv.Itoa(secs) + "s"
	}

	dbStats := qgen.Builder.GetConn().Stats()
	openConnCount := dbStats.OpenConnections
	// Disk I/O?
	// TODO: Fetch the adapter from Builder rather than getting it from a global?
	goroutines := runtime.NumGoroutine()
	cpus := runtime.NumCPU()
	httpConns := c.ConnWatch.Count()

	debugTasks := c.DebugPageTasks{c.Tasks.HalfSec.Count(), c.Tasks.Sec.Count(), c.Tasks.FifteenMin.Count(), c.Tasks.Hour.Count(), c.Tasks.Day.Count(), c.Tasks.Shutdown.Count()}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var tlen, ulen, rlen int
	var tcap, ucap, rcap int
	tc := c.Topics.GetCache()
	if tc != nil {
		tlen, tcap = tc.Length(), tc.GetCapacity()
	}
	uc := c.Users.GetCache()
	if uc != nil {
		ulen, ucap = uc.Length(), uc.GetCapacity()
	}
	rc := c.Rstore.GetCache()
	if rc != nil {
		rlen, rcap = rc.Length(), rc.GetCapacity()
	}
	topicListThawed := c.TopicListThaw.Thawed()

	debugCache := c.DebugPageCache{tlen, ulen, rlen, tcap, ucap, rcap, topicListThawed}

	var fErr error
	acc := qgen.NewAcc()
	count := func(tbl string) int {
		if fErr != nil {
			return 0
		}
		c, err := acc.Count(tbl).Total()
		fErr = err
		return c
	}

	// TODO: Call Count on an attachment store
	attachs := count("attachments")
	// TODO: Implement a PollStore and call Count on that instead
	//polls := count("polls")
	polls := c.Polls.Count()
	//pollsOptions := count("polls_options") // TODO: Add this
	//pollsVotes := count("polls_votes") // TODO: Add this

	//loginLogs := count("login_logs")
	loginLogs := c.LoginLogs.Count()
	//regLogs := count("registration_logs")
	regLogs := c.RegLogs.Count()
	//modLogs := count("moderation_logs")
	modLogs := c.ModLogs.Count()
	//adminLogs := count("administration_logs")
	adminLogs := c.AdminLogs.Count()

	views := count("viewchunks")
	viewsAgents := count("viewchunks_agents")
	viewsForums := count("viewchunks_forums")
	viewsLangs := count("viewchunks_langs")
	viewsReferrers := count("viewchunks_referrers")
	viewsSystems := count("viewchunks_systems")
	postChunks := count("postchunks")
	topicChunks := count("topicchunks")
	if fErr != nil {
		return c.InternalError(fErr, w, r)
	}

	debugDatabase := c.DebugPageDatabase{c.Topics.Count(), c.Users.Count(), c.Rstore.Count(), c.Prstore.Count(), c.Activity.Count(), c.Likes.Count(), attachs, polls, loginLogs, regLogs, modLogs, adminLogs, views, viewsAgents, viewsForums, viewsLangs, viewsReferrers, viewsSystems, postChunks, topicChunks}

	dirSize := func(path string) int {
		if fErr != nil {
			return 0
		}
		c, err := c.DirSize(path)
		fErr = err
		return c
	}

	staticSize := dirSize("./public/")
	attachSize := dirSize("./attachs/")
	uploadsSize := dirSize("./uploads/")
	logsSize := dirSize(c.Config.LogDir)
	backupsSize := dirSize("./backups/")
	if fErr != nil {
		return c.InternalError(fErr, w, r)
	}
	// TODO: How can we measure this without freezing up the entire page?
	//gitSize, _ := c.DirSize("./.git")
	gitSize := 0

	debugDisk := c.DebugPageDisk{staticSize, attachSize, uploadsSize, logsSize, backupsSize, gitSize}

	pi := c.PanelDebugPage{bp, goVersion, dbVersion, uptime, openConnCount, qgen.Builder.GetAdapter().GetName(), goroutines, cpus, httpConns, debugTasks, memStats, debugCache, debugDatabase, debugDisk}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_dashboard_right", "debug_page", "panel_debug", pi})
}

func DebugTasks(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "debug", "debug")
	if ferr != nil {
		return ferr
	}

	var tasks []c.PanelTaskTask
	var taskTypes []c.PanelTaskType

	pi := c.PanelTaskPage{bp, tasks, taskTypes}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_dashboard_right", "debug_page", "panel_debug_task", pi})
}
