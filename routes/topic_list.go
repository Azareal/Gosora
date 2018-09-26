package routes

import (
	"log"
	"net/http"
	"strconv"

	"../common"
)

func TopicList(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("topics")
	header.Zone = "topics"
	header.MetaDesc = header.Settings["meta_desc"].(string)

	group, err := common.Groups.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
		return common.LocalError("Something weird happened", w, r, user)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))

	// TODO: Pass a struct back rather than passing back so many variables
	var topicList []*common.TopicsRow
	var forumList []common.Forum
	var paginator common.Paginator
	if user.IsSuperAdmin {
		topicList, forumList, paginator, err = common.TopicList.GetList(page, "")
	} else {
		topicList, forumList, paginator, err = common.TopicList.GetListByGroup(group, page, "")
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// ! Need an inline error not a page level error
	if len(topicList) == 0 {
		return common.NotFound(w, r, header)
	}

	pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{"lastupdated", false}, paginator}
	if common.RunPreRenderHook("pre_render_topic_list", w, r, &user, &pi) {
		return nil
	}
	err = common.RunThemeTemplate(header.Theme.Name, "topics", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func TopicListMostViewed(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("topics")
	header.Zone = "topics"
	header.MetaDesc = header.Settings["meta_desc"].(string)

	group, err := common.Groups.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
		return common.LocalError("Something weird happened", w, r, user)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))

	// TODO: Pass a struct back rather than passing back so many variables
	var topicList []*common.TopicsRow
	var forumList []common.Forum
	var paginator common.Paginator
	if user.IsSuperAdmin {
		topicList, forumList, paginator, err = common.TopicList.GetList(page, "most-viewed")
	} else {
		topicList, forumList, paginator, err = common.TopicList.GetListByGroup(group, page, "most-viewed")
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// ! Need an inline error not a page level error
	if len(topicList) == 0 {
		return common.NotFound(w, r, header)
	}

	pi := common.TopicListPage{header, topicList, forumList, common.Config.DefaultForum, common.TopicListSort{"mostviewed", false}, paginator}
	if common.RunPreRenderHook("pre_render_topic_list", w, r, &user, &pi) {
		return nil
	}
	err = common.RunThemeTemplate(header.Theme.Name, "topics", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
