package panel

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

func Debug(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "debug", "debug")
	if ferr != nil {
		return ferr
	}

	goVersion := runtime.Version()
	dbVersion := qgen.Builder.DbVersion()
	var uptime string
	upDuration := time.Since(c.StartTime)
	hours := int(upDuration.Hours())
	minutes := int(upDuration.Minutes())
	if hours > 24 {
		days := hours / 24
		hours -= days * 24
		uptime += strconv.Itoa(days) + "d"
		uptime += strconv.Itoa(hours) + "h"
	} else if hours >= 1 {
		uptime += strconv.Itoa(hours) + "h"
	}
	uptime += strconv.Itoa(minutes) + "m"

	dbStats := qgen.Builder.GetConn().Stats()
	openConnCount := dbStats.OpenConnections
	// Disk I/O?
	// TODO: Fetch the adapter from Builder rather than getting it from a global?
	goroutines := runtime.NumGoroutine()
	cpus := runtime.NumCPU()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var tlen, ulen, rlen int
	var tcap, ucap, rcap int
	tcache := c.Topics.GetCache()
	if tcache != nil {
		tlen = tcache.Length()
		tcap = tcache.GetCapacity()
	}
	ucache := c.Users.GetCache()
	if ucache != nil {
		ulen = ucache.Length()
		ucap = ucache.GetCapacity()
	}
	rcache := c.Rstore.GetCache()
	if rcache != nil {
		rlen = rcache.Length()
		rcap = rcache.GetCapacity()
	}
	topicListThawed := c.TopicListThaw.Thawed()

	debugCache := c.DebugPageCache{tlen, ulen, rlen, tcap, ucap, rcap, topicListThawed}

	// TODO: Implement a LikeStore and call Count on that instead
	likes, err := qgen.NewAcc().Count("likes").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	// TODO: Call Count on an attachment store
	attachs, err := qgen.NewAcc().Count("attachments").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	// TODO: Implement a PollStore and call Count on that instead
	polls, err := qgen.NewAcc().Count("polls").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	views, err := qgen.NewAcc().Count("viewchunks").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	viewsAgents, err := qgen.NewAcc().Count("viewchunks_agents").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	viewsForums, err := qgen.NewAcc().Count("viewchunks_forums").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	viewsLangs, err := qgen.NewAcc().Count("viewchunks_langs").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	viewsReferrers, err := qgen.NewAcc().Count("viewchunks_referrers").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	viewsSystems, err := qgen.NewAcc().Count("viewchunks_systems").Total()
	if err != nil {
		return c.InternalError(err,w,r)
	}
	debugDatabase := c.DebugPageDatabase{c.Topics.Count(),c.Users.Count(),c.Rstore.Count(),c.Prstore.Count(),c.Activity.Count(),likes,attachs,polls,views,viewsAgents,viewsForums,viewsLangs,viewsReferrers,viewsSystems}

	staticSize, err := c.DirSize("./public/")
	if err != nil {
		return c.InternalError(err,w,r)
	}
	attachSize, err := c.DirSize("./attachs/")
	if err != nil {
		return c.InternalError(err,w,r)
	}
	uploadsSize, err := c.DirSize("./uploads/")
	if err != nil {
		return c.InternalError(err,w,r)
	}
	logsSize, err := c.DirSize("./logs/")
	if err != nil {
		return c.InternalError(err,w,r)
	}
	backupsSize, err := c.DirSize("./backups/")
	if err != nil {
		return c.InternalError(err,w,r)
	}
	debugDisk := c.DebugPageDisk{staticSize,attachSize,uploadsSize,logsSize,backupsSize}

	pi := c.PanelDebugPage{basePage, goVersion, dbVersion, uptime, openConnCount, qgen.Builder.GetAdapter().GetName(), goroutines, cpus, memStats, debugCache, debugDatabase, debugDisk}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_dashboard_right", "debug_page", "panel_debug", pi})
}
