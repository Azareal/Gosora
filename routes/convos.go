package routes

import (
	"database/sql"
	"errors"
	"html"
	"net/http"
	"strconv"
	"strings"

	//"log"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Convos(w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header) c.RouteError {
	accountEditHead("convos", w, r, user, h)
	h.AddScript("convo.js")
	h.AddSheet(h.Theme.Name + "/convo.css")
	h.AddNotice("convo_dev")
	ccount := c.Convos.GetUserCount(user.ID)
	page, _ := strconv.Atoi(r.FormValue("page"))
	offset, page, lastPage := c.PageOffset(ccount, page, c.Config.ItemsPerPage)
	pageList := c.Paginate(page, lastPage, 5)

	convos, err := c.Convos.GetUserExtra(user.ID, offset)
	//log.Printf("convos: %+v\n", convos)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	}

	var cRows []c.ConvoListRow
	for _, convo := range convos {
		var parti []*c.User
		notMe := false
		for _, u := range convo.Users {
			if u.ID == user.ID {
				continue
			}
			parti = append(parti, u)
			notMe = true
		}
		if !notMe {
			parti = convo.Users
		}
		cRows = append(cRows, c.ConvoListRow{convo, parti, len(parti) == 1})
	}

	pi := c.Account{h, "dashboard", "convos", c.ConvoListPage{h, cRows, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, h, pi)
}

func Convo(w http.ResponseWriter, r *http.Request, user *c.User, header *c.Header, scid string) c.RouteError {
	accountEditHead("convo", w, r, user, header)
	header.AddSheet(header.Theme.Name + "/convo.css")
	header.AddNotice("convo_dev")
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	convo, err := c.Convos.Get(cid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	pcount := convo.PostsCount()
	if pcount == 0 {
		return c.NotFound(w, r, header)
	}

	page, _ := strconv.Atoi(r.FormValue("page"))
	offset, page, lastPage := c.PageOffset(pcount, page, c.Config.ItemsPerPage)
	pageList := c.Paginate(page, lastPage, 5)

	posts, err := convo.Posts(offset, c.Config.ItemsPerPage)
	// TODO: Report a better error for no posts
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	uids, err := convo.Uids()
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	umap, err := c.Users.BulkGetMap(uids)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	users := make([]*c.User, len(umap))
	i := 0
	for _, user := range umap {
		users[i] = user
		i++
	}

	pitems := make([]c.ConvoViewRow, len(posts))
	for i, post := range posts {
		uuser, ok := umap[post.CreatedBy]
		if !ok {
			return c.InternalError(errors.New("convo post creator not in umap"), w, r)
		}
		canModify := user.ID == post.CreatedBy || user.IsSuperMod
		pitems[i] = c.ConvoViewRow{post, uuser, "", 4, canModify}
	}

	canReply := user.Perms.UseConvos || user.Perms.UseConvosOnlyWithMod
	if !user.Perms.UseConvos && user.Perms.UseConvosOnlyWithMod {
		u, err := c.Users.Get(convo.CreatedBy)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		if !u.IsSuperMod {
			canReply = false
		}
	}

	pi := c.Account{header, "dashboard", "convo", c.ConvoViewPage{header, convo, pitems, users, canReply, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, header, pi)
}

func ConvosCreate(w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header) c.RouteError {
	accountEditHead("create_convo", w, r, user, h)
	if !user.Perms.UseConvos && !user.Perms.UseConvosOnlyWithMod {
		return c.NoPermissions(w, r, user)
	}
	h.AddNotice("convo_dev")
	uid, err := strconv.Atoi(r.FormValue("with"))
	if err != nil {
		return c.LocalError("invalid integer in parameter with", w, r, user)
	}
	u, err := c.Users.Get(uid)
	if err != nil {
		return c.LocalError("Unable to fetch user", w, r, user)
	}
	// TODO: Avoid potential double escape?
	pi := c.Account{h, "dashboard", "create_convo", c.ConvoCreatePage{h, html.EscapeString(u.Name)}}
	return renderTemplate("account", w, r, h, pi)
}

func ConvosCreateSubmit(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.UseConvos && !user.Perms.UseConvosOnlyWithMod {
		return c.NoPermissions(w, r, user)
	}

	sRecps := c.SanitiseSingleLine(r.PostFormValue("recp"))
	body := c.PreparseMessage(r.PostFormValue("body"))
	rlist := []int{}

	// De-dupe recipients
	var recps []string
	unames := make(map[string]struct{})
	for _, recp := range strings.Split(sRecps, ",") {
		recp = strings.TrimSpace(recp)
		_, exists := unames[recp]
		if !exists {
			recps = append(recps, recp)
			unames[recp] = struct{}{}
		}
	}

	max := 10 // max number of recipients that can be added at once
	for i, recp := range recps {
		if i >= max {
			break
		}

		u, err := c.Users.GetByName(recp)
		if err == sql.ErrNoRows {
			return c.LocalError("One of the recipients doesn't exist", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		// TODO: Should we kick them out of existing conversations if they're moved into a group without permission or the permission is revoked from their group? We might want to give them a chance to delete their messages though to avoid privacy headaches here and it may only be temporarily to tackle a specific incident.
		if !u.Perms.UseConvos && !u.Perms.UseConvosOnlyWithMod {
			return c.LocalError("One of the recipients doesn't have permission to use the conversations system", w, r, user)
		}
		if !user.Perms.UseConvos && !u.IsSuperMod && user.Perms.UseConvosOnlyWithMod {
			return c.LocalError("You are only allowed to message global moderators.", w, r, user)
		}
		if !user.IsSuperMod && !u.Perms.UseConvos && u.Perms.UseConvosOnlyWithMod {
			return c.LocalError("One of the recipients doesn't have permission to engage with conversation with you.", w, r, user)
		}
		blocked, err := c.UserBlocks.IsBlockedBy(u.ID, user.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		// Supermods can bypass blocks so they can tell people off when they do something stupid or have to convey important information
		if blocked && !user.IsSuperMod {
			return c.LocalError("You don't have permission to send messages to one of these users.", w, r, user)
		}

		rlist = append(rlist, u.ID)
	}

	cid, err := c.Convos.Create(body, user.ID, rlist)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Don't bother making the subscription if the convo creator is the only recipient?
	for _, uid := range rlist {
		if uid == user.ID {
			continue
		}
		err := c.Subscriptions.Add(uid, cid, "convo")
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	err = c.Subscriptions.Add(user.ID, cid, "convo")
	if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.AddActivityAndNotifyAll(c.Alert{ActorID: user.ID, Event: "create", ElementType: "convo", ElementID: cid, Actor: user})
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convo/"+strconv.Itoa(cid), http.StatusSeeOther)
	return nil
}

/*func ConvosDeleteSubmit(w http.ResponseWriter, r *http.Request, user *c.User, scid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}
	if err := c.Convos.Delete(cid); err != nil {
		return c.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/convos/", http.StatusSeeOther)
	return nil
}*/

func ConvosCreateReplySubmit(w http.ResponseWriter, r *http.Request, user *c.User, scid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.UseConvos && !user.Perms.UseConvosOnlyWithMod {
		return c.NoPermissions(w, r, user)
	}
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	convo, err := c.Convos.Get(cid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	pcount := convo.PostsCount()
	if pcount == 0 {
		return c.NotFound(w, r, nil)
	}
	if !convo.Has(user.ID) {
		return c.LocalError("You are not in this conversation.", w, r, user)
	}
	// TODO: Let the user reply if they're the convo creator in a convo with a mod
	if !user.Perms.UseConvos && user.Perms.UseConvosOnlyWithMod {
		u, err := c.Users.Get(convo.CreatedBy)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		if !u.IsSuperMod {
			return c.LocalError("You're only allowed to talk to global moderators.", w, r, user)
		}
	}

	body := c.PreparseMessage(r.PostFormValue("content"))
	post := &c.ConversationPost{CID: cid, Body: body, CreatedBy: user.ID}
	pid, err := post.Create()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.AddActivityAndNotifyAll(c.Alert{ActorID: user.ID, Event: "reply", ElementType: "convo", ElementID: cid, Actor: user, Extra: strconv.Itoa(pid)})
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convo/"+strconv.Itoa(convo.ID), http.StatusSeeOther)
	return nil
}

func ConvosDeleteReplySubmit(w http.ResponseWriter, r *http.Request, user *c.User, scpid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	cpid, err := strconv.Atoi(scpid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	post := &c.ConversationPost{ID: cpid}
	err = post.Fetch()
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	convo, err := c.Convos.Get(post.CID)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	pcount := convo.PostsCount()
	if pcount == 0 {
		return c.NotFound(w, r, nil)
	}
	if user.ID != post.CreatedBy && !user.IsSuperMod {
		return c.NoPermissions(w, r, user)
	}

	posts, err := convo.Posts(0, c.Config.ItemsPerPage)
	// TODO: Report a better error for no posts
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if post.ID == posts[0].ID {
		err = c.Convos.Delete(convo.ID)
	} else {
		err = post.Delete()
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convo/"+strconv.Itoa(post.CID), http.StatusSeeOther)
	return nil
}

func ConvosEditReplySubmit(w http.ResponseWriter, r *http.Request, user *c.User, scpid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	cpid, err := strconv.Atoi(scpid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}
	if !user.Perms.UseConvos {
		return c.NoPermissions(w, r, user)
	}
	js := r.PostFormValue("js") == "1"

	post := &c.ConversationPost{ID: cpid}
	err = post.Fetch()
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	convo, err := c.Convos.Get(post.CID)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	pcount := convo.PostsCount()
	if pcount == 0 {
		return c.NotFound(w, r, nil)
	}
	if user.ID != post.CreatedBy && !user.IsSuperMod {
		return c.NoPermissions(w, r, user)
	}
	if !convo.Has(user.ID) {
		return c.LocalError("You are not in this conversation.", w, r, user)
	}

	post.Body = c.PreparseMessage(r.PostFormValue("edit_item"))
	err = post.Update()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return actionSuccess(w, r, "/user/convo/"+strconv.Itoa(post.CID), js)
}

func RelationsBlockCreate(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header, spid string) c.RouteError {
	h.Title = p.GetTitlePhrase("create_block")
	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, u)
	}
	puser, err := c.Users.Get(pid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to block doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.Page{h, nil, c.AreYouSure{"/user/block/create/submit/" + strconv.Itoa(puser.ID), p.GetTmplPhrase("create_block_msg")}}
	return renderTemplate("are_you_sure", w, r, h, pi)
}

func RelationsBlockCreateSubmit(w http.ResponseWriter, r *http.Request, u *c.User, spid string) c.RouteError {
	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, u)
	}
	puser, err := c.Users.Get(pid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to block doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if u.ID == puser.ID {
		return c.LocalError("You can't block yourself.", w, r, u)
	}

	err = c.UserBlocks.Add(u.ID, puser.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(puser.ID), http.StatusSeeOther)
	return nil
}

func RelationsBlockRemove(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header, spid string) c.RouteError {
	h.Title = p.GetTitlePhrase("remove_block")
	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, u)
	}
	puser, err := c.Users.Get(pid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to block doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.Page{h, nil, c.AreYouSure{"/user/block/remove/submit/" + strconv.Itoa(puser.ID), p.GetTmplPhrase("remove_block_msg")}}
	return renderTemplate("are_you_sure", w, r, h, pi)
}

func RelationsBlockRemoveSubmit(w http.ResponseWriter, r *http.Request, u *c.User, spid string) c.RouteError {
	pid, err := strconv.Atoi(spid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, u)
	}
	puser, err := c.Users.Get(pid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to unblock doesn't exist.", w, r, u)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.UserBlocks.Remove(u.ID, puser.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(puser.ID), http.StatusSeeOther)
	return nil
}
