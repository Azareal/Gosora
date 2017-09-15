package main

import (
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ? - Should we add a new permission or permission zone (like per-forum permissions) specifically for profile comment creation
// ? - Should we allow banned users to make reports? How should we handle report abuse?
// TODO: Add a permission to stop certain users from using custom avatars
// ? - Log username changes and put restrictions on this?

func routeTopicCreate(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			PreError("The provided ForumID is not a valid number.", w, r)
			return
		}
	}

	headerVars, ok := ForumUserCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		NoPermissions(w, r, user)
		return
	}

	BuildWidgets("create_topic", nil, headerVars, r)

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strictmode bool
	if vhooks["topic_create_pre_loop"] != nil {
		runVhook("topic_create_pre_loop", w, r, fid, &headerVars, &user, &strictmode)
	}

	// TODO: Re-add support for plugin_socialgroups
	var forumList []Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		group, err := gstore.Get(user.Group)
		if err != nil {
			LocalError("Something weird happened behind the scenes", w, r, user)
			log.Print("Group #" + strconv.Itoa(user.Group) + " doesn't exist, but it's set on User #" + strconv.Itoa(user.ID))
			return
		}
		canSee = group.CanSee
	}

	// TODO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TODO: Surely, there's a better way of doing this. I've added it in for now to support plugin_socialgroups, but we really need to clean this up
		if strictmode && ffid != fid {
			continue
		}

		// Do a bulk forum fetch, just in case it's the SqlForumStore?
		forum := fstore.DirtyGet(ffid)
		fcopy := *forum
		if hooks["topic_create_frow_assign"] != nil {
			// TODO: Add the skip feature to all the other row based hooks?
			if runHook("topic_create_frow_assign", &fcopy).(bool) {
				continue
			}
		}
		forumList = append(forumList, fcopy)
	}

	ctpage := CreateTopicPage{"Create Topic", user, headerVars, forumList, fid}
	if preRenderHooks["pre_render_create_topic"] != nil {
		if runPreRenderHook("pre_render_create_topic", w, r, &user, &ctpage) {
			return
		}
	}

	template_create_topic_handle(ctpage, w)
}

