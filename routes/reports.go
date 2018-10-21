package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"../common"
	"../common/counters"
)

func ReportSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	headerLite, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("isJs") == "1")

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalError("Bad ID", w, r, user)
	}
	itemType := r.FormValue("type")

	// TODO: Localise these titles and bodies
	var title, content string
	if itemType == "reply" {
		reply, err := common.Rstore.Get(itemID)
		if err == sql.ErrNoRows {
			return common.LocalError("We were unable to find the reported post", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		topic, err := common.Topics.Get(reply.ParentID)
		if err == sql.ErrNoRows {
			return common.LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	} else if itemType == "user-reply" {
		userReply, err := common.Prstore.Get(itemID)
		if err == sql.ErrNoRows {
			return common.LocalError("We weren't able to find the reported post", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}

		profileOwner, err := common.Users.Get(userReply.ParentID)
		if err == sql.ErrNoRows {
			return common.LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		title = "Profile: " + profileOwner.Name
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	} else if itemType == "topic" {
		topic, err := common.Topics.Get(itemID)
		if err == sql.ErrNoRows {
			return common.NotFound(w, r, nil)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
		title = "Topic: " + topic.Title
		content = topic.Content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		_, hasHook := headerLite.Hooks.VhookNeedHook("report_preassign", &itemID, &itemType)
		if hasHook {
			return nil
		}

		// Don't try to guess the type
		return common.LocalError("Unknown type", w, r, user)
	}

	// TODO: Repost attachments in the reports forum, so that the mods can see them
	_, err = common.Reports.Create(title, content, &user, itemType, itemID)
	if err == common.ErrAlreadyReported {
		return common.LocalError("Someone has already reported this!", w, r, user)
	}
	counters.PostCounter.Bump()

	if !isJs {
		// TODO: Redirect back to where we came from
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
