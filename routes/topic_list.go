package routes

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// TODO: Implement search
func TopicList(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
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
	}

	// TODO: Pass a struct back rather than passing back so many variables
	var topicList []*common.TopicsRow
	var forumList []common.Forum
	var paginator common.Paginator
	if user.IsSuperAdmin {
		topicList, forumList, paginator, err = common.TopicList.GetList(page, "", fids)
	} else {
		topicList, forumList, paginator, err = common.TopicList.GetListByGroup(group, page, "", fids)
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

	header.Title = phrases.GetTitlePhrase("topics")
	header.Zone = "topics"
	header.Path = "/topics/"
	header.MetaDesc = header.Settings["meta_desc"].(string)
	if len(fids) == 1 {
		forum, err := common.Forums.Get(fids[0])
		if err != nil {
			return common.LocalError("Invalid fid forum", w, r, user)
		}
		header.Title = forum.Name
		header.ZoneID = forum.ID
	}

	pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{"lastupdated", false}, paginator}
	return renderTemplate("topics", w, r, header, pi)
}

func wsTopicList(topicList []*common.TopicsRow, lastPage int) *common.WsTopicList {
	wsTopicList := make([]*common.WsTopicsRow, len(topicList))
	for i, topicRow := range topicList {
		wsTopicList[i] = topicRow.WebSockets()
	}
	return &common.WsTopicList{wsTopicList, lastPage}
}

func TopicListMostViewed(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
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

	// TODO: Pass a struct back rather than passing back so many variables
	var topicList []*common.TopicsRow
	var forumList []common.Forum
	var paginator common.Paginator
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

	//MarshalJSON() ([]byte, error)
	// TODO: Reduce the amount of boilerplate here
	if r.FormValue("js") == "1" {
		outBytes, err := wsTopicList(topicList, paginator.LastPage).MarshalJSON()
		if err != nil {
			return common.InternalError(err, w, r)
		}
		w.Write(outBytes)
		return nil
	}

	pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{"mostviewed", false}, paginator}
	return renderTemplate("topics", w, r, header, pi)
}
