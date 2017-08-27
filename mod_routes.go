package main

import (
	//"log"
	//"fmt"
	"strconv"
	"time"
	"net"
	"net/http"
	"html"
)

func route_edit_topic(w http.ResponseWriter, r *http.Request, user User) {
	//log.Print("in route_edit_topic")
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
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided TopicID is not a valid number.",w,r,is_js)
		return
	}

	old_topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreErrorJSQ("The topic you tried to edit doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,old_topic.ParentID)
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

	if old_topic.Is_Closed != is_closed {
		var action string
		if is_closed {
			action = "lock"
		} else {
			action = "unlock"
		}

		err = addModLog(action,tid,"topic",ipaddress,user.ID)
		if err != nil {
			InternalError(err,w)
			return
		}
		_, err = create_action_reply_stmt.Exec(tid,action,ipaddress,user.ID)
		if err != nil {
			InternalError(err,w)
			return
		}
		_, err = add_replies_to_topic_stmt.Exec(1, user.ID, tid)
		if err != nil {
			InternalError(err,w)
			return
		}
		err = fstore.UpdateLastTopic(topic_name,tid,user.Name,user.ID,time.Now().Format("2006-01-02 15:04:05"),old_topic.ParentID)
		if err != nil && err != ErrNoRows {
			InternalError(err,w)
			return
		}
	}

	err = topics.Load(tid)
	if err == ErrNoRows {
		LocalErrorJSQ("This topic no longer exists!",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	if is_js == "0" {
		http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}
}

func route_delete_topic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to delete doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
		NoPermissions(w,r,user)
		return
	}

	_, err = delete_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("delete",tid,"topic",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	// Might need soft-delete before we can do an action reply for this
	/*_, err = create_action_reply_stmt.Exec(tid,"delete",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}*/

	//log.Print("Topic #" + strconv.Itoa(tid) + " was deleted by User #" + strconv.Itoa(user.ID))
	http.Redirect(w,r,"/",http.StatusSeeOther)

	wcount := word_count(topic.Content)
	err = decrease_post_user_stats(wcount,topic.CreatedBy,true,user)
	if err != nil {
		InternalError(err,w)
		return
	}

	err = fstore.DecrementTopicCount(topic.ParentID)
	if err != nil && err != ErrNoRows {
		InternalError(err,w)
		return
	}
	topics.Remove(tid)
}

func route_stick_topic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to pin doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}

	_, err = stick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("stick",tid,"topic",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}
	_, err = create_action_reply_stmt.Exec(tid,"stick",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	err = topics.Load(tid)
	if err != nil {
		LocalError("This topic doesn't exist!",w,r,user)
		return
	}
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_unstick_topic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to unpin doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w,r,user)
		return
	}

	_, err = unstick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("unstick",tid,"topic",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}
	_, err = create_action_reply_stmt.Exec(tid,"unstick",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	err = topics.Load(tid)
	if err != nil {
		LocalError("This topic doesn't exist!",w,r,user)
		return
	}
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_reply_edit_submit(w http.ResponseWriter, r *http.Request, user User) {
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
	err = get_reply_tid_stmt.QueryRow(rid).Scan(&tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	var fid int
	err = get_topic_fid_stmt.QueryRow(tid).Scan(&fid)
	if err == ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,fid)
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
		w.Write(success_json_bytes)
	}
}

