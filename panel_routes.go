package main

import (
	"log"
	"fmt"
	"errors"
	"strings"
	"strconv"
	"html"
	"time"
	"runtime"
	"encoding/json"
	"net/http"
	"html/template"
	"database/sql"
)

import _ "github.com/go-sql-driver/mysql"
import "github.com/shirou/gopsutil/cpu"
import "github.com/shirou/gopsutil/mem"

func route_panel(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}

	var cpustr, cpuColour string
	perc2, err := cpu.Percent(time.Duration(time.Second),true)
	if err != nil {
		cpustr = "Unknown"
	} else {
		calcperc := int(perc2[0]) / runtime.NumCPU()
		cpustr = strconv.Itoa(calcperc)
		if calcperc < 30 {
			cpuColour = "stat_green"
		} else if calcperc < 75 {
			cpuColour = "stat_orange"
		} else {
			cpuColour = "stat_red"
		}
	}

	var ramstr, ramColour string
	memres, err := mem.VirtualMemory()
	if err != nil {
		ramstr = "Unknown"
	} else {
		total_count, total_unit := convert_byte_unit(float64(memres.Total))
		used_count := convert_byte_in_unit(float64(memres.Total - memres.Available),total_unit)

		// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
		//fmt.Println(used_count)
		var totstr string
		if (total_count - float64(int(total_count))) > 0.85 {
			used_count += 1.0 - (total_count - float64(int(total_count)))
			totstr = strconv.Itoa(int(total_count) + 1)
		} else {
			totstr = fmt.Sprintf("%.1f",total_count)
		}
		//fmt.Println(used_count)

		if used_count > total_count {
			used_count = total_count
		}
		ramstr = fmt.Sprintf("%.1f",used_count) + " / " + totstr + total_unit

		ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
		//fmt.Println(ramperc)
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
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}
	var postInterval string = "day"

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
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}
	var topicInterval string = "day"

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
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}
	var reportInterval string = "week"

	var newUserCount int
	err = todays_newuser_count_stmt.QueryRow().Scan(&newUserCount)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}
	var newUserInterval string = "week"

	var gridElements []GridElement = []GridElement{
		GridElement{"dash-version","v" + version.String(),0,"grid_istat stat_green","","","Gosora is up-to-date :)"},
		GridElement{"dash-cpu","CPU: " + cpustr + "%",1,"grid_istat " + cpuColour,"","","The global CPU usage of this server"},
		GridElement{"dash-ram","RAM: " + ramstr,2,"grid_istat " + ramColour,"","","The global RAM usage of this server"},
	}

	if enable_websockets {
		uonline := ws_hub.user_count()
		gonline := ws_hub.guest_count()
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

		totonline, totunit := convert_friendly_unit(totonline)
		uonline, uunit := convert_friendly_unit(uonline)
		gonline, gunit := convert_friendly_unit(gonline)

		gridElements = append(gridElements, GridElement{"dash-totonline",strconv.Itoa(totonline) + totunit + " online",3,"grid_stat " + onlineColour,"","","The number of people who are currently online"})
		gridElements = append(gridElements, GridElement{"dash-gonline",strconv.Itoa(gonline) + gunit + " guests online",4,"grid_stat " + onlineGuestsColour,"","","The number of guests who are currently online"})
		gridElements = append(gridElements, GridElement{"dash-uonline",strconv.Itoa(uonline) + uunit + " users online",5,"grid_stat " + onlineUsersColour,"","","The number of logged-in users who are currently online"})
	}

	gridElements = append(gridElements, GridElement{"dash-postsperday",strconv.Itoa(postCount) + " posts / " + postInterval,6,"grid_stat " + postColour,"","","The number of new posts over the last 24 hours"})
	gridElements = append(gridElements, GridElement{"dash-topicsperday",strconv.Itoa(topicCount) + " topics / " + topicInterval,7,"grid_stat " + topicColour,"","","The number of new topics over the last 24 hours"})
	gridElements = append(gridElements, GridElement{"dash-totonlineperday","20 online / day",8,"grid_stat stat_disabled","","","Coming Soon!"/*"The people online over the last 24 hours"*/})

	gridElements = append(gridElements, GridElement{"dash-searches","8 searches / week",9,"grid_stat stat_disabled","","","Coming Soon!"/*"The number of searches over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-newusers",strconv.Itoa(newUserCount) + " new users / " + newUserInterval,10,"grid_stat","","","The number of new users over the last 7 days"})
	gridElements = append(gridElements, GridElement{"dash-reports",strconv.Itoa(reportCount) + " reports / " + reportInterval,11,"grid_stat","","","The number of reports over the last 7 days"})

	gridElements = append(gridElements, GridElement{"dash-minperuser","2 minutes / user / week",12,"grid_stat stat_disabled","","","Coming Soon!"/*"The average number of number of minutes spent by each active user over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-visitorsperweek","2 visitors / week",13,"grid_stat stat_disabled","","","Coming Soon!"/*"The number of unique visitors we've had over the last 7 days"*/})
	gridElements = append(gridElements, GridElement{"dash-postsperuser","5 posts / user / week",14,"grid_stat stat_disabled","","","Coming Soon!"/*"The average number of posts made by each active user over the past week"*/})

	pi := PanelDashboardPage{"Control Panel Dashboard",user,headerVars,gridElements,extData}
	templates.ExecuteTemplate(w,"panel-dashboard.html",pi)
}

