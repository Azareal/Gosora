/* Copyright Azareal 2016 - 2017 */
package main

import "log"
import "fmt"
import "strconv"
import "bytes"
import "regexp"
import "strings"
import "time"
import "io"
import "os"
import "net"
import "net/http"
import "html"
import "html/template"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "golang.org/x/crypto/bcrypt"

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var nList []string

// GET functions
func route_static(w http.ResponseWriter, r *http.Request){
	//name := r.URL.Path[len("/static/"):]
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
	h.Set("Content-Length", strconv.FormatInt(file.Length, 10)) // Avoid doing a type conversion every time?
	//http.ServeContent(w,r,r.URL.Path,file.Info.ModTime(),file)
	//w.Write(file.Data)
	io.Copy(w, bytes.NewReader(file.Data)) // Use w.Write instead?
	//io.CopyN(w, bytes.NewReader(file.Data), static_files[r.URL.Path].Length)
}

/*func route_exit(w http.ResponseWriter, r *http.Request){
	db.Close()
	os.Exit(0)
}*/

func route_fstatic(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, r.URL.Path)
}

func route_overview(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	pi := Page{"Overview",user,noticeList,tList,nil}
	err := templates.ExecuteTemplate(w,"overview.html", pi)
    if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_custom_page(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	name := r.URL.Path[len("/pages/"):]
	if templates.Lookup("page_" + name) == nil {
		NotFound(w,r,user)
		return
	}
	pi := Page{"Page",user,noticeList,tList,0}
	err := templates.ExecuteTemplate(w,"page_" + name,pi)
	if err != nil {
		InternalError(err, w, r, user)
	}
}
	
func route_topics(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	// I'll have to find a solution which doesn't involve shutting down all of the routes for a user, if they don't have ANY permissions
	/*if !user.Perms.ViewTopic {
		NoPermissions(w,r,user)
		return
	}*/
	
	var topicList []TopicUser
	rows, err := get_topic_list_stmt.Query()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	topicItem := TopicUser{ID: 0,}
	for rows.Next() {
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.ParentID, &topicItem.CreatedByName, &topicItem.Avatar)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if topicItem.Avatar != "" {
			if topicItem.Avatar[0] == '.' {
				topicItem.Avatar = "/uploads/avatar_" + strconv.Itoa(topicItem.CreatedBy) + topicItem.Avatar
			}
		} else {
			topicItem.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(topicItem.CreatedBy),1)
		}
		
		if hooks["trow_assign"] != nil {
			topicItem = run_hook("trow_assign", topicItem).(TopicUser)
		}
		topicList = append(topicList, topicItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	rows.Close()
	
	pi := TopicsPage{"Topic List",user,noticeList,topicList,nil}
	if template_topics_handle != nil {
		template_topics_handle(pi,w)
	} else {
		err = templates.ExecuteTemplate(w,"topics.html", pi)
		if err != nil {
			InternalError(err, w, r, user)
		}
	}
}

func route_forum(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var topicList []TopicUser
	fid, err := strconv.Atoi(r.URL.Path[len("/forum/"):])
	if err != nil {
		LocalError("The provided ForumID is not a valid number.",w,r,user)
		return
	}
	
	_, ok = forums[fid]
	if !ok {
		NotFound(w,r,user)
		return
	}
	if !user.Perms.ViewTopic {
		NoPermissions(w,r,user)
		return
	}
	
	rows, err := get_forum_topics_stmt.Query(fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	topicItem := TopicUser{ID: 0}
	for rows.Next() {
		err := rows.Scan(&topicItem.ID, &topicItem.Title, &topicItem.Content, &topicItem.CreatedBy, &topicItem.Is_Closed, &topicItem.Sticky, &topicItem.CreatedAt, &topicItem.ParentID, &topicItem.CreatedByName, &topicItem.Avatar)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if topicItem.Avatar != "" {
			if topicItem.Avatar[0] == '.' {
				topicItem.Avatar = "/uploads/avatar_" + strconv.Itoa(topicItem.CreatedBy) + topicItem.Avatar
			}
		} else {
			topicItem.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(topicItem.CreatedBy),1)
		}
		
		if hooks["trow_assign"] != nil {
			topicItem = run_hook("trow_assign", topicItem).(TopicUser)
		}
		topicList = append(topicList, topicItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	rows.Close()
	
	pi := ForumPage{forums[fid].Name,user,noticeList,topicList,nil}
	if template_forum_handle != nil {
		template_forum_handle(pi,w)
	} else {
		err = templates.ExecuteTemplate(w,"forum.html", pi)
		if err != nil {
			InternalError(err, w, r, user)
		}
	}
}

func route_forums(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var forumList []Forum
	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList, forum)
		}
	}
	
	pi := ForumsPage{"Forum List",user,noticeList,forumList,nil}
	if template_forums_handle != nil {
		template_forums_handle(pi,w)
	} else {
		err := templates.ExecuteTemplate(w,"forums.html",pi)
		if err != nil {
			InternalError(err,w,r,user)
		}
	}
}
	
