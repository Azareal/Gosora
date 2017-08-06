package main

import (
	//"os"
	"log"
	"bytes"
	"strings"
	"strconv"
	//"math/rand"
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
	"html/template"
	"io/ioutil"
	"database/sql"
	//"runtime/pprof"

	//_ "github.com/go-sql-driver/mysql"
	//"github.com/erikstmartin/go-testdb"
	//"github.com/husobee/vestigo"
)

var db_test *sql.DB
var db_prod *sql.DB
var gloinited bool

func gloinit() {
	dev.DebugMode = false
	//nogrouplog = true

	// init_database is a little noisy for a benchmark
	//discard := ioutil.Discard
	//log.SetOutput(discard)

	startTime = time.Now()
	//timeLocation = startTime.Location()
	process_config()

	init_themes()
	err := init_database()
	if err != nil {
		log.Fatal(err)
	}

	db_prod = db
	//db_test, err = sql.Open("testdb","")
	//if err != nil {
	//	log.Fatal(err)
	//}

	init_templates()
	db_prod.SetMaxOpenConns(64)
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
	//log.SetOutput(os.Stdout)
	auth = NewDefaultAuth()

	router = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	gloinited = true
}

func init() {
	gloinit()
}

func BenchmarkTopicTemplateSerial(b *testing.B) {
	b.ReportAllocs()

	user := User{0,"bob","Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,make(map[string]bool),"",false,"","","","","",0,0,"127.0.0.1"}
	admin := User{1,"admin-alice","Admin Alice","admin@localhost",0,true,true,true,true,true,false,AllPerms,make(map[string]bool),"",false,"","","","","",-1,58,"127.0.0.1"}

	topic := TopicUser{Title: "Lol",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",ClassName: "",Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"}

	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry","Jerry",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry2","Jerry2",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry3","Jerry3",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry4","Jerry4",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry5","Jerry5",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry6","Jerry6",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry7","Jerry7",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry8","Jerry8",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry9","Jerry9",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!","Hey everyone!",0,"jerry10","Jerry10",config.DefaultGroup,"",0,0,"","",0,"","","","",0,"127.0.0.1",false,1,"",""})

	headerVars := HeaderVars{
		NoticeList:[]string{"test"},
		Stylesheets:[]string{"panel"},
		Scripts:[]string{"whatever"},
		Widgets:PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
		Site:site,
	}

	tpage := TopicPage{"Topic Blah",user,headerVars,replyList,topic,1,1,extData}
	tpage2 := TopicPage{"Topic Blah",admin,headerVars,replyList,topic,1,1,extData}
	w := ioutil.Discard

	b.Run("compiled_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(tpage2,w)
		}
	})
	b.Run("interpreted_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", tpage2)
		}
	})
	b.Run("compiled_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(tpage,w)
		}
	})
	b.Run("interpreted_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", tpage)
		}
	})

	w2 := httptest.NewRecorder()
	b.Run("compiled_useradmin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w2.Body.Reset()
			template_topic(tpage2,w2)
		}
	})
	b.Run("interpreted_useradmin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w2.Body.Reset()
			templates.ExecuteTemplate(w2,"topic.html", tpage2)
		}
	})
	b.Run("compiled_userguest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w2.Body.Reset()
			template_topic(tpage,w2)
		}
	})
	b.Run("interpreted_userguest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w2.Body.Reset()
			templates.ExecuteTemplate(w2,"topic.html", tpage)
		}
	})

	/*f, err := os.Create("topic_bench.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()*/
}

func BenchmarkTopicsTemplateSerial(b *testing.B) {
	b.ReportAllocs()

	user := User{0,build_profile_url("bob",0),"Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,make(map[string]bool),"",false,"","","","","",0,0,"127.0.0.1"}
	admin := User{1,build_profile_url("admin-alice",1),"Admin Alice","admin@localhost",0,true,true,true,true,true,false,AllPerms,make(map[string]bool),"",false,"","","","","Admin",58,580,"127.0.0.1"}

	var topicList []*TopicsRow
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})
	topicList = append(topicList, &TopicsRow{Title: "Hey everyone!",Content: "Hey everyone!",CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,Creator:&admin,LastUser:&user,ClassName:"", IpAddress: "127.0.0.1"})

	headerVars := HeaderVars{
		NoticeList:[]string{"test"},
		Stylesheets:[]string{"panel"},
		Scripts:[]string{"whatever"},
		Widgets:PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
		Site:site,
	}

	w := ioutil.Discard
	tpage := TopicsPage{"Topic Blah",user,headerVars,topicList,extData}
	tpage2 := TopicsPage{"Topic Blah",admin,headerVars,topicList,extData}

	b.Run("compiled_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topics(tpage2,w)
		}
	})
	b.Run("interpreted_useradmin",func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topics.html", tpage2)
		}
	})
	b.Run("compiled_userguest",func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topics(tpage,w)
		}
	})
	b.Run("interpreted_userguest",func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topics.html", tpage)
		}
	})
}

func BenchmarkStaticRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	b.RunParallel(func(pb *testing.PB) {
		static_w := httptest.NewRecorder()
		static_req := httptest.NewRequest("get","/static/global.js",bytes.NewReader(nil))
		static_handler := http.HandlerFunc(route_static)
		for pb.Next() {
			//static_w.Code = 200
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
			//if static_w.Code != 200 {
			//	fmt.Println(static_w.Body)
			//	panic("HTTP Error!")
			//}
		}
	})
}

// TO-DO: Swap out LocalError for a panic for this?
func BenchmarkTopicAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.CascadeGet(1)
		if err != nil {
			b.Fatal(err)
		}
		if !admin.Is_Admin {
			b.Fatal("UID1 is not an admin")
		}
		admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

		topic_w := httptest.NewRecorder()
		topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
		topic_req_admin := topic_req
		topic_req_admin.AddCookie(&admin_uid_cookie)
		topic_req_admin.AddCookie(&admin_session_cookie)

		// Deal with the session stuff, etc.
		user, ok := PreRoute(topic_w,topic_req_admin)
		if !ok {
			b.Fatal("Mysterious error!")
		}

		for pb.Next() {
			topic_w.Body.Reset()
			route_topic_id(topic_w,topic_req_admin,user)
		}
	})
}

func BenchmarkTopicGuestRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		topic_w := httptest.NewRecorder()
		topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
		for pb.Next() {
			topic_w.Body.Reset()
			route_topic_id(topic_w,topic_req,guest_user)
		}
	})
}

// TO-DO: Make these routes compatible with the changes to the router
/*
func BenchmarkForumsAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.CascadeGet(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

		forums_w := httptest.NewRecorder()
		forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forums_req_admin := forums_req
		forums_req_admin.AddCookie(&admin_uid_cookie)
		forums_req_admin.AddCookie(&admin_session_cookie)
		forums_handler := http.HandlerFunc(route_forums)

		for pb.Next() {
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
}

func BenchmarkForumsAdminRouteParallelProf(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.CascadeGet(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path: "/",MaxAge: year}

		forums_w := httptest.NewRecorder()
		forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forums_req_admin := forums_req
		forums_req_admin.AddCookie(&admin_uid_cookie)
		forums_req_admin.AddCookie(&admin_session_cookie)
		forums_handler := http.HandlerFunc(route_forums)
		f, err := os.Create("cpu_forums_admin_parallel.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		for pb.Next() {
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
		pprof.StopCPUProfile()
	})
}

func BenchmarkForumsGuestRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		forums_w := httptest.NewRecorder()
		forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forums_handler := http.HandlerFunc(route_forums)
		for pb.Next() {
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
	})
}
*/

