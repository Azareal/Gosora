package routes

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func wsTopicList(topicList []*common.TopicsRow, lastPage int) *common.WsTopicList {
	wsTopicList := make([]*common.WsTopicsRow, len(topicList))
	for i, topicRow := range topicList {
		wsTopicList[i] = topicRow.WebSockets()
	}
	return &common.WsTopicList{wsTopicList, lastPage}
}

func TopicList(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	return TopicListCommon(w, r, user, header, "lastupdated")
}

func TopicListMostViewed(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	return TopicListCommon(w, r, user, header, "mostviewed")
}

// TODO: Implement search
func TopicListCommon(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header, torder string) common.RouteError {
	header.Title = phrases.GetTitlePhrase("topics")
	header.Zone = "topics"
	header.Path = "/topics/"
	header.MetaDesc = header.Settings["meta_desc"].(string)

	group, err := common.Groups.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
		return common.LocalError("Something weird happened", w, r, user)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))
	sfids := r.FormValue("fids")
	var fids []int
	if sfids != "" {
		for _, sfid := range strings.Split(sfids, ",") {
			fid, err := strconv.Atoi(sfid)
			if err != nil {
				return common.LocalError("Invalid fid", w, r, user)
			}
			fids = append(fids, fid)
		}
		if len(fids) == 1 {
			forum, err := common.Forums.Get(fids[0])
			if err != nil {
				return common.LocalError("Invalid fid forum", w, r, user)
			}
			header.Title = forum.Name
			header.ZoneID = forum.ID
		}
	}

	//(t *Topic) WsTopicsRows() *WsTopicsRow
	// TODO: Allow multiple forums in searches
	// TODO: Simplify this block after initially landing search
	var topicList []*common.TopicsRow
	var forumList []common.Forum
	var paginator common.Paginator
	q := r.FormValue("q")
	if q != "" {
		var canSee []int
		if user.IsSuperAdmin {
			canSee, err = common.Forums.GetAllVisibleIDs()
			if err != nil {
				return common.InternalError(err, w, r)
			}
		} else {
			canSee = group.CanSee
		}

		var cfids []int
		if len(fids) > 0 {
			var inSlice = func(haystack []int, needle int) bool {
				for _, item := range haystack {
					if needle == item {
						return true
					}
				}
				return false
			}
			for _, fid := range fids {
				if inSlice(canSee, fid) {
					forum := common.Forums.DirtyGet(fid)
					if forum.Name != "" && forum.Active && (forum.ParentType == "" || forum.ParentType == "forum") {
						// TODO: Add a hook here for plugin_guilds?
						cfids = append(cfids, fid)
					}
				}
			}
		} else {
			cfids = canSee
		}

		tids, err := common.RepliesSearch.Query(q, cfids)
		if err != nil && err != sql.ErrNoRows {
			return common.InternalError(err, w, r)
		}
		//fmt.Printf("tids %+v\n", tids)
		// TODO: Handle the case where there aren't any items...
		// TODO: Add a BulkGet method which returns a slice?
		tMap, err := common.Topics.BulkGetMap(tids)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		var reqUserList = make(map[int]bool)
		for _, topic := range tMap {
			reqUserList[topic.CreatedBy] = true
			reqUserList[topic.LastReplyBy] = true
			topicList = append(topicList, topic.TopicsRow())
		}
		//fmt.Printf("reqUserList %+v\n", reqUserList)

		// Convert the user ID map to a slice, then bulk load the users
		var idSlice = make([]int, len(reqUserList))
		var i int
		for userID := range reqUserList {
			idSlice[i] = userID
			i++
		}

		// TODO: What if a user is deleted via the Control Panel?
		//fmt.Printf("idSlice %+v\n", idSlice)
		userList, err := common.Users.BulkGetMap(idSlice)
		if err != nil {
			return nil // TODO: Implement this!
		}

		// TODO: De-dupe this logic in common/topic_list.go?
		for _, topic := range topicList {
			topic.Link = common.BuildTopicURL(common.NameToSlug(topic.Title), topic.ID)
			// TODO: Pass forum to something like topic.Forum and use that instead of these two properties? Could be more flexible.
			forum := common.Forums.DirtyGet(topic.ParentID)
			topic.ForumName = forum.Name
			topic.ForumLink = forum.Link

			// TODO: Create a specialised function with a bit less overhead for getting the last page for a post count
			_, _, lastPage := common.PageOffset(topic.PostCount, 1, common.Config.ItemsPerPage)
			topic.LastPage = lastPage
			topic.Creator = userList[topic.CreatedBy]
			topic.LastUser = userList[topic.LastReplyBy]
		}

		// TODO: Reduce the amount of boilerplate here
		if r.FormValue("js") == "1" {
			outBytes, err := wsTopicList(topicList, paginator.LastPage).MarshalJSON()
			if err != nil {
				return common.InternalError(err, w, r)
			}
			w.Write(outBytes)
			return nil
		}

		pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{torder, false}, paginator}
		return renderTemplate("topics", w, r, header, pi)
	}

	// TODO: Pass a struct back rather than passing back so many variables
	if user.IsSuperAdmin {
		topicList, forumList, paginator, err = common.TopicList.GetList(page, "most-viewed", fids)
	} else {
		topicList, forumList, paginator, err = common.TopicList.GetListByGroup(group, page, "most-viewed", fids)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}
	// ! Need an inline error not a page level error
	if len(topicList) == 0 {
		return common.NotFound(w, r, header)
	}

	// TODO: Reduce the amount of boilerplate here
	if r.FormValue("js") == "1" {
		outBytes, err := wsTopicList(topicList, paginator.LastPage).MarshalJSON()
		if err != nil {
			return common.InternalError(err, w, r)
		}
		w.Write(outBytes)
		return nil
	}

	pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{torder, false}, paginator}
	return renderTemplate("topics", w, r, header, pi)
}
