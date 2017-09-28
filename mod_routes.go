package main

import (
	//"log"
	//"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_socialgroups
func routeEditTopic(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided TopicID is not a valid number.", w, r, isJs)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreErrorJSQ("The topic you tried to edit doesn't exist.", w, r, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	topicName := r.PostFormValue("topic_name")
	topicContent := html.EscapeString(r.PostFormValue("topic_content"))
	log.Print("topicContent ", topicContent)

	err = topic.Update(topicName, topicContent)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	err = fstore.UpdateLastTopic(topic.ID, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
}

// TODO: Add support for soft-deletion and add a permission just for hard delete
// TODO: Disable stat updates in posts handled by plugin_socialgroups
func routeDeleteTopic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to delete doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
		NoPermissions(w, r, user)
		return
	}

	// We might be able to handle this err better
	err = topics.Delete(topic.CreatedBy)
	if err != nil {
		InternalError(err, w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("delete", tid, "topic", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// ? - We might need to add soft-delete before we can do an action reply for this
	/*_, err = createActionReplyStmt.Exec(tid,"delete",ipaddress,user.ID)
	if err != nil {
		InternalError(err,w)
		return
	}*/

	//log.Print("Topic #" + strconv.Itoa(tid) + " was deleted by User #" + strconv.Itoa(user.ID))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func routeStickTopic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to pin doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w, r, user)
		return
	}

	err = topic.Stick()
	if err != nil {
		InternalError(err, w)
		return
	}

	// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("stick", tid, "topic", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = topic.CreateActionReply("stick", ipaddress, user)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func routeUnstickTopic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to unpin doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		NoPermissions(w, r, user)
		return
	}

	err = topic.Unstick()
	if err != nil {
		InternalError(err, w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("unstick", tid, "topic", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = topic.CreateActionReply("unstick", ipaddress, user)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func routeLockTopic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/lock/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to pin doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		NoPermissions(w, r, user)
		return
	}

	err = topic.Lock()
	if err != nil {
		InternalError(err, w)
		return
	}

	// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("lock", tid, "topic", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = topic.CreateActionReply("lock", ipaddress, user)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func routeUnlockTopic(w http.ResponseWriter, r *http.Request, user User) {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unlock/submit/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The topic you tried to pin doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		NoPermissions(w, r, user)
		return
	}

	err = topic.Unlock()
	if err != nil {
		InternalError(err, w)
		return
	}

	// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("unlock", tid, "topic", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = topic.CreateActionReply("unlock", ipaddress, user)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

// TODO: Disable stat updates in posts handled by plugin_socialgroups
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func routeReplyEditSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
		return
	}

	content := html.EscapeString(preparseMessage(r.PostFormValue("edit_item")))
	_, err = editReplyStmt.Exec(content, parseMessage(content), rid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	// Get the Reply ID..
	var tid int
	err = getReplyTIDStmt.QueryRow(rid).Scan(&tid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	var fid int
	err = getTopicFIDStmt.QueryRow(tid).Scan(&fid)
	if err == ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
}

// TODO: Disable stat updates in posts handled by plugin_socialgroups
func routeReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
		return
	}

	reply, err := getReply(rid)
	if err == ErrNoRows {
		PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	var fid int
	err = getTopicFIDStmt.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	_, err = deleteReplyStmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}
	//log.Print("Reply #" + strconv.Itoa(rid) + " was deleted by User #" + strconv.Itoa(user.ID))
	if !isJs {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	replyCreator, err := users.Get(reply.CreatedBy)
	if err == nil {
		wcount := wordCount(reply.Content)
		err = replyCreator.decreasePostStats(wcount, false)
		if err != nil {
			InternalErrorJSQ(err, w, r, isJs)
			return
		}
	} else if err != ErrNoRows {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}
	_, err = removeRepliesFromTopicStmt.Exec(1, reply.ParentID)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("delete", reply.ParentID, "reply", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	tcache, ok := topics.(TopicCache)
	if ok {
		tcache.CacheRemove(reply.ParentID)
	}
}

func routeProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
		return
	}

	// Get the Reply ID..
	var uid int
	err = getUserReplyUIDStmt.QueryRow(rid).Scan(&uid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	if user.ID != uid && !user.Perms.EditReply {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	content := html.EscapeString(preparseMessage(r.PostFormValue("edit_item")))
	_, err = editProfileReplyStmt.Exec(content, parseMessage(content), rid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(uid)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
}

func routeProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
		return
	}

	var uid int
	err = getUserReplyUIDStmt.QueryRow(rid).Scan(&uid)
	if err == ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.", w, r, user, isJs)
		return
	} else if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}

	if user.ID != uid && !user.Perms.DeleteReply {
		NoPermissionsJSQ(w, r, user, isJs)
		return
	}

	_, err = deleteProfileReplyStmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err, w, r, isJs)
		return
	}
	//log.Print("The profile post '" + strconv.Itoa(rid) + "' was deleted by User #" + strconv.Itoa(user.ID))

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
}

