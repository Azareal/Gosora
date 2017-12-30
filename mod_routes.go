package main

import (
	//"log"
	//"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"./common"
)

// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Make sure this route is member only
func routeEditTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	err := r.ParseForm()
	if err != nil {
		return common.PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("js") == "1")

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		return common.PreErrorJSQ("The provided TopicID is not a valid number.", w, r, isJs)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("The topic you tried to edit doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditTopic {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	topicName := r.PostFormValue("topic_name")
	topicContent := common.PreparseMessage(r.PostFormValue("topic_content"))
	// TODO: Fully parse the post and store it in the parsed column
	err = topic.Update(topicName, topicContent)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.Forums.UpdateLastTopic(topic.ID, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
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
// TODO: Make sure this route is member only
func routeDeleteTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if common.ReqIsJson(r) {
		if r.Body == nil {
			return common.PreErrorJS("No request body", w, r)
		}
		//log.Print("r.Body: ", r.Body)
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			//log.Print("parse err: ", err)
			return common.PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
		if err != nil {
			return common.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return common.LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := common.Topics.Get(tid)
		if err == ErrNoRows {
			return common.PreErrorJSQ("The topic you tried to delete doesn't exist.", w, r, isJs)
		} else if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.DeleteTopic {
			return common.NoPermissionsJSQ(w, r, user, isJs)
		}

		// We might be able to handle this err better
		err = topic.Delete()
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		err = common.ModLogs.Create("delete", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// ? - We might need to add soft-delete before we can do an action reply for this
		/*_, err = stmts.createActionReply.Exec(tid,"delete",ipaddress,user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err,w,r,isJs)
		}*/

		log.Printf("Topic #%d was deleted by common.User #%d", tid, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeStickTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreError("The topic you tried to pin doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Stick()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("stick", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("stick", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeUnstickTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreError("The topic you tried to unpin doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.PinTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Unstick()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unstick", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unstick", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeLockTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Move this to some sort of middleware
	var tids []int
	var isJs = false
	if common.ReqIsJson(r) {
		if r.Body == nil {
			return common.PreErrorJS("No request body", w, r)
		}
		err := json.NewDecoder(r.Body).Decode(&tids)
		if err != nil {
			return common.PreErrorJS("We weren't able to parse your data", w, r)
		}
		isJs = true
	} else {
		tid, err := strconv.Atoi(r.URL.Path[len("/topic/lock/submit/"):])
		if err != nil {
			return common.PreError("The provided TopicID is not a valid number.", w, r)
		}
		tids = append(tids, tid)
	}
	if len(tids) == 0 {
		return common.LocalErrorJSQ("You haven't provided any IDs", w, r, user, isJs)
	}

	for _, tid := range tids {
		topic, err := common.Topics.Get(tid)
		if err == ErrNoRows {
			return common.PreErrorJSQ("The topic you tried to lock doesn't exist.", w, r, isJs)
		} else if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		// TODO: Add hooks to make use of headerLite
		_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
			return common.NoPermissionsJSQ(w, r, user, isJs)
		}

		err = topic.Lock()
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}

		err = common.ModLogs.Create("lock", tid, "topic", user.LastIP, user.ID)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
		err = topic.CreateActionReply("lock", user.LastIP, user)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
	}

	if len(tids) == 1 {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tids[0]), http.StatusSeeOther)
	}
	return nil
}

func routeUnlockTopic(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unlock/submit/"):])
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreError("The topic you tried to unlock doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CloseTopic {
		return common.NoPermissions(w, r, user)
	}

	err = topic.Unlock()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unlock", tid, "topic", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	err = topic.CreateActionReply("unlock", user.LastIP, user)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func routeReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	err := r.ParseForm()
	if err != nil {
		return common.PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	// Get the Reply ID..
	var tid int
	err = stmts.getReplyTID.QueryRow(rid).Scan(&tid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(tid).Scan(&fid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	content := common.PreparseMessage(r.PostFormValue("edit_item"))
	_, err = stmts.editReply.Exec(content, common.ParseMessage(content, fid, "forums"), rid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
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
func routeReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	err := r.ParseForm()
	if err != nil {
		return common.PreError("Bad Form", w, r)
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	//log.Printf("Reply #%d was deleted by common.User #%d", rid, user.ID)
	if !isJs {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	replyCreator, err := common.Users.Get(reply.CreatedBy)
	if err == nil {
		wcount := common.WordCount(reply.Content)
		err = replyCreator.DecreasePostStats(wcount, false)
		if err != nil {
			return common.InternalErrorJSQ(err, w, r, isJs)
		}
	} else if err != ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.ModLogs.Create("delete", reply.ParentID, "reply", user.LastIP, user.ID)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return nil
}

func routeProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	err := r.ParseForm()
	if err != nil {
		return common.LocalError("Bad Form", w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/edit/submit/"):])
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	// Get the Reply ID..
	var uid int
	err = stmts.getUserReplyUID.QueryRow(rid).Scan(&uid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != uid && !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	content := common.PreparseMessage(r.PostFormValue("edit_item"))
	_, err = stmts.editProfileReply.Exec(content, common.ParseMessage(content, 0, ""), rid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(uid)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routeProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	err := r.ParseForm()
	if err != nil {
		return common.LocalError("Bad Form", w, r, user)
	}
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(r.URL.Path[len("/profile/reply/delete/submit/"):])
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	var uid int
	err = stmts.getUserReplyUID.QueryRow(rid).Scan(&uid)
	if err == ErrNoRows {
		return common.LocalErrorJSQ("The reply you tried to delete doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != uid && !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	_, err = stmts.deleteProfileReply.Exec(rid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//log.Printf("The profile post '%d' was deleted by common.User #%d", rid, user.ID)

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func routeIps(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewIPs {
		return common.NoPermissions(w, r, user)
	}

	var ip = r.FormValue("ip")
	var uid int
	var reqUserList = make(map[int]bool)

	rows, err := stmts.findUsersByIPUsers.Query(ip)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	rows2, err := stmts.findUsersByIPTopics.Query(ip)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows2.Close()

	for rows2.Next() {
		err := rows2.Scan(&uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows2.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	rows3, err := stmts.findUsersByIPReplies.Query(ip)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows3.Close()

	for rows3.Next() {
		err := rows3.Scan(&uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		reqUserList[uid] = true
	}
	err = rows3.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := common.Users.BulkGetMap(idSlice)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.IPSearchPage{common.GetTitlePhrase("ip-search"), user, headerVars, userList, ip}
	if common.PreRenderHooks["pre_render_ips"] != nil {
		if common.RunPreRenderHook("pre_render_ips", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "ip-search.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeBanSubmit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.BanUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}
	if uid == -2 {
		return common.LocalError("Why don't you like Merlin?", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Is there a difference between IsMod and IsSuperMod? Should we delete the redundant one?
	if targetUser.IsMod {
		return common.LocalError("You may not ban another staff member.", w, r, user)
	}
	if uid == user.ID {
		return common.LocalError("Why are you trying to ban yourself? Stop that.", w, r, user)
	}
	if targetUser.IsBanned {
		return common.LocalError("The user you're trying to unban is already banned.", w, r, user)
	}

	durationDays, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of days", w, r, user)
	}

	durationWeeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of weeks", w, r, user)
	}

	durationMonths, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of months", w, r, user)
	}

	var duration time.Duration
	if durationDays > 1 && durationWeeks > 1 && durationMonths > 1 {
		duration, _ = time.ParseDuration("0")
	} else {
		var seconds int
		seconds += durationDays * common.Day
		seconds += durationWeeks * common.Week
		seconds += durationMonths * common.Month
		duration, _ = time.ParseDuration(strconv.Itoa(seconds) + "s")
	}

	err = targetUser.Ban(duration, user.ID)
	if err == ErrNoRows {
		return common.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("ban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeUnban(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.BanUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if !targetUser.IsBanned {
		return common.LocalError("The user you're trying to unban isn't banned.", w, r, user)
	}

	err = targetUser.Unban()
	if err == common.ErrNoTempGroup {
		return common.LocalError("The user you're trying to unban is not banned", w, r, user)
	} else if err == ErrNoRows {
		return common.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeActivate(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.ActivateUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The account you're trying to activate no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.Active {
		return common.LocalError("The account you're trying to activate has already been activated.", w, r, user)
	}
	err = targetUser.Activate()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("activate", targetUser.ID, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}
