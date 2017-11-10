package main

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"io"
	"log"
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

func routeTopicCreate(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			return LocalError("You didn't provide a valid number for the forum ID.", w, r, user)
		}
	}
	if fid == 0 {
		fid = config.DefaultForum
	}

	headerVars, ferr := ForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return NoPermissions(w, r, user)
	}

	BuildWidgets("create_topic", nil, headerVars, r)

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strictmode bool
	if vhooks["topic_create_pre_loop"] != nil {
		runVhook("topic_create_pre_loop", w, r, fid, &headerVars, &user, &strictmode)
	}

	// TODO: Re-add support for plugin_guilds
	var forumList []Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		group, err := gstore.Get(user.Group)
		if err != nil {
			// TODO: Refactor this
			LocalError("Something weird happened behind the scenes", w, r, user)
			log.Printf("Group #%d doesn't exist, but it's set on User #%d", user.Group, user.ID)
			return nil
		}
		canSee = group.CanSee
	}

	// TODO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TODO: Surely, there's a better way of doing this. I've added it in for now to support plugin_guilds, but we really need to clean this up
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
			return nil
		}
	}

	err = RunThemeTemplate(headerVars.ThemeName, "create-topic", ctpage, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

// POST functions. Authorised users only.
func routeTopicCreateSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Reduce this to 1MB for attachments for each file?
	if r.ContentLength > int64(config.MaxRequestSize) {
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		return CustomError("Your attachments are too big. Your files need to be smaller than "+strconv.Itoa(int(size))+unit+".", http.StatusExpectationFailed, "Error", w, r, user)
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		return LocalError("Unable to parse the form", w, r, user)
	}

	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		return LocalError("The provided ForumID is not a valid number.", w, r, user)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateTopic {
		return NoPermissions(w, r, user)
	}

	topicName := html.EscapeString(r.PostFormValue("topic-name"))
	content := html.EscapeString(preparseMessage(r.PostFormValue("topic-content")))
	tid, err := topics.Create(fid, topicName, content, user.ID, user.LastIP)
	if err != nil {
		switch err {
		case ErrNoRows:
			return LocalError("Something went wrong, perhaps the forum got deleted?", w, r, user)
		case ErrNoTitle:
			return LocalError("This topic doesn't have a title", w, r, user)
		case ErrNoBody:
			return LocalError("This topic doesn't have a body", w, r, user)
		default:
			return InternalError(err, w, r)
		}
	}

	_, err = stmts.addSubscription.Exec(user.ID, tid, "topic")
	if err != nil {
		return InternalError(err, w, r)
	}

	err = user.increasePostStats(wordCount(content), true)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				return LocalError("You can't attach more than five files", w, r, user)
			}

			for _, file := range files {
				if dev.DebugMode {
					log.Print("file.Filename ", file.Filename)
				}
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					return LocalError("Bad file", w, r, user)
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return LocalError("Bad file extension", w, r, user)
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !allowedFileExts.Contains(ext) {
					return LocalError("You're not allowed to upload files with this extension", w, r, user)
				}

				infile, err := file.Open()
				if err != nil {
					return LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					return LocalError("Upload failed [Hashing Failed]", w, r, user)
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					return LocalError("Upload failed [File Creation Failed]", w, r, user)
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					return LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					return LocalError("Upload failed [Copy Failed]", w, r, user)
				}

				_, err = stmts.addAttachment.Exec(fid, "forums", tid, "topics", user.ID, filename)
				if err != nil {
					return InternalError(err, w, r)
				}
			}
		}
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeCreateReply(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Reduce this to 1MB for attachments for each file?
	// TODO: Reuse this code more
	if r.ContentLength > int64(config.MaxRequestSize) {
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		return CustomError("Your attachments are too big. Your files need to be smaller than "+strconv.Itoa(int(size))+unit+".", http.StatusExpectationFailed, "Error", w, r, user)
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		return LocalError("Unable to parse the form", w, r, user)
	}

	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		return PreError("Failed to convert the Topic ID", w, r)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreError("Couldn't find the parent topic", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return NoPermissions(w, r, user)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				return LocalError("You can't attach more than five files", w, r, user)
			}

			for _, file := range files {
				log.Print("file.Filename ", file.Filename)
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					return LocalError("Bad file", w, r, user)
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return LocalError("Bad file extension", w, r, user)
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !allowedFileExts.Contains(ext) {
					return LocalError("You're not allowed to upload files with this extension", w, r, user)
				}

				infile, err := file.Open()
				if err != nil {
					return LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					return LocalError("Upload failed [Hashing Failed]", w, r, user)
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					return LocalError("Upload failed [File Creation Failed]", w, r, user)
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					return LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					return LocalError("Upload failed [Copy Failed]", w, r, user)
				}

				_, err = stmts.addAttachment.Exec(topic.ParentID, "forums", tid, "replies", user.ID, filename)
				if err != nil {
					return InternalError(err, w, r)
				}
			}
		}
	}

	content := preparseMessage(html.EscapeString(r.PostFormValue("reply-content")))
	_, err = rstore.Create(topic, content, user.LastIP, user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}

	err = fstore.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		return InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	_, err = stmts.notifyWatchers.Exec(lastID)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enableWebsockets {
		go notifyWatchers(lastID)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)

	wcount := wordCount(content)
	err = user.increasePostStats(wcount, false)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

