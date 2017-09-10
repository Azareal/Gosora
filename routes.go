/*
*
* Gosora Route Handlers
* Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"log"
	//"fmt"
	"bytes"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"./query_gen/lib"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

//var nList []string
var hvars *HeaderVars // We might need to rethink this now that it's a pointer
var successJSONBytes = []byte(`{"success":"1"}`)
var cacheControlMaxAge = "max-age=" + strconv.Itoa(day)

func init() {
	hvars = &HeaderVars{Site: site}
}

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
func route_static(w http.ResponseWriter, r *http.Request) {
	//log.Print("Outputting static file '" + r.URL.Path + "'")
	file, ok := staticFiles[r.URL.Path]
	if !ok {
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
	//io.CopyN(w, bytes.NewReader(file.Data), static_files[r.URL.Path].Length)
}

// Deprecated: Test route for stopping the server during a performance analysis
/*func route_exit(w http.ResponseWriter, r *http.Request){
	db.Close()
	os.Exit(0)
}

// Deprecated: Test route to see which file serving method is faster
func route_fstatic(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w,r,r.URL.Path)
}*/

// TODO: Make this a static file somehow? Is it possible for us to put this file somewhere else?
// TODO: Add a sitemap
// TODO: Add an API so that plugins can register disallowed areas. E.g. /groups/join for plugin_socialgroups
func route_robots_txt(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`User-agent: *
Disallow: /panel/
Disallow: /topics/create/
Disallow: /user/edit/
Disallow: /accounts/
`))
}

