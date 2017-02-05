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
	//"runtime/pprof"
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

var templates = template.New("")
var no_css_tmpl = template.CSS("")
var staff_css_tmpl = template.CSS(staff_css)
var settings map[string]interface{} = make(map[string]interface{})
var external_sites map[string]string = make(map[string]string)
var groups []Group
var forums []Forum // The IDs for a forum tend to be low and sequential for the most part, so we can get more performance out of using a slice instead of a map AND it has better concurrency
var forum_perms map[int]map[int]ForumPerms // [gid][fid]Perms
var groupCapCount int
var forumCapCount int
var static_files map[string]SFile = make(map[string]SFile)

var template_topic_handle func(TopicPage,io.Writer) = nil
var template_topic_alt_handle func(TopicPage,io.Writer) = nil
var template_topics_handle func(TopicsPage,io.Writer) = nil
var template_forum_handle func(ForumPage,io.Writer) = nil
var template_forums_handle func(ForumsPage,io.Writer) = nil
var template_profile_handle func(ProfilePage,io.Writer) = nil
var template_create_topic_handle func(CreateTopicPage,io.Writer) = nil

func compile_templates() {
	var c CTemplateSet
	user := User{62,"","compiler@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","","",0,0,"0.0.0.0.0"}
	noticeList := []string{"test"}
	
	log.Print("Compiling the templates")
	
	topic := TopicUser{1,"Blah",template.HTML("Hey there!"),0,false,false,"",0,"","127.0.0.1",0,"","",no_css_tmpl,0,"","","","",58}
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"",template.HTML("Yo!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	
	var varList map[string]VarItem = make(map[string]VarItem)
	tpage := TopicPage{"Title",user,noticeList,replyList,topic,1,1,false}
	topic_id_tmpl := c.compile_template("topic.html","templates/","TopicPage", tpage, varList)
	topic_id_alt_tmpl := c.compile_template("topic_alt.html","templates/","TopicPage", tpage, varList)
	
	varList = make(map[string]VarItem)
	ppage := ProfilePage{"User 526",user,noticeList,replyList,user,false}
	profile_tmpl := c.compile_template("profile.html","templates/","ProfilePage", ppage, varList)
	
	var forumList []Forum
	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList, forum)
		}
	}
	varList = make(map[string]VarItem)
	forums_page := ForumsPage{"Forum List",user,noticeList,forumList,0}
	forums_tmpl := c.compile_template("forums.html","templates/","ForumsPage", forums_page, varList)
	
	var topicList []TopicUser
	topicList = append(topicList, TopicUser{1,"Topic Title","The topic content.",1,false,false,"",1,"","127.0.0.1",0,"Admin","","",0,"","","","",58})
	topics_page := TopicsPage{"Topic List",user,noticeList,topicList,""}
	topics_tmpl := c.compile_template("topics.html","templates/","TopicsPage", topics_page, varList)
	
	forum_item := Forum{1,"General Forum",true,"all",0,"",0,"",0,""}
	forum_page := ForumPage{"General Forum",user,noticeList,topicList,forum_item,1,1,nil}
	forum_tmpl := c.compile_template("forum.html","templates/","ForumPage", forum_page, varList)
	
	log.Print("Writing the templates")
	go write_template("topic", topic_id_tmpl)
	go write_template("topic_alt", topic_id_alt_tmpl)
	go write_template("profile", profile_tmpl)
	go write_template("forums", forums_tmpl)
	go write_template("topics", topics_tmpl)
	go write_template("forum", forum_tmpl)
	go write_file("./template_list.go", "package main\n\n" + c.FragOut)
}

func write_template(name string, content string) {
	write_file("./template_" + name + ".go", content)
}

func init_templates() {
	compile_templates()
	
	// Filler functions for now...
	fmap := make(map[string]interface{})
	fmap["add"] = func(in interface{}, in2 interface{})interface{} {
		return 1
	}
	fmap["subtract"] = func(in interface{}, in2 interface{})interface{} {
		return 1
	}
	fmap["multiply"] = func(in interface{}, in2 interface{})interface{} {
		return 1
	}
	fmap["divide"] = func(in interface{}, in2 interface{})interface{} {
		return 1
	}
	
	// The interpreted templates...
	templates.Funcs(fmap)
	template.Must(templates.ParseGlob("templates/*"))
	template.Must(templates.ParseGlob("pages/*"))
}

