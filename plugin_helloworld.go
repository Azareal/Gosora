package main

func init() {
	plugins["helloworld"] = NewPlugin("helloworld", "Hello World", "Azareal", "http://github.com/Azareal", "", "", "", initHelloworld, nil, deactivateHelloworld, nil, nil)
}

// init_helloworld is separate from init() as we don't want the plugin to run if the plugin is disabled
func initHelloworld() error {
	plugins["helloworld"].AddHook("rrow_assign", helloworldReply)
	return nil
}

func deactivateHelloworld() {
	plugins["helloworld"].RemoveHook("rrow_assign", helloworldReply)
}

func helloworldReply(data interface{}) interface{} {
	reply := data.(*Reply)
	reply.Content = "Hello World!"
	reply.ContentHtml = "Hello World!"
	reply.Tag = "Auto"
	return nil
}