func route_overview(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	BuildWidgets("overview", nil, headerVars, r)

	pi := Page{"Overview", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_overview"] != nil {
		if runPreRenderHook("pre_render_overview", w, r, &user, &pi) {
			return
		}
	}

	err := templates.ExecuteTemplate(w, "overview.html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func route_custom_page(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}

	name := r.URL.Path[len("/pages/"):]
	if templates.Lookup("page_"+name) == nil {
		NotFound(w, r)
		return
	}
	BuildWidgets("custom_page", name, headerVars, r)

	pi := Page{"Page", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_custom_page"] != nil {
		if runPreRenderHook("pre_render_custom_page", w, r, &user, &pi) {
			return
		}
	}

	err := templates.ExecuteTemplate(w, "page_"+name, pi)
	if err != nil {
		InternalError(err, w)
	}
}

// TODO: Paginate this
func route_topics(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	BuildWidgets("topics", nil, headerVars, r)

	var qlist string
	var fidList []interface{}
	group := groups[user.Group]
	for _, fid := range group.CanSee {
		if fstore.DirtyGet(fid).Name != "" {
			fidList = append(fidList, strconv.Itoa(fid))
			qlist += "?,"
		}
	}
	qlist = qlist[0 : len(qlist)-1]

	var topicList []*TopicsRow
	//stmt, err := qgen.Builder.SimpleLeftJoin("topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, topics.postCount, topics.likeCount, users.name, users.avatar","topics.createdBy = users.uid","parentID IN("+qlist+")","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC","")
	stmt, err := qgen.Builder.SimpleSelect("topics", "tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount", "parentID IN("+qlist+")", "sticky DESC, lastReplyAt DESC, createdBy DESC", "")
	if err != nil {
		InternalError(err, w)
		return
	}

	rows, err := stmt.Query(fidList...)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	var reqUserList = make(map[int]bool)
	for rows.Next() {
		topicItem := TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			InternalError(err, w)
			return
		}

		topicItem.Link = buildTopicURL(nameToSlug(topicItem.Title), topicItem.ID)

		forum := fstore.DirtyGet(topicItem.ParentID)
		if topicItem.ParentID >= 0 {
			topicItem.ForumName = forum.Name
			topicItem.ForumLink = forum.Link
		} else {
			topicItem.ForumName = ""
			//topicItem.ForumLink = ""
		}

		/*topicItem.CreatedAt, err = relativeTime(topicItem.CreatedAt)
		if err != nil {
			replyItem.CreatedAt = ""
		}*/
		topicItem.LastReplyAt, err = relativeTime(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err, w)
		}

		if vhooks["topics_topic_row_assign"] != nil {
			runVhook("topics_topic_row_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkCascadeGetMap(idSlice)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := TopicsPage{"Topic List", user, headerVars, topicList}
	if preRenderHooks["pre_render_topic_list"] != nil {
		if runPreRenderHook("pre_render_topic_list", w, r, &user, &pi) {
			return
		}
	}
	RunThemeTemplate(headerVars.ThemeName, "topics", pi, w)
}

func route_forum(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	page, _ := strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(sfid, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	fid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("The provided ForumID is not a valid number.", w, r)
		return
	}

	headerVars, ok := ForumSessionCheck(w, r, &user, fid)
	if !ok {
		return
	}
	//log.Printf("groups[user.Group]: %+v\n", groups[user.Group].Forums)
	if !user.Perms.ViewTopic {
		NoPermissions(w, r, user)
		return
	}

	// TODO: Fix this double-check
	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		NotFound(w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	BuildWidgets("view_forum", forum, headerVars, r)

	// Calculate the offset
	var offset int
	lastPage := (forum.TopicCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = lastPage
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}
	rows, err := get_forum_topics_offset_stmt.Query(fid, offset, config.ItemsPerPage)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	// TODO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	var topicList []*TopicsRow
	var reqUserList = make(map[int]bool)
	for rows.Next() {
		var topicItem = TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.IsClosed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			InternalError(err, w)
			return
		}

		topicItem.Link = buildTopicURL(nameToSlug(topicItem.Title), topicItem.ID)
		topicItem.LastReplyAt, err = relativeTime(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err, w)
		}

		if vhooks["forum_trow_assign"] != nil {
			runVhook("forum_trow_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TODO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkCascadeGetMap(idSlice)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Second pass to the add the user data
	// TODO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := ForumPage{forum.Name, user, headerVars, topicList, *forum, page, lastPage}
	if preRenderHooks["pre_render_view_forum"] != nil {
		if runPreRenderHook("pre_render_view_forum", w, r, &user, &pi) {
			return
		}
	}
	RunThemeTemplate(headerVars.ThemeName, "forum", pi, w)
}

func route_forums(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	BuildWidgets("forums", nil, headerVars, r)

	var err error
	var forumList []Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			InternalError(err, w)
			return
		}
		//log.Print("canSee",canSee)
	} else {
		group := groups[user.Group]
		canSee = group.CanSee
		//log.Print("group.CanSee",group.CanSee)
	}

	for _, fid := range canSee {
		//log.Print(forums[fid])
		var forum = *fstore.DirtyGet(fid)
		if forum.ParentID == 0 {
			if forum.LastTopicID != 0 {
				forum.LastTopicTime, err = relativeTime(forum.LastTopicTime)
				if err != nil {
					InternalError(err, w)
				}
			} else {
				forum.LastTopic = "None"
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
			return
		}
	}
	RunThemeTemplate(headerVars.ThemeName, "forums", pi, w)
}

func route_topic_id(w http.ResponseWriter, r *http.Request, user User) {
	var err error
	var page, offset int
	var replyList []Reply

	page, _ = strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/topic/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	tid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("The provided TopicID is not a valid number.", w, r)
		return
	}

	// Get the topic...
	topic, err := getTopicuser(tid)
	if err == ErrNoRows {
		NotFound(w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}
	topic.ClassName = ""

	headerVars, ok := ForumSessionCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic {
		//log.Printf("user.Perms: %+v\n", user.Perms)
		NoPermissions(w, r, user)
		return
	}

	BuildWidgets("view_topic", &topic, headerVars, r)

	topic.Content = parseMessage(topic.Content)
	topic.ContentLines = strings.Count(topic.Content, "\n")

	// We don't want users posting in locked topics...
	if topic.IsClosed && !user.IsMod {
		user.Perms.CreateReply = false
	}

	topic.Tag = groups[topic.Group].Tag
	if groups[topic.Group].IsMod || groups[topic.Group].IsAdmin {
		topic.ClassName = config.StaffCss
	}

	/*if headerVars.Settings["url_tags"] == false {
		topic.URLName = ""
	} else {
		topic.URL, ok = external_sites[topic.URLPrefix]
		if !ok {
			topic.URL = topic.URLName
		} else {
			topic.URL = topic.URL + topic.URLName
		}
	}*/

	topic.CreatedAt, err = relativeTime(topic.CreatedAt)
	if err != nil {
		topic.CreatedAt = ""
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

	// Get the replies..
	rows, err := get_topic_replies_offset_stmt.Query(topic.ID, offset, config.ItemsPerPage)
	if err == ErrNoRows {
		LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	replyItem := Reply{ClassName: ""}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IPAddress, &replyItem.LikeCount, &replyItem.ActionType)
		if err != nil {
			InternalError(err, w)
			return
		}

		replyItem.UserLink = buildProfileURL(nameToSlug(replyItem.CreatedByName), replyItem.CreatedBy)
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = parseMessage(replyItem.Content)
		replyItem.ContentLines = strings.Count(replyItem.Content, "\n")

		if groups[replyItem.Group].IsMod || groups[replyItem.Group].IsAdmin {
			replyItem.ClassName = config.StaffCss
		} else {
			replyItem.ClassName = ""
		}

		if replyItem.Avatar != "" {
			if replyItem.Avatar[0] == '.' {
				replyItem.Avatar = "/uploads/avatar_" + strconv.Itoa(replyItem.CreatedBy) + replyItem.Avatar
			}
		} else {
			replyItem.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(replyItem.CreatedBy), 1)
		}

		replyItem.Tag = groups[replyItem.Group].Tag

		/*if headerVars.Settings["url_tags"] == false {
			replyItem.URLName = ""
		} else {
			replyItem.URL, ok = external_sites[replyItem.URLPrefix]
			if !ok {
				replyItem.URL = replyItem.URLName
			} else {
				replyItem.URL = replyItem.URL + replyItem.URLName
			}
		}*/

		replyItem.CreatedAt, err = relativeTime(replyItem.CreatedAt)
		if err != nil {
			replyItem.CreatedAt = ""
		}

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

		// TODO: Rename this to topic_rrow_assign
		if hooks["rrow_assign"] != nil {
			runHook("rrow_assign", &replyItem)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	tpage := TopicPage{topic.Title, user, headerVars, replyList, topic, page, lastPage}
	if preRenderHooks["pre_render_view_topic"] != nil {
		if runPreRenderHook("pre_render_view_topic", w, r, &user, &tpage) {
			return
		}
	}
	RunThemeTemplate(headerVars.ThemeName, "topic", tpage, w)
}

func route_profile(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}

	var err error
	var replyContent, replyCreatedByName, replyCreatedAt, replyAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []Reply

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/user/"):], ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}

	pid, err := strconv.Atoi(halves[1])
	if err != nil {
		LocalError("The provided User ID is not a valid number.", w, r, user)
		return
	}

	var puser *User
	if pid == user.ID {
		user.IsMod = true
		puser = &user
	} else {
		// Fetch the user data
		puser, err = users.CascadeGet(pid)
		if err == ErrNoRows {
			NotFound(w, r)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}
	}

	// Get the replies..
	rows, err := get_profile_replies_stmt.Query(puser.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			InternalError(err, w)
			return
		}

		replyLines = strings.Count(replyContent, "\n")
		if groups[replyGroup].IsMod || groups[replyGroup].IsAdmin {
			replyClassName = config.StaffCss
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

		if groups[replyGroup].Tag != "" {
			replyTag = groups[replyGroup].Tag
		} else if puser.ID == replyCreatedBy {
			replyTag = "Profile Owner"
		} else {
			replyTag = ""
		}

		replyLiked := false
		replyLikeCount := 0

		// TODO: Add a hook here

		replyList = append(replyList, Reply{rid, puser.ID, replyContent, parseMessage(replyContent), replyCreatedBy, buildProfileURL(nameToSlug(replyCreatedByName), replyCreatedBy), replyCreatedByName, replyGroup, replyCreatedAt, replyLastEdit, replyLastEditBy, replyAvatar, replyClassName, replyLines, replyTag, "", "", "", 0, "", replyLiked, replyLikeCount, "", ""})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	ppage := ProfilePage{puser.Name + "'s Profile", user, headerVars, replyList, *puser}
	if preRenderHooks["pre_render_profile"] != nil {
		if runPreRenderHook("pre_render_profile", w, r, &user, &ppage) {
			return
		}
	}

	template_profile_handle(ppage, w)
}

func route_topic_create(w http.ResponseWriter, r *http.Request, user User, sfid string) {
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			PreError("The provided ForumID is not a valid number.", w, r)
			return
		}
	}

	headerVars, ok := ForumSessionCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w, r, user)
		return
	}

	BuildWidgets("create_topic", nil, headerVars, r)

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strictmode bool
	if vhooks["topic_create_pre_loop"] != nil {
		runVhook("topic_create_pre_loop", w, r, fid, &headerVars, &user, &strictmode)
	}

	// TODO: Re-add support for plugin_socialgroups
	var forumList []Forum
	var canSee []int
	if user.IsSuperAdmin {
		canSee, err = fstore.GetAllVisibleIDs()
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		group := groups[user.Group]
		canSee = group.CanSee
	}

	// TODO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TODO: Surely, there's a better way of doing this. I've added it in for now to support plugin_socialgroups, but we really need to clean this up
		if strictmode && ffid != fid {
			continue
		}

		// Do a bulk forum fetch, just in case it's the SqlForumStore?
		forum := fstore.DirtyGet(ffid)
		fcopy := *forum
		if hooks["topic_create_frow_assign"] != nil {
			// TODO: Add the skip feature to all the other row based hooks?
			if runHook("topic_create_frow_assign", &fcopy).(bool) {
				continue
			}
		}
		forumList = append(forumList, fcopy)
	}

	ctpage := CreateTopicPage{"Create Topic", user, headerVars, forumList, fid}
	if preRenderHooks["pre_render_create_topic"] != nil {
		if runPreRenderHook("pre_render_create_topic", w, r, &user, &ctpage) {
			return
		}
	}

	template_create_topic_handle(ctpage, w)
}

