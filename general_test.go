package main
import "testing"
import "io/ioutil"
import "html/template"

func BenchmarkTemplates(b *testing.B) {
	user := User{0,"Bob",0,false,false,false,false,false,false,"",false,"","","","",""}
	admin := User{1,"Admin",0,true,true,true,true,true,false,"",false,"","","","",""}
	var noticeList map[int]string = make(map[int]string)
	noticeList[0] = "test"
	
	topic := TopicUser{0,"Lol",template.HTML("Hey everyone!"),0,false,false,"",0,"","","",no_css_tmpl,0,"","","",""}
	
	var replyList map[int]interface{} = make(map[int]interface{})
	replyList[0] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[1] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[2] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[3] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[4] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[5] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[6] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[7] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[8] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	replyList[9] = Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""}
	
	var replyList2 []interface{}
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	replyList2 = append(replyList2, Reply{0,0,"Hey everyone!",template.HTML("Hey everyone!"),0,"","",0,0,"",no_css_tmpl,0,"","","",""})
	
	pi := Page2{"Topic Blah","topic",user,noticeList,replyList,topic}
	pi2 := Page2{"Topic Blah","topic",admin,noticeList,replyList,topic}
	pi3 := Page{"Topic Blah","topic",user,noticeList,replyList2,topic}
	pi4 := Page{"Topic Blah","topic",admin,noticeList,replyList2,topic}
	w := ioutil.Discard
	
	b.Run("compiled_sliceloop_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(pi4,w)
		}
	})
	b.Run("compiled_maploop_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic_maploop(pi2,w)
		}
	})
	b.Run("interpreted_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi2)
		}
	})
	
	b.Run("compiled_sliceloop_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(pi3,w)
		}
	})
	b.Run("compiled_maploop_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic_maploop(pi,w)
		}
	})
	b.Run("interpreted_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi)
		}
	})
}
