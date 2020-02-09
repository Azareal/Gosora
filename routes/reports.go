package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
)

func ReportSubmit(w http.ResponseWriter, r *http.Request, user c.User, sItemID string) c.RouteError {
	headerLite, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"

	itemID, err := strconv.Atoi(sItemID)
	if err != nil {
		return c.LocalError("Bad ID", w, r, user)
	}
	itemType := r.FormValue("type")

	// TODO: Localise these titles and bodies
	var title, content string
	switch itemType {
	case "reply":
		reply, err := c.Rstore.Get(itemID)
		if err == sql.ErrNoRows {
			return c.LocalError("We were unable to find the reported post", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}

		topic, err := c.Topics.Get(reply.ParentID)
		if err == sql.ErrNoRows {
			return c.LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	case "user-reply":
		userReply, err := c.Prstore.Get(itemID)
		if err == sql.ErrNoRows {
			return c.LocalError("We weren't able to find the reported post", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}

		profileOwner, err := c.Users.Get(userReply.ParentID)
		if err == sql.ErrNoRows {
			return c.LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		title = "Profile: " + profileOwner.Name
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	case "topic":
		topic, err := c.Topics.Get(itemID)
		if err == sql.ErrNoRows {
			return c.NotFound(w, r, nil)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		title = "Topic: " + topic.Title
		content = topic.Content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	case "convo-reply":
		post := &c.ConversationPost{ID: itemID}
		err := post.Fetch()
		if err == sql.ErrNoRows {
			return c.NotFound(w, r, nil)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}

		post, err = c.ConvoPostProcess.OnLoad(post)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		user, err := c.Users.Get(post.CreatedBy)
		if err != nil {
			return c.InternalError(err, w, r)
		}

		title = "Convo Post: " + user.Name
		content = post.Body + "\n\nOriginal Post: #cpid-" + strconv.Itoa(itemID)
	default:
		_, hasHook := headerLite.Hooks.VhookNeedHook("report_preassign", &itemID, &itemType)
		if hasHook {
			return nil
		}

		// Don't try to guess the type
		return c.LocalError("Unknown type", w, r, user)
	}

	// TODO: Repost attachments in the reports forum, so that the mods can see them
	_, err = c.Reports.Create(title, content, &user, itemType, itemID)
	if err == c.ErrAlreadyReported {
		return c.LocalError("Someone has already reported this!", w, r, user)
	}
	counters.PostCounter.Bump()

	if !js {
		// TODO: Redirect back to where we came from
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
