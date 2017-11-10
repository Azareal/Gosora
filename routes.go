/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"log"
	//"fmt"
	"bytes"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"./query_gen/lib"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

//var nList []string
var successJSONBytes = []byte(`{"success":"1"}`)
var cacheControlMaxAge = "max-age=" + strconv.Itoa(day) // TODO: Make this a config value

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {
}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	dest := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		dest += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

// GET functions
func routeStatic(w http.ResponseWriter, r *http.Request) {
	//log.Print("Outputting static file '" + r.URL.Path + "'")
	file, ok := staticFiles[r.URL.Path]
	if !ok {
		if dev.DebugMode {
			log.Print("Failed to find '" + r.URL.Path + "'")
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
	//Cache-Control: max-age=31536000
	h.Set("Cache-Control", cacheControlMaxAge)
	h.Set("Vary", "Accept-Encoding")
	//http.ServeContent(w,r,r.URL.Path,file.Info.ModTime(),file)
	//w.Write(file.Data)
	if strings.Contains(h.Get("Accept-Encoding"), "gzip") {
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Length", strconv.FormatInt(file.GzipLength, 10))
		io.Copy(w, bytes.NewReader(file.GzipData)) // Use w.Write instead?
	} else {
		h.Set("Content-Length", strconv.FormatInt(file.Length, 10)) // Avoid doing a type conversion every time?
		io.Copy(w, bytes.NewReader(file.Data))
	}
	//io.CopyN(w, bytes.NewReader(file.Data), staticFiles[r.URL.Path].Length)
}

// Deprecated: Test route for stopping the server during a performance analysis
/*func routeExit(w http.ResponseWriter, r *http.Request, user User) RouteError{
	db.Close()
	os.Exit(0)
}*/

// TODO: Make this a static file somehow? Is it possible for us to put this file somewhere else?
// TODO: Add a sitemap
// TODO: Add an API so that plugins can register disallowed areas. E.g. /guilds/join for plugin_guilds
func routeRobotsTxt(w http.ResponseWriter, r *http.Request) RouteError {
	_, _ = w.Write([]byte(`User-agent: *
Disallow: /panel/
Disallow: /topics/create/
Disallow: /user/edit/
Disallow: /accounts/
`))
	return nil
}

func routeOverview(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	BuildWidgets("overview", nil, headerVars, r)

	pi := Page{"Overview", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_overview"] != nil {
		if runPreRenderHook("pre_render_overview", w, r, &user, &pi) {
			return nil
		}
	}

	err := templates.ExecuteTemplate(w, "overview.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeCustomPage(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	name := r.URL.Path[len("/pages/"):]
	if templates.Lookup("page_"+name) == nil {
		return NotFound(w, r)
	}
	BuildWidgets("custom_page", name, headerVars, r)

	pi := Page{"Page", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_custom_page"] != nil {
		if runPreRenderHook("pre_render_custom_page", w, r, &user, &pi) {
			return nil
		}
	}

	err := templates.ExecuteTemplate(w, "page_"+name, pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeTopics(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	BuildWidgets("topics", nil, headerVars, r)

	// TODO: Add a function for the qlist stuff
	var qlist string
	group, err := gstore.Get(user.Group)
	if err != nil {
		log.Printf("Group #%d doesn't exist despite being used by User #%d", user.Group, user.ID)
		return LocalError("Something weird happened", w, r, user)
	}

	// TODO: Make CanSee a method on *Group with a canSee field?
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		canSee = group.CanSee
	}

	// We need a list of the visible forums for Quick Topic
	var forumList []Forum
	var argList []interface{}

	for _, fid := range canSee {
		forum := fstore.DirtyGet(fid)
		if forum.Name != "" && forum.Active {
			if forum.ParentType == "" || forum.ParentType == "forum" {
				// Optimise Quick Topic away for guests
				if user.Loggedin {
					fcopy := forum.Copy()
					// TODO: Add a hook here for plugin_guilds
					forumList = append(forumList, fcopy)
				}
			}
			// ? - Should we be showing plugin_guilds posts on /topics/?
			// ? - Would it be useful, if we could post in social groups from /topics/?
			argList = append(argList, strconv.Itoa(fid))
			qlist += "?,"

		}
	}

	// ! Need an inline error not a page level error
	if qlist == "" {
		return NotFound(w, r)
	}
	qlist = qlist[0 : len(qlist)-1]

	topicCountStmt, err := qgen.Builder.SimpleCount("topics", "parentID IN("+qlist+")", "")
	if err != nil {
		return InternalError(err, w, r)
	}

	var topicCount int
	err = topicCountStmt.QueryRow(argList...).Scan(&topicCount)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Get the current page
	page, _ := strconv.Atoi(r.FormValue("page"))

	// Calculate the offset
	var offset int
	lastPage := (topicCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = lastPage
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}

	var topicList []*TopicsRow
	stmt, err := qgen.Builder.SimpleSelect("topics", "tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount", "parentID IN("+qlist+")", "sticky DESC, lastReplyAt DESC, createdBy DESC", "?,?")
	if err != nil {
		return InternalError(err, w, r)
	}

	argList = append(argList, offset)
	argList = append(argList, config.ItemsPerPage)

	rows, err := stmt.Query(argList...)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	var reqUserList = make(map[int]bool)
	for rows.Next() {
		topicItem := TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return InternalError(err, w, r)
		}

		topicItem.Link = buildTopicURL(nameToSlug(topicItem.Title), topicItem.ID)

		forum := fstore.DirtyGet(topicItem.ParentID)
		topicItem.ForumName = forum.Name
		topicItem.ForumLink = forum.Link

		/*topicItem.CreatedAt, err = relativeTimeFromString(topicItem.CreatedAt)
		if err != nil {
			replyItem.CreatedAt = ""
		}*/
		topicItem.RelativeLastReplyAt = relativeTime(topicItem.LastReplyAt)

		if vhooks["topics_topic_row_assign"] != nil {
			runVhook("topics_topic_row_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkGetMap(idSlice)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := TopicsPage{"All Topics", user, headerVars, topicList, forumList, config.DefaultForum}
	if preRenderHooks["pre_render_topic_list"] != nil {
		if runPreRenderHook("pre_render_topic_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = RunThemeTemplate(headerVars.ThemeName, "topics", pi, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeForum(w http.ResponseWriter, r *http.Request, user User, sfid string) RouteError {
	page, _ := strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(sfid, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	fid, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("The provided ForumID is not a valid number.", w, r)
	}

	headerVars, ferr := ForumUserCheck(w, r, &user, fid)
	if ferr != nil {
		return ferr
	}

	if !user.Perms.ViewTopic {
		return NoPermissions(w, r, user)
	}

	// TODO: Fix this double-check
	forum, err := fstore.Get(fid)
	if err == ErrNoRows {
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	BuildWidgets("view_forum", forum, headerVars, r)

	// Calculate the offset
	var offset int
	// TODO: Does forum.TopicCount take the deleted items into consideration for guests?
	lastPage := (forum.TopicCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = lastPage
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}

	// TODO: Move this to *Forum
	rows, err := stmts.getForumTopicsOffset.Query(fid, offset, config.ItemsPerPage)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	// TODO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	var topicList []*TopicsRow
	var reqUserList = make(map[int]bool)
	for rows.Next() {
		var topicItem = TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			return InternalError(err, w, r)
		}

		topicItem.Link = buildTopicURL(nameToSlug(topicItem.Title), topicItem.ID)
		topicItem.RelativeLastReplyAt = relativeTime(topicItem.LastReplyAt)

		if vhooks["forum_trow_assign"] != nil {
			runVhook("forum_trow_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkGetMap(idSlice)
	if err != nil {
		return InternalError(err, w, r)
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := ForumPage{forum.Name, user, headerVars, topicList, forum, page, lastPage}
	if preRenderHooks["pre_render_view_forum"] != nil {
		if runPreRenderHook("pre_render_view_forum", w, r, &user, &pi) {
			return nil
		}
	}
	err = RunThemeTemplate(headerVars.ThemeName, "forum", pi, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeForums(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	BuildWidgets("forums", nil, headerVars, r)

	var err error
	var forumList []Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			return InternalError(err, w, r)
		}
		//log.Print("canSee ", canSee)
	} else {
		group, err := gstore.Get(user.Group)
		if err != nil {
			log.Printf("Group #%d doesn't exist despite being used by User #%d", user.Group, user.ID)
			return LocalError("Something weird happened", w, r, user)
		}
		canSee = group.CanSee
		//log.Print("group.CanSee ", group.CanSee)
	}

	for _, fid := range canSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		var forum = fstore.DirtyGet(fid).Copy()
		if forum.ParentID == 0 && forum.Name != "" && forum.Active {
			if forum.LastTopicID != 0 {
				//topic, user := forum.GetLast()
				//if topic.ID != 0 && user.ID != 0 {
				if forum.LastTopic.ID != 0 && forum.LastReplyer.ID != 0 {
					forum.LastTopicTime = relativeTime(forum.LastTopic.LastReplyAt)
				} else {
					forum.LastTopicTime = ""
				}
			} else {
				forum.LastTopicTime = ""
			}
			if hooks["forums_frow_assign"] != nil {
				runHook("forums_frow_assign", &forum)
			}
			forumList = append(forumList, forum)
		}
	}

	pi := ForumsPage{"Forum List", user, headerVars, forumList}
	if preRenderHooks["pre_render_forum_list"] != nil {
		if runPreRenderHook("pre_render_forum_list", w, r, &user, &pi) {
			return nil
		}
	}
	err = RunThemeTemplate(headerVars.ThemeName, "forums", pi, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeTopicID(w http.ResponseWriter, r *http.Request, user User) RouteError {
	var err error
	var page, offset int
	var replyList []ReplyUser

	page, _ = strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/topic/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	tid, err := strconv.Atoi(halves[1])
	if err != nil {
		return PreError("The provided TopicID is not a valid number.", w, r)
	}

	// Get the topic...
	topic, err := getTopicUser(tid)
	if err == ErrNoRows {
		return NotFound(w, r)
	} else if err != nil {
		return InternalError(err, w, r)
	}
	topic.ClassName = ""
	//log.Printf("topic: %+v\n", topic)

	headerVars, ferr := ForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		//log.Printf("user.Perms: %+v\n", user.Perms)
		return NoPermissions(w, r, user)
	}

	BuildWidgets("view_topic", &topic, headerVars, r)

	topic.ContentHTML = parseMessage(topic.Content, topic.ParentID, "forums")
	topic.ContentLines = strings.Count(topic.Content, "\n")

	// We don't want users posting in locked topics...
	if topic.IsClosed && !user.IsMod {
		user.Perms.CreateReply = false
	}

	postGroup, err := gstore.Get(topic.Group)
	if err != nil {
		return InternalError(err, w, r)
	}

	topic.Tag = postGroup.Tag
	if postGroup.IsMod || postGroup.IsAdmin {
		topic.ClassName = config.StaffCSS
	}
	topic.RelativeCreatedAt = relativeTime(topic.CreatedAt)

	// TODO: Make a function for this? Build a more sophisticated noavatar handling system?
	if topic.Avatar != "" {
		if topic.Avatar[0] == '.' {
			topic.Avatar = "/uploads/avatar_" + strconv.Itoa(topic.CreatedBy) + topic.Avatar
		}
	} else {
		topic.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(topic.CreatedBy), 1)
	}

	// Calculate the offset
	lastPage := (topic.PostCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = lastPage
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}

	tpage := TopicPage{topic.Title, user, headerVars, replyList, topic, page, lastPage}

	// Get the replies..
	rows, err := stmts.getTopicRepliesOffset.Query(topic.ID, offset, config.ItemsPerPage)
	if err == ErrNoRows {
		return LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	replyItem := ReplyUser{ClassName: ""}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IPAddress, &replyItem.LikeCount, &replyItem.ActionType)
		if err != nil {
			return InternalError(err, w, r)
		}

		replyItem.UserLink = buildProfileURL(nameToSlug(replyItem.CreatedByName), replyItem.CreatedBy)
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = parseMessage(replyItem.Content, topic.ParentID, "forums")
		replyItem.ContentLines = strings.Count(replyItem.Content, "\n")

		postGroup, err = gstore.Get(replyItem.Group)
		if err != nil {
			return InternalError(err, w, r)
		}

		if postGroup.IsMod || postGroup.IsAdmin {
			replyItem.ClassName = config.StaffCSS
		} else {
			replyItem.ClassName = ""
		}

		// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the UserStore initialise this?
		if replyItem.Avatar != "" {
			if replyItem.Avatar[0] == '.' {
				replyItem.Avatar = "/uploads/avatar_" + strconv.Itoa(replyItem.CreatedBy) + replyItem.Avatar
			}
		} else {
			replyItem.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(replyItem.CreatedBy), 1)
		}

		replyItem.Tag = postGroup.Tag
		replyItem.RelativeCreatedAt = relativeTime(replyItem.CreatedAt)

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
			default:
				replyItem.ActionType = replyItem.ActionType + " has happened"
				replyItem.ActionIcon = ""
			}
		}
		replyItem.Liked = false

		if vhooks["topic_reply_row_assign"] != nil {
			runVhook("topic_reply_row_assign", &tpage, &replyItem)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	tpage.ItemList = replyList
	if preRenderHooks["pre_render_view_topic"] != nil {
		if runPreRenderHook("pre_render_view_topic", w, r, &user, &tpage) {
			return nil
		}
	}
	err = RunThemeTemplate(headerVars.ThemeName, "topic", tpage, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeProfile(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyRelativeCreatedAt, replyAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []ReplyUser

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/user/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	pid, err := strconv.Atoi(halves[1])
	if err != nil {
		return LocalError("The provided User ID is not a valid number.", w, r, user)
	}

	var puser *User
	if pid == user.ID {
		user.IsMod = true
		puser = &user
	} else {
		// Fetch the user data
		puser, err = users.Get(pid)
		if err == ErrNoRows {
			return NotFound(w, r)
		} else if err != nil {
			return InternalError(err, w, r)
		}
	}

	// Get the replies..
	rows, err := stmts.getProfileReplies.Query(puser.ID)
	if err != nil {
		return InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			return InternalError(err, w, r)
		}

		group, err := gstore.Get(replyGroup)
		if err != nil {
			return InternalError(err, w, r)
		}

		replyLines = strings.Count(replyContent, "\n")
		if group.IsMod || group.IsAdmin {
			replyClassName = config.StaffCSS
		} else {
			replyClassName = ""
		}

		if replyAvatar != "" {
			if replyAvatar[0] == '.' {
				replyAvatar = "/uploads/avatar_" + strconv.Itoa(replyCreatedBy) + replyAvatar
			}
		} else {
			replyAvatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(replyCreatedBy), 1)
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
		replyRelativeCreatedAt = relativeTime(replyCreatedAt)

		// TODO: Add a hook here

		replyList = append(replyList, ReplyUser{rid, puser.ID, replyContent, parseMessage(replyContent, 0, ""), replyCreatedBy, buildProfileURL(nameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyRelativeCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, "", ""})
	}
	err = rows.Err()
	if err != nil {
		return InternalError(err, w, r)
	}

	ppage := ProfilePage{puser.Name + "'s Profile", user, headerVars, replyList, *puser}
	if preRenderHooks["pre_render_profile"] != nil {
		if runPreRenderHook("pre_render_profile", w, r, &user, &ppage) {
			return nil
		}
	}

	err = template_profile_handle(ppage, w)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeLogin(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return LocalError("You're already logged in.", w, r, user)
	}
	pi := Page{"Login", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_login"] != nil {
		if runPreRenderHook("pre_render_login", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "login.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func routeLoginSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	if user.Loggedin {
		return LocalError("You're already logged in.", w, r, user)
	}
	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}

	uid, err := auth.Authenticate(html.EscapeString(r.PostFormValue("username")), r.PostFormValue("password"))
	if err != nil {
		return LocalError(err.Error(), w, r, user)
	}

	userPtr, err := users.Get(uid)
	if err != nil {
		return LocalError("Bad account", w, r, user)
	}
	user = *userPtr

	var session string
	if user.Session == "" {
		session, err = auth.CreateSession(uid)
		if err != nil {
			return InternalError(err, w, r)
		}
	} else {
		session = user.Session
	}

	auth.SetCookies(w, uid, session)
	if user.IsAdmin {
		// Is this error check reundant? We already check for the error in PreRoute for the same IP
		// TODO: Should we be logging this?
		log.Printf("#%d has logged in with IP %s", uid, user.LastIP)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func routeRegister(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerVars, ferr := UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return LocalError("You're already logged in.", w, r, user)
	}
	pi := Page{"Registration", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_register"] != nil {
		if runPreRenderHook("pre_render_register", w, r, &user, &pi) {
			return nil
		}
	}
	err := templates.ExecuteTemplate(w, "register.html", pi)
	if err != nil {
		return InternalError(err, w, r)
	}
	return nil
}

func routeRegisterSubmit(w http.ResponseWriter, r *http.Request, user User) RouteError {
	headerLite, _ := SimpleUserCheck(w, r, &user)

	err := r.ParseForm()
	if err != nil {
		return LocalError("Bad Form", w, r, user)
	}

	username := html.EscapeString(r.PostFormValue("username"))
	if username == "" {
		return LocalError("You didn't put in a username.", w, r, user)
	}
	email := html.EscapeString(r.PostFormValue("email"))
	if email == "" {
		return LocalError("You didn't put in an email.", w, r, user)
	}

	password := r.PostFormValue("password")
	if password == "" {
		return LocalError("You didn't put in a password.", w, r, user)
	}
	if password == username {
		return LocalError("You can't use your username as your password.", w, r, user)
	}
	if password == email {
		return LocalError("You can't use your email as your password.", w, r, user)
	}

	err = weakPassword(password)
	if err != nil {
		return LocalError(err.Error(), w, r, user)
	}

	confirmPassword := r.PostFormValue("confirm_password")
	log.Print("Registration Attempt! Username: " + username) // TODO: Add more controls over what is logged when?

	// Do the two inputted passwords match..?
	if password != confirmPassword {
		return LocalError("The two passwords don't match.", w, r, user)
	}

	var active bool
	var group int
	switch headerLite.Settings["activation_type"] {
	case 1: // Activate All
		active = true
		group = config.DefaultGroup
	default: // Anything else. E.g. Admin Activation or Email Activation.
		group = config.ActivationGroup
	}

	uid, err := users.Create(username, password, email, group, active)
	if err == errAccountExists {
		return LocalError("This username isn't available. Try another.", w, r, user)
	} else if err != nil {
		return InternalError(err, w, r)
	}

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if site.EnableEmails {
		token, err := GenerateSafeString(80)
		if err != nil {
			return InternalError(err, w, r)
		}
		_, err = stmts.addEmail.Exec(email, uid, 0, token)
		if err != nil {
			return InternalError(err, w, r)
		}

		if !SendValidationEmail(username, email, token) {
			return LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.", w, r, user)
		}
	}

	session, err := auth.CreateSession(uid)
	if err != nil {
		return InternalError(err, w, r)
	}

	auth.SetCookies(w, uid, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// TODO: Set the cookie domain
func routeChangeTheme(w http.ResponseWriter, r *http.Request, user User) RouteError {
	//headerLite, _ := SimpleUserCheck(w, r, &user)
	err := r.ParseForm()
	if err != nil {
		return PreError("Bad Form", w, r)
	}

	// TODO: Rename isJs to something else, just in case we rewrite the JS side in WebAssembly?
	isJs := (r.PostFormValue("isJs") == "1")

	newTheme := html.EscapeString(r.PostFormValue("newTheme"))

	theme, ok := themes[newTheme]
	if !ok || theme.HideFromThemes {
		// TODO: Should we be logging this?
		log.Print("Bad Theme: ", newTheme)
		return LocalErrorJSQ("That theme doesn't exist", w, r, user, isJs)
	}

	// TODO: Store the current theme in the user's account?
	/*if user.Loggedin {
		_, err = stmts.changeTheme.Exec(newTheme, user.ID)
		if err != nil {
			return InternalError(err, w, r)
		}
	}*/

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)

	if !isJs {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}

// TODO: We don't need support XML here to support sitemaps, we could handle those elsewhere
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)

func routeAPI(w http.ResponseWriter, r *http.Request, user User) RouteError {
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return PreErrorJS("Bad Form", w, r)
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		return PreErrorJS("Invalid Action", w, r)
	}

	module := r.FormValue("module")
	switch module {
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			return PreErrorJS("Invalid asid", w, r)
		}

		_, err = stmts.deleteActivityStreamMatch.Exec(user.ID, asid)
		if err != nil {
			return InternalError(err, w, r)
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
			return PreErrorJS("Couldn't find the parent topic", w, r)
		} else if err != nil {
			return InternalErrorJS(err, w, r)
		}

		rows, err := stmts.getActivityFeedByWatcher.Query(user.ID)
		if err != nil {
			return InternalErrorJS(err, w, r)
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&asid, &actorID, &targetUserID, &event, &elementType, &elementID)
			if err != nil {
				return InternalErrorJS(err, w, r)
			}
			res, err := buildAlert(asid, event, elementType, actorID, targetUserID, elementID, user)
			if err != nil {
				return LocalErrorJS(err.Error(), w, r)
			}
			msglist += res + ","
		}

		err = rows.Err()
		if err != nil {
			return InternalErrorJS(err, w, r)
		}

		if len(msglist) != 0 {
			msglist = msglist[0 : len(msglist)-1]
		}
		_, _ = w.Write([]byte(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`))
		//log.Print(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`)
	//case "topics":
	//case "forums":
	//case "users":
	//case "pages":
	// This might not be possible. We might need .xml paths for sitemaps
	/*case "sitemap":
	if format != "xml" {
		PreError("You can only fetch sitemaps in the XML format!",w,r)
		return
	}*/
	default:
		return PreErrorJS("Invalid Module", w, r)
	}
	return nil
}
