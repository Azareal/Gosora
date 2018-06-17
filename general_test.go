package main

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"./common"
	"./install/install"
	"./query_gen/lib"
	"./routes"
	//"github.com/husobee/vestigo"
)

//var dbTest *sql.DB
var dbProd *sql.DB
var gloinited bool
var installAdapter install.InstallAdapter

func ResetTables() (err error) {
	err = installAdapter.InitDatabase()
	if err != nil {
		return err
	}

	err = installAdapter.TableDefs()
	if err != nil {
		return err
	}

	err = installAdapter.CreateAdmin()
	if err != nil {
		return err
	}

	return installAdapter.InitialData()
}

func gloinit() (err error) {
	// TODO: Make these configurable via flags to the go test command
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false
	common.Dev.TemplateDebug = false
	qgen.LogPrepares = false
	//nogrouplog = true
	common.StartTime = time.Now()

	err = common.LoadConfig()
	if err != nil {
		return err
	}
	err = common.ProcessConfig()
	if err != nil {
		return err
	}

	common.Themes, err = common.NewThemeList()
	if err != nil {
		return err
	}
	common.SwitchToTestDB()

	var ok bool
	installAdapter, ok = install.Lookup(dbAdapter)
	if !ok {
		return errors.New("We couldn't find the adapter '" + dbAdapter + "'")
	}
	installAdapter.SetConfig(common.DbConfig.Host, common.DbConfig.Username, common.DbConfig.Password, common.DbConfig.Dbname, common.DbConfig.Port)

	err = ResetTables()
	if err != nil {
		return err
	}
	err = InitDatabase()
	if err != nil {
		return err
	}
	err = afterDBInit()
	if err != nil {
		return err
	}

	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		return err
	}
	gloinited = true
	return nil
}

func init() {
	err := gloinit()
	if err != nil {
		log.Print("Something bad happened")
		log.Fatal(err)
	}
}