func route_panel_forums(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}

	var forumList []interface{}
	for _, forum := range forums {
		if forum.Name != "" {
			fadmin := ForumAdmin{forum.ID,forum.Name,forum.Desc,forum.Active,forum.Preset,forum.TopicCount,preset_to_lang(forum.Preset)}
			if fadmin.Preset == "" {
				fadmin.Preset = "custom"
			}
			forumList = append(forumList,fadmin)
		}
	}
	pi := Page{"Forum Manager",user,headerVars,forumList,nil}
	err := templates.ExecuteTemplate(w,"panel-forums.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_forums_create_submit(w http.ResponseWriter, r *http.Request){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	fname := r.PostFormValue("forum-name")
	fdesc := r.PostFormValue("forum-desc")
	fpreset := strip_invalid_preset(r.PostFormValue("forum-preset"))
	factive := r.PostFormValue("forum-name")
	active := (factive == "on" || factive == "1" )

	fid, err := create_forum(fname,fdesc,active,fpreset)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	permmap_to_query(preset_to_permmap(fpreset),fid)
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_delete(w http.ResponseWriter, r *http.Request, sfid string){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}

	if !forum_exists(fid) {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}

	confirm_msg := "Are you sure you want to delete the '" + forums[fid].Name + "' forum?"
	yousure := AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid),confirm_msg}

	pi := Page{"Delete Forum",user,headerVars,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html",pi)
}

