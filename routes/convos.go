package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	//"log"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Convos(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	accountEditHead("convos", w, r, &user, header)
	header.AddSheet(header.Theme.Name + "/convo.css")
	header.AddNotice("convo_dev")
	ccount := c.Convos.GetUserCount(user.ID)
	page, _ := strconv.Atoi(r.FormValue("page"))
	offset, page, lastPage := c.PageOffset(ccount, page, c.Config.ItemsPerPage)
	pageList := c.Paginate(page, lastPage, 5)

	convos, err := c.Convos.GetUserExtra(user.ID, offset)
	//log.Printf("convos: %+v\n", convos)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.Account{header, "dashboard", "convos", c.ConvoListPage{header, convos, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, header, pi)
}

func Convo(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header, scid string) c.RouteError {
	accountEditHead("convo", w, r, &user, header)
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

	pi := c.Account{header, "dashboard", "convo", c.ConvoViewPage{header, convo, pitems, users, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, header, pi)
}

func ConvosCreate(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	accountEditHead("create_convo", w, r, &user, header)
	header.AddNotice("convo_dev")
	recpName := ""
	pi := c.Account{header, "dashboard", "create_convo", c.ConvoCreatePage{header, recpName}}
	return renderTemplate("account", w, r, header, pi)
}

func ConvosCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.IsBanned {
		return c.NoPermissions(w, r, user)
	}

	recps := c.SanitiseSingleLine(r.PostFormValue("recp"))
	body := c.PreparseMessage(r.PostFormValue("body"))
	rlist := []int{}
	max := 10 // max number of recipients that can be added at once
	for i, recp := range strings.Split(recps, ",") {
		if i >= max {
			break
		}

		u, err := c.Users.GetByName(recp)
		if err == sql.ErrNoRows {
			return c.LocalError("One of the recipients doesn't exist", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}

		rlist = append(rlist, u.ID)
	}

	cid, err := c.Convos.Create(body, user.ID, rlist)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convo/"+strconv.Itoa(cid), http.StatusSeeOther)
	return nil
}

/*func ConvosDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, scid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	err = c.Convos.Delete(cid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convos/", http.StatusSeeOther)
	return nil
}*/

func ConvosCreateReplySubmit(w http.ResponseWriter, r *http.Request, user c.User, scid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.IsBanned {
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

	body := c.PreparseMessage(r.PostFormValue("content"))
	post := &c.ConversationPost{CID: cid, Body: body, CreatedBy: user.ID}
	_, err = post.Create()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convo/"+strconv.Itoa(convo.ID), http.StatusSeeOther)
	return nil
}

func ConvosDeleteReplySubmit(w http.ResponseWriter, r *http.Request, user c.User, scpid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
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

func ConvosEditReplySubmit(w http.ResponseWriter, r *http.Request, user c.User, scpid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	cpid, err := strconv.Atoi(scpid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}
	isJs := (r.PostFormValue("js") == "1")

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

	if !isJs {
		http.Redirect(w, r, "/user/convo/"+strconv.Itoa(post.CID), http.StatusSeeOther)
	} else {
		w.Write(successJSONBytes)
	}
	return nil
}
