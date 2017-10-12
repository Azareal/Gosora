package main

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	if fid == 0 {
		fid = config.DefaultForum
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
		if forum.Name != "" && forum.Active {
			fcopy := forum.Copy()
			if hooks["topic_create_frow_assign"] != nil {
				// TODO: Add the skip feature to all the other row based hooks?
				if runHook("topic_create_frow_assign", &fcopy).(bool) {
					continue
				}
			}
			forumList = append(forumList, fcopy)
		}
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
	// TODO: Reduce this to 1MB for attachments for each file?
	if r.ContentLength > int64(config.MaxRequestSize) {
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		CustomError("Your attachments are too big. Your files need to be smaller than "+strconv.Itoa(int(size))+unit+".", http.StatusExpectationFailed, "Error", w, r, user)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		LocalError("Unable to parse the form", w, r, user)
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

	tid, err := topics.Create(fid, topicName, content, user.ID, ipaddress)
	if err != nil {
		switch err {
		case ErrNoRows:
			LocalError("Something went wrong, perhaps the forum got deleted?", w, r, user)
		case ErrNoTitle:
			LocalError("This topic doesn't have a title", w, r, user)
		case ErrNoBody:
			LocalError("This topic doesn't have a body", w, r, user)
		default:
			InternalError(err, w)
		}
		return
	}

	_, err = addSubscriptionStmt.Exec(user.ID, tid, "topic")
	if err != nil {
		InternalError(err, w)
		return
	}

	err = user.increasePostStats(wordCount(content), true)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				LocalError("You can't attach more than five files", w, r, user)
				return
			}

			for _, file := range files {
				log.Print("file.Filename ", file.Filename)
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					LocalError("Bad file", w, r, user)
					return
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					LocalError("Bad file extension", w, r, user)
					return
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !allowedFileExts.Contains(ext) {
					LocalError("You're not allowed to upload files with this extension", w, r, user)
					return
				}

				infile, err := file.Open()
				if err != nil {
					LocalError("Upload failed", w, r, user)
					return
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					LocalError("Upload failed [Hashing Failed]", w, r, user)
					return
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					LocalError("Upload failed [File Creation Failed]", w, r, user)
					return
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					LocalError("Upload failed", w, r, user)
					return
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					LocalError("Upload failed [Copy Failed]", w, r, user)
					return
				}

				_, err = addAttachmentStmt.Exec(fid, "forums", tid, "topics", user.ID, filename)
				if err != nil {
					InternalError(err, w)
					return
				}
			}
		}
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func routeCreateReply(w http.ResponseWriter, r *http.Request, user User) {
	// TODO: Reduce this to 1MB for attachments for each file?
	if r.ContentLength > int64(config.MaxRequestSize) {
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		CustomError("Your attachments are too big. Your files need to be smaller than "+strconv.Itoa(int(size))+unit+".", http.StatusExpectationFailed, "Error", w, r, user)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		LocalError("Unable to parse the form", w, r, user)
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

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				LocalError("You can't attach more than five files", w, r, user)
				return
			}

			for _, file := range files {
				log.Print("file.Filename ", file.Filename)
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					LocalError("Bad file", w, r, user)
					return
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					LocalError("Bad file extension", w, r, user)
					return
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !allowedFileExts.Contains(ext) {
					LocalError("You're not allowed to upload files with this extension", w, r, user)
					return
				}

				infile, err := file.Open()
				if err != nil {
					LocalError("Upload failed", w, r, user)
					return
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					LocalError("Upload failed [Hashing Failed]", w, r, user)
					return
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					LocalError("Upload failed [File Creation Failed]", w, r, user)
					return
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					LocalError("Upload failed", w, r, user)
					return
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					LocalError("Upload failed [Copy Failed]", w, r, user)
					return
				}

				_, err = addAttachmentStmt.Exec(topic.ParentID, "forums", tid, "replies", user.ID, filename)
				if err != nil {
					InternalError(err, w)
					return
				}
			}
		}
	}

	content := preparseMessage(html.EscapeString(r.PostFormValue("reply-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	_, err = rstore.Create(tid, content, ipaddress, topic.ParentID, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	err = fstore.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	res, err := addActivityStmt.Exec(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notifyWatchersStmt.Exec(lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enableWebsockets {
		go notifyWatchers(lastID)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)

	wcount := wordCount(content)
	err = user.increasePostStats(wcount, false)
	if err != nil {
		InternalError(err, w)
		return
	}
}

// TODO: Refactor this
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

	err = hasLikedTopicStmt.QueryRow(user.ID, tid).Scan(&tid)
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
	_, err = createLikeStmt.Exec(score, tid, "topics", user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = addLikesToTopicStmt.Exec(1, tid)
	if err != nil {
		InternalError(err, w)
		return
	}

	res, err := addActivityStmt.Exec(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notifyOneStmt.Exec(topic.CreatedBy, lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(topic.CreatedBy, int(lastID), "like", "topic", user.ID, topic.CreatedBy, tid)

	// Flush the topic out of the cache
	tcache, ok := topics.(TopicCache)
	if ok {
		tcache.CacheRemove(tid)
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

	reply, err := rstore.Get(rid)
	if err == ErrNoRows {
		PreError("You can't like something which doesn't exist!", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	var fid int
	err = getTopicFIDStmt.QueryRow(reply.ParentID).Scan(&fid)
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

	_, err = users.Get(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		LocalError("The target user doesn't exist", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	err = reply.Like(user.ID)
	if err == ErrAlreadyLiked {
		LocalError("You've already liked this!", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	res, err := addActivityStmt.Exec(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notifyOneStmt.Exec(reply.CreatedBy, lastID)
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

	content := html.EscapeString(preparseMessage(r.PostFormValue("reply-content")))
	_, err = createProfileReplyStmt.Exec(uid, content, parseMessage(content, 0, ""), user.ID, ipaddress)
	if err != nil {
		InternalError(err, w)
		return
	}

	var userName string
	err = getUserNameStmt.QueryRow(uid).Scan(&userName)
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
		reply, err := rstore.Get(itemID)
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
		userReply, err := prstore.Get(itemID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		err = getUserNameStmt.QueryRow(userReply.ParentID).Scan(&title)
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
		err = getTopicBasicStmt.QueryRow(itemID).Scan(&title, &content)
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
	rows, err := reportExistsStmt.Query(itemType + "_" + strconv.Itoa(itemID))
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

	// TODO: Repost attachments in the reports forum, so that the mods can see them
	// ? - Can we do this via the TopicStore?
	res, err := createReportStmt.Exec(title, content, parseMessage(content, 0, ""), user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		InternalError(err, w)
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = addTopicsToForumStmt.Exec(1, fid)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = fstore.UpdateLastTopic(int(lastID), user.ID, fid)
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

	err = getPasswordStmt.QueryRow(user.ID).Scan(&realPassword, &salt)
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
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		CustomError("Your avatar's too big. Avatars must be smaller than "+strconv.Itoa(int(size))+unit, http.StatusExpectationFailed, "Error", w, r, user)
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

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		LocalError("Upload failed", w, r, user)
		return
	}

	var filename, ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open()
			if err != nil {
				LocalError("Upload failed", w, r, user)
				return
			}
			defer infile.Close()

			// We don't want multiple files
			// TODO: Check the length of r.MultipartForm.File and error rather than doing this x.x
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

				// TODO: Can we do this without a regex?
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

	err = user.ChangeAvatar("." + ext)
	if err != nil {
		InternalError(err, w)
		return
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext

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
	err = user.ChangeName(newUsername)
	if err != nil {
		LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
		return
	}
	user.Name = newUsername

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
	rows, err := getEmailsByUserStmt.Query(user.ID)
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
	rows, err := getEmailsByUserStmt.Query(user.ID)
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

	_, err = verifyEmailStmt.Exec(user.Email)
	if err != nil {
		InternalError(err, w)
		return
	}

	// If Email Activation is on, then activate the account while we're here
	if headerVars.Settings["activation_type"] == 2 {
		_, err = activateUserStmt.Exec(user.ID)
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

func routeShowAttachment(w http.ResponseWriter, r *http.Request, user User, filename string) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	filename = Stripslashes(filename)
	var ext = filepath.Ext("./attachs/" + filename)
	//log.Print("ext ", ext)
	//log.Print("filename ", filename)
	if !allowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		LocalError("Bad extension", w, r, user)
		return
	}

	sectionID, err := strconv.Atoi(r.FormValue("sectionID"))
	if err != nil {
		LocalError("The sectionID is not an integer", w, r, user)
		return
	}
	var sectionTable = r.FormValue("sectionType")

	var originTable string
	var originID, uploadedBy int
	err = getAttachmentStmt.QueryRow(filename, sectionID, sectionTable).Scan(&sectionID, &sectionTable, &originID, &originTable, &uploadedBy, &filename)
	if err == ErrNoRows {
		NotFound(w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	if sectionTable == "forums" {
		_, ok := SimpleForumUserCheck(w, r, &user, sectionID)
		if !ok {
			return
		}
		if !user.Perms.ViewTopic {
			NoPermissions(w, r, user)
			return
		}
	} else {
		LocalError("Unknown section", w, r, user)
		return
	}

	if originTable != "topics" && originTable != "replies" {
		LocalError("Unknown origin", w, r, user)
		return
	}

	// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
	http.ServeFile(w, r, "./attachs/"+filename)
}