func route_panel_forums_delete_submit(w http.ResponseWriter, r *http.Request, sfid string) {
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	if !forum_exists(fid) {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}

	err = delete_forum(fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_edit(w http.ResponseWriter, r *http.Request, sfid string) {
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	if !forum_exists(fid) {
		LocalError("The forum you're trying to edit doesn't exist.",w,r,user)
		return
	}

	var forum Forum = forums[fid]
	if forum.Preset == "" {
		forum.Preset = "custom"
	}

	var glist []Group = groups
	var gplist []GroupForumPermPreset
	for gid, group := range glist {
		if gid == 0 {
			continue
		}
		gplist = append(gplist,GroupForumPermPreset{group,forum_perms_to_group_forum_preset(group.Forums[fid])})
	}

	pi := EditForumPage{"Forum Editor",user,headerVars,forum.ID,forum.Name,forum.Desc,forum.Active,forum.Preset,gplist,extData}
	err = templates.ExecuteTemplate(w,"panel-forum-edit.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_forums_edit_submit(w http.ResponseWriter, r *http.Request, sfid string) {
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalErrorJSQ("The provided Forum ID is not a valid number.",w,r,user,is_js)
		return
	}

	forum_name := r.PostFormValue("forum_name")
	forum_desc := r.PostFormValue("forum_desc")
	forum_preset := strip_invalid_preset(r.PostFormValue("forum_preset"))
	forum_active := r.PostFormValue("forum_active")
    if !forum_exists(fid) {
		LocalErrorJSQ("The forum you're trying to edit doesn't exist.",w,r,user,is_js)
		return
	}

	if forum_name == "" {
		forum_name = forums[fid].Name
	}

	var active bool
	if forum_active == "" {
		active = forums[fid].Active
	} else if forum_active == "1" || forum_active == "Show" {
		active = true
	} else {
		active = false
	}

	_, err = update_forum_stmt.Exec(forum_name,forum_desc,active,forum_preset,fid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	if forums[fid].Name != forum_name {
		forums[fid].Name = forum_name
	}
	if forums[fid].Desc != forum_desc {
		forums[fid].Desc = forum_desc
	}
	if forums[fid].Active != active {
		forums[fid].Active = active
	}
	if forums[fid].Preset != forum_preset {
		forums[fid].Preset = forum_preset
	}

	permmap_to_query(preset_to_permmap(forum_preset),fid)

	if is_js == "0" {
		http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}
}

func route_panel_forums_edit_perms_submit(w http.ResponseWriter, r *http.Request, sfid string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		LocalErrorJSQ("The provided Forum ID is not a valid number.",w,r,user,is_js)
		return
	}

	gid, err := strconv.Atoi(r.PostFormValue("gid"))
	if err != nil {
		LocalErrorJSQ("Invalid Group ID",w,r,user,is_js)
		return
	}

	perm_preset := strip_invalid_group_forum_preset(r.PostFormValue("perm_preset"))
	fperms, changed := group_forum_preset_to_forum_perms(perm_preset)
	if changed {
		permupdate_mutex.Lock()
		groups[gid].Forums[fid] = fperms

		perms, err := json.Marshal(fperms)
		if err != nil {
			InternalErrorJSQ(err,w,r,is_js)
			return
		}

		_, err = add_forum_perms_to_group_stmt.Exec(gid,fid,perm_preset,perms)
		if err != nil {
			InternalErrorJSQ(err,w,r,is_js)
			return
		}
		permupdate_mutex.Unlock()

		_, err = update_forum_stmt.Exec(forums[fid].Name,forums[fid].Desc,forums[fid].Active,"",fid)
		if err != nil {
			InternalErrorJSQ(err,w,r,is_js)
			return
		}
		forums[fid].Preset = ""
	}

	if is_js == "0" {
		http.Redirect(w,r,"/panel/forums/edit/" + strconv.Itoa(fid),http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}
}

func route_panel_settings(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}

	var settingList map[string]interface{} = make(map[string]interface{})
	rows, err := get_settings_stmt.Query()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	var sname, scontent, stype string
	for rows.Next() {
		err := rows.Scan(&sname,&scontent,&stype)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		if stype == "list" {
			llist := settingLabels[sname]
			labels := strings.Split(llist,",")
			conv, err := strconv.Atoi(scontent)
			if err != nil {
				LocalError("The setting '" + sname + "' can't be converted to an integer",w,r,user)
				return
			}
			scontent = labels[conv - 1]
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
		InternalError(err,w,r)
		return
	}

	pi := Page{"Setting Manager",user,headerVars,tList,settingList}
	templates.ExecuteTemplate(w,"panel-settings.html",pi)
}

func route_panel_setting(w http.ResponseWriter, r *http.Request, sname string){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}
	setting := Setting{sname,"","",""}

	err := get_setting_stmt.QueryRow(setting.Name).Scan(&setting.Content,&setting.Type)
	if err == sql.ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	var itemList []interface{}
	if setting.Type == "list" {
		llist, ok := settingLabels[setting.Name]
		if !ok {
			LocalError("The labels for this setting don't exist",w,r,user)
			return
		}

		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			LocalError("The value of this setting couldn't be converted to an integer",w,r,user)
			return
		}

		labels := strings.Split(llist,",")
		for index, label := range labels {
			itemList = append(itemList, OptionLabel{
				Label: label,
				Value: index + 1,
				Selected: conv == (index + 1),
			})
		}
	}

	pi := Page{"Edit Setting",user,headerVars,itemList,setting}
	templates.ExecuteTemplate(w,"panel-setting.html",pi)
}

func route_panel_setting_edit(w http.ResponseWriter, r *http.Request, sname string) {
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	var stype, sconstraints string
	scontent := r.PostFormValue("setting-value")

	err = get_full_setting_stmt.QueryRow(sname).Scan(&sname, &stype, &sconstraints)
	if err == sql.ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	if stype == "bool" {
		if scontent == "on" || scontent == "1" {
			scontent = "1"
		} else {
			scontent = "0"
		}
	}

	_, err = update_setting_stmt.Exec(scontent,sname)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	errmsg := parseSetting(sname, scontent, stype, sconstraints)
	if errmsg != "" {
		LocalError(errmsg,w,r,user)
		return
	}
	http.Redirect(w,r,"/panel/settings/",http.StatusSeeOther)
}

func route_panel_plugins(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}

	var pluginList []interface{}
	for _, plugin := range plugins {
		pluginList = append(pluginList,plugin)
	}

	pi := Page{"Plugin Manager",user,headerVars,pluginList,nil}
	templates.ExecuteTemplate(w,"panel-plugins.html",pi)
}