func main(){
	//if profiling {
	//	f, err := os.Create("startup_cpu.prof")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	pprof.StartCPUProfile(f)
	//}
	
	init_themes()
	var err error
	init_database(err)
	init_templates()
	db.SetMaxOpenConns(64)
	
	err = init_errors()
	if err != nil {
		log.Fatal(err)
	}
	
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
	
	init_plugins()
	
	router := NewRouter()
	router.HandleFunc("/static/", route_static) // In a directory to stop it clashing with the other paths
	
	fs_u := http.FileServer(http.Dir("./uploads"))
	router.Handle("/uploads/", http.StripPrefix("/uploads/",fs_u))
	
	router.HandleFunc("/overview/", route_overview)
	router.HandleFunc("/topics/create/", route_topic_create)
	router.HandleFunc("/topics/", route_topics)
	router.HandleFunc("/forums/", route_forums)
	router.HandleFunc("/forum/", route_forum)
	router.HandleFunc("/topic/create/submit/", route_create_topic)
	router.HandleFunc("/topic/", route_topic_id)
	router.HandleFunc("/reply/create/", route_create_reply)
	//router.HandleFunc("/reply/edit/", route_reply_edit)
	//router.HandleFunc("/reply/delete/", route_reply_delete)
	router.HandleFunc("/reply/edit/submit/", route_reply_edit_submit)
	router.HandleFunc("/reply/delete/submit/", route_reply_delete_submit)
	router.HandleFunc("/report/submit/", route_report_submit)
	router.HandleFunc("/topic/edit/submit/", route_edit_topic)
	router.HandleFunc("/topic/delete/submit/", route_delete_topic)
	router.HandleFunc("/topic/stick/submit/", route_stick_topic)
	router.HandleFunc("/topic/unstick/submit/", route_unstick_topic)
	
	// Custom Pages
	router.HandleFunc("/pages/", route_custom_page)
	
	// Accounts
	router.HandleFunc("/accounts/login/", route_login)
	router.HandleFunc("/accounts/create/", route_register)
	router.HandleFunc("/accounts/logout/", route_logout)
	router.HandleFunc("/accounts/login/submit/", route_login_submit)
	router.HandleFunc("/accounts/create/submit/", route_register_submit)
	
	//router.HandleFunc("/accounts/list/", route_login) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	//router.HandleFunc("/accounts/create/full/", route_logout) // Advanced account creator for admins?
	//router.HandleFunc("/user/edit/", route_logout)
	router.HandleFunc("/user/edit/critical/", route_account_own_edit_critical) // Password & Email
	router.HandleFunc("/user/edit/critical/submit/", route_account_own_edit_critical_submit)
	router.HandleFunc("/user/edit/avatar/", route_account_own_edit_avatar)
	router.HandleFunc("/user/edit/avatar/submit/", route_account_own_edit_avatar_submit)
	router.HandleFunc("/user/edit/username/", route_account_own_edit_username)
	router.HandleFunc("/user/edit/username/submit/", route_account_own_edit_username_submit)
	router.HandleFunc("/user/edit/email/", route_account_own_edit_email)
	router.HandleFunc("/user/edit/email/token/", route_account_own_edit_email_token_submit)
	router.HandleFunc("/user/", route_profile)
	router.HandleFunc("/profile/reply/create/", route_profile_reply_create)
	router.HandleFunc("/profile/reply/edit/submit/", route_profile_reply_edit_submit)
	router.HandleFunc("/profile/reply/delete/submit/", route_profile_reply_delete_submit)
	//router.HandleFunc("/user/edit/submit/", route_logout)
	router.HandleFunc("/users/ban/", route_ban)
	router.HandleFunc("/users/ban/submit/", route_ban_submit)
	router.HandleFunc("/users/unban/", route_unban)
	router.HandleFunc("/users/activate/", route_activate)
	
	// Admin
	router.HandleFunc("/panel/", route_panel)
	router.HandleFunc("/panel/forums/", route_panel_forums)
	router.HandleFunc("/panel/forums/create/", route_panel_forums_create_submit)
	router.HandleFunc("/panel/forums/delete/", route_panel_forums_delete)
	router.HandleFunc("/panel/forums/delete/submit/", route_panel_forums_delete_submit)
	router.HandleFunc("/panel/forums/edit/", route_panel_forums_edit)
	router.HandleFunc("/panel/forums/edit/submit/", route_panel_forums_edit_submit)
	router.HandleFunc("/panel/settings/", route_panel_settings)
	router.HandleFunc("/panel/settings/edit/", route_panel_setting)
	router.HandleFunc("/panel/settings/edit/submit/", route_panel_setting_edit)
	router.HandleFunc("/panel/themes/", route_panel_themes)
	router.HandleFunc("/panel/themes/default/", route_panel_themes_default)
	router.HandleFunc("/panel/plugins/", route_panel_plugins)
	router.HandleFunc("/panel/plugins/activate/", route_panel_plugins_activate)
	router.HandleFunc("/panel/plugins/deactivate/", route_panel_plugins_deactivate)
	router.HandleFunc("/panel/users/", route_panel_users)
	router.HandleFunc("/panel/users/edit/", route_panel_users_edit)
	router.HandleFunc("/panel/users/edit/submit/", route_panel_users_edit_submit)
	router.HandleFunc("/panel/groups/", route_panel_groups)
	//router.HandleFunc("/exit/", route_exit)
	
	router.HandleFunc("/", default_route)
	defer db.Close()
	
	//if profiling {
	//	pprof.StopCPUProfile()
	//}
	
	if !enable_ssl {
		if server_port == "" {
			 server_port = "80"
		}
		http.ListenAndServe(":" + server_port, router)
	} else {
		if server_port == "" {
			 server_port = "443"
		}
		http.ListenAndServeTLS(":" + server_port, ssl_fullchain, ssl_privkey, router)
	}
}