package main
import "html/template"

func init() {
	plugins["helloworld"] = Plugin{"helloworld","Hello World","Azareal","http://github.com/Azareal","",false,"",init_helloworld}
}

// init_helloworld is separate from init() as we don't want the plugin to run if the plugin is disabled
func init_helloworld() {
	add_hook("rrow_assign", helloworld_reply)
}

func helloworld_reply(data interface{}) interface{} {
	reply := data.(Reply)
	reply.Content = "Hello World!"
	reply.ContentHtml = template.HTML("Hello World!")
	reply.Tag = "Automated"
	return reply
}