func route_panel_plugins_activate(w http.ResponseWriter, r *http.Request, uname string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system",w,r,user)
		return
	}

	var active bool
	err := is_plugin_active_stmt.QueryRow(uname).Scan(&active)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}

	if plugins[uname].Activate != nil {
		err = plugins[uname].Activate()
		if err != nil {
			LocalError(err.Error(),w,r,user)
			return
		}
	}

	has_plugin := err != sql.ErrNoRows
	if has_plugin {
		if active {
			LocalError("The plugin is already active",w,r,user)
			return
		}
		_, err = update_plugin_stmt.Exec(1,uname)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	} else {
		_, err := add_plugin_stmt.Exec(uname,1)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	}

	log.Print("Activating plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugins[uname] = plugin
	plugins[uname].Init()
	http.Redirect(w,r,"/panel/plugins/",http.StatusSeeOther)
}

func route_panel_plugins_deactivate(w http.ResponseWriter, r *http.Request, uname string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}

	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system",w,r,user)
		return
	}

	var active bool
	err := is_plugin_active_stmt.QueryRow(uname).Scan(&active)
	if err == sql.ErrNoRows {
		LocalError("The plugin you're trying to deactivate isn't active",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	if !active {
		LocalError("The plugin you're trying to deactivate isn't active",w,r,user)
		return
	}
	_, err = update_plugin_stmt.Exec(0,uname)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	plugin.Active = false
	plugins[uname] = plugin
	plugins[uname].Deactivate()

	http.Redirect(w,r,"/panel/plugins/",http.StatusSeeOther)
}

func route_panel_users(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}

	var userList []interface{}
	rows, err := get_users_stmt.Query()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	for rows.Next() {
		puser := User{ID: 0,}
		err := rows.Scan(&puser.ID, &puser.Name, &puser.Group, &puser.Active, &puser.Is_Super_Admin, &puser.Avatar)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		init_user_perms(&puser)
		if puser.Avatar != "" {
			if puser.Avatar[0] == '.' {
				puser.Avatar = "/uploads/avatar_" + strconv.Itoa(puser.ID) + puser.Avatar
			}
		} else {
			puser.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(puser.ID),1)
		}

		if groups[puser.Group].Tag != "" {
			puser.Tag = groups[puser.Group].Tag
		} else {
			puser.Tag = ""
		}
		userList = append(userList,puser)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	pi := Page{"User Manager",user,headerVars,userList,nil}
	err = templates.ExecuteTemplate(w,"panel-users.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_users_edit(w http.ResponseWriter, r *http.Request,suid string){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}

	if !user.Perms.EditUser {
		NoPermissions(w,r,user)
		return
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}

	targetUser, err := users.CascadeGet(uid)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	if targetUser.Is_Admin && !user.Is_Admin {
		LocalError("Only administrators can edit the account of an administrator.",w,r,user)
		return
	}

	var groupList []interface{}
	for _, group := range groups[1:] {
		if !user.Perms.EditUserGroupAdmin && group.Is_Admin {
			continue
		}
		if !user.Perms.EditUserGroupSuperMod && group.Is_Mod {
			continue
		}
		groupList = append(groupList,group)
	}

	pi := Page{"User Editor",user,headerVars,groupList,targetUser}
	err = templates.ExecuteTemplate(w,"panel-user-edit.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_users_edit_submit(w http.ResponseWriter, r *http.Request, suid string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditUser {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}

	targetUser, err := users.CascadeGet(uid)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	if targetUser.Is_Admin && !user.Is_Admin {
		LocalError("Only administrators can edit the account of an administrator.",w,r,user)
		return
	}

	newname := html.EscapeString(r.PostFormValue("user-name"))
	if newname == "" {
		LocalError("You didn't put in a username.",w,r,user)
		return
	}

	newemail := html.EscapeString(r.PostFormValue("user-email"))
	if newemail == "" {
		LocalError("You didn't put in an email address.",w,r,user)
		return
	}
	if (newemail != targetUser.Email) && !user.Perms.EditUserEmail {
		LocalError("You need the EditUserEmail permission to edit the email address of a user.",w,r,user)
		return
	}

	newpassword := r.PostFormValue("user-password")
	if newpassword != "" && !user.Perms.EditUserPassword {
		LocalError("You need the EditUserPassword permission to edit the password of a user.",w,r,user)
		return
	}

	newgroup, err := strconv.Atoi(r.PostFormValue("user-group"))
	if err != nil {
		LocalError("The provided GroupID is not a valid number.",w,r,user)
		return
	}

	if (newgroup > groupCapCount) || (newgroup < 0) || groups[newgroup].Name=="" {
		LocalError("The group you're trying to place this user in doesn't exist.",w,r,user)
		return
	}

	if !user.Perms.EditUserGroupAdmin && groups[newgroup].Is_Admin {
		LocalError("You need the EditUserGroupAdmin permission to assign someone to an administrator group.",w,r,user)
		return
	}
	if !user.Perms.EditUserGroupSuperMod && groups[newgroup].Is_Mod {
		LocalError("You need the EditUserGroupAdmin permission to assign someone to a super mod group.",w,r,user)
		return
	}

	_, err = update_user_stmt.Exec(newname,newemail,newgroup,targetUser.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	if newpassword != "" {
		SetPassword(targetUser.ID,newpassword)
	}

	err = users.Load(targetUser.ID)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
		return
	}

	http.Redirect(w,r,"/panel/users/edit/" + strconv.Itoa(targetUser.ID),http.StatusSeeOther)
}

func route_panel_groups(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}

	var groupList []interface{}
	for _, group := range groups[1:] {
		var rank string
		var rank_class string
		var can_edit bool
		var can_delete bool = false

		if group.Is_Admin {
			rank = "Admin"
			rank_class = "admin"
		} else if group.Is_Mod {
			rank = "Mod"
			rank_class = "mod"
		} else if group.Is_Banned {
			rank = "Banned"
			rank_class = "banned"
		} else if group.ID == 6 {
			rank = "Guest"
			rank_class = "guest"
		} else {
			rank = "Member"
			rank_class = "member"
		}

		can_edit = user.Perms.EditGroup && (!group.Is_Admin || user.Perms.EditGroupAdmin) && (!group.Is_Mod || user.Perms.EditGroupSuperMod)
		groupList = append(groupList, GroupAdmin{group.ID,group.Name,rank,rank_class,can_edit,can_delete})
	}
	//fmt.Printf("%+v\n", groupList)

	pi := Page{"Group Manager",user,headerVars,groupList,nil}
	templates.ExecuteTemplate(w,"panel-groups.html",pi)
}

func route_panel_groups_edit(w http.ResponseWriter, r *http.Request, sgid string){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w,r,user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.",w,r,user)
		return
	}

	if !group_exists(gid) {
		//fmt.Println("aaaaa monsters")
		NotFound(w,r)
		return
	}

	group := groups[gid]
	if group.Is_Admin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.",w,r,user)
		return
	}
	if group.Is_Mod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.",w,r,user)
		return
	}

	var rank string
	if group.Is_Admin {
		rank = "Admin"
	} else if group.Is_Mod {
		rank = "Mod"
	} else if group.Is_Banned {
		rank = "Banned"
	} else if group.ID == 6 {
		rank = "Guest"
	} else {
		rank = "Member"
	}

	disable_rank := !user.Perms.EditGroupGlobalPerms || (group.ID == 6)

	pi := EditGroupPage{"Group Editor",user,headerVars,group.ID,group.Name,group.Tag,rank,disable_rank,extData}
	err = templates.ExecuteTemplate(w,"panel-group-edit.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_groups_edit_perms(w http.ResponseWriter, r *http.Request, sgid string){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w,r,user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.",w,r,user)
		return
	}

	if !group_exists(gid) {
		//fmt.Println("aaaaa monsters")
		NotFound(w,r)
		return
	}

	group := groups[gid]
	if group.Is_Admin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.",w,r,user)
		return
	}
	if group.Is_Mod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.",w,r,user)
		return
	}

	var localPerms []NameLangToggle
	localPerms = append(localPerms, NameLangToggle{"ViewTopic",GetLocalPermPhrase("ViewTopic"),group.Perms.ViewTopic})
	localPerms = append(localPerms, NameLangToggle{"LikeItem",GetLocalPermPhrase("LikeItem"),group.Perms.LikeItem})
	localPerms = append(localPerms, NameLangToggle{"CreateTopic",GetLocalPermPhrase("CreateTopic"),group.Perms.CreateTopic})
	//<--
	localPerms = append(localPerms, NameLangToggle{"EditTopic",GetLocalPermPhrase("EditTopic"),group.Perms.EditTopic})
	localPerms = append(localPerms, NameLangToggle{"DeleteTopic",GetLocalPermPhrase("DeleteTopic"),group.Perms.DeleteTopic})
	localPerms = append(localPerms, NameLangToggle{"CreateReply",GetLocalPermPhrase("CreateReply"),group.Perms.CreateReply})
	localPerms = append(localPerms, NameLangToggle{"EditReply",GetLocalPermPhrase("EditReply"),group.Perms.EditReply})
	localPerms = append(localPerms, NameLangToggle{"DeleteReply",GetLocalPermPhrase("DeleteReply"),group.Perms.DeleteReply})
	localPerms = append(localPerms, NameLangToggle{"PinTopic",GetLocalPermPhrase("PinTopic"),group.Perms.PinTopic})
	localPerms = append(localPerms, NameLangToggle{"CloseTopic",GetLocalPermPhrase("CloseTopic"),group.Perms.CloseTopic})

	var globalPerms []NameLangToggle
	globalPerms = append(globalPerms, NameLangToggle{"BanUsers",GetGlobalPermPhrase("BanUsers"),group.Perms.BanUsers})
	globalPerms = append(globalPerms, NameLangToggle{"ActivateUsers",GetGlobalPermPhrase("ActivateUsers"),group.Perms.ActivateUsers})
	globalPerms = append(globalPerms, NameLangToggle{"EditUser",GetGlobalPermPhrase("EditUser"),group.Perms.EditUser})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserEmail",GetGlobalPermPhrase("EditUserEmail"),group.Perms.EditUserEmail})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserPassword",GetGlobalPermPhrase("EditUserPassword"),group.Perms.EditUserPassword})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroup",GetGlobalPermPhrase("EditUserGroup"),group.Perms.EditUserGroup})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroupSuperMod",GetGlobalPermPhrase("EditUserGroupSuperMod"),group.Perms.EditUserGroupSuperMod})
	globalPerms = append(globalPerms, NameLangToggle{"EditUserGroupAdmin",GetGlobalPermPhrase("EditUserGroupAdmin"),group.Perms.EditUserGroupAdmin})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroup",GetGlobalPermPhrase("EditGroup"),group.Perms.EditGroup})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupLocalPerms",GetGlobalPermPhrase("EditGroupLocalPerms"),group.Perms.EditGroupLocalPerms})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupGlobalPerms",GetGlobalPermPhrase("EditGroupGlobalPerms"),group.Perms.EditGroupGlobalPerms})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupSuperMod",GetGlobalPermPhrase("EditGroupSuperMod"),group.Perms.EditGroupSuperMod})
	globalPerms = append(globalPerms, NameLangToggle{"EditGroupAdmin",GetGlobalPermPhrase("EditGroupAdmin"),group.Perms.EditGroupAdmin})
	globalPerms = append(globalPerms, NameLangToggle{"ManageForums",GetGlobalPermPhrase("ManageForums"),group.Perms.ManageForums})
	globalPerms = append(globalPerms, NameLangToggle{"EditSettings",GetGlobalPermPhrase("EditSettings"),group.Perms.EditSettings})
	globalPerms = append(globalPerms, NameLangToggle{"ManageThemes",GetGlobalPermPhrase("ManageThemes"),group.Perms.ManageThemes})
	globalPerms = append(globalPerms, NameLangToggle{"ManagePlugins",GetGlobalPermPhrase("ManagePlugins"),group.Perms.ManagePlugins})
	globalPerms = append(globalPerms, NameLangToggle{"ViewAdminLogs",GetGlobalPermPhrase("ViewAdminLogs"),group.Perms.ViewAdminLogs})
	globalPerms = append(globalPerms, NameLangToggle{"ViewIPs",GetGlobalPermPhrase("ViewIPs"),group.Perms.ViewIPs})

	pi := EditGroupPermsPage{"Group Editor",user,headerVars,group.ID,group.Name,localPerms,globalPerms,extData}
	err = templates.ExecuteTemplate(w,"panel-group-edit-perms.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_panel_groups_edit_submit(w http.ResponseWriter, r *http.Request, sgid string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.",w,r,user)
		return
	}

	if !group_exists(gid) {
		//fmt.Println("aaaaa monsters")
		NotFound(w,r)
		return
	}

	group := groups[gid]
	if group.Is_Admin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.",w,r,user)
		return
	}
	if group.Is_Mod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.",w,r,user)
		return
	}

	gname := r.FormValue("group-name")
	if gname == "" {
		LocalError("The group name can't be left blank.",w,r,user)
		return
	}
	gtag := r.FormValue("group-tag")
	rank := r.FormValue("group-type")

	var original_rank string
	if group.Is_Admin {
		original_rank = "Admin"
	} else if group.Is_Mod {
		original_rank = "Mod"
	} else if group.Is_Banned {
		original_rank = "Banned"
	} else if group.ID == 6 {
		original_rank = "Guest"
	} else {
		original_rank = "Member"
	}

	group_update_mutex.Lock()
	defer group_update_mutex.Unlock()
	if rank != original_rank {
		if !user.Perms.EditGroupGlobalPerms {
			LocalError("You need the EditGroupGlobalPerms permission to change the group type.",w,r,user)
			return
		}

		switch(rank) {
			case "Admin":
				if !user.Perms.EditGroupAdmin {
					LocalError("You need the EditGroupAdmin permission to designate this group as an admin group.",w,r,user)
					return
				}

				_, err = update_group_rank_stmt.Exec(1,1,0,gid)
				if err != nil {
					InternalError(err,w,r)
					return
				}
				groups[gid].Is_Admin = true
				groups[gid].Is_Mod = true
				groups[gid].Is_Banned = false
			case "Mod":
				if !user.Perms.EditGroupSuperMod {
					LocalError("You need the EditGroupSuperMod permission to designate this group as a super-mod group.",w,r,user)
					return
				}

				_, err = update_group_rank_stmt.Exec(0,1,0,gid)
				if err != nil {
					InternalError(err,w,r)
					return
				}
				groups[gid].Is_Admin = false
				groups[gid].Is_Mod = true
				groups[gid].Is_Banned = false
			case "Banned":
				_, err = update_group_rank_stmt.Exec(0,0,1,gid)
				if err != nil {
					InternalError(err,w,r)
					return
				}
				groups[gid].Is_Admin = false
				groups[gid].Is_Mod = false
				groups[gid].Is_Banned = true
			case "Guest":
				LocalError("You can't designate a group as a guest group.",w,r,user)
				return
			case "Member":
				_, err = update_group_rank_stmt.Exec(0,0,0,gid)
				if err != nil {
					InternalError(err,w,r)
					return
				}
				groups[gid].Is_Admin = false
				groups[gid].Is_Mod = false
				groups[gid].Is_Banned = false
			default:
				LocalError("Invalid group type.",w,r,user)
				return
		}
	}

	_, err = update_group_stmt.Exec(gname,gtag,gid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	groups[gid].Name = gname
	groups[gid].Tag = gtag

	http.Redirect(w,r,"/panel/groups/edit/" + strconv.Itoa(gid),http.StatusSeeOther)
}