// POST functions. Authorised users only.
func routeTopicCreateSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		PreError("The provided ForumID is not a valid number.", w, r)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		NoPermissions(w, r, user)
		return
	}

	topicName := html.EscapeString(r.PostFormValue("topic-name"))
	content := html.EscapeString(preparseMessage(r.PostFormValue("topic-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	wcount := wordCount(content)
	res, err := create_topic_stmt.Exec(fid, topicName, content, parseMessage(content), user.ID, ipaddress, wcount, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	err = fstore.IncrementTopicCount(fid)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_subscription_stmt.Exec(user.ID, lastID, "topic")
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
	err = user.increasePostStats(wcount, true)
	if err != nil {
		InternalError(err, w)
		return
	}

	err = fstore.UpdateLastTopic(topicName, int(lastID), user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), fid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
	}
}

func routeCreateReply(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		PreError("Failed to convert the Topic ID", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("Couldn't find the parent topic", w, r)
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
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		NoPermissions(w, r, user)
		return
	}

	content := preparseMessage(html.EscapeString(r.PostFormValue("reply-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	wcount := wordCount(content)
	_, err = create_reply_stmt.Exec(tid, content, parseMessage(content), ipaddress, wcount, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_replies_to_topic_stmt.Exec(1, user.ID, tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = fstore.UpdateLastTopic(topic.Title, tid, user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), topic.ParentID)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_watchers_stmt.Exec(lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enableWebsockets {
		go notifyWatchers(lastID)
	}

	// Reload the topic...
	err = topics.Reload(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The destination no longer exists", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	err = user.increasePostStats(wcount, false)
	if err != nil {
		InternalError(err, w)
		return
	}
}

func routeLikeTopic(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/like/submit/"):])
	if err != nil {
		PreError("Topic IDs can only ever be numbers.", w, r)
		return
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		PreError("The requested topic doesn't exist.", w, r)
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
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w, r, user)
		return
	}

	if topic.CreatedBy == user.ID {
		LocalError("You can't like your own topics", w, r, user)
		return
	}

	err = has_liked_topic_stmt.QueryRow(user.ID, tid).Scan(&tid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!", w, r, user)
		return
	}

	_, err = users.Get(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		LocalError("The target user doesn't exist", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score, tid, "topics", user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_likes_to_topic_stmt.Exec(1, tid)
	if err != nil {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_one_stmt.Exec(topic.CreatedBy, lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(topic.CreatedBy, int(lastID), "like", "topic", user.ID, topic.CreatedBy, tid)

	// Reload the topic...
	err = topics.Reload(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The liked topic no longer exists", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func routeReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/like/submit/"):])
	if err != nil {
		PreError("The provided Reply ID is not a valid number.", w, r)
		return
	}

	reply, err := getReply(rid)
	if err == ErrNoRows {
		PreError("You can't like something which doesn't exist!", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	var fid int
	err = get_topic_fid_stmt.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		PreError("The parent topic doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumUserCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w, r, user)
		return
	}

	if reply.CreatedBy == user.ID {
		LocalError("You can't like your own replies", w, r, user)
		return
	}

	err = has_liked_reply_stmt.QueryRow(user.ID, rid).Scan(&rid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!", w, r, user)
		return
	}

	_, err = users.Get(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		LocalError("The target user doesn't exist", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score, rid, "replies", user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_likes_to_reply_stmt.Exec(1, rid)
	if err != nil {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_one_stmt.Exec(reply.CreatedBy, lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(reply.CreatedBy, int(lastID), "like", "post", user.ID, reply.CreatedBy, rid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
}

func routeProfileReplyCreate(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		LocalError("Invalid UID", w, r, user)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	_, err = create_profile_reply_stmt.Exec(uid, html.EscapeString(preparseMessage(r.PostFormValue("reply-content"))), parseMessage(html.EscapeString(preparseMessage(r.PostFormValue("reply-content")))), user.ID, ipaddress)
	if err != nil {
		InternalError(err, w)
		return
	}

	var userName string
	err = get_user_name_stmt.QueryRow(uid).Scan(&userName)
	if err == ErrNoRows {
		LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
}

func routeReportSubmit(w http.ResponseWriter, r *http.Request, user User, sitemID string) {
	if !user.Loggedin {
		LoginRequired(w, r, user)
		return
	}
	if user.IsBanned {
		Banned(w, r, user)
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

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		LocalError("Bad ID", w, r, user)
		return
	}

	itemType := r.FormValue("type")

	var fid = 1
	var title, content string
	if itemType == "reply" {
		reply, err := getReply(itemID)
		if err == ErrNoRows {
			LocalError("We were unable to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		topic, err := topics.Get(reply.ParentID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	} else if itemType == "user-reply" {
		userReply, err := getUserReply(itemID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		err = get_user_name_stmt.QueryRow(userReply.ParentID).Scan(&title)
		if err == ErrNoRows {
			LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}
		title = "Profile: " + title
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	} else if itemType == "topic" {
		err = get_topic_basic_stmt.QueryRow(itemID).Scan(&title, &content)
		if err == ErrNoRows {
			NotFound(w, r)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		if vhooks["report_preassign"] != nil {
			runVhookNoreturn("report_preassign", &itemID, &itemType)
			return
		}
		// Don't try to guess the type
		LocalError("Unknown type", w, r, user)
		return
	}

	var count int
	rows, err := report_exists_stmt.Query(itemType + "_" + strconv.Itoa(itemID))
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			InternalError(err, w)
			return
		}
	}
	if count != 0 {
		LocalError("Someone has already reported this!", w, r, user)
		return
	}

	res, err := create_report_stmt.Exec(title, content, parseMessage(content), user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		InternalError(err, w)
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_topics_to_forum_stmt.Exec(1, fid)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = fstore.UpdateLastTopic(title, int(lastID), user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), fid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
}

func routeAccountOwnEditCritical(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit.html", pi)
}

func routeAccountOwnEditCriticalSubmit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	err = get_password_stmt.QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == ErrNoRows {
		LocalError("Your account no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	err = CheckPassword(realPassword, currentPassword, salt)
	if err == ErrMismatchedHashAndPassword {
		LocalError("That's not the correct password.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}
	if newPassword != confirmPassword {
		LocalError("The two passwords don't match.", w, r, user)
		return
	}
	SetPassword(user.ID, newPassword)

	// Log the user out as a safety precaution
	auth.ForceLogout(user.ID)

	headerVars.NoticeList = append(headerVars.NoticeList, "Your password was successfully updated")
	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit.html", pi)
}

func routeAccountOwnEditAvatar(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
}

func routeAccountOwnEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user User) {
	if r.ContentLength > int64(config.MaxRequestSize) {
		http.Error(w, "Request too large", http.StatusExpectationFailed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	err := r.ParseMultipartForm(int64(config.MaxRequestSize))
	if err != nil {
		LocalError("Upload failed", w, r, user)
		return
	}

	var filename string
	var ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open()
			if err != nil {
				LocalError("Upload failed", w, r, user)
				return
			}
			defer infile.Close()

			// We don't want multiple files
			if filename != "" {
				if filename != hdr.Filename {
					os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
					LocalError("You may only upload one avatar", w, r, user)
					return
				}
			} else {
				filename = hdr.Filename
			}

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					LocalError("Bad file", w, r, user)
					return
				}
				ext = extarr[len(extarr)-1]

				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					LocalError("Bad file extension", w, r, user)
					return
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
			if err != nil {
				LocalError("Upload failed [File Creation Failed]", w, r, user)
				return
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile)
			if err != nil {
				LocalError("Upload failed [Copy Failed]", w, r, user)
				return
			}
		}
	}

	_, err = set_avatar_stmt.Exec("."+ext, strconv.Itoa(user.ID))
	if err != nil {
		InternalError(err, w)
		return
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	err = users.Reload(user.ID)
	if err != nil {
		LocalError("This user no longer exists!", w, r, user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList, "Your avatar was successfully updated")
	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
}

func routeAccountOwnEditUsername(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	pi := Page{"Edit Username", user, headerVars, tList, user.Name}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
}

func routeAccountOwnEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	newUsername := html.EscapeString(r.PostFormValue("account-new-username"))
	_, err = set_username_stmt.Exec(newUsername, strconv.Itoa(user.ID))
	if err != nil {
		LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
		return
	}

	// TODO: Use the reloaded data instead for the name?
	user.Name = newUsername
	err = users.Reload(user.ID)
	if err != nil {
		LocalError("Your account doesn't exist!", w, r, user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList, "Your username was successfully updated")
	pi := Page{"Edit Username", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
}

func routeAccountOwnEditEmail(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	email := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := get_emails_by_user_stmt.Query(user.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			log.Fatal(err)
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user.
	// This also applies when the admin switches site.EnableEmails on after having it off for a while.
	if len(emailList) == 0 {
		email.Email = user.Email
		email.Validated = false
		email.Primary = true
		emailList = append(emailList, email)
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	pi := Page{"Email Manager", user, headerVars, emailList, nil}
	if preRenderHooks["pre_render_account_own_edit_email"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
}

func routeAccountOwnEditEmailTokenSubmit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := UserCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	token := r.URL.Path[len("/user/edit/token/"):]

	email := Email{UserID: user.ID}
	targetEmail := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := get_emails_by_user_stmt.Query(user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			InternalError(err, w)
			return
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		if email.Token == token {
			targetEmail = email
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	if len(emailList) == 0 {
		LocalError("A verification email was never sent for you!", w, r, user)
		return
	}
	if targetEmail.Token == "" {
		LocalError("That's not a valid token!", w, r, user)
		return
	}

	_, err = verify_email_stmt.Exec(user.Email)
	if err != nil {
		InternalError(err, w)
		return
	}

	// If Email Activation is on, then activate the account while we're here
	if headerVars.Settings["activation_type"] == 2 {
		_, err = activate_user_stmt.Exec(user.ID)
		if err != nil {
			InternalError(err, w)
			return
		}
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	headerVars.NoticeList = append(headerVars.NoticeList, "Your email was successfully verified")
	pi := Page{"Email Manager", user, headerVars, emailList, nil}
	if preRenderHooks["pre_render_account_own_edit_email"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
}

func routeLogout(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Loggedin {
		LocalError("You can't logout without logging in first.", w, r, user)
		return
	}
	auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
