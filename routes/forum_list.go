package routes

import (
	"log"
	"net/http"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func ForumList(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	skip, rerr := header.Hooks.VhookSkippable("route_forum_list_start", w, r, &user, header)
	if skip || rerr != nil {
		return rerr
	}
	header.Title = phrases.GetTitlePhrase("forums")
	header.Zone = "forums"
	header.Path = "/forums/"
	header.MetaDesc = header.Settings["meta_desc"].(string)

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
			header.Hooks.Hook("forums_frow_assign", &f)
			forumList = append(forumList, f)
		}
	}

	return renderTemplate("forums", w, r, header, c.ForumsPage{header, forumList})
}
