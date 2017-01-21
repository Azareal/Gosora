package main
import "os"
import "log"
import "bytes"
import "strconv"
import "math/rand"
import "testing"
import "net/http"
import "net/http/httptest"
import "io/ioutil"
import "html/template"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
//import "github.com/husobee/vestigo"
import "runtime/pprof"

var gloinited bool = false
func gloinit() {
	debug = false
	nogrouplog = true
	
	// init_database is a little noisy for a benchmark
	//discard := ioutil.Discard
	//log.SetOutput(discard)
	
	var err error
	init_database(err)
	db.SetMaxOpenConns(64)
	external_sites["YT"] = "https://www.youtube.com/"
	hooks["trow_assign"] = nil
	hooks["rrow_assign"] = nil
	//log.SetOutput(os.Stdout)
	gloinited = true
}

func BenchmarkTopicTemplate(b *testing.B) {
	b.ReportAllocs()
	
	user := User{0,"Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","","",0,0,"127.0.0.1"}
	admin := User{1,"Admin","admin@localhost",0,true,true,true,true,true,false,AllPerms,"",false,"","","","","",-1,58,"127.0.0.1"}
	noticeList := []string{"test"}
	
	topic := TopicUser{Title: "Lol",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"}
	
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","","",0,"127.0.0.1"})
	
	tpage := TopicPage{"Topic Blah",user,noticeList,replyList,topic,1,1,false}
	tpage2 := TopicPage{"Topic Blah",admin,noticeList,replyList,topic,1,1,false}
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

func BenchmarkTopicsTemplate(b *testing.B) {
	b.ReportAllocs()
	
	user := User{0,"Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","","",0,0,"127.0.0.1"}
	admin := User{1,"Admin","admin@localhost",0,true,true,true,true,true,false,AllPerms,"",false,"","","","","",-1,58,"127.0.0.1"}
	noticeList := []string{"test"}
	
	var topicList []TopicUser
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	topicList = append(topicList, TopicUser{Title: "Hey everyone!",Content: template.HTML("Hey everyone!"),CreatedBy: 1,CreatedAt: "0000-00-00 00:00:00",ParentID: 1,CreatedByName:"Admin",Css: no_css_tmpl,Tag: "Admin", Level: 58, IpAddress: "127.0.0.1"})
	
	w := ioutil.Discard
	tpage := TopicsPage{"Topic Blah",user,noticeList,topicList,nil}
	tpage2 := TopicsPage{"Topic Blah",admin,noticeList,topicList,nil}
	
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
	
	b.RunParallel(func(pb *testing.PB) {
		static_w := httptest.NewRecorder()
		static_req := httptest.NewRequest("get","/static/global.js",bytes.NewReader(nil))
		static_handler := http.HandlerFunc(route_static)
		for pb.Next() {
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
		}
	})
}

/*func BenchmarkStaticRouteParallelWithPlugins(b *testing.B) {
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
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
		}
	})
}*/

func BenchmarkTopicAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}
	
	b.RunParallel(func(pb *testing.PB) {
		admin_uid_cookie := http.Cookie{Name: "uid",Value: "1",Path: "/",MaxAge: year}
		// TO-DO: Stop hard-coding this value. Seriously.
		admin_session_cookie := http.Cookie{Name: "session",Value: "TKBh5Z-qEQhWDBnV6_XVmOhKAowMYPhHeRlrQjjbNc0QRrRiglvWOYFDc1AaMXQIywvEsyA2AOBRYUrZ5kvnGhThY1GhOW6FSJADnRWm_bI=",Path: "/",MaxAge: year}
		
		topic_w := httptest.NewRecorder()
		topic_req := httptest.NewRequest("get","/topic/1",bytes.NewReader(nil))
		topic_req_admin := topic_req
		topic_req_admin.AddCookie(&admin_uid_cookie)
		topic_req_admin.AddCookie(&admin_session_cookie)
		topic_handler := http.HandlerFunc(route_topic_id)
		
		for pb.Next() {
			//topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
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
		topic_handler := http.HandlerFunc(route_topic_id)
		for pb.Next() {
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
	})
}


func BenchmarkForumsAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}
	
	b.RunParallel(func(pb *testing.PB) {
		admin_uid_cookie := http.Cookie{Name: "uid",Value: "1",Path: "/",MaxAge: year}
		// TO-DO: Stop hard-coding this value. Seriously.
		admin_session_cookie := http.Cookie{Name: "session",Value: "TKBh5Z-qEQhWDBnV6_XVmOhKAowMYPhHeRlrQjjbNc0QRrRiglvWOYFDc1AaMXQIywvEsyA2AOBRYUrZ5kvnGhThY1GhOW6FSJADnRWm_bI=",Path: "/",MaxAge: year}
		
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
		admin_uid_cookie := http.Cookie{Name: "uid",Value: "1",Path: "/",MaxAge: year}
		// TO-DO: Stop hard-coding this value. Seriously.
		admin_session_cookie := http.Cookie{Name: "session",Value: "TKBh5Z-qEQhWDBnV6_XVmOhKAowMYPhHeRlrQjjbNc0QRrRiglvWOYFDc1AaMXQIywvEsyA2AOBRYUrZ5kvnGhThY1GhOW6FSJADnRWm_bI=",Path: "/",MaxAge: year}
		
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


func BenchmarkRoutesSerial(b *testing.B) {
	b.ReportAllocs()
	
	admin_uid_cookie := http.Cookie{Name: "uid",Value: "1",Path: "/",MaxAge: year}
	// TO-DO: Stop hard-coding this value. Seriously.
	admin_session_cookie := http.Cookie{Name: "session",Value: "TKBh5Z-qEQhWDBnV6_XVmOhKAowMYPhHeRlrQjjbNc0QRrRiglvWOYFDc1AaMXQIywvEsyA2AOBRYUrZ5kvnGhThY1GhOW6FSJADnRWm_bI=",Path: "/",MaxAge: year}
	
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
	
	/*f, err := os.Create("routes_bench_cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)*/
	//defer pprof.StopCPUProfile()
	//pprof.StopCPUProfile()
	
	b.Run("static_recorder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			static_w.Body.Reset()
			static_handler.ServeHTTP(static_w,static_req)
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
		/*f, err := os.Create("routes_bench_forums_cpu_2.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)*/
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
		/*f, err := os.Create("routes_bench_topic_cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)*/
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
		/*f, err := os.Create("routes_bench_topic_cpu_2.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)*/
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
		/*f, err := os.Create("routes_bench_forums_cpu_2.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)*/
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
		//pprof.StopCPUProfile()
	})
}

func BenchmarkQueryTopicParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}
	
	b.RunParallel(func(pb *testing.PB) {
		topic := TopicUser{Css: no_css_tmpl}
		var content string
		var is_super_admin bool
		var group int
		for pb.Next() {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name, users.avatar, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, topics.ipaddress from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group, &topic.URLPrefix, &topic.URLName, &topic.Level, &topic.IpAddress)
			if err == sql.ErrNoRows {
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
		topic := TopicUser{Css: no_css_tmpl}
		var content string
		var is_super_admin bool
		var group int
		for pb.Next() {
			err := get_topic_user_stmt.QueryRow(1).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group, &topic.URLPrefix, &topic.URLName, &topic.Level, &topic.IpAddress)
			if err == sql.ErrNoRows {
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
	topic := TopicUser{Css: no_css_tmpl}
	var content string
	var is_super_admin bool
	var group int
	b.Run("topic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name, users.avatar, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, topics.ipaddress from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&topic.Title, &content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.CreatedByName, &topic.Avatar, &is_super_admin, &group, &topic.URLPrefix, &topic.URLName, &topic.Level, &topic.IpAddress)
			if err == sql.ErrNoRows {
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
	replyItem := Reply{Css: no_css_tmpl}
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

func addEmptyRoutesToMux(routes []string, serveMux *http.ServeMux) {
	for _, route := range routes {
		serveMux.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkDefaultGoRouter(b *testing.B) {
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

/*func addEmptyRoutesToVestigo(routes []string, router *vestigo.Router) {
	for _, route := range routes {
		router.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkVestigoRouter(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("get","/topics/",bytes.NewReader(nil))
	routes := make([]string, 0)
	
	routes = append(routes,"/test/")
	router := vestigo.NewRouter()
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
	router = vestigo.NewRouter()
	addEmptyRoutesToVestigo(routes, router)
	b.Run("five-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
	routes = append(routes,"/panel/plugins/")
	routes = append(routes,"/panel/groups/")
	routes = append(routes,"/panel/settings/")
	routes = append(routes,"/panel/users/")
	routes = append(routes,"/panel/forums/")
	addEmptyRoutesToVestigo(routes, router)
	b.Run("ten-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("twenty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("thirty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("forty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("fifty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("sixty-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
	
	router = vestigo.NewRouter()
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
	addEmptyRoutesToVestigo(routes, router)
	b.Run("seventy-routes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req = httptest.NewRequest("get",routes[rand.Intn(len(routes))],bytes.NewReader(nil))
			router.ServeHTTP(w,req)
		}
	})
}*/

func addEmptyRoutesToCustom(routes []string, router *Router) {
	for _, route := range routes {
		router.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkCustomRouter(b *testing.B) {
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
}

func BenchmarkParser(b *testing.B) {
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

func BenchmarkBBCodePluginWithRegexp(b *testing.B) {
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

func BenchmarkBBCodePluginWithoutCodeTag(b *testing.B) {
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

func BenchmarkBBCodePluginWithFullParser(b *testing.B) {
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
		log.Print("Level: " + strconv.Itoa(level) + " Score: " + sscore)
	}
}

/*func TestRoute(t *testing.T) {
	
}*/