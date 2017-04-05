package main

import "log"
import "fmt"
import "strconv"
import "net"
import "net/http"
import "html"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func route_edit_topic(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	var tid int
	var fid int
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided TopicID is not a valid number.",w,r,is_js)
		return
	}
	
	var old_is_closed bool
	err = db.QueryRow("select parentID, is_closed from topics where tid = ?", tid).Scan(&fid,&old_is_closed)
	if err == sql.ErrNoRows {
		PreErrorJSQ("The topic you tried to edit doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	topic_name := r.PostFormValue("topic_name")
	topic_status := r.PostFormValue("topic_status")
	is_closed := (topic_status == "closed")
	
	topic_content := html.EscapeString(r.PostFormValue("topic_content"))
	_, err = edit_topic_stmt.Exec(topic_name, preparse_message(topic_content), parse_message(html.EscapeString(preparse_message(topic_content))), is_closed, tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	
	if old_is_closed != is_closed {
		var action string
		if is_closed {
			action = "lock"
		} else {
			action = "unlock"
		}
		
		err = addModLog(action,tid,"topic",ipaddress,user.ID)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		_, err = create_action_reply_stmt.Exec(tid,action,ipaddress,user.ID)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		_, err = add_replies_to_topic_stmt.Exec(1, tid)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		_, err = update_forum_cache_stmt.Exec(topic_name, tid, user.Name, user.ID, 1)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	}
	
	err = topics.Load(tid)
	if err != nil {
		LocalErrorJSQ("This topic no longer exists!",w,r,user,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,`{"success":"1"}`)
	}
}

func route_delete_topic(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}
	
	var content string
	var createdBy int
	var fid int
	err = db.QueryRow("select content, createdBy, parentID from topics where tid = ?", tid).Scan(&content, &createdBy, &fid)
	if err == sql.ErrNoRows {
		PreError("The topic you tried to delete doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
		NoPermissions(w,r,user)
		return
	}
	
	_, err = delete_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("delete",tid,"topic",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	/*_, err = create_action_reply_stmt.Exec(tid,"delete",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}*/
	
	log.Print("The topic '" + strconv.Itoa(tid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	http.Redirect(w,r,"/",http.StatusSeeOther)
	
	wcount := word_count(content)
	err = decrease_post_user_stats(wcount,createdBy,true,user)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = remove_topics_from_forum_stmt.Exec(1, fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	forums[fid].TopicCount -= 1
	topics.Remove(tid)
}

func route_stick_topic(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}
	
	topic, err := topics.CascadeGet(tid)
	if err == sql.ErrNoRows {
		PreError("The topic you tried to pin doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}
	
	_, err = stick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	//topic.Sticky = true
	err = topics.Load(tid)
	if err != nil {
		LocalError("This topic doesn't exist!",w,r,user)
		return
	}
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_unstick_topic(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}
	
	topic, err := topics.CascadeGet(tid)
	if err == sql.ErrNoRows {
		PreError("The topic you tried to unpin doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}
	
	_, err = unstick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	//topic.Sticky = false
	err = topics.Load(tid)
	if err != nil {
		LocalError("This topic doesn't exist!",w,r,user)
		return
	}
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided Reply ID is not a valid number.",w,r,is_js)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	// Get the Reply ID..
	var tid int
	err = db.QueryRow("select tid from replies where rid = ?", rid).Scan(&tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	var fid int
	err = db.QueryRow("select parentID from topics where tid = ?", tid).Scan(&fid)
	if err == sql.ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/topic/" + strconv.Itoa(tid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,`{"success":"1"}`)
	}
}

func route_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return          
	}
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided Reply ID is not a valid number.",w,r,is_js)
		return
	}
	
	var tid int
	var content string
	var createdBy int
	err = db.QueryRow("select tid, content, createdBy from replies where rid = ?", rid).Scan(&tid, &content, &createdBy)
	if err == sql.ErrNoRows {
		PreErrorJSQ("The reply you tried to delete doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	var fid int
	err = db.QueryRow("select parentID from topics where tid = ?", tid).Scan(&fid)
	if err == sql.ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	_, err = delete_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	if is_js == "0" {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,`{"success":"1"}`)
	}
	
	wcount := word_count(content)
	err = decrease_post_user_stats(wcount, createdBy, false, user)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	_, err = remove_replies_from_topic_stmt.Exec(1,tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
	}
	
	err = topics.Load(tid)
	if err != nil {
		LocalError("This topic no longer exists!",w,r,user)
		return
	}
}

func route_profile_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	// Get the Reply ID..
	var uid int
	err = db.QueryRow("select uid from users_replies where rid = ?", rid).Scan(&uid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	if user.ID != uid && !user.Perms.EditReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	content := html.EscapeString(preparse_message(r.PostFormValue("edit_item")))
	_, err = edit_profile_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/user/" + strconv.Itoa(uid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,`{"success":"1"}`)
	}
}

func route_profile_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
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
	err = db.QueryRow("select uid from users_replies where rid = ?", rid).Scan(&uid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	
	if user.ID != uid && !user.Perms.DeleteReply {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	_, err = delete_profile_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,`{"success":"1"}`)
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
	err = db.QueryRow("select name from users where uid = ?", uid).Scan(&uname)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	confirm_msg := "Are you sure you want to ban '" + uname + "'?"
	yousure := AreYouSure{"/users/ban/submit/" + strconv.Itoa(uid),confirm_msg}
	
	pi := Page{"Ban User",user,noticeList,tList,yousure}
	templates.ExecuteTemplate(w,"areyousure.html",pi)
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
	err = db.QueryRow("select `group`,`is_super_admin` from `users` where `uid` = ?", uid).Scan(&group, &is_super_admin)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
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
		InternalError(err,w,r)
		return
	}
	
	err = users.Load(uid)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
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
	err = db.QueryRow("select `name`, `group` from users where `uid` = ?", uid).Scan(&uname, &group)
	if err == sql.ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	if !groups[group].Is_Banned {
		LocalError("The user you're trying to unban isn't banned.",w,r,user)
		return
	}
	
	_, err = change_group_stmt.Exec(default_group, uid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	err = users.Load(uid)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
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
	err = db.QueryRow("select `name`,`active` from users where `uid` = ?", uid).Scan(&uname, &active)
	if err == sql.ErrNoRows {
		LocalError("The account you're trying to activate no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	if active {
		LocalError("The account you're trying to activate has already been activated.",w,r,user)
		return
	}
	_, err = activate_user_stmt.Exec(uid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = change_group_stmt.Exec(default_group, uid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	err = users.Load(uid)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
		return
	}
	http.Redirect(w,r,"/users/" + strconv.Itoa(uid),http.StatusSeeOther)
}
