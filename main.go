/* Copyright Azareal 2016 - 2018 */
package main

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"io"
	"html/template"
	//"runtime/pprof"
)

var version Version = Version{Major:0,Minor:1,Patch:0,Tag:"dev"}

const hour int = 60 * 60
const day int = hour * 24
const month int = day * 30
const year int = day * 365
const kilobyte int = 1024
const megabyte int = kilobyte * 1024
const gigabyte int = megabyte * 1024
const terabyte int = gigabyte * 1024
const saltLength int = 32
const sessionLength int = 80
var enable_websockets bool = false // Don't change this, the value is overwritten by an initialiser

var startTime time.Time
var timeLocation *time.Location
var templates = template.New("")
var no_css_tmpl template.CSS = template.CSS("")
var staff_css_tmpl template.CSS = template.CSS(staff_css)
var settings map[string]interface{} = make(map[string]interface{})
var external_sites map[string]string = make(map[string]string)
var groups []Group
var groupCapCount int
var static_files map[string]SFile = make(map[string]SFile)

var template_topic_handle func(TopicPage,io.Writer) = nil
var template_topic_alt_handle func(TopicPage,io.Writer) = nil
var template_topics_handle func(TopicsPage,io.Writer) = nil
var template_forum_handle func(ForumPage,io.Writer) = nil
var template_forums_handle func(ForumsPage,io.Writer) = nil
var template_profile_handle func(ProfilePage,io.Writer) = nil
var template_create_topic_handle func(CreateTopicPage,io.Writer) = nil

