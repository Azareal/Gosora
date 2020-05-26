package routes

import (
	"log"
	"net/http"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func ForumList(w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header) c.RouteError {
	/*skip, rerr := h.Hooks.VhookSkippable("route_forum_list_start", w, r, user, h)
	if skip || rerr != nil {
		return rerr
	}*/
	skip, rerr := c.H_route_forum_list_start_hook(h.Hooks, w, r, user, h)
	if skip || rerr != nil {
		return rerr
	}
	h.Title = phrases.GetTitlePhrase("forums")
	h.Zone = "forums"
	h.Path = "/forums/"
	h.MetaDesc = h.Settings["meta_desc"].(string)

	var err error
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = c.Forums.GetAllVisibleIDs()
		if err != nil {
			return c.InternalError(err, w, r)
		}
	} else {
		group, err := c.Groups.Get(user.Group)
		if err != nil {
			log.Printf("Group #%d doesn't exist despite being used by c.User #%d", user.Group, user.ID)
			return c.LocalError("Something weird happened", w, r, user)
		}
		canSee = group.CanSee
	}

	var forumList []c.Forum
	for _, fid := range canSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		f := c.Forums.DirtyGet(fid).Copy()
		if f.ParentID == 0 && f.Name != "" && f.Active {
			if f.LastTopicID != 0 {
				if f.LastTopic.ID != 0 && f.LastReplyer.ID != 0 {
					f.LastTopicTime = c.RelativeTime(f.LastTopic.LastReplyAt)
				}
			}
			//h.Hooks.Hook("forums_frow_assign", &f)
			c.H_forums_frow_assign_hook(h.Hooks, &f)
			forumList = append(forumList, f)
		}
	}

	return renderTemplate("forums", w, r, h, c.ForumsPage{h, forumList})
}