// TODO: Refactor this
func routeLikeTopic(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/like/submit/"):])
	if err != nil {
		return PreError("Topic IDs can only ever be numbers.", w, r)
	}

	topic, err := topics.Get(tid)
	if err == ErrNoRows {
		return PreError("The requested topic doesn't exist.", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return NoPermissions(w, r, user)
	}
	if topic.CreatedBy == user.ID {
		return LocalError("You can't like your own topics", w, r, user)
	}

	_, err = users.Get(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		return LocalError("The target user doesn't exist", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	score := 1
	err = topic.Like(score, user.ID)
	if err == ErrAlreadyLiked {
		return LocalError("You already liked this", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		return InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	_, err = stmts.notifyOne.Exec(topic.CreatedBy, lastID)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(topic.CreatedBy, int(lastID), "like", "topic", user.ID, topic.CreatedBy, tid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/like/submit/"):])
	if err != nil {
		return PreError("The provided Reply ID is not a valid number.", w, r)
	}

	reply, err := rstore.Get(rid)
	if err == ErrNoRows {
		return PreError("You can't like something which doesn't exist!", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		return PreError("The parent topic doesn't exist.", w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return NoPermissions(w, r, user)
	}
	if reply.CreatedBy == user.ID {
		return LocalError("You can't like your own replies", w, r, user)
	}

	_, err = users.Get(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		return LocalError("The target user doesn't exist", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	err = reply.Like(user.ID)
	if err == ErrAlreadyLiked {
		return LocalError("You've already liked this!", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		return InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	_, err = stmts.notifyOne.Exec(reply.CreatedBy, lastID)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(reply.CreatedBy, int(lastID), "like", "post", user.ID, reply.CreatedBy, rid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	return nil
}

func routeProfileReplyCreate(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return NoPermissions(w, r, user)
	}

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		return LocalError("Invalid UID", w, r, user)
	}

	content := html.EscapeString(preparseMessage(r.PostFormValue("reply-content")))
	_, err = prstore.Create(uid, content, user.ID, user.LastIP)
	if err != nil {
		return InternalError(err, w, r)
	}

	if !users.Exists(uid) {
		return LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeReportSubmit(w http.ResponseWriter, r *http.Request, user User, sitemID string) RouteError {
	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return LocalError("Bad ID", w, r, user)
	}
	itemType := r.FormValue("type")

	var fid = 1
	var title, content string
	if itemType == "reply" {
		reply, err := rstore.Get(itemID)
		if err == ErrNoRows {
			return LocalError("We were unable to find the reported post", w, r, user)
		} else if err != nil {
			return InternalError(err, w, r)
		}

		topic, err := topics.Get(reply.ParentID)
		if err == ErrNoRows {
			return LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
		} else if err != nil {
			return InternalError(err, w, r)
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	} else if itemType == "user-reply" {
		userReply, err := prstore.Get(itemID)
		if err == ErrNoRows {
			return LocalError("We weren't able to find the reported post", w, r, user)
		} else if err != nil {
			return InternalError(err, w, r)
		}

		err = stmts.getUserName.QueryRow(userReply.ParentID).Scan(&title)
		if err == ErrNoRows {
			return LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
		} else if err != nil {
			return InternalError(err, w, r)
		}
		title = "Profile: " + title
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	} else if itemType == "topic" {
		err = stmts.getTopicBasic.QueryRow(itemID).Scan(&title, &content)
		if err == ErrNoRows {
			return NotFound(w, r)
		} else if err != nil {
			return InternalError(err, w, r)
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		if vhooks["report_preassign"] != nil {
			runVhookNoreturn("report_preassign", &itemID, &itemType)
			return nil
		}
		// Don't try to guess the type
		return LocalError("Unknown type", w, r, user)
	}

	var count int
	err = stmts.reportExists.QueryRow(itemType + "_" + strconv.Itoa(itemID)).Scan(&count)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}
	if count != 0 {
		return LocalError("Someone has already reported this!", w, r, user)
	}

	// TODO: Repost attachments in the reports forum, so that the mods can see them
	// ? - Can we do this via the TopicStore? Should we do a ReportStore?
	res, err := stmts.createReport.Exec(title, content, parseMessage(content, 0, ""), user.ID, user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		return InternalError(err, w, r)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return InternalError(err, w, r)
	}

	err = fstore.AddTopic(int(lastID), user.ID, fid)
	if err != nil && err != ErrNoRows {
		return InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
	return nil
}

func routeAccountEditCritical(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "account-own-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditCriticalSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	err := stmts.getPassword.QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == ErrNoRows {
		return LocalError("Your account no longer exists.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	err = CheckPassword(realPassword, currentPassword, salt)
	if err == ErrMismatchedHashAndPassword {
		return LocalError("That's not the correct password.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}
	if newPassword != confirmPassword {
		return LocalError("The two passwords don't match.", w, r, user)
	}
	SetPassword(user.ID, newPassword)

	// Log the user out as a safety precaution
	auth.ForceLogout(user.ID)

	headerVars.NoticeList = append(headerVars.NoticeList, "Your password was successfully updated")
	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "account-own-edit.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditAvatar(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if r.ContentLength > int64(config.MaxRequestSize) {
		size, unit := convertByteUnit(float64(config.MaxRequestSize))
		return CustomError("Your avatar's too big. Avatars must be smaller than "+strconv.Itoa(int(size))+unit, http.StatusExpectationFailed, "Error", w, r, user)
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	err := r.ParseMultipartForm(int64(megabyte))
	if err != nil {
		return LocalError("Upload failed", w, r, user)
	}

	var filename, ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open()
			if err != nil {
				return LocalError("Upload failed", w, r, user)
			}
			defer infile.Close()

			// We don't want multiple files
			// TODO: Check the length of r.MultipartForm.File and error rather than doing this x.x
			if filename != "" {
				if filename != hdr.Filename {
					os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
					return LocalError("You may only upload one avatar", w, r, user)
				}
			} else {
				filename = hdr.Filename
			}

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					return LocalError("Bad file", w, r, user)
				}
				ext = extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return LocalError("Bad file extension", w, r, user)
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
			if err != nil {
				return LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile)
			if err != nil {
				return LocalError("Upload failed [Copy Failed]", w, r, user)
			}
		}
	}

	err = user.ChangeAvatar("." + ext)
	if err != nil {
		return InternalError(err, w, r)
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext

	headerVars.NoticeList = append(headerVars.NoticeList, "Your avatar was successfully updated")
	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditUsername(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := Page{"Edit Username", user, headerVars, tList, user.Name}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	newUsername := html.EscapeString(r.PostFormValue("account-new-username"))
	err := user.ChangeName(newUsername)
	if err != nil {
		return LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}
	user.Name = newUsername

	headerVars.NoticeList = append(headerVars.NoticeList, "Your username was successfully updated")
	pi := Page{"Edit Username", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditEmail(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	email := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := stmts.getEmailsByUser.Query(user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			return InternalError(err, w, r)
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
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
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

// TODO: Do a session check on this?
func routeAccountEditEmailTokenSubmit(w http.ResponseWriter, r *http.Request, user User, token string) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	email := Email{UserID: user.ID}
	targetEmail := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := stmts.getEmailsByUser.Query(user.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			return InternalError(err, w, r)
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
		return InternalError(err, w, r)
	}

	if len(emailList) == 0 {
		return LocalError("A verification email was never sent for you!", w, r, user)
	}
	if targetEmail.Token == "" {
		return LocalError("That's not a valid token!", w, r, user)
	}

	_, err = stmts.verifyEmail.Exec(user.Email)
	if err != nil {
		return InternalError(err, w, r)
	}

	// If Email Activation is on, then activate the account while we're here
	if headerVars.Settings["activation_type"] == 2 {
		_, err = stmts.activateUser.Exec(user.ID)
		if err != nil {
			return InternalError(err, w, r)
		}
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	headerVars.NoticeList = append(headerVars.NoticeList, "Your email was successfully verified")
	pi := Page{"Email Manager", user, headerVars, emailList, nil}
	if preRenderHooks["pre_render_account_own_edit_email"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return nil
		}
	}
	err = templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeLogout(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if !user.Loggedin {
		return LocalError("You can't logout without logging in first.", w, r, user)
	}
	auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeShowAttachment(w http.ResponseWriter, r *http.Request, user User, filename string) RouteError {
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}

	filename = Stripslashes(filename)
	var ext = filepath.Ext("./attachs/" + filename)
	//log.Print("ext ", ext)
	//log.Print("filename ", filename)
	if !allowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		return LocalError("Bad extension", w, r, user)
	}

	sectionID, err := strconv.Atoi(r.FormValue("sectionID"))
	if err != nil {
		return LocalError("The sectionID is not an integer", w, r, user)
	}
	var sectionTable = r.FormValue("sectionType")

	var originTable string
	var originID, uploadedBy int
	err = stmts.getAttachment.QueryRow(filename, sectionID, sectionTable).Scan(&sectionID, &sectionTable, &originID, &originTable, &uploadedBy, &filename)
	if err == ErrNoRows {
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	if sectionTable == "forums" {
		_, ferr := SimpleForumUserCheck(w, r, &user, sectionID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic {
			return NoPermissions(w, r, user)
		}
	} else {
		return LocalError("Unknown section", w, r, user)
	}

	if originTable != "topics" && originTable != "replies" {
		return LocalError("Unknown origin", w, r, user)
	}

	// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
	http.ServeFile(w, r, "./attachs/"+filename)
	return nil
}
