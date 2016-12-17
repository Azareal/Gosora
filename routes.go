/* Copyright Azareal 2016 - 2017 */
package main

import "errors"
import "log"
import "fmt"
import "strconv"
import "bytes"
import "regexp"
import "strings"
import "time"
import "io"
import "os"
import "net/http"
import "html"
import "html/template"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "golang.org/x/crypto/bcrypt"

// A blank list to fill out that parameter in Page for routes which don't use it
var tList map[int]interface{}
var nList map[int]string

// GET functions
func route_static(w http.ResponseWriter, r *http.Request){
	//name := r.URL.Path[len("/static/"):]
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && static_files[r.URL.Path].Info.ModTime().Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	h := w.Header()
	h.Set("Last-Modified", static_files[r.URL.Path].FormattedModTime)
	h.Set("Content-Type", static_files[r.URL.Path].Mimetype)
	h.Set("Content-Length", strconv.FormatInt(static_files[r.URL.Path].Length, 10))
	//http.ServeContent(w,r,r.URL.Path,static_files[r.URL.Path].Info.ModTime(),static_files[r.URL.Path])
	//w.Write(static_files[r.URL.Path].Data)
	io.Copy(w, bytes.NewReader(static_files[r.URL.Path].Data))
	//io.CopyN(w, bytes.NewReader(static_files[r.URL.Path].Data), static_files[r.URL.Path].Length)
}

func route_fstatic(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, r.URL.Path)
}

