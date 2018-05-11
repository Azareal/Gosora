package main

import (
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"./common"
	"./common/counters"
)

// TODO: Refactor this
func routeLikeTopicSubmit(w http.ResponseWriter, r *http.Request, user common.User, stid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")
	tid, err := strconv.Atoi(stid)
	if err != nil {
		return common.PreErrorJSQ("Topic IDs can only ever be numbers.", w, r, isJs)
	}

	topic, err := common.Topics.Get(tid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("The requested topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}
	if topic.CreatedBy == user.ID {
		return common.LocalErrorJSQ("You can't like your own topics", w, r, user, isJs)
	}

	_, err = common.Users.Get(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		return common.LocalErrorJSQ("The target user doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	score := 1
	err = topic.Like(score, user.ID)
	//log.Print("likeErr: ", err)
	if err == common.ErrAlreadyLiked {
		return common.LocalErrorJSQ("You already liked this", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.AddActivityAndNotifyTarget(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

func routeReplyLikeSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == ErrNoRows {
		return common.PreErrorJSQ("You can't like something which doesn't exist!", w, r, isJs)
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
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}
	if reply.CreatedBy == user.ID {
		return common.LocalErrorJSQ("You can't like your own replies", w, r, user, isJs)
	}

	_, err = common.Users.Get(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		return common.LocalErrorJSQ("The target user doesn't exist", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = reply.Like(user.ID)
	if err == common.ErrAlreadyLiked {
		return common.LocalErrorJSQ("You've already liked this!", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.AddActivityAndNotifyTarget(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
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

	profileOwner, err := common.Users.Get(uid)
	if err == ErrNoRows {
		return common.LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	content := common.PreparseMessage(r.PostFormValue("reply-content"))
	// TODO: Fully parse the post and store it in the parsed column
	_, err = common.Prstore.Create(profileOwner.ID, content, user.ID, user.LastIP)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.AddActivityAndNotifyTarget(user.ID, profileOwner.ID, "reply", "user", profileOwner.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	counters.PostCounter.Bump()
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
			return common.NotFound(w, r, nil)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		_, hasHook := common.RunVhookNeedHook("report_preassign", &itemID, &itemType)
		if hasHook {
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
	counters.PostCounter.Bump()

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
	return nil
}

func routeAccountEditCriticalSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeAccountEditAvatar(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
		return nil
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
	headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_avatar_updated"))

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
		return nil
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
	if common.RunPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
		return nil
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

	headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_username_updated"))
	pi := common.Page{"Edit Username", user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
		return nil
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
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_mail_disabled"))
	}

	pi := common.Page{"Email Manager", user, headerVars, emailList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
		return nil
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
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_mail_disabled"))
	}
	headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_mail_verify_success"))
	pi := common.Page{"Email Manager", user, headerVars, emailList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
		return nil
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
		return common.NotFound(w, r, nil)
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
