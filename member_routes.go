package main

import (
	"net/http"
	"strconv"

	"./common"
	"./common/counters"
)

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
	if r.FormValue("verified") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_mail_verify_success"))
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
	http.Redirect(w, r, "/user/edit/email/?verified=1", http.StatusSeeOther)

	return nil
}
