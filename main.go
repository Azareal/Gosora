/* Copyright Azareal 2016 - 2018 */
package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"io"
	"os"
	"net/http"
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

var router *GenRouter
var startTime time.Time
//var timeLocation *time.Location
var templates = template.New("")
//var no_css_tmpl template.CSS = template.CSS("")
var settings map[string]interface{} = make(map[string]interface{})
var external_sites map[string]string = map[string]string{
	"YT":"https://www.youtube.com/",
}
var groups []Group
var groupCapCount int
var static_files map[string]SFile = make(map[string]SFile)
var logWriter io.Writer = io.MultiWriter(os.Stderr)

func interpreted_topic_template(pi TopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["topic"]
	if !ok {
		mapping = "topic"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}

var template_topic_handle func(TopicPage,http.ResponseWriter) = interpreted_topic_template
var template_topic_alt_handle func(TopicPage,http.ResponseWriter) = interpreted_topic_template
var template_topics_handle func(TopicsPage,http.ResponseWriter) = func(pi TopicsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["topics"]
	if !ok {
		mapping = "topics"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}
var template_forum_handle func(ForumPage,http.ResponseWriter) = func(pi ForumPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["forum"]
	if !ok {
		mapping = "forum"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}
var template_forums_handle func(ForumsPage,http.ResponseWriter) = func(pi ForumsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["forums"]
	if !ok {
		mapping = "forums"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}
var template_profile_handle func(ProfilePage,http.ResponseWriter) = func(pi ProfilePage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["profile"]
	if !ok {
		mapping = "profile"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}
var template_create_topic_handle func(CreateTopicPage,http.ResponseWriter) = func(pi CreateTopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["create-topic"]
	if !ok {
		mapping = "create-topic"
	}
	err := templates.ExecuteTemplate(w,mapping + ".html", pi)
	if err != nil {
		InternalError(err,w)
	}
}

func compile_templates() error {
	var c CTemplateSet
	user := User{62,build_profile_url("fake-user",62),"Fake User","compiler@localhost",0,false,false,false,false,false,false,GuestPerms,make(map[string]bool),"",false,"","","","","",0,0,"0.0.0.0.0"}
	// TO-DO: Do a more accurate level calculation for this?
	user2 := User{1,build_profile_url("admin-alice",1),"Admin Alice","alice@localhost",1,true,true,true,true,false,false,AllPerms,make(map[string]bool),"",true,"","","","","",58,1000,"127.0.0.1"}
	user3 := User{2,build_profile_url("admin-fred",62),"Admin Fred","fred@localhost",1,true,true,true,true,false,false,AllPerms,make(map[string]bool),"",true,"","","","","",42,900,"::1"}
	headerVars := HeaderVars{
		Site:site,
		NoticeList:[]string{"test"},
		Stylesheets:[]string{"panel"},
		Scripts:[]string{"whatever"},
		Widgets:PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	log.Print("Compiling the templates")

	topic := TopicUser{1,"blah","Blah","Hey there!",0,false,false,"Date","Date",0,"","127.0.0.1",0,1,"classname","weird-data",build_profile_url("fake-user",62),"Fake User",config.DefaultGroup,"",0,"","","","",58,false}
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"Yo!","Yo!",0,"alice","Alice",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})

	var varList map[string]VarItem = make(map[string]VarItem)
	tpage := TopicPage{"Title",user,headerVars,replyList,topic,1,1,extData}
	topic_id_tmpl, err := c.compile_template("topic.html","templates/","TopicPage", tpage, varList)
	if err != nil {
		return err
	}
	topic_id_alt_tmpl, err := c.compile_template("topic_alt.html","templates/","TopicPage", tpage, varList)
	if err != nil {
		return err
	}

	varList = make(map[string]VarItem)
	ppage := ProfilePage{"User 526",user,headerVars,replyList,user,extData}
	profile_tmpl, err := c.compile_template("profile.html","templates/","ProfilePage", ppage, varList)
	if err != nil {
		return err
	}

	var forumList []Forum
	forums, err := fstore.GetAll()
	if err != nil {
		return err
	}

	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList,*forum)
		}
	}
	varList = make(map[string]VarItem)
	forums_page := ForumsPage{"Forum List",user,headerVars,forumList,extData}
	forums_tmpl, err := c.compile_template("forums.html","templates/","ForumsPage",forums_page,varList)
	if err != nil {
		return err
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList,&TopicsRow{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",user3.ID,1,"","127.0.0.1",0,1,"classname","",&user2,"",0,&user3,"General","/forum/general.2"})
	topics_page := TopicsPage{"Topic List",user,headerVars,topicsList,extData}
	topics_tmpl, err := c.compile_template("topics.html","templates/","TopicsPage",topics_page,varList)
	if err != nil {
		return err
	}

	//var topicList []TopicUser
	//topicList = append(topicList,TopicUser{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","","admin-fred","Admin Fred",config.DefaultGroup,"",0,"","","","",58,false})
	forum_item := Forum{1,"general","General Forum","Where the general stuff happens",true,"all",0,"",0,"","",0,"",0,""}
	forum_page := ForumPage{"General Forum",user,headerVars,topicsList,forum_item,1,1,extData}
	forum_tmpl, err := c.compile_template("forum.html","templates/","ForumPage",forum_page,varList)
	if err != nil {
		return err
	}

	log.Print("Writing the templates")
	go write_template("topic", topic_id_tmpl)
	go write_template("topic_alt", topic_id_alt_tmpl)
	go write_template("profile", profile_tmpl)
	go write_template("forums", forums_tmpl)
	go write_template("topics", topics_tmpl)
	go write_template("forum", forum_tmpl)
	go func() {
		err := write_file("./template_list.go","package main\n\n" + c.FragOut)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func write_template(name string, content string) {
	err := write_file("./template_" + name + ".go", content)
	if err != nil {
		log.Fatal(err)
	}
}

func init_templates() {
	if dev.DebugMode {
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
	if dev.DebugMode {
		log.Print("Loading the template files...")
	}
	templates.Funcs(fmap)
	template.Must(templates.ParseGlob("templates/*"))
	template.Must(templates.ParseGlob("pages/*"))
}

func process_config() {
	config.Noavatar = strings.Replace(config.Noavatar,"{site_url}",site.Url,-1)
	if site.Port != "80" && site.Port != "443" {
		site.Url = strings.TrimSuffix(site.Url,"/")
		site.Url = strings.TrimSuffix(site.Url,"\\")
		site.Url = strings.TrimSuffix(site.Url,":")
		site.Url = site.Url + ":" + site.Port
	}
}

func main(){
	// TO-DO: Have a file for each run with the time/date the server started as the file name?
	// TO-DO: Log panics with recover()
	f, err := os.OpenFile("./operations.log",os.O_WRONLY|os.O_APPEND|os.O_CREATE,0755)
	if err != nil {
		log.Fatal(err)
	}
	logWriter = io.MultiWriter(os.Stderr,f)
	log.SetOutput(logWriter)

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
	//timeLocation = startTime.Location()

	log.Print("Processing configuration data")
	process_config()

	init_themes()
	err = init_database()
	if err != nil {
		log.Fatal(err)
	}

	init_templates()
	err = init_errors()
	if err != nil {
		log.Fatal(err)
	}

	if config.CacheTopicUser == CACHE_STATIC {
		users = NewMemoryUserStore(config.UserCacheCapacity)
		topics = NewMemoryTopicStore(config.TopicCacheCapacity)
	} else {
		users = NewSqlUserStore()
		topics = NewSqlTopicStore()
	}

	init_static_files()

	log.Print("Initialising the widgets")
	err = init_widgets()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the authentication system")
	auth = NewDefaultAuth()

	log.Print("Initialising the router")
	router = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	///router.HandleFunc("/static/", route_static)
	///router.HandleFunc("/overview/", route_overview)
	///router.HandleFunc("/topics/create/", route_topic_create)
	///router.HandleFunc("/topics/", route_topics)
	///router.HandleFunc("/forums/", route_forums)
	///router.HandleFunc("/forum/", route_forum)
	router.HandleFunc("/topic/create/submit/", route_topic_create_submit)
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

	log.Print("Initialising the plugins")
	init_plugins()

	defer db.Close()

	//if profiling {
	//	pprof.StopCPUProfile()
	//}

	// TO-DO: Let users run *both* HTTP and HTTPS
	log.Print("Initialising the HTTP server")
	if !site.EnableSsl {
		if site.Port == "" {
			site.Port = "80"
		}
		err = http.ListenAndServe(":" + site.Port, router)
	} else {
		if site.Port == "" {
			site.Port = "443"
		}
		if site.Port == "80" || site.Port == "443" {
			// We should also run the server on port 80
			// TO-DO: Redirect to port 443
			go func() {
				err = http.ListenAndServe(":80", &HttpsRedirect{})
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
		err = http.ListenAndServeTLS(":" + site.Port, config.SslFullchain, config.SslPrivkey, router)
	}

	// Why did the server stop?
	if err != nil {
		log.Fatal(err)
	}
}