func route_reply_delete_submit(w http.ResponseWriter, r *http.Request, user User) {
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

	reply, err := get_reply(rid)
	if err == ErrNoRows {
		PreErrorJSQ("The reply you tried to delete doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	var fid int
	err = get_topic_fid_stmt.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.",w,r,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,fid)
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
	//log.Print("Reply #" + strconv.Itoa(rid) + " was deleted by User #" + strconv.Itoa(user.ID))
	if is_js == "0" {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}

	wcount := word_count(reply.Content)
	err = decrease_post_user_stats(wcount, reply.CreatedBy, false, user)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
		return
	}
	_, err = remove_replies_from_topic_stmt.Exec(1,reply.ParentID)
	if err != nil {
		InternalErrorJSQ(err,w,r,is_js)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("delete",reply.ParentID,"reply",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	err = topics.Load(reply.ParentID)
	if err != nil {
		LocalError("This topic no longer exists!",w,r,user)
		return
	}
}

func route_profile_reply_edit_submit(w http.ResponseWriter, r *http.Request, user User) {
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
	err = get_user_reply_uid_stmt.QueryRow(rid).Scan(&uid)
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
		http.Redirect(w,r,"/user/" + strconv.Itoa(uid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}
}

func route_profile_reply_delete_submit(w http.ResponseWriter, r *http.Request, user User) {
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
	err = get_user_reply_uid_stmt.QueryRow(rid).Scan(&uid)
	if err == ErrNoRows {
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
	//log.Print("The profile post '" + strconv.Itoa(rid) + "' was deleted by User #" + strconv.Itoa(user.ID))

	if is_js == "0" {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		w.Write(success_json_bytes)
	}
}

// TO-DO: This is being replaced with the new ban route system
/*func route_ban(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
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
	err = get_user_name_stmt.QueryRow(uid).Scan(&uname)
	if err == ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	confirm_msg := "Are you sure you want to ban '" + uname + "'?"
	yousure := AreYouSure{"/users/ban/submit/" + strconv.Itoa(uid),confirm_msg}

	pi := Page{"Ban User",user,headerVars,tList,yousure}
	if pre_render_hooks["pre_render_ban"] != nil {
		if run_pre_render_hook("pre_render_ban", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"areyousure.html",pi)
}*/

func route_ban_submit(w http.ResponseWriter, r *http.Request, user User) {
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
	/*if uid == -2 {
		LocalError("Stop trying to ban Merlin! Ban admin! Bad! No!",w,r,user)
		return
	}*/

	targetUser, err := users.CascadeGet(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	if targetUser.Is_Super_Admin || targetUser.Is_Admin || targetUser.Is_Mod {
		LocalError("You may not ban another staff member.",w,r,user)
		return
	}
	if uid == user.ID {
		LocalError("Why are you trying to ban yourself? Stop that.",w,r,user)
		return
	}

	if targetUser.Is_Banned {
		LocalError("The user you're trying to unban is already banned.",w,r,user)
		return
	}

	duration_days, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of days",w,r,user)
		return
	}

	duration_weeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of weeks",w,r,user)
		return
	}

	duration_months, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of months",w,r,user)
		return
	}

	var duration time.Duration
	if duration_days > 1 && duration_weeks > 1 && duration_months > 1 {
		duration, _ = time.ParseDuration("0")
	} else {
		var seconds int
		seconds += duration_days * day
		seconds += duration_weeks * week
		seconds += duration_months * month
		duration, _ = time.ParseDuration(strconv.Itoa(seconds) + "s")
	}

	err = targetUser.Ban(duration,user.ID)
	if err == ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("ban",uid,"user",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	http.Redirect(w,r,"/user/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_unban(w http.ResponseWriter, r *http.Request, user User) {
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

	targetUser, err := users.CascadeGet(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	if !targetUser.Is_Banned {
		LocalError("The user you're trying to unban isn't banned.",w,r,user)
		return
	}

	err = targetUser.Unban()
	if err == ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("unban",uid,"user",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	http.Redirect(w,r,"/user/" + strconv.Itoa(uid),http.StatusSeeOther)
}

func route_activate(w http.ResponseWriter, r *http.Request, user User) {
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

	var active bool
	err = get_user_active_stmt.QueryRow(uid).Scan(&active)
	if err == ErrNoRows {
		LocalError("The account you're trying to activate no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w)
		return
	}

	if active {
		LocalError("The account you're trying to activate has already been activated.",w,r,user)
		return
	}
	_, err = activate_user_stmt.Exec(uid)
	if err != nil {
		InternalError(err,w)
		return
	}

	_, err = change_group_stmt.Exec(config.DefaultGroup, uid)
	if err != nil {
		InternalError(err,w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	err = addModLog("activate",uid,"user",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}

	err = users.Load(uid)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
		return
	}
	http.Redirect(w,r,"/user/" + strconv.Itoa(uid),http.StatusSeeOther)
}
