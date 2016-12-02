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
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "golang.org/x/crypto/bcrypt"

// A blank list to fill out that parameter in Page for routes which don't use it
var tList map[int]interface{}

// GET functions
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
		http.Error(w,errpage,404)
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
	)
	topicList = make(map[int]interface{})
	currentID = 0
	
	rows, err := db.Query("select tid, title, content, createdBy, is_closed, sticky, createdAt, parentID from topics")
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&tid, &title, &content, &createdBy, &is_closed, &sticky, &createdAt, &parentID)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if is_closed {
			status = "closed"
		} else {
			status = "open"
		}
		topicList[currentID] = Topic{tid, title, content, createdBy, is_closed, sticky, createdAt,parentID, status}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	pi := Page{"Topic List","topics",user,topicList,0}
	templates.ExecuteTemplate(w,"topics.html", pi)
}
	
func route_topic_id(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
	var(
		tid int
		rid int
		parentID int
		title string
		content string
		createdBy int
		createdByName string
		createdAt string
		replyContent string
		replyCreatedBy int
		replyCreatedByName string
		replyCreatedAt string
		replyLastEdit int
		replyLastEditBy int
		replyAvatar string
		replyHasAvatar bool
		is_closed bool
		sticky bool
		
		currentID int
		replyList map[int]interface{}
	)
	replyList = make(map[int]interface{})
	currentID = 0
	
	tid, err := strconv.Atoi(r.URL.Path[len("/topic/"):])
	if err != nil {
		LocalError("The provided TopicID is not a valid number.",w,r,user)
		return
	}
	
	// Get the topic..
	//err = db.QueryRow("select title, content, createdBy, status, is_closed from topics where tid = ?", tid).Scan(&title, &content, &createdBy, &status, &is_closed)
	err = db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name from topics left join users ON topics.createdBy = users.uid where tid = ?", tid).Scan(&title, &content, &createdBy, &createdAt, &is_closed, &sticky, &parentID, &createdByName)
	if err == sql.ErrNoRows {
		errmsg := "The requested topic doesn't exist."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,404)
		return
	} else if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	var tdata map[string]string
	tdata = make(map[string]string)
	tdata["tid"] = strconv.Itoa(tid)
	tdata["title"] = title
	tdata["content"] = content
	tdata["createdBy"] = string(createdBy)
	tdata["createdAt"] = string(createdAt)
	tdata["parentID"] = string(parentID)
	if is_closed {
		tdata["status"] = "closed"
	} else {
		tdata["status"] = "open"
	}
	//tdata["sticky"] = sticky
	tdata["createdByName"] = createdByName
	
	// Get the replies..
	//rows, err := db.Query("select rid, content, createdBy, createdAt from replies where tid = ?", tid)
	rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name from replies left join users ON replies.createdBy = users.uid where tid = ?", tid)
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName)
		if err != nil {
			InternalError(err,w,r,user)
			return
		}
		
		if replyAvatar != "" {
			replyHasAvatar = true
			if replyAvatar[0] == '.' {
				replyAvatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + replyAvatar
			}
		} else {
			replyHasAvatar = false
		}
		
		replyList[currentID] = Reply{rid,tid,replyContent,replyCreatedBy,replyCreatedByName,replyCreatedAt,replyLastEdit,replyLastEditBy,replyAvatar,replyHasAvatar}
		currentID++
	}
	err = rows.Err()
	if err != nil {
		InternalError(err,w,r,user)
		return
	}
	
	pi := Page{title,"topic",user,replyList,tdata}
	templates.ExecuteTemplate(w,"topic.html", pi)
}
	
func route_topic_create(w http.ResponseWriter, r *http.Request){
	user := SessionCheck(w,r)
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
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	success := 1
	
	res, err := create_topic_stmt.Exec(html.EscapeString(r.PostFormValue("topic-name")),html.EscapeString(r.PostFormValue("topic-content")),int32(time.Now().Unix()),user.ID)
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	if success != 1 {
		errmsg := "Unable to create the topic"
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
	} else {
		http.Redirect(w, r, "/topic/" + strconv.FormatInt(lastId, 10), http.StatusSeeOther)
	}
}

func route_create_reply(w http.ResponseWriter, r *http.Request) {
	var tid int
	
	
	user := SessionCheck(w,r)
	if !user.Loggedin {
		LoginRequired(w,r,user)
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		LocalError("Bad Form", w, r, user)
		return          
	}
	
	success := 1
	tid, err = strconv.Atoi(r.PostFormValue("tid"))
	if err != nil {
		log.Print(err)
		success = 0
		
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
		return
	}
	//log.Println("A reply is being created")
	
	_, err = create_reply_stmt.Exec(tid,html.EscapeString(r.PostFormValue("reply-content")),int32(time.Now().Unix()),user.ID)
	if err != nil {
		log.Print(err)
		success = 0
	}
	
	if success != 1 {
		errmsg := "Unable to create the reply"
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
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
	
	var tid int
	tid, err = strconv.Atoi(r.URL.Path[len("/topic/edit/submit/"):])
	if err != nil {
		LocalErrorJSQ("The provided TopicID is not a valid number.",w,r,user,is_js)
		return
	}
	
	topic_name := r.PostFormValue("topic_name")
	topic_status := r.PostFormValue("topic_status")
	var is_closed bool
	if topic_status == "closed" {
		is_closed = true
	} else {
		is_closed = false
	}
	topic_content := html.EscapeString(r.PostFormValue("topic_content"))
	_, err = edit_topic_stmt.Exec(topic_name, topic_content, is_closed, tid)
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
	_, err = edit_reply_stmt.Exec(content, rid)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
	
	user.HasAvatar = true
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	
	pi := Page{"Edit Avatar","account-own-edit-avatar-success",user,tList,0}
	templates.ExecuteTemplate(w,"account-own-edit-avatar-success.html", pi)
}
func route_logout(w http.ResponseWriter, r *http.Request) {
	user := SessionCheck(w,r)
	if !user.Loggedin {
		errmsg := "You can't logout without logging in first."
		pi := Page{"Error","error",user,tList,errmsg}
		
		var b bytes.Buffer
		templates.ExecuteTemplate(&b,"error.html", pi)
		errpage := b.String()
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
			http.Error(w,errpage,500)
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
			http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
		http.Error(w,errpage,500)
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