func routeIps(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Perms.ViewIPs {
		NoPermissions(w, r, user)
		return
	}

	ip := r.FormValue("ip")
	var uid int
	var reqUserList = make(map[int]bool)

	rows, err := findUsersByIPUsersStmt.Query(ip)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			InternalError(err, w)
			return
		}
		reqUserList[uid] = true
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	rows2, err := findUsersByIPTopicsStmt.Query(ip)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		err := rows2.Scan(&uid)
		if err != nil {
			InternalError(err, w)
			return
		}
		reqUserList[uid] = true
	}
	err = rows2.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	rows3, err := findUsersByIPRepliesStmt.Query(ip)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows3.Close()

	for rows3.Next() {
		err := rows3.Scan(&uid)
		if err != nil {
			InternalError(err, w)
			return
		}
		reqUserList[uid] = true
	}
	err = rows3.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkGetMap(idSlice)
	if err != nil {
		InternalError(err, w)
		return
	}

	pi := IPSearchPage{"IP Search", user, headerVars, userList, ip}
	if preRenderHooks["pre_render_ips"] != nil {
		if runPreRenderHook("pre_render_ips", w, r, &user, &pi) {
			return
		}
	}
	err = templates.ExecuteTemplate(w, "ip-search.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// TODO: This is being replaced with the new ban route system
/*func routeBan(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w,r,&user)
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
	if preRenderHooks["pre_render_ban"] != nil {
		if runPreRenderHook("pre_render_ban", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"areyousure.html",pi)
}*/

func routeBanSubmit(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Perms.BanUsers {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/submit/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}
	/*if uid == -2 {
		LocalError("Stop trying to ban Merlin! Ban admin! Bad! No!",w,r,user)
		return
	}*/

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if targetUser.IsSuperAdmin || targetUser.IsAdmin || targetUser.IsMod {
		LocalError("You may not ban another staff member.", w, r, user)
		return
	}
	if uid == user.ID {
		LocalError("Why are you trying to ban yourself? Stop that.", w, r, user)
		return
	}

	if targetUser.IsBanned {
		LocalError("The user you're trying to unban is already banned.", w, r, user)
		return
	}

	durationDays, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of days", w, r, user)
		return
	}

	durationWeeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of weeks", w, r, user)
		return
	}

	durationMonths, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		LocalError("You can only use whole numbers for the number of months", w, r, user)
		return
	}

	var duration time.Duration
	if durationDays > 1 && durationWeeks > 1 && durationMonths > 1 {
		duration, _ = time.ParseDuration("0")
	} else {
		var seconds int
		seconds += durationDays * day
		seconds += durationWeeks * week
		seconds += durationMonths * month
		duration, _ = time.ParseDuration(strconv.Itoa(seconds) + "s")
	}

	err = targetUser.Ban(duration, user.ID)
	if err == ErrNoRows {
		LocalError("The user you're trying to ban no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("ban", uid, "user", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
}

func routeUnban(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Perms.BanUsers {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/unban/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if !targetUser.IsBanned {
		LocalError("The user you're trying to unban isn't banned.", w, r, user)
		return
	}

	err = targetUser.Unban()
	if err == ErrNoRows {
		LocalError("The user you're trying to unban no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("unban", uid, "user", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
}

func routeActivate(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Perms.ActivateUsers {
		NoPermissions(w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/activate/"):])
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		LocalError("The account you're trying to activate no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if targetUser.Active {
		LocalError("The account you're trying to activate has already been activated.", w, r, user)
		return
	}
	err = targetUser.Activate()
	if err != nil {
		InternalError(err, w)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}
	err = addModLog("activate", targetUser.ID, "user", ipaddress, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	http.Redirect(w, r, "/user/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
}