// TODO: Swap out LocalError for a panic for this?
func BenchmarkTopicAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	admin, err := common.Users.Get(1)
	if err != nil {
		b.Fatal(err)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	adminUIDCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: common.Year}
	adminSessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: common.Year}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topicW := httptest.NewRecorder()
			topicReqAdmin := httptest.NewRequest("get", "/topic/hm.1", bytes.NewReader(nil))
			topicReqAdmin.AddCookie(&adminUIDCookie)
			topicReqAdmin.AddCookie(&adminSessionCookie)

			// Deal with the session stuff, etc.
			user, ok := common.PreRoute(topicW, topicReqAdmin)
			if !ok {
				b.Fatal("Mysterious error!")
			}
			//topicW.Body.Reset()
			routes.ViewTopic(topicW, topicReqAdmin, user, "1")
			if topicW.Code != 200 {
				b.Log(topicW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkTopicAdminRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	admin, err := common.Users.Get(1)
	if err != nil {
		b.Fatal(err)
	}
	if !admin.IsAdmin {
		b.Fatal("UID1 is not an admin")
	}
	adminUIDCookie := http.Cookie{Name: "uid", Value: "1", Path: "/", MaxAge: common.Year}
	adminSessionCookie := http.Cookie{Name: "session", Value: admin.Session, Path: "/", MaxAge: common.Year}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topicW := httptest.NewRecorder()
			topicReqAdmin := httptest.NewRequest("get", "/topic/hm.1", bytes.NewReader(nil))
			topicReqAdmin.AddCookie(&adminUIDCookie)
			topicReqAdmin.AddCookie(&adminSessionCookie)
			topicReqAdmin.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			topicReqAdmin.Header.Set("Host", "localhost")
			topicReqAdmin.Host = "localhost"
			//topicW.Body.Reset()
			router.ServeHTTP(topicW, topicReqAdmin)
			if topicW.Code != 200 {
				b.Log(topicW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkTopicAdminRouteParallelAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallel(b)
}

func BenchmarkTopicAdminRouteParallelWithRouterAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallelWithRouter(b)
}

func BenchmarkTopicAdminRouteParallelAltAlt(b *testing.B) {
	BenchmarkTopicAdminRouteParallel(b)
}

func BenchmarkTopicGuestRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topicW := httptest.NewRecorder()
			topicReq := httptest.NewRequest("get", "/topic/hm.1", bytes.NewReader(nil))
			//topicW.Body.Reset()
			routes.ViewTopic(topicW, topicReq, common.GuestUser, "1")
			if topicW.Code != 200 {
				b.Log(topicW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkTopicGuestRouteParallelDebugMode(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = true
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topicW := httptest.NewRecorder()
			topicReq := httptest.NewRequest("get", "/topic/hm.1", bytes.NewReader(nil))
			//topicW.Body.Reset()
			routes.ViewTopic(topicW, topicReq, common.GuestUser, "1")
			if topicW.Code != 200 {
				b.Log(topicW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkTopicGuestRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	/*f, err := os.Create("BenchmarkTopicGuestRouteParallelWithRouter.prof")
	if err != nil {
		b.Fatal(err)
	}
	pprof.StartCPUProfile(f)*/

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topicW := httptest.NewRecorder()
			topicReq := httptest.NewRequest("GET", "/topic/hm.1", bytes.NewReader(nil))
			topicReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			topicReq.Header.Set("Host", "localhost")
			topicReq.Host = "localhost"
			//topicW.Body.Reset()
			router.ServeHTTP(topicW, topicReq)
			if topicW.Code != 200 {
				b.Log(topicW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	//defer pprof.StopCPUProfile()
	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkBadRouteGuestRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			badW := httptest.NewRecorder()
			badReq := httptest.NewRequest("GET", "/garble/haa", bytes.NewReader(nil))
			badReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			badReq.Header.Set("Host", "localhost")
			badReq.Host = "localhost"
			router.ServeHTTP(badW, badReq)
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkTopicsGuestRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			listW := httptest.NewRecorder()
			listReq := httptest.NewRequest("GET", "/topics/", bytes.NewReader(nil))
			listReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			listReq.Header.Set("Host", "localhost")
			listReq.Host = "localhost"
			router.ServeHTTP(listW, listReq)
			if listW.Code != 200 {
				b.Log(listW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkForumsGuestRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			listW := httptest.NewRecorder()
			listReq := httptest.NewRequest("GET", "/forums/", bytes.NewReader(nil))
			listReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			listReq.Header.Set("Host", "localhost")
			listReq.Host = "localhost"
			router.ServeHTTP(listW, listReq)
			if listW.Code != 200 {
				b.Log(listW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

func BenchmarkForumGuestRouteParallelWithRouter(b *testing.B) {
	b.ReportAllocs()
	var err error
	if !gloinited {
		err = gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		b.Fatal(err)
	}
	prev := common.Dev.DebugMode
	prev2 := common.Dev.SuperDebug
	common.Dev.DebugMode = false
	common.Dev.SuperDebug = false

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			listW := httptest.NewRecorder()
			listReq := httptest.NewRequest("GET", "/forum/general.2", bytes.NewReader(nil))
			listReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36")
			listReq.Header.Set("Host", "localhost")
			listReq.Host = "localhost"
			router.ServeHTTP(listW, listReq)
			if listW.Code != 200 {
				b.Log(listW.Body)
				b.Fatal("HTTP Error!")
			}
		}
	})

	common.Dev.DebugMode = prev
	common.Dev.SuperDebug = prev2
}

// TODO: Make these routes compatible with the changes to the router
/*
func BenchmarkForumsAdminRouteParallel(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.Get(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		adminUidCookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		adminSessionCookie := http.Cookie{Name:"session",Value: admin.Session,Path:"/",MaxAge: year}

		forumsW := httptest.NewRecorder()
		forumsReq := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forumsReq_admin := forums_req
		forumsReq_admin.AddCookie(&adminUidCookie)
		forumsReq_admin.AddCookie(&adminSessionCookie)
		forumsHandler := http.HandlerFunc(route_forums)

		for pb.Next() {
			forumsW.Body.Reset()
			forumsHandler.ServeHTTP(forumsW,forumsReqAdmin)
		}
	})
}

func BenchmarkForumsAdminRouteParallelProf(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		gloinit()
	}

	b.RunParallel(func(pb *testing.PB) {
		admin, err := users.Get(1)
		if err != nil {
			panic(err)
		}
		if !admin.Is_Admin {
			panic("UID1 is not an admin")
		}
		adminUidCookie := http.Cookie{Name:"uid",Value:"1",Path:"/",MaxAge: year}
		adminSessionCookie := http.Cookie{Name:"session",Value: admin.Session,Path: "/",MaxAge: year}

		forumsW := httptest.NewRecorder()
		forumsReq := httptest.NewRequest("get","/forums/",bytes.NewReader(nil))
		forumsReqAdmin := forumsReq
		forumsReqAdmin.AddCookie(&admin_uid_cookie)
		forumsReqAdmin.AddCookie(&admin_session_cookie)
		forumsHandler := http.HandlerFunc(route_forums)
		f, err := os.Create("cpu_forums_admin_parallel.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		for pb.Next() {
			forumsW.Body.Reset()
			forumsHandler.ServeHTTP(forumsW,forumsReqAdmin)
		}
		pprof.StopCPUProfile()
	})
}

func BenchmarkRoutesSerial(b *testing.B) {
	b.ReportAllocs()
	admin, err := users.Get(1)
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
			//	b.Print(static_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
	})

	b.Run("topic_admin_recorder", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu.prof")
		//if err != nil {
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	b.Print(topic_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder", func(b *testing.B) {
		f, err := os.Create("routes_bench_topic_cpu_2.prof")
		if err != nil {
			b.Fatal(err)
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
		//	b.Fatal(err)
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
		//	b.Fatal(err)
		//}
		//pprof.StartCPUProfile(f)
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	b.Print(topic_w.Body)
			//	b.Fatal("HTTP Error!")
			//}
		}
		//pprof.StopCPUProfile()
	})
	b.Run("topic_guest_recorder_with_plugins", func(b *testing.B) {
		//f, err := os.Create("routes_bench_topic_cpu_2.prof")
		//if err != nil {
		//	b.Fatal(err)
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
		//	b.Fatal(err)
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
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu common.TopicUser
		for pb.Next() {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
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
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		var tu common.TopicUser

		getTopicUser, err := qgen.Builder.SimpleLeftJoin("topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level", "topics.createdBy = users.uid", "tid = ?", "", "")
		if err != nil {
			b.Fatal(err)
		}
		defer getTopicUser.Close()

		for pb.Next() {
			err := getTopicUser.QueryRow(1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
			if err == ErrNoRows {
				b.Fatal("No rows found!")
				return
			} else if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkUserGet(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		var err error
		for pb.Next() {
			_, err = common.Users.Get(1)
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkUserBypassGet(b *testing.B) {
	b.ReportAllocs()
	if !gloinited {
		err := gloinit()
		if err != nil {
			b.Fatal(err)
		}
	}

	// Bypass the cache and always hit the database
	b.RunParallel(func(pb *testing.PB) {
		var err error
		for pb.Next() {
			_, err = common.Users.BypassGet(1)
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkQueriesSerial(b *testing.B) {
	b.ReportAllocs()
	var tu common.TopicUser
	b.Run("topic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.QueryRow("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?", 1).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
			if err == ErrNoRows {
				b.Fatal("No rows found!")
				return
			} else if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
	b.Run("topic_replies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				b.Fatal(err)
				return
			}
			defer rows.Close()

			for rows.Next() {
			}
			err = rows.Err()
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})

	var replyItem common.ReplyUser
	var isSuperAdmin bool
	var group int
	b.Run("topic_replies_scan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rows, err := db.Query("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?", 1)
			if err != nil {
				b.Fatal(err)
				return
			}
			for rows.Next() {
				err := rows.Scan(&replyItem.ID, &replyItem.Content, &replyItem.CreatedBy, &replyItem.CreatedAt, &replyItem.LastEdit, &replyItem.LastEditBy, &replyItem.Avatar, &replyItem.CreatedByName, &isSuperAdmin, &group, &replyItem.URLPrefix, &replyItem.URLName, &replyItem.Level, &replyItem.IPAddress)
				if err != nil {
					b.Fatal(err)
					return
				}
			}
			defer rows.Close()

			err = rows.Err()
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

// TODO: Take the attachment system into account in these parser benches
func BenchmarkParserSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("", 0, "")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("Hey everyone, how's it going?", 0, "")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("Hey everyone, how's it going? :)", 0, "")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("Hey everyone, how's it going? :):):):):)", 0, "")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("Hey everyone, how's it going? :):):):):):):):):):)", 0, "")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.ParseMessage("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)", 0, "")
		}
	})
}

func BenchmarkBBCodePluginWithRegexpSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeRegexParse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func BenchmarkBBCodePluginWithoutCodeTagSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeParseWithoutCode("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func BenchmarkBBCodePluginWithFullParserSerial(b *testing.B) {
	b.ReportAllocs()
	b.Run("empty_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("")
		}
	})
	b.Run("short_post", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("Hey everyone, how's it going?")
		}
	})
	b.Run("one_smily", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("Hey everyone, how's it going? :)")
		}
	})
	b.Run("five_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("Hey everyone, how's it going? :):):):):)")
		}
	})
	b.Run("ten_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("Hey everyone, how's it going? :):):):):):):):):):)")
		}
	})
	b.Run("twenty_smilies", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("Hey everyone, how's it going? :):):):):):):):):):):):):):):):):):):):)")
		}
	})
	b.Run("one_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("[b]H[/b]ey everyone, how's it going?")
		}
	})
	b.Run("five_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b]eryone, how's it going?")
		}
	})
	b.Run("ten_bold", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bbcodeFullParse("[b]H[/b][b]e[/b][b]y[/b] [b]e[/b][b]v[/b][b]e[/b][b]r[/b][b]y[/b][b]o[/b][b]n[/b]e, how's it going?")
		}
	})
}

func TestLevels(t *testing.T) {
	levels := common.GetLevels(40)
	for level, score := range levels {
		sscore := strconv.FormatFloat(score, 'f', -1, 64)
		t.Log("Level: " + strconv.Itoa(level) + " Score: " + sscore)
	}
}

// TODO: Make this compatible with the changes to the router
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

	admin, err := users.Get(1)
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
		t.Print(topic_w.Body)
		t.Fatal("HTTP Error!")
	}
	t.Print("No problems found in the topic-admin route!")
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
		t.Print(topic_w.Body)
		t.Fatal("HTTP Error!")
	}
	t.Print("No problems found in the topic-guest route!")
}*/

// TODO: Make these routes compatible with the changes to the router
/*
func TestForumsAdminRoute(t *testing.T) {
	if !gloinited {
		gloinit()
	}
	if !plugins_inited {
		init_plugins()
	}

	admin, err := users.Get(1)
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

	admin, err := users.Get(1)
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
		t.Print(forum_w.Body)
		t.Fatal("HTTP Error!")
	}
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
		t.Print(forum_w.Body)
		t.Fatal("HTTP Error!")
	}
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
	t.Print(alert_w.Body)
	if alert_w.Code != 200 {
		t.Fatal("HTTP Error!")
	}
	db = db_prod
}*/

func TestSplittyThing(t *testing.T) {
	var extraData string
	var path = "/pages/hohoho"
	t.Log("Raw Path:", path)
	if path[len(path)-1] != '/' {
		extraData = path[strings.LastIndexByte(path, '/')+1:]
		path = path[:strings.LastIndexByte(path, '/')+1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extraData)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extraData))

	t.Log("Splitty thing test")
	path = "/topics/"
	extraData = ""
	t.Log("Raw Path:", path)
	if path[len(path)-1] != '/' {
		extraData = path[strings.LastIndexByte(path, '/')+1:]
		path = path[:strings.LastIndexByte(path, '/')+1]
	}
	t.Log("Path:", path)
	t.Log("Extra Data:", extraData)
	t.Log("Path Bytes:", []byte(path))
	t.Log("Extra Data Bytes:", []byte(extraData))
}
