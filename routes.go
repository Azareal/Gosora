/* Copyright Azareal 2016 - 2017 */
package main

import (
	"log"
	//"fmt"
	"strconv"
	"bytes"
	"regexp"
	"strings"
	"time"
	"io"
	"os"
	"net"
	"net/http"
	"html"

	"./query_gen/lib"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var nList []string
var hvars HeaderVars
var extData ExtData
var success_json_bytes []byte = []byte(`{"success":"1"}`)

func init() {
	hvars.Site = site
}

// GET functions
func route_static(w http.ResponseWriter, r *http.Request){
	//log.Print("Outputting static file '" + r.URL.Path + "'")
	file, ok := static_files[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Surely, there's a more efficient way of doing this?
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && file.Info.ModTime().Before(t.Add(1 * time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h := w.Header()
	h.Set("Last-Modified", file.FormattedModTime)
	h.Set("Content-Type", file.Mimetype)
	//http.ServeContent(w,r,r.URL.Path,file.Info.ModTime(),file)
	//w.Write(file.Data)
	if strings.Contains(r.Header.Get("Accept-Encoding"),"gzip") {
		h.Set("Content-Encoding","gzip")
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

// TO-DO: Make this a static file somehow? Is it possible for us to put this file somewhere else?
// TO-DO: Add a sitemap
// TO-DO: Add an API so that plugins can register disallowed areas. E.g. /groups/join for plugin_socialgroups
func route_robots_txt(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`User-agent: *
Disallow: /panel/
Disallow: /topics/create/
Disallow: /user/edit/
Disallow: /accounts/
`))
}

func route_overview(w http.ResponseWriter, r *http.Request, user User){
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	BuildWidgets("overview",nil,&headerVars,r)

	pi := Page{"Overview",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_overview"] != nil {
		if run_pre_render_hook("pre_render_overview", w, r, &user, &pi) {
			return
		}
	}

	err := templates.ExecuteTemplate(w,"overview.html",pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_custom_page(w http.ResponseWriter, r *http.Request, user User){
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}

	name := r.URL.Path[len("/pages/"):]
	if templates.Lookup("page_" + name) == nil {
		NotFound(w,r)
		return
	}
	BuildWidgets("custom_page",name,&headerVars,r)

	pi := Page{"Page",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_custom_page"] != nil {
		if run_pre_render_hook("pre_render_custom_page", w, r, &user, &pi) {
			return
		}
	}

	err := templates.ExecuteTemplate(w,"page_" + name,pi)
	if err != nil {
		InternalError(err,w,r)
	}
}

func route_topics(w http.ResponseWriter, r *http.Request, user User){
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	BuildWidgets("topics",nil,&headerVars,r)

	var qlist string
	var fidList []interface{}
	group := groups[user.Group]
	for _, fid := range group.CanSee {
		if fstore.DirtyGet(fid).Name != "" {
			fidList = append(fidList,strconv.Itoa(fid))
			qlist += "?,"
		}
	}
	qlist = qlist[0:len(qlist) - 1]

	var topicList []*TopicsRow
	//stmt, err := qgen.Builder.SimpleLeftJoin("topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, topics.postCount, topics.likeCount, users.name, users.avatar","topics.createdBy = users.uid","parentID IN("+qlist+")","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC","")
	stmt, err := qgen.Builder.SimpleSelect("topics","tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount","parentID IN("+qlist+")","sticky DESC, lastReplyAt DESC, createdBy DESC","")
	if err != nil {
		InternalError(err,w,r)
		return
	}

	rows, err := stmt.Query(fidList...)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	var reqUserList map[int]bool = make(map[int]bool)
	for rows.Next() {
		topicItem := TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		topicItem.Link = build_topic_url(name_to_slug(topicItem.Title),topicItem.ID)

		forum := fstore.DirtyGet(topicItem.ParentID)
		if topicItem.ParentID >= 0 {
			topicItem.ForumName = forum.Name
			topicItem.ForumLink = forum.Link
		} else {
			topicItem.ForumName = ""
		}

		/*topicItem.CreatedAt, err = relative_time(topicItem.CreatedAt)
		if err != nil {
			replyItem.CreatedAt = ""
		}*/
		topicItem.LastReplyAt, err = relative_time(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err,w,r)
		}

		if hooks["topics_topic_row_assign"] != nil {
			run_vhook("topics_topic_row_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice []int = make([]int,len(reqUserList))
	var i int
	for userID, _ := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TO-DO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkCascadeGetMap(idSlice)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Second pass to the add the user data
	// TO-DO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := TopicsPage{"Topic List",user,headerVars,topicList,extData}
	if pre_render_hooks["pre_render_topic_list"] != nil {
		if run_pre_render_hook("pre_render_topic_list", w, r, &user, &pi) {
			return
		}
	}

	if template_topics_handle != nil {
		template_topics_handle(pi,w)
	} else {
		mapping, ok := themes[defaultTheme].TemplatesMap["topics"]
		if !ok {
			mapping = "topics"
		}
		err = templates.ExecuteTemplate(w,mapping + ".html", pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_forum(w http.ResponseWriter, r *http.Request, user User, sfid string){
	page, _ := strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(sfid,".")
	if len(halves) < 2 {
		halves = append(halves,halves[0])
	}
	fid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("The provided ForumID is not a valid number.",w,r)
		return
	}

	headerVars, ok := ForumSessionCheck(w,r,&user,fid)
	if !ok {
		return
	}
	//fmt.Printf("%+v\n", groups[user.Group].Forums)
	if !user.Perms.ViewTopic {
		NoPermissions(w,r,user)
		return
	}

	// TO-DO: Fix this double-check
	forum, err := fstore.CascadeGet(fid)
	if err == ErrNoRows {
		NotFound(w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	BuildWidgets("view_forum",forum,&headerVars,r)

	// Calculate the offset
	var offset int
	last_page := int(forum.TopicCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = last_page
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}
	rows, err := get_forum_topics_offset_stmt.Query(fid,offset,config.ItemsPerPage)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	// TO-DO: Use something other than TopicsRow as we don't need to store the forum name and link on each and every topic item?
	var topicList []*TopicsRow
	var reqUserList map[int]bool = make(map[int]bool)
	for rows.Next() {
		var topicItem TopicsRow = TopicsRow{ID: 0}
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.LastReplyBy, &topicItem.ParentID, &topicItem.PostCount, &topicItem.LikeCount)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		topicItem.Link = build_topic_url(name_to_slug(topicItem.Title),topicItem.ID)
		topicItem.LastReplyAt, err = relative_time(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err,w,r)
		}

		if hooks["forum_trow_assign"] != nil {
			run_vhook("forum_trow_assign", &topicItem, &forum)
		}
		topicList = append(topicList, &topicItem)
		reqUserList[topicItem.CreatedBy] = true
		reqUserList[topicItem.LastReplyBy] = true
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Convert the user ID map to a slice, then bulk load the users
	var idSlice []int = make([]int,len(reqUserList))
	var i int
	for userID, _ := range reqUserList {
		idSlice[i] = userID
		i++
	}

	// TO-DO: What if a user is deleted via the Control Panel?
	userList, err := users.BulkCascadeGetMap(idSlice)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Second pass to the add the user data
	// TO-DO: Use a pointer to TopicsRow instead of TopicsRow itself?
	for _, topicItem := range topicList {
		topicItem.Creator = userList[topicItem.CreatedBy]
		topicItem.LastUser = userList[topicItem.LastReplyBy]
	}

	pi := ForumPage{forum.Name,user,headerVars,topicList,*forum,page,last_page,extData}
	if pre_render_hooks["pre_render_view_forum"] != nil {
		if run_pre_render_hook("pre_render_view_forum", w, r, &user, &pi) {
			return
		}
	}

	if template_forum_handle != nil {
		template_forum_handle(pi,w)
	} else {
		mapping, ok := themes[defaultTheme].TemplatesMap["forum"]
		if !ok {
			mapping = "forum"
		}
		err = templates.ExecuteTemplate(w,mapping + ".html", pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_forums(w http.ResponseWriter, r *http.Request, user User){
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	BuildWidgets("forums",nil,&headerVars,r)

	var err error
	var forumList []Forum
	var canSee []int
	if user.Is_Super_Admin {
		canSee, err = fstore.GetAllIDs()
		if err != nil {
			InternalError(err,w,r)
			return
		}
		//fmt.Println("canSee",canSee)
	} else {
		group := groups[user.Group]
		canSee = group.CanSee
		//fmt.Println("group.CanSee",group.CanSee)
	}

	for _, fid := range canSee {
		//fmt.Println(forums[fid])
		var forum Forum = *fstore.DirtyGet(fid)
		if forum.Active && forum.Name != "" && forum.ParentID == 0 {
			if forum.LastTopicID != 0 {
				forum.LastTopicTime, err = relative_time(forum.LastTopicTime)
				if err != nil {
					InternalError(err,w,r)
				}
			} else {
				forum.LastTopic = "None"
				forum.LastTopicTime = ""
			}
			if hooks["forums_frow_assign"] != nil {
				run_hook("forums_frow_assign", &forum)
			}
			forumList = append(forumList, forum)
		}
	}

	pi := ForumsPage{"Forum List",user,headerVars,forumList,extData}
	if pre_render_hooks["pre_render_forum_list"] != nil {
		if run_pre_render_hook("pre_render_forum_list", w, r, &user, &pi) {
			return
		}
	}

	if template_forums_handle != nil {
		template_forums_handle(pi,w)
	} else {
		mapping, ok := themes[defaultTheme].TemplatesMap["forums"]
		if !ok {
			mapping = "forums"
		}
		err = templates.ExecuteTemplate(w,mapping + ".html", pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_topic_id(w http.ResponseWriter, r *http.Request, user User){
	var err error
	var page, offset int
	var replyList []Reply

	page, _ = strconv.Atoi(r.FormValue("page"))

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/topic/"):],".")
	if len(halves) < 2 {
		halves = append(halves,halves[0])
	}

	tid, err := strconv.Atoi(halves[1])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}

	// Get the topic...
	topic, err := get_topicuser(tid)
	if err == ErrNoRows {
		NotFound(w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	topic.ClassName = ""

	headerVars, ok := ForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic {
		//fmt.Printf("%+v\n", user.Perms)
		NoPermissions(w,r,user)
		return
	}

	BuildWidgets("view_topic",&topic,&headerVars,r)

	topic.Content = parse_message(topic.Content)
	topic.ContentLines = strings.Count(topic.Content,"\n")

	// We don't want users posting in locked topics...
	if topic.Is_Closed && !user.Is_Mod {
		user.Perms.CreateReply = false
	}

	topic.Tag = groups[topic.Group].Tag
	if groups[topic.Group].Is_Mod || groups[topic.Group].Is_Admin {
		topic.ClassName = config.StaffCss
	}

	/*if settings["url_tags"] == false {
		topic.URLName = ""
	} else {
		topic.URL, ok = external_sites[topic.URLPrefix]
		if !ok {
			topic.URL = topic.URLName
		} else {
			topic.URL = topic.URL + topic.URLName
		}
	}*/

	topic.CreatedAt, err = relative_time(topic.CreatedAt)
	if err != nil {
		topic.CreatedAt = ""
	}

	// Calculate the offset
	last_page := int(topic.PostCount / config.ItemsPerPage) + 1
	if page > 1 {
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else if page == -1 {
		page = last_page
		offset = (config.ItemsPerPage * page) - config.ItemsPerPage
	} else {
		page = 1
	}

	// Get the replies..
	rows, err := get_topic_replies_offset_stmt.Query(topic.ID, offset, config.ItemsPerPage)
	if err == ErrNoRows {
		LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	replyItem := Reply{ClassName:""}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IpAddress, &replyItem.LikeCount, &replyItem.ActionType)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		replyItem.UserLink = build_profile_url(name_to_slug(replyItem.CreatedByName),replyItem.CreatedBy)
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = parse_message(replyItem.Content)
		replyItem.ContentLines = strings.Count(replyItem.Content,"\n")

		if groups[replyItem.Group].Is_Mod || groups[replyItem.Group].Is_Admin {
			replyItem.ClassName = config.StaffCss
		} else {
			replyItem.ClassName = ""
		}

		if replyItem.Avatar != "" {
			if replyItem.Avatar[0] == '.' {
				replyItem.Avatar = "/uploads/avatar_" + strconv.Itoa(replyItem.CreatedBy) + replyItem.Avatar
			}
		} else {
			replyItem.Avatar = strings.Replace(config.Noavatar,"{id}",strconv.Itoa(replyItem.CreatedBy),1)
		}

		replyItem.Tag = groups[replyItem.Group].Tag

		/*if settings["url_tags"] == false {
			replyItem.URLName = ""
		} else {
			replyItem.URL, ok = external_sites[replyItem.URLPrefix]
			if !ok {
				replyItem.URL = replyItem.URLName
			} else {
				replyItem.URL = replyItem.URL + replyItem.URLName
			}
		}*/

		replyItem.CreatedAt, err = relative_time(replyItem.CreatedAt)
		if err != nil {
			replyItem.CreatedAt = ""
		}

		// We really shouldn't have inline HTML, we should do something about this...
		if replyItem.ActionType != "" {
			switch(replyItem.ActionType) {
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

		// TO-DO: Rename this to topic_rrow_assign
		if hooks["rrow_assign"] != nil {
			run_hook("rrow_assign", &replyItem)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	tpage := TopicPage{topic.Title,user,headerVars,replyList,topic,page,last_page,extData}
	if pre_render_hooks["pre_render_view_topic"] != nil {
		if run_pre_render_hook("pre_render_view_topic", w, r, &user, &tpage) {
			return
		}
	}

	if template_topic_handle != nil {
		template_topic_handle(tpage,w)
	} else {
		mapping, ok := themes[defaultTheme].TemplatesMap["topic"]
		if !ok {
			mapping = "topic"
		}
		err = templates.ExecuteTemplate(w,mapping + ".html", tpage)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_profile(w http.ResponseWriter, r *http.Request, user User){
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}

	var err error
	var replyContent, replyCreatedByName, replyCreatedAt, replyAvatar, replyTag, replyClassName string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyList []Reply

	// SEO URLs...
	halves := strings.Split(r.URL.Path[len("/user/"):],".")
	if len(halves) < 2 {
		halves = append(halves,halves[0])
	}

	pid, err := strconv.Atoi(halves[1])
	if err != nil {
		LocalError("The provided User ID is not a valid number.",w,r,user)
		return
	}

	var puser *User
	if pid == user.ID {
		user.Is_Mod = true
		puser = &user
	} else {
		// Fetch the user data
		puser, err = users.CascadeGet(pid)
		if err == ErrNoRows {
			NotFound(w,r)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
	}

	// Get the replies..
	rows, err := get_profile_replies_stmt.Query(puser.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		replyLines = strings.Count(replyContent,"\n")
		if groups[replyGroup].Is_Mod || groups[replyGroup].Is_Admin {
			replyClassName = config.StaffCss
		} else {
			replyClassName = ""
		}
		if replyAvatar != "" {
			if replyAvatar[0] == '.' {
				replyAvatar = "/uploads/avatar_" + strconv.Itoa(replyCreatedBy) + replyAvatar
			}
		} else {
			replyAvatar = strings.Replace(config.Noavatar,"{id}",strconv.Itoa(replyCreatedBy),1)
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

		// TO-DO: Add a hook here

		replyList = append(replyList, Reply{rid,puser.ID,replyContent,parse_message(replyContent),replyCreatedBy,build_profile_url(name_to_slug(replyCreatedByName),replyCreatedBy),replyCreatedByName,replyGroup,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyClassName,replyLines,replyTag,"","","",0,"",replyLiked,replyLikeCount,"",""})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	ppage := ProfilePage{puser.Name + "'s Profile",user,headerVars,replyList,*puser,extData}
	if pre_render_hooks["pre_render_profile"] != nil {
		if run_pre_render_hook("pre_render_profile", w, r, &user, &ppage) {
			return
		}
	}

	if template_profile_handle != nil {
		template_profile_handle(ppage,w)
	} else {
		err = templates.ExecuteTemplate(w,"profile.html",ppage)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_topic_create(w http.ResponseWriter, r *http.Request, user User, sfid string){
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			PreError("The provided ForumID is not a valid number.",w,r)
			return
		}
	}

	headerVars, ok := ForumSessionCheck(w,r,&user,fid)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w,r,user)
		return
	}

	BuildWidgets("create_topic",nil,&headerVars,r)

	// Lock this to the forum being linked?
	// Should we always put it in strictmode when it's linked from another forum? Well, the user might end up changing their mind on what forum they want to post in and it would be a hassle, if they had to switch pages, even if it is a single click for many (exc. mobile)
	var strictmode bool
	if vhooks["topic_create_pre_loop"] != nil {
		run_vhook("topic_create_pre_loop", w, r, fid, &headerVars, &user, &strictmode)
	}

	var forumList []Forum
	var canSee []int
	if user.Is_Super_Admin {
		canSee, err = fstore.GetAllIDs()
		if err != nil {
			InternalError(err,w,r)
			return
		}
	} else {
		group := groups[user.Group]
		canSee = group.CanSee
	}

	// TO-DO: plugin_superadmin needs to be able to override this loop. Skip flag on topic_create_pre_loop?
	for _, ffid := range canSee {
		// TO-DO: Surely, there's a better way of doing this. I've added it in for now to support plugin_socialgroups, but we really need to clean this up
		if strictmode && ffid != fid {
			continue
		}

		// Do a bulk forum fetch, just in case it's the SqlForumStore?
		forum := fstore.DirtyGet(ffid)
		if forum.Active && forum.Name != "" {
			fcopy := *forum
			if hooks["topic_create_frow_assign"] != nil {
				// TO-DO: Add the skip feature to all the other row based hooks?
				if run_hook("topic_create_frow_assign", &fcopy).(bool) {
					continue
				}
			}
			forumList = append(forumList, fcopy)
		}
	}

	ctpage := CreateTopicPage{"Create Topic",user,headerVars,forumList,fid,extData}
	if pre_render_hooks["pre_render_create_topic"] != nil {
		if run_pre_render_hook("pre_render_create_topic", w, r, &user, &ctpage) {
			return
		}
	}

	if template_create_topic_handle != nil {
		template_create_topic_handle(ctpage,w)
	} else {
		err = templates.ExecuteTemplate(w,"create-topic.html",ctpage)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

// POST functions. Authorised users only.
func route_topic_create_submit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return
	}

	fid, err := strconv.Atoi(r.PostFormValue("topic-board"))
	if err != nil {
		PreError("The provided ForumID is not a valid number.",w,r)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,fid)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w,r,user)
		return
	}

	topic_name := html.EscapeString(r.PostFormValue("topic-name"))
	content := html.EscapeString(preparse_message(r.PostFormValue("topic-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}

	wcount := word_count(content)
	res, err := create_topic_stmt.Exec(fid,topic_name,content,parse_message(content),user.ID,ipaddress,wcount,user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	err = fstore.IncrementTopicCount(fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = add_subscription_stmt.Exec(user.ID,lastId,"topic")
	if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,"/topic/" + strconv.FormatInt(lastId,10), http.StatusSeeOther)
	err = increase_post_user_stats(wcount,user.ID,true,user)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	err = fstore.UpdateLastTopic(topic_name,int(lastId),user.Name,user.ID,"",fid)
	if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
	}
}

func route_create_reply(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return
	}
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		PreError("Failed to convert the Topic ID",w,r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("Couldn't find the parent topic",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateReply {
		NoPermissions(w,r,user)
		return
	}

	content := preparse_message(html.EscapeString(r.PostFormValue("reply-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}

	wcount := word_count(content)
	_, err = create_reply_stmt.Exec(tid,content,parse_message(content),ipaddress,wcount,user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = add_replies_to_topic_stmt.Exec(1,user.ID,tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	_, err = update_forum_cache_stmt.Exec(topic.Title,tid,user.Name,user.ID,1)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID,topic.CreatedBy,"reply","topic",tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = notify_watchers_stmt.Exec(lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Alert the subscribers about this post without blocking this post from being posted
	if enable_websockets {
		go notify_watchers(lastId)
	}

	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The destination no longer exists",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	err = increase_post_user_stats(wcount, user.ID, false, user)
	if err != nil {
		InternalError(err,w,r)
		return
	}
}

func route_like_topic(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return
	}

	tid, err := strconv.Atoi(r.URL.Path[len("/topic/like/submit/"):])
	if err != nil {
		PreError("Topic IDs can only ever be numbers.",w,r)
		return
	}

	topic, err := topics.CascadeGet(tid)
	if err == ErrNoRows {
		PreError("The requested topic doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w,r,user)
		return
	}

	if topic.CreatedBy == user.ID {
		LocalError("You can't like your own topics",w,r,user)
		return
	}

	err = has_liked_topic_stmt.QueryRow(user.ID,tid).Scan(&tid)
	if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!",w,r,user)
		return
	}

	_, err = users.CascadeGet(topic.CreatedBy)
	if err != nil && err == ErrNoRows {
		LocalError("The target user doesn't exist",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score,tid,"topics",user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = add_likes_to_topic_stmt.Exec(1,tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID,topic.CreatedBy,"like","topic",tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = notify_one_stmt.Exec(topic.CreatedBy,lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = ws_hub.push_alert(topic.CreatedBy,"like","topic",user.ID,topic.CreatedBy,tid)

	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err == ErrNoRows {
		LocalError("The liked topic no longer exists",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_reply_like_submit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		PreError("Bad Form",w,r)
		return
	}

	rid, err := strconv.Atoi(r.URL.Path[len("/reply/like/submit/"):])
	if err != nil {
		PreError("The provided Reply ID is not a valid number.",w,r)
		return
	}

	reply, err := get_reply(rid)
	if err == ErrNoRows {
		PreError("You can't like something which doesn't exist!",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	var fid int
	err = get_topic_fid_stmt.QueryRow(reply.ParentID).Scan(&fid)
	if err == ErrNoRows {
		PreError("The parent topic doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	ok := SimpleForumSessionCheck(w,r,&user,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w,r,user)
		return
	}

	if reply.CreatedBy == user.ID {
		LocalError("You can't like your own replies",w,r,user)
		return
	}

	err = has_liked_reply_stmt.QueryRow(user.ID, rid).Scan(&rid)
	if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
		return
	} else if err != ErrNoRows {
		LocalError("You already liked this!",w,r,user)
		return
	}

	_, err = users.CascadeGet(reply.CreatedBy)
	if err != nil && err != ErrNoRows {
		LocalError("The target user doesn't exist",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	score := 1
	_, err = create_like_stmt.Exec(score,rid,"replies",user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = add_likes_to_reply_stmt.Exec(1,rid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	res, err := add_activity_stmt.Exec(user.ID,reply.CreatedBy,"like","post",rid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = notify_one_stmt.Exec(reply.CreatedBy,lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// Live alerts, if the poster is online and WebSockets is enabled
	_ = ws_hub.push_alert(reply.CreatedBy,"like","post",user.ID,reply.CreatedBy,rid)

	http.Redirect(w,r,"/topic/" + strconv.Itoa(reply.ParentID),http.StatusSeeOther)
}

func route_profile_reply_create(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Loggedin || !user.Perms.CreateReply {
		NoPermissions(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		LocalError("Invalid UID",w,r,user)
		return
	}

	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}

	_, err = create_profile_reply_stmt.Exec(uid,html.EscapeString(preparse_message(r.PostFormValue("reply-content"))),parse_message(html.EscapeString(preparse_message(r.PostFormValue("reply-content")))),user.ID,ipaddress)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	var user_name string
	err = get_user_name_stmt.QueryRow(uid).Scan(&user_name)
	if err == ErrNoRows {
		LocalError("The profile you're trying to post on doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w, r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
}

func route_report_submit(w http.ResponseWriter, r *http.Request, user User, sitem_id string) {
	if !user.Loggedin {
		LoginRequired(w,r,user)
		return
	}
	if user.Is_Banned {
		Banned(w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}

	item_id, err := strconv.Atoi(sitem_id)
	if err != nil {
		LocalError("Bad ID",w,r,user)
		return
	}

	item_type := r.FormValue("type")

	var fid int = 1
	var title, content string
	if item_type == "reply" {
		reply, err := get_reply(item_id)
		if err == ErrNoRows {
			LocalError("We were unable to find the reported post",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}

		topic, err := topics.CascadeGet(reply.ParentID)
		if err == ErrNoRows {
			LocalError("We weren't able to find the topic the reported post is supposed to be in",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}

		title = "Reply: " + topic.Title
		content = reply.Content + "\n\nOriginal Post: #rid-" + strconv.Itoa(item_id)
	} else if item_type == "user-reply" {
		user_reply, err := get_user_reply(item_id)
		if err == ErrNoRows {
			LocalError("We weren't able to find the reported post",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}

		err = get_user_name_stmt.QueryRow(user_reply.ParentID).Scan(&title)
		if err == ErrNoRows {
			LocalError("We weren't able to find the profile the reported post is supposed to be on",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		title = "Profile: " + title
		content = user_reply.Content + "\n\nOriginal Post: @" + strconv.Itoa(user_reply.ParentID)
	} else if item_type == "topic" {
		err = get_topic_basic_stmt.QueryRow(item_id).Scan(&title,&content)
		if err == ErrNoRows {
			NotFound(w,r)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		title = "Topic: " + title
		content = content + "\n\nOriginal Post: #tid-" + strconv.Itoa(item_id)
	} else {
		if vhooks["report_preassign"] != nil {
			run_vhook_noreturn("report_preassign", &item_id, &item_type)
			return
		}
		// Don't try to guess the type
		LocalError("Unknown type",w,r,user)
		return
	}

	var count int
	rows, err := report_exists_stmt.Query(item_type + "_" + strconv.Itoa(item_id))
	if err != nil && err != ErrNoRows {
		InternalError(err,w,r)
		return
	}

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	}
	if count != 0 {
		LocalError("Someone has already reported this!",w,r,user)
		return
	}

	res, err := create_report_stmt.Exec(title,content,parse_message(content),user.ID,item_type + "_" + strconv.Itoa(item_id))
	if err != nil {
		InternalError(err,w,r)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}

	_, err = add_topics_to_forum_stmt.Exec(1, fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	_, err = update_forum_cache_stmt.Exec(title, lastId, user.Name, user.ID, fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	http.Redirect(w,r,"/topic/" + strconv.FormatInt(lastId, 10), http.StatusSeeOther)
}

func route_account_own_edit_critical(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}

	pi := Page{"Edit Password",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_account_own_edit_critical"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_critical_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}

	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}

	var real_password, salt string
	current_password := r.PostFormValue("account-current-password")
	new_password := r.PostFormValue("account-new-password")
	confirm_password := r.PostFormValue("account-confirm-password")

	err = get_password_stmt.QueryRow(user.ID).Scan(&real_password, &salt)
	if err == ErrNoRows {
		LocalError("Your account no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	err = CheckPassword(real_password,current_password,salt)
	if err == ErrMismatchedHashAndPassword {
		LocalError("That's not the correct password.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	if new_password != confirm_password {
		LocalError("The two passwords don't match.",w,r,user)
		return
	}
	SetPassword(user.ID, new_password)

	// Log the user out as a safety precaution
	auth.ForceLogout(user.ID)

	headerVars.NoticeList = append(headerVars.NoticeList,"Your password was successfully updated")
	pi := Page{"Edit Password",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_account_own_edit_critical"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_avatar(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	pi := Page{"Edit Avatar",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_account_own_edit_avatar"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html",pi)
}

func route_account_own_edit_avatar_submit(w http.ResponseWriter, r *http.Request, user User) {
	if r.ContentLength > int64(config.MaxRequestSize) {
		http.Error(w,"Request too large",http.StatusExpectationFailed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxRequestSize))

	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}

	err := r.ParseMultipartForm(int64(config.MaxRequestSize))
	if  err != nil {
		LocalError("Upload failed",w,r,user)
		return
	}

	var filename string
	var ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			infile, err := hdr.Open();
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
				extarr := strings.Split(hdr.Filename,".")
				if len(extarr) < 2 {
					LocalError("Bad file", w, r, user)
					return
				}
				ext = extarr[len(extarr) - 1]

				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					LocalError("Bad file extension", w, r, user)
					return
				}
				ext = reg.ReplaceAllString(ext,"")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext);
			if  err != nil {
				LocalError("Upload failed [File Creation Failed]",w,r,user)
				return
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile);
			if  err != nil {
				LocalError("Upload failed [Copy Failed]",w,r,user)
				return
			}
		}
	}

	_, err = set_avatar_stmt.Exec("." + ext, strconv.Itoa(user.ID))
	if err != nil {
		InternalError(err,w,r)
		return
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	err = users.Load(user.ID)
	if err != nil {
		LocalError("This user no longer exists!",w,r,user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList, "Your avatar was successfully updated")
	pi := Page{"Edit Avatar",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_account_own_edit_avatar"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html", pi)
}

func route_account_own_edit_username(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	pi := Page{"Edit Username",user,headerVars,tList,user.Name}
	if pre_render_hooks["pre_render_account_own_edit_username"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-username.html",pi)
}

func route_account_own_edit_username_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}

	new_username := html.EscapeString(r.PostFormValue("account-new-username"))
	_, err = set_username_stmt.Exec(new_username, strconv.Itoa(user.ID))
	if err != nil {
		LocalError("Unable to change the username. Does someone else already have this name?",w,r,user)
		return
	}

	user.Name = new_username
	err = users.Load(user.ID)
	if err != nil {
		LocalError("Your account doesn't exist!",w,r,user)
		return
	}

	headerVars.NoticeList = append(headerVars.NoticeList,"Your username was successfully updated")
	pi := Page{"Edit Username",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_account_own_edit_username"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_username", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_account_own_edit_email(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
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
		headerVars.NoticeList = append(headerVars.NoticeList,"The mail system is currently disabled.")
	}
	pi := Page{"Email Manager",user,headerVars,emailList,nil}
	if pre_render_hooks["pre_render_account_own_edit_email"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-email.html", pi)
}

func route_account_own_edit_email_token_submit(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	token := r.URL.Path[len("/user/edit/token/"):]

	email := Email{UserID: user.ID}
	targetEmail := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := get_emails_by_user_stmt.Query(user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated, &email.Token)
		if err != nil {
			InternalError(err,w,r)
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
		InternalError(err,w,r)
		return
	}

	if len(emailList) == 0 {
		LocalError("A verification email was never sent for you!",w,r,user)
		return
	}
	if targetEmail.Token == "" {
		LocalError("That's not a valid token!",w,r,user)
		return
	}

	_, err = verify_email_stmt.Exec(user.Email)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	// If Email Activation is on, then activate the account while we're here
	if settings["activation_type"] == 2 {
		_, err = activate_user_stmt.Exec(user.ID)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	}

	if !site.EnableEmails {
		headerVars.NoticeList = append(headerVars.NoticeList,"The mail system is currently disabled.")
	}
	headerVars.NoticeList = append(headerVars.NoticeList,"Your email was successfully verified")
	pi := Page{"Email Manager",user,headerVars,emailList,nil}
	if pre_render_hooks["pre_render_account_own_edit_email"] != nil {
		if run_pre_render_hook("pre_render_account_own_edit_email", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"account-own-edit-email.html", pi)
}

func route_logout(w http.ResponseWriter, r *http.Request, user User) {
	if !user.Loggedin {
		LocalError("You can't logout without logging in first.",w,r,user)
		return
	}
	auth.Logout(w, user.ID)
	http.Redirect(w,r, "/", http.StatusSeeOther)
}

func route_login(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	pi := Page{"Login",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_login"] != nil {
		if run_pre_render_hook("pre_render_login", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"login.html",pi)
}

func route_login_submit(w http.ResponseWriter, r *http.Request, user User) {
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}

	uid, err := auth.Authenticate(html.EscapeString(r.PostFormValue("username")), r.PostFormValue("password"))
	if err != nil {
		LocalError(err.Error(),w,r,user)
		return
	}

	var session string
	if user.Session == "" {
		session, err = auth.CreateSession(uid)
		if err != nil {
			InternalError(err,w,r)
			return
		}
	} else {
		session = user.Session
	}

	auth.SetCookies(w,uid,session)
	http.Redirect(w,r,"/",http.StatusSeeOther)
}

func route_register(w http.ResponseWriter, r *http.Request, user User) {
	headerVars, ok := SessionCheck(w,r,&user)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	pi := Page{"Registration",user,headerVars,tList,nil}
	if pre_render_hooks["pre_render_register"] != nil {
		if run_pre_render_hook("pre_render_register", w, r, &user, &pi) {
			return
		}
	}
	templates.ExecuteTemplate(w,"register.html",pi)
}

func route_register_submit(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return
	}

	username := html.EscapeString(r.PostFormValue("username"))
	if username == "" {
		LocalError("You didn't put in a username.",w,r,user)
		return
	}
	email := html.EscapeString(r.PostFormValue("email"))
	if email == "" {
		LocalError("You didn't put in an email.",w,r,user)
		return
	}

	password := r.PostFormValue("password")
	if password == "" {
		LocalError("You didn't put in a password.",w,r,user)
		return
	}

	if password == username {
		LocalError("You can't use your username as your password.",w,r,user)
		return
	}

	if password == email {
		LocalError("You can't use your email as your password.",w,r,user)
		return
	}

	err = weak_password(password)
	if err != nil {
		LocalError(err.Error(),w,r,user)
		return
	}

	confirm_password := r.PostFormValue("confirm_password")
	log.Print("Registration Attempt! Username: " + username)

	// Do the two inputted passwords match..?
	if password != confirm_password {
		LocalError("The two passwords don't match.",w,r,user)
		return
	}

	var active, group int
	switch settings["activation_type"] {
		case 1: // Activate All
			active = 1
			group = config.DefaultGroup
		default: // Anything else. E.g. Admin Activation or Email Activation.
			group = config.ActivationGroup
	}

	uid, err := users.CreateUser(username, password, email, group, active)
	if err == err_account_exists {
		LocalError("This username isn't available. Try another.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if site.EnableEmails {
		token, err := GenerateSafeString(80)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		_, err = add_email_stmt.Exec(email, uid, 0, token)
		if err != nil {
			InternalError(err,w,r)
			return
		}

		if !SendValidationEmail(username, email, token) {
			LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.",w,r,user)
			return
		}
	}

	session, err := auth.CreateSession(uid)
	if err != nil {
		InternalError(err,w,r)
		return
	}

	auth.SetCookies(w,uid,session)
	http.Redirect(w,r,"/",http.StatusSeeOther)
}

var phrase_login_alerts []byte = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)
func route_api(w http.ResponseWriter, r *http.Request, user User) {
	err := r.ParseForm()
	format := r.FormValue("format")
	// TO-DO: Change is_js from a string to a boolean value
	var is_js string
	if format == "json" {
		is_js = "1"
	} else { // html
		is_js = "0"
	}
	if err != nil {
		PreErrorJSQ("Bad Form",w,r,is_js)
		return
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		PreErrorJSQ("Invalid Action",w,r,is_js)
		return
	}

	module := r.FormValue("module")
	switch(module) {
		case "alerts": // A feed of events tailored for a specific user
			if format != "json" {
				PreError("You can only fetch alerts in the JSON format!",w,r)
				return
			}

			w.Header().Set("Content-Type","application/json")
			if !user.Loggedin {
				w.Write(phrase_login_alerts)
				return
			}

			var msglist, event, elementType string
			var asid, actor_id, targetUser_id, elementID int
			var msgCount int

			err = get_activity_count_by_watcher_stmt.QueryRow(user.ID).Scan(&msgCount)
			if err == ErrNoRows {
				PreError("Couldn't find the parent topic",w,r)
				return
			} else if err != nil {
				InternalError(err,w,r)
				return
			}

			rows, err := get_activity_feed_by_watcher_stmt.Query(user.ID)
			if err != nil {
				InternalErrorJS(err,w,r)
				return
			}
			defer rows.Close()

			for rows.Next() {
				err = rows.Scan(&asid,&actor_id,&targetUser_id,&event,&elementType,&elementID)
				if err != nil {
					InternalErrorJS(err,w,r)
					return
				}
				res, err := build_alert(event, elementType, actor_id, targetUser_id, elementID, user)
				if err != nil {
					LocalErrorJS(err.Error(),w,r)
					return
				}
				msglist += res + ","
			}

			err = rows.Err()
			if err != nil {
				InternalErrorJS(err,w,r)
				return
			}

			if len(msglist) != 0 {
				msglist = msglist[0:len(msglist)-1]
			}
			w.Write([]byte(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`))
			//fmt.Println(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`)
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
			PreErrorJSQ("Invalid Module",w,r,is_js)
	}
}
