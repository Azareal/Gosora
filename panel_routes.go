package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/gopsutil/mem"
)

func route_panel(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
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
	err = todays_post_count_stmt.QueryRow().Scan(&postCount)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
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
	err = todays_topic_count_stmt.QueryRow().Scan(&topicCount)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
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
	err = todays_report_count_stmt.QueryRow().Scan(&reportCount)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}
	var reportInterval = "week"

	var newUserCount int
	err = todays_newuser_count_stmt.QueryRow().Scan(&newUserCount)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
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
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-dashboard.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_forums(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}

	var forumList []interface{}
	forums, err := fstore.GetAll()
	if err != nil {
		InternalError(err, w)
		return
	}

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
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-forums.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_forums_create_submit(w http.ResponseWriter, r *http.Request, user User) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := stripInvalidPreset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-name")
	active := (factive == "on" || factive == "1")

	_, err = fstore.CreateForum(fname, fdesc, active, fpreset)
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
}

// TODO: Revamp this
func route_panel_forums_delete(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.", w, r, user)
		return
	}

	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	confirmMsg := "Are you sure you want to delete the '" + forum.Name + "' forum?"
	yousure := AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := PanelPage{"Delete Forum", user, headerVars, stats, tList, yousure}
	if preRenderHooks["pre_render_panel_delete_forum"] != nil {
		if runPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "areyousure.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_forums_delete_submit(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.", w, r, user)
		return
	}

	err = fstore.CascadeDelete(fid)
	if err == ErrNoRows {
		LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
}