func route_topic_id(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	var(
		err error
		content string
		is_super_admin bool
		group int
		replyList []Reply
	)
	
	topic := TopicUser{Css: no_css_tmpl}
	topic.ID, err = strconv.Atoi(r.URL.Path[len("/topic/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	if !user.Perms.ViewTopic {
		//fmt.Printf("%+v\n", user)
		//fmt.Printf("%+v\n", user.Perms)
		NoPermissions(w,r,user)
		return
	}
	
	// Get the topic..
	err = get_topic_user_stmt.QueryRow(topic.ID).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group, &topic.URLPrefix, &topic.URLName, &topic.Level, &topic.IpAddress)
	if err == sql.ErrNoRows {
		NotFound(w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	topic.Content = template.HTML(parse_message(content))
	topic.ContentLines = strings.Count(content,"\n")
	
	// We don't want users posting in locked topics...
	if topic.Is_Closed && !user.Is_Mod {
		user.Perms.CreateReply = false
	}
	
	if topic.Avatar != "" {
		if topic.Avatar[0] == '.' {
			topic.Avatar = "/uploads/avatar_" + strconv.Itoa(topic.CreatedBy) + topic.Avatar
		}
	} else {
		topic.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(topic.CreatedBy),1)
	}
	if is_super_admin || groups[group].Is_Mod || groups[group].Is_Admin {
		topic.Css = staff_css_tmpl
		topic.Level = -1
	}
	
	topic.Tag = groups[group].Tag
	
	if settings["url_tags"] == false {
		topic.URLName = ""
	} else {
		topic.URL, ok = external_sites[topic.URLPrefix]
		if !ok {
			topic.URL = topic.URLName
		} else {
			topic.URL = topic.URL + topic.URLName
		}
	}
	
	// Get the replies..
	rows, err := get_topic_replies_stmt.Query(topic.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	replyItem := Reply{Css: no_css_tmpl}
	for rows.Next() {
		err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &is_super_admin, &group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IpAddress)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		replyItem.ParentID = topic.ID
		replyItem.ContentHtml = template.HTML(parse_message(replyItem.Content))
		replyItem.ContentLines = strings.Count(replyItem.Content,"\n")
		if is_super_admin || groups[group].Is_Mod || groups[group].Is_Admin {
			replyItem.Css = staff_css_tmpl
			replyItem.Level = -1
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
		
		replyItem.Tag = groups[group].Tag
		
		if settings["url_tags"] == false {
			replyItem.URLName = ""
		} else {
			replyItem.URL, ok = external_sites[replyItem.URLPrefix]
			if !ok {
				replyItem.URL = replyItem.URLName
			} else {
				replyItem.URL = replyItem.URL + replyItem.URLName
			}
		}
		
		if hooks["rrow_assign"] != nil {
			replyItem = run_hook("rrow_assign", replyItem).(Reply)
		}
		replyList = append(replyList, replyItem)
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	rows.Close()
	
	tpage := TopicPage{topic.Title,user,noticeList,replyList,topic,nil}
	if template_topic_handle != nil {
		template_topic_handle(tpage,w)
	} else {
		err = templates.ExecuteTemplate(w,"topic.html", tpage)
		if err != nil {
			InternalError(err, w, r, user)
		}
	}
}

func route_profile(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var(
		err error
		rid int
		replyContent string
		replyCreatedBy int
		replyCreatedByName string
		replyCreatedAt string
		replyLastEdit int
		replyLastEditBy int
		replyAvatar string
		replyCss template.CSS
		replyLines int
		replyTag string
		is_super_admin bool
		group int
		
		replyList []Reply
	)
	
	puser := User{ID: 0,}
	puser.ID, err = strconv.Atoi(r.URL.Path[len("/user/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	if puser.ID == user.ID {
		user.Is_Mod = true
		puser = user
	} else {
		// Fetch the user data
		err = db.QueryRow("select `name`,`group`,`is_super_admin`,`avatar`,`message`,`url_prefix`,`url_name`,`level` from `users` where `uid` = ?", puser.ID).Scan(&puser.Name, &puser.Group, &puser.Is_Super_Admin, &puser.Avatar, &puser.Message, &puser.URLPrefix, &puser.URLName, &puser.Level)
		if err == sql.ErrNoRows {
			NotFound(w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		puser.Is_Admin = puser.Is_Super_Admin || groups[puser.Group].Is_Admin
		puser.Is_Super_Mod = puser.Is_Admin || groups[puser.Group].Is_Mod
		puser.Is_Mod = puser.Is_Super_Mod
		puser.Is_Banned = groups[puser.Group].Is_Banned
		if puser.Is_Banned && puser.Is_Super_Mod {
			puser.Is_Banned = false
		}
	}
	
	puser.Tag = groups[puser.Group].Tag
	
	if puser.Avatar != "" {
		if puser.Avatar[0] == '.' {
			puser.Avatar = "/uploads/avatar_" + strconv.Itoa(puser.ID) + puser.Avatar
		}
	} else {
		puser.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(puser.ID),1)
	}
	
	// Get the replies..
	rows, err := db.Query("select users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group from users_replies left join users ON users_replies.createdBy = users.uid where users_replies.uid = ?", puser.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &is_super_admin, &group)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		replyLines = strings.Count(replyContent,"\n")
		if is_super_admin || groups[group].Is_Mod || groups[group].Is_Admin {
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
		if groups[group].Tag != "" {
			replyTag = groups[group].Tag
		} else if puser.ID == replyCreatedBy {
			replyTag = "Profile Owner"
		} else {
			replyTag = ""
		}
		
		replyList = append(replyList, Reply{rid,puser.ID,replyContent,template.HTML(parse_message(replyContent)),replyCreatedBy,replyCreatedByName,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyCss,replyLines,replyTag,"","","",0,""})
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	ppage := ProfilePage{puser.Name + "'s Profile",user,noticeList,replyList,puser,false}
	if template_profile_handle != nil {
		template_profile_handle(ppage,w)
	} else {
		err = templates.ExecuteTemplate(w,"profile.html", ppage)
		if err != nil {
			InternalError(err, w, r, user)
		}
	}
}

func route_topic_create(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w,r,user)
		return
	}
	pi := Page{"Create Topic",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"create-topic.html", pi)
}
	
// POST functions. Authorised users only.
func route_create_topic(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	if !user.Loggedin || !user.Perms.CreateTopic {
		NoPermissions(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	topic_name := html.EscapeString(r.PostFormValue("topic-name"))
	content := html.EscapeString(preparse_message(r.PostFormValue("topic-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	
	res, err := create_topic_stmt.Exec(topic_name,content,parse_message(content),ipaddress,user.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	_, err = update_forum_cache_stmt.Exec(topic_name, lastId, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.FormatInt(lastId,10), http.StatusSeeOther)
	wcount := word_count(content)
	err = increase_post_user_stats(wcount, user.ID, true, user)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
}

func route_create_reply(w http.ResponseWriter, r *http.Request) {
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
		LocalError("Bad Form", w, r, user)
		return          
	}
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		LocalError("Failed to convert the TopicID", w, r, user)
		return
	}
	
	content := preparse_message(html.EscapeString(r.PostFormValue("reply-content")))
	ipaddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return
	}
	
	_, err = create_reply_stmt.Exec(tid,content,parse_message(content),ipaddress,user.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	var topic_name string
	err = db.QueryRow("select title from topics where tid = ?", tid).Scan(&topic_name)
	if err == sql.ErrNoRows {
		LocalError("Couldn't find the parent topic", w, r, user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = update_forum_cache_stmt.Exec(topic_name, tid, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	wcount := word_count(content)
	err = increase_post_user_stats(wcount, user.ID, false, user)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
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
		LocalError("Bad Form", w, r, user)
		return          
	}
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		LocalError("Invalid UID",w,r,user)
		return
	}
	
	_, err = create_profile_reply_stmt.Exec(uid,html.EscapeString(preparse_message(r.PostFormValue("reply-content"))),parse_message(html.EscapeString(preparse_message(r.PostFormValue("reply-content")))),user.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	var user_name string
	err = db.QueryRow("select name from users where uid = ?", uid).Scan(&user_name)
	if err == sql.ErrNoRows {
		LocalError("The profile you're trying to post on doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
}

func route_report_submit(w http.ResponseWriter, r *http.Request) {
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
		LocalError("Bad Form", w, r, user)
		return
	}
	if r.FormValue("session") != user.Session {
		SecurityError(w,r,user)
		return
	}
	
	item_id, err := strconv.Atoi(r.URL.Path[len("/report/submit/"):])
	if err != nil {
		LocalError("Bad ID", w, r, user)
		return
	}
	
	item_type := r.FormValue("type")
	success := 1
	
	var tid int
	var title string
	var content string
	var data string
	if item_type == "reply" {
		err = db.QueryRow("select tid, content from replies where rid = ?", item_id).Scan(&tid, &content)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		err = db.QueryRow("select title, data from topics where tid = ?", tid).Scan(&title,&data)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the topic which the reported post is supposed to be in", w, r, user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		content = content + "<br><br>Original Post: <a href='/topic/" + strconv.Itoa(tid) + "'>" + title + "</a>"
	} else if item_type == "user-reply" {
		err = db.QueryRow("select uid, content from users_replies where rid = ?", item_id).Scan(&tid, &content)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the reported post", w, r, user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		err = db.QueryRow("select name from users where uid = ?", tid).Scan(&title)
		if err == sql.ErrNoRows {
			LocalError("We were unable to find the profile which the reported post is supposed to be on", w, r, user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		content = content + "<br><br>Original Post: <a href='/user/" + strconv.Itoa(tid) + "'>" + title + "</a>"
	} else if item_type == "topic" {
		err = db.QueryRow("select title, content from topics where tid = ?", item_id).Scan(&title,&content)
		if err == sql.ErrNoRows {
			NotFound(w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
		content = content + "<br><br>Original Post: <a href='/topic/" + strconv.Itoa(item_id) + "'>" + title + "</a>"
	} else {
		if vhooks["report_preassign"] != nil {
			run_vhook_noreturn("report_preassign", &item_id, &item_type)
			return
		}
		
		// Don't try to guess the type
		LocalError("Unknown type", w, r, user)
		return  
	}
	
	var count int
	rows, err := db.Query("select count(*) as count from topics where data = ? and data != '' and parentID = -1", item_type + "_" + strconv.Itoa(item_id))
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r,user)
		return
	}
	
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
	}
	if count != 0 {
		LocalError("Someone has already reported this!", w, r, user)
		return
	}
	
	title = "Report: " + title
	res, err := create_report_stmt.Exec(title,content,content,user.ID,item_type + "_" + strconv.Itoa(item_id))
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	_, err = update_forum_cache_stmt.Exec(title, lastId, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if success != 1 {
		errmsg := "Unable to create the report"
		pi := Page{"Error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Redirect(w, r, "/topic/" + strconv.FormatInt(lastId, 10), http.StatusSeeOther)
	}
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
	pi := Page{"Edit Password",user,noticeList,tList,0}
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
		LocalError("Bad Form", w, r, user)
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
		InternalError(err,w,r,user)
		return
	}
	
	current_password = current_password + salt
	err = bcrypt.CompareHashAndPassword([]byte(real_password), []byte(current_password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		LocalError("That's not the correct password.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
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
		InternalError(err,w,r,user)
		return
	}
	
	noticeList[len(noticeList)] = "Your password was successfully updated"
	pi := Page{"Edit Password",user,noticeList,tList,0}
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
	pi := Page{"Edit Avatar",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html", pi)
}

func route_account_own_edit_avatar_submit(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > int64(max_request_size) {
		http.Error(w, "request too large", http.StatusExpectationFailed)
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
		LocalError("Upload failed", w, r, user)
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
				LocalError("Upload failed [File Creation Failed]", w, r, user)
				return
			}
			defer outfile.Close()
			
			_, err = io.Copy(outfile, infile);
			if  err != nil {
				LocalError("Upload failed [Copy Failed]", w, r, user)
				return
			}
		}
	}
	
	_, err = set_avatar_stmt.Exec("." + ext, strconv.Itoa(user.ID))
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	noticeList = append(noticeList, "Your avatar was successfully updated")
	
	pi := Page{"Edit Avatar",user,noticeList,tList,0}
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
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
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
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	new_username := html.EscapeString(r.PostFormValue("account-new-username"))
	_, err = set_username_stmt.Exec(new_username, strconv.Itoa(user.ID))
	if err != nil {
		LocalError("Unable to change the username. Does someone else already have this name?",w,r,user)
		return
	}
	user.Name = new_username
	
	noticeList = append(noticeList,"Your username was successfully updated")
	pi := Page{"Edit Username",user,noticeList,tList,0}
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
	pi := Page{"Email Manager",user,noticeList,emailList,0}
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
	token := r.URL.Path[len("/user/edit/email/token/"):]
	
	email := Email{UserID: user.ID}
	targetEmail := Email{UserID: user.ID}
	var emailList []interface{}
	rows, err := db.Query("select email, validated, token from emails where uid = ?", user.ID)
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
		if email.Token == token {
			targetEmail = email
		}
		emailList = append(emailList, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
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
		InternalError(err,w,r,user)
		return
	}
	
	// If Email Activation is on, then activate the account while we're here
	if settings["activation_type"] == 2 {
		_, err = activate_user_stmt.Exec(user.ID)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
	}
	
	if !enable_emails {
		noticeList = append(noticeList,"The email system has been turned off. All features involving sending emails have been disabled.")
	}
	noticeList = append(noticeList,"Your email was successfully verified")
	pi := Page{"Email Manager",user,noticeList,emailList,0}
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
		InternalError(err,w,r,user)
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
	pi := Page{"Login",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"login.html", pi)
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
		LocalError("Bad Form", w, r, user)
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
		InternalError(err,w,r,user)
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
			InternalError(err,w,r,user)
			return
		}
		
		err := bcrypt.CompareHashAndPassword([]byte(real_password), []byte(password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			LocalError("That's not the correct password.",w,r,user)
			return
		} else if err != nil {
			InternalError(err,w,r,user)
			return
		}
	}
	
	session, err = GenerateSafeString(sessionLength)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = update_session_stmt.Exec(session, uid)
	if err != nil {
		InternalError(err,w,r,user)
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
	pi := Page{"Registration",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"register.html", pi)
}

func route_register_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
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
	if password == "test" || password == "123456" || password == "123" || password == "password" {
		LocalError("Your password is too weak.", w, r, user)
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
		InternalError(err,w,r,user)
		return
	} else if err != sql.ErrNoRows {
		LocalError("This username isn't available. Try another.",w,r,user)
		return
	}
	
	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	session, err := GenerateSafeString(sessionLength)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	password = password + salt
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		InternalError(err,w,r,user)
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
		InternalError(err,w,r,user)
		return
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if enable_emails {
		token, err := GenerateSafeString(80)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		_, err = add_email_stmt.Exec(email, lastId, 0, token)
		if err != nil {
			InternalError(err,w,r,user)
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