func route_panel_groups_edit_perms_submit(w http.ResponseWriter, r *http.Request, sgid string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	gid, err := strconv.Atoi(sgid)
	if err != nil {
		LocalError("The Group ID is not a valid integer.",w,r,user)
		return
	}

	if !group_exists(gid) {
		//fmt.Println("aaaaa monsters")
		NotFound(w,r)
		return
	}

	group := groups[gid]
	if group.Is_Admin && !user.Perms.EditGroupAdmin {
		LocalError("You need the EditGroupAdmin permission to edit an admin group.",w,r,user)
		return
	}
	if group.Is_Mod && !user.Perms.EditGroupSuperMod {
		LocalError("You need the EditGroupSuperMod permission to edit a super-mod group.",w,r,user)
		return
	}

	//var lpmap map[string]bool = make(map[string]bool)
	var pmap map[string]bool = make(map[string]bool)
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
		LocalError("Unable to marshal the data",w,r,user)
		return
	}

	_, err = update_group_perms_stmt.Exec(pjson,gid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	err = rebuild_group_permissions(gid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,"/panel/groups/edit/perms/" + strconv.Itoa(gid),http.StatusSeeOther)
}

func route_panel_groups_create_submit(w http.ResponseWriter, r *http.Request){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.EditGroup {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	group_name := r.PostFormValue("group-name")
	if group_name == "" {
		LocalError("You need a name for this group!",w,r,user)
		return
	}
	group_tag := r.PostFormValue("group-tag")

	var is_admin, is_mod, is_banned bool
	if user.Perms.EditGroupGlobalPerms {
		group_type := r.PostFormValue("group-type")
		if group_type == "Admin" {
			if !user.Perms.EditGroupAdmin {
				LocalError("You need the EditGroupAdmin permission to create admin groups",w,r,user)
				return
			}
			is_admin = true
			is_mod = true
		} else if group_type == "Mod" {
			if !user.Perms.EditGroupSuperMod {
				LocalError("You need the EditGroupSuperMod permission to create admin groups",w,r,user)
				return
			}
			is_mod = true
		} else if group_type == "Banned" {
			is_banned = true
		}
	}

	gid, err := create_group(group_name, group_tag, is_admin, is_mod, is_banned)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	fmt.Println(groups)
	http.Redirect(w,r,"/panel/groups/edit/" + strconv.Itoa(gid),http.StatusSeeOther)
}

func route_panel_themes(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageThemes {
		NoPermissions(w,r,user)
		return
	}

	var pThemeList, vThemeList []Theme
	for _, theme := range themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList,theme)
		} else {
			vThemeList = append(vThemeList,theme)
		}

	}

	pi := ThemesPage{"Theme Manager",user,headerVars,pThemeList,vThemeList,extData}
	err := templates.ExecuteTemplate(w,"panel-themes.html",pi)
	if err != nil {
		log.Print(err)
	}
}

