package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
)

func ProfileReplyCreateSubmit(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	if !user.Perms.CreateProfileReply {
		return c.NoPermissions(w, r, user)
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		return c.LocalError("Invalid UID", w, r, user)
	}

	profileOwner, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	blocked, err := c.UserBlocks.IsBlockedBy(profileOwner.ID, user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	// Supermods can bypass blocks so they can tell people off when they do something stupid or have to convey important information
	if blocked && !user.IsSuperMod {
		return c.LocalError("You don't have permission to send messages to one of these users.", w, r, user)
	}

	content := c.PreparseMessage(r.PostFormValue("content"))
	if len(content) == 0 {
		return c.LocalError("You can't make a blank post", w, r, user)
	}
	// TODO: Fully parse the post and store it in the parsed column
	prid, err := c.Prstore.Create(profileOwner.ID, content, user.ID, user.GetIP())
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// ! Be careful about leaking per-route permission state with user ptr
	alert := c.Alert{ActorID: user.ID, TargetUserID: profileOwner.ID, Event: "reply", ElementType: "user", ElementID: profileOwner.ID, Actor: user, Extra: strconv.Itoa(prid)}
	err = c.AddActivityAndNotifyTarget(alert)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	counters.PostCounter.Bump()
	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func ProfileReplyEditSubmit(w http.ResponseWriter, r *http.Request, user *c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, js)
	}

	reply, err := c.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The target reply doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	creator, err := c.Users.Get(reply.CreatedBy)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	if !user.Perms.CreateProfileReply {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	// ? Does the admin understand that this group perm affects this?
	if user.ID != creator.ID && !user.Perms.EditReply {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	// TODO: Stop blocked users from modifying profile replies?

	err = reply.SetBody(r.PostFormValue("edit_item"))
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	if !js {
		http.Redirect(w, r, "/user/"+strconv.Itoa(creator.ID)+"#reply-"+strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}

func ProfileReplyDeleteSubmit(w http.ResponseWriter, r *http.Request, user *c.User, srid string) c.RouteError {
	js := r.PostFormValue("js") == "1"
	rid, err := strconv.Atoi(srid)
	if err != nil {
		return c.LocalErrorJSQ("The provided Reply ID is not a valid number.", w, r, user, js)
	}

	reply, err := c.Prstore.Get(rid)
	if err == sql.ErrNoRows {
		return c.PreErrorJSQ("The target reply doesn't exist.", w, r, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	creator, err := c.Users.Get(reply.CreatedBy)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	if user.ID != creator.ID && !user.Perms.DeleteReply {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	err = reply.Delete()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	//log.Printf("The profile post '%d' was deleted by c.User #%d", reply.ID, user.ID)

	if !js {
		//http.Redirect(w,r, "/user/" + strconv.Itoa(creator.ID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}

	err = c.ModLogs.Create("delete", reply.ParentID, "profile-reply", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	return nil
}
