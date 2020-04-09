package routes

import (
	"database/sql"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	p "github.com/Azareal/Gosora/common/phrases"
)

// TODO: Retire this in favour of an alias for /topics/?
func ViewForum(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header, sfid string) c.RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))
	_, fid, err := ParseSEOURL(sfid)
	if err != nil {
		return c.SimpleError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, h)
	}

	ferr := c.ForumUserCheck(h, w, r, u, fid)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ViewTopic {
		return c.NoPermissions(w, r, u)
	}
	h.Path = "/forums/"

	// TODO: Fix this double-check
	forum, err := c.Forums.Get(fid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, h)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	h.Title = forum.Name
	h.OGDesc = forum.Desc

	topicList, pagi, err := c.TopicList.GetListByForum(forum, page, 0)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	h.Zone = "view_forum"
	h.ZoneID = forum.ID

	// TODO: Reduce the amount of boilerplate here
	if r.FormValue("js") == "1" {
		outBytes, err := wsTopicList(topicList, pagi.LastPage).MarshalJSON()
		if err != nil {
			return c.InternalError(err, w, r)
		}
		w.Write(outBytes)
		return nil
	}

	//pageList := c.Paginate(page, lastPage, 5)
	pi := c.ForumPage{h, topicList, forum, pagi}
	tmpl := forum.Tmpl
	if tmpl == "" {
		ferr = renderTemplate("forum", w, r, h, pi)
	} else {
		tmpl = "forum_" + tmpl
		err = renderTemplate3(tmpl, tmpl, w, r, h, pi)
		if err != nil {
			ferr = renderTemplate("forum", w, r, h, pi)
		}
	}
	counters.ForumViewCounter.Bump(forum.ID)
	return ferr
}