/*func BenchmarkRoutesSerial(b *testing.B) {
	b.ReportAllocs()
	admin, err := users.CascadeGet(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}

	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path: "/",MaxAge: year}

	if plugins_inited {
		b.Log("Plugins have already been initialised, they can't be deinitialised so these tests will run with plugins on")
	}
	static_w := httptest.NewRecorder()
	static_req := httptest.NewRequest("get","/static/global.js",bytes.NewReader(nil))
	static_handler := http.HandlerFunc(route_static)

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_req_admin := topic_req
	topic_req_admin.AddCookie(&admin_uid_cookie)
	topic_req_admin.AddCookie(&admin_session_cookie)
	topic_handler := http.HandlerFunc(route_topic_id)

	topics_w := httptest.NewRecorder()
	topics_req := httptest.NewRequest("get","/topics/",bytes.NewReader(nil))
	topics_req_admin := topics_req
	topics_req_admin.AddCookie(&admin_uid_cookie)
	topics_req_admin.AddCookie(&admin_session_cookie)
	topics_handler := http.HandlerFunc(route_topics)

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/1",bytes.NewReader(nil))
	forum_req_admin := forum_req
	forum_req_admin.AddCookie(&admin_uid_cookie)
	forum_req_admin.AddCookie(&admin_session_cookie)
	forum_handler := http.HandlerFunc(route_forum)

	forums_w := httptest.NewRecorder()
	forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forums_req_admin := forums_req
	forums_req_admin.AddCookie(&admin_uid_cookie)
	forums_req_admin.AddCookie(&admin_session_cookie)
	forums_handler := http.HandlerFunc(route_forums)

	if !gloinited {
		gloinit()
	}

	//f, err := os.Create("routes_bench_cpu.prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.StartCPUProfile(f)
	///defer pprof.StopCPUProfile()
	///pprof.StopCPUProfile()

	b.Run("static_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//static_w.Code = 200
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
			//if static_w.Code != 200 {
			//	fmt.Println(static_w.Body)
			//	panic("HTTP Error!")
			//}
		}
	})

	b.Run("topic_admin_recorder", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu.prof")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	fmt.Println(topic_w.Body)
			//	panic("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder", func(b *testing.B) {
		f, err := os.Create("routes_bench_topic_cpu_2.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
		pprof.StopCPUProfile()
	})
	b.Run("topics_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req_admin)
		}
	})
	b.Run("topics_guest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req)
		}
	})
	b.Run("forum_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req_admin)
		}
	})
	b.Run("forum_guest_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req)
		}
	})
	b.Run("forums_admin_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
	b.Run("forums_guest_recorder", func(b *testing.B) {
		//f, err := os.Create("routes_bench_forums_cpu_2.prof")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
		//pprof.StopCPUProfile()
	})

	if !plugins_inited {
		init_plugins()
	}

	b.Run("topic_admin_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu.prof")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	fmt.Println(topic_w.Body)
			//	panic("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu_2.prof")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topics_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req_admin)
		}
	})
	b.Run("topics_guest_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req)
		}
	})
	b.Run("forum_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req_admin)
		}
	})
	b.Run("forum_guest_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req)
		}
	})
	b.Run("forums_admin_recorder_with_plugins", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
	b.Run("forums_guest_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_forums_cpu_2.prof")
		//if err != nil {
		//	log.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
		//pprof.StopCPUProfile()
	})
}*/

func BenchmarkQueryTopicParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu TopicUser
		for pb.Next() {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.Is_Closed, &tu.Sticky, &tu.ParentID, &tu.IpAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
			if err == ErrNoRows {
				log.Fatal("No rows found!")
				return
			} else if err != nil {
				log.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkQueryPreparedTopicParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu TopicUser
		for pb.Next() {
			err := get_topic_user_stmt.QueryRow(1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.Is_Closed, &tu.Sticky, &tu.ParentID, &tu.IpAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
			if err == ErrNoRows {
				log.Fatal("No rows found!")
				return
			} else if err != nil {
				log.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkQueriesSerial(b *testing.B) {
	b.ReportAllocs()
	var tu TopicUser
	b.Run("topic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.Is_Closed, &tu.Sticky, &tu.ParentID, &tu.IpAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
			if err == ErrNoRows {
				log.Fatal("No rows found!")
				return
			} else if err != nil {
				log.Fatal(err)
				return
			}
		}
	})
	b.Run("topic_replies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				log.Fatal(err)
				return
			}
			for rows.Next() {}
			err = rows.Err()
			if err != nil {
				log.Fatal(err)
				return
			}
			defer rows.Close()
		}
	})

	var replyItem Reply
	var is_super_admin bool
	var group int
	b.Run("topic_replies_scan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				log.Fatal(err)
				return
			}
			for rows.Next() {
				err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &is_super_admin, &group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IpAddress)
				if err != nil {
					log.Fatal(err)
					return
				}
			}
			err = rows.Err()
			if err != nil {
				log.Fatal(err)
				return
			}
			defer rows.Close()
		}
	})
}

