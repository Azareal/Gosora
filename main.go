/* Copyright Azareal 2016 - 2017 */
package main

import (
	"net/http"
	"log"
	"mime"
	"strings"
	"path/filepath"
	"io"
	"io/ioutil"
	"os"
	"html/template"
)

const hour int = 60 * 60
const day int = hour * 24
const month int = day * 30
const year int = day * 365
const kilobyte int = 1024
const megabyte int = 1024 * 1024
const saltLength int = 32
const sessionLength int = 80
var nogrouplog bool = false // This is mainly for benchmarks, as we don't want a lot of information getting in the way of the results

var templates = template.Must(template.ParseGlob("templates/*"))
var no_css_tmpl = template.CSS("")
var staff_css_tmpl = template.CSS(staff_css)
var settings map[string]interface{} = make(map[string]interface{})
var external_sites map[string]string = make(map[string]string)
var groups map[int]Group = make(map[int]Group)
var forums map[int]Forum = make(map[int]Forum)
var static_files map[string]SFile = make(map[string]SFile)
var ctemplates []string
var template_topic_handle func(TopicPage,io.Writer) = nil
var template_topics_handle func(Page,io.Writer) = nil
var template_forum_handle func(Page,io.Writer) = nil
var template_forums_handle func(Page,io.Writer) = nil
var template_profile_handle func(Page,io.Writer) = nil

func compile_templates() {
	var c CTemplateSet
	user := User{0,"","compiler@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","",""}
	var noticeList map[int]string = make(map[int]string)
	noticeList[0] = "test"
	
	log.Print("Compiling the templates")
	
	topic := TopicUser{0,"",template.HTML(""),0,false,false,"",0,"","","",no_css_tmpl,0,"","","",""}
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"",template.HTML(""),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	
	var varList map[string]VarItem = make(map[string]VarItem)
	tpage := TopicPage{"Title","name",user,noticeList,replyList,topic,false}
	topic_id_tmpl := c.compile_template("topic.html","templates/","TopicPage", tpage, varList)
	
	varList = make(map[string]VarItem)
	varList["extra_data"] = VarItem{"extra_data","tmpl_profile_vars.Something.(User)","User"}
	//pi := Page{"Title","name",user,noticeList,replyList,user}
	//profile_tmpl := c.compile_template("profile.html","templates/","Page", pi, varList)
	
	var forumList []interface{}
	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList, forum)
		}
	}
	varList = make(map[string]VarItem)
	pi := Page{"Forum List","forums",user,noticeList,forumList,0}
	forums_tmpl := c.compile_template("forums.html","templates/","Page", pi, varList)
	
	var topicList []interface{}
	topicList = append(topicList, TopicUser{1,"Topic Title","The topic content.",1,false,false,"",1,"open","Admin","","",0,"","","",""})
	pi = Page{"Topic List","topics",user,noticeList,topicList,""}
	topics_tmpl := c.compile_template("topics.html","templates/","Page", pi, varList)
	
	pi = Page{"General Forum","forum",user,noticeList,topicList,"There aren't any topics in this forum yet."}
	forum_tmpl := c.compile_template("forum.html","templates/","Page", pi, varList)
	
	log.Print("Writing the templates")
	write_template("topic", topic_id_tmpl)
	//write_template("profile", profile_tmpl)
	write_template("forums", forums_tmpl)
	write_template("topics", topics_tmpl)
	write_template("forum", forum_tmpl)
}

func write_template(name string, content string) {
	f, err := os.Create("./template_" + name + ".go")
	if err != nil {
		log.Fatal(err)
	}
	
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
	f.Close()
}

