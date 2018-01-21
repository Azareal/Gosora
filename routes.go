/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"bytes"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"./common"
	"./query_gen/lib"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

//var nList []string
var successJSONBytes = []byte(`{"success":"1"}`)
var cacheControlMaxAge = "max-age=" + strconv.Itoa(common.Day) // TODO: Make this a common.Config value

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {
}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	dest := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		dest += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

// Temporary stubs for view tracking
func routeDynamic() {
}
func routeUploads() {
}

// GET functions
func routeStatic(w http.ResponseWriter, r *http.Request) {
	file, ok := common.StaticFiles.Get(r.URL.Path)
	if !ok {
		if common.Dev.DebugMode {
			log.Printf("Failed to find '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h := w.Header()

	// Surely, there's a more efficient way of doing this?
	t, err := time.Parse(http.TimeFormat, h.Get("If-Modified-Since"))
	if err == nil && file.Info.ModTime().Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h.Set("Last-Modified", file.FormattedModTime)
	h.Set("Content-Type", file.Mimetype)
	h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000
	h.Set("Vary", "Accept-Encoding")
	if strings.Contains(h.Get("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Length", strconv.FormatInt(file.GzipLength, 10))
		io.Copy(w, bytes.NewReader(file.GzipData)) // Use w.Write instead?
	} else {
		h.Set("Content-Length", strconv.FormatInt(file.Length, 10)) // Avoid doing a type conversion every time?
		io.Copy(w, bytes.NewReader(file.Data))
	}
	// Other options instead of io.Copy: io.CopyN(), w.Write(), http.ServeContent()
}

func routeOverview(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "overview"

	pi := common.Page{common.GetTitlePhrase("overview"), user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_overview"] != nil {
		if common.RunPreRenderHook("pre_render_overview", w, r, &user, &pi) {
			return nil
		}
	}

	err := common.Templates.ExecuteTemplate(w, "overview.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeCustomPage(w http.ResponseWriter, r *http.Request, user common.User, name string) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// ! Is this safe?
	if common.Templates.Lookup("page_"+name+".html") == nil {
		return common.NotFound(w, r)
	}
	headerVars.Zone = "custom_page"

	pi := common.Page{common.GetTitlePhrase("page"), user, headerVars, tList, nil}
	// TODO: Pass the page name to the pre-render hook?
	if common.PreRenderHooks["pre_render_custom_page"] != nil {
		if common.RunPreRenderHook("pre_render_custom_page", w, r, &user, &pi) {
			return nil
		}
	}

	err := common.Templates.ExecuteTemplate(w, "page_"+name+".html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeTopics(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "topics"
	headerVars.MetaDesc = headerVars.Settings["meta_desc"].(string)

	group, err := common.Groups.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
		return common.LocalError("Something weird happened", w, r, user)
	}

	// TODO: Make CanSee a method on *Group with a canSee field? Have a CanSee method on *User to cover the case of superadmins?
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = common.Forums.GetAllVisibleIDs()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		canSee = group.CanSee
	}

	// We need a list of the visible forums for Quick Topic
	// ? - Would it be useful, if we could post in social groups from /topics/?
	var forumList []common.Forum
	for _, fid := range canSee {
		forum := common.Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active && (forum.ParentType == "" || forum.ParentType == "forum") {
			fcopy := forum.Copy()
			// TODO: Add a hook here for plugin_guilds
			forumList = append(forumList, fcopy)
		}
	}

	// ? - Should we be showing plugin_guilds posts on /topics/?
	argList, qlist := common.ForumListToArgQ(forumList)

	// ! Need an inline error not a page level error
	if qlist == "" {
		return common.NotFound(w, r)
	}

	topicCount, err := common.ArgQToTopicCount(argList, qlist)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))
	offset, page, lastPage := common.PageOffset(topicCount, page, common.Config.ItemsPerPage)

	var topicList []*common.TopicsRow
	stmt, err := qgen.Builder.SimpleSelect("topics", "tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount", "parentID IN("+qlist+")", "sticky DESC, lastReplyAt DESC, createdBy DESC", "?,?")
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer stmt.Close()

	argList = append(argList, offset)
	argList = append(argList, common.Config.ItemsPerPage)

	rows, err := stmt.Query(argList...)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	var reqUserList = make(map[int]bool)
	for rows.Next() {
		topicItem := common.TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		topicItem.Link = common.BuildTopicURL(common.NameToSlug(topicItem.Title), topicItem.ID)

		forum := common.Forums.DirtyGet(topicItem.ParentID)
		topicItem.ForumName = forum.Name
		topicItem.ForumLink = forum.Link

		//topicItem.CreatedAt = common.RelativeTime(topicItem.CreatedAt)
		topicItem.RelativeLastReplyAt = common.RelativeTime(topicItem.LastReplyAt)

		if common.Vhooks["topics_topic_row_assign"] != nil {
			common.RunVhook("topics_topic_row_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := common.Users.BulkGetMap(idSlice)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pageList := common.Paginate(topicCount, common.Config.ItemsPerPage, 5)
	pi := common.TopicsPage{common.GetTitlePhrase("topics"), user, headerVars, topicList, forumList, common.Config.DefaultForum, pageList, page, lastPage}
	if common.PreRenderHooks["pre_render_topic_list"] != nil {
		if common.RunPreRenderHook("pre_render_topic_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.RunThemeTemplate(headerVars.Theme.Name, "topics", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeForum(w http.ResponseWriter, r *http.Request, user common.User, sfid string) common.RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(sfid, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	fid, err := strconv.Atoi(halves[1])
	if err != nil {
		return common.PreError("The provided ForumID is not a valid number.", w, r)
	}

	headerVars, ferr := common.ForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		return common.NoPermissions(w, r, user)
	}

	// TODO: Fix this double-check
	forum, err := common.Forums.Get(fid)
	if err == ErrNoRows {
		return common.NotFound(w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	headerVars.Zone = "view_forum"

	// TODO: Does forum.TopicCount take the deleted items into consideration for guests? We don't have soft-delete yet, only hard-delete
	offset, page, lastPage := common.PageOffset(forum.TopicCount, page, common.Config.ItemsPerPage)

	// TODO: Move this to *Forum
	rows, err := stmts.getForumTopicsOffset.Query(fid, offset, common.Config.ItemsPerPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	// TODO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	var topicList []*common.TopicsRow
	var reqUserList = make(map[int]bool)
	for rows.Next() {
		var topicItem = common.TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		topicItem.Link = common.BuildTopicURL(common.NameToSlug(topicItem.Title), topicItem.ID)
		topicItem.RelativeLastReplyAt = common.RelativeTime(topicItem.LastReplyAt)

		if common.Vhooks["forum_trow_assign"] != nil {
			common.RunVhook("forum_trow_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := common.Users.BulkGetMap(idSlice)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := common.ForumPage{forum.Name, user, headerVars, topicList, forum, page, lastPage}
	if common.PreRenderHooks["pre_render_forum"] != nil {
		if common.RunPreRenderHook("pre_render_forum", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.RunThemeTemplate(headerVars.Theme.Name, "forum", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeForums(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	headerVars.Zone = "forums"
	headerVars.MetaDesc = headerVars.Settings["meta_desc"].(string)

	var err error
	var forumList []common.Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = common.Forums.GetAllVisibleIDs()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		group, err := common.Groups.Get(user.Group)
		if err != nil {
			log.Printf("Group #%d doesn't exist despite being used by common.User #%d", user.Group, user.ID)
			return common.LocalError("Something weird happened", w, r, user)
		}
		canSee = group.CanSee
	}

	for _, fid := range canSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		var forum = common.Forums.DirtyGet(fid).Copy()
		if forum.ParentID == 0 && forum.Name != "" && forum.Active {
			if forum.LastTopicID != 0 {
				if forum.LastTopic.ID != 0 && forum.LastReplyer.ID != 0 {
					forum.LastTopicTime = common.RelativeTime(forum.LastTopic.LastReplyAt)
				} else {
					forum.LastTopicTime = ""
				}
			} else {
				forum.LastTopicTime = ""
			}
			if common.Hooks["forums_frow_assign"] != nil {
				common.RunHook("forums_frow_assign", &forum)
			}
			forumList = append(forumList, forum)
		}
	}

	pi := common.ForumsPage{common.GetTitlePhrase("forums"), user, headerVars, forumList}
	if common.PreRenderHooks["pre_render_forum_list"] != nil {
		if common.RunPreRenderHook("pre_render_forum_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = common.RunThemeTemplate(headerVars.Theme.Name, "forums", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeTopicID(w http.ResponseWriter, r *http.Request, user common.User, urlBit string) common.RouteError {
	var err error
	var replyList []common.ReplyUser

	page, _ := strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	// TODO: Make a shared function for this
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	tid, err := strconv.Atoi(halves[1])
	if err != nil {
		return common.PreError("The provided TopicID is not a valid number.", w, r)
	}

	// Get the topic...
	topic, err := common.GetTopicUser(tid)
	if err == ErrNoRows {
		return common.NotFound(w, r)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	topic.ClassName = ""
	//log.Printf("topic: %+v\n", topic)

	headerVars, ferr := common.ForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		//log.Printf("user.Perms: %+v\n", user.Perms)
		return common.NoPermissions(w, r, user)
	}
	headerVars.Zone = "view_topic"

	topic.ContentHTML = common.ParseMessage(topic.Content, topic.ParentID, "forums")
	topic.ContentLines = strings.Count(topic.Content, "\n")

	// We don't want users posting in locked topics...
	if topic.IsClosed && !user.IsMod {
		user.Perms.CreateReply = false
	}

	postGroup, err := common.Groups.Get(topic.Group)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	topic.Tag = postGroup.Tag
	if postGroup.IsMod || postGroup.IsAdmin {
		topic.ClassName = common.Config.StaffCSS
	}
	topic.RelativeCreatedAt = common.RelativeTime(topic.CreatedAt)

	// TODO: Make a function for this? Build a more sophisticated noavatar handling system?
	if topic.Avatar != "" {
		if topic.Avatar[0] == '.' {
			topic.Avatar = "/uploads/avatar_" + strconv.Itoa(topic.CreatedBy) + topic.Avatar
		}
	} else {
		topic.Avatar = strings.Replace(common.Config.Noavatar, "{id}", strconv.Itoa(topic.CreatedBy), 1)
	}

	// Calculate the offset
	offset, page, lastPage := common.PageOffset(topic.PostCount, page, common.Config.ItemsPerPage)
	tpage := common.TopicPage{topic.Title, user, headerVars, replyList, topic, page, lastPage}

	// Get the replies..
	rows, err := stmts.getTopicRepliesOffset.Query(topic.ID, offset, common.Config.ItemsPerPage)
	if err == ErrNoRows {
		return common.LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	replyItem := common.ReplyUser{ClassName: ""}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IPAddress, &replyItem.LikeCount, &replyItem.ActionType)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		replyItem.UserLink = common.BuildProfileURL(common.NameToSlug(replyItem.CreatedByName), replyItem.CreatedBy)
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = common.ParseMessage(replyItem.Content, topic.ParentID, "forums")
		replyItem.ContentLines = strings.Count(replyItem.Content, "\n")

		postGroup, err = common.Groups.Get(replyItem.Group)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		if postGroup.IsMod || postGroup.IsAdmin {
			replyItem.ClassName = common.Config.StaffCSS
		} else {
			replyItem.ClassName = ""
		}

		// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the common.UserStore initialise this?
		if replyItem.Avatar != "" {
			if replyItem.Avatar[0] == '.' {
				replyItem.Avatar = "/uploads/avatar_" + strconv.Itoa(replyItem.CreatedBy) + replyItem.Avatar
			}
		} else {
			replyItem.Avatar = strings.Replace(common.Config.Noavatar, "{id}", strconv.Itoa(replyItem.CreatedBy), 1)
		}

		replyItem.Tag = postGroup.Tag
		replyItem.RelativeCreatedAt = common.RelativeTime(replyItem.CreatedAt)

		// We really shouldn't have inline HTML, we should do something about this...
		if replyItem.ActionType != "" {
			switch replyItem.ActionType {
			case "lock":
				replyItem.ActionType = "This topic has been locked by <a href='" + replyItem.UserLink + "'>" + replyItem.CreatedByName + "</a>"
				replyItem.ActionIcon = "&#x1F512;&#xFE0E"
			case "unlock":
				replyItem.ActionType = "This topic has been reopened by <a href='" + replyItem.UserLink + "'>" + replyItem.CreatedByName + "</a>"
				replyItem.ActionIcon = "&#x1F513;&#xFE0E"
			case "stick":
				replyItem.ActionType = "This topic has been pinned by <a href='" + replyItem.UserLink + "'>" + replyItem.CreatedByName + "</a>"
				replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
			case "unstick":
				replyItem.ActionType = "This topic has been unpinned by <a href='" + replyItem.UserLink + "'>" + replyItem.CreatedByName + "</a>"
				replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
			case "move":
				replyItem.ActionType = "This topic has been moved by <a href='" + replyItem.UserLink + "'>" + replyItem.CreatedByName + "</a>"
			default:
				replyItem.ActionType = replyItem.ActionType + " has happened"
				replyItem.ActionIcon = ""
			}
		}
		replyItem.Liked = false

		if common.Vhooks["topic_reply_row_assign"] != nil {
			common.RunVhook("topic_reply_row_assign", &tpage, &replyItem)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	tpage.ItemList = replyList
	if common.PreRenderHooks["pre_render_view_topic"] != nil {
		if common.RunPreRenderHook("pre_render_view_topic", w, r, &user, &tpage) {
			return nil
		}
	}
	err = common.RunThemeTemplate(headerVars.Theme.Name, "topic", tpage, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	common.TopicViewCounter.Bump(topic.ID) // TODO Move this into the router?
	return nil
}

func routeProfile(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyRelativeCreatedAt, replyAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []common.ReplyUser

	// SEO URLs...
	// TODO: Do a 301 if it's the wrong username? Do a canonical too?
	halves := strings.Split(r.URL.Path[len("/user/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	pid, err := strconv.Atoi(halves[1])
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	var puser *common.User
	if pid == user.ID {
		user.IsMod = true
		puser = &user
	} else {
		// Fetch the user data
		// TODO: Add a shared function for checking for ErrNoRows and internal erroring if it's not that case?
		puser, err = common.Users.Get(pid)
		if err == ErrNoRows {
			return common.NotFound(w, r)
		} else if err != nil {
			return common.InternalError(err, w, r)
		}
	}

	// Get the replies..
	rows, err := stmts.getProfileReplies.Query(puser.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		group, err := common.Groups.Get(replyGroup)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		replyLines = strings.Count(replyContent, "\n")
		if group.IsMod || group.IsAdmin {
			replyClassName = common.Config.StaffCSS
		} else {
			replyClassName = ""
		}

		if replyAvatar != "" {
			if replyAvatar[0] == '.' {
				replyAvatar = "/uploads/avatar_" + strconv.Itoa(replyCreatedBy) + replyAvatar
			}
		} else {
			replyAvatar = strings.Replace(common.Config.Noavatar, "{id}", strconv.Itoa(replyCreatedBy), 1)
		}

		if group.Tag != "" {
			replyTag = group.Tag
		} else if puser.ID == replyCreatedBy {
			replyTag = "Profile Owner"
		} else {
			replyTag = ""
		}

		replyLiked := false
		replyLikeCount := 0
		replyRelativeCreatedAt = common.RelativeTime(replyCreatedAt)

		// TODO: Add a hook here

		replyList = append(replyList, common.ReplyUser{rid, puser.ID, replyContent, common.ParseMessage(replyContent, 0, ""), replyCreatedBy, common.BuildProfileURL(common.NameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyRelativeCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, "", ""})
	}
	err = rows.Err()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Add a phrase for this title
	ppage := common.ProfilePage{puser.Name + "'s Profile", user, headerVars, replyList, *puser}
	if common.PreRenderHooks["pre_render_profile"] != nil {
		if common.RunPreRenderHook("pre_render_profile", w, r, &user, &ppage) {
			return nil
		}
	}

	err = common.RunThemeTemplate(headerVars.Theme.Name, "profile", ppage, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeLogin(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	pi := common.Page{common.GetTitlePhrase("login"), user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_login"] != nil {
		if common.RunPreRenderHook("pre_render_login", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "login.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func routeLoginSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}

	username := html.EscapeString(strings.Replace(r.PostFormValue("username"), "\n", "", -1))
	uid, err := common.Auth.Authenticate(username, r.PostFormValue("password"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	userPtr, err := common.Users.Get(uid)
	if err != nil {
		return common.LocalError("Bad account", w, r, user)
	}
	user = *userPtr

	var session string
	if user.Session == "" {
		session, err = common.Auth.CreateSession(uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		session = user.Session
	}

	common.Auth.SetCookies(w, uid, session)
	if user.IsAdmin {
		// Is this error check redundant? We already check for the error in PreRoute for the same IP
		// TODO: Should we be logging this?
		log.Printf("#%d has logged in with IP %s", uid, user.LastIP)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeRegister(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	pi := common.Page{common.GetTitlePhrase("register"), user, headerVars, tList, nil}
	if common.PreRenderHooks["pre_render_register"] != nil {
		if common.RunPreRenderHook("pre_render_register", w, r, &user, &pi) {
			return nil
		}
	}
	err := common.Templates.ExecuteTemplate(w, "register.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func routeRegisterSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerLite, _ := common.SimpleUserCheck(w, r, &user)

	username := html.EscapeString(strings.Replace(r.PostFormValue("username"), "\n", "", -1))
	if username == "" {
		return common.LocalError("You didn't put in a username.", w, r, user)
	}
	email := html.EscapeString(strings.Replace(r.PostFormValue("email"), "\n", "", -1))
	if email == "" {
		return common.LocalError("You didn't put in an email.", w, r, user)
	}

	password := r.PostFormValue("password")
	switch password {
	case "":
		return common.LocalError("You didn't put in a password.", w, r, user)
	case username:
		return common.LocalError("You can't use your username as your password.", w, r, user)
	case email:
		return common.LocalError("You can't use your email as your password.", w, r, user)
	}

	// ?  Move this into Create()? What if we want to programatically set weak passwords for tests?
	err := common.WeakPassword(password)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	confirmPassword := r.PostFormValue("confirm_password")
	if common.Dev.DebugMode {
		log.Print("Registration Attempt! Username: " + username) // TODO: Add more controls over what is logged when?
	}

	// Do the two inputted passwords match..?
	if password != confirmPassword {
		return common.LocalError("The two passwords don't match.", w, r, user)
	}

	var active bool
	var group int
	switch headerLite.Settings["activation_type"] {
	case 1: // Activate All
		active = true
		group = common.Config.DefaultGroup
	default: // Anything else. E.g. Admin Activation or Email Activation.
		group = common.Config.ActivationGroup
	}

	uid, err := common.Users.Create(username, password, email, group, active)
	if err == common.ErrAccountExists {
		return common.LocalError("This username isn't available. Try another.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if common.Site.EnableEmails {
		token, err := common.GenerateSafeString(80)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		_, err = stmts.addEmail.Exec(email, uid, 0, token)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		if !common.SendValidationEmail(username, email, token) {
			return common.LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.", w, r, user)
		}
	}

	session, err := common.Auth.CreateSession(uid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	common.Auth.SetCookies(w, uid, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// TODO: Set the cookie domain
func routeChangeTheme(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	//headerLite, _ := SimpleUserCheck(w, r, &user)
	// TODO: Rename isJs to something else, just in case we rewrite the JS side in WebAssembly?
	isJs := (r.PostFormValue("isJs") == "1")
	newTheme := html.EscapeString(r.PostFormValue("newTheme"))

	theme, ok := common.Themes[newTheme]
	if !ok || theme.HideFromThemes {
		return common.LocalErrorJSQ("That theme doesn't exist", w, r, user, isJs)
	}

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: common.Year}
	http.SetCookie(w, &cookie)

	if !isJs {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

// TODO: Refactor this
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)

// TODO: Refactor this endpoint
func routeAPI(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return common.PreErrorJS("Bad Form", w, r)
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		return common.PreErrorJS("Invalid Action", w, r)
	}

	module := r.FormValue("module")
	switch module {
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			return common.PreErrorJS("Invalid asid", w, r)
		}
		_, err = stmts.deleteActivityStreamMatch.Exec(user.ID, asid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			w.Write(phraseLoginAlerts)
			return nil
		}

		var msglist, event, elementType string
		var asid, actorID, targetUserID, elementID int
		var msgCount int

		err = stmts.getActivityCountByWatcher.QueryRow(user.ID).Scan(&msgCount)
		if err == ErrNoRows {
			return common.PreErrorJS("Couldn't find the parent topic", w, r)
		} else if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		rows, err := stmts.getActivityFeedByWatcher.Query(user.ID)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&asid, &actorID, &targetUserID, &event, &elementType, &elementID)
			if err != nil {
				return common.InternalErrorJS(err, w, r)
			}
			res, err := buildAlert(asid, event, elementType, actorID, targetUserID, elementID, user)
			if err != nil {
				return common.LocalErrorJS(err.Error(), w, r)
			}
			msglist += res + ","
		}
		err = rows.Err()
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		if len(msglist) != 0 {
			msglist = msglist[0 : len(msglist)-1]
		}
		_, _ = w.Write([]byte(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`))
	default:
		return common.PreErrorJS("Invalid Module", w, r)
	}
	return nil
}