func compile_templates() error {
	var c CTemplateSet
	user := User{62,"fake-user","Fake User","compiler@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","","",0,0,"0.0.0.0.0"}
	headerVars := HeaderVars{
		NoticeList:[]string{"test"},
		Stylesheets:[]string{"panel"},
		Scripts:[]string{"whatever"},
		Widgets:PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	log.Print("Compiling the templates")

	topic := TopicUser{1,"blah","Blah","Hey there!",0,false,false,"Date","Date",0,"","127.0.0.1",0,1,"classname","weird-data","fake-user","Fake User",default_group,"",no_css_tmpl,0,"","","","",58,false}
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"Yo!","Yo!",0,"alice","Alice",default_group,"",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1",false,1,"",""})

	var varList map[string]VarItem = make(map[string]VarItem)
	tpage := TopicPage{"Title",user,headerVars,replyList,topic,1,1,extData}
	topic_id_tmpl := c.compile_template("topic.html","templates/","TopicPage", tpage, varList)
	topic_id_alt_tmpl := c.compile_template("topic_alt.html","templates/","TopicPage", tpage, varList)

	varList = make(map[string]VarItem)
	ppage := ProfilePage{"User 526",user,headerVars,replyList,user,extData}
	profile_tmpl := c.compile_template("profile.html","templates/","ProfilePage", ppage, varList)

	var forumList []Forum
	forums, err := fstore.GetAll()
	if err != nil {
		return err
	}

	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList,forum)
		}
	}
	varList = make(map[string]VarItem)
	forums_page := ForumsPage{"Forum List",user,headerVars,forumList,extData}
	forums_tmpl := c.compile_template("forums.html","templates/","ForumsPage",forums_page,varList)

	var topicsList []TopicsRow
	topicsList = append(topicsList,TopicsRow{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","admin-alice","Admin Alice","","",0,"","","","",58,"General"})
	topics_page := TopicsPage{"Topic List",user,headerVars,topicsList,extData}
	topics_tmpl := c.compile_template("topics.html","templates/","TopicsPage",topics_page,varList)

	var topicList []TopicUser
	topicList = append(topicList,TopicUser{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","","admin-fred","Admin Fred",default_group,"","",0,"","","","",58,false})
	forum_item := Forum{1,"general","General Forum","Where the general stuff happens",true,"all",0,"","",0,"",0,""}
	forum_page := ForumPage{"General Forum",user,headerVars,topicList,forum_item,1,1,extData}
	forum_tmpl := c.compile_template("forum.html","templates/","ForumPage",forum_page,varList)

	log.Print("Writing the templates")
	go write_template("topic", topic_id_tmpl)
	go write_template("topic_alt", topic_id_alt_tmpl)
	go write_template("profile", profile_tmpl)
	go write_template("forums", forums_tmpl)
	go write_template("topics", topics_tmpl)
	go write_template("forum", forum_tmpl)
	go write_file("./template_list.go","package main\n\n" + c.FragOut)

	return nil
}

func write_template(name string, content string) {
	err := write_file("./template_" + name + ".go", content)
	if err != nil {
		log.Fatal(err)
	}
}

func init_templates() {
	if debug {
		log.Print("Initialising the template system")
	}
	compile_templates()

	// Filler functions for now...
	filler_func := func(in interface{}, in2 interface{})interface{} {
		return 1
	}
	fmap := make(map[string]interface{})
	fmap["add"] = filler_func
	fmap["subtract"] = filler_func
	fmap["multiply"] = filler_func
	fmap["divide"] = filler_func

	// The interpreted templates...
	if debug {
		log.Print("Loading the template files...")
	}
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

	log.Print("Running Gosora v" + version.String())
	fmt.Println("")
	startTime = time.Now()
	timeLocation = startTime.Location()

	init_themes()
	err := init_database()
	if err != nil {
		log.Fatal(err)
	}

	init_templates()
	err = init_errors()
	if err != nil {
		log.Fatal(err)
	}

	if cache_topicuser == CACHE_STATIC {
		users = NewMemoryUserStore(user_cache_capacity)
		topics = NewMemoryTopicStore(topic_cache_capacity)
	} else {
		users = NewSqlUserStore()
		topics = NewSqlTopicStore()
	}

	init_static_files()
	external_sites["YT"] = "https://www.youtube.com/"
	hooks["trow_assign"] = nil
	hooks["rrow_assign"] = nil
	init_plugins()

	log.Print("Initialising the widgets")
	err = init_widgets()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the authentication system")
	auth = NewDefaultAuth()

	log.Print("Initialising the router")
	router := NewGenRouter(http.FileServer(http.Dir("./uploads")))
	///router.HandleFunc("/static/", route_static)
	///router.HandleFunc("/overview/", route_overview)
	///router.HandleFunc("/topics/create/", route_topic_create)
	///router.HandleFunc("/topics/", route_topics)
	///router.HandleFunc("/forums/", route_forums)
	///router.HandleFunc("/forum/", route_forum)
	router.HandleFunc("/topic/create/submit/", route_create_topic)
	router.HandleFunc("/topic/", route_topic_id)
	router.HandleFunc("/reply/create/", route_create_reply)
	//router.HandleFunc("/reply/edit/", route_reply_edit)
	//router.HandleFunc("/reply/delete/", route_reply_delete)
	router.HandleFunc("/reply/edit/submit/", route_reply_edit_submit)
	router.HandleFunc("/reply/delete/submit/", route_reply_delete_submit)
	router.HandleFunc("/reply/like/submit/", route_reply_like_submit)
	///router.HandleFunc("/report/submit/", route_report_submit)
	router.HandleFunc("/topic/edit/submit/", route_edit_topic)
	router.HandleFunc("/topic/delete/submit/", route_delete_topic)
	router.HandleFunc("/topic/stick/submit/", route_stick_topic)
	router.HandleFunc("/topic/unstick/submit/", route_unstick_topic)
	router.HandleFunc("/topic/like/submit/", route_like_topic)

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
	router.HandleFunc("/user/edit/token/", route_account_own_edit_email_token_submit)
	router.HandleFunc("/user/", route_profile)
	router.HandleFunc("/profile/reply/create/", route_profile_reply_create)
	router.HandleFunc("/profile/reply/edit/submit/", route_profile_reply_edit_submit)
	router.HandleFunc("/profile/reply/delete/submit/", route_profile_reply_delete_submit)
	//router.HandleFunc("/user/edit/submit/", route_logout)
	router.HandleFunc("/users/ban/", route_ban)
	router.HandleFunc("/users/ban/submit/", route_ban_submit)
	router.HandleFunc("/users/unban/", route_unban)
	router.HandleFunc("/users/activate/", route_activate)

	// The Control Panel
	///router.HandleFunc("/panel/", route_panel)
	///router.HandleFunc("/panel/forums/", route_panel_forums)
	///router.HandleFunc("/panel/forums/create/", route_panel_forums_create_submit)
	///router.HandleFunc("/panel/forums/delete/", route_panel_forums_delete)
	///router.HandleFunc("/panel/forums/delete/submit/", route_panel_forums_delete_submit)
	///router.HandleFunc("/panel/forums/edit/", route_panel_forums_edit)
	///router.HandleFunc("/panel/forums/edit/submit/", route_panel_forums_edit_submit)
	///router.HandleFunc("/panel/forums/edit/perms/submit/", route_panel_forums_edit_perms_submit)
	///router.HandleFunc("/panel/settings/", route_panel_settings)
	///router.HandleFunc("/panel/settings/edit/", route_panel_setting)
	///router.HandleFunc("/panel/settings/edit/submit/", route_panel_setting_edit)
	///router.HandleFunc("/panel/themes/", route_panel_themes)
	///router.HandleFunc("/panel/themes/default/", route_panel_themes_default)
	///router.HandleFunc("/panel/plugins/", route_panel_plugins)
	///router.HandleFunc("/panel/plugins/activate/", route_panel_plugins_activate)
	///router.HandleFunc("/panel/plugins/deactivate/", route_panel_plugins_deactivate)
	///router.HandleFunc("/panel/users/", route_panel_users)
	///router.HandleFunc("/panel/users/edit/", route_panel_users_edit)
	///router.HandleFunc("/panel/users/edit/submit/", route_panel_users_edit_submit)
	///router.HandleFunc("/panel/groups/", route_panel_groups)
	///router.HandleFunc("/panel/groups/edit/", route_panel_groups_edit)
	///router.HandleFunc("/panel/groups/edit/perms/", route_panel_groups_edit_perms)
	///router.HandleFunc("/panel/groups/edit/submit/", route_panel_groups_edit_submit)
	///router.HandleFunc("/panel/groups/edit/perms/submit/", route_panel_groups_edit_perms_submit)
	///router.HandleFunc("/panel/groups/create/", route_panel_groups_create_submit)
	///router.HandleFunc("/panel/logs/mod/", route_panel_logs_mod)

	///router.HandleFunc("/api/", route_api)
	//router.HandleFunc("/exit/", route_exit)
	///router.HandleFunc("/", default_route)
	router.HandleFunc("/ws/", route_websockets)
	defer db.Close()

	//if profiling {
	//	pprof.StopCPUProfile()
	//}

	log.Print("Initialising the HTTP server")
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
