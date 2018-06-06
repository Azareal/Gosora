package panel

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"../../common"
	"../../query_gen/lib"
)

func Debug(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("panel_debug")

	goVersion := runtime.Version()
	dbVersion := qgen.Builder.DbVersion()
	var uptime string
	upDuration := time.Since(common.StartTime)
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

	pi := common.PanelDebugPage{&common.BasePanelPage{header, stats, "debug", common.ReportForumID}, goVersion, dbVersion, uptime, openConnCount, qgen.Builder.GetAdapter().GetName()}
	return panelRenderTemplate("panel_debug", w, r, user, &pi)
}
