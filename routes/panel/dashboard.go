package panel

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"runtime"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/gopsutil/mem"
	"github.com/pkg/errors"
)

type dashStmts struct {
	todaysPostCount         *sql.Stmt
	todaysTopicCount        *sql.Stmt
	todaysTopicCountByForum *sql.Stmt
	todaysNewUserCount      *sql.Stmt
	weeklyTopicCountByForum *sql.Stmt
}

// TODO: Stop hard-coding these queries
func dashMySQLStmts() (stmts dashStmts, err error) {
	db := qgen.Builder.GetConn()

	var prepareStmt = func(table string, ext string, dur string) *sql.Stmt {
		if err != nil {
			return nil
		}
		stmt, ierr := db.Prepare("select count(*) from " + table + " where createdAt BETWEEN (utc_timestamp() - interval 1 "+dur+") and utc_timestamp() " + ext)
		err = errors.WithStack(ierr)
		return stmt
	}

	stmts.todaysPostCount = prepareStmt("replies", "","day")
	stmts.todaysTopicCount = prepareStmt("topics", "","day")
	stmts.todaysNewUserCount = prepareStmt("users", "","day")
	stmts.todaysTopicCountByForum = prepareStmt("topics", " and parentID = ?","day")
	stmts.weeklyTopicCountByForum = prepareStmt("topics", " and parentID = ?","week")

	return stmts, err
}

// TODO: Stop hard-coding these queries
func dashMSSQLStmts() (stmts dashStmts, err error) {
	db := qgen.Builder.GetConn()

	var prepareStmt = func(table string, ext string, dur string) *sql.Stmt {
		if err != nil {
			return nil
		}
		stmt, ierr := db.Prepare("select count(*) from " + table + " where createdAt >= DATEADD("+dur+", -1, GETUTCDATE())" + ext)
		err = errors.WithStack(ierr)
		return stmt
	}

	stmts.todaysPostCount = prepareStmt("replies", "","DAY")
	stmts.todaysTopicCount = prepareStmt("topics", "","DAY")
	stmts.todaysNewUserCount = prepareStmt("users", "","DAY")
	stmts.todaysTopicCountByForum = prepareStmt("topics", " and parentID = ?","DAY")
	stmts.weeklyTopicCountByForum = prepareStmt("topics", " and parentID = ?","WEEK")

	return stmts, err
}

type GE = c.GridElement

