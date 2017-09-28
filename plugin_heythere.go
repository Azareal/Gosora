package main

func init() {
	plugins["heythere"] = NewPlugin("heythere", "Hey There", "Azareal", "http://github.com/Azareal", "", "", "", initHeythere, nil, deactivateHeythere, nil, nil)
}

// init_heythere is separate from init() as we don't want the plugin to run if the plugin is disabled
func initHeythere() error {
	plugins["heythere"].AddHook("topic_reply_row_assign", heythereReply)
	return nil
}

func deactivateHeythere() {
	plugins["heythere"].RemoveHook("topic_reply_row_assign", heythereReply)
}

func heythereReply(data ...interface{}) interface{} {
	currentUser := data[0].(*TopicPage).CurrentUser
	reply := data[1].(*ReplyUser)
	reply.Content = "Hey there, " + currentUser.Name + "!"
	reply.ContentHtml = "Hey there, " + currentUser.Name + "!"
	reply.Tag = "Auto"
	return nil
}