// Commented until I add logic for profiling the router generator, I'm not sure what the best way of doing that is
/*func addEmptyRoutesToMux(routes []string, serveMux *http.ServeMux) {
	for _, route := range routes {
		serveMux.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkDefaultGoRouterSerial(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("get","/topics/",bytes.NewReader(nil))
	routes := make([]string, 0)

	routes = append(routes,"/test/")
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/test/", func(_ http.ResponseWriter,_ *http.Request){})
	b.Run("one-route", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	routes = append(routes,"/topic/")
	routes = append(routes,"/forums/")
	routes = append(routes,"/forum/")
	routes = append(routes,"/panel/")
	serveMux = http.NewServeMux()
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("five-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/panel/plugins/")
	routes = append(routes,"/panel/groups/")
	routes = append(routes,"/panel/settings/")
	routes = append(routes,"/panel/users/")
	routes = append(routes,"/panel/forums/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("ten-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/panel/forums/create/submit/")
	routes = append(routes,"/panel/forums/delete/")
	routes = append(routes,"/users/ban/")
	routes = append(routes,"/panel/users/edit/")
	routes = append(routes,"/panel/forums/create/")
	routes = append(routes,"/users/unban/")
	routes = append(routes,"/pages/")
	routes = append(routes,"/users/activate/")
	routes = append(routes,"/panel/forums/edit/submit/")
	routes = append(routes,"/panel/plugins/activate/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("twenty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/panel/plugins/deactivate/")
	routes = append(routes,"/panel/plugins/install/")
	routes = append(routes,"/panel/plugins/uninstall/")
	routes = append(routes,"/panel/templates/")
	routes = append(routes,"/panel/templates/edit/")
	routes = append(routes,"/panel/templates/create/")
	routes = append(routes,"/panel/templates/delete/")
	routes = append(routes,"/panel/templates/edit/submit/")
	routes = append(routes,"/panel/themes/")
	routes = append(routes,"/panel/themes/edit/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("thirty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/panel/themes/create/")
	routes = append(routes,"/panel/themes/delete/")
	routes = append(routes,"/panel/themes/delete/submit/")
	routes = append(routes,"/panel/templates/create/submit/")
	routes = append(routes,"/panel/templates/delete/submit/")
	routes = append(routes,"/panel/widgets/")
	routes = append(routes,"/panel/widgets/edit/")
	routes = append(routes,"/panel/widgets/activate/")
	routes = append(routes,"/panel/widgets/deactivate/")
	routes = append(routes,"/panel/magical/wombat/path")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("forty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/report/")
	routes = append(routes,"/report/submit/")
	routes = append(routes,"/topic/create/submit/")
	routes = append(routes,"/topics/create/")
	routes = append(routes,"/overview/")
	routes = append(routes,"/uploads/")
	routes = append(routes,"/static/")
	routes = append(routes,"/reply/edit/submit/")
	routes = append(routes,"/reply/delete/submit/")
	routes = append(routes,"/topic/edit/submit/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("fifty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/topic/delete/submit/")
	routes = append(routes,"/topic/stick/submit/")
	routes = append(routes,"/topic/unstick/submit/")
	routes = append(routes,"/accounts/login/")
	routes = append(routes,"/accounts/create/")
	routes = append(routes,"/accounts/logout/")
	routes = append(routes,"/accounts/login/submit/")
	routes = append(routes,"/accounts/create/submit/")
	routes = append(routes,"/user/edit/critical/")
	routes = append(routes,"/user/edit/critical/submit/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("sixty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})

	serveMux = http.NewServeMux()
	routes = append(routes,"/user/edit/avatar/")
	routes = append(routes,"/user/edit/avatar/submit/")
	routes = append(routes,"/user/edit/username/")
	routes = append(routes,"/user/edit/username/submit/")
	routes = append(routes,"/profile/reply/create/")
	routes = append(routes,"/profile/reply/edit/submit/")
	routes = append(routes,"/profile/reply/delete/submit/")
	routes = append(routes,"/arcane/tower/")
	routes = append(routes,"/magical/kingdom/")
	routes = append(routes,"/insert/name/here/")
	addEmptyRoutesToMux(routes, serveMux)
	b.Run("seventy-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			serveMux.ServeHTTP(w,req)
		}
	})
}

func addEmptyRoutesToCustom(routes []string, router *Router) {
	for _, route := range routes {
		router.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkCustomRouterSerial(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("get","/topics/",bytes.NewReader(nil))
	routes := make([]string, 0)

	routes = append(routes,"/test/")
	router := NewRouter()
	router.HandleFunc("/test/", func(_ http.ResponseWriter,_ *http.Request){})
	b.Run("one-route", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	routes = append(routes,"/topic/")
	routes = append(routes,"/forums/")
	routes = append(routes,"/forum/")
	routes = append(routes,"/panel/")
	router = NewRouter()
	addEmptyRoutesToCustom(routes, router)
	b.Run("five-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/panel/plugins/")
	routes = append(routes,"/panel/groups/")
	routes = append(routes,"/panel/settings/")
	routes = append(routes,"/panel/users/")
	routes = append(routes,"/panel/forums/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("ten-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/panel/forums/create/submit/")
	routes = append(routes,"/panel/forums/delete/")
	routes = append(routes,"/users/ban/")
	routes = append(routes,"/panel/users/edit/")
	routes = append(routes,"/panel/forums/create/")
	routes = append(routes,"/users/unban/")
	routes = append(routes,"/pages/")
	routes = append(routes,"/users/activate/")
	routes = append(routes,"/panel/forums/edit/submit/")
	routes = append(routes,"/panel/plugins/activate/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("twenty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/panel/plugins/deactivate/")
	routes = append(routes,"/panel/plugins/install/")
	routes = append(routes,"/panel/plugins/uninstall/")
	routes = append(routes,"/panel/templates/")
	routes = append(routes,"/panel/templates/edit/")
	routes = append(routes,"/panel/templates/create/")
	routes = append(routes,"/panel/templates/delete/")
	routes = append(routes,"/panel/templates/edit/submit/")
	routes = append(routes,"/panel/themes/")
	routes = append(routes,"/panel/themes/edit/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("thirty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/panel/themes/create/")
	routes = append(routes,"/panel/themes/delete/")
	routes = append(routes,"/panel/themes/delete/submit/")
	routes = append(routes,"/panel/templates/create/submit/")
	routes = append(routes,"/panel/templates/delete/submit/")
	routes = append(routes,"/panel/widgets/")
	routes = append(routes,"/panel/widgets/edit/")
	routes = append(routes,"/panel/widgets/activate/")
	routes = append(routes,"/panel/widgets/deactivate/")
	routes = append(routes,"/panel/magical/wombat/path")
	addEmptyRoutesToCustom(routes, router)
	b.Run("forty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/report/")
	routes = append(routes,"/report/submit/")
	routes = append(routes,"/topic/create/submit/")
	routes = append(routes,"/topics/create/")
	routes = append(routes,"/overview/")
	routes = append(routes,"/uploads/")
	routes = append(routes,"/static/")
	routes = append(routes,"/reply/edit/submit/")
	routes = append(routes,"/reply/delete/submit/")
	routes = append(routes,"/topic/edit/submit/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("fifty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/topic/delete/submit/")
	routes = append(routes,"/topic/stick/submit/")
	routes = append(routes,"/topic/unstick/submit/")
	routes = append(routes,"/accounts/login/")
	routes = append(routes,"/accounts/create/")
	routes = append(routes,"/accounts/logout/")
	routes = append(routes,"/accounts/login/submit/")
	routes = append(routes,"/accounts/create/submit/")
	routes = append(routes,"/user/edit/critical/")
	routes = append(routes,"/user/edit/critical/submit/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("sixty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})

	router = NewRouter()
	routes = append(routes,"/user/edit/avatar/")
	routes = append(routes,"/user/edit/avatar/submit/")
	routes = append(routes,"/user/edit/username/")
	routes = append(routes,"/user/edit/username/submit/")
	routes = append(routes,"/profile/reply/create/")
	routes = append(routes,"/profile/reply/edit/submit/")
	routes = append(routes,"/profile/reply/delete/submit/")
	routes = append(routes,"/arcane/tower/")
	routes = append(routes,"/magical/kingdom/")
	routes = append(routes,"/insert/name/here/")
	addEmptyRoutesToCustom(routes, router)
	b.Run("seventy-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
}*/

func BenchmarkParserSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = parse_message("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
}

func BenchmarkBBCodePluginWithRegexpSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_regex_parse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func BenchmarkBBCodePluginWithoutCodeTagSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_parse_without_code("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func BenchmarkBBCodePluginWithFullParserSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcode_full_parse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func TestLevels(t *testing.T) {
	levels := getLevels(40)
	for level, score := range levels {
		sscore := strconv.FormatFloat(score, 'f', -1, 64)
		t.Log("Level: " + strconv.Itoa(level) + " Score: " + sscore)
	}
}

// TO-DO: Make this compatible with the changes to the router
/*
func TestStaticRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	static_w := httptest.NewRecorder()
	static_req := httptest.NewRequest("get","/static/global.js",bytes.NewReader(nil))
	static_handler := http.HandlerFunc(route_static)

	static_handler.ServeHTTP(static_w,static_req)
	if static_w.Code != 200 {
		t.Fatal(static_w.Body)
	}
}
*/

/*func TestTopicAdminRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.CascadeGet(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}

	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_req_admin := topic_req
	topic_req_admin.AddCookie(&admin_uid_cookie)
	topic_req_admin.AddCookie(&admin_session_cookie)
	topic_handler := http.HandlerFunc(route_topic_id)

	topic_handler.ServeHTTP(topic_w,topic_req_admin)
	if topic_w.Code != 200 {
		fmt.Println(topic_w.Body)
		panic("HTTP Error!")
	}
	fmt.Println("No problems found in the topic-admin route!")
}*/

