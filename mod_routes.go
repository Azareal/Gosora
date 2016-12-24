package main

import "log"
import "fmt"
import "strings"
import "strconv"
import "net/http"
import "html"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func route_edit_topic(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	var tid int
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided TopicID is not a valid number.",w,r,user,is_js)
		return
	}
	
	topic_name := r.PostFormValue("topic_name")
	topic_status := r.PostFormValue("topic_status")
	
	var is_closed bool
	if topic_status == "closed" {
		is_closed = true
	} else {
		is_closed = false
	}
	
	topic_content := html.EscapeString(r.PostFormValue("topic_content"))
	_, err = edit_topic_stmt.Exec(topic_name, preparse_message(topic_content), parse_message(html.EscapeString(preparse_message(topic_content))), is_closed, tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_delete_topic(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	err = db.QueryRow("SELECT tid from topics where tid = ?", tid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalError("The topic you tried to delete doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = delete_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	log.Print("The topic '" + strconv.Itoa(tid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func route_stick_topic(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = stick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_unstick_topic(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = unstick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		LocalError("The provided Reply ID is not a valid number.",w,r,user)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Get the Reply ID..
	var tid int
	err = db.QueryRow("select tid from replies where rid = ?", rid).Scan(&tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/topic/" + strconv.Itoa(tid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	var tid int
	err = db.QueryRow("SELECT tid from replies where rid = ?", rid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	_, err = delete_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_profile_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		LocalError("The provided Reply ID is not a valid number.",w,r,user)
		return
	}
	
	// Get the Reply ID..
	var uid int
	err = db.QueryRow("select uid from users_replies where rid = ?", rid).Scan(&uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if user.ID != uid && !user.Perms.EditReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_profile_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/user/" + strconv.Itoa(uid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_profile_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	var uid int
	err = db.QueryRow("SELECT uid from users_replies where rid = ?", rid).Scan(&uid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	if user.ID != uid && !user.Perms.DeleteReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	_, err = delete_profile_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_ban(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Perms.BanUsers {
		NoPermissions(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var uname string
	err = db.QueryRow("SELECT name from users where uid = ?", uid).Scan(&uname)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	confirm_msg := "Are you sure you want to ban '" + uname + "'?"
	yousure := AreYouSure{"/users/ban/submit/" + strconv.Itoa(uid),confirm_msg}
	
	pi := Page{"Ban User","ban-user",user,noticeList,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html", pi)
}

func route_ban_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Perms.BanUsers {
		NoPermissions(w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/submit/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var group int
	var is_super_admin bool
	err = db.QueryRow("SELECT `group`, `is_super_admin` from `users` where `uid` = ?", uid).Scan(&group, &is_super_admin)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_super_admin || groups[group].Is_Admin || groups[group].Is_Mod {
		LocalError("You may not ban another staff member.",w,r,user)
		return
	}
	if uid == user.ID {
		LocalError("You may not ban yourself.",w,r,user)
		return
	}
	if uid == -2 {
		LocalError("You may not ban me. Fine, I will offer up some guidance unto thee. Come to my lair, young one. /arcane-tower/",w,r,user)
		return
	}
	
	if groups[group].Is_Banned {
		LocalError("The user you're trying to unban is already banned.",w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(4, uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_unban(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.BanUsers {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/unban/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var uname string
	var group int
	err = db.QueryRow("SELECT `name`, `group` from users where `uid` = ?", uid).Scan(&uname, &group)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if !groups[group].Is_Banned {
		LocalError("The user you're trying to unban isn't banned.",w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(default_group, uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_activate(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Perms.ActivateUsers {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uid, err := strconv.Atoi(r.URL.Path[len("/users/activate/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	var uname string
	var active bool
	err = db.QueryRow("SELECT `name`, `active` from users where `uid` = ?", uid).Scan(&uname, &active)
	if err == sql.ErrNoRows {
		LocalError("The account you're trying to activate no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if active {
		LocalError("The account you're trying to activate has already been activated.",w,r,user)
		return
	}
	
	_, err = activate_user_stmt.Exec(uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(default_group, uid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}

/* Control Panel*/
func route_panel(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod {
		NoPermissions(w,r,user)
		return
	}
	
	pi := Page{"Control Panel Dashboard","panel",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"panel-dashboard.html", pi)
}

func route_panel_forums(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	
	var forumList []interface{}
	for _, forum := range forums {
		if forum.ID > -1 {
			forumList = append(forumList, forum)
		}
	}
	
	pi := Page{"Forum Manager","panel-forums",user,noticeList,forumList,0}
	templates.ExecuteTemplate(w,"panel-forums.html", pi)
}

func route_panel_forums_create_submit(w http.ResponseWriter, r *http.Request){
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fname := r.PostFormValue("forum-name")
	res, err := create_forum_stmt.Exec(fname)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	forums[int(lastId)] = Forum{int(lastId),fname,true,"",0,"",0,""}
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_delete(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/delete/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	_, ok = forums[fid];
    if !ok {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}
	
	confirm_msg := "Are you sure you want to delete the '" + forums[fid].Name + "' forum?"
	yousure := AreYouSure{"/panel/forums/delete/submit/" + strconv.Itoa(fid),confirm_msg}
	
	pi := Page{"Delete Forum","panel-forums-delete",user,noticeList,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html", pi)
}

func route_panel_forums_delete_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/delete/submit/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	_, ok = forums[fid];
    if !ok {
		LocalError("The forum you're trying to delete doesn't exist.",w,r,user)
		return
	}
	
	_, err = delete_forum_stmt.Exec(fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Remove this forum from the forum cache
	delete(forums,fid);
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_forums_edit_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManageForums {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	fid, err := strconv.Atoi(r.URL.Path[len("/panel/forums/edit/submit/"):])
	if err != nil {
		LocalError("The provided Forum ID is not a valid number.",w,r,user)
		return
	}
	
	forum_name := r.PostFormValue("edit_item")
	forum, ok := forums[fid];
    if !ok {
		LocalError("The forum you're trying to edit doesn't exist.",w,r,user)
		return
	}
	
	_, err = update_forum_stmt.Exec(forum_name, fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	forum.Name = forum_name
	forums[fid] = forum
	
	http.Redirect(w,r,"/panel/forums/",http.StatusSeeOther)
}

func route_panel_settings(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}
	
	var settingList map[string]interface{} = make(map[string]interface{})
	rows, err := db.Query("SELECT name, content, type FROM settings")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	var sname string
	var scontent string
	var stype string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype)
		if err != nil {
			InternalError(err,w,r,user)
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
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{"Setting Manager","panel-settings",user, noticeList,tList,settingList}
	templates.ExecuteTemplate(w,"panel-settings.html", pi)
}

func route_panel_setting(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}
	
	setting := Setting{"","","",""}
	setting.Name = r.URL.Path[len("/panel/settings/edit/"):]
	
	err := db.QueryRow("SELECT content, type from settings where name = ?", setting.Name).Scan(&setting.Content, &setting.Type)
	if err == sql.ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
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
	
	pi := Page{"Edit Setting","panel-setting",user,noticeList,itemList,setting}
	templates.ExecuteTemplate(w,"panel-setting.html", pi)
}

func route_panel_setting_edit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.EditSettings {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	var stype string
	var sconstraints string
	sname := r.URL.Path[len("/panel/settings/edit/submit/"):]
	scontent := r.PostFormValue("setting-value")
	
	err = db.QueryRow("SELECT name, type, constraints from settings where name = ?", sname).Scan(&sname, &stype, &sconstraints)
	if err == sql.ErrNoRows {
		LocalError("The setting you want to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
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
		InternalError(err,w,r,user)
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
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}
	
	var pluginList []interface{}
	for _, plugin := range plugins {
		pluginList = append(pluginList, plugin)
	}
	
	pi := Page{"Plugin Manager","panel-plugins",user,noticeList,pluginList,0}
	templates.ExecuteTemplate(w,"panel-plugins.html", pi)
}

func route_panel_plugins_activate(w http.ResponseWriter, r *http.Request){
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uname := r.URL.Path[len("/panel/plugins/activate/"):]
	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system",w,r,user)
		return
	}
	
	var active bool
	err := db.QueryRow("SELECT active from plugins where uname = ?", uname).Scan(&active)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r,user)
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
		_, err = update_plugin_stmt.Exec(1, uname)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
	} else {
		_, err := add_plugin_stmt.Exec(uname,1)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
	}
	
	log.Print("Activating plugin '" + plugin.Name + "'")
	plugin.Active = true
	plugins[uname] = plugin
	plugins[uname].Init()
	http.Redirect(w,r,"/panel/plugins/",http.StatusSeeOther)
}

func route_panel_plugins_deactivate(w http.ResponseWriter, r *http.Request){
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.ManagePlugins {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	uname := r.URL.Path[len("/panel/plugins/deactivate/"):]
	plugin, ok := plugins[uname]
	if !ok {
		LocalError("The plugin isn't registered in the system",w,r,user)
		return
	}
	
	var active bool
	err := db.QueryRow("SELECT active from plugins where uname = ?", uname).Scan(&active)
	if err == sql.ErrNoRows {
		LocalError("The plugin you're trying to deactivate isn't active",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if !active {
		LocalError("The plugin you're trying to deactivate isn't active",w,r,user)
		return
	}
	_, err = update_plugin_stmt.Exec(0, uname)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	plugin.Active = false
	plugins[uname] = plugin
	plugins[uname].Deactivate()
	
	http.Redirect(w,r,"/panel/plugins/",http.StatusSeeOther)
}

func route_panel_users(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod {
		NoPermissions(w,r,user)
		return
	}
	
	var userList []interface{}
	rows, err := db.Query("SELECT `uid`,`name`,`group`,`active`,`is_super_admin`,`avatar` FROM users")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		puser := User{ID: 0,}
		err := rows.Scan(&puser.ID, &puser.Name, &puser.Group, &puser.Active, &puser.Is_Super_Admin, &puser.Avatar)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		puser.Is_Admin = puser.Is_Super_Admin || groups[puser.Group].Is_Admin
		puser.Is_Super_Mod = puser.Is_Admin || groups[puser.Group].Is_Mod
		puser.Is_Mod = puser.Is_Super_Mod
		puser.Is_Banned = groups[puser.Group].Is_Banned
		if puser.Is_Banned && puser.Is_Super_Mod {
			puser.Is_Banned = false
		}
		
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
		
		userList = append(userList, puser)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{"User Manager","panel-users",user,noticeList,userList,0}
	err = templates.ExecuteTemplate(w,"panel-users.html", pi)
	if err != nil {
		InternalError(err, w, r, user)
	}
}

func route_panel_users_edit(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	// Even if they have the right permissions, the control panel is only open to supermods+. There are many areas without subpermissions which assume that the current user is a supermod+ and admins are extremely unlikely to give these permissions to someone who isn't at-least a supermod to begin with
	if !user.Is_Super_Mod || !user.Perms.EditUser {
		NoPermissions(w,r,user)
		return
	}
	
	var err error
	targetUser := User{ID: 0,}
	targetUser.ID, err = strconv.Atoi(r.URL.Path[len("/panel/users/edit/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	err = db.QueryRow("SELECT `name`, `email`, `group` from `users` where `uid` = ?", targetUser.ID).Scan(&targetUser.Name, &targetUser.Email, &targetUser.Group)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	targetUser.Is_Admin = targetUser.Is_Super_Admin || groups[targetUser.Group].Is_Admin
	targetUser.Is_Super_Mod = groups[targetUser.Group].Is_Mod || targetUser.Is_Admin
	if targetUser.Is_Admin && !user.Is_Admin {
		LocalError("Only administrators can edit the account of an administrator.",w,r,user)
		return
	}
	
	var groupList []interface{}
	for _, group := range groups {
		if !user.Perms.EditUserGroupAdmin && group.Is_Admin {
			continue
		}
		if !user.Perms.EditUserGroupSuperMod && group.Is_Mod {
			continue
		}
		groupList = append(groupList, group)
	}
	
	pi := Page{"User Editor","panel-user-edit",user,noticeList,groupList,targetUser}
	err = templates.ExecuteTemplate(w,"panel-user-edit.html", pi)
	if err != nil {
		InternalError(err, w, r, user)
	}
}

func route_panel_users_edit_submit(w http.ResponseWriter, r *http.Request){
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Is_Super_Mod || !user.Perms.EditUser {
		NoPermissions(w,r,user)
		return
	}
	
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	var err error
	targetUser := User{ID: 0,}
	targetUser.ID, err = strconv.Atoi(r.URL.Path[len("/panel/users/edit/submit/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}
	
	err = db.QueryRow("SELECT `name`, `email`, `group` from `users` where `uid` = ?", targetUser.ID).Scan(&targetUser.Name, &targetUser.Email, &targetUser.Group)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to edit doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	targetUser.Is_Admin = targetUser.Is_Super_Admin || groups[targetUser.Group].Is_Admin
	targetUser.Is_Super_Mod = groups[targetUser.Group].Is_Mod || targetUser.Is_Admin
	if targetUser.Is_Admin && !user.Is_Admin {
		LocalError("Only administrators can edit the account of an administrator.",w,r,user)
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
		LocalError("The provided GroupID is not a valid number.",w,r,user)
		return
	}
	
	_, ok = groups[newgroup]
	if !ok {
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
		InternalError(err,w,r,user)
		return
	}
	
	if newpassword != "" {
		SetPassword(targetUser.ID, newpassword)
	}
	
	http.Redirect(w,r,"/panel/users/edit/" + strconv.Itoa(targetUser.ID),http.StatusSeeOther)
}
