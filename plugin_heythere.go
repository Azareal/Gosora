package main

import "./common"

func init() {
	common.Plugins.Add(&common.Plugin{UName: "heythere", Name: "Hey There", Author: "Azareal", URL: "https://github.com/Azareal", Init: initHeythere, Deactivate: deactivateHeythere})
}

// init_heythere is separate from init() as we don't want the plugin to run if the plugin is disabled
func initHeythere() error {
	common.Plugins["heythere"].AddHook("topic_reply_row_assign", heythereReply)
	return nil
}

func deactivateHeythere() {
	common.Plugins["heythere"].RemoveHook("topic_reply_row_assign", heythereReply)
}

func heythereReply(data ...interface{}) interface{} {
	currentUser := data[0].(*common.TopicPage).Header.CurrentUser
	reply := data[1].(*common.ReplyUser)
	reply.Content = "Hey there, " + currentUser.Name + "!"
	reply.ContentHtml = "Hey there, " + currentUser.Name + "!"
	reply.Tag = "Auto"
	return nil
}
