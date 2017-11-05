package main

import (
	//"log"
	//"fmt"
	"encoding/json"
	"html"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_guilds
func routeEditTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("js") == "1")

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		return PreErrorJSQ("The provided TopicID is not a valid number.", w, r, isJs)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreErrorJSQ("The topic you tried to edit doesn't exist.", w, r, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	topicName := r.PostFormValue("topic_name")
	topicContent := html.EscapeString(r.PostFormValue("topic_content"))
	err = topic.Update(topicName, topicContent)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	err = fstore.UpdateLastTopic(topic.ID, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Add support for soft-deletion and add a permission for hard delete in addition to the usual
// TODO: Disable stat updates in posts handled by plugin_guilds
func routeDeleteTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if r.Header.Get("Content-type") == "application/json" {
		if r.Body == nil {
			return PreErrorJS("No request body", w, r)
		}
		//log.Print("r.Body: ", r.Body)
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			//log.Print("parse err: ", err)
			return PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
		if err != nil {
			return PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := topics.Get(tid)
		if err == ErrNoRows {
			return PreErrorJSQ("The topic you tried to delete doesn't exist.", w, r, isJs)
		} else if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
			return NoPermissionsJSQ(w, r, user, isJs)
		}

		// We might be able to handle this err better
		err = topic.Delete()
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return LocalErrorJSQ("Bad IP", w, r, user, isJs)
		}
		err = addModLog("delete", tid, "topic", ipaddress, user.ID)
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		// ? - We might need to add soft-delete before we can do an action reply for this
		/*_, err = stmts.createActionReply.Exec(tid,"delete",ipaddress,user.ID)
		if err != nil {
			return InternalErrorJSQ(err,w,r,isJs)
		}*/

		log.Printf("Topic #%d was deleted by User #%d", tid, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeStickTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		return PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreError("The topic you tried to pin doesn't exist.", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return NoPermissions(w, r, user)
	}

	err = topic.Stick()
	if err != nil {
		return InternalError(err, w, r)
	}

	// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("stick", tid, "topic", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	err = topic.CreateActionReply("stick", ipaddress, user)
	if err != nil {
		return InternalError(err, w, r)
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeUnstickTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		return PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreError("The topic you tried to unpin doesn't exist.", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return NoPermissions(w, r, user)
	}

	err = topic.Unstick()
	if err != nil {
		return InternalError(err, w, r)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("unstick", tid, "topic", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unstick", ipaddress, user)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeLockTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if r.Header.Get("Content-type") == "application/json" {
		if r.Body == nil {
			return PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/lock/submit/"):])
		if err != nil {
			return PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := topics.Get(tid)
		if err == ErrNoRows {
			return PreErrorJSQ("The topic you tried to lock doesn't exist.", w, r, isJs)
		} else if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
			return NoPermissionsJSQ(w, r, user, isJs)
		}

		err = topic.Lock()
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}

		// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
		ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return LocalErrorJSQ("Bad IP", w, r, user, isJs)
		}
		err = addModLog("lock", tid, "topic", ipaddress, user.ID)
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}
		err = topic.CreateActionReply("lock", ipaddress, user)
		if err != nil {
			return InternalErrorJSQ(err, w, r, isJs)
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func routeUnlockTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unlock/submit/"):])
	if err != nil {
		return PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreError("The topic you tried to unlock doesn't exist.", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		return NoPermissions(w, r, user)
	}

	err = topic.Unlock()
	if err != nil {
		return InternalError(err, w, r)
	}

	// ! - Can we use user.LastIP here? It might be racey, if another thread mutates it... We need to fix this.
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("unlock", tid, "topic", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unlock", ipaddress, user)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func routeReplyEditSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		return PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	// Get the Reply ID..
	var tid int
	err = stmts.getReplyTID.QueryRow(rid).Scan(&tid)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(tid).Scan(&fid)
	if err == ErrNoRows {
		return PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	content := html.EscapeString(preparseMessage(r.PostFormValue("edit_item")))
	_, err = stmts.editReply.Exec(content, parseMessage(content, fid, "forums"), rid)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Refactor this
// TODO: Disable stat updates in posts handled by plugin_guilds
func routeReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		return PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := rstore.Get(rid)
	if err == ErrNoRows {
		return PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		return PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	//log.Printf("Reply #%d was deleted by User #%d", rid, user.ID)
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
			return InternalErrorJSQ(err, w, r, isJs)
		}
	} else if err != ErrNoRows {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalErrorJSQ("Bad IP", w, r, user, isJs)
	}
	err = addModLog("delete", reply.ParentID, "reply", ipaddress, user.ID)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}
	return nil
}

func routeProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		return LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	// Get the Reply ID..
	var uid int
	err = stmts.getUserReplyUID.QueryRow(rid).Scan(&uid)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != uid && !user.Perms.EditReply {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	content := html.EscapeString(preparseMessage(r.PostFormValue("edit_item")))
	_, err = stmts.editProfileReply.Exec(content, parseMessage(content, 0, ""), rid)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(uid)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routeProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/delete/submit/"):])
	if err != nil {
		return LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	var uid int
	err = stmts.getUserReplyUID.QueryRow(rid).Scan(&uid)
	if err == ErrNoRows {
		return LocalErrorJSQ("The reply you tried to delete doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != uid && !user.Perms.DeleteReply {
		return NoPermissionsJSQ(w, r, user, isJs)
	}

	_, err = stmts.deleteProfileReply.Exec(rid)
	if err != nil {
		return InternalErrorJSQ(err, w, r, isJs)
	}
	//log.Printf("The profile post '%d' was deleted by User #%d", rid, user.ID)

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routeIps(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewIPs {
		return NoPermissions(w, r, user)
	}

	var ip = r.FormValue("ip")
	var uid int
	var reqUserList = make(map[int]bool)

	rows, err := stmts.findUsersByIPUsers.Query(ip)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			return InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	rows2, err := stmts.findUsersByIPTopics.Query(ip)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows2.Close()

	for rows2.Next() {
		err := rows2.Scan(&uid)
		if err != nil {
			return InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows2.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	rows3, err := stmts.findUsersByIPReplies.Query(ip)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows3.Close()

	for rows3.Next() {
		err := rows3.Scan(&uid)
		if err != nil {
			return InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows3.Err()
	if err != nil {
		return InternalError(err, w, r)
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
		return InternalError(err, w, r)
	}

	pi := IPSearchPage{"IP Search", user, headerVars, userList, ip}
	if preRenderHooks["pre_render_ips"] != nil {
		if runPreRenderHook("pre_render_ips", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "ip-search.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeBanSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Perms.BanUsers {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/ban/submit/"):])
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}
	/*if uid == -2 {
		return LocalError("Stop trying to ban Merlin! Ban admin! Bad! No!",w,r,user)
	}*/

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		return LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Is there a difference between IsMod and IsSuperMod? Should we delete the redundant one?
	if targetUser.IsSuperAdmin || targetUser.IsAdmin || targetUser.IsMod {
		return LocalError("You may not ban another staff member.", w, r, user)
	}
	if uid == user.ID {
		return LocalError("Why are you trying to ban yourself? Stop that.", w, r, user)
	}

	if targetUser.IsBanned {
		return LocalError("The user you're trying to unban is already banned.", w, r, user)
	}

	durationDays, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		return LocalError("You can only use whole numbers for the number of days", w, r, user)
	}

	durationWeeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		return LocalError("You can only use whole numbers for the number of weeks", w, r, user)
	}

	durationMonths, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		return LocalError("You can only use whole numbers for the number of months", w, r, user)
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
		return LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("ban", uid, "user", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeUnban(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Perms.BanUsers {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/unban/"):])
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		return LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if !targetUser.IsBanned {
		return LocalError("The user you're trying to unban isn't banned.", w, r, user)
	}

	err = targetUser.Unban()
	if err == ErrNoTempGroup {
		return LocalError("The user you're trying to unban is not banned", w, r, user)
	} else if err == ErrNoRows {
		return LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("unban", uid, "user", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeActivate(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Perms.ActivateUsers {
		return NoPermissions(w, r, user)
	}
	if r.FormValue("session") != user.Session {
		return SecurityError(w, r, user)
	}

	uid, err := strconv.Atoi(r.URL.Path[len("/users/activate/"):])
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}

	targetUser, err := users.Get(uid)
	if err == ErrNoRows {
		return LocalError("The account you're trying to activate no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if targetUser.Active {
		return LocalError("The account you're trying to activate has already been activated.", w, r, user)
	}
	err = targetUser.Activate()
	if err != nil {
		return InternalError(err, w, r)
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return LocalError("Bad IP", w, r, user)
	}
	err = addModLog("activate", targetUser.ID, "user", ipaddress, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}