func main(){
	var err error
	init_database(err);
	compile_templates()
	
	log.Print("Loading the static files.")
	err = filepath.Walk("./public", func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		
		path = strings.Replace(path,"\\","/",-1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		
		path = strings.TrimPrefix(path,"public/")
		log.Print("Added the '" + path + "' static file.")
		static_files["/static/" + path] = SFile{data,0,int64(len(data)),mime.TypeByExtension(filepath.Ext("/public/" + path)),f,f.ModTime().UTC().Format(http.TimeFormat)}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	
	external_sites["YT"] = "https://www.youtube.com/"
	hooks["trow_assign"] = nil
	hooks["rrow_assign"] = nil
	templates.ParseGlob("pages/*")
	
	for name, body := range plugins {
		log.Print("Added plugin " + name)
		if body.Active {
			log.Print("Initialised plugin " + name)
			plugins[name].Init()
		}
	}
	
	// In a directory to stop it clashing with the other paths
	http.HandleFunc("/static/", route_static)
	//http.HandleFunc("/static/", route_fstatic)
	//fs_p := http.FileServer(http.Dir("./public"))
	//http.Handle("/static/", http.StripPrefix("/static/",fs_p))
	fs_u := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/",fs_u))
	
	http.HandleFunc("/overview/", route_overview)
	http.HandleFunc("/topics/create/", route_topic_create)
	http.HandleFunc("/topics/", route_topics)
	http.HandleFunc("/forums/", route_forums)
	http.HandleFunc("/forum/", route_forum)
	http.HandleFunc("/topic/create/submit/", route_create_topic)
	http.HandleFunc("/topic/", route_topic_id)
	http.HandleFunc("/reply/create/", route_create_reply)
	//http.HandleFunc("/reply/edit/", route_reply_edit)
	//http.HandleFunc("/reply/delete/", route_reply_delete)
	http.HandleFunc("/reply/edit/submit/", route_reply_edit_submit)
	http.HandleFunc("/reply/delete/submit/", route_reply_delete_submit)
	http.HandleFunc("/report/submit/", route_report_submit)
	http.HandleFunc("/topic/edit/submit/", route_edit_topic)
	http.HandleFunc("/topic/delete/submit/", route_delete_topic)
	http.HandleFunc("/topic/stick/submit/", route_stick_topic)
	http.HandleFunc("/topic/unstick/submit/", route_unstick_topic)
	
	// Custom Pages
	http.HandleFunc("/pages/", route_custom_page)
	
	// Accounts
	http.HandleFunc("/accounts/login/", route_login)
	http.HandleFunc("/accounts/create/", route_register)
	http.HandleFunc("/accounts/logout/", route_logout)
	http.HandleFunc("/accounts/login/submit/", route_login_submit)
	http.HandleFunc("/accounts/create/submit/", route_register_submit)
	
	//http.HandleFunc("/accounts/list/", route_login) // Redirect /accounts/ and /user/ to here..
	//http.HandleFunc("/accounts/create/full/", route_logout)
	//http.HandleFunc("/user/edit/", route_logout)
	http.HandleFunc("/user/edit/critical/", route_account_own_edit_critical) // Password & Email
	http.HandleFunc("/user/edit/critical/submit/", route_account_own_edit_critical_submit)
	http.HandleFunc("/user/edit/avatar/", route_account_own_edit_avatar)
	http.HandleFunc("/user/edit/avatar/submit/", route_account_own_edit_avatar_submit)
	http.HandleFunc("/user/edit/username/", route_account_own_edit_username)
	http.HandleFunc("/user/edit/username/submit/", route_account_own_edit_username_submit)
	http.HandleFunc("/user/", route_profile)
	http.HandleFunc("/profile/reply/create/", route_profile_reply_create)
	http.HandleFunc("/profile/reply/edit/submit/", route_profile_reply_edit_submit)
	http.HandleFunc("/profile/reply/delete/submit/", route_profile_reply_delete_submit)
	//http.HandleFunc("/user/edit/submit/", route_logout)
	http.HandleFunc("/users/ban/", route_ban)
	http.HandleFunc("/users/ban/submit/", route_ban_submit)
	http.HandleFunc("/users/unban/", route_unban)
	http.HandleFunc("/users/activate/", route_activate)
	
	// Admin
	http.HandleFunc("/panel/", route_panel)
	http.HandleFunc("/panel/forums/", route_panel_forums)
	http.HandleFunc("/panel/forums/create/", route_panel_forums_create_submit)
	http.HandleFunc("/panel/forums/delete/", route_panel_forums_delete)
	http.HandleFunc("/panel/forums/delete/submit/", route_panel_forums_delete_submit)
	http.HandleFunc("/panel/forums/edit/submit/", route_panel_forums_edit_submit)
	http.HandleFunc("/panel/settings/", route_panel_settings)
	http.HandleFunc("/panel/settings/edit/", route_panel_setting)
	http.HandleFunc("/panel/settings/edit/submit/", route_panel_setting_edit)
	http.HandleFunc("/panel/plugins/", route_panel_plugins)
	http.HandleFunc("/panel/plugins/activate/", route_panel_plugins_activate)
	http.HandleFunc("/panel/plugins/deactivate/", route_panel_plugins_deactivate)
	http.HandleFunc("/panel/users/", route_panel_users)
	http.HandleFunc("/panel/users/edit/", route_panel_users_edit)
	http.HandleFunc("/panel/users/edit/submit/", route_panel_users_edit_submit)
	
	http.HandleFunc("/", default_route)
	
	defer db.Close()
    http.ListenAndServe(":8080", nil)
}