func route_panel_forums_edit(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.", w, r, user)
		return
	}

	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	var glist = groups
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
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-forum-edit.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_forums_edit_submit(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
		return
	}

	forumName := r.PostFormValue("forum_name")
	forumDesc := r.PostFormValue("forum_desc")
	forumPreset := stripInvalidPreset(r.PostFormValue("forum_preset"))
	forumActive := r.PostFormValue("forum_active")

	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		LocalErrorJSQ("The forum you're trying to edit doesn't exist.", w, r, user, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	if forumName == "" {
		forumName = forum.Name
	}

	var active bool
	if forumActive == "" {
		active = forum.Active
	} else if forumActive == "1" || forumActive == "Show" {
		active = true
	} else {
		active = false
	}

	forumUpdateMutex.Lock()
	_, err = update_forum_stmt.Exec(forumName, forumDesc, active, forumPreset, fid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}
	if forum.Name != forumName {
		forum.Name = forumName
	}
	if forum.Desc != forumDesc {
		forum.Desc = forumDesc
	}
	if forum.Active != active {
		forum.Active = active
	}
	if forum.Preset != forumPreset {
		forum.Preset = forumPreset
	}
	forumUpdateMutex.Unlock()

	permmapToQuery(presetToPermmap(forumPreset), fid)

	if !isJs {
		http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
}

func route_panel_forums_edit_perms_submit(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalErrorJSQ("The provided Forum ID is not a valid number.", w, r, user, isJs)
		return
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		LocalErrorJSQ("Invalid Group ID", w, r, user, isJs)
		return
	}

	permPreset := stripInvalidGroupForumPreset(r.PostFormValue("perm_preset"))
	fperms, changed := groupForumPresetToForumPerms(permPreset)

	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		LocalErrorJSQ("This forum doesn't exist", w, r, user, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()
	if changed {
		permUpdateMutex.Lock()
		groups[gid].Forums[fid] = fperms

		perms, err := json.Marshal(fperms)
		if err != nil {
			InternalErrorJSQ(err, w, r, isJs)
			return
		}

		_, err = add_forum_perms_to_group_stmt.Exec(gid, fid, permPreset, perms)
		if err != nil {
			InternalErrorJSQ(err, w, r, isJs)
			return
		}
		permUpdateMutex.Unlock()

		_, err = update_forum_stmt.Exec(forum.Name, forum.Desc, forum.Active, "", fid)
		if err != nil {
			InternalErrorJSQ(err, w, r, isJs)
			return
		}
		forum.Preset = ""
	}

	if !isJs {
		http.Redirect(w, r, "/panel/forums/edit/"+strconv.Itoa(fid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
}

func route_panel_settings(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}

	//log.Print("headerVars.Settings",headerVars.Settings)
	var settingList = make(map[string]interface{})
	rows, err := get_settings_stmt.Query()
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	// nolint need the type so people viewing this file understand what it returns without visiting setting.go
	var settingLabels map[string]string = GetAllSettingLabels()
	var sname, scontent, stype string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype)
		if err != nil {
			InternalError(err, w)
			return
		}

		if stype == "list" {
			llist := settingLabels[sname]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(scontent)
			if err != nil {
				LocalError("The setting '"+sname+"' can't be converted to an integer", w, r, user)
				return
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
		InternalError(err, w)
		return
	}

	pi := PanelPage{"Setting Manager", user, headerVars, stats, tList, settingList}
	if preRenderHooks["pre_render_panel_settings"] != nil {
		if runPreRenderHook("pre_render_panel_settings", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-settings.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_setting(w http.ResponseWriter, r *http.Request, user User, sname string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}
	setting := Setting{sname, "", "", ""}

	err := get_setting_stmt.QueryRow(setting.Name).Scan(&setting.Content, &setting.Type)
	if err == ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	var itemList []interface{}
	if setting.Type == "list" {
		llist := GetSettingLabel(setting.Name)
		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			LocalError("The value of this setting couldn't be converted to an integer", w, r, user)
			return
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
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-setting.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_setting_edit(w http.ResponseWriter, r *http.Request, user User, sname string) {
	headerLite, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	var stype, sconstraints string
	scontent := r.PostFormValue("setting-value")

	err = get_full_setting_stmt.QueryRow(sname).Scan(&sname, &stype, &sconstraints)
	if err == ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if stype == "bool" {
		if scontent == "on" || scontent == "1" {
			scontent = "1"
		} else {
			scontent = "0"
		}
	}

	_, err = update_setting_stmt.Exec(scontent, sname)
	if err != nil {
		InternalError(err, w)
		return
	}

	errmsg := headerLite.Settings.ParseSetting(sname, scontent, stype, sconstraints)
	if errmsg != "" {
		LocalError(errmsg, w, r, user)
		return
	}
	settingBox.Store(headerLite.Settings)

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
}

func route_panel_word_filters(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}

	var filterList = wordFilterBox.Load().(WordFilterBox)
	pi := PanelPage{"Word Filter Manager", user, headerVars, stats, tList, filterList}
	if preRenderHooks["pre_render_panel_word_filters"] != nil {
		if runPreRenderHook("pre_render_panel_word_filters", w, r, &user, &pi) {
			return
		}
	}
	err := templates.ExecuteTemplate(w, "panel-word-filters.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_word_filters_create(w http.ResponseWriter, r *http.Request, user User) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
		return
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	res, err := create_word_filter_stmt.Exec(find, replacement)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	addWordFilter(int(lastID), find, replacement)
	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther) // TODO: Return json for JS?
}

func route_panel_word_filters_edit(w http.ResponseWriter, r *http.Request, user User, wfid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w, r, user)
		return
	}

	_ = wfid

	pi := PanelPage{"Edit Word Filter", user, headerVars, stats, tList, nil}
	if preRenderHooks["pre_render_panel_word_filters_edit"] != nil {
		if runPreRenderHook("pre_render_panel_word_filters_edit", w, r, &user, &pi) {
			return
		}
	}
	err := templates.ExecuteTemplate(w, "panel-word-filters-edit.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_word_filters_edit_submit(w http.ResponseWriter, r *http.Request, user User, wfid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	// TODO: Either call it isJs or js rather than flip-flopping back and forth across the routes x.x
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
		return
	}

	find := strings.TrimSpace(r.PostFormValue("find"))
	if find == "" {
		LocalErrorJSQ("You need to specify what word you want to match", w, r, user, isJs)
		return
	}

	// Unlike with find, it's okay if we leave this blank, as this means that the admin wants to remove the word entirely with no replacement
	replacement := strings.TrimSpace(r.PostFormValue("replacement"))

	_, err = update_word_filter_stmt.Exec(find, replacement, id)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	wordFilters := wordFilterBox.Load().(WordFilterBox)
	wordFilters[id] = WordFilter{ID: id, Find: find, Replacement: replacement}
	wordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
}

func route_panel_word_filters_delete_submit(w http.ResponseWriter, r *http.Request, user User, wfid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	isJs := (r.PostFormValue("isJs") == "1")
	if !user.Perms.EditSettings {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	id, err := strconv.Atoi(wfid)
	if err != nil {
		LocalErrorJSQ("The word filter ID must be an integer.", w, r, user, isJs)
		return
	}

	_, err = delete_word_filter_stmt.Exec(id)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	wordFilters := wordFilterBox.Load().(WordFilterBox)
	delete(wordFilters, id)
	wordFilterBox.Store(wordFilters)

	http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
}

func route_panel_plugins(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w, r, user)
		return
	}

	var pluginList []interface{}
	for _, plugin := range plugins {
		//log.Print("plugin.Name",plugin.Name)
		//log.Print("plugin.Installed",plugin.Installed)
		pluginList = append(pluginList, plugin)
	}

	pi := PanelPage{"Plugin Manager", user, headerVars, stats, pluginList, nil}
	if preRenderHooks["pre_render_panel_plugins"] != nil {
		if runPreRenderHook("pre_render_panel_plugins", w, r, &user, &pi) {
			return
		}
	}
	err := templates.ExecuteTemplate(w, "panel-plugins.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_plugins_activate(w http.ResponseWriter, r *http.Request, user User, uname string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	//log.Print("uname","'"+uname+"'")
	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system", w, r, user)
		return
	}

	if plugin.Installable && !plugin.Installed {
		LocalError("You can't activate this plugin without installing it first", w, r, user)
		return
	}

	var active bool
	err := is_plugin_active_stmt.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}
	var hasPlugin = (err == nil)

	if plugins[uname].Activate != nil {
		err = plugins[uname].Activate()
		if err != nil {
			LocalError(err.Error(), w, r, user)
			return
		}
	}

	//log.Print("err",err)
	//log.Print("active",active)
	if hasPlugin {
		if active {
			LocalError("The plugin is already active", w, r, user)
			return
		}
		//log.Print("update_plugin")
		_, err = update_plugin_stmt.Exec(1, uname)
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		//log.Print("add_plugin")
		_, err := add_plugin_stmt.Exec(uname, 1, 0)
		if err != nil {
			InternalError(err, w)
			return
		}
	}

	log.Print("Activating plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugins[uname] = plugin
	err = plugins[uname].Init()
	if err != nil {
		LocalError(err.Error(), w, r, user)
		return
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
}

func route_panel_plugins_deactivate(w http.ResponseWriter, r *http.Request, user User, uname string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w, r, user)
		return
	}

	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system", w, r, user)
		return
	}

	var active bool
	err := is_plugin_active_stmt.QueryRow(uname).Scan(&active)
	if err == ErrNoRows {
		LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if !active {
		LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
		return
	}
	_, err = update_plugin_stmt.Exec(0, uname)
	if err != nil {
		InternalError(err, w)
		return
	}

	plugin.Active = false
	plugins[uname] = plugin
	plugins[uname].Deactivate()

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
}

func route_panel_plugins_install(w http.ResponseWriter, r *http.Request, user User, uname string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system", w, r, user)
		return
	}

	if !plugin.Installable {
		LocalError("This plugin is not installable", w, r, user)
		return
	}

	if plugin.Installed {
		LocalError("This plugin has already been installed", w, r, user)
		return
	}

	var active bool
	err := is_plugin_active_stmt.QueryRow(uname).Scan(&active)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}
	var hasPlugin = (err == nil)

	if plugins[uname].Install != nil {
		err = plugins[uname].Install()
		if err != nil {
			LocalError(err.Error(), w, r, user)
			return
		}
	}

	if plugins[uname].Activate != nil {
		err = plugins[uname].Activate()
		if err != nil {
			LocalError(err.Error(), w, r, user)
			return
		}
	}

	if hasPlugin {
		_, err = update_plugin_install_stmt.Exec(1, uname)
		if err != nil {
			InternalError(err, w)
			return
		}
		_, err = update_plugin_stmt.Exec(1, uname)
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		_, err := add_plugin_stmt.Exec(uname, 1, 1)
		if err != nil {
			InternalError(err, w)
			return
		}
	}

	log.Print("Installing plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugin.Installed = true
	plugins[uname] = plugin
	err = plugins[uname].Init()
	if err != nil {
		LocalError(err.Error(), w, r, user)
		return
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
}

func route_panel_users(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := pageOffset(stats.Users, page, perPage)

	var userList []User
	rows, err := get_users_offset_stmt.Query(offset, perPage)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	// TODO: Add a UserStore method for iterating over global users and global user offsets
	for rows.Next() {
		puser := User{ID: 0}
		err := rows.Scan(&puser.ID, &puser.Name, &puser.Group, &puser.Active, &puser.IsSuperAdmin, &puser.Avatar)
		if err != nil {
			InternalError(err, w)
			return
		}

		initUserPerms(&puser)
		if puser.Avatar != "" {
			if puser.Avatar[0] == '.' {
				puser.Avatar = "/uploads/avatar_" + strconv.Itoa(puser.ID) + puser.Avatar
			}
		} else {
			puser.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(puser.ID), 1)
		}

		if groups[puser.Group].Tag != "" {
			puser.Tag = groups[puser.Group].Tag
		} else {
			puser.Tag = ""
		}
		userList = append(userList, puser)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	pageList := paginate(stats.Users, perPage, 5)
	pi := PanelUserPage{"User Manager", user, headerVars, stats, userList, pageList, page, lastPage}
	if preRenderHooks["pre_render_panel_users"] != nil {
		if runPreRenderHook("pre_render_panel_users", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-users.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_users_edit(w http.ResponseWriter, r *http.Request, user User, suid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	if !user.Perms.EditUser {
		NoPermissions(w, r, user)
		return
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}

	targetUser, err := users.CascadeGet(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		LocalError("Only administrators can edit the account of an administrator.", w, r, user)
		return
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
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-user-edit.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_users_edit_submit(w http.ResponseWriter, r *http.Request, user User, suid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditUser {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}

	targetUser, err := users.CascadeGet(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if targetUser.IsAdmin && !user.IsAdmin {
		LocalError("Only administrators can edit the account of an administrator.", w, r, user)
		return
	}

	newname := html.EscapeString(r.PostFormValue("user-name"))
	if newname == "" {
		LocalError("You didn't put in a username.", w, r, user)
		return
	}

	newemail := html.EscapeString(r.PostFormValue("user-email"))
	if newemail == "" {
		LocalError("You didn't put in an email address.", w, r, user)
		return
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		LocalError("You need the EditUserEmail permission to edit the email address of a user.", w, r, user)
		return
	}

	newpassword := r.PostFormValue("user-password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		LocalError("You need the EditUserPassword permission to edit the password of a user.", w, r, user)
		return
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("user-group"))
	if err != nil {
		LocalError("The provided GroupID is not a valid number.", w, r, user)
		return
	}

	if (newgroup > groupCapCount) || (newgroup < 0) || groups[newgroup].Name == "" {
		LocalError("The group you're trying to place this user in doesn't exist.", w, r, user)
		return
	}

	if !user.Perms.EditUserGroupAdmin && groups[newgroup].IsAdmin {
		LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.", w, r, user)
		return
	}
	if !user.Perms.EditUserGroupSuperMod && groups[newgroup].IsMod {
		LocalError("You need the EditUserGroupAdmin permission to assign someone to a super mod group.", w, r, user)
		return
	}

	_, err = update_user_stmt.Exec(newname, newemail, newgroup, targetUser.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	if newpassword != "" {
		SetPassword(targetUser.ID, newpassword)
	}

	err = users.Load(targetUser.ID)
	if err != nil {
		LocalError("This user no longer exists!", w, r, user)
		return
	}

	http.Redirect(w, r, "/panel/users/edit/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
}

func route_panel_groups(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 9
	offset, page, lastPage := pageOffset(stats.Groups, page, perPage)

	// Skip the System group
	offset++

	var count int
	var groupList []GroupAdmin
	for _, group := range groups[offset:] {
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
			return
		}
	}

	err := templates.ExecuteTemplate(w, "panel-groups.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_groups_edit(w http.ResponseWriter, r *http.Request, user User, sgid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w, r, user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.", w, r, user)
		return
	}

	if !groupExists(gid) {
		//log.Print("aaaaa monsters")
		NotFound(w, r)
		return
	}

	group := groups[gid]
	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
		return
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
		return
	}

	var rank string
	if group.IsAdmin {
		rank = "Admin"
	} else if group.IsMod {
		rank = "Mod"
	} else if group.IsBanned {
		rank = "Banned"
	} else if group.ID == 6 {
		rank = "Guest"
	} else {
		rank = "Member"
	}

	disableRank := !user.Perms.EditGroupGlobalPerms || (group.ID == 6)

	pi := PanelEditGroupPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, group.Tag, rank, disableRank}
	if preRenderHooks["pre_render_panel_edit_group"] != nil {
		if runPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-group-edit.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_groups_edit_perms(w http.ResponseWriter, r *http.Request, user User, sgid string) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w, r, user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.", w, r, user)
		return
	}

	if !groupExists(gid) {
		//log.Print("aaaaa monsters")
		NotFound(w, r)
		return
	}

	group := groups[gid]
	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
		return
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
		return
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

	pi := PanelEditGroupPermsPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, localPerms, globalPerms}
	if preRenderHooks["pre_render_panel_edit_group_perms"] != nil {
		if runPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-group-edit-perms.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_groups_edit_submit(w http.ResponseWriter, r *http.Request, user User, sgid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.", w, r, user)
		return
	}

	if !groupExists(gid) {
		//log.Print("aaaaa monsters")
		NotFound(w, r)
		return
	}

	group := groups[gid]
	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
		return
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
		return
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		LocalError("The group name can't be left blank.", w, r, user)
		return
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
			LocalError("You need the EditGroupGlobalPerms permission to change the group type.", w, r, user)
			return
		}

		switch rank {
		case "Admin":
			if !user.Perms.EditGroupAdmin {
				LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.", w, r, user)
				return
			}

			_, err = update_group_rank_stmt.Exec(1, 1, 0, gid)
			if err != nil {
				InternalError(err, w)
				return
			}
			groups[gid].IsAdmin = true
			groups[gid].IsMod = true
			groups[gid].IsBanned = false
		case "Mod":
			if !user.Perms.EditGroupSuperMod {
				LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.", w, r, user)
				return
			}

			_, err = update_group_rank_stmt.Exec(0, 1, 0, gid)
			if err != nil {
				InternalError(err, w)
				return
			}
			groups[gid].IsAdmin = false
			groups[gid].IsMod = true
			groups[gid].IsBanned = false
		case "Banned":
			_, err = update_group_rank_stmt.Exec(0, 0, 1, gid)
			if err != nil {
				InternalError(err, w)
				return
			}
			groups[gid].IsAdmin = false
			groups[gid].IsMod = false
			groups[gid].IsBanned = true
		case "Guest":
			LocalError("You can't designate a group as a guest group.", w, r, user)
			return
		case "Member":
			_, err = update_group_rank_stmt.Exec(0, 0, 0, gid)
			if err != nil {
				InternalError(err, w)
				return
			}
			groups[gid].IsAdmin = false
			groups[gid].IsMod = false
			groups[gid].IsBanned = false
		default:
			LocalError("Invalid group type.", w, r, user)
			return
		}
	}

	_, err = update_group_stmt.Exec(gname, gtag, gid)
	if err != nil {
		InternalError(err, w)
		return
	}
	groups[gid].Name = gname
	groups[gid].Tag = gtag

	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
}

func route_panel_groups_edit_perms_submit(w http.ResponseWriter, r *http.Request, user User, sgid string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.", w, r, user)
		return
	}

	if !groupExists(gid) {
		//log.Print("aaaaa monsters o.o")
		NotFound(w, r)
		return
	}

	group := groups[gid]
	if group.IsAdmin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.", w, r, user)
		return
	}
	if group.IsMod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.", w, r, user)
		return
	}

	//var lpmap map[string]bool = make(map[string]bool)
	var pmap = make(map[string]bool)
	if user.Perms.EditGroupLocalPerms {
		pplist := LocalPermList
		for _, perm := range pplist {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	//var gpmap map[string]bool = make(map[string]bool)
	if user.Perms.EditGroupGlobalPerms {
		gplist := GlobalPermList
		for _, perm := range gplist {
			pvalue := r.PostFormValue("group-perm-" + perm)
			pmap[perm] = (pvalue == "1")
		}
	}

	pjson, err := json.Marshal(pmap)
	if err != nil {
		LocalError("Unable to marshal the data", w, r, user)
		return
	}

	_, err = update_group_perms_stmt.Exec(pjson, gid)
	if err != nil {
		InternalError(err, w)
		return
	}

	err = rebuildGroupPermissions(gid)
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/panel/groups/edit/perms/"+strconv.Itoa(gid), http.StatusSeeOther)
}

func route_panel_groups_create_submit(w http.ResponseWriter, r *http.Request, user User) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	groupName := r.PostFormValue("group-name")
	if groupName == "" {
		LocalError("You need a name for this group!", w, r, user)
		return
	}
	groupTag := r.PostFormValue("group-tag")

	var isAdmin, isMod, isBanned bool
	if user.Perms.EditGroupGlobalPerms {
		groupType := r.PostFormValue("group-type")
		if groupType == "Admin" {
			if !user.Perms.EditGroupAdmin {
				LocalError("You need the EditGroupAdmin permission to create admin groups", w, r, user)
				return
			}
			isAdmin = true
			isMod = true
		} else if groupType == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				LocalError("You need the EditGroupSuperMod permission to create admin groups", w, r, user)
				return
			}
			isMod = true
		} else if groupType == "Banned" {
			isBanned = true
		}
	}

	gid, err := createGroup(groupName, groupTag, isAdmin, isMod, isBanned)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/panel/groups/edit/"+strconv.Itoa(gid), http.StatusSeeOther)
}