func route_panel_themes_default(w http.ResponseWriter, r *http.Request, uname string){
	user, ok := SimplePanelSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ManageThemes {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	theme, ok := themes[uname]
	if !ok {
		LocalError("The theme isn't registered in the system",w,r,user)
		return
	}
	if theme.Disabled {
		LocalError("You must not enable this theme",w,r,user)
		return
	}

	var isDefault bool
	fmt.Println("uname",uname)
	err := is_theme_default_stmt.QueryRow(uname).Scan(&isDefault)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	}

	has_theme := err != sql.ErrNoRows
	if has_theme {
		fmt.Println("isDefault",isDefault)
		if isDefault {
			LocalError("The theme is already active",w,r,user)
			return
		}
		_, err = update_theme_stmt.Exec(1,uname)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	} else {
		_, err := add_theme_stmt.Exec(uname,1)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	}

	_, err = update_theme_stmt.Exec(0,defaultTheme)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	log.Print("Setting theme '" + theme.Name + "' as the default theme")
	theme.Active = true
	themes[uname] = theme

	dTheme, ok := themes[defaultTheme]
	if !ok {
		InternalError(errors.New("The default theme is missing"),w,r)
		return
	}
	dTheme.Active = false
	themes[defaultTheme] = dTheme

	defaultTheme = uname
	reset_template_overrides()
	add_theme_static_files(themes[uname])
	map_theme_templates(theme)

	http.Redirect(w,r,"/panel/themes/",http.StatusSeeOther)
}