func Dashboard(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "dashboard", "dashboard")
	if ferr != nil {
		return ferr
	}

	// We won't calculate this on the spot anymore, as the system doesn't seem to like it if we do multiple fetches simultaneously. Should we constantly calculate this on a background thread? Perhaps, the watchdog to scale back heavy features under load? One plus side is that we'd get immediate CPU percentages here instead of waiting it to kick in with WebSockets
	var cpustr = "Unknown"
	var cpuColour string

	lessThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number < lowerBound:
			return "stat_green"
		case number < midBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	var ramstr, ramColour string
	memres, err := mem.VirtualMemory()
	if err != nil {
		ramstr = "Unknown"
	} else {
		totalCount, totalUnit := c.ConvertByteUnit(float64(memres.Total))
		usedCount := c.ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

		// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
		var totstr string
		if (totalCount - float64(int(totalCount))) > 0.85 {
			usedCount += 1.0 - (totalCount - float64(int(totalCount)))
			totstr = strconv.Itoa(int(totalCount) + 1)
		} else {
			totstr = fmt.Sprintf("%.1f", totalCount)
		}

		if usedCount > totalCount {
			usedCount = totalCount
		}
		ramstr = fmt.Sprintf("%.1f", usedCount) + " / " + totstr + totalUnit

		ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
		ramColour = lessThanSwitch(int(ramperc), 50, 75)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memCount, memUnit := c.ConvertByteUnit(float64(m.Sys))

	greaterThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number > midBound:
			return "stat_green"
		case number > lowerBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	// TODO: Add a stat store for this?
	var intErr error
	var extractStat = func(stmt *sql.Stmt, args ...interface{}) (stat int) {
		err := stmt.QueryRow(args...).Scan(&stat)
		if err != nil && err != sql.ErrNoRows {
			intErr = err
		}
		return stat
	}

	var stmts dashStmts
	switch qgen.Builder.GetAdapter().GetName() {
	case "mysql":
		stmts, err = dashMySQLStmts()
	case "mssql":
		stmts, err = dashMSSQLStmts()
	default:
		return c.InternalError(errors.New("Unknown database adapter on dashboard"), w, r)
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Allow for more complex phrase structures than just suffixes
	var postCount = extractStat(stmts.todaysPostCount)
	var postInterval = p.GetTmplPhrase("panel_dashboard_day_suffix")
	var postColour = greaterThanSwitch(postCount, 5, 25)

	var topicCount = extractStat(stmts.todaysTopicCount)
	var topicInterval = p.GetTmplPhrase("panel_dashboard_day_suffix")
	var topicColour = greaterThanSwitch(topicCount, 0, 8)

	var reportCount = extractStat(stmts.weeklyTopicCountByForum, c.ReportForumID)
	var reportInterval = p.GetTmplPhrase("panel_dashboard_week_suffix")

	var newUserCount = extractStat(stmts.todaysNewUserCount)
	var newUserInterval = p.GetTmplPhrase("panel_dashboard_week_suffix")

	// Did any of the extractStats fail?
	if intErr != nil {
		return c.InternalError(intErr, w, r)
	}

	var gridElements = []GE{}
	var addElem = func(id string, href string, body string, order int, class string, back string, textColour string, tooltip string) {
		gridElements = append(gridElements, GE{id,href,body,order,class,back,textColour,tooltip})
	}

	// TODO: Implement a check for new versions of Gosora
	// TODO: Localise this
	//addElem("dash-version", "", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)")
	addElem("dash-version", "","v" + c.SoftwareVersion.String(), 0, "grid_istat", "", "", "")

	addElem("dash-cpu","", p.GetTmplPhrasef("panel_dashboard_cpu",cpustr), 1, "grid_istat " + cpuColour, "", "", p.GetTmplPhrase("panel_dashboard_cpu_desc"))
	addElem("dash-ram","", p.GetTmplPhrasef("panel_dashboard_ram",ramstr), 2, "grid_istat " + ramColour, "", "", p.GetTmplPhrase("panel_dashboard_ram_desc"))
	addElem("dash-memused","/panel/analytics/memory/", p.GetTmplPhrasef("panel_dashboard_memused",memCount, memUnit), 2, "grid_istat", "", "", p.GetTmplPhrase("panel_dashboard_memused_desc"))

	if c.EnableWebsockets {
		uonline := c.WsHub.UserCount()
		gonline := c.WsHub.GuestCount()
		totonline := uonline + gonline
		//reqCount := 0

		var onlineColour = greaterThanSwitch(totonline, 3, 10)
		var onlineGuestsColour = greaterThanSwitch(gonline, 1, 10)
		var onlineUsersColour = greaterThanSwitch(uonline, 1, 5)

		totonline, totunit := c.ConvertFriendlyUnit(totonline)
		uonline, uunit := c.ConvertFriendlyUnit(uonline)
		gonline, gunit := c.ConvertFriendlyUnit(gonline)

		addElem("dash-totonline", "",p.GetTmplPhrasef("panel_dashboard_online", totonline, totunit), 3, "grid_stat " + onlineColour, "", "", p.GetTmplPhrase("panel_dashboard_online_desc"))
		addElem("dash-gonline","", p.GetTmplPhrasef("panel_dashboard_guests_online", gonline, gunit), 4, "grid_stat " + onlineGuestsColour, "", "", p.GetTmplPhrase("panel_dashboard_guests_online_desc"))
		addElem("dash-uonline","", p.GetTmplPhrasef("panel_dashboard_users_online", uonline, uunit), 5, "grid_stat " + onlineUsersColour, "", "", p.GetTmplPhrase("panel_dashboard_users_online_desc"))
		//addElem("dash-reqs","", strconv.Itoa(reqCount) + " reqs / second", 7, "grid_stat grid_end_group " + topicColour, "", "", "The number of requests over the last 24 hours")
	}

	addElem("dash-postsperday", "",strconv.Itoa(postCount) + " posts" + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours")
	addElem("dash-topicsperday", "",strconv.Itoa(topicCount) + " topics" + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours")
	addElem("dash-totonlineperday","", "?? online / day", 8, "grid_stat stat_disabled", "", "", p.GetTmplPhrase("panel_dashboard_coming_soon") /*, "The people online over the last 24 hours"*/)

	addElem("dash-searches","", "?? searches / week", 9, "grid_stat stat_disabled", "", "", p.GetTmplPhrase("panel_dashboard_coming_soon") /*"The number of searches over the last 7 days"*/)
	addElem("dash-newusers","", strconv.Itoa(newUserCount) + " new users" + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days")
	addElem("dash-reports","", strconv.Itoa(reportCount) + " reports" + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days")

	if false {
		addElem("dash-minperuser","", "?? minutes / user / week", 12, "grid_stat stat_disabled", "", "", p.GetTmplPhrase("panel_dashboard_coming_soon") /*"The average number of number of minutes spent by each active user over the last 7 days"*/)
		addElem("dash-visitorsperweek","", "?? visitors / week", 13, "grid_stat stat_disabled", "", "", p.GetTmplPhrase("panel_dashboard_coming_soon") /*"The number of unique visitors we've had over the last 7 days"*/)
		addElem("dash-postsperuser","", "?? posts / user / week", 14, "grid_stat stat_disabled", "", "", p.GetTmplPhrase("panel_dashboard_coming_soon") /*"The average number of posts made by each active user over the past week"*/)
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_dashboard_right","","panel_dashboard", gridElements})
}
