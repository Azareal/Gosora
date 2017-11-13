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

	"./common"
	"github.com/Azareal/gopsutil/mem"
)

func routePanel(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
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
		totalCount, totalUnit := common.ConvertByteUnit(float64(memres.Total))
		usedCount := common.ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

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
	err = stmts.todaysPostCount.QueryRow().Scan(&postCount)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
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
	err = stmts.todaysTopicCount.QueryRow().Scan(&topicCount)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
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
	err = stmts.todaysReportCount.QueryRow().Scan(&reportCount)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	var reportInterval = "week"

	var newUserCount int
	err = stmts.todaysNewUserCount.QueryRow().Scan(&newUserCount)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	var newUserInterval = "week"

	var gridElements = []common.GridElement{
		common.GridElement{"dash-version", "v" + version.String(), 0, "grid_istat stat_green", "", "", "Gosora is up-to-date :)"},
		common.GridElement{"dash-cpu", "CPU: " + cpustr, 1, "grid_istat " + cpuColour, "", "", "The global CPU usage of this server"},
		common.GridElement{"dash-ram", "RAM: " + ramstr, 2, "grid_istat " + ramColour, "", "", "The global RAM usage of this server"},
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

		totonline, totunit := common.ConvertFriendlyUnit(totonline)
		uonline, uunit := common.ConvertFriendlyUnit(uonline)
		gonline, gunit := common.ConvertFriendlyUnit(gonline)

		gridElements = append(gridElements, common.GridElement{"dash-totonline", strconv.Itoa(totonline) + totunit + " online", 3, "grid_stat " + onlineColour, "", "", "The number of people who are currently online"})
		gridElements = append(gridElements, common.GridElement{"dash-gonline", strconv.Itoa(gonline) + gunit + " guests online", 4, "grid_stat " + onlineGuestsColour, "", "", "The number of guests who are currently online"})
		gridElements = append(gridElements, common.GridElement{"dash-uonline", strconv.Itoa(uonline) + uunit + " users online", 5, "grid_stat " + onlineUsersColour, "", "", "The number of logged-in users who are currently online"})
	}

	gridElements = append(gridElements, common.GridElement{"dash-postsperday", strconv.Itoa(postCount) + " posts / " + postInterval, 6, "grid_stat " + postColour, "", "", "The number of new posts over the last 24 hours"})
	gridElements = append(gridElements, common.GridElement{"dash-topicsperday", strconv.Itoa(topicCount) + " topics / " + topicInterval, 7, "grid_stat " + topicColour, "", "", "The number of new topics over the last 24 hours"})
	gridElements = append(gridElements, common.GridElement{"dash-totonlineperday", "20 online / day", 8, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The people online over the last 24 hours"*/})

	gridElements = append(gridElements, common.GridElement{"dash-searches", "8 searches / week", 9, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of searches over the last 7 days"*/})
	gridElements = append(gridElements, common.GridElement{"dash-newusers", strconv.Itoa(newUserCount) + " new users / " + newUserInterval, 10, "grid_stat", "", "", "The number of new users over the last 7 days"})
	gridElements = append(gridElements, common.GridElement{"dash-reports", strconv.Itoa(reportCount) + " reports / " + reportInterval, 11, "grid_stat", "", "", "The number of reports over the last 7 days"})

	gridElements = append(gridElements, common.GridElement{"dash-minperuser", "2 minutes / user / week", 12, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of number of minutes spent by each active user over the last 7 days"*/})
	gridElements = append(gridElements, common.GridElement{"dash-visitorsperweek", "2 visitors / week", 13, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The number of unique visitors we've had over the last 7 days"*/})
	gridElements = append(gridElements, common.GridElement{"dash-postsperuser", "5 posts / user / week", 14, "grid_stat stat_disabled", "", "", "Coming Soon!" /*"The average number of posts made by each active user over the past week"*/})

	pi := common.PanelDashboardPage{"Control Panel Dashboard", user, headerVars, stats, gridElements}
	if common.PreRenderHooks["pre_render_panel_dashboard"] != nil {
		if common.RunPreRenderHook("pre_render_panel_dashboard", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-dashboard.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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
	forums, err := common.Fstore.GetAll()
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
	pi := common.PanelPage{"Forum Manager", user, headerVars, stats, forumList, nil}
	if common.PreRenderHooks["pre_render_panel_forums"] != nil {
		if common.RunPreRenderHook("pre_render_panel_forums", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-forums.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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

	_, err := common.Fstore.Create(fname, fdesc, active, fpreset)
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

	forum, err := common.Fstore.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to delete doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	confirmMsg := "Are you sure you want to delete the '" + forum.Name + "' forum?"
	yousure := common.AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid), confirmMsg}

	pi := common.PanelPage{"Delete Forum", user, headerVars, stats, tList, yousure}
	if common.PreRenderHooks["pre_render_panel_delete_forum"] != nil {
		if common.RunPreRenderHook("pre_render_panel_delete_forum", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "areyousure.html", pi)
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

	err = common.Fstore.Delete(fid)
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

	forum, err := common.Fstore.Get(fid)
	if err == ErrNoRows {
		return common.LocalError("The forum you're trying to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	glist, err := common.Gstore.GetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	var gplist []common.GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		gplist = append(gplist, common.GroupForumPermPreset{group, common.ForumPermsToGroupForumPreset(group.Forums[fid])})
	}

	pi := common.PanelEditForumPage{"Forum Editor", user, headerVars, stats, forum.ID, forum.Name, forum.Desc, forum.Active, forum.Preset, gplist}
	if common.PreRenderHooks["pre_render_panel_edit_forum"] != nil {
		if common.RunPreRenderHook("pre_render_panel_edit_forum", w, r, &user, &pi) {
			return nil
		}
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

	forum, err := common.Fstore.Get(fid)
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

	if !isJs {
		http.Redirect(w, r, "/panel/forums/", http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
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

	forum, err := common.Fstore.Get(fid)
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

	if !isJs {
		http.Redirect(w, r, "/panel/forums/edit/"+strconv.Itoa(fid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
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
	rows, err := stmts.getSettings.Query()
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	// nolint need the type so people viewing this file understand what it returns without visiting setting.go
	var settingLabels map[string]string = common.GetAllSettingLabels()
	var sname, scontent, stype string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		if stype == "list" {
			llist := settingLabels[sname]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(scontent)
			if err != nil {
				return common.LocalError("The setting '"+sname+"' can't be converted to an integer", w, r, user)
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
		return common.InternalError(err, w, r)
	}

	pi := common.PanelPage{"Setting Manager", user, headerVars, stats, tList, settingList}
	if common.PreRenderHooks["pre_render_panel_settings"] != nil {
		if common.RunPreRenderHook("pre_render_panel_settings", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-settings.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelSetting(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
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

	pi := common.PanelPage{"Edit Setting", user, headerVars, stats, itemList, setting}
	if common.PreRenderHooks["pre_render_panel_setting"] != nil {
		if common.RunPreRenderHook("pre_render_panel_setting", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-setting.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelSettingEdit(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
	headerLite, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	scontent := r.PostFormValue("setting-value")
	setting, err := headerLite.Settings.BypassGet(sname)
	if err == ErrNoRows {
		return common.LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if setting.Type == "bool" {
		if scontent == "on" || scontent == "1" {
			scontent = "1"
		} else {
			scontent = "0"
		}
	}

	// TODO: Make this a method or function?
	_, err = stmts.updateSetting.Exec(scontent, sname)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	errmsg := headerLite.Settings.ParseSetting(sname, scontent, setting.Type, setting.Constraint)
	if errmsg != "" {
		return common.LocalError(errmsg, w, r, user)
	}
	// TODO: Do a reload instead?
	common.SettingBox.Store(headerLite.Settings)

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
	pi := common.PanelPage{"Word Filter Manager", user, headerVars, stats, tList, filterList}
	if common.PreRenderHooks["pre_render_panel_word_filters"] != nil {
		if common.RunPreRenderHook("pre_render_panel_word_filters", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "panel-word-filters.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelWordFiltersCreate(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
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

	if !isJs {
		http.Redirect(w, r, "/panel/settings/word-filters/", http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routePanelWordFiltersEdit(w http.ResponseWriter, r *http.Request, user common.User, wfid string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	_ = wfid

	pi := common.PanelPage{"Edit Word Filter", user, headerVars, stats, tList, nil}
	if common.PreRenderHooks["pre_render_panel_word_filters_edit"] != nil {
		if common.RunPreRenderHook("pre_render_panel_word_filters_edit", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "panel-word-filters-edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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

	pi := common.PanelPage{"Plugin Manager", user, headerVars, stats, pluginList, nil}
	if common.PreRenderHooks["pre_render_panel_plugins"] != nil {
		if common.RunPreRenderHook("pre_render_panel_plugins", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "panel-plugins.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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
		if puser.Avatar != "" {
			if puser.Avatar[0] == '.' {
				puser.Avatar = "/uploads/avatar_" + strconv.Itoa(puser.ID) + puser.Avatar
			}
		} else {
			puser.Avatar = strings.Replace(common.Config.Noavatar, "{id}", strconv.Itoa(puser.ID), 1)
		}

		if common.Gstore.DirtyGet(puser.Group).Tag != "" {
			puser.Tag = common.Gstore.DirtyGet(puser.Group).Tag
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
	pi := common.PanelUserPage{"User Manager", user, headerVars, stats, userList, pageList, page, lastPage}
	if common.PreRenderHooks["pre_render_panel_users"] != nil {
		if common.RunPreRenderHook("pre_render_panel_users", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-users.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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
		return common.LocalError("The provided common.User ID is not a valid number.", w, r, user)
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
	groups, err := common.Gstore.GetRange(1, 0) // ? - 0 = Go to the end
	if err != nil {
		return common.InternalError(err, w, r)
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

	pi := common.PanelPage{"User Editor", user, headerVars, stats, groupList, targetUser}
	if common.PreRenderHooks["pre_render_panel_edit_user"] != nil {
		if common.RunPreRenderHook("pre_render_panel_edit_user", w, r, &user, &pi) {
			return nil
		}
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
		return common.LocalError("The provided common.User ID is not a valid number.", w, r, user)
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

	newname := html.EscapeString(r.PostFormValue("user-name"))
	if newname == "" {
		return common.LocalError("You didn't put in a username.", w, r, user)
	}

	newemail := html.EscapeString(r.PostFormValue("user-email"))
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

	group, err := common.Gstore.Get(newgroup)
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

	_, err = stmts.updateUser.Exec(newname, newemail, newgroup, targetUser.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	if newpassword != "" {
		common.SetPassword(targetUser.ID, newpassword)
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
	groups, _ := common.Gstore.GetRange(offset, 0)
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
		groupList = append(groupList, common.GroupAdmin{group.ID, group.Name, rank, rankClass, canEdit, canDelete})
		count++
	}
	//log.Printf("groupList: %+v\n", groupList)

	pageList := common.Paginate(stats.Groups, perPage, 5)
	pi := common.PanelGroupPage{"Group Manager", user, headerVars, stats, groupList, pageList, page, lastPage}
	if common.PreRenderHooks["pre_render_panel_groups"] != nil {
		if common.RunPreRenderHook("pre_render_panel_groups", w, r, &user, &pi) {
			return nil
		}
	}

	err := common.Templates.ExecuteTemplate(w, "panel-groups.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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

	group, err := common.Gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r)
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

	pi := common.PanelEditGroupPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, group.Tag, rank, disableRank}
	if common.PreRenderHooks["pre_render_panel_edit_group"] != nil {
		if common.RunPreRenderHook("pre_render_panel_edit_group", w, r, &user, &pi) {
			return nil
		}
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

	group, err := common.Gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r)
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
	localPerms = append(localPerms, common.NameLangToggle{"ViewTopic", common.GetLocalPermPhrase("ViewTopic"), group.Perms.ViewTopic})
	localPerms = append(localPerms, common.NameLangToggle{"LikeItem", common.GetLocalPermPhrase("LikeItem"), group.Perms.LikeItem})
	localPerms = append(localPerms, common.NameLangToggle{"CreateTopic", common.GetLocalPermPhrase("CreateTopic"), group.Perms.CreateTopic})
	//<--
	localPerms = append(localPerms, common.NameLangToggle{"EditTopic", common.GetLocalPermPhrase("EditTopic"), group.Perms.EditTopic})
	localPerms = append(localPerms, common.NameLangToggle{"DeleteTopic", common.GetLocalPermPhrase("DeleteTopic"), group.Perms.DeleteTopic})
	localPerms = append(localPerms, common.NameLangToggle{"CreateReply", common.GetLocalPermPhrase("CreateReply"), group.Perms.CreateReply})
	localPerms = append(localPerms, common.NameLangToggle{"EditReply", common.GetLocalPermPhrase("EditReply"), group.Perms.EditReply})
	localPerms = append(localPerms, common.NameLangToggle{"DeleteReply", common.GetLocalPermPhrase("DeleteReply"), group.Perms.DeleteReply})
	localPerms = append(localPerms, common.NameLangToggle{"PinTopic", common.GetLocalPermPhrase("PinTopic"), group.Perms.PinTopic})
	localPerms = append(localPerms, common.NameLangToggle{"CloseTopic", common.GetLocalPermPhrase("CloseTopic"), group.Perms.CloseTopic})

	var globalPerms []common.NameLangToggle
	globalPerms = append(globalPerms, common.NameLangToggle{"BanUsers", common.GetGlobalPermPhrase("BanUsers"), group.Perms.BanUsers})
	globalPerms = append(globalPerms, common.NameLangToggle{"ActivateUsers", common.GetGlobalPermPhrase("ActivateUsers"), group.Perms.ActivateUsers})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUser", common.GetGlobalPermPhrase("EditUser"), group.Perms.EditUser})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUserEmail", common.GetGlobalPermPhrase("EditUserEmail"), group.Perms.EditUserEmail})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUserPassword", common.GetGlobalPermPhrase("EditUserPassword"), group.Perms.EditUserPassword})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUserGroup", common.GetGlobalPermPhrase("EditUserGroup"), group.Perms.EditUserGroup})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUserGroupSuperMod", common.GetGlobalPermPhrase("EditUserGroupSuperMod"), group.Perms.EditUserGroupSuperMod})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditUserGroupAdmin", common.GetGlobalPermPhrase("EditUserGroupAdmin"), group.Perms.EditUserGroupAdmin})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditGroup", common.GetGlobalPermPhrase("EditGroup"), group.Perms.EditGroup})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditGroupLocalPerms", common.GetGlobalPermPhrase("EditGroupLocalPerms"), group.Perms.EditGroupLocalPerms})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditGroupGlobalPerms", common.GetGlobalPermPhrase("EditGroupGlobalPerms"), group.Perms.EditGroupGlobalPerms})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditGroupSuperMod", common.GetGlobalPermPhrase("EditGroupSuperMod"), group.Perms.EditGroupSuperMod})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditGroupAdmin", common.GetGlobalPermPhrase("EditGroupAdmin"), group.Perms.EditGroupAdmin})
	globalPerms = append(globalPerms, common.NameLangToggle{"ManageForums", common.GetGlobalPermPhrase("ManageForums"), group.Perms.ManageForums})
	globalPerms = append(globalPerms, common.NameLangToggle{"EditSettings", common.GetGlobalPermPhrase("EditSettings"), group.Perms.EditSettings})
	globalPerms = append(globalPerms, common.NameLangToggle{"ManageThemes", common.GetGlobalPermPhrase("ManageThemes"), group.Perms.ManageThemes})
	globalPerms = append(globalPerms, common.NameLangToggle{"ManagePlugins", common.GetGlobalPermPhrase("ManagePlugins"), group.Perms.ManagePlugins})
	globalPerms = append(globalPerms, common.NameLangToggle{"ViewAdminLogs", common.GetGlobalPermPhrase("ViewAdminLogs"), group.Perms.ViewAdminLogs})
	globalPerms = append(globalPerms, common.NameLangToggle{"ViewIPs", common.GetGlobalPermPhrase("ViewIPs"), group.Perms.ViewIPs})
	globalPerms = append(globalPerms, common.NameLangToggle{"UploadFiles", common.GetGlobalPermPhrase("UploadFiles"), group.Perms.UploadFiles})

	pi := common.PanelEditGroupPermsPage{"Group Editor", user, headerVars, stats, group.ID, group.Name, localPerms, globalPerms}
	if common.PreRenderHooks["pre_render_panel_edit_group_perms"] != nil {
		if common.RunPreRenderHook("pre_render_panel_edit_group_perms", w, r, &user, &pi) {
			return nil
		}
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

	group, err := common.Gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters")
		return common.NotFound(w, r)
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
	common.Gstore.Reload(gid)

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

	group, err := common.Gstore.Get(gid)
	if err == ErrNoRows {
		//log.Print("aaaaa monsters o.o")
		return common.NotFound(w, r)
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

	gid, err := common.Gstore.Create(groupName, groupTag, isAdmin, isMod, isBanned)
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

	var pThemeList, vThemeList []common.Theme
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

	pi := common.PanelThemesPage{"Theme Manager", user, headerVars, stats, pThemeList, vThemeList}
	if common.PreRenderHooks["pre_render_panel_themes"] != nil {
		if common.RunPreRenderHook("pre_render_panel_themes", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "panel-themes.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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
	common.MapThemeTemplates(theme)
	common.ChangeDefaultThemeMutex.Unlock()

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func routePanelBackups(w http.ResponseWriter, r *http.Request, user common.User, backupURL string) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.IsSuperAdmin {
		return common.NoPermissions(w, r, user)
	}

	if backupURL != "" {
		// We don't want them trying to break out of this directory, it shouldn't hurt since it's a super admin, but it's always good to practice good security hygiene, especially if this is one of many instances on a managed server not controlled by the superadmin/s
		backupURL = common.Stripslashes(backupURL)

		var ext = filepath.Ext("./backups/" + backupURL)
		if ext == ".sql" {
			info, err := os.Stat("./backups/" + backupURL)
			if err != nil {
				return common.NotFound(w, r)
			}
			// TODO: Change the served filename to gosora_backup_%timestamp%.sql, the time the file was generated, not when it was modified aka what the name of it should be
			w.Header().Set("Content-Disposition", "attachment; filename=gosora_backup.sql")
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
			// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
			http.ServeFile(w, r, "./backups/"+backupURL)
			return nil
		}
		return common.NotFound(w, r)
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

	pi := common.PanelBackupPage{"Backups", user, headerVars, stats, backupList}
	err = common.Templates.ExecuteTemplate(w, "panel-backups.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routePanelLogsMod(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var logCount int
	err := stmts.modlogCount.QueryRow().Scan(&logCount)
	if err != nil {
		return common.InternalError(err, w, r)
	}

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

		actor, err := common.Users.Get(actorID)
		if err != nil {
			actor = &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
		}

		switch action {
		case "lock":
			topic, err := common.Topics.Get(elementID)
			if err != nil {
				topic = &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was locked by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unlock":
			topic, err := common.Topics.Get(elementID)
			if err != nil {
				topic = &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was reopened by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "stick":
			topic, err := common.Topics.Get(elementID)
			if err != nil {
				topic = &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was pinned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unstick":
			topic, err := common.Topics.Get(elementID)
			if err != nil {
				topic = &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + topic.Link + "'>" + topic.Title + "</a> was unpinned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "delete":
			if elementType == "topic" {
				action = "Topic #" + strconv.Itoa(elementID) + " was deleted by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
			} else {
				reply := common.BlankReply()
				reply.ID = elementID
				topic, err := reply.Topic()
				if err != nil {
					topic = &common.Topic{Title: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
				}
				action = "A reply in <a href='" + topic.Link + "'>" + topic.Title + "</a> was deleted by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
			}
		case "ban":
			targetUser, err := common.Users.Get(elementID)
			if err != nil {
				targetUser = &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was banned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "unban":
			targetUser, err := common.Users.Get(elementID)
			if err != nil {
				targetUser = &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was unbanned by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		case "activate":
			targetUser, err := common.Users.Get(elementID)
			if err != nil {
				targetUser = &common.User{Name: "Unknown", Link: common.BuildProfileURL("unknown", 0)}
			}
			action = "<a href='" + targetUser.Link + "'>" + targetUser.Name + "</a> was activated by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		default:
			action = "Unknown action '" + action + "' by <a href='" + actor.Link + "'>" + actor.Name + "</a>"
		}
		logs = append(logs, common.LogItem{Action: template.HTML(action), IPAddress: ipaddress, DoneAt: doneAt})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.PanelLogsPage{"Moderation Logs", user, headerVars, stats, logs, pageList, page, lastPage}
	if common.PreRenderHooks["pre_render_panel_mod_log"] != nil {
		if common.RunPreRenderHook("pre_render_panel_mod_log", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "panel-modlogs.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
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

	pi := common.PanelDebugPage{"Debug", user, headerVars, stats, uptime, openConnCount, dbAdapter}
	err := common.Templates.ExecuteTemplate(w, "panel-debug.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
