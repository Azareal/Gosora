package main

import (
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"path/filepath"
	"html/template"
)

const hour int = 60 * 60
const day int = hour * 24
const month int = day * 30
const year int = day * 365
const saltLength int = 32
const sessionLength int = 80
var db *sql.DB
var get_session_stmt *sql.Stmt
var create_topic_stmt *sql.Stmt
var create_reply_stmt *sql.Stmt
var edit_topic_stmt *sql.Stmt
var edit_reply_stmt *sql.Stmt
var delete_reply_stmt *sql.Stmt
var login_stmt *sql.Stmt
var update_session_stmt *sql.Stmt
var logout_stmt *sql.Stmt
var set_password_stmt *sql.Stmt
var register_stmt *sql.Stmt
var username_exists_stmt *sql.Stmt
var custom_pages map[string]string = make(map[string]string)
var templates = template.Must(template.ParseGlob("templates/*"))

func init_database(err error) {
	user := "root"
	password := "password"
	if(password != ""){
		password = ":" + password
	}
	dbname := "grosolo"
	db, err = sql.Open("mysql",user + password + "@tcp(127.0.0.1:3306)/" + dbname)
	if err != nil {
		log.Fatal(err)
	}
	
	// Make sure that the connection is alive..
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_session statement.")
	get_session_stmt, err = db.Prepare("SELECT `uid`, `name`, `group`, `is_super_admin`, `session` FROM `users` WHERE `uid` = ? AND `session` = ? AND `session` <> ''")
	if err != nil {
		log.Fatal(err)
	}
	_ = get_session_stmt // Bug fix, compiler isn't recognising this despite it being used, probably because it's hidden behind if statements
	
	log.Print("Preparing create_topic statement.")
	create_topic_stmt, err = db.Prepare("INSERT INTO topics(title,createdAt,lastReplyAt,createdBy) VALUES(?,?,0,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_reply statement.")
	create_reply_stmt, err = db.Prepare("INSERT INTO replies(tid,content,createdAt,createdBy) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing edit_topic statement.")
	edit_topic_stmt, err = db.Prepare("UPDATE topics SET title = ? WHERE tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing edit_reply statement.")
	edit_reply_stmt, err = db.Prepare("UPDATE replies SET content = ? WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing delete_reply statement.")
	delete_reply_stmt, err = db.Prepare("DELETE FROM replies WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing login statement.")
	login_stmt, err = db.Prepare("SELECT `uid`, `name`, `password`, `salt` FROM `users` WHERE `name` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_session statement.")
	update_session_stmt, err = db.Prepare("UPDATE users SET session = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing logout statement.")
	logout_stmt, err = db.Prepare("UPDATE users SET session = '' WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing set_password statement.")
	set_password_stmt, err = db.Prepare("UPDATE users SET password = ?, salt = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	// Add an admin version of register_stmt with more flexibility
	// create_account_stmt, err = db.Prepare("INSERT INTO 
	
	log.Print("Preparing register statement.")
	register_stmt, err = db.Prepare("INSERT INTO users(`name`,`password`,`salt`,`group`,`is_super_admin`,`session`) VALUES(?,?,?,0,0,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_session statement.")
	username_exists_stmt, err = db.Prepare("SELECT `name` FROM `users` WHERE `name` = ?")
	if err != nil {
		log.Fatal(err)
	}
}

func main(){
	var err error
	init_database(err);
	
	log.Print("Loading the custom pages.")
	err = filepath.Walk("pages/", add_custom_page)
	if err != nil {
		log.Fatal(err)
	}
	
	// In a directory to stop it clashing with the other paths
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static/",fs))
	
	http.HandleFunc("/overview/", route_overview)
	http.HandleFunc("/topics/create/", route_topic_create)
	http.HandleFunc("/topics/", route_topics)
	http.HandleFunc("/topic/create/submit/", route_create_topic) //POST
	http.HandleFunc("/topic/", route_topic_id)
	http.HandleFunc("/reply/create/", route_create_reply) //POST
	//http.HandleFunc("/reply/edit/", route_reply_edit) //POST
	//http.HandleFunc("/reply/delete/", route_reply_delete) //POST
	http.HandleFunc("/reply/edit/submit/", route_reply_edit_submit) //POST
	http.HandleFunc("/reply/delete/submit/", route_reply_delete_submit) //POST
	http.HandleFunc("/topic/edit/submit/", route_edit_topic) //POST
	
	// Custom Pages
	http.HandleFunc("/pages/:name/", route_custom_page)
	
	// Accounts
	http.HandleFunc("/accounts/login/", route_login)
	http.HandleFunc("/accounts/create/", route_register)
	http.HandleFunc("/accounts/logout/", route_logout)
	http.HandleFunc("/accounts/login/submit/", route_login_submit) // POST
	http.HandleFunc("/accounts/create/submit/", route_register_submit) // POST
	
	//http.HandleFunc("/accounts/list/", route_login) // Redirect /accounts/ and /user/ to here..
	//http.HandleFunc("/accounts/create/full/", route_logout)
	//http.HandleFunc("/user/edit/", route_logout)
	http.HandleFunc("/user/edit/critical/", route_account_own_edit_critical) // Password & Email
	http.HandleFunc("/user/edit/critical/submit/", route_account_own_edit_critical_submit)
	//http.HandleFunc("/user/:id/edit/", route_logout)
	//http.HandleFunc("/user/:id/ban/", route_logout)
	http.HandleFunc("/", route_topics)
	
	defer db.Close()
    http.ListenAndServe(":8080", nil)
}