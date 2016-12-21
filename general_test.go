package main
import "testing"
import "io/ioutil"
import "html/template"

func BenchmarkTemplates(b *testing.B) {
	user := User{0,"Bob",0,false,false,false,false,false,false,GuestPerms,"",false,"","","","",""}
	admin := User{1,"Admin",0,true,true,true,true,true,false,AllPerms,"",false,"","","","",""}
	var noticeList map[int]string = make(map[int]string)
	noticeList[0] = "test"
	
	topic := TopicUser{0,"Lol",template.HTML("Hey everyone!"),0,false,false,"",0,"","","",no_css_tmpl,0,"","","",""}
	
	var replyList []interface{}
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
	
	
	var replyList2 []Reply
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
	
	pi := Page{"Topic Blah","topic",user,noticeList,replyList,topic}
	pi2 := Page{"Topic Blah","topic",admin,noticeList,replyList,topic}
	tpage := TopicPage{"Topic Blah","topic",user,noticeList,replyList2,topic,false}
	tpage2 := TopicPage{"Topic Blah","topic",admin,noticeList,replyList2,topic,false}
	w := ioutil.Discard
	
	b.Run("compiled_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic_2(pi2,w)
		}
	})
	/*b.Run("interpreted_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi2)
		}
	})*/
	b.Run("compiled_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic_2(pi,w)
		}
	})
	/*b.Run("interpreted_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi)
		}
	})*/
	
	
	b.Run("compiled_customstruct_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(tpage2,w)
		}
	})
	b.Run("interpreted_customstruct_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", tpage2)
		}
	})
	b.Run("compiled_customstruct_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(tpage,w)
		}
	})
	b.Run("interpreted_customstruct_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", tpage)
		}
	})
}