/*func TestTopicGuestRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	topic_w := httptest.NewRecorder()
	topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
	topic_handler := http.HandlerFunc(route_topic_id)

	topic_handler.ServeHTTP(topic_w,topic_req)
	if topic_w.Code != 200 {
		fmt.Println(topic_w.Body)
		panic("HTTP Error!")
	}
	fmt.Println("No problems found in the topic-guest route!")
}*/

// TO-DO: Make these routes compatible with the changes to the router
/*
func TestForumsAdminRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.CascadeGet(1)
	if err != nil {
		t.Fatal(err)
	}
	if !admin.Is_Admin {
		t.Fatal("UID1 is not an admin")
	}
	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	forums_w := httptest.NewRecorder()
	forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forums_req_admin := forums_req
	forums_req_admin.AddCookie(&admin_uid_cookie)
	forums_req_admin.AddCookie(&admin_session_cookie)
	forums_handler := http.HandlerFunc(route_forums)

	forums_handler.ServeHTTP(forums_w,forums_req_admin)
	if forums_w.Code != 200 {
		t.Fatal(forums_w.Body)
	}
}

func TestForumsGuestRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	forums_w := httptest.NewRecorder()
	forums_req := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
	forums_handler := http.HandlerFunc(route_forums)

	forums_handler.ServeHTTP(forums_w,forums_req)
	if forums_w.Code != 200 {
		t.Fatal(forums_w.Body)
	}
}
*/

/*func TestForumAdminRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.CascadeGet(1)
	if err != nil {
		panic(err)
	}
	if !admin.Is_Admin {
		panic("UID1 is not an admin")
	}
	admin_uid_cookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
	admin_session_cookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/1",bytes.NewReader(nil))
	forum_req_admin := forum_req
	forum_req_admin.AddCookie(&admin_uid_cookie)
	forum_req_admin.AddCookie(&admin_session_cookie)
	forum_handler := http.HandlerFunc(route_forum)

	forum_handler.ServeHTTP(forum_w,forum_req_admin)
	if forum_w.Code != 200 {
		fmt.Println(forum_w.Body)
		panic("HTTP Error!")
	}
	fmt.Println("No problems found in the forum-admin route!")
}*/

/*func TestForumGuestRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	forum_w := httptest.NewRecorder()
	forum_req := httptest.NewRequest("get","/forum/2",bytes.NewReader(nil))
	forum_handler := http.HandlerFunc(route_forum)

	forum_handler.ServeHTTP(forum_w,forum_req)
	if forum_w.Code != 200 {
		fmt.Println(forum_w.Body)
		panic("HTTP Error!")
	}
	fmt.Println("No problems found in the forum-guest route!")
}*/

/*func TestAlerts(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}
	db = db_test
	alert_w := httptest.NewRecorder()
	alert_req := httptest.NewRequest("get","/api/?action=get&module=alerts&format=json",bytes.NewReader(nil))
	alert_handler := http.HandlerFunc(route_api)
	//testdb.StubQuery()
	testdb.SetQueryFunc(func(query string) (result sql.Rows, err error) {
		cols := []string{"asid","actor","targetUser","event","elementType","elementID"}
		rows := `1,1,0,like,post,5
		1,1,0,friend_invite,user,2`
		return testdb.RowsFromCSVString(cols,rows), nil
	})

	alert_handler.ServeHTTP(alert_w,alert_req)
	fmt.Println(alert_w.Body)
	if alert_w.Code != 200 {
		panic("HTTP Error!")
	}

	fmt.Println("No problems found in the alert handler!")
	db = db_prod
}*/

func TestSplittyThing(t *testing.T) {
	var extra_data string
	var path string = "/pages/hohoho"
	t.Log("Raw Path:",path)
	if path[len(path) - 1] != '/' {
		extra_data = path[strings.LastIndexByte(path,'/') + 1:]
		path = path[:strings.LastIndexByte(path,'/') + 1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extra_data)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extra_data))

	t.Log("Splitty thing test")
	path = "/topics/"
	extra_data = ""
	t.Log("Raw Path:",path)
	if path[len(path) - 1] != '/' {
		extra_data = path[strings.LastIndexByte(path,'/') + 1:]
		path = path[:strings.LastIndexByte(path,'/') + 1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extra_data)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extra_data))
}
