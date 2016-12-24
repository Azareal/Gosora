package main
//import "fmt"
import "log"
import "bytes"
import "math/rand"
import "testing"
import "net/http"
import "net/http/httptest"
import "io/ioutil"
import "html/template"

func BenchmarkTopicTemplate(b *testing.B) {
	b.ReportAllocs()
	
	user := User{0,"Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","",""}
	admin := User{1,"Admin","admin@localhost",0,true,true,true,true,true,false,AllPerms,"",false,"","","","",""}
	var noticeList map[int]string = make(map[int]string)
	noticeList[0] = "test"
	
	topic := TopicUser{0,"Lol",template.HTML("Hey everyone!"),0,false,false,"",0,"","","",no_css_tmpl,0,"","","",""}
	
	var replyList []Reply
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList = append(replyList, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	
	tpage := TopicPage{"Topic Blah","topic",user,noticeList,replyList,topic,false}
	tpage2 := TopicPage{"Topic Blah","topic",admin,noticeList,replyList,topic,false}
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
}

func BenchmarkTopicsTemplate(b *testing.B) {
	b.ReportAllocs()
	
	user := User{0,"Bob","bob@localhost",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","",""}
	admin := User{1,"Admin","admin@localhost",0,true,true,true,true,true,false,AllPerms,"",false,"","","","",""}
	var noticeList map[int]string = make(map[int]string)
	noticeList[0] = "test"
	
	var topicList []interface{}
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	topicList = append(topicList, TopicUser{0,"Hey everyone!",template.HTML("Hey everyone!"),0,false,false,"0000-00-00 00:00:00",1,"open","Admin","",no_css_tmpl,0,"Admin","","",""})
	
	tpage := Page{"Topic Blah","topic",user,noticeList,topicList,0}
	tpage2 := Page{"Topic Blah","topic",admin,noticeList,topicList,0}
	w := ioutil.Discard
	
	b.Run("compiled_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topics(tpage2,w)
		}
	})
	b.Run("interpreted_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topics.html", tpage2)
		}
	})
	b.Run("compiled_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topics(tpage,w)
		}
	})
	b.Run("interpreted_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topics.html", tpage)
		}
	})
}

func BenchmarkRoute(b *testing.B) {
	b.ReportAllocs()
	
	admin_uid_cookie := http.Cookie{Name: "uid",Value: "1",Path: "/",MaxAge: year}
	// TO-DO: Stop hard-coding this value
	admin_session_cookie := http.Cookie{Name: "session",Value: "TKBh5Z-qEQhWDBnV6_XVmOhKAowMYPhHeRlrQjjbNc0QRrRiglvWOYFDc1AaMXQIywvEsyA2AOBRYUrZ5kvnGhThY1GhOW6FSJADnRWm_bI=",Path: "/",MaxAge: year}
	
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
	
	debug = false
	nogrouplog = true
	
	// init_database is a little noisy for a benchmark
	discard := ioutil.Discard
	log.SetOutput(discard)
	
	var err error
	init_database(err);
	external_sites["YT"] = "https://www.youtube.com/"
	hooks["trow_assign"] = nil
	hooks["rrow_assign"] = nil
	
	for name, body := range plugins {
		if body.Active {
			plugins[name].Init()
		}
	}
	
	b.Run("topic_admin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req_admin)
			//if topic_w.Code != 200 {
			//	fmt.Println(topic_w.Body)
			//	panic("HTTP Error!")
			//}
		}
	})
	b.Run("topic_guest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topic_w.Code = 200
			topic_w.Body.Reset()
			topic_handler.ServeHTTP(topic_w,topic_req)
		}
	})
	b.Run("topics_admin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req_admin)
		}
	})
	b.Run("topics_guest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//topics_w.Code = 200
			topics_w.Body.Reset()
			topics_handler.ServeHTTP(topics_w,topics_req)
		}
	})
	b.Run("forum_admin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req_admin)
		}
	})
	b.Run("forum_guest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forum_w.Code = 200
			forum_w.Body.Reset()
			forum_handler.ServeHTTP(forum_w,forum_req)
		}
	})
	b.Run("forums_admin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req_admin)
		}
	})
	b.Run("forums_guest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//forums_w.Code = 200
			forums_w.Body.Reset()
			forums_handler.ServeHTTP(forums_w,forums_req)
		}
	})
}

func addEmptyRoutesToMux(routes []string, serveMux *http.ServeMux) {
	for _, route := range routes {
		serveMux.HandleFunc(route, func(_ http.ResponseWriter,_ *http.Request){})
	}
}

func BenchmarkRouter(b *testing.B) {
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

/*func TestRoute(t *testing.T) {
}*/