func route_overview(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	pi := Page{"Overview","overview",user,noticeList,tList,0}
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
	}
	pi := Page{"Page","page",user,noticeList,tList,0}
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
	
	var(
		topicList map[int]interface{}
		currentID int
		
		tid int
		title string
		content string
		createdBy int
		is_closed bool
		sticky bool
		createdAt string
		parentID int
		status string
		name string
		avatar string
	)
	topicList = make(map[int]interface{})
	currentID = 0
	
	rows, err := db.Query("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&tid, &title, &content, &createdBy, &is_closed, &sticky, &createdAt, &parentID, &name, &avatar)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if is_closed {
			status = "closed"
		} else {
			status = "open"
		}
		if avatar != "" {
			if avatar[0] == '.' {
				avatar = "/uploads/avatar_" + strconv.Itoa(createdBy) + avatar
			}
		} else {
			avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(createdBy),1)
		}
		
		topicList[currentID] = TopicUser{tid,title,content,createdBy,is_closed,sticky, createdAt,parentID,status,name,avatar,"",0,"","","",""}
		
		if hooks["trow_assign"] != nil {
			topicList[currentID] = run_hook("trow_assign", topicList[currentID])
		}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	var msg string
	if len(topicList) == 0 {
		msg = "There aren't any topics yet."
	} else {
		msg = ""
	}
	
	pi := Page{"Topic List","topics",user,noticeList,topicList,msg}
	err = templates.ExecuteTemplate(w,"topics.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_forum(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var(
		topicList map[int]interface{}
		currentID int
		
		tid int
		title string
		content string
		createdBy int
		is_closed bool
		sticky bool
		createdAt string
		parentID int
		status string
		name string
		avatar string
	)
	topicList = make(map[int]interface{})
	currentID = 0
	
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
	
	rows, err := db.Query("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid WHERE topics.parentID = ? order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC", fid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&tid, &title, &content, &createdBy, &is_closed, &sticky, &createdAt, &parentID, &name, &avatar)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if is_closed {
			status = "closed"
		} else {
			status = "open"
		}
		if avatar != "" {
			if avatar[0] == '.' {
				avatar = "/uploads/avatar_" + strconv.Itoa(createdBy) + avatar
			}
		} else {
			avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(createdBy),1)
		}
		
		topicList[currentID] = TopicUser{tid,title,content,createdBy,is_closed,sticky,createdAt,parentID,status,name,avatar,"",0,"","","",""}
		
		if hooks["trow_assign"] != nil {
			topicList[currentID] = run_hook("trow_assign", topicList[currentID])
		}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	var msg string
	if len(topicList) == 0 {
		msg = "There aren't any topics in this forum yet."
	} else {
		msg = ""
	}
	
	pi := Page{forums[fid].Name,"forum",user,noticeList,topicList,msg}
	err = templates.ExecuteTemplate(w,"forum.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_forums(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var forumList map[int]interface{} = make(map[int]interface{})
	currentID := 0
	
	for _, forum := range forums {
		if forum.Active {
			forumList[currentID] = forum
			currentID++
		}
	}
	
	if len(forums) == 0 {
		InternalError(errors.New("No forums"),w,r,user)
		return
	}
	
	pi := Page{"Forum List","forums",user,noticeList,forumList,0}
	err := templates.ExecuteTemplate(w,"forums.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}
	
func route_topic_id(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	var(
		err error
		rid int
		content string
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
		replyURL string
		replyURLPrefix string
		replyURLName string
		is_super_admin bool
		group int
		
		currentID int
		replyList map[int]interface{}
	)
	replyList = make(map[int]interface{})
	currentID = 0
	topic := TopicUser{0,"","",0,false,false,"",0,"","","",no_css_tmpl,0,"","","",""}
	
	topic.ID, err = strconv.Atoi(r.URL.Path[len("/topic/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	// Get the topic..
	err = db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name, users.avatar, users.is_super_admin, users.group, users.url_prefix, users.url_name from topics left join users ON topics.createdBy = users.uid where tid = ?", topic.ID).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group, &topic.URLPrefix, &topic.URLName)
	if err == sql.ErrNoRows {
		NotFound(w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	topic.Content = template.HTML(parse_message(content))
	topic.ContentLines = strings.Count(content,"\n")
	
	if topic.Is_Closed {
		topic.Status = "closed"
	} else {
		topic.Status = "open"
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
	}
	if groups[group].Tag != "" {
			topic.Tag = groups[group].Tag
	} else {
		topic.Tag = ""
	}
	if settings["url_tags"] == false {
		topic.URLName = ""
	} else {
		topic.URL, ok = external_sites[topic.URLPrefix]
		if !ok {
			topic.URL = topic.URLName
		} else {
			topic.URL = replyURL + topic.URLName
		}
	}
	
	// Get the replies..
	rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name from replies left join users ON replies.createdBy = users.uid where tid = ?", topic.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &is_super_admin, &group, &replyURLPrefix, &replyURLName)
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
		} else {
			replyTag = ""
		}
		if settings["url_tags"] == false {
			replyURLName = ""
		} else {
			replyURL, ok = external_sites[replyURLPrefix]
			if !ok {
				replyURL = replyURLName
			} else {
				replyURL = replyURL + replyURLName
			}
		}
		
		replyList[currentID] = Reply{rid,topic.ID,replyContent,template.HTML(parse_message(replyContent)),replyCreatedBy,replyCreatedByName,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyCss,replyLines,replyTag,replyURL,replyURLPrefix,replyURLName}
		
		if hooks["rrow_assign"] != nil {
			replyList[currentID] = run_hook("rrow_assign", replyList[currentID])
		}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{topic.Title,"topic",user,noticeList,replyList,topic}
	if ctemplates["topic"] != nil {
		ctemplates["topic"](pi,w)
	} else {
		err = templates.ExecuteTemplate(w,"topic.html", pi)
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
		
		currentID int
		replyList map[int]interface{}
	)
	replyList = make(map[int]interface{})
	currentID = 0
	
	puser := User{0,"",0,false,false,false,false,false,false,"",false,"","","","",""}
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
		err = db.QueryRow("SELECT `name`, `group`, `is_super_admin`, `avatar`, `message`, `url_prefix`, `url_name` FROM `users` WHERE `uid` = ?", puser.ID).Scan(&puser.Name, &puser.Group, &puser.Is_Super_Admin, &puser.Avatar, &puser.Message, &puser.URLPrefix, &puser.URLName)
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
	
	if groups[puser.Group].Tag != "" {
			puser.Tag = groups[puser.Group].Tag
	} else {
		puser.Tag = ""
	}
	
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
		
		replyList[currentID] = Reply{rid,puser.ID,replyContent,template.HTML(parse_message(replyContent)),replyCreatedBy,replyCreatedByName,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyCss,replyLines,replyTag,"","",""}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{puser.Name + "'s Profile","profile",user,noticeList,replyList,puser}
	//w.Write([]byte(template_profile(pi)))
	err = templates.ExecuteTemplate(w,"profile.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_topic_create(w http.ResponseWriter, r *http.Request){
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if user.Is_Banned {
		Banned(w,r,user)
		return
	}
	
	pi := Page{"Create Topic","create-topic",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"create-topic.html", pi)
}
	
// POST functions. Authorised users only.
func route_create_topic(w http.ResponseWriter, r *http.Request) {
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
	success := 1
	topic_name := html.EscapeString(r.PostFormValue("topic-name"))
	
	res, err := create_topic_stmt.Exec(topic_name,html.EscapeString(preparse_message(r.PostFormValue("topic-content"))),parse_message(html.EscapeString(preparse_message(r.PostFormValue("topic-content")))),user.ID)
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	_, err = update_forum_cache_stmt.Exec(topic_name, lastId, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if success != 1 {
		errmsg := "Unable to create the topic"
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Redirect(w, r, "/topic/" + strconv.FormatInt(lastId, 10), http.StatusSeeOther)
	}
}

func route_create_reply(w http.ResponseWriter, r *http.Request) {
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
	
	success := 1
	tid, err := strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		log.Print(err)
		success = 0
		
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	content := preparse_message(html.EscapeString(r.PostFormValue("reply-content")))
	log.Print(content)
	_, err = create_reply_stmt.Exec(tid,content,parse_message(content),user.ID)
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	var topic_name string
	err = db.QueryRow("select title from topics where tid = ?", tid).Scan(&topic_name)
	if err == sql.ErrNoRows {
		log.Print(err)
		success = 0
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = update_forum_cache_stmt.Exec(topic_name, tid, user.Name, user.ID, 1)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if success != 1 {
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	}
}

func route_profile_reply_create(w http.ResponseWriter, r *http.Request) {
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
	
	success := 1
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		log.Print(err)
		success = 0
		
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	_, err = create_profile_reply_stmt.Exec(uid,html.EscapeString(preparse_message(r.PostFormValue("reply-content"))),parse_message(html.EscapeString(preparse_message(r.PostFormValue("reply-content")))),user.ID)
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	var user_name string
	err = db.QueryRow("select name from users where uid = ?", uid).Scan(&user_name)
	if err == sql.ErrNoRows {
		log.Print(err)
		success = 0
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if success != 1 {
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Redirect(w, r, "/user/" + strconv.Itoa(uid), http.StatusSeeOther)
	}
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
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
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
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Password","account-own-edit",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_critical_submit(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
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
		pi := Page{"Error","error",user,nList,tList,"Your account doesn't exist."}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	current_password = current_password + salt
	err = bcrypt.CompareHashAndPassword([]byte(real_password), []byte(current_password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		pi := Page{"Error","error",user,nList,tList,"That's not the correct password."}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	if new_password != confirm_password {
		pi := Page{"Error","error",user,nList,tList,"The two passwords don't match."}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	SetPassword(user.ID, new_password)
	
	// Log the user out as a safety precaution
	_, err = logout_stmt.Exec(user.ID)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{"Edit Password","account-own-edit-success",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-success.html", pi)
}

func route_account_own_edit_avatar(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Avatar","account-own-edit-avatar",user,noticeList,tList,0}
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
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	err := r.ParseMultipartForm(int64(max_request_size))
	if  err != nil {
		LocalError("Upload failed", w, r, user)
		return
	}
	
	var filename string = ""
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
	noticeList[len(noticeList)] = "Your avatar was successfully updated"
	
	pi := Page{"Edit Avatar","account-own-edit-avatar",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html", pi)
}

func route_account_own_edit_username(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Username","account-own-edit-username",user,noticeList,tList,user.Name}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_account_own_edit_username_submit(w http.ResponseWriter, r *http.Request) {
	user, noticeList, ok := SessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
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
	
	pi := Page{"Edit Username","account-own-edit-username",user,noticeList,tList,user.Name}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_logout(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	if !user.Loggedin {
		errmsg := "You can't logout without logging in first."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
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
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Login","login",user,noticeList,tList,0}
	templates.ExecuteTemplate(w,"login.html", pi)
}

func route_login_submit(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	
	if user.Loggedin {
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
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
		errmsg := "That username doesn't exist."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Emergency password reset mechanism..
	if salt == "" {
		if password != real_password {
			errmsg := "That's not the correct password."
			pi := Page{"Error","error",user,nList,tList,errmsg}
			
			var b bytes.Buffer
			templates.ExecuteTemplate(&b,"error.html", pi)
			errpage := b.String()
			w.WriteHeader(500)
			fmt.Fprintln(w,errpage)
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
			errmsg := "That's not the correct password."
			pi := Page{"Error","error",user,nList,tList,errmsg}
			
			var b bytes.Buffer
			templates.ExecuteTemplate(&b,"error.html", pi)
			errpage := b.String()
			w.WriteHeader(500)
			fmt.Fprintln(w,errpage)
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
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Registration","register",user,noticeList,tList,0}
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
		errmsg := "The two passwords don't match."
		pi := Page{"Password Mismatch","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	// Is this username already taken..?
	err = username_exists_stmt.QueryRow(username).Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		InternalError(err,w,r,user)
		return
	} else if err != sql.ErrNoRows {
		errmsg := "This username isn't available. Try another."
		pi := Page{"Username Taken","error",user,nList,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
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
	
	res, err := register_stmt.Exec(username,email,string(hashed_password),salt,session)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	cookie := http.Cookie{Name: "uid",Value: strconv.FormatInt(lastId, 10),Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name: "session",Value: session,Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	http.Redirect(w,r, "/", http.StatusSeeOther)
}
