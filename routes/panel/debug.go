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

	pi := c.PanelDebugPage{basePage, goVersion, dbVersion, uptime, openConnCount, qgen.Builder.GetAdapter().GetName(), goroutines, cpus, memStats}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage,"panel_dashboard_right","","panel_debug", pi})
}
