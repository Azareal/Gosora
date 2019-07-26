package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func Convos(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	accountEditHead("convos", w, r, &user, header)
	ccount := c.Convos.GetUserCount(user.ID)
	page, _ := strconv.Atoi(r.FormValue("page"))
	offset, page, lastPage := c.PageOffset(ccount, page, c.Config.ItemsPerPage)
	pageList := c.Paginate(page, lastPage, 5)

	convos, err := c.Convos.GetUser(user.ID, offset)
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
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user)
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

	posts, err := convo.Posts(offset)
	// TODO: Report a better error for no posts
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.Account{header, "dashboard", "convo", c.ConvoViewPage{header, posts, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, header, pi)
}

func ConvosCreate(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	accountEditHead("create_convo", w, r, &user, header)
	recpName := ""
	pi := c.Account{header, "dashboard", "create_convo", c.ConvoCreatePage{header, recpName}}
	return renderTemplate("account", w, r, header, pi)
}

func ConvosCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
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

/*type ConversationPost struct {
	ID   int
	CID int
	Body string
	Post string // aes, ''
}*/

func ConvosDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, scid string) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	cid, err := strconv.Atoi(scid)
	if err != nil {
		return c.LocalError(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user)
	}

	err = c.Convos.Delete(cid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/convos/", http.StatusSeeOther)
	return nil
}

func ConvosCreateReplySubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	http.Redirect(w, r, "/user/convo/id", http.StatusSeeOther)
	return nil
}

func ConvosDeleteReplySubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	http.Redirect(w, r, "/user/convo/id", http.StatusSeeOther)
	return nil
}