func route_panel_logs_mod(w http.ResponseWriter, r *http.Request){
	user, headerVars, ok := PanelSessionCheck(w,r)
	if !ok {
		return
	}

	rows, err := get_modlogs_stmt.Query()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	var logs []Log
	var action, elementType, ipaddress, doneAt string
	var elementID, actorID int
	for rows.Next() {
		err := rows.Scan(&action,&elementID,&elementType, &ipaddress, &actorID, &doneAt)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		actor, err := users.CascadeGet(actorID)
		if err != nil {
			actor = &User{Name:"Unknown"}
		}

		switch(action) {
			case "lock":
				topic, err := topics.CascadeGet(elementID)
				if err != nil {
					topic = &Topic{Title:"Unknown"}
				}
				action = "<a href='" + build_topic_url(elementID) + "'>" + topic.Title + "</a> was locked by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "unlock":
				topic, err := topics.CascadeGet(elementID)
				if err != nil {
					topic = &Topic{Title:"Unknown"}
				}
				action = "<a href='" + build_topic_url(elementID) + "'>" + topic.Title + "</a> was reopened by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "stick":
				topic, err := topics.CascadeGet(elementID)
				if err != nil {
					topic = &Topic{Title:"Unknown"}
				}
				action = "<a href='" + build_topic_url(elementID) + "'>" + topic.Title + "</a> was pinned by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "unstick":
				topic, err := topics.CascadeGet(elementID)
				if err != nil {
					topic = &Topic{Title:"Unknown"}
				}
				action = "<a href='" + build_topic_url(elementID) + "'>" + topic.Title + "</a> was unpinned by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "delete":
				if elementType == "topic" {
					action = "Topic #" + strconv.Itoa(elementID) + " was deleted by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
				} else {
					topic, err := get_topic_by_reply(elementID)
					if err != nil {
						topic = &Topic{Title:"Unknown"}
					}
					action = "A reply in <a href='" + build_topic_url(elementID) + "'>" + topic.Title + "</a> was deleted by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
				}
			case "ban":
				targetUser, err := users.CascadeGet(elementID)
				if err != nil {
					targetUser = &User{Name:"Unknown"}
				}
				action = "<a href='" + build_profile_url(elementID) + "'>" + targetUser.Name + "</a> was banned by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "unban":
				targetUser, err := users.CascadeGet(elementID)
				if err != nil {
					targetUser = &User{Name:"Unknown"}
				}
				action = "<a href='" + build_profile_url(elementID) + "'>" + targetUser.Name + "</a> was unbanned by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			case "activate":
				targetUser, err := users.CascadeGet(elementID)
				if err != nil {
					targetUser = &User{Name:"Unknown"}
				}
				action = "<a href='" + build_profile_url(elementID) + "'>" + targetUser.Name + "</a> was activated by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
			default:
				action = "Unknown action '" + action + "' by <a href='" + build_profile_url(actorID) + "'>"+actor.Name+"</a>"
		}
		logs = append(logs, Log{Action:template.HTML(action),IPAddress:ipaddress,DoneAt:doneAt})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	pi := LogsPage{"Moderation Logs",user,headerVars,logs,extData}
	err = templates.ExecuteTemplate(w,"panel-modlogs.html",pi)
	if err != nil {
		log.Print(err)
	}
}
