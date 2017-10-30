package main

import (
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

	"github.com/Azareal/gopsutil/mem"
)

func routePanel(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// We won't calculate this on the spot anymore, as the system doesn't seem to like it if we do multiple fetches simultaneously. Should we constantly calculate this on a background thread? Perhaps, the watchdog to scale back heavy features under load? One plus side is that we'd get immediate CPU percentages here instead of waiting it to kick in with WebSockets
	var cpustr = "Unknown"
	var cpuColour string

	var ramstr, ramColour string
	memres, err := mem.VirtualMemory()
	if err != nil {
		ramstr = "Unknown"
	} else {
		totalCount, totalUnit := convertByteUnit(float64(memres.Total))
		usedCount := convertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

		// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
		//log.Print("pre used_count",used_count)
		var totstr string
		if (totalCount - float64(int(totalCount))) > 0.85 {
			usedCount += 1.0 - (totalCount - float64(int(totalCount)))
			totstr = strconv.Itoa(int(totalCount) + 1)
		} else {
			totstr = fmt.Sprintf("%.1f", totalCount)
		}
		//log.Print("post used_count",used_count)

		if usedCount > totalCount {
			usedCount = totalCount
		}
		ramstr = fmt.Sprintf("%.1f", usedCount) + " / " + totstr + totalUnit

		ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
		//log.Print("ramperc",ramperc)
		if ramperc < 50 {
			ramColour = "stat_green"
		} else if ramperc < 75 {
			ramColour = "stat_orange"
		} else {
			ramColour = "stat_red"
		}
	}

	var postCount int
	err = todaysPostCountStmt.QueryRow().Scan(&postCount)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var postInterval = "day"

	var postColour string
	if postCount > 25 {
		postColour = "stat_green"
	} else if postCount > 5 {
		postColour = "stat_orange"
	} else {
		postColour = "stat_red"
	}

	var topicCount int
	err = todaysTopicCountStmt.QueryRow().Scan(&topicCount)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var topicInterval = "day"

	var topicColour string
	if topicCount > 8 {
		topicColour = "stat_green"
	} else if topicCount > 0 {
		topicColour = "stat_orange"
	} else {
		topicColour = "stat_red"
	}

	var reportCount int
	err = todaysReportCountStmt.QueryRow().Scan(&reportCount)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var reportInterval = "week"

	var newUserCount int
	err = todaysNewUserCountStmt.QueryRow().Scan(&newUserCount)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var newUserInterval = "week"

	var gridElements = []GridElement{
		GridElement{"dash-version", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)"},
		GridElement{"dash-cpu", "CPU: " + cpustr, 1, "grid_istat " + cpuColour, "", "", "The global CPU usage of this server"},
		GridElement{"dash-ram", "RAM: " + ramstr, 2, "grid_istat " + ramColour, "", "", "The global RAM usage of this server"},
	}

	if enableWebsockets {
		uonline := wsHub.userCount()
		gonline := wsHub.guestCount()
		totonline := uonline + gonline

		var onlineColour string
		if totonline > 10 {
			onlineColour = "stat_green"
		} else if totonline > 3 {
			onlineColour = "stat_orange"
		} else {
			onlineColour = "stat_red"
		}

		var onlineGuestsColour string
		if gonline > 10 {
			onlineGuestsColour = "stat_green"
		} else if gonline > 1 {
			onlineGuestsColour = "stat_orange"
		} else {
			onlineGuestsColour = "stat_red"
		}

		var onlineUsersColour string
		if uonline > 5 {
			onlineUsersColour = "stat_green"
		} else if uonline > 1 {
			onlineUsersColour = "stat_orange"
		} else {
			onlineUsersColour = "stat_red"
		}

		totonline, totunit := convertFriendlyUnit(totonline)
		uonline, uunit := convertFriendlyUnit(uonline)
		gonline, gunit := convertFriendlyUnit(gonline)

		gridElements = append(gridElements, GridElement{"dash-totonline", strconv.Itoa(totonline) + totunit + " online", 3, "grid_stat " + onlineColour, "", "", "The number of people who are currently online"})
		gridElements = append(gridElements, GridElement{"dash-gonline", strconv.Itoa(gonline) + gunit + " guests online", 4, "grid_stat " + onlineGuestsColour, "", "", "The number of guests who are currently online"})
		gridElements = append(gridElements, GridElement{"dash-uonline", strconv.Itoa(uonline) + uunit + " users online", 5, "grid_stat " + onlineUsersColour, "", "", "The number of logged-in users who are currently online"})
	}

	gridElements = append(gridElements, GridElement{"dash-postsperday", strconv.Itoa(postCount) + " posts / " + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours"})
	gridElements = append(gridElements, GridElement{"dash-topicsperday", strconv.Itoa(topicCount) + " topics / " + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours"})
	gridElements = append(gridElements, GridElement{"dash-totonlineperday", "20 online / day", 8, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The people online over the last 24 hours"*/})

	gridElements = append(gridElements, GridElement{"dash-searches", "8 searches / week", 9, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of searches over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-newusers", strconv.Itoa(newUserCount) + " new users / " + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days"})
	gridElements = append(gridElements, GridElement{"dash-reports", strconv.Itoa(reportCount) + " reports / " + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days"})

	gridElements = append(gridElements, GridElement{"dash-minperuser", "2 minutes / user / week", 12, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of number of minutes spent by each active user over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-visitorsperweek", "2 visitors / week", 13, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of unique visitors we've had over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-postsperuser", "5 posts / user / week", 14, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of posts made by each active user over the past week"*/})

	pi := PanelDashboardPage{"Control Panel Dashboard", user, headerVars, stats, gridElements}
	if preRenderHooks["pre_render_panel_dashboard"] != nil {
		if runPreRenderHook("pre_render_panel_dashboard", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-dashboard.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelForums(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}

	// TODO: Paginate this?
	var forumList []interface{}
	forums, err := fstore.GetAll()
	if err != nil {
		return InternalError(err, w, r)
	}

	// ? - Should we generate something similar to the forumView? It might be a little overkill for a page which is rarely loaded in comparison to /forums/
	for _, forum := range forums {
		if forum.Name != "" && forum.ParentID == 0 {
			fadmin := ForumAdmin{forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, forum.TopicCount, presetToLang(forum.Preset)}
			if fadmin.Preset == "" {
				fadmin.Preset = "custom"
			}
			forumList = append(forumList, fadmin)
		}
	}
	pi := PanelPage{"Forum Manager", user, headerVars, stats, forumList, nil}
	if preRenderHooks["pre_render_panel_forums"] != nil {
		if runPreRenderHook("pre_render_panel_forums", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-forums.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelForumsCreateSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := stripInvalidPreset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-name")
	active := (factive == "on" || factive == "1")

	_, err = fstore.Create(fname, fdesc, active, fpreset)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	return nil
}

// TODO: Revamp this
func routePanelForumsDelete(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := fstore.Get(fid)
	if err == ErrNoRows {
		return LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	confirmMsg := "Are you sure you want to delete the '" + forum.Name + "' forum?"
	yousure := AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := PanelPage{"Delete Forum", user, headerVars, stats, tList, yousure}
	if preRenderHooks["pre_render_panel_delete_forum"] != nil {
		if runPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "areyousure.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelForumsDeleteSubmit(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	err = fstore.Delete(fid)
	if err == ErrNoRows {
		return LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	return nil
}

func routePanelForumsEdit(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return LocalError("The provided Forum ID is not a valid number.", w, r, user)
	}

	forum, err := fstore.Get(fid)
	if err == ErrNoRows {
		return LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {

		return InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	glist, err := gstore.GetAll()
	if err != nil {
		return InternalError(err, w, r)
	}

	var gplist []GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		gplist = append(gplist, GroupForumPermPreset{group, forumPermsToGroupForumPreset(group.Forums[fid])})
	}

	pi := PanelEditForumPage{"Forum Editor", user, headerVars, stats, forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, gplist}
	if preRenderHooks["pre_render_panel_edit_forum"] != nil {
		if runPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-forum-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelForumsEditSubmit(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	forum, err := fstore.Get(fid)
	if err == ErrNoRows {
		return LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	forumName := r.PostFormValue("forum_name")
	forumDesc := r.PostFormValue("forum_desc")
	forumPreset := stripInvalidPreset(r.PostFormValue("forum_preset"))
	forumActive := r.PostFormValue("forum_active")

	var active = false
	if forumActive == "" {
		active = forum.Active
	} else if forumActive == "1" || forumActive == "Show" {
		active = true
	}

	err = forum.Update(forumName, forumDesc, active, forumPreset)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routePanelForumsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageForums {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		return LocalErrorJSQ("Invalid Group ID", w, r, user, isJs)
	}

	permPreset := stripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	fperms, changed := groupForumPresetToForumPerms(permPreset)

	forum, err := fstore.Get(fid)
	if err == ErrNoRows {
		return LocalErrorJSQ("This forum doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()
	if changed {
		permUpdateMutex.Lock()
		defer permUpdateMutex.Unlock()
		group, err := gstore.Get(gid)
		if err != nil {
			return LocalError("The group whose permissions you're updating doesn't exist.", w, r, user)
		}
		group.Forums[fid] = fperms

		err = replaceForumPermsForGroup(gid, map[int]string{fid: permPreset}, map[int]ForumPerms{fid: fperms})
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add this and replaceForumPermsForGroup into a transaction?
		_, err = updateForumStmt.Exec(forum.Name, forum.Desc, forum.Active, "", fid)
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}
		err = fstore.Reload(fid)
		if err != nil {
			// TODO: Log this? -- Another admin might have deleted it
			return LocalErrorJSQ("Unable to reload forum", w, r, user, isJs)
		}
	}

	if !isJs {
		http.Redirect(w, r, "/panel/forums/edit/"+strconv.Itoa(fid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routePanelSettings(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}

	//log.Print("headerVars.Settings",headerVars.Settings)
	var settingList = make(map[string]interface{})
	rows, err := getSettingsStmt.Query()
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	// nolint need the type so people viewing this file understand what it returns without visiting setting.go
	var settingLabels map[string]string = GetAllSettingLabels()
	var sname, scontent, stype string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype)
		if err != nil {
			return InternalError(err, w, r)
		}

		if stype == "list" {
			llist := settingLabels[sname]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(scontent)
			if err != nil {
				return LocalError("The setting '"+sname+"' can't be converted to an integer", w, r, user)
			}
			scontent = labels[conv-1]
		} else if stype == "bool" {
			if scontent == "1" {
				scontent = "Yes"
			} else {
				scontent = "No"
			}
		}
		settingList[sname] = scontent
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	pi := PanelPage{"Setting Manager", user, headerVars, stats, tList, settingList}
	if preRenderHooks["pre_render_panel_settings"] != nil {
		if runPreRenderHook("pre_render_panel_settings", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-settings.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelSetting(w http.ResponseWriter, r *http.Request, user User, sname string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}
	setting := Setting{sname, "", "", ""}

	err := getSettingStmt.QueryRow(setting.Name).Scan(&setting.Content, &setting.Type)
	if err == ErrNoRows {
		return LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	var itemList []interface{}
	if setting.Type == "list" {
		llist := GetSettingLabel(setting.Name)
		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			return LocalError("The value of this setting couldn't be converted to an integer", w, r, user)
		}

		labels := strings.Split(llist, ",")
		for index, label := range labels {
			itemList = append(itemList, OptionLabel{
				Label:    label,
				Value:    index + 1,
				Selected: conv == (index + 1),
			})
		}
	}

	pi := PanelPage{"Edit Setting", user, headerVars, stats, itemList, setting}
	if preRenderHooks["pre_render_panel_setting"] != nil {
		if runPreRenderHook("pre_render_panel_setting", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-setting.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelSettingEdit(w http.ResponseWriter, r *http.Request, user User, sname string) RouteError {
	headerLite, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	var stype, sconstraints string
	scontent := r.PostFormValue("setting-value")

	err = getFullSettingStmt.QueryRow(sname).Scan(&sname, &stype, &sconstraints)
	if err == ErrNoRows {
		return LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if stype == "bool" {
		if scontent == "on" || scontent == "1" {
			scontent = "1"
		} else {
			scontent = "0"
		}
	}

	// TODO: Make this a method or function?
	_, err = updateSettingStmt.Exec(scontent, sname)
	if err != nil {
		return InternalError(err, w, r)
	}

	errmsg := headerLite.Settings.ParseSetting(sname, scontent, stype, sconstraints)
	if errmsg != "" {
		return LocalError(errmsg, w, r, user)
	}
	settingBox.Store(headerLite.Settings)

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
	return nil
}

func routePanelWordFilters(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return nil
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}

	var filterList = wordFilterBox.Load().(WordFilterBox)
	pi := PanelPage{"Word Filter Manager", user, headerVars, stats, tList, filterList}
	if preRenderHooks["pre_render_panel_word_filters"] != nil {
		if runPreRenderHook("pre_render_panel_word_filters", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "panel-word-filters.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelWordFiltersCreate(w http.ResponseWriter, r *http.Request, user User) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("js") == "1")

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	res, err := createWordFilterStmt.Exec(find, replacement)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	addWordFilter(int(lastID), find, replacement)
	if !isJs {
		http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routePanelWordFiltersEdit(w http.ResponseWriter, r *http.Request, user User, wfid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return NoPermissions(w, r, user)
	}

	_ = wfid

	pi := PanelPage{"Edit Word Filter", user, headerVars, stats, tList, nil}
	if preRenderHooks["pre_render_panel_word_filters_edit"] != nil {
		if runPreRenderHook("pre_render_panel_word_filters_edit", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "panel-word-filters-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelWordFiltersEditSubmit(w http.ResponseWriter, r *http.Request, user User, wfid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		return LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	_, err = updateWordFilterStmt.Exec(find, replacement, id)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := wordFilterBox.Load().(WordFilterBox)
	wordFilters[id] = WordFilter{ID: id, Find: find, Replacement: replacement}
	wordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelWordFiltersDeleteSubmit(w http.ResponseWriter, r *http.Request, user User, wfid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		return LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
	}

	_, err = deleteWordFilterStmt.Exec(id)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	wordFilters := wordFilterBox.Load().(WordFilterBox)
	delete(wordFilters, id)
	wordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	return nil
}

func routePanelPlugins(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return NoPermissions(w, r, user)
	}

	var pluginList []interface{}
	for _, plugin := range plugins {
		//log.Print("plugin.Name ", plugin.Name)
		//log.Print("plugin.Installed ", plugin.Installed)
		pluginList = append(pluginList, plugin)
	}

	pi := PanelPage{"Plugin Manager", user, headerVars, stats, pluginList, nil}
	if preRenderHooks["pre_render_panel_plugins"] != nil {
		if runPreRenderHook("pre_render_panel_plugins", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "panel-plugins.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelPluginsActivate(w http.ResponseWriter, r *http.Request, user User, uname string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	//log.Print("uname","'"+uname+"'")
	plugin, ok := plugins[uname]
	if !ok {
		return LocalError("The plugin isn't registered in the system", w, r, user)
	}

	if plugin.Installable && !plugin.Installed {
		return LocalError("You can't activate this plugin without installing it first", w, r, user)
	}

	var active bool
	err := isPluginActiveStmt.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var hasPlugin = (err == nil)

	if plugins[uname].Activate != nil {
		err = plugins[uname].Activate()
		if err != nil {
			return LocalError(err.Error(), w, r, user)
		}
	}

	//log.Print("err", err)
	//log.Print("active", active)
	if hasPlugin {
		if active {
			return LocalError("The plugin is already active", w, r, user)
		}
		//log.Print("updatePlugin")
		_, err = updatePluginStmt.Exec(1, uname)
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		//log.Print("addPlugin")
		_, err := addPluginStmt.Exec(uname, 1, 0)
		if err != nil {
			return InternalError(err, w, r)
		}
	}

	log.Print("Activating plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugins[uname] = plugin
	err = plugins[uname].Init()
	if err != nil {
		return LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelPluginsDeactivate(w http.ResponseWriter, r *http.Request, user User, uname string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	plugin, ok := plugins[uname]
	if !ok {
		return LocalError("The plugin isn't registered in the system", w, r, user)
	}

	var active bool
	err := isPluginActiveStmt.QueryRow(uname).Scan(&active)
	if err == ErrNoRows {
		return LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if !active {
		return LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	}
	_, err = updatePluginStmt.Exec(0, uname)
	if err != nil {
		return InternalError(err, w, r)
	}

	plugin.Active = false
	plugins[uname] = plugin
	plugins[uname].Deactivate()

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelPluginsInstall(w http.ResponseWriter, r *http.Request, user User, uname string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	plugin, ok := plugins[uname]
	if !ok {
		return LocalError("The plugin isn't registered in the system", w, r, user)
	}

	if !plugin.Installable {
		return LocalError("This plugin is not installable", w, r, user)
	}

	if plugin.Installed {
		return LocalError("This plugin has already been installed", w, r, user)
	}

	var active bool
	err := isPluginActiveStmt.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	var hasPlugin = (err == nil)

	if plugins[uname].Install != nil {
		err = plugins[uname].Install()
		if err != nil {
			return LocalError(err.Error(), w, r, user)
		}
	}

	if plugins[uname].Activate != nil {
		err = plugins[uname].Activate()
		if err != nil {
			return LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		_, err = updatePluginInstallStmt.Exec(1, uname)
		if err != nil {
			return InternalError(err, w, r)
		}
		_, err = updatePluginStmt.Exec(1, uname)
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		_, err := addPluginStmt.Exec(uname, 1, 1)
		if err != nil {
			return InternalError(err, w, r)
		}
	}

	log.Print("Installing plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugin.Installed = true
	plugins[uname] = plugin
	err = plugins[uname].Init()
	if err != nil {
		return LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func routePanelUsers(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := pageOffset(stats.Users, page, perPage)

	var userList []User
	rows, err := getUsersOffsetStmt.Query(offset, perPage)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	// TODO: Add a UserStore method for iterating over global users and global user offsets
	for rows.Next() {
		puser := &User{ID: 0}
		err := rows.Scan(&puser.ID, &puser.Name, &puser.Group, &puser.Active, &puser.IsSuperAdmin, &puser.Avatar)
		if err != nil {
			return InternalError(err, w, r)
		}

		puser.initPerms()
		if puser.Avatar != "" {
			if puser.Avatar[0] == '.' {
				puser.Avatar = "/uploads/avatar_" + strconv.Itoa(puser.ID) + puser.Avatar
			}
		} else {
			puser.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(puser.ID), 1)
		}

		if gstore.DirtyGet(puser.Group).Tag != "" {
			puser.Tag = gstore.DirtyGet(puser.Group).Tag
		} else {
			puser.Tag = ""
		}
		userList = append(userList, *puser)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	pageList := paginate(stats.Users, perPage, 5)
	pi := PanelUserPage{"User Manager", user, headerVars, stats, userList, pageList, page, lastPage}
	if preRenderHooks["pre_render_panel_users"] != nil {
		if runPreRenderHook("pre_render_panel_users", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-users.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelUsersEdit(w http.ResponseWriter, r *http.Request, user User, suid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	if !user.Perms.EditUser {
		return NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		return LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return LocalError("Only administrators can edit the account of an administrator.", w, r, user)
	}

	// ? - Should we stop admins from deleting all the groups? Maybe, protect the group they're currently using?
	groups, err := gstore.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return InternalError(err, w, r)
	}

	var groupList []interface{}
	for _, group := range groups[1:] {
		if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
			continue
		}
		if !user.Perms.EditUserGroupSuperMod && group.IsMod {
			continue
		}
		groupList = append(groupList, group)
	}

	pi := PanelPage{"User Editor", user, headerVars, stats, groupList, targetUser}
	if preRenderHooks["pre_render_panel_edit_user"] != nil {
		if runPreRenderHook("pre_render_panel_edit_user", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-user-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelUsersEditSubmit(w http.ResponseWriter, r *http.Request, user User, suid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditUser {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		return LocalError("The user you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		return LocalError("Only administrators can edit the account of other administrators.", w, r, user)
	}

	newname := html.EscapeString(r.PostFormValue("user-name"))
	if newname == "" {
		return LocalError("You didn't put in a username.", w, r, user)
	}

	newemail := html.EscapeString(r.PostFormValue("user-email"))
	if newemail == "" {
		return LocalError("You didn't put in an email address.", w, r, user)
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		return LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
	}

	newpassword := r.PostFormValue("user-password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		return LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("user-group"))
	if err != nil {
		return LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := gstore.Get(newgroup)
	if err == ErrNoRows {
		return LocalError("The group you're trying to place this user in doesn't exist.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if !user.Perms.EditUserGroupAdmin && group.IsAdmin {
		return LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.", w, r, user)
	}
	if !user.Perms.EditUserGroupSuperMod && group.IsMod {
		return LocalError("You need the EditUserGroupSuperMod permission to assign someone to a super mod group.", w, r, user)
	}

	_, err = updateUserStmt.Exec(newname, newemail, newgroup, targetUser.ID)
	if err != nil {
		return InternalError(err, w, r)
	}

	if newpassword != "" {
		SetPassword(targetUser.ID, newpassword)
	}

	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(targetUser.ID)
	}
	http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}

func routePanelGroups(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 9
	offset, page, lastPage := pageOffset(stats.Groups, page, perPage)

	// Skip the 'Unknown' group
	offset++

	var count int
	var groupList []GroupAdmin
	groups, _ := gstore.GetRange(offset, 0)
	for _, group := range groups {
		if count == perPage {
			break
		}

		var rank string
		var rankClass string
		var canEdit bool
		var canDelete = false

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
		groupList = append(groupList, GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}
	//log.Printf("groupList: %+v\n", groupList)

	pageList := paginate(stats.Groups, perPage, 5)
	pi := PanelGroupPage{"Group Manager", user, headerVars, stats, groupList, pageList, page, lastPage}
	if preRenderHooks["pre_render_panel_groups"] != nil {
		if runPreRenderHook("pre_render_panel_groups", w, r, &user, &pi) {
			return nil
		}
	}

	err := templates.ExecuteTemplate(w, "panel-groups.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEdit(w http.ResponseWriter, r *http.Request, user User, sgid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
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

	pi := PanelEditGroupPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, group.Tag, rank, disableRank}
	if preRenderHooks["pre_render_panel_edit_group"] != nil {
		if runPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-group-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditPerms(w http.ResponseWriter, r *http.Request, user User, sgid string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return NoPermissions(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	// TODO: Load the phrases in bulk for efficiency?
	var localPerms []NameLangToggle
	localPerms = append(localPerms, NameLangToggle{"ViewTopic", GetLocalPermPhrase("ViewTopic"), group.Perms.ViewTopic})
	localPerms = append(localPerms, NameLangToggle{"LikeItem", GetLocalPermPhrase("LikeItem"), group.Perms.LikeItem})
	localPerms = append(localPerms, NameLangToggle{"CreateTopic", GetLocalPermPhrase("CreateTopic"), group.Perms.CreateTopic})
	//<--
	localPerms = append(localPerms, NameLangToggle{"EditTopic", GetLocalPermPhrase("EditTopic"), group.Perms.EditTopic})
	localPerms = append(localPerms, NameLangToggle{"DeleteTopic", GetLocalPermPhrase("DeleteTopic"), group.Perms.DeleteTopic})
	localPerms = append(localPerms, NameLangToggle{"CreateReply", GetLocalPermPhrase("CreateReply"), group.Perms.CreateReply})
	localPerms = append(localPerms, NameLangToggle{"EditReply", GetLocalPermPhrase("EditReply"), group.Perms.EditReply})
	localPerms = append(localPerms, NameLangToggle{"DeleteReply", GetLocalPermPhrase("DeleteReply"), group.Perms.DeleteReply})
	localPerms = append(localPerms, NameLangToggle{"PinTopic", GetLocalPermPhrase("PinTopic"), group.Perms.PinTopic})
	localPerms = append(localPerms, NameLangToggle{"CloseTopic", GetLocalPermPhrase("CloseTopic"), group.Perms.CloseTopic})

	var globalPerms []NameLangToggle
	globalPerms = append(globalPerms, NameLangToggle{"BanUsers", GetGlobalPermPhrase("BanUsers"), group.Perms.BanUsers})
	globalPerms = append(globalPerms, NameLangToggle{"ActivateUsers", GetGlobalPermPhrase("ActivateUsers"), group.Perms.ActivateUsers})
	globalPerms = append(globalPerms, NameLangToggle{"EditUser", GetGlobalPermPhrase("EditUser"), group.Perms.EditUser})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserEmail", GetGlobalPermPhrase("EditUserEmail"), group.Perms.EditUserEmail})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserPassword", GetGlobalPermPhrase("EditUserPassword"), group.Perms.EditUserPassword})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroup", GetGlobalPermPhrase("EditUserGroup"), group.Perms.EditUserGroup})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroupSuperMod", GetGlobalPermPhrase("EditUserGroupSuperMod"), group.Perms.EditUserGroupSuperMod})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroupAdmin", GetGlobalPermPhrase("EditUserGroupAdmin"), group.Perms.EditUserGroupAdmin})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroup", GetGlobalPermPhrase("EditGroup"), group.Perms.EditGroup})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupLocalPerms", GetGlobalPermPhrase("EditGroupLocalPerms"), group.Perms.EditGroupLocalPerms})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupGlobalPerms", GetGlobalPermPhrase("EditGroupGlobalPerms"), group.Perms.EditGroupGlobalPerms})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupSuperMod", GetGlobalPermPhrase("EditGroupSuperMod"), group.Perms.EditGroupSuperMod})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupAdmin", GetGlobalPermPhrase("EditGroupAdmin"), group.Perms.EditGroupAdmin})
	globalPerms = append(globalPerms, NameLangToggle{"ManageForums", GetGlobalPermPhrase("ManageForums"), group.Perms.ManageForums})
	globalPerms = append(globalPerms, NameLangToggle{"EditSettings", GetGlobalPermPhrase("EditSettings"), group.Perms.EditSettings})
	globalPerms = append(globalPerms, NameLangToggle{"ManageThemes", GetGlobalPermPhrase("ManageThemes"), group.Perms.ManageThemes})
	globalPerms = append(globalPerms, NameLangToggle{"ManagePlugins", GetGlobalPermPhrase("ManagePlugins"), group.Perms.ManagePlugins})
	globalPerms = append(globalPerms, NameLangToggle{"ViewAdminLogs", GetGlobalPermPhrase("ViewAdminLogs"), group.Perms.ViewAdminLogs})
	globalPerms = append(globalPerms, NameLangToggle{"ViewIPs", GetGlobalPermPhrase("ViewIPs"), group.Perms.ViewIPs})
	globalPerms = append(globalPerms, NameLangToggle{"UploadFiles", GetGlobalPermPhrase("UploadFiles"), group.Perms.UploadFiles})

	pi := PanelEditGroupPermsPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, localPerms, globalPerms}
	if preRenderHooks["pre_render_panel_edit_group_perms"] != nil {
		if runPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-group-edit-perms.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelGroupsEditSubmit(w http.ResponseWriter, r *http.Request, user User, sgid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return LocalError("You need to provide a whole number for the group ID", w, r, user)
	}

	group, err := gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		return LocalError("The group name can't be left blank.", w, r, user)
	}
	gtag := r.FormValue("group-tag")
	rank := r.FormValue("group-type")

	var originalRank string
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

	groupUpdateMutex.Lock()
	defer groupUpdateMutex.Unlock()
	if rank != originalRank {
		if !user.Perms.EditGroupGlobalPerms {
			return LocalError("You need the EditGroupGlobalPerms permission to change the group type.", w, r, user)
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				return LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.", w, r, user)
			}
			err = group.ChangeRank(true, true, false)
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				return LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.", w, r, user)
			}
			err = group.ChangeRank(false, true, false)
		case "Banned":
			err = group.ChangeRank(false, false, true)
		case "Guest":
			return LocalError("You can't designate a group as a guest group.", w, r, user)
		case "Member":
			err = group.ChangeRank(false, false, false)
		default:
			return LocalError("Invalid group type.", w, r, user)
		}
		if err != nil {
			return InternalError(err, w, r)
		}
	}

	_, err = updateGroupStmt.Exec(gname, gtag, gid)
	if err != nil {
		return InternalError(err, w, r)
	}
	group.Name = gname
	group.Tag = gtag

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsEditPermsSubmit(w http.ResponseWriter, r *http.Request, user User, sgid string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		return LocalError("The Group ID is not a valid integer.", w, r, user)
	}

	group, err := gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		return LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		return LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
	}

	////var lpmap map[string]bool = make(map[string]bool)
	var pmap = make(map[string]bool)
	if user.Perms.EditGroupLocalPerms {
		pplist := LocalPermList
		for _, perm := range pplist {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	////var gpmap map[string]bool = make(map[string]bool)
	if user.Perms.EditGroupGlobalPerms {
		gplist := GlobalPermList
		for _, perm := range gplist {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	pjson, err := json.Marshal(pmap)
	if err != nil {
		return LocalError("Unable to marshal the data", w, r, user)
	}
	_, err = updateGroupPermsStmt.Exec(pjson, gid)
	if err != nil {
		return InternalError(err, w, r)
	}
	err = rebuildGroupPermissions(gid)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelGroupsCreateSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditGroup {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		return LocalError("You need a name for this group!", w, r, user)
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				return LocalError("You need the EditGroupAdmin permission to create admin groups", w, r, user)
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				return LocalError("You need the EditGroupSuperMod permission to create admin groups", w, r, user)
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := gstore.Create(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		return InternalError(err, w, r)
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
	return nil
}

func routePanelThemes(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return NoPermissions(w, r, user)
	}

	var pThemeList, vThemeList []Theme
	for _, theme := range themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList, theme)
		} else {
			vThemeList = append(vThemeList, theme)
		}

	}

	pi := PanelThemesPage{"Theme Manager", user, headerVars, stats, pThemeList, vThemeList}
	if preRenderHooks["pre_render_panel_themes"] != nil {
		if runPreRenderHook("pre_render_panel_themes", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "panel-themes.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelThemesSetDefault(w http.ResponseWriter, r *http.Request, user User, uname string) RouteError {
	_, ferr := SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	theme, ok := themes[uname]
	if !ok {
		return LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return LocalError("You must not enable this theme", w, r, user)
	}

	var isDefault bool
	log.Print("uname", uname) // TODO: Do we need to log this?
	err := isThemeDefaultStmt.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	hasTheme := err != ErrNoRows
	if hasTheme {
		log.Print("isDefault", isDefault) // TODO: Do we need to log this?
		if isDefault {
			return LocalError("The theme is already active", w, r, user)
		}
		_, err = updateThemeStmt.Exec(1, uname)
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		_, err := addThemeStmt.Exec(uname, 1)
		if err != nil {
			return InternalError(err, w, r)
		}
	}

	// TODO: Make this less racey
	changeDefaultThemeMutex.Lock()
	defaultTheme := defaultThemeBox.Load().(string)
	_, err = updateThemeStmt.Exec(0, defaultTheme)
	if err != nil {
		return InternalError(err, w, r)
	}

	log.Print("Setting theme '" + theme.Name + "' as the default theme")
	theme.Active = true
	themes[uname] = theme

	dTheme, ok := themes[defaultTheme]
	if !ok {
		return InternalError(errors.New("The default theme is missing"), w, r)
	}
	dTheme.Active = false
	themes[defaultTheme] = dTheme

	defaultThemeBox.Store(uname)
	resetTemplateOverrides()
	mapThemeTemplates(theme)
	changeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func routePanelBackups(w http.ResponseWriter, r *http.Request, user User, backupURL string) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.IsSuperAdmin {
		return NoPermissions(w, r, user)
	}

	if backupURL != "" {
		// We don't want them trying to break out of this directory, it shouldn't hurt since it's a super admin, but it's always good to practice good security hygiene, especially if this is one of many instances on a managed server not controlled by the superadmin/s
		backupURL = Stripslashes(backupURL)

		var ext = filepath.Ext("./backups/" + backupURL)
		if ext == ".sql" {
			info, err := os.Stat("./backups/" + backupURL)
			if err != nil {
				return NotFound(w, r)
			}
			// TODO: Change the served filename to gosora_backup_%timestamp%.sql, the time the file was generated, not when it was modified aka what the name of it should be
			w.Header().Set("Content-Disposition", "attachment; filename=gosora_backup.sql")
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
			// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
			http.ServeFile(w, r, "./backups/"+backupURL)
			return nil
		}
		return NotFound(w, r)
	}

	var backupList []backupItem
	backupFiles, err := ioutil.ReadDir("./backups")
	if err != nil {
		return InternalError(err, w, r)
	}
	for _, backupFile := range backupFiles {
		var ext = filepath.Ext(backupFile.Name())
		if ext != ".sql" {
			continue
		}
		backupList = append(backupList, backupItem{backupFile.Name(), backupFile.ModTime()})
	}

	pi := PanelBackupPage{"Backups", user, headerVars, stats, backupList}
	err = templates.ExecuteTemplate(w, "panel-backups.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelLogsMod(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var logCount int
	err := modlogCountStmt.QueryRow().Scan(&logCount)
	if err != nil {
		return InternalError(err, w, r)
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := pageOffset(logCount, page, perPage)

	rows, err := getModlogsOffsetStmt.Query(offset, perPage)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	var logs []logItem
	var action, elementType, ipaddress, doneAt string
	var elementID, actorID int
	for rows.Next() {
		err := rows.Scan(&action, &elementID, &elementType, &ipaddress, &actorID, &doneAt)
		if err != nil {
			return InternalError(err, w, r)
		}

		actor, err := users.Get(actorID)
		if err != nil {
			actor = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
		}

		switch action {
		case "lock":
			topic, err := topics.Get(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was locked by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unlock":
			topic, err := topics.Get(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was reopened by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "stick":
			topic, err := topics.Get(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was pinned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unstick":
			topic, err := topics.Get(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was unpinned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "delete":
			if elementType == "topic" {
				action = "Topic #" + strconv.Itoa(elementID) + " was deleted by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
			} else {
				topic, err := getTopicByReply(elementID)
				if err != nil {
					topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
				}
				action = "A reply in <a href='" + topic.Link + "'>" + topic.Title + "</a> was deleted by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
			}
		case "ban":
			targetUser, err := users.Get(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was banned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unban":
			targetUser, err := users.Get(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was unbanned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "activate":
			targetUser, err := users.Get(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was activated by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		default:
			action = "Unknown action '" + action + "' by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		}
		logs = append(logs, logItem{Action: template.HTML(action), IPAddress: ipaddress, DoneAt: doneAt})
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	pageList := paginate(logCount, perPage, 5)
	pi := PanelLogsPage{"Moderation Logs", user, headerVars, stats, logs, pageList, page, lastPage}
	if preRenderHooks["pre_render_panel_mod_log"] != nil {
		if runPreRenderHook("pre_render_panel_mod_log", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "panel-modlogs.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routePanelDebug(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, stats, ferr := PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.IsAdmin {
		return NoPermissions(w, r, user)
	}

	uptime := "..."
	dbStats := db.Stats()
	openConnCount := dbStats.OpenConnections
	// Disk I/O?

	pi := PanelDebugPage{"Debug", user, headerVars, stats, uptime, openConnCount, dbAdapter}
	err := templates.ExecuteTemplate(w, "panel-debug.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}
