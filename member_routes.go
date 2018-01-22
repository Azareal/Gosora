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

	"./common"
)

func routeCreateReplySubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		return common.PreError("Failed to convert the Topic ID", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreError("Couldn't find the parent topic", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return common.NoPermissions(w, r, user)
	}

	// Handle the file attachments
	// TODO: Stop duplicating this code
	if user.Perms.UploadFiles {
		files, ok := r.MultipartForm.File["upload_files"]
		if ok {
			if len(files) > 5 {
				return common.LocalError("You can't attach more than five files", w, r, user)
			}

			for _, file := range files {
				log.Print("file.Filename ", file.Filename)
				extarr := strings.Split(file.Filename, ".")
				if len(extarr) < 2 {
					return common.LocalError("Bad file", w, r, user)
				}
				ext := extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return common.LocalError("Bad file extension", w, r, user)
				}
				ext = strings.ToLower(reg.ReplaceAllString(ext, ""))
				if !common.AllowedFileExts.Contains(ext) {
					return common.LocalError("You're not allowed to upload files with this extension", w, r, user)
				}

				infile, err := file.Open()
				if err != nil {
					return common.LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				hasher := sha256.New()
				_, err = io.Copy(hasher, infile)
				if err != nil {
					return common.LocalError("Upload failed [Hashing Failed]", w, r, user)
				}
				infile.Close()

				checksum := hex.EncodeToString(hasher.Sum(nil))
				filename := checksum + "." + ext
				outfile, err := os.Create("." + "/attachs/" + filename)
				if err != nil {
					return common.LocalError("Upload failed [File Creation Failed]", w, r, user)
				}
				defer outfile.Close()

				infile, err = file.Open()
				if err != nil {
					return common.LocalError("Upload failed", w, r, user)
				}
				defer infile.Close()

				_, err = io.Copy(outfile, infile)
				if err != nil {
					return common.LocalError("Upload failed [Copy Failed]", w, r, user)
				}

				err = common.Attachments.Add(topic.ParentID, "forums", tid, "replies", user.ID, filename)
				if err != nil {
					return common.InternalError(err, w, r)
				}
			}
		}
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and put that in the parsed column
	_, err = common.Rstore.Create(topic, content, user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.Forums.UpdateLastTopic(tid, user.ID, topic.ParentID)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	_, err = stmts.notifyWatchers.Exec(lastID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enableWebsockets {
		go notifyWatchers(lastID)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)

	wcount := common.WordCount(content)
	err = user.IncreasePostStats(wcount, false)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	common.PostCounter.Bump()
	return nil
}

// TODO: Refactor this
func routeLikeTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreError("Topic IDs can only ever be numbers.", w, r)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreError("The requested topic doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return common.NoPermissions(w, r, user)
	}
	if topic.CreatedBy == user.ID {
		return common.LocalError("You can't like your own topics", w, r, user)
	}

	_, err = common.Users.Get(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		return common.LocalError("The target user doesn't exist", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	score := 1
	err = topic.Like(score, user.ID)
	if err == common.ErrAlreadyLiked {
		return common.LocalError("You already liked this", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	_, err = stmts.notifyOne.Exec(topic.CreatedBy, lastID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(topic.CreatedBy, int(lastID), "like", "topic", user.ID, topic.CreatedBy, tid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	return nil
}

func routeReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreError("The provided Reply ID is not a valid number.", w, r)
	}

	reply, err := common.Rstore.Get(rid)
	if err == ErrNoRows {
		return common.PreError("You can't like something which doesn't exist!", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var fid int
	err = stmts.getTopicFID.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		return common.PreError("The parent topic doesn't exist.", w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return common.NoPermissions(w, r, user)
	}
	if reply.CreatedBy == user.ID {
		return common.LocalError("You can't like your own replies", w, r, user)
	}

	_, err = common.Users.Get(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		return common.LocalError("The target user doesn't exist", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = reply.Like(user.ID)
	if err == common.ErrAlreadyLiked {
		return common.LocalError("You've already liked this!", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	res, err := stmts.addActivity.Exec(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	_, err = stmts.notifyOne.Exec(reply.CreatedBy, lastID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(reply.CreatedBy, int(lastID), "like", "post", user.ID, reply.CreatedBy, rid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	return nil
}

func routeProfileReplyCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if !user.Perms.ViewTopic || !user.Perms.CreateReply {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		return common.LocalError("Invalid UID", w, r, user)
	}
	if !common.Users.Exists(uid) {
		return common.LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and store it in the parsed column
	_, err = common.Prstore.Create(uid, content, user.ID, user.LastIP)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	common.PostCounter.Bump()
	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func routeReportSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalError("Bad ID", w, r, user)
	}
	itemType := r.FormValue("type")

	var fid = 1
	var title, content string
	if itemType == "reply" {
		reply, err := common.Rstore.Get(itemID)
		if err == ErrNoRows {
			return common.LocalError("We were unable to find the reported post", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		topic, err := common.Topics.Get(reply.ParentID)
		if err == ErrNoRows {
			return common.LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	} else if itemType == "user-reply" {
		userReply, err := common.Prstore.Get(itemID)
		if err == ErrNoRows {
			return common.LocalError("We weren't able to find the reported post", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		err = stmts.getUserName.QueryRow(userReply.ParentID).Scan(&title)
		if err == ErrNoRows {
			return common.LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		title = "Profile: " + title
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	} else if itemType == "topic" {
		err = stmts.getTopicBasic.QueryRow(itemID).Scan(&title, &content)
		if err == ErrNoRows {
			return common.NotFound(w, r)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		if common.Vhooks["report_preassign"] != nil {
			common.RunVhookNoreturn("report_preassign", &itemID, &itemType)
			return nil
		}
		// Don't try to guess the type
		return common.LocalError("Unknown type", w, r, user)
	}

	var count int
	err = stmts.reportExists.QueryRow(itemType + "_" + strconv.Itoa(itemID)).Scan(&count)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	if count != 0 {
		return common.LocalError("Someone has already reported this!", w, r, user)
	}

	// TODO: Repost attachments in the reports forum, so that the mods can see them
	// ? - Can we do this via the TopicStore? Should we do a ReportStore?
	res, err := stmts.createReport.Exec(title, content, common.ParseMessage(content, 0, ""), user.ID, user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		return common.InternalError(err, w, r)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.Forums.AddTopic(int(lastID), user.ID, fid)
	if err != nil && err != ErrNoRows {
		return common.InternalError(err, w, r)
	}
	common.PostCounter.Bump()

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
	return nil
}

func routeAccountEditCritical(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Password", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditCriticalSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	err := stmts.getPassword.QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == ErrNoRows {
		return common.LocalError("Your account no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.CheckPassword(realPassword, currentPassword, salt)
	if err == common.ErrMismatchedHashAndPassword {
		return common.LocalError("That's not the correct password.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	if newPassword != confirmPassword {
		return common.LocalError("The two passwords don't match.", w, r, user)
	}
	common.SetPassword(user.ID, newPassword)

	// Log the user out as a safety precaution
	common.Auth.ForceLogout(user.ID)

	headerVars.NoticeList = append(headerVars.NoticeList, "Your password was successfully updated")
	pi := common.Page{"Edit Password", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditAvatar(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit_avatar.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var filename, ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open()
			if err != nil {
				return common.LocalError("Upload failed", w, r, user)
			}
			defer infile.Close()

			// We don't want multiple files
			// TODO: Check the length of r.MultipartForm.File and error rather than doing this x.x
			if filename != "" {
				if filename != hdr.Filename {
					os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
					return common.LocalError("You may only upload one avatar", w, r, user)
				}
			} else {
				filename = hdr.Filename
			}

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					return common.LocalError("Bad file", w, r, user)
				}
				ext = extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return common.LocalError("Bad file extension", w, r, user)
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
			if err != nil {
				return common.LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile)
			if err != nil {
				return common.LocalError("Upload failed [Copy Failed]", w, r, user)
			}
		}
	}

	err := user.ChangeAvatar("." + ext)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	headerVars.NoticeList = append(headerVars.NoticeList, "Your avatar was successfully updated")

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_avatar.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditUsername(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Username", user, headerVars, tList, user.Name}
	if common.PreRenderHooks["pre_render_account_own_edit_username"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit_username.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	newUsername := html.EscapeString(strings.Replace(r.PostFormValue("account-new-username"), "\n", "", -1))
	err := user.ChangeName(newUsername)
	if err != nil {
		return common.LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}
	user.Name = newUsername

	headerVars.NoticeList = append(headerVars.NoticeList, "Your username was successfully updated")
	pi := common.Page{"Edit Username", user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_username"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_username.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeAccountEditEmail(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	email := common.Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := stmts.getEmailsByUser.Query(user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user.
	// This also applies when the admin switches site.EnableEmails on after having it off for a while.
	if len(emailList) == 0 {
		email.Email = user.Email
		email.Validated = false
		email.Primary = true
		emailList = append(emailList, email)
	}
	if !common.Site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}

	pi := common.Page{"Email Manager", user, headerVars, emailList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_email"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_email.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Do a session check on this?
func routeAccountEditEmailTokenSubmit(w http.ResponseWriter, r *http.Request, user common.User, token string) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	email := common.Email{UserID: user.ID}
	targetEmail := common.Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := stmts.getEmailsByUser.Query(user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			return common.InternalError(err, w, r)
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
		return common.InternalError(err, w, r)
	}

	if len(emailList) == 0 {
		return common.LocalError("A verification email was never sent for you!", w, r, user)
	}
	if targetEmail.Token == "" {
		return common.LocalError("That's not a valid token!", w, r, user)
	}

	_, err = stmts.verifyEmail.Exec(user.Email)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// If Email Activation is on, then activate the account while we're here
	if headerVars.Settings["activation_type"] == 2 {
		err = user.Activate()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	if !common.Site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	headerVars.NoticeList = append(headerVars.NoticeList, "Your email was successfully verified")
	pi := common.Page{"Email Manager", user, headerVars, emailList, nil}
	if common.PreRenderHooks["pre_render_account_own_edit_email"] != nil {
		if common.RunPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_email.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeLogout(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if !user.Loggedin {
		return common.LocalError("You can't logout without logging in first.", w, r, user)
	}
	common.Auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeShowAttachment(w http.ResponseWriter, r *http.Request, user common.User, filename string) common.RouteError {
	filename = common.Stripslashes(filename)
	var ext = filepath.Ext("./attachs/" + filename)
	//log.Print("ext ", ext)
	//log.Print("filename ", filename)
	if !common.AllowedFileExts.Contains(strings.TrimPrefix(ext, ".")) {
		return common.LocalError("Bad extension", w, r, user)
	}

	sectionID, err := strconv.Atoi(r.FormValue("sectionID"))
	if err != nil {
		return common.LocalError("The sectionID is not an integer", w, r, user)
	}
	var sectionTable = r.FormValue("sectionType")

	var originTable string
	var originID, uploadedBy int
	err = stmts.getAttachment.QueryRow(filename, sectionID, sectionTable).Scan(&sectionID, &sectionTable, &originID, &originTable, &uploadedBy, &filename)
	if err == ErrNoRows {
		return common.NotFound(w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if sectionTable == "forums" {
		_, ferr := common.SimpleForumUserCheck(w, r, &user, sectionID)
		if ferr != nil {
			return ferr
		}
		if !user.Perms.ViewTopic {
			return common.NoPermissions(w, r, user)
		}
	} else {
		return common.LocalError("Unknown section", w, r, user)
	}

	if originTable != "topics" && originTable != "replies" {
		return common.LocalError("Unknown origin", w, r, user)
	}

	// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
	http.ServeFile(w, r, "./attachs/"+filename)
	return nil
}