// POST functions. Authorised users only.
func route_topic_create_submit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		PreError("The provided ForumID is not a valid number.", w, r)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumSessionCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w, r, user)
		return
	}

	topicName := html.EscapeString(r.PostFormValue("topic-name"))
	content := html.EscapeString(preparseMessage(r.PostFormValue("topic-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	wcount := wordCount(content)
	res, err := create_topic_stmt.Exec(fid, topicName, content, parseMessage(content), user.ID, ipaddress, wcount, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	err = fstore.IncrementTopicCount(fid)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_subscription_stmt.Exec(user.ID, lastID, "topic")
	if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
	err = user.increasePostStats(wcount, true)
	if err != nil {
		InternalError(err, w)
		return
	}

	err = fstore.UpdateLastTopic(topicName, int(lastID), user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), fid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
	}
}

func route_create_reply(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		PreError("Failed to convert the Topic ID", w, r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("Couldn't find the parent topic", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumSessionCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateReply {
		NoPermissions(w, r, user)
		return
	}

	content := preparseMessage(html.EscapeString(r.PostFormValue("reply-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	wcount := wordCount(content)
	_, err = create_reply_stmt.Exec(tid, content, parseMessage(content), ipaddress, wcount, user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_replies_to_topic_stmt.Exec(1, user.ID, tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = fstore.UpdateLastTopic(topic.Title, tid, user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), topic.ParentID)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, topic.CreatedBy, "reply", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_watchers_stmt.Exec(lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enableWebsockets {
		go notifyWatchers(lastID)
	}

	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The destination no longer exists", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
	err = user.increasePostStats(wcount, false)
	if err != nil {
		InternalError(err, w)
		return
	}
}

func route_like_topic(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/like/submit/"):])
	if err != nil {
		PreError("Topic IDs can only ever be numbers.", w, r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("The requested topic doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumSessionCheck(w, r, &user, topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w, r, user)
		return
	}

	if topic.CreatedBy == user.ID {
		LocalError("You can't like your own topics", w, r, user)
		return
	}

	err = has_liked_topic_stmt.QueryRow(user.ID, tid).Scan(&tid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!", w, r, user)
		return
	}

	_, err = users.CascadeGet(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		LocalError("The target user doesn't exist", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score, tid, "topics", user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_likes_to_topic_stmt.Exec(1, tid)
	if err != nil {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, topic.CreatedBy, "like", "topic", tid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_one_stmt.Exec(topic.CreatedBy, lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(topic.CreatedBy, int(lastID), "like", "topic", user.ID, topic.CreatedBy, tid)

	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The liked topic no longer exists", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(tid), http.StatusSeeOther)
}

func route_reply_like_submit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/like/submit/"):])
	if err != nil {
		PreError("The provided Reply ID is not a valid number.", w, r)
		return
	}

	reply, err := getReply(rid)
	if err == ErrNoRows {
		PreError("You can't like something which doesn't exist!", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	var fid int
	err = get_topic_fid_stmt.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		PreError("The parent topic doesn't exist.", w, r)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// TODO: Add hooks to make use of headerLite
	_, ok := SimpleForumSessionCheck(w, r, &user, fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w, r, user)
		return
	}

	if reply.CreatedBy == user.ID {
		LocalError("You can't like your own replies", w, r, user)
		return
	}

	err = has_liked_reply_stmt.QueryRow(user.ID, rid).Scan(&rid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!", w, r, user)
		return
	}

	_, err = users.CascadeGet(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		LocalError("The target user doesn't exist", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score, rid, "replies", user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_likes_to_reply_stmt.Exec(1, rid)
	if err != nil {
		InternalError(err, w)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID, reply.CreatedBy, "like", "post", rid)
	if err != nil {
		InternalError(err, w)
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = notify_one_stmt.Exec(reply.CreatedBy, lastID)
	if err != nil {
		InternalError(err, w)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = wsHub.pushAlert(reply.CreatedBy, int(lastID), "like", "post", user.ID, reply.CreatedBy, rid)

	http.Redirect(w, r, "/topic/"+strconv.Itoa(reply.ParentID), http.StatusSeeOther)
}

func route_profile_reply_create(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Loggedin || !user.Perms.CreateReply {
		NoPermissions(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		LocalError("Invalid UID", w, r, user)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP", w, r, user)
		return
	}

	_, err = create_profile_reply_stmt.Exec(uid, html.EscapeString(preparseMessage(r.PostFormValue("reply-content"))), parseMessage(html.EscapeString(preparseMessage(r.PostFormValue("reply-content")))), user.ID, ipaddress)
	if err != nil {
		InternalError(err, w)
		return
	}

	var userName string
	err = get_user_name_stmt.QueryRow(uid).Scan(&userName)
	if err == ErrNoRows {
		LocalError("The profile you're trying to post on doesn't exist.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
}

func route_report_submit(w http.ResponseWriter, r *http.Request, user User, sitemID string) {
	if !user.Loggedin {
		LoginRequired(w, r, user)
		return
	}
	if user.IsBanned {
		Banned(w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w, r, user)
		return
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		LocalError("Bad ID", w, r, user)
		return
	}

	itemType := r.FormValue("type")

	var fid = 1
	var title, content string
	if itemType == "reply" {
		reply, err := getReply(itemID)
		if err == ErrNoRows {
			LocalError("We were unable to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		topic, err := topics.CascadeGet(reply.ParentID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the topic the reported post is supposed to be in", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(itemID)
	} else if itemType == "user-reply" {
		userReply, err := getUserReply(itemID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}

		err = get_user_name_stmt.QueryRow(userReply.ParentID).Scan(&title)
		if err == ErrNoRows {
			LocalError("We weren't able to find the profile the reported post is supposed to be on", w, r, user)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}
		title = "Profile: " + title
		content = userReply.Content + "\n\nOriginal Post: @" + strconv.Itoa(userReply.ParentID)
	} else if itemType == "topic" {
		err = get_topic_basic_stmt.QueryRow(itemID).Scan(&title, &content)
		if err == ErrNoRows {
			NotFound(w, r)
			return
		} else if err != nil {
			InternalError(err, w)
			return
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(itemID)
	} else {
		if vhooks["report_preassign"] != nil {
			runVhookNoreturn("report_preassign", &itemID, &itemType)
			return
		}
		// Don't try to guess the type
		LocalError("Unknown type", w, r, user)
		return
	}

	var count int
	rows, err := report_exists_stmt.Query(itemType + "_" + strconv.Itoa(itemID))
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			InternalError(err, w)
			return
		}
	}
	if count != 0 {
		LocalError("Someone has already reported this!", w, r, user)
		return
	}

	res, err := create_report_stmt.Exec(title, content, parseMessage(content), user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		InternalError(err, w)
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		InternalError(err, w)
		return
	}

	_, err = add_topics_to_forum_stmt.Exec(1, fid)
	if err != nil {
		InternalError(err, w)
		return
	}
	err = fstore.UpdateLastTopic(title, int(lastID), user.Name, user.ID, time.Now().Format("2006-01-02 15:04:05"), fid)
	if err != nil && err != ErrNoRows {
		InternalError(err, w)
		return
	}

	http.Redirect(w, r, "/topic/"+strconv.FormatInt(lastID, 10), http.StatusSeeOther)
}

func route_account_own_edit_critical(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit.html", pi)
}

func route_account_own_edit_critical_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	err = get_password_stmt.QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == ErrNoRows {
		LocalError("Your account no longer exists.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	err = CheckPassword(realPassword, currentPassword, salt)
	if err == ErrMismatchedHashAndPassword {
		LocalError("That's not the correct password.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}
	if newPassword != confirmPassword {
		LocalError("The two passwords don't match.", w, r, user)
		return
	}
	SetPassword(user.ID, newPassword)

	// Log the user out as a safety precaution
	auth.ForceLogout(user.ID)

	headerVars.NoticeList = append(headerVars.NoticeList, "Your password was successfully updated")
	pi := Page{"Edit Password", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_critical"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit.html", pi)
}

func route_account_own_edit_avatar(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
}

func route_account_own_edit_avatar_submit(w http.ResponseWriter, r *http.Request, user User) {
	if r.ContentLength > int64(config.MaxRequestSize) {
		http.Error(w, "Request too large", http.StatusExpectationFailed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	err := r.ParseMultipartForm(int64(config.MaxRequestSize))
	if err != nil {
		LocalError("Upload failed", w, r, user)
		return
	}

	var filename string
	var ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open()
			if err != nil {
				LocalError("Upload failed", w, r, user)
				return
			}
			defer infile.Close()

			// We don't want multiple files
			if filename != "" {
				if filename != hdr.Filename {
					os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
					LocalError("You may only upload one avatar", w, r, user)
					return
				}
			} else {
				filename = hdr.Filename
			}

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					LocalError("Bad file", w, r, user)
					return
				}
				ext = extarr[len(extarr)-1]

				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					LocalError("Bad file extension", w, r, user)
					return
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
			if err != nil {
				LocalError("Upload failed [File Creation Failed]", w, r, user)
				return
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile)
			if err != nil {
				LocalError("Upload failed [Copy Failed]", w, r, user)
				return
			}
		}
	}

	_, err = set_avatar_stmt.Exec("."+ext, strconv.Itoa(user.ID))
	if err != nil {
		InternalError(err, w)
		return
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	err = users.Load(user.ID)
	if err != nil {
		LocalError("This user no longer exists!", w, r, user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList, "Your avatar was successfully updated")
	pi := Page{"Edit Avatar", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_avatar"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-avatar.html", pi)
}

func route_account_own_edit_username(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	pi := Page{"Edit Username", user, headerVars, tList, user.Name}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
}

func route_account_own_edit_username_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	newUsername := html.EscapeString(r.PostFormValue("account-new-username"))
	_, err = set_username_stmt.Exec(newUsername, strconv.Itoa(user.ID))
	if err != nil {
		LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
		return
	}

	// TODO: Use the reloaded data instead for the name?
	user.Name = newUsername
	err = users.Load(user.ID)
	if err != nil {
		LocalError("Your account doesn't exist!", w, r, user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList, "Your username was successfully updated")
	pi := Page{"Edit Username", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_account_own_edit_username"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-username.html", pi)
}

func route_account_own_edit_email(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}

	email := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := get_emails_by_user_stmt.Query(user.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			log.Fatal(err)
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user.
	// This also applies when the admin switches site.EnableEmails on after having it off for a while.
	if len(emailList) == 0 {
		email.Email = user.Email
		email.Validated = false
		email.Primary = true
		emailList = append(emailList, email)
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	pi := Page{"Email Manager", user, headerVars, emailList, nil}
	if preRenderHooks["pre_render_account_own_edit_email"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
}

func route_account_own_edit_email_token_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.", w, r, user)
		return
	}
	token := r.URL.Path[len("/user/edit/token/"):]

	email := Email{UserID: user.ID}
	targetEmail := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := get_emails_by_user_stmt.Query(user.ID)
	if err != nil {
		InternalError(err, w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			InternalError(err, w)
			return
		}

		if email.Email == user.Email {
			email.Primary = true
		}
		if email.Token == token {
			targetEmail = email
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err, w)
		return
	}

	if len(emailList) == 0 {
		LocalError("A verification email was never sent for you!", w, r, user)
		return
	}
	if targetEmail.Token == "" {
		LocalError("That's not a valid token!", w, r, user)
		return
	}

	_, err = verify_email_stmt.Exec(user.Email)
	if err != nil {
		InternalError(err, w)
		return
	}

	// If Email Activation is on, then activate the account while we're here
	if headerVars.Settings["activation_type"] == 2 {
		_, err = activate_user_stmt.Exec(user.ID)
		if err != nil {
			InternalError(err, w)
			return
		}
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList, "The mail system is currently disabled.")
	}
	headerVars.NoticeList = append(headerVars.NoticeList, "Your email was successfully verified")
	pi := Page{"Email Manager", user, headerVars, emailList, nil}
	if preRenderHooks["pre_render_account_own_edit_email"] != nil {
		if runPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "account-own-edit-email.html", pi)
}

// TODO: Move this into member_routes.go
func route_logout(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Loggedin {
		LocalError("You can't logout without logging in first.", w, r, user)
		return
	}
	auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func route_login(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.", w, r, user)
		return
	}
	pi := Page{"Login", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_login"] != nil {
		if runPreRenderHook("pre_render_login", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "login.html", pi)
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func route_login_submit(w http.ResponseWriter, r *http.Request, user User) {
	if user.Loggedin {
		LocalError("You're already logged in.", w, r, user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	uid, err := auth.Authenticate(html.EscapeString(r.PostFormValue("username")), r.PostFormValue("password"))
	if err != nil {
		LocalError(err.Error(), w, r, user)
		return
	}

	userPtr, err := users.CascadeGet(uid)
	if err != nil {
		LocalError("Bad account", w, r, user)
		return
	}
	user = *userPtr

	var session string
	if user.Session == "" {
		session, err = auth.CreateSession(uid)
		if err != nil {
			InternalError(err, w)
			return
		}
	} else {
		session = user.Session
	}

	auth.SetCookies(w, uid, session)
	if user.IsAdmin {
		// Is this error check reundant? We already check for the error in PreRoute for the same IP
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			InternalError(err, w)
			return
		}
		log.Print("#" + strconv.Itoa(uid) + " has logged in with IP " + host)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func route_register(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w, r, &user)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.", w, r, user)
		return
	}
	pi := Page{"Registration", user, headerVars, tList, nil}
	if preRenderHooks["pre_render_register"] != nil {
		if runPreRenderHook("pre_render_register", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w, "register.html", pi)
}

func route_register_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerLite, _ := SimpleSessionCheck(w, r, &user)

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return
	}

	username := html.EscapeString(r.PostFormValue("username"))
	if username == "" {
		LocalError("You didn't put in a username.", w, r, user)
		return
	}
	email := html.EscapeString(r.PostFormValue("email"))
	if email == "" {
		LocalError("You didn't put in an email.", w, r, user)
		return
	}

	password := r.PostFormValue("password")
	if password == "" {
		LocalError("You didn't put in a password.", w, r, user)
		return
	}

	if password == username {
		LocalError("You can't use your username as your password.", w, r, user)
		return
	}

	if password == email {
		LocalError("You can't use your email as your password.", w, r, user)
		return
	}

	err = weakPassword(password)
	if err != nil {
		LocalError(err.Error(), w, r, user)
		return
	}

	confirmPassword := r.PostFormValue("confirm_password")
	log.Print("Registration Attempt! Username: " + username) // TODO: Add controls over what is logged when?

	// Do the two inputted passwords match..?
	if password != confirmPassword {
		LocalError("The two passwords don't match.", w, r, user)
		return
	}

	var active, group int
	switch headerLite.Settings["activation_type"] {
	case 1: // Activate All
		active = 1
		group = config.DefaultGroup
	default: // Anything else. E.g. Admin Activation or Email Activation.
		group = config.ActivationGroup
	}

	uid, err := users.CreateUser(username, password, email, group, active)
	if err == errAccountExists {
		LocalError("This username isn't available. Try another.", w, r, user)
		return
	} else if err != nil {
		InternalError(err, w)
		return
	}

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if site.EnableEmails {
		token, err := GenerateSafeString(80)
		if err != nil {
			InternalError(err, w)
			return
		}
		_, err = add_email_stmt.Exec(email, uid, 0, token)
		if err != nil {
			InternalError(err, w)
			return
		}

		if !SendValidationEmail(username, email, token) {
			LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.", w, r, user)
			return
		}
	}

	session, err := auth.CreateSession(uid)
	if err != nil {
		InternalError(err, w)
		return
	}

	auth.SetCookies(w, uid, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// TODO: Set the cookie domain
func route_change_theme(w http.ResponseWriter, r *http.Request, user User) {
	//headerLite, _ := SimpleSessionCheck(w, r, &user)
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form", w, r)
		return
	}

	// TODO: Rename isJs to something else, just in case we rewrite the JS side in WebAssembly?
	isJs := (r.PostFormValue("isJs") == "1")

	newTheme := html.EscapeString(r.PostFormValue("newTheme"))

	theme, ok := themes[newTheme]
	if !ok || theme.HideFromThemes {
		log.Print("Bad Theme: ", newTheme)
		LocalErrorJSQ("That theme doesn't exist", w, r, user, isJs)
		return
	}

	// TODO: Store the current theme in the user's account?
	/*if user.Loggedin {
		_, err = change_theme_stmt.Exec(newTheme, user.ID)
		if err != nil {
			InternalError(err, w)
			return
		}
	}*/

	cookie := http.Cookie{Name: "current_theme", Value: newTheme, Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)

	if !isJs {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
}

// TODO: We don't need support XML here to support sitemaps, we could handle those elsewhere
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)

func route_api(w http.ResponseWriter, r *http.Request, user User) {
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		PreErrorJS("Bad Form", w, r)
		return
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		PreErrorJS("Invalid Action", w, r)
		return
	}

	module := r.FormValue("module")
	switch module {
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			PreErrorJS("Invalid asid", w, r)
			return
		}

		_, err = delete_activity_stream_match_stmt.Exec(user.ID, asid)
		if err != nil {
			InternalError(err, w)
			return
		}
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			w.Write(phraseLoginAlerts)
			return
		}

		var msglist, event, elementType string
		var asid, actorID, targetUserID, elementID int
		var msgCount int

		err = get_activity_count_by_watcher_stmt.QueryRow(user.ID).Scan(&msgCount)
		if err == ErrNoRows {
			PreErrorJS("Couldn't find the parent topic", w, r)
			return
		} else if err != nil {
			InternalErrorJS(err, w, r)
			return
		}

		rows, err := get_activity_feed_by_watcher_stmt.Query(user.ID)
		if err != nil {
			InternalErrorJS(err, w, r)
			return
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&asid, &actorID, &targetUserID, &event, &elementType, &elementID)
			if err != nil {
				InternalErrorJS(err, w, r)
				return
			}
			res, err := buildAlert(asid, event, elementType, actorID, targetUserID, elementID, user)
			if err != nil {
				LocalErrorJS(err.Error(), w, r)
				return
			}
			msglist += res + ","
		}

		err = rows.Err()
		if err != nil {
			InternalErrorJS(err, w, r)
			return
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
		PreErrorJS("Invalid Module", w, r)
	}
}
