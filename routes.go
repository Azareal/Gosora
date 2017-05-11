/* Copyright Azareal 2016 - 2017 */
package main

import (
	"log"
//	"fmt"
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
	"html/template"
	"database/sql"
)

import _ "github.com/go-sql-driver/mysql"
import "golang.org/x/crypto/bcrypt"

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var nList []string
var success_json_bytes []byte = []byte(`{"success":"1"}`)

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

/*func route_exit(w http.ResponseWriter, r *http.Request){
	db.Close()
	os.Exit(0)
}

func route_fstatic(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w,r,r.URL.Path)
}*/

func route_overview(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	pi := Page{"Overview",user,noticeList,tList,nil}
	err := templates.ExecuteTemplate(w,"overview.html",pi)
    if err != nil {
        InternalError(err,w,r)
    }
}

func route_custom_page(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	name := r.URL.Path[len("/pages/"):]
	if templates.Lookup("page_" + name) == nil {
		NotFound(w,r)
		return
	}
	
	err := templates.ExecuteTemplate(w,"page_" + name,Page{"Page",user,noticeList,tList,nil})
	if err != nil {
		InternalError(err,w,r)
	}
}
	
func route_topics(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var fidList []string
	group := groups[user.Group]
	for _, fid := range group.CanSee {
		if forums[fid].Name != "" {
			fidList = append(fidList,strconv.Itoa(fid))
		}
	}
	
	var topicList []TopicsRow
	rows, err := db.Query("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, topics.likeCount, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid where parentID in("+strings.Join(fidList,",")+") order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	//rows, err := get_topic_list_stmt.Query()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	topicItem := TopicsRow{ID: 0}
	for rows.Next() {
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.ParentID, &topicItem.LikeCount, &topicItem.CreatedByName, &topicItem.Avatar)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		if topicItem.Avatar != "" {
			if topicItem.Avatar[0] == '.' {
				topicItem.Avatar = "/uploads/avatar_" + strconv.Itoa(topicItem.CreatedBy) + topicItem.Avatar
			}
		} else {
			topicItem.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(topicItem.CreatedBy),1)
		}
		
		if topicItem.ParentID >= 0 {
			topicItem.ForumName = forums[topicItem.ParentID].Name
		} else {
			topicItem.ForumName = ""
		}
		
		/*topicItem.CreatedAt, err = relative_time(topicItem.CreatedAt)
		if err != nil {
			InternalError(err,w,r)
		}*/
		topicItem.LastReplyAt, err = relative_time(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err,w,r)
		}
		
		if hooks["trow_assign"] != nil {
			topicItem = run_hook("trow_assign", topicItem).(TopicsRow)
		}
		topicList = append(topicList, topicItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	rows.Close()
	
	pi := TopicsPage{"Topic List",user,noticeList,topicList,nil}
	if template_topics_handle != nil {
		template_topics_handle(pi,w)
	} else {
		err = templates.ExecuteTemplate(w,"topics.html",pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_forum(w http.ResponseWriter, r *http.Request, sfid string){
	page, _ := strconv.Atoi(r.FormValue("page"))
	fid, err := strconv.Atoi(sfid)
	if err != nil {
		PreError("The provided ForumID is not a valid number.",w,r)
		return
	}
	
	user, noticeList, ok := ForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	//fmt.Printf("%+v\n", groups[user.Group].Forums)
	if !user.Perms.ViewTopic {
		NoPermissions(w,r,user)
		return
	}
	
	// Calculate the offset
	var offset int
	last_page := int(forums[fid].TopicCount / items_per_page) + 1
	if page > 1 {
		offset = (items_per_page * page) - items_per_page
	} else if page == -1 {
		page = last_page
		offset = (items_per_page * page) - items_per_page
	} else {
		page = 1
	}
	rows, err := get_forum_topics_offset_stmt.Query(fid,offset)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	var topicList []TopicUser
	topicItem := TopicUser{ID: 0}
	for rows.Next() {
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.LastReplyAt, &topicItem.ParentID, &topicItem.LikeCount, &topicItem.CreatedByName, &topicItem.Avatar)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		if topicItem.Avatar != "" {
			if topicItem.Avatar[0] == '.' {
				topicItem.Avatar = "/uploads/avatar_" + strconv.Itoa(topicItem.CreatedBy) + topicItem.Avatar
			}
		} else {
			topicItem.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(topicItem.CreatedBy),1)
		}
		
		topicItem.LastReplyAt, err = relative_time(topicItem.LastReplyAt)
		if err != nil {
			InternalError(err,w,r)
		}
		
		if hooks["trow_assign"] != nil {
			topicItem = run_hook("trow_assign", topicItem).(TopicUser)
		}
		topicList = append(topicList, topicItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	rows.Close()
	
	pi := ForumPage{forums[fid].Name,user,noticeList,topicList,forums[fid],page,last_page,nil}
	if template_forum_handle != nil {
		template_forum_handle(pi,w)
	} else {
		err = templates.ExecuteTemplate(w,"forum.html",pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_forums(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var forumList []Forum
	var err error
	group := groups[user.Group]
	//fmt.Println(group.CanSee)
	for _, fid := range group.CanSee {
		//fmt.Println(forums[fid])
		forum := forums[fid]
		if forum.Active && forum.Name != "" {
			if forum.LastTopicID != 0 {
				forum.LastTopicTime, err = relative_time(forum.LastTopicTime)
				if err != nil {
					InternalError(err,w,r)
				}
			} else {
				forum.LastTopic = "None"
				forum.LastTopicTime = ""
			}
			forumList = append(forumList, forum)
		}
	}
	
	pi := ForumsPage{"Forum List",user,noticeList,forumList,nil}
	if template_forums_handle != nil {
		template_forums_handle(pi,w)
	} else {
		err := templates.ExecuteTemplate(w,"forums.html",pi)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}
	
func route_topic_id(w http.ResponseWriter, r *http.Request){
	var err error
	var page, offset int
	var replyList []Reply
	
	page, _ = strconv.Atoi(r.FormValue("page"))
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/"):])
	if err != nil {
		PreError("The provided TopicID is not a valid number.",w,r)
		return
	}
	
	// Get the topic...
	topic, err := get_topicuser(tid)
	if err == sql.ErrNoRows {
		NotFound(w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	topic.Css = no_css_tmpl
	
	user, noticeList, ok := ForumSessionCheck(w,r,topic.ParentID)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic {
		//fmt.Printf("%+v\n", user.Perms)
		NoPermissions(w,r,user)
		return
	}
	
	topic.Content = parse_message(topic.Content)
	topic.ContentLines = strings.Count(topic.Content,"\n")
	
	// We don't want users posting in locked topics...
	if topic.Is_Closed && !user.Is_Mod {
		user.Perms.CreateReply = false
	}
	
	topic.Tag = groups[topic.Group].Tag
	if groups[topic.Group].Is_Mod || groups[topic.Group].Is_Admin {
		topic.Css = staff_css_tmpl
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
	
	// Calculate the offset
	last_page := int(topic.PostCount / items_per_page) + 1
	if page > 1 {
		offset = (items_per_page * page) - items_per_page
	} else if page == -1 {
		page = last_page
		offset = (items_per_page * page) - items_per_page
	} else {
		page = 1
	}
	
	// Get the replies..
	rows, err := get_topic_replies_offset_stmt.Query(topic.ID, offset)
	if err == sql.ErrNoRows {
		LocalError("Bad Page. Some of the posts may have been deleted or you got here by directly typing in the page number.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	replyItem := Reply{Css: no_css_tmpl}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &replyItem.Group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IpAddress, &replyItem.LikeCount, &replyItem.ActionType)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = parse_message(replyItem.Content)
		replyItem.ContentLines = strings.Count(replyItem.Content,"\n")
		
		if groups[replyItem.Group].Is_Mod || groups[replyItem.Group].Is_Admin {
			replyItem.Css = staff_css_tmpl
		} else {
			replyItem.Css = no_css_tmpl
		}
		
		if replyItem.Avatar != "" {
			if replyItem.Avatar[0] == '.' {
				replyItem.Avatar = "/uploads/avatar_" + strconv.Itoa(replyItem.CreatedBy) + replyItem.Avatar
			}
		} else {
			replyItem.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(replyItem.CreatedBy),1)
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
		
		// We really shouldn't have inline HTML, we should do something about this...
		if replyItem.ActionType != "" {
			switch(replyItem.ActionType) {
				case "lock":
					replyItem.ActionType = "This topic has been locked by <a href='" + build_profile_url(replyItem.CreatedBy) + "'>" + replyItem.CreatedByName + "</a>"
					replyItem.ActionIcon = "&#x1F512;&#xFE0E"
				case "unlock":
					replyItem.ActionType = "This topic has been reopened by <a href='" + build_profile_url(replyItem.CreatedBy) + "'>" + replyItem.CreatedByName + "</a>"
					replyItem.ActionIcon = "&#x1F513;&#xFE0E"
				case "stick":
					replyItem.ActionType = "This topic has been pinned by <a href='" + build_profile_url(replyItem.CreatedBy) + "'>" + replyItem.CreatedByName + "</a>"
					replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
				case "unstick":
					replyItem.ActionType = "This topic has been unpinned by <a href='" + build_profile_url(replyItem.CreatedBy) + "'>" + replyItem.CreatedByName + "</a>"
					replyItem.ActionIcon = "&#x1F4CC;&#xFE0E"
				default:
					replyItem.ActionType = replyItem.ActionType + " has happened"
					replyItem.ActionIcon = ""
			}
		}
		replyItem.Liked = false
		
		if hooks["rrow_assign"] != nil {
			replyItem = run_hook("rrow_assign", replyItem).(Reply)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	rows.Close()
	
	tpage := TopicPage{topic.Title,user,noticeList,replyList,topic,page,last_page,nil}
	if template_topic_handle != nil {
		template_topic_handle(tpage,w)
	} else {
		err = templates.ExecuteTemplate(w,"topic.html", tpage)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_profile(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var err error
	var replyContent, replyCreatedByName, replyCreatedAt, replyAvatar, replyTag string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyLines, replyGroup int
	var replyCss template.CSS
	var replyList []Reply
	
	pid, err := strconv.Atoi(r.URL.Path[len("/user/"):])
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
		if err == sql.ErrNoRows {
			NotFound(w,r)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
	}
	
	// Get the replies..
	rows, err := db.Query("select users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group from users_replies left join users ON users_replies.createdBy = users.uid where users_replies.uid = ?", puser.ID)
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
			replyCss = staff_css_tmpl
		} else {
			replyCss = no_css_tmpl
		}
		if replyAvatar != "" {
			if replyAvatar[0] == '.' {
				replyAvatar = "/uploads/avatar_" + strconv.Itoa(replyCreatedBy) + replyAvatar
			}
		} else {
			replyAvatar = strings.Replace(noavatar,"{id}",strconv.Itoa(replyCreatedBy),1)
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
		
		replyList = append(replyList, Reply{rid,puser.ID,replyContent,parse_message(replyContent),replyCreatedBy,replyCreatedByName,replyGroup,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyCss,replyLines,replyTag,"","","",0,"",replyLiked,replyLikeCount,"",""})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	ppage := ProfilePage{puser.Name + "'s Profile",user,noticeList,replyList,*puser,false}
	if template_profile_handle != nil {
		template_profile_handle(ppage,w)
	} else {
		err = templates.ExecuteTemplate(w,"profile.html",ppage)
		if err != nil {
			InternalError(err,w,r)
		}
	}
}

func route_topic_create(w http.ResponseWriter, r *http.Request, sfid string){
	var fid int
	var err error
	if sfid != "" {
		fid, err = strconv.Atoi(sfid)
		if err != nil {
			PreError("The provided ForumID is not a valid number.",w,r)
			return
		}
	}
	
	user, noticeList, ok := ForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w,r,user)
		return
	}
	
	var forumList []Forum
	group := groups[user.Group]
	for _, fid := range group.CanSee {
		if forums[fid].Active && forums[fid].Name != "" {
			forumList = append(forumList, forums[fid])
		}
	}
	
	ctpage := CreateTopicPage{"Create Topic",user,noticeList,forumList,fid,nil}
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
func route_create_topic(w http.ResponseWriter, r *http.Request) {
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
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
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
	res, err := create_topic_stmt.Exec(fid,topic_name,content,parse_message(content),ipaddress,wcount,user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = add_topics_to_forum_stmt.Exec(1,fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	forums[fid].TopicCount -= 1
	
	_, err = update_forum_cache_stmt.Exec(topic_name,lastId,user.Name,user.ID,fid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	forums[fid].LastTopic = topic_name
	forums[fid].LastTopicID = int(lastId)
	forums[fid].LastReplyer = user.Name
	forums[fid].LastReplyerID = user.ID
	forums[fid].LastTopicTime = ""
	
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
}

func route_create_reply(w http.ResponseWriter, r *http.Request) {
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
	
	var topic_name string
	var fid int
	var createdBy int
	err = db.QueryRow("select title, parentID, createdBy from topics where tid = ?",tid).Scan(&topic_name,&fid,&createdBy)
	if err == sql.ErrNoRows {
		PreError("Couldn't find the parent topic",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
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
	_, err = create_reply_stmt.Exec(tid,content,parse_message(content),ipaddress,wcount, user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = add_replies_to_topic_stmt.Exec(1, tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	_, err = update_forum_cache_stmt.Exec(topic_name, tid, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	res, err := add_activity_stmt.Exec(user.ID,createdBy,"reply","topic",tid)
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
	
	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err != sql.ErrNoRows {
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

func route_like_topic(w http.ResponseWriter, r *http.Request) {
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
	
	var words int
	var fid int
	var createdBy int
	err = db.QueryRow("select parentID, words, createdBy from topics where tid = ?", tid).Scan(&fid,&words,&createdBy)
	if err == sql.ErrNoRows {
		PreError("The requested topic doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w,r,user)
		return
	}
	
	err = db.QueryRow("select targetItem from likes where sentBy = ? and targetItem = ? and targetType = 'topics'", user.ID, tid).Scan(&tid)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	} else if err != sql.ErrNoRows {
		LocalError("You already liked this!",w,r,user)
		return
	}
	
	_, err = users.CascadeGet(createdBy)
	if err != nil && err == sql.ErrNoRows {
		LocalError("The target user doesn't exist",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	//score := words_to_score(words,true)
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
	
	res, err := add_activity_stmt.Exec(user.ID,createdBy,"like","topic",tid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	/*_, err = notify_watchers_stmt.Exec(lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}*/
	_, err = notify_one_stmt.Exec(createdBy,lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	// Reload the topic...
	err = topics.Load(tid)
	if err != nil && err != sql.ErrNoRows {
		LocalError("The liked topic no longer exists",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_reply_like_submit(w http.ResponseWriter, r *http.Request) {
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
	
	var tid int
	var words int
	var createdBy int
	err = db.QueryRow("select tid, words, createdBy from replies where rid = ?", rid).Scan(&tid, &words, &createdBy)
	if err == sql.ErrNoRows {
		PreError("You can't like something which doesn't exist!",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	var fid int
	err = db.QueryRow("select parentID from topics where tid = ?", tid).Scan(&fid)
	if err == sql.ErrNoRows {
		PreError("The parent topic doesn't exist.",w,r)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	user, ok := SimpleForumSessionCheck(w,r,fid)
	if !ok {
		return
	}
	if !user.Perms.ViewTopic || !user.Perms.LikeItem {
		NoPermissions(w,r,user)
		return
	}
	
	err = db.QueryRow("select targetItem from likes where sentBy = ? and targetItem = ? and targetType = 'replies'", user.ID, rid).Scan(&rid)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	} else if err != sql.ErrNoRows {
		LocalError("You already liked this!",w,r,user)
		return
	}
	
	_, err = users.CascadeGet(createdBy)
	if err != nil && err != sql.ErrNoRows {
		LocalError("The target user doesn't exist",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	//score := words_to_score(words,false)
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
	
	res, err := add_activity_stmt.Exec(user.ID,createdBy,"like","post",rid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = notify_one_stmt.Exec(createdBy,lastId)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	http.Redirect(w,r,"/topic/" + strconv.Itoa(tid),http.StatusSeeOther)
}

func route_profile_reply_create(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
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
	
	_, err = create_profile_reply_stmt.Exec(uid,html.EscapeString(preparse_message(r.PostFormValue("reply-content"))),parse_message(html.EscapeString(preparse_message(r.PostFormValue("reply-content")))),user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	var user_name string
	err = db.QueryRow("select name from users where uid = ?", uid).Scan(&user_name)
	if err == sql.ErrNoRows {
		LocalError("The profile you're trying to post on doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	http.Redirect(w, r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
}

func route_report_submit(w http.ResponseWriter, r *http.Request, sitem_id string) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
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
	var tid int
	var title, content, data string
	if item_type == "reply" {
		err = db.QueryRow("select tid, content from replies where rid = ?", item_id).Scan(&tid, &content)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the reported post",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		
		err = db.QueryRow("select title, data from topics where tid = ?",tid).Scan(&title,&data)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the topic which the reported post is supposed to be in",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		content = content + "\n\nOriginal Post: #rid-" + strconv.Itoa(item_id)
	} else if item_type == "user-reply" {
		err = db.QueryRow("select uid, content from users_replies where rid = ?", item_id).Scan(&tid, &content)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the reported post",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		
		err = db.QueryRow("select name from users where uid = ?", tid).Scan(&title)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the profile which the reported post is supposed to be on",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
		content = content + "\n\nOriginal Post: @" + strconv.Itoa(tid)
	} else if item_type == "topic" {
		err = db.QueryRow("select title, content from topics where tid = ?", item_id).Scan(&title,&content)
		if err == sql.ErrNoRows {
			NotFound(w,r)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
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
	rows, err := db.Query("select count(*) as count from topics where data = ? and data != '' and parentID = 1", item_type + "_" + strconv.Itoa(item_id))
	if err != nil && err != sql.ErrNoRows {
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
	
	title = "Report: " + title
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

func route_account_own_edit_critical(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	pi := Page{"Edit Password",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_critical_submit(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
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
	
	var real_password string
	var salt string
	current_password := r.PostFormValue("account-current-password")
	new_password := r.PostFormValue("account-new-password")
	confirm_password := r.PostFormValue("account-confirm-password")
	
	err = get_password_stmt.QueryRow(user.ID).Scan(&real_password, &salt)
	if err == sql.ErrNoRows {
		LocalError("Your account no longer exists.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	current_password = current_password + salt
	err = bcrypt.CompareHashAndPassword([]byte(real_password), []byte(current_password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
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
	_, err = logout_stmt.Exec(user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	noticeList = append(noticeList,"Your password was successfully updated")
	pi := Page{"Edit Password",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_avatar(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	pi := Page{"Edit Avatar",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html",pi)
}

func route_account_own_edit_avatar_submit(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > int64(max_request_size) {
		http.Error(w,"Request too large",http.StatusExpectationFailed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(max_request_size))
	
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	
	err := r.ParseMultipartForm(int64(max_request_size))
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
	noticeList = append(noticeList, "Your avatar was successfully updated")
	
	pi := Page{"Edit Avatar",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html", pi)
}

func route_account_own_edit_username(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	pi := Page{"Edit Username",user,noticeList,tList,user.Name}
	templates.ExecuteTemplate(w,"account-own-edit-username.html",pi)
}

func route_account_own_edit_username_submit(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
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
	
	noticeList = append(noticeList,"Your username was successfully updated")
	pi := Page{"Edit Username",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_account_own_edit_email(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You need to login to edit your account.",w,r,user)
		return
	}
	
	email := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := db.Query("select email, validated from emails where uid = ?", user.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&email.Email, &email.Validated)
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
	
	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user. This also applies when the admin switches enable_emails on after having it off for a while
	if len(emailList) == 0 {
		email.Email = user.Email
		email.Validated = false
		email.Primary = true
		emailList = append(emailList, email)
	}
	
	if !enable_emails {
		noticeList = append(noticeList, "The email system has been turned off. All features involving sending emails have been disabled.")
	}
	pi := Page{"Email Manager",user,noticeList,emailList,nil}
	templates.ExecuteTemplate(w,"account-own-edit-email.html", pi)
}

func route_account_own_edit_email_token_submit(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
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
	rows, err := db.Query("select email, validated, token from emails where uid = ?", user.ID)
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
	
	if !enable_emails {
		noticeList = append(noticeList,"The email system has been turned off. All features involving sending emails have been disabled.")
	}
	noticeList = append(noticeList,"Your email was successfully verified")
	pi := Page{"Email Manager",user,noticeList,emailList,nil}
	templates.ExecuteTemplate(w,"account-own-edit-email.html", pi)
}

func route_logout(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin {
		LocalError("You can't logout without logging in first.",w,r,user)
		return
	}
	
	_, err := logout_stmt.Exec(user.ID)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	err = users.Load(user.ID)
	if err != nil {
		LocalError("Your account doesn't exist!",w,r,user)
		return
	}
	http.Redirect(w,r, "/", http.StatusSeeOther)
}
	
func route_login(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	pi := Page{"Login",user,noticeList,tList,nil}
	templates.ExecuteTemplate(w,"login.html",pi)
}

func route_login_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form",w,r,user)
		return          
	}
	
	var uid int
	var real_password string
	var salt string
	var session string
	username := html.EscapeString(r.PostFormValue("username"))
	password := r.PostFormValue("password")
	
	err = login_stmt.QueryRow(username).Scan(&uid, &username, &real_password, &salt)
	if err == sql.ErrNoRows {
		LocalError("That username doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r)
		return
	}
	
	// Emergency password reset mechanism..
	if salt == "" {
		if password != real_password {
			LocalError("That's not the correct password.",w,r,user)
			return
		}
		
		// Re-encrypt the password
		SetPassword(uid, password)
	} else { // Normal login..
		password = password + salt
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		err := bcrypt.CompareHashAndPassword([]byte(real_password), []byte(password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			LocalError("That's not the correct password.",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r)
			return
		}
	}
	
	session, err = GenerateSafeString(sessionLength)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	_, err = update_session_stmt.Exec(session, uid)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	cookie := http.Cookie{Name: "uid",Value: strconv.Itoa(uid),Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name: "session",Value: session,Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	http.Redirect(w,r, "/", http.StatusSeeOther)
}

func route_register(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if user.Loggedin {
		LocalError("You're already logged in.",w,r,user)
		return
	}
	templates.ExecuteTemplate(w,"register.html",Page{"Registration",user,noticeList,tList,nil})
}

func route_register_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
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
	if password == "test" || password == "123456" || password == "123" || password == "password" {
		LocalError("Your password is too weak.",w,r,user)
		return  
	}
	
	confirm_password := r.PostFormValue("confirm_password")
	log.Print("Registration Attempt! Username: " + username)
	
	// Do the two inputted passwords match..?
	if password != confirm_password {
		LocalError("The two passwords don't match.",w,r,user)
		return
	}
	
	// Is this username already taken..?
	err = username_exists_stmt.QueryRow(username).Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r)
		return
	} else if err != sql.ErrNoRows {
		LocalError("This username isn't available. Try another.",w,r,user)
		return
	}
	
	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	session, err := GenerateSafeString(sessionLength)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	password = password + salt
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	var active int
	var group int
	switch settings["activation_type"] {
		case 1: // Activate All
			active = 1
			group = default_group
		default: // Anything else. E.g. Admin Activation or Email Activation.
			group = activation_group
	}
	
	res, err := register_stmt.Exec(username,email,string(hashed_password),salt,group,session,active)
	if err != nil {
		InternalError(err,w,r)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r)
		return
	}
	
	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if enable_emails {
		token, err := GenerateSafeString(80)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		_, err = add_email_stmt.Exec(email, lastId, 0, token)
		if err != nil {
			InternalError(err,w,r)
			return
		}
		
		if !SendValidationEmail(username, email, token) {
			LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.",w,r,user)
			return
		}
	}
	
	cookie := http.Cookie{Name: "uid",Value: strconv.FormatInt(lastId, 10),Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name: "session",Value: session,Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	http.Redirect(w,r, "/", http.StatusSeeOther)
}

var phrase_login_alerts []byte = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)
func route_api(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	format := r.FormValue("format")
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
	
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
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
			
			var msglist string
			var asid int
			var actor_id int
			var targetUser_id int
			var event string
			var elementType string
			var elementID int
			//---
			var targetUser *User
			
			rows, err := get_activity_feed_by_watcher_stmt.Query(user.ID)
			if err != nil {
				InternalErrorJS(err,w,r)
				return
			}
			
			for rows.Next() {
				err = rows.Scan(&asid,&actor_id,&targetUser_id,&event,&elementType,&elementID)
				if err != nil {
					InternalErrorJS(err,w,r)
					return
				}
				
				actor, err := users.CascadeGet(actor_id)
				if err != nil {
					LocalErrorJS("Unable to find the actor",w,r)
					return
				}
				
				/*if elementType != "forum" {
					targetUser, err = users.CascadeGet(targetUser_id)
					if err != nil {
						LocalErrorJS("Unable to find the target user",w,r)
						return
					}
				}*/
				
				if event == "friend_invite" {
					msglist += `{"msg":"You received a friend invite from {0}","sub":["` + actor.Name + `"],"path":"\/user\/`+strconv.Itoa(actor.ID)+`","avatar":"`+strings.Replace(actor.Avatar,"/","\\/",-1)+`"},`
					continue
				}
				
				/*
				"You received a friend invite from {user}"
				"{x}{mentioned you on}{user}{'s profile}"
				"{x}{mentioned you in}{topic}"
				"{x}{likes}{you}"
				"{x}{liked}{your topic}{topic}"
				"{x}{liked}{your post on}{user}{'s profile}" todo
				"{x}{liked}{your post in}{topic}"
				"{x}{replied to}{your post in}{topic}" todo
				"{x}{replied to}{topic}"
				"{x}{replied to}{your topic}{topic}"
				"{x}{created a new topic}{topic}"
				*/
				
				var act string
				var post_act string
				var url string
				var area string
				var start_frag string
				var end_frag string
				switch(elementType) {
					case "forum":
						if event == "reply" {
							act = "created a new topic"
							topic, err := topics.CascadeGet(elementID)
							if err != nil {
								LocalErrorJS("Unable to find the linked topic",w,r)
								return
							}
							url = build_topic_url(elementID)
							area = topic.Title
							// Store the forum ID in the targetUser column instead of making a new one? o.O
							// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now..
						} else {
							act = "did something in a forum"
						}
					case "topic":
						topic, err := topics.CascadeGet(elementID)
						if err != nil {
							LocalErrorJS("Unable to find the linked topic",w,r)
							return
						}
						url = build_topic_url(elementID)
						area = topic.Title
						
						if targetUser_id == user.ID {
							post_act = " your topic"
						}
					case "user":
						targetUser, err = users.CascadeGet(elementID)
						if err != nil {
							LocalErrorJS("Unable to find the target user",w,r)
							return
						}
						area = targetUser.Name
						end_frag = "'s profile"
						url = build_profile_url(elementID)
					case "post":
						topic, err := get_topic_by_reply(elementID)
						if err != nil {
							LocalErrorJS("Unable to find the target reply or parent topic",w,r)
							return
						}
						url = build_topic_url(topic.ID)
						area = topic.Title
						if targetUser_id == user.ID {
							post_act = " your post in"
						}
					default:
						LocalErrorJS("Invalid elementType",w,r)
				}
				
				switch(event) {
					case "like":
						if elementType == "user" {
							act = "likes"
							end_frag = ""
							if targetUser.ID == user.ID {
								area = "you"
							}
						} else {
							act = "liked"
						}
					case "mention":
						if elementType == "user" {
							act = "mentioned you on"
						} else {
							act = "mentioned you in"
							post_act = ""
						}
					case "reply": act = "replied to"
				}
				
				msglist += `{"msg":"{0} ` + start_frag + act + post_act + ` {1}` + end_frag + `","sub":["` + actor.Name + `","` + area + `"],"path":"` + url + `","avatar":"` + actor.Avatar + `"},`
			}
			
			err = rows.Err()
			if err != nil {
				InternalErrorJS(err,w,r)
				return
			}
			rows.Close()
			
			if len(msglist) != 0 {
				msglist = msglist[0:len(msglist)-1]
			}
			w.Write([]byte(`{"msgs":[`+msglist+`]}`))
			//fmt.Println(`{"msgs":[`+msglist+`]}`)
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
