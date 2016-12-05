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
import "net/http"
import "html"
import "html/template"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "golang.org/x/crypto/bcrypt"

// A blank list to fill out that parameter in Page for routes which don't use it
var tList map[int]interface{}

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
	user := SessionCheck(w,r)
	pi := Page{"Overview","overview",user,tList,0}
	err := templates.ExecuteTemplate(w,"overview.html", pi)
    if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_custom_page(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	name := r.URL.Path[len("/pages/"):]
	
	val, ok := custom_pages[name];
	if ok {
		pi := Page{"Page","page",user,tList,val}
		templates.ExecuteTemplate(w,"custom_page.html", pi)
	} else {
		errmsg := "The requested page doesn't exist."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(404)
		fmt.Fprintln(w,errpage)
	}
}
	
func route_topics(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
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
		if avatar != "" && avatar[0] == '.' {
			avatar = "/uploads/avatar_" + strconv.Itoa(createdBy) + avatar
		}
		
		topicList[currentID] = TopicUser{tid, title, content, createdBy, is_closed, sticky, createdAt,parentID, status, name, avatar, ""}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	pi := Page{"Topic List","topics",user,topicList,0}
	err = templates.ExecuteTemplate(w,"topics.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_forum(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	var(
		topicList map[int]interface{}
		currentID int
		fname string
		
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
	
	err = db.QueryRow("select name from forums where fid = ?", fid).Scan(&fname)
	if err == sql.ErrNoRows {
		pi := Page{"Error","error",user,tList,"The requested forum doesn't exist."}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(404)
		fmt.Fprintln(w,errpage)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
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
		if avatar != "" && avatar[0] == '.' {
			avatar = "/uploads/avatar_" + strconv.Itoa(createdBy) + avatar
		}
		
		topicList[currentID] = TopicUser{tid, title, content, createdBy, is_closed, sticky, createdAt,parentID, status, name, avatar, ""}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	pi := Page{fname,"forum",user,topicList,0}
	err = templates.ExecuteTemplate(w,"forum.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}

func route_forums(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	var forumList map[int]interface{}
	forumList = make(map[int]interface{})
	currentID := 0
	
	rows, err := db.Query("select fid, name, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime from forums")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		forum := Forum{0,"","",0,"",0,""}
		err := rows.Scan(&forum.ID, &forum.Name, &forum.LastTopic,&forum.LastTopicID,&forum.LastReplyer,&forum.LastReplyerID,&forum.LastTopicTime)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if forum.LastTopicID != 0 {
			forum.LastTopicTime, err = relative_time(forum.LastTopicTime)
			if err != nil {
				InternalError(err,w,r,user)
				return
			}
		} else {
			forum.LastTopic = "None"
			forum.LastTopicTime = ""
		}
		
		forumList[currentID] = forum
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	pi := Page{"Forum List","forums",user,forumList,0}
	err = templates.ExecuteTemplate(w,"forums.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}
	
func route_topic_id(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
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
		is_super_admin bool
		group int
		
		currentID int
		replyList map[int]interface{}
	)
	replyList = make(map[int]interface{})
	currentID = 0
	topic := TopicUser{0,"","",0,false,false,"",0,"","","",no_css_tmpl}
	
	topic.ID, err = strconv.Atoi(r.URL.Path[len("/topic/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	// Get the topic..
	//err = db.QueryRow("select title, content, createdBy, status, is_closed from topics where tid = ?", tid).Scan(&title, &content, &createdBy, &status, &is_closed)
	err = db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name, users.avatar, users.is_super_admin, users.group from topics left join users ON topics.createdBy = users.uid where tid = ?", topic.ID).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group)
	if err == sql.ErrNoRows {
		errmsg := "The requested topic doesn't exist."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(404)
		fmt.Fprintln(w,errpage)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	topic.Content = template.HTML(parse_message(content))
	if topic.Is_Closed {
		topic.Status = "closed"
	} else {
		topic.Status = "open"
	}
	if topic.Avatar != "" && topic.Avatar[0] == '.' {
		topic.Avatar = "/uploads/avatar_" + strconv.Itoa(topic.CreatedBy) + topic.Avatar
	}
	if is_super_admin || groups[group].Is_Admin {
		topic.Css = staff_css_tmpl
	}
	
	// Get the replies..
	//rows, err := db.Query("select rid, content, createdBy, createdAt from replies where tid = ?", tid)
	rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group from replies left join users ON replies.createdBy = users.uid where tid = ?", topic.ID)
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
		
		if is_super_admin || groups[group].Is_Admin {
			replyCss = staff_css_tmpl
		} else {
			replyCss = no_css_tmpl
		}
		if replyAvatar != "" && replyAvatar[0] == '.' {
			replyAvatar = "/uploads/avatar_" + strconv.Itoa(replyCreatedBy) + replyAvatar
		}
		
		replyList[currentID] = Reply{rid,topic.ID,replyContent,template.HTML(parse_message(replyContent)),replyCreatedBy,replyCreatedByName,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyCss}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{topic.Title,"topic",user,replyList,topic}
	err = templates.ExecuteTemplate(w,"topic.html", pi)
	if err != nil {
        InternalError(err, w, r, user)
    }
}
	
func route_topic_create(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if user.Is_Banned {
		Banned(w,r,user)
		return
	}
	
	pi := Page{"Create Topic","create-topic",user,tList,0}
	templates.ExecuteTemplate(w,"create-topic.html", pi)
}
	
// POST functions. Authorised users only.
func route_create_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
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
	
	res, err := create_topic_stmt.Exec(topic_name,html.EscapeString(r.PostFormValue("topic-content")),parse_message(html.EscapeString(r.PostFormValue("topic-content"))),user.ID)
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
		pi := Page{"Error","error",user,tList,errmsg}
		
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
	user := SessionCheck(w,r)
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
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	_, err = create_reply_stmt.Exec(tid,html.EscapeString(r.PostFormValue("reply-content")),parse_message(html.EscapeString(r.PostFormValue("reply-content"))),user.ID)
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
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
	} else {
		http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	}
}

func route_edit_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	if user.Is_Banned {
		BannedJSQ(w,r,user,is_js)
		return
	}
	
	var tid int
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided TopicID is not a valid number.",w,r,user,is_js)
		return
	}
	
	topic_name := r.PostFormValue("topic_name")
	topic_status := r.PostFormValue("topic_status")
	//log.Print(topic_name)
	//log.Print(topic_status)
	var is_closed bool
	if topic_status == "closed" {
		is_closed = true
	} else {
		is_closed = false
	}
	topic_content := html.EscapeString(r.PostFormValue("topic_content"))
	//log.Print(topic_content)
	_, err = edit_topic_stmt.Exec(topic_name, topic_content, parse_message(topic_content), is_closed, tid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_delete_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/delete/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	err = db.QueryRow("SELECT tid from topics where tid = ?", tid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalError("The topic you tried to delete doesn't exist.",w,r,user)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	_, err = delete_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	log.Print("The topic '" + strconv.Itoa(tid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func route_stick_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/stick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = stick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_unstick_topic(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/unstick/submit/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	_, err = unstick_topic_stmt.Exec(tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w, r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
}

func route_reply_edit_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/edit/submit/"):])
	if err != nil {
		LocalError("The provided Reply ID is not a valid number.",w,r,user)
		return
	}
	
	content := html.EscapeString(r.PostFormValue("edit_item"))
	_, err = edit_reply_stmt.Exec(content, parse_message(content), rid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	// Get the Reply ID..
	var tid int
	err = db.QueryRow("select tid from replies where rid = ?", rid).Scan(&tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	if is_js == "0" {
		http.Redirect(w,r, "/topic/" + strconv.Itoa(tid) + "#reply-" + strconv.Itoa(rid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_reply_delete_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	is_js := r.PostFormValue("is_js")
	if is_js == "" {
		is_js = "0"
	}
	
	if !user.Is_Admin {
		NoPermissionsJSQ(w,r,user,is_js)
		return
	}
	
	rid, err := strconv.Atoi(r.URL.Path[len("/reply/delete/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided Reply ID is not a valid number.",w,r,user,is_js)
		return
	}
	
	var tid int
	err = db.QueryRow("SELECT tid from replies where rid = ?", rid).Scan(&tid)
	if err == sql.ErrNoRows {
		LocalErrorJSQ("The reply you tried to delete doesn't exist.",w,r,user,is_js)
		return
	} else if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	
	_, err = delete_reply_stmt.Exec(rid)
	if err != nil {
		InternalErrorJSQ(err,w,r,user,is_js)
		return
	}
	log.Print("The reply '" + strconv.Itoa(rid) + "' was deleted by User ID #" + strconv.Itoa(user.ID) + ".")
	
	if is_js == "0" {
		//http.Redirect(w,r, "/topic/" + strconv.Itoa(tid), http.StatusSeeOther)
	} else {
		fmt.Fprintf(w,"{'success': '1'}")
	}
}

func route_account_own_edit_critical(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Password","account-own-edit",user,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit.html", pi)
}

func route_account_own_edit_critical_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
		pi := Page{"Error","error",user,tList,"Your account doesn't exist."}
		
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
		pi := Page{"Error","error",user,tList,"That's not the correct password."}
		
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
		pi := Page{"Error","error",user,tList,"The two passwords don't match."}
		
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
	
	pi := Page{"Edit Password","account-own-edit-success",user,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-success.html", pi)
}

func route_account_own_edit_avatar(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Avatar","account-own-edit-avatar",user,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-avatar.html", pi)
}

func route_account_own_edit_avatar_submit(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > int64(max_request_size) {
		http.Error(w, "request too large", http.StatusExpectationFailed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(max_request_size))
	
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
	
	pi := Page{"Edit Avatar","account-own-edit-avatar-success",user,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-avatar-success.html", pi)
}

func route_account_own_edit_username(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Edit Username","account-own-edit-username",user,tList,user.Name}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_account_own_edit_username_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You need to login to edit your own account."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
		InternalError(err,w,r,user)
		return
	}
	user.Name = new_username
	
	pi := Page{"Edit Username","account-own-edit-username",user,tList,user.Name}
	templates.ExecuteTemplate(w,"account-own-edit-username.html", pi)
}

func route_logout(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You can't logout without logging in first."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
	user := SessionCheck(w,r)
	if user.Loggedin {
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Login","login",user,tList,0}
	templates.ExecuteTemplate(w,"login.html", pi)
}

func route_login_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if user.Loggedin {
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
	log.Print("Username: " + username)
	password := r.PostFormValue("password")
	log.Print("Password: " + password)
	
	err = login_stmt.QueryRow(username).Scan(&uid, &username, &real_password, &salt)
	if err == sql.ErrNoRows {
		errmsg := "That username doesn't exist."
		pi := Page{"Error","error",user,tList,errmsg}
		
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
			pi := Page{"Error","error",user,tList,errmsg}
			
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
		//hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		//log.Print("Hashed: " + string(hashed_password))
		//log.Print("Real:   " + real_password)
		//if string(hashed_password) != real_password {
		err := bcrypt.CompareHashAndPassword([]byte(real_password), []byte(password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			errmsg := "That's not the correct password."
			pi := Page{"Error","error",user,tList,errmsg}
			
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
	
	log.Print("Successful Login")
	log.Print("Session: " + session)
	cookie := http.Cookie{Name: "uid",Value: strconv.Itoa(uid),Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name: "session",Value: session,Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	http.Redirect(w,r, "/", http.StatusSeeOther)
}

func route_register(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if user.Loggedin {
		errmsg := "You're already logged in."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		w.WriteHeader(500)
		fmt.Fprintln(w,errpage)
		return
	}
	
	pi := Page{"Registration","register",user,tList,0}
	templates.ExecuteTemplate(w,"register.html", pi)
}

func route_register_submit(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	username := html.EscapeString(r.PostFormValue("username"))
	password := r.PostFormValue("password")
	confirm_password := r.PostFormValue("confirm_password")
	log.Print("Registration Attempt! Username: " + username)
	
	// Do the two inputted passwords match..?
	if password != confirm_password {
		errmsg := "The two passwords don't match."
		pi := Page{"Password Mismatch","error",user,tList,errmsg}
		
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
		pi := Page{"Username Taken","error",user,tList,errmsg}
		
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
	
	res, err := register_stmt.Exec(username,string(hashed_password),salt,session)
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

func route_panel_forums(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	
	var forumList map[int]interface{} = make(map[int]interface{})
	currentID := 0
	
	rows, err := db.Query("select fid, name from forums")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		forum := ForumSimple{0,""}
		err := rows.Scan(&forum.ID, &forum.Name)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		forumList[currentID] = forum
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{"Forum Manager","panel-forums",user,forumList,0}
	templates.ExecuteTemplate(w,"panel-forums.html", pi)
}

func route_panel_forums_create_submit(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	if !user.Is_Admin {
		NoPermissions(w,r,user)
		return
	}
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	_, err = create_forum_stmt.Exec(r.PostFormValue("forum-name"))
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	http.Redirect(w,r, "/panel/forums/", http.StatusSeeOther)
}