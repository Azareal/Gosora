package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func wsTopicList(topicList []*c.TopicsRow, lastPage int) *c.WsTopicList {
	wsTopicList := make([]*c.WsTopicsRow, len(topicList))
	for i, topicRow := range topicList {
		wsTopicList[i] = topicRow.WebSockets()
	}
	return &c.WsTopicList{wsTopicList, lastPage, 0}
}

func TopicList(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header) c.RouteError {
	/*skip, rerr := h.Hooks.VhookSkippable("route_topic_list_start", w, r, u, h)
	if skip || rerr != nil {
		return rerr
	}*/
	skip, rerr := c.H_route_topic_list_start_hook(h.Hooks, w, r, u, h)
	if skip || rerr != nil {
		return rerr
	}
	return TopicListCommon(w, r, u, h, "lastupdated", 0)
}

func TopicListMostViewed(w http.ResponseWriter, r *http.Request, u *c.User, h *c.Header) c.RouteError {
	skip, rerr := h.Hooks.VhookSkippable("route_topic_list_mostviewed_start", w, r, u, h)
	if skip || rerr != nil {
		return rerr
	}
	return TopicListCommon(w, r, u, h, "mostviewed", c.TopicListMostViewed)
}

// TODO: Implement search
func TopicListCommon(w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header, torder string, tsorder int) c.RouteError {
	h.Title = phrases.GetTitlePhrase("topics")
	h.Zone = "topics"
	h.Path = "/topics/"
	h.MetaDesc = h.Settings["meta_desc"].(string)

	group, err := c.Groups.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by c.User #%d", user.Group, user.ID)
		return c.LocalError("Something weird happened", w, r, user)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))
	sfids := r.FormValue("fids")
	var fids []int
	if sfids != "" {
		for _, sfid := range strings.Split(sfids, ",") {
			fid, err := strconv.Atoi(sfid)
			if err != nil {
				return c.LocalError("Invalid fid", w, r, user)
			}
			fids = append(fids, fid)
		}
		if len(fids) == 1 {
			forum, err := c.Forums.Get(fids[0])
			if err != nil {
				return c.LocalError("Invalid fid forum", w, r, user)
			}
			h.Title = forum.Name
			h.ZoneID = forum.ID
		}
	}

	// TODO: Allow multiple forums in searches
	// TODO: Simplify this block after initially landing search
	var topicList []*c.TopicsRow
	var forumList []c.Forum
	var pagi c.Paginator
	var canLock, ccanLock, canMove, ccanMove bool
	q := r.FormValue("q")
	if q != "" && c.RepliesSearch != nil {
		var canSee []int
		if user.IsSuperAdmin {
			canSee, err = c.Forums.GetAllVisibleIDs()
			if err != nil {
				return c.InternalError(err, w, r)
			}
		} else {
			canSee = group.CanSee
		}

		var cfids []int
		if len(fids) > 0 {
			inSlice := func(haystack []int, needle int) bool {
				for _, item := range haystack {
					if needle == item {
						return true
					}
				}
				return false
			}
			for _, fid := range fids {
				if inSlice(canSee, fid) {
					f := c.Forums.DirtyGet(fid)
					if f.Name != "" && f.Active && (f.ParentType == "" || f.ParentType == "forum") && f.TopicCount != 0 {
						// TODO: Add a hook here for plugin_guilds?
						cfids = append(cfids, fid)
					}
				}
			}
		} else {
			cfids = canSee
		}

		tids, err := c.RepliesSearch.Query(q, cfids)
		if err != nil && err != sql.ErrNoRows {
			return c.InternalError(err, w, r)
		}
		//log.Printf("tids %+v\n", tids)
		// TODO: Handle the case where there aren't any items...
		// TODO: Add a BulkGet method which returns a slice?
		tMap, err := c.Topics.BulkGetMap(tids)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		// TODO: Cache emptied map across requests with sync pool
		reqUserList := make(map[int]bool)
		for _, t := range tMap {
			reqUserList[t.CreatedBy] = true
			reqUserList[t.LastReplyBy] = true
			topicList = append(topicList, t.TopicsRow())
		}
		//fmt.Printf("reqUserList %+v\n", reqUserList)

		// Convert the user ID map to a slice, then bulk load the users
		idSlice := make([]int, len(reqUserList))
		var i int
		for userID := range reqUserList {
			idSlice[i] = userID
			i++
		}

		// TODO: What if a user is deleted via the Control Panel?
		//fmt.Printf("idSlice %+v\n", idSlice)
		userList, err := c.Users.BulkGetMap(idSlice)
		if err != nil {
			return nil // TODO: Implement this!
		}

		// TODO: De-dupe this logic in common/topic_list.go?
		//var sb strings.Builder
		for _, t := range topicList {
			//c.BuildTopicURLSb(&sb, c.NameToSlug(t.Title), t.ID)
			//t.Link = sb.String()
			//sb.Reset()
			t.Link = c.BuildTopicURL(c.NameToSlug(t.Title), t.ID)
			// TODO: Pass forum to something like t.Forum and use that instead of these two properties? Could be more flexible.
			forum := c.Forums.DirtyGet(t.ParentID)
			t.ForumName = forum.Name
			t.ForumLink = forum.Link

			fp, err := c.FPStore.Get(forum.ID, user.Group)
			if err == c.ErrNoRows {
				fp = c.BlankForumPerms()
			} else if err != nil {
				return c.InternalError(err, w, r)
			}
			if fp.Overrides && !user.IsSuperAdmin {
				ccanLock = fp.CloseTopic
				ccanMove = fp.MoveTopic
			} else {
				ccanLock = user.Perms.CloseTopic
				ccanMove = user.Perms.MoveTopic
			}
			if ccanLock {
				canLock = true
			}
			if ccanMove {
				canMove = true
			}

			// TODO: Create a specialised function with a bit less overhead for getting the last page for a post count
			_, _, lastPage := c.PageOffset(t.PostCount, 1, c.Config.ItemsPerPage)
			t.LastPage = lastPage
			// TODO: Avoid map if either is equal to the current user
			t.Creator = userList[t.CreatedBy]
			t.LastUser = userList[t.LastReplyBy]
		}

		// TODO: Reduce the amount of boilerplate here
		if r.FormValue("js") == "1" {
			outBytes, err := wsTopicList(topicList, pagi.LastPage).MarshalJSON()
			if err != nil {
				return c.InternalError(err, w, r)
			}
			w.Write(outBytes)
			return nil
		}

		h.Title = phrases.GetTitlePhrase("topics_search")
		pi := c.TopicListPage{h, topicList, forumList, c.Config.DefaultForum, c.TopicListSort{torder, false}, canLock, canMove, pagi}
		return renderTemplate("topics", w, r, h, pi)
	}

	// TODO: Pass a struct back rather than passing back so many variables
	if user.IsSuperAdmin {
		topicList, forumList, pagi, err = c.TopicList.GetList(page, tsorder, fids)
		canLock, canMove = true, true
	} else {
		topicList, forumList, pagi, err = c.TopicList.GetListByGroup(group, page, tsorder, fids)
		for _, forum := range forumList {
			fp, err := c.FPStore.Get(forum.ID, user.Group)
			if err == c.ErrNoRows {
				fp = c.BlankForumPerms()
			} else if err != nil {
				return c.InternalError(err, w, r)
			}
			if fp.Overrides {
				ccanLock = fp.CloseTopic
				ccanMove = fp.MoveTopic
			} else {
				ccanLock = user.Perms.CloseTopic
				ccanMove = user.Perms.MoveTopic
			}
			if ccanLock {
				canLock = true
			}
			if ccanMove {
				canMove = true
			}
		}
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Reduce the amount of boilerplate here
	if r.FormValue("js") == "1" {
		outBytes, err := wsTopicList(topicList, pagi.LastPage).MarshalJSON()
		if err != nil {
			return c.InternalError(err, w, r)
		}
		w.Write(outBytes)
		return nil
	}

	pi := c.TopicListPage{h, topicList, forumList, c.Config.DefaultForum, c.TopicListSort{torder, false}, canLock, canMove, pagi}
	if r.FormValue("i") == "1" {
		return renderTemplate("topics_mini", w, r, h, pi)
	}
	return renderTemplate("topics", w, r, h, pi)
}
