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
	pi := Page{"Topic Blah","topic",user,noticeList,replyList,topic}
	pi2 := Page{"Topic Blah","topic",admin,noticeList,replyList,topic}
	w := ioutil.Discard
	
	b.Run("compiled_writer_collated_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(pi2,w)
		}
	})
	b.Run("compiled_writer_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic2(pi2,w)
		}
	})
	b.Run("compiled_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w.Write([]byte(template_topic3(pi2)))
		}
	})
	b.Run("interpreted_useradmin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi2)
		}
	})
	b.Run("compiled_writer_collated_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic(pi,w)
		}
	})
	b.Run("compiled_writer_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			template_topic2(pi,w)
		}
	})
	b.Run("compiled_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w.Write([]byte(template_topic3(pi)))
		}
	})
	b.Run("interpreted_userguest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			templates.ExecuteTemplate(w,"topic.html", pi)
		}
	})
}
