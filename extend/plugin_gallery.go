package extend

import c "github.com/Azareal/Gosora/common"

func init() {
	c.Plugins.Add(&c.Plugin{UName: "gallery", Name: "Gallery", Author: "Azareal", URL: "https://github.com/Azareal", Init: initGallery, Deactivate: deactivateGallery})
}

// init_heythere is separate from init() as we don't want the plugin to run if the plugin is disabled
func initGallery(plugin *c.Plugin) error {
	plugin.AddHook("topic_reply_row_assign", galleryReply)
	return nil
}

func deactivateGallery(plugin *c.Plugin) {
	plugin.RemoveHook("topic_reply_row_assign", galleryReply)
}

func galleryReply(data ...interface{}) interface{} {
	currentUser := data[0].(*c.TopicPage).Header.CurrentUser
	reply := data[1].(*c.ReplyUser)
	reply.Content = "Hey there, " + currentUser.Name + "!"
	reply.ContentHtml = "Hey there, " + currentUser.Name + "!"
	reply.Tag = "Auto"
	return nil
}
