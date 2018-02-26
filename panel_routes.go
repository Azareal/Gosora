package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"./common"
	"./query_gen/lib"
	"github.com/Azareal/gopsutil/mem"
)

// We're trying to reduce the amount of boilerplate in here, so I added these two functions, they might wind up circulating outside this file in the future
func panelSuccessRedirect(dest string, w http.ResponseWriter, r *http.Request, isJs bool) common.RouteError {
	if !isJs {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}
func panelRenderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, user common.User, pi interface{}) common.RouteError {
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &user, pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, tmplName+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelDashboard(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
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
		totalCount, totalUnit := common.ConvertByteUnit(float64(memres.Total))
		usedCount := common.ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

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
	var extractStat = func(stmt *sql.Stmt) (stat int) {
		err := stmt.QueryRow().Scan(&stat)
		if err != nil && err != ErrNoRows {
			intErr = err
		}
		return stat
	}

	var postCount = extractStat(stmts.todaysPostCount)
	var postInterval = "day"
	var postColour = greaterThanSwitch(postCount, 5, 25)

	var topicCount = extractStat(stmts.todaysTopicCount)
	var topicInterval = "day"
	var topicColour = greaterThanSwitch(topicCount, 0, 8)

	var reportCount = extractStat(stmts.todaysReportCount)
	var reportInterval = "week"

	var newUserCount = extractStat(stmts.todaysNewUserCount)
	var newUserInterval = "week"

	// Did any of the extractStats fail?
	if intErr != nil {
		return common.InternalError(intErr, w, r)
	}

	var gridElements = []common.GridElement{
		common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)"},
		common.GridElement{"dash-cpu", "CPU: " + cpustr, 1, "grid_istat " + cpuColour, "", "", "The global CPU usage of this server"},
		common.GridElement{"dash-ram", "RAM: " + ramstr, 2, "grid_istat " + ramColour, "", "", "The global RAM usage of this server"},
	}

	if enableWebsockets {
		uonline := wsHub.userCount()
		gonline := wsHub.guestCount()
		totonline := uonline + gonline
		reqCount := 0

		var onlineColour = greaterThanSwitch(totonline, 3, 10)
		var onlineGuestsColour = greaterThanSwitch(gonline, 1, 10)
		var onlineUsersColour = greaterThanSwitch(uonline, 1, 5)

		totonline, totunit := common.ConvertFriendlyUnit(totonline)
		uonline, uunit := common.ConvertFriendlyUnit(uonline)
		gonline, gunit := common.ConvertFriendlyUnit(gonline)

		gridElements = append(gridElements, common.GridElement{"dash-totonline", strconv.Itoa(totonline) + totunit + " online", 3, "grid_stat " + onlineColour, "", "", "The number of people who are currently online"})
		gridElements = append(gridElements, common.GridElement{"dash-gonline", strconv.Itoa(gonline) + gunit + " guests online", 4, "grid_stat " + onlineGuestsColour, "", "", "The number of guests who are currently online"})
		gridElements = append(gridElements, common.GridElement{"dash-uonline", strconv.Itoa(uonline) + uunit + " users online", 5, "grid_stat " + onlineUsersColour, "", "", "The number of logged-in users who are currently online"})
		gridElements = append(gridElements, common.GridElement{"dash-reqs", strconv.Itoa(reqCount) + " reqs / second", 7, "grid_stat grid_end_group " + topicColour, "", "", "The number of requests over the last 24 hours"})
	}

	gridElements = append(gridElements, common.GridElement{"dash-postsperday", strconv.Itoa(postCount) + " posts / " + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours"})
	gridElements = append(gridElements, common.GridElement{"dash-topicsperday", strconv.Itoa(topicCount) + " topics / " + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours"})
	gridElements = append(gridElements, common.GridElement{"dash-totonlineperday", "20 online / day", 8, "grid_stat stat_disabled", "", "", "Coming Soon!" /*, "The people online over the last 24 hours"*/})

	gridElements = append(gridElements, common.GridElement{"dash-searches", "8 searches / week", 9, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of searches over the last 7 days"*/})
	gridElements = append(gridElements, common.GridElement{"dash-newusers", strconv.Itoa(newUserCount) + " new users / " + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days"})
	gridElements = append(gridElements, common.GridElement{"dash-reports", strconv.Itoa(reportCount) + " reports / " + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days"})

	if false {
		gridElements = append(gridElements, common.GridElement{"dash-minperuser", "2 minutes / user / week", 12, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of number of minutes spent by each active user over the last 7 days"*/})
		gridElements = append(gridElements, common.GridElement{"dash-visitorsperweek", "2 visitors / week", 13, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of unique visitors we've had over the last 7 days"*/})
		gridElements = append(gridElements, common.GridElement{"dash-postsperuser", "5 posts / user / week", 14, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of posts made by each active user over the past week"*/})
	}

	pi := common.PanelDashboardPage{common.GetTitlePhrase("panel_dashboard"), user, headerVars, stats, "dashboard", gridElements}
	return panelRenderTemplate("panel_dashboard", w, r, user, &pi)
}

func routePanelForums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	// TODO: Paginate this?
	var forumList []interface{}
	forums, err := common.Forums.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// ? - Should we generate something similar to the forumView? It might be a little overkill for a page which is rarely loaded in comparison to /forums/
	for _, forum := range forums {
		if forum.Name != "" && forum.ParentID == 0 {
			fadmin := common.ForumAdmin{forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, forum.TopicCount, common.PresetToLang(forum.Preset)}
			if fadmin.Preset == "" {
				fadmin.Preset = "custom"
			}
			forumList = append(forumList, fadmin)
		}
	}
	pi := common.PanelPage{common.GetTitlePhrase("panel_forums"), user, headerVars, stats, "forums", forumList, nil}
	return panelRenderTemplate("panel_forums", w, r, user, &pi)
}

func routePanelForumsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := common.StripInvalidPreset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-name")
	active := (factive == "on" || factive == "1")

	_, err := common.Forums.Create(fname, fdesc, active, fpreset)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	return nil
}

// TODO: Revamp this
func routePanelForumsDelete(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Make this a phrase
	confirmMsg := "Are you sure you want to delete the '" + forum.Name + "' forum?"
	yousure := common.AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := common.PanelPage{common.GetTitlePhrase("panel_delete_forum"), user, headerVars, stats, "forums", tList, yousure}
	if common.RunPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "are_you_sure.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelForumsDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	err = common.Forums.Delete(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	return nil
}

func routePanelForumsEdit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	glist, err := common.Groups.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var gplist []common.GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		// TODO: Don't access the cache on the group directly
		gplist = append(gplist, common.GroupForumPermPreset{group, common.ForumPermsToGroupForumPreset(group.Forums[fid])})
	}

	pi := common.PanelEditForumPage{common.GetTitlePhrase("panel_edit_forum"), user, headerVars, stats, "forums", forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, gplist}
	if common.RunPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-forum-edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	return nil
}

func routePanelForumsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	forumName := r.PostFormValue("forum_name")
	forumDesc := r.PostFormValue("forum_desc")
	forumPreset := common.StripInvalidPreset(r.PostFormValue("forum_preset"))
	forumActive := r.PostFormValue("forum_active")

	var active = false
	if forumActive == "" {
		active = forum.Active
	} else if forumActive == "1" || forumActive == "Show" {
		active = true
	}

	err = forum.Update(forumName, forumDesc, active, forumPreset)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return panelSuccessRedirect("/panel/forums/", w, r, isJs)
}

func routePanelForumsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		return common.LocalErrorJSQ("Invalid Group ID", w, r, user, isJs)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("This forum doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	permPreset := common.StripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	err = forum.SetPreset(permPreset, gid)
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return panelSuccessRedirect("/panel/forums/edit/"+strconv.Itoa(fid), w, r, isJs)
}

// A helper function for the Advanced portion of the Forum Perms Editor
func panelForumPermsExtractDash(paramList string) (fid int, gid int, err error) {
	params := strings.Split(paramList, "-")
	if len(params) != 2 {
		return fid, gid, errors.New("Parameter count mismatch")
	}

	fid, err = strconv.Atoi(params[0])
	if err != nil {
		return fid, gid, errors.New("The provided Forum ID is not a valid number.")
	}

	gid, err = strconv.Atoi(params[1])
	if err != nil {
		err = errors.New("The provided Group ID is not a valid number.")
	}

	return fid, gid, err
}

func routePanelForumsEditPermsAdvance(w http.ResponseWriter, r *http.Request, user common.User, paramList string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}

	fid, gid, err := panelForumPermsExtractDash(paramList)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	forumPerms, err := common.FPStore.Get(fid, gid)
	if err == ErrNoRows {
		return common.LocalError("The requested group doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var formattedPermList []common.NameLangToggle

	// TODO: Load the phrases in bulk for efficiency?
	// TODO: Reduce the amount of code duplication between this and the group editor. Also, can we grind this down into one line or use a code generator to stay current more easily?
	var addNameLangToggle = func(permStr string, perm bool) {
		formattedPermList = append(formattedPermList, common.NameLangToggle{permStr, common.GetLocalPermPhrase(permStr), perm})
	}
	addNameLangToggle("ViewTopic", forumPerms.ViewTopic)
	addNameLangToggle("LikeItem", forumPerms.LikeItem)
	addNameLangToggle("CreateTopic", forumPerms.CreateTopic)
	//<--
	addNameLangToggle("EditTopic", forumPerms.EditTopic)
	addNameLangToggle("DeleteTopic", forumPerms.DeleteTopic)
	addNameLangToggle("CreateReply", forumPerms.CreateReply)
	addNameLangToggle("EditReply", forumPerms.EditReply)
	addNameLangToggle("DeleteReply", forumPerms.DeleteReply)
	addNameLangToggle("PinTopic", forumPerms.PinTopic)
	addNameLangToggle("CloseTopic", forumPerms.CloseTopic)
	addNameLangToggle("MoveTopic", forumPerms.MoveTopic)

	pi := common.PanelEditForumGroupPage{common.GetTitlePhrase("panel_edit_forum"), user, headerVars, stats, "forums", forum.ID, gid, forum.Name, forum.Desc, forum.Active, forum.Preset, formattedPermList}
	if common.RunPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-forum-edit-perms.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	return nil
}

func routePanelForumsEditPermsAdvanceSubmit(w http.ResponseWriter, r *http.Request, user common.User, paramList string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, gid, err := panelForumPermsExtractDash(paramList)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	forumPerms, err := common.FPStore.GetCopy(fid, gid)
	if err == ErrNoRows {
		return common.LocalError("The requested group doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var extractPerm = func(name string) bool {
		pvalue := r.PostFormValue("forum-perm-" + name)
		return (pvalue == "1")
	}

	// TODO: Generate this code?
	forumPerms.ViewTopic = extractPerm("ViewTopic")
	forumPerms.LikeItem = extractPerm("LikeItem")
	forumPerms.CreateTopic = extractPerm("CreateTopic")
	forumPerms.EditTopic = extractPerm("EditTopic")
	forumPerms.DeleteTopic = extractPerm("DeleteTopic")
	forumPerms.CreateReply = extractPerm("CreateReply")
	forumPerms.EditReply = extractPerm("EditReply")
	forumPerms.DeleteReply = extractPerm("DeleteReply")
	forumPerms.PinTopic = extractPerm("PinTopic")
	forumPerms.CloseTopic = extractPerm("CloseTopic")
	forumPerms.MoveTopic = extractPerm("MoveTopic")

	err = forum.SetPerms(&forumPerms, "custom", gid)
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	return panelSuccessRedirect("/panel/forums/edit/perms/"+strconv.Itoa(fid)+"-"+strconv.Itoa(gid), w, r, isJs)
}

type AnalyticsTimeRange struct {
	Quantity   int
	Unit       string
	Slices     int
	SliceWidth int
	Range      string
}

func panelAnalyticsTimeRange(rawTimeRange string) (timeRange AnalyticsTimeRange, err error) {
	timeRange.Quantity = 6
	timeRange.Unit = "hour"
	timeRange.Slices = 12
	timeRange.SliceWidth = 60 * 30
	timeRange.Range = "six-hours"

	switch rawTimeRange {
	case "one-month":
		timeRange.Quantity = 30
		timeRange.Unit = "day"
		timeRange.Slices = 30
		timeRange.SliceWidth = 60 * 60 * 24
		timeRange.Range = "one-month"
	case "one-week":
		timeRange.Quantity = 7
		timeRange.Unit = "day"
		timeRange.Slices = 14
		timeRange.SliceWidth = 60 * 60 * 12
		timeRange.Range = "one-week"
	case "two-days": // Two days is experimental
		timeRange.Quantity = 2
		timeRange.Unit = "day"
		timeRange.Slices = 24
		timeRange.SliceWidth = 60 * 60 * 2
		timeRange.Range = "two-days"
	case "one-day":
		timeRange.Quantity = 1
		timeRange.Unit = "day"
		timeRange.Slices = 24
		timeRange.SliceWidth = 60 * 60
		timeRange.Range = "one-day"
	case "twelve-hours":
		timeRange.Quantity = 12
		timeRange.Slices = 24
		timeRange.Range = "twelve-hours"
	case "six-hours", "":
		timeRange.Range = "six-hours"
	default:
		return timeRange, errors.New("Unknown time range")
	}
	return timeRange, nil
}

func panelAnalyticsTimeRangeToLabelList(timeRange AnalyticsTimeRange) (revLabelList []int64, labelList []int64, viewMap map[int64]int64) {
	viewMap = make(map[int64]int64)
	var currentTime = time.Now().Unix()
	for i := 1; i <= timeRange.Slices; i++ {
		var label = currentTime - int64(i*timeRange.SliceWidth)
		revLabelList = append(revLabelList, label)
		viewMap[label] = 0
	}
	for _, value := range revLabelList {
		labelList = append(labelList, value)
	}
	return revLabelList, labelList, viewMap
}

func panelAnalyticsRowsToViewMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		err := rows.Scan(&count, &createdAt)
		if err != nil {
			return viewMap, err
		}

		var unixCreatedAt = createdAt.Unix()
		// TODO: Bulk log this
		if common.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
		}

		for _, value := range labelList {
			if unixCreatedAt > value {
				viewMap[value] += count
				break
			}
		}
	}
	return viewMap, rows.Err()
}

func routePanelAnalyticsViews(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsViews")
	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks").Columns("count, createdAt").Where("route = ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", graph, viewItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_views", w, r, user, &pi)
}

func routePanelAnalyticsRouteViews(w http.ResponseWriter, r *http.Request, user common.User, route string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsRouteViews")
	acc := qgen.Builder.Accumulator()
	// TODO: Validate the route is valid
	rows, err := acc.Select("viewchunks").Columns("count, createdAt").Where("route = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(route)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsRoutePage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", html.EscapeString(route), graph, viewItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_route_views", w, r, user, &pi)
}

func routePanelAnalyticsAgentViews(w http.ResponseWriter, r *http.Request, user common.User, agent string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsAgentViews")
	acc := qgen.Builder.Accumulator()
	// TODO: Verify the agent is valid
	rows, err := acc.Select("viewchunks_agents").Columns("count, createdAt").Where("browser = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(agent)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	// ? Only allow valid agents? The problem with this is that agents wind up getting renamed and it would take a migration to get them all up to snuff
	agent = html.EscapeString(agent)
	friendlyAgent, ok := common.GetUserAgentPhrase(agent)
	if !ok {
		friendlyAgent = agent
	}

	pi := common.PanelAnalyticsAgentPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", agent, friendlyAgent, graph, timeRange.Range}
	return panelRenderTemplate("panel_analytics_agent_views", w, r, user, &pi)
}

func routePanelAnalyticsForumViews(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsForumViews")
	acc := qgen.Builder.Accumulator()
	// TODO: Verify the agent is valid
	rows, err := acc.Select("viewchunks_forums").Columns("count, createdAt").Where("forum = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(sfid)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	forum, err := common.Forums.Get(fid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelAnalyticsAgentPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", sfid, forum.Name, graph, timeRange.Range}
	return panelRenderTemplate("panel_analytics_forum_views", w, r, user, &pi)
}

func routePanelAnalyticsSystemViews(w http.ResponseWriter, r *http.Request, user common.User, system string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsSystemViews")
	acc := qgen.Builder.Accumulator()
	// TODO: Verify the agent is valid
	rows, err := acc.Select("viewchunks_systems").Columns("count, createdAt").Where("system = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(system)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	system = html.EscapeString(system)
	friendlySystem, ok := common.GetOSPhrase(system)
	if !ok {
		friendlySystem = system
	}

	pi := common.PanelAnalyticsAgentPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", system, friendlySystem, graph, timeRange.Range}
	return panelRenderTemplate("panel_analytics_system_views", w, r, user, &pi)
}

func routePanelAnalyticsReferrerViews(w http.ResponseWriter, r *http.Request, user common.User, domain string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsReferrerViews")
	acc := qgen.Builder.Accumulator()
	// TODO: Verify the agent is valid
	rows, err := acc.Select("viewchunks_referrers").Columns("count, createdAt").Where("domain = ?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(domain)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsAgentPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", html.EscapeString(domain), "", graph, timeRange.Range}
	return panelRenderTemplate("panel_analytics_referrer_views", w, r, user, &pi)
}

func routePanelAnalyticsTopics(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsTopics")
	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("topicchunks").Columns("count, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", graph, viewItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_topics", w, r, user, &pi)
}

func routePanelAnalyticsPosts(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Stylesheets = append(headerVars.Stylesheets, "chartist/chartist.min.css")
	headerVars.Scripts = append(headerVars.Scripts, "chartist/chartist.min.js")
	headerVars.Scripts = append(headerVars.Scripts, "analytics.js")

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := panelAnalyticsTimeRangeToLabelList(timeRange)

	common.DebugLog("in routePanelAnalyticsPosts")
	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("postchunks").Columns("count, createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	viewMap, err = panelAnalyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []common.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, common.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := common.PanelTimeGraph{Series: viewList, Labels: labelList}
	common.DebugLogf("graph: %+v\n", graph)

	pi := common.PanelAnalyticsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", graph, viewItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_posts", w, r, user, &pi)
}

func panelAnalyticsRowsToNameMap(rows *sql.Rows) (map[string]int, error) {
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int
		var name string
		err := rows.Scan(&count, &name)
		if err != nil {
			return nameMap, err
		}

		// TODO: Bulk log this
		if common.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
		}
		nameMap[name] += count
	}
	return nameMap, rows.Err()
}

func routePanelAnalyticsForums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks_forums").Columns("count, forum").Where("forum != ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	forumMap, err := panelAnalyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var forumItems []common.PanelAnalyticsAgentsItem
	for sfid, count := range forumMap {
		fid, err := strconv.Atoi(sfid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		forum, err := common.Forums.Get(fid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		forumItems = append(forumItems, common.PanelAnalyticsAgentsItem{
			Agent:         sfid,
			FriendlyAgent: forum.Name,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsAgentsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", forumItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_forums", w, r, user, &pi)
}

func routePanelAnalyticsRoutes(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks").Columns("count, route").Where("route != ''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	routeMap, err := panelAnalyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var routeItems []common.PanelAnalyticsRoutesItem
	for route, count := range routeMap {
		routeItems = append(routeItems, common.PanelAnalyticsRoutesItem{
			Route: route,
			Count: count,
		})
	}

	pi := common.PanelAnalyticsRoutesPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", routeItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_routes", w, r, user, &pi)
}

func routePanelAnalyticsAgents(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks_agents").Columns("count, browser").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	agentMap, err := panelAnalyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var agentItems []common.PanelAnalyticsAgentsItem
	for agent, count := range agentMap {
		aAgent, ok := common.GetUserAgentPhrase(agent)
		if !ok {
			aAgent = agent
		}
		agentItems = append(agentItems, common.PanelAnalyticsAgentsItem{
			Agent:         agent,
			FriendlyAgent: aAgent,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsAgentsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", agentItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_agents", w, r, user, &pi)
}

func routePanelAnalyticsSystems(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks_systems").Columns("count, system").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	osMap, err := panelAnalyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var systemItems []common.PanelAnalyticsAgentsItem
	for system, count := range osMap {
		sSystem, ok := common.GetOSPhrase(system)
		if !ok {
			sSystem = system
		}
		systemItems = append(systemItems, common.PanelAnalyticsAgentsItem{
			Agent:         system,
			FriendlyAgent: sSystem,
			Count:         count,
		})
	}

	pi := common.PanelAnalyticsAgentsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", systemItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_systems", w, r, user, &pi)
}

func routePanelAnalyticsReferrers(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := panelAnalyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	acc := qgen.Builder.Accumulator()
	rows, err := acc.Select("viewchunks_referrers").Columns("count, domain").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	refMap, err := panelAnalyticsRowsToNameMap(rows)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Sort this slice
	var refItems []common.PanelAnalyticsAgentsItem
	for domain, count := range refMap {
		refItems = append(refItems, common.PanelAnalyticsAgentsItem{
			Agent: html.EscapeString(domain),
			Count: count,
		})
	}

	pi := common.PanelAnalyticsAgentsPage{common.GetTitlePhrase("panel_analytics"), user, headerVars, stats, "analytics", refItems, timeRange.Range}
	return panelRenderTemplate("panel_analytics_referrers", w, r, user, &pi)
}

func routePanelSettings(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	var settingList = make(map[string]interface{})

	settings, err := headerVars.Settings.BypassGetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// nolint need the type so people viewing this file understand what it returns without visiting phrases.go
	var settingLabels map[string]string = common.GetAllSettingLabels()
	for _, setting := range settings {
		if setting.Type == "list" {
			llist := settingLabels[setting.Name]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(setting.Content)
			if err != nil {
				return common.LocalError("The setting '"+setting.Name+"' can't be converted to an integer", w, r, user)
			}
			setting.Content = labels[conv-1]
		} else if setting.Type == "bool" {
			if setting.Content == "1" {
				setting.Content = "Yes"
			} else {
				setting.Content = "No"
			}
		}
		settingList[setting.Name] = setting.Content
	}

	pi := common.PanelPage{common.GetTitlePhrase("panel_settings"), user, headerVars, stats, "settings", tList, settingList}
	return panelRenderTemplate("panel_settings", w, r, user, &pi)
}

func routePanelSettingEdit(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	setting, err := headerVars.Settings.BypassGet(sname)
	if err == ErrNoRows {
		return common.LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var itemList []interface{}
	if setting.Type == "list" {
		llist := common.GetSettingLabel(setting.Name)
		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			return common.LocalError("The value of this setting couldn't be converted to an integer", w, r, user)
		}

		for index, label := range strings.Split(llist, ",") {
			itemList = append(itemList, common.OptionLabel{
				Label:    label,
				Value:    index + 1,
				Selected: conv == (index + 1),
			})
		}
	}

	pi := common.PanelPage{common.GetTitlePhrase("panel_edit_setting"), user, headerVars, stats, "settings", itemList, setting}
	return panelRenderTemplate("panel_setting", w, r, user, &pi)
}

func routePanelSettingEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
	headerLite, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	scontent := r.PostFormValue("setting-value")
	err := headerLite.Settings.Update(sname, scontent)
	if err != nil {
		if common.SafeSettingError(err) {
			return common.LocalError(err.Error(), w, r, user)
		}
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
	return nil
}

func routePanelWordFilters(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	var filterList = common.WordFilterBox.Load().(common.WordFilterMap)
	pi := common.PanelPage{common.GetTitlePhrase("panel_word_filters"), user, headerVars, stats, "word-filters", tList, filterList}
	return panelRenderTemplate("panel_word_filters", w, r, user, &pi)
}

func routePanelWordFiltersCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	res, err := stmts.createWordFilter.Exec(find, replacement)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	common.AddWordFilter(int(lastID), find, replacement)
	return panelSuccessRedirect("/panel/settings/word-filters/", w, r, isJs)
}

// TODO: Implement this as a non-JS fallback
func routePanelWordFiltersEdit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	_ = wfid

	pi := common.PanelPage{common.GetTitlePhrase("panel_edit_word_filter"), user, headerVars, stats, "word-filters", tList, nil}
	return panelRenderTemplate("panel_word_filters_edit", w, r, user, &pi)
}

func routePanelWordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return common.LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	_, err = stmts.updateWordFilter.Exec(find, replacement, id)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := common.WordFilterBox.Load().(common.WordFilterMap)
	wordFilters[id] = common.WordFilter{ID: id, Find: find, Replacement: replacement}
	common.WordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelWordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return common.LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	_, err = stmts.deleteWordFilter.Exec(id)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := common.WordFilterBox.Load().(common.WordFilterMap)
	delete(wordFilters, id)
	common.WordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelPlugins(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	var pluginList []interface{}
	for _, plugin := range common.Plugins {
		pluginList = append(pluginList, plugin)
	}

	pi := common.PanelPage{common.GetTitlePhrase("panel_plugins"), user, headerVars, stats, "plugins", pluginList, nil}
	return panelRenderTemplate("panel_plugins", w, r, user, &pi)
}

func routePanelPluginsActivate(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if plugin.Installable && !plugin.Installed {
		return common.LocalError("You can't activate this plugin without installing it first", w, r, user)
	}

	var active bool
	err := stmts.isPluginActive.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	var hasPlugin = (err == nil)

	if common.Plugins[uname].Activate != nil {
		err = common.Plugins[uname].Activate()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		if active {
			return common.LocalError("The plugin is already active", w, r, user)
		}
		_, err = stmts.updatePlugin.Exec(1, uname)
	} else {
		_, err = stmts.addPlugin.Exec(uname, 1, 0)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Activating plugin '%s'", plugin.Name)
	plugin.Active = true
	common.Plugins[uname] = plugin
	err = common.Plugins[uname].Init()
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelPluginsDeactivate(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}

	var active bool
	err := stmts.isPluginActive.QueryRow(uname).Scan(&active)
	if err == ErrNoRows {
		return common.LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if !active {
		return common.LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	}
	_, err = stmts.updatePlugin.Exec(0, uname)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	plugin.Active = false
	common.Plugins[uname] = plugin
	common.Plugins[uname].Deactivate()

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelPluginsInstall(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if !plugin.Installable {
		return common.LocalError("This plugin is not installable", w, r, user)
	}
	if plugin.Installed {
		return common.LocalError("This plugin has already been installed", w, r, user)
	}

	var active bool
	err := stmts.isPluginActive.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	var hasPlugin = (err == nil)

	if common.Plugins[uname].Install != nil {
		err = common.Plugins[uname].Install()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if common.Plugins[uname].Activate != nil {
		err = common.Plugins[uname].Activate()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		_, err = stmts.updatePluginInstall.Exec(1, uname)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		_, err = stmts.updatePlugin.Exec(1, uname)
	} else {
		_, err = stmts.addPlugin.Exec(uname, 1, 1)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Installing plugin '%s'", plugin.Name)
	plugin.Active = true
	plugin.Installed = true
	common.Plugins[uname] = plugin
	err = common.Plugins[uname].Init()
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelUsers(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(stats.Users, page, perPage)

	var userList []common.User
	// TODO: Move this into the common.UserStore
	rows, err := stmts.getUsersOffset.Query(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	// TODO: Add a common.UserStore method for iterating over global users and global user offsets
	for rows.Next() {
		puser := &common.User{ID: 0}
		err := rows.Scan(&puser.ID, &puser.Name, &puser.Group, &puser.Active, &puser.IsSuperAdmin, &puser.Avatar)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		puser.InitPerms()
		puser.Avatar = common.BuildAvatar(puser.ID, puser.Avatar)
		if common.Groups.DirtyGet(puser.Group).Tag != "" {
			puser.Tag = common.Groups.DirtyGet(puser.Group).Tag
		} else {
			puser.Tag = ""
		}
		userList = append(userList, *puser)
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(stats.Users, perPage, 5)
	pi := common.PanelUserPage{common.GetTitlePhrase("panel_users"), user, headerVars, stats, "users", userList, pageList, page, lastPage}
	return panelRenderTemplate("panel_users", w, r, user, &pi)
}

func routePanelUsersEdit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return common.LocalError("Only administrators can edit the account of an administrator.", w, r, user)
	}

	// ? - Should we stop admins from deleting all the groups? Maybe, protect the group they're currently using?
	groups, err := common.Groups.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var groupList []interface{}
	for _, group := range groups {
		if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
			continue
		}
		if !user.Perms.EditUserGroupSuperMod && group.IsMod {
			continue
		}
		groupList = append(groupList, group)
	}

	pi := common.PanelPage{common.GetTitlePhrase("panel_edit_user"), user, headerVars, stats, "users", groupList, targetUser}
	if common.RunPreRenderHook("pre_render_panel_edit_user", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-user-edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelUsersEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return common.LocalError("Only administrators can edit the account of other administrators.", w, r, user)
	}

	newname := html.EscapeString(strings.Replace(r.PostFormValue("user-name"), "\n", "", -1))
	if newname == "" {
		return common.LocalError("You didn't put in a username.", w, r, user)
	}

	// TODO: How should activation factor into admin set emails?
	// TODO: How should we handle secondary emails? Do we even have secondary emails implemented?
	newemail := html.EscapeString(strings.Replace(r.PostFormValue("user-email"), "\n", "", -1))
	if newemail == "" {
		return common.LocalError("You didn't put in an email address.", w, r, user)
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		return common.LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
	}

	newpassword := r.PostFormValue("user-password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		return common.LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("user-group"))
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(newgroup)
	if err == ErrNoRows {
		return common.LocalError("The group you're trying to place this user in doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
		return common.LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.", w, r, user)
	}
	if !user.Perms.EditUserGroupSuperMod && group.IsMod {
		return common.LocalError("You need the EditUserGroupSuperMod permission to assign someone to a super mod group.", w, r, user)
	}

	// TODO: Move this query into common
	_, err = stmts.updateUser.Exec(newname, newemail, newgroup, targetUser.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	if newpassword != "" {
		common.SetPassword(targetUser.ID, newpassword)
		// Log the user out as a safety precaution
		common.Auth.ForceLogout(targetUser.ID)
	}

	targetUser.CacheRemove()
	http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}

func routePanelGroups(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 9
	offset, page, lastPage := common.PageOffset(stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []common.GroupAdmin
	groups, _ := common.Groups.GetRange(offset, 0)
	for _, group := range groups {
		if count == perPage {
			break
		}

		var rank string
		var rankClass string
		var canEdit bool
		var canDelete = false

		// TODO: Use a switch for this
		if group.IsAdmin {
			rank = "Admin"
			rankClass = "admin"
		} else if group.IsMod {
			rank = "Mod"
			rankClass = "mod"
		} else if group.IsBanned {
			rank = "Banned"
			rankClass = "banned"
		} else if group.ID == 6 {
			rank = "Guest"
			rankClass = "guest"
		} else {
			rank = "Member"
			rankClass = "member"
		}

		canEdit = user.Perms.EditGroup && (!group.IsAdmin || user.Perms.EditGroupAdmin) && (!group.IsMod || user.Perms.EditGroupSuperMod)
		groupList = append(groupList, common.GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}
	//log.Printf("groupList: %+v\n", groupList)

	pageList := common.Paginate(stats.Groups, perPage, 5)
	pi := common.PanelGroupPage{common.GetTitlePhrase("panel_groups"), user, headerVars, stats, "groups", groupList, pageList, page, lastPage}
	return panelRenderTemplate("panel_groups", w, r, user, &pi)
}

func routePanelGroupsEdit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, headerVars)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	var rank string
	switch {
	case group.IsAdmin:
		rank = "Admin"
	case group.IsMod:
		rank = "Mod"
	case group.IsBanned:
		rank = "Banned"
	case group.ID == 6:
		rank = "Guest"
	default:
		rank = "Member"
	}

	disableRank := !user.Perms.EditGroupGlobalPerms || (group.ID == 6)

	pi := common.PanelEditGroupPage{common.GetTitlePhrase("panel_edit_group"), user, headerVars, stats, "groups", group.ID, group.Name, group.Tag, rank, disableRank}
	if common.RunPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-group-edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditPerms(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, headerVars)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []common.NameLangToggle

	var addLocalPerm = func(permStr string, perm bool) {
		localPerms = append(localPerms, common.NameLangToggle{permStr, common.GetLocalPermPhrase(permStr), perm})
	}

	addLocalPerm("ViewTopic", group.Perms.ViewTopic)
	addLocalPerm("LikeItem", group.Perms.LikeItem)
	addLocalPerm("CreateTopic", group.Perms.CreateTopic)
	//<--
	addLocalPerm("EditTopic", group.Perms.EditTopic)
	addLocalPerm("DeleteTopic", group.Perms.DeleteTopic)
	addLocalPerm("CreateReply", group.Perms.CreateReply)
	addLocalPerm("EditReply", group.Perms.EditReply)
	addLocalPerm("DeleteReply", group.Perms.DeleteReply)
	addLocalPerm("PinTopic", group.Perms.PinTopic)
	addLocalPerm("CloseTopic", group.Perms.CloseTopic)
	addLocalPerm("MoveTopic", group.Perms.MoveTopic)

	var globalPerms []common.NameLangToggle
	var addGlobalPerm = func(permStr string, perm bool) {
		globalPerms = append(globalPerms, common.NameLangToggle{permStr, common.GetGlobalPermPhrase(permStr), perm})
	}

	addGlobalPerm("BanUsers", group.Perms.BanUsers)
	addGlobalPerm("ActivateUsers", group.Perms.ActivateUsers)
	addGlobalPerm("EditUser", group.Perms.EditUser)
	addGlobalPerm("EditUserEmail", group.Perms.EditUserEmail)
	addGlobalPerm("EditUserPassword", group.Perms.EditUserPassword)
	addGlobalPerm("EditUserGroup", group.Perms.EditUserGroup)
	addGlobalPerm("EditUserGroupSuperMod", group.Perms.EditUserGroupSuperMod)
	addGlobalPerm("EditUserGroupAdmin", group.Perms.EditUserGroupAdmin)
	addGlobalPerm("EditGroup", group.Perms.EditGroup)
	addGlobalPerm("EditGroupLocalPerms", group.Perms.EditGroupLocalPerms)
	addGlobalPerm("EditGroupGlobalPerms", group.Perms.EditGroupGlobalPerms)
	addGlobalPerm("EditGroupSuperMod", group.Perms.EditGroupSuperMod)
	addGlobalPerm("EditGroupAdmin", group.Perms.EditGroupAdmin)
	addGlobalPerm("ManageForums", group.Perms.ManageForums)
	addGlobalPerm("EditSettings", group.Perms.EditSettings)
	addGlobalPerm("ManageThemes", group.Perms.ManageThemes)
	addGlobalPerm("ManagePlugins", group.Perms.ManagePlugins)
	addGlobalPerm("ViewAdminLogs", group.Perms.ViewAdminLogs)
	addGlobalPerm("ViewIPs", group.Perms.ViewIPs)
	addGlobalPerm("UploadFiles", group.Perms.UploadFiles)

	pi := common.PanelEditGroupPermsPage{common.GetTitlePhrase("panel_edit_group"), user, headerVars, stats, "groups", group.ID, group.Name, localPerms, globalPerms}
	if common.RunPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "panel-group-edit-perms.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return common.LocalError("The group name can't be left blank.", w, r, user)
	}
	gtag := r.FormValue("group-tag")
	rank := r.FormValue("group-type")

	var originalRank string
	// TODO: Use a switch for this
	if group.IsAdmin {
		originalRank = "Admin"
	} else if group.IsMod {
		originalRank = "Mod"
	} else if group.IsBanned {
		originalRank = "Banned"
	} else if group.ID == 6 {
		originalRank = "Guest"
	} else {
		originalRank = "Member"
	}

	if rank != originalRank {
		if !user.Perms.EditGroupGlobalPerms {
			return common.LocalError("You need the EditGroupGlobalPerms permission to change the group type.", w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.", w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.", w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return common.LocalError("You can't designate a group as a guest group.", w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return common.LocalError("Invalid group type.", w, r, user)
		}
		if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	// TODO: Move this to *Group
	_, err = stmts.updateGroup.Exec(gname, gtag, gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	common.Groups.Reload(gid)

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user common.User, sgid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return common.LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := common.Groups.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return common.LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return common.LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	var pmap = make(map[string]bool)
	if user.Perms.EditGroupLocalPerms {
		for _, perm := range common.LocalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	if user.Perms.EditGroupGlobalPerms {
		for _, perm := range common.GlobalPermList {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	// TODO: Abstract this
	pjson, err := json.Marshal(pmap)
	if err != nil {
		return common.LocalError("Unable to marshal the data", w, r, user)
	}
	_, err = stmts.updateGroupPerms.Exec(pjson, gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = common.RebuildGroupPermissions(gid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return common.NoPermissions(w, r, user)
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		return common.LocalError("You need a name for this group!", w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				return common.LocalError("You need the EditGroupAdmin permission to create admin groups", w, r, user)
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				return common.LocalError("You need the EditGroupSuperMod permission to create admin groups", w, r, user)
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := common.Groups.Create(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelThemes(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	var pThemeList, vThemeList []*common.Theme
	for _, theme := range common.Themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList, theme)
		} else {
			vThemeList = append(vThemeList, theme)
		}

	}

	pi := common.PanelThemesPage{common.GetTitlePhrase("panel_themes"), user, headerVars, stats, "themes", pThemeList, vThemeList}
	return panelRenderTemplate("panel_themes", w, r, user, &pi)
}

func routePanelThemesSetDefault(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	theme, ok := common.Themes[uname]
	if !ok {
		return common.LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return common.LocalError("You must not enable this theme", w, r, user)
	}

	var isDefault bool
	err := stmts.isThemeDefault.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	hasTheme := err != ErrNoRows
	if hasTheme {
		if isDefault {
			return common.LocalError("The theme is already active", w, r, user)
		}
		_, err = stmts.updateTheme.Exec(1, uname)
	} else {
		_, err = stmts.addTheme.Exec(uname, 1)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Make this less racey
	// TODO: Move this to common
	common.ChangeDefaultThemeMutex.Lock()
	defaultTheme := common.DefaultThemeBox.Load().(string)
	_, err = stmts.updateTheme.Exec(0, defaultTheme)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Setting theme '%s' as the default theme", theme.Name)
	theme.Active = true
	common.Themes[uname] = theme

	dTheme, ok := common.Themes[defaultTheme]
	if !ok {
		return common.InternalError(errors.New("The default theme is missing"), w, r)
	}
	dTheme.Active = false
	common.Themes[defaultTheme] = dTheme

	common.DefaultThemeBox.Store(uname)
	common.ResetTemplateOverrides()
	theme.MapTemplates()
	common.ChangeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func routePanelBackups(w http.ResponseWriter, r *http.Request, user common.User, backupURL string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	if backupURL != "" {
		// We don't want them trying to break out of this directory, it shouldn't hurt since it's a super admin, but it's always good to practice good security hygiene, especially if this is one of many instances on a managed server not controlled by the superadmin/s
		backupURL = common.Stripslashes(backupURL)

		var ext = filepath.Ext("./backups/" + backupURL)
		if ext == ".sql" {
			info, err := os.Stat("./backups/" + backupURL)
			if err != nil {
				return common.NotFound(w, r, headerVars)
			}
			// TODO: Change the served filename to gosora_backup_%timestamp%.sql, the time the file was generated, not when it was modified aka what the name of it should be
			w.Header().Set("Content-Disposition", "attachment; filename=gosora_backup.sql")
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
			// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
			http.ServeFile(w, r, "./backups/"+backupURL)
			return nil
		}
		return common.NotFound(w, r, headerVars)
	}

	var backupList []common.BackupItem
	backupFiles, err := ioutil.ReadDir("./backups")
	if err != nil {
		return common.InternalError(err, w, r)
	}
	for _, backupFile := range backupFiles {
		var ext = filepath.Ext(backupFile.Name())
		if ext != ".sql" {
			continue
		}
		backupList = append(backupList, common.BackupItem{backupFile.Name(), backupFile.ModTime()})
	}

	pi := common.PanelBackupPage{common.GetTitlePhrase("panel_backups"), user, headerVars, stats, "backups", backupList}
	return panelRenderTemplate("panel_backups", w, r, user, &pi)
}

// TODO: Log errors when something really screwy is going on?
func handleUnknownUser(user *common.User, err error) *common.User {
	if err != nil {
		return &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
	}
	return user
}
func handleUnknownTopic(topic *common.Topic, err error) *common.Topic {
	if err != nil {
		return &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
	}
	return topic
}

// TODO: Move the log building logic into /common/ and it's own abstraction
func topicElementTypeAction(action string, elementType string, elementID int, actor *common.User, topic *common.Topic) (out string) {
	if action == "delete" {
		return fmt.Sprintf("Topic #%d was deleted by <a href='%s'>%s</a>", elementID, actor.Link, actor.Name)
	}

	switch action {
	case "lock":
		out = "<a href='%s'>%s</a> was locked by <a href='%s'>%s</a>"
	case "unlock":
		out = "<a href='%s'>%s</a> was reopened by <a href='%s'>%s</a>"
	case "stick":
		out = "<a href='%s'>%s</a> was pinned by <a href='%s'>%s</a>"
	case "unstick":
		out = "<a href='%s'>%s</a> was unpinned by <a href='%s'>%s</a>"
	case "move":
		out = "<a href='%s'>%s</a> was moved by <a href='%s'>%s</a>" // TODO: Add where it was moved to, we'll have to change the source data for that, most likely? Investigate that and try to work this in
	default:
		return fmt.Sprintf("Unknown action '%s' on elementType '%s' by <a href='%s'>%s</a>", action, elementType, actor.Link, actor.Name)
	}
	return fmt.Sprintf(out, topic.Link, topic.Title, actor.Link, actor.Name)
}

func modlogsElementType(action string, elementType string, elementID int, actor *common.User) (out string) {
	switch elementType {
	case "topic":
		topic := handleUnknownTopic(common.Topics.Get(elementID))
		out = topicElementTypeAction(action, elementType, elementID, actor, topic)
	case "user":
		targetUser := handleUnknownUser(common.Users.Get(elementID))
		switch action {
		case "ban":
			out = "<a href='%s'>%s</a> was banned by <a href='%s'>%s</a>"
		case "unban":
			out = "<a href='%s'>%s</a> was unbanned by <a href='%s'>%s</a>"
		case "activate":
			out = "<a href='%s'>%s</a> was activated by <a href='%s'>%s</a>"
		}
		out = fmt.Sprintf(out, targetUser.Link, targetUser.Name, actor.Link, actor.Name)
	case "reply":
		if action == "delete" {
			topic := handleUnknownTopic(common.TopicByReplyID(elementID))
			out = fmt.Sprintf("A reply in <a href='%s'>%s</a> was deleted by <a href='%s'>%s</a>", topic.Link, topic.Title, actor.Link, actor.Name)
		}
	}

	if out == "" {
		out = fmt.Sprintf("Unknown action '%s' on elementType '%s' by <a href='%s'>%s</a>", action, elementType, actor.Link, actor.Name)
	}
	return out
}

func routePanelLogsMod(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	logCount := common.ModLogs.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	rows, err := stmts.getModlogsOffset.Query(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	var logs []common.LogItem
	var action, elementType, ipaddress, doneAt string
	var elementID, actorID int
	for rows.Next() {
		err := rows.Scan(&action, &elementID, &elementType, &ipaddress, &actorID, &doneAt)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		actor := handleUnknownUser(common.Users.Get(actorID))
		action = modlogsElementType(action, elementType, elementID, actor)
		logs = append(logs, common.LogItem{Action: template.HTML(action), IPAddress: ipaddress, DoneAt: doneAt})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelLogsPage{common.GetTitlePhrase("panel_mod_logs"), user, headerVars, stats, "logs", logs, pageList, page, lastPage}
	return panelRenderTemplate("panel_modlogs", w, r, user, &pi)
}

func routePanelLogsAdmin(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	logCount := common.ModLogs.GlobalCount()
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	rows, err := stmts.getAdminlogsOffset.Query(offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	var logs []common.LogItem
	var action, elementType, ipaddress, doneAt string
	var elementID, actorID int
	for rows.Next() {
		err := rows.Scan(&action, &elementID, &elementType, &ipaddress, &actorID, &doneAt)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		actor := handleUnknownUser(common.Users.Get(actorID))
		action = modlogsElementType(action, elementType, elementID, actor)
		logs = append(logs, common.LogItem{Action: template.HTML(action), IPAddress: ipaddress, DoneAt: doneAt})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelLogsPage{common.GetTitlePhrase("panel_admin_logs"), user, headerVars, stats, "logs", logs, pageList, page, lastPage}
	return panelRenderTemplate("panel_adminlogs", w, r, user, &pi)
}

func routePanelDebug(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	uptime := "..."
	dbStats := db.Stats()
	openConnCount := dbStats.OpenConnections
	// Disk I/O?

	pi := common.PanelDebugPage{common.GetTitlePhrase("panel_debug"), user, headerVars, stats, "debug", uptime, openConnCount, dbAdapter}
	return panelRenderTemplate("panel_debug", w, r, user, &pi)
}