func route_panel_themes(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageThemes {
		NoPermissions(w, r, user)
		return
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
			return
		}
	}
	err := templates.ExecuteTemplate(w, "panel-themes.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_themes_set_default(w http.ResponseWriter, r *http.Request, user User, uname string) {
	_, ok := SimplePanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ManageThemes {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	theme, ok := themes[uname]
	if !ok {
		LocalError("The theme isn't registered in the system", w, r, user)
		return
	}
	if theme.Disabled {
		LocalError("You must not enable this theme", w, r, user)
		return
	}

	var isDefault bool
	log.Print("uname", uname)
	err := is_theme_default_stmt.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	hasTheme := err != ErrNoRows
	if hasTheme {
		log.Print("isDefault", isDefault)
		if isDefault {
			LocalError("The theme is already active", w, r, user)
			return
		}
		_, err = update_theme_stmt.Exec(1, uname)
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		_, err := add_theme_stmt.Exec(uname, 1)
		if err != nil {
			InternalError(err, w)
			return
		}
	}

	// TODO: Make this less racey
	changeDefaultThemeMutex.Lock()
	defaultTheme := defaultThemeBox.Load().(string)
	_, err = update_theme_stmt.Exec(0, defaultTheme)
	if err != nil {
		InternalError(err, w)
		return
	}

	log.Print("Setting theme '" + theme.Name + "' as the default theme")
	theme.Active = true
	themes[uname] = theme

	dTheme, ok := themes[defaultTheme]
	if !ok {
		InternalError(errors.New("The default theme is missing"), w)
		return
	}
	dTheme.Active = false
	themes[defaultTheme] = dTheme

	defaultThemeBox.Store(uname)
	resetTemplateOverrides()
	mapThemeTemplates(theme)
	changeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
}

func route_panel_logs_mod(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}

	var logCount int
	err := modlog_count_stmt.QueryRow().Scan(&logCount)
	if err != nil {
		InternalError(err, w)
		return
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 10
	offset, page, lastPage := pageOffset(logCount, page, perPage)

	rows, err := get_modlogs_offset_stmt.Query(offset, perPage)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	var logs []Log
	var action, elementType, ipaddress, doneAt string
	var elementID, actorID int
	for rows.Next() {
		err := rows.Scan(&action, &elementID, &elementType, &ipaddress, &actorID, &doneAt)
		if err != nil {
			InternalError(err, w)
			return
		}

		actor, err := users.CascadeGet(actorID)
		if err != nil {
			actor = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
		}

		switch action {
		case "lock":
			topic, err := topics.CascadeGet(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was locked by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unlock":
			topic, err := topics.CascadeGet(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was reopened by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "stick":
			topic, err := topics.CascadeGet(elementID)
			if err != nil {
				topic = &Topic{Title: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was pinned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unstick":
			topic, err := topics.CascadeGet(elementID)
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
			targetUser, err := users.CascadeGet(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was banned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unban":
			targetUser, err := users.CascadeGet(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was unbanned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "activate":
			targetUser, err := users.CascadeGet(elementID)
			if err != nil {
				targetUser = &User{Name: "Unknown", Link: buildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was activated by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		default:
			action = "Unknown action '" + action + "' by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		}
		logs = append(logs, Log{Action: template.HTML(action), IPAddress: ipaddress, DoneAt: doneAt})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	pageList := paginate(logCount, perPage, 5)
	pi := PanelLogsPage{"Moderation Logs", user, headerVars, stats, logs, pageList, page, lastPage}
	if preRenderHooks["pre_render_panel_mod_log"] != nil {
		if runPreRenderHook("pre_render_panel_mod_log", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "panel-modlogs.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_panel_debug(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, stats, ok := PanelSessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.IsAdmin {
		NoPermissions(w, r, user)
		return
	}

	uptime := "..."
	dbStats := db.Stats()
	openConnCount := dbStats.OpenConnections
	// Disk I/O?

	pi := PanelDebugPage{"Debug", user, headerVars, stats, uptime, openConnCount, dbAdapter}
	err := templates.ExecuteTemplate(w, "panel-debug.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}
