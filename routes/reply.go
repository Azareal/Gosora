package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	"../common"
)

// TODO: Disable stat updates in posts handled by plugin_guilds
// TODO: Update the stats after edits so that we don't under or over decrement stats during deletes
func ReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := reply.Topic()
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic || !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.SetPost(r.PostFormValue("edit_item"))
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Refactor this
// TODO: Disable stat updates in posts handled by plugin_guilds
func ReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.PreErrorJSQ("The provided Reply ID is not a valid number.", w, r, isJs)
	}

	reply, err := common.Rstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The reply you tried to delete doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	topic, err := common.Topics.Get(reply.ParentID)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The parent topic doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := common.SimpleForumUserCheck(w, r, &user, topic.ParentID)
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
		http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
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
	} else if err != sql.ErrNoRows {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	err = common.ModLogs.Create("delete", reply.ParentID, "reply", user.LastIP, user.ID)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return nil
}

func ProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("js") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	// ? Does the admin understand that this group perm affects this?
	if user.ID != creator.ID && !user.Perms.EditReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.SetBody(r.PostFormValue("edit_item"))
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if !isJs {
		http.Redirect(w, r, "/user/"+strconv.Itoa(creator.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func ProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, srid string) common.RouteError {
	isJs := (r.PostFormValue("isJs") == "1")

	rid, err := strconv.Atoi(srid)
	if err != nil {
		return common.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, isJs)
	}

	reply, err := common.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return common.PreErrorJSQ("The target reply doesn't exist.", w, r, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	creator, err := common.Users.Get(reply.CreatedBy)
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	if user.ID != creator.ID && !user.Perms.DeleteReply {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	err = reply.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//log.Printf("The profile post '%d' was deleted by common.User #%d", reply.ID, user.ID)

	if !isJs {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(creator.ID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}
