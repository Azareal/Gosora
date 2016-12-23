package main
//import "fmt"
import "log"
import "bytes"
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
	
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/topics/", route_topics)
	b.Run("topics_guest_plus_router", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			topics_w.Body.Reset()
			serveMux.ServeHTTP(topics_w,topics_req)
		}
	})
}

/*func TestRoute(b *testing.T) {
	
}*/