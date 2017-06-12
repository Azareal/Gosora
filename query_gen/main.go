/* WIP Under Construction */
package main

import "log"

var db_registry []DB_Adapter
var blank_order []DB_Order

type DB_Column struct
{
	Table string
	Left string // Could be a function or a column, so I'm naming this Left
	Alias string // aka AS Blah, if it's present
	Type string // function or column
}

type DB_Field struct
{
	Name string
	Type string
}

type DB_Where struct
{
	LeftTable string
	LeftColumn string
	RightTable string
	RightColumn string
	Operator string
	LeftType string
	RightType string
}

type DB_Joiner struct
{
	LeftTable string
	LeftColumn string
	RightTable string
	RightColumn string
	Operator string
}

type DB_Order struct
{
	Column string
	Order string
}

type DB_Token struct {
	Contents string
	Type string // function, operator, column, number, string, substitute
}

type DB_Setter struct {
	Column string
	Expr []DB_Token // Simple expressions, the innards of functions are opaque for now.
}

type DB_Adapter interface {
	get_name() string
	simple_insert(string,string,string,string) error
	simple_replace(string,string,string,string) error
	simple_update(string,string,string,string) error
	simple_delete(string,string,string) error
	purge(string,string) error
	simple_select(string,string,string,string,string/*,int,int*/) error
	simple_left_join(string,string,string,string,string,string,string/*,int,int*/) error
	simple_inner_join(string,string,string,string,string,string,string/*,int,int*/) error
	write() error
	
	// TO-DO: Add a simple query builder
}

func main() {
	log.Println("Running the query generator")
	for _, adapter := range db_registry {
		log.Println("Building the queries for the " + adapter.get_name() + " adapter")
		write_statements(adapter)
		adapter.write()
	}
}

func write_statements(adapter DB_Adapter) error {
	err := write_selects(adapter)
	if err != nil {
		return err
	}
	err = write_left_joins(adapter)
	if err != nil {
		return err
	}
	err = write_inner_joins(adapter)
	if err != nil {
		return err
	}
	err = write_inserts(adapter)
	if err != nil {
		return err
	}
	err = write_replaces(adapter)
	if err != nil {
		return err
	}
	err = write_updates(adapter)
	if err != nil {
		return err
	}
	err = write_deletes(adapter)
	if err != nil {
		return err
	}
	return nil
}

func write_selects(adapter DB_Adapter) error {
	// url_prefix and url_name will be removed from this query in a later commit
	adapter.simple_select("get_user","users","name, group, is_super_admin, avatar, message, url_prefix, url_name, level","uid = ?","")
	
	adapter.simple_select("get_full_user","users","name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip","uid = ?","")
		
	adapter.simple_select("get_topic","topics","title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data","tid = ?","")
	
	adapter.simple_select("get_reply","replies","content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount","rid = ?","")
		
	adapter.simple_select("login","users","uid, name, password, salt","name = ?","")
		
	adapter.simple_select("get_password","users","password,salt","uid = ?","")
	
	adapter.simple_select("username_exists","users","name","name = ?","")
	
	
	adapter.simple_select("get_settings","settings","name, content, type","","")
	
	adapter.simple_select("get_setting","settings","content, type","name = ?","")
	
	adapter.simple_select("get_full_setting","settings","name, type, constraints","name = ?","")
	
	adapter.simple_select("is_plugin_active","plugins","active","uname = ?","")
	
	adapter.simple_select("get_users","users","uid, name, group, active, is_super_admin, avatar","","")
	
	adapter.simple_select("is_theme_default","themes","default","uname = ?","")
	
	adapter.simple_select("get_modlogs","moderation_logs","action, elementID, elementType, ipaddress, actorID, doneAt","","")
	
	adapter.simple_select("get_reply_tid","replies","tid","rid = ?","")
	
	adapter.simple_select("get_topic_fid","topics","parentID","tid = ?","")
	
	adapter.simple_select("get_user_reply_uid","users_replies","uid","rid = ?","")
	
	adapter.simple_select("has_liked_topic","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'topics'","")
	
	adapter.simple_select("has_liked_reply","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'replies'","")
	
	adapter.simple_select("get_user_name","users","name","uid = ?","")
	
	adapter.simple_select("get_user_rank","users","group, is_super_admin","uid = ?","")
	
	adapter.simple_select("get_user_active","users","active","uid = ?","")
	
	adapter.simple_select("get_user_group","users","group","uid = ?","")
	
	adapter.simple_select("get_emails_by_user","emails","email, validated","uid = ?","")
	
	adapter.simple_select("get_topic_basic","topics","title, content","tid = ?","")
	
	adapter.simple_select("get_activity_entry","activity_stream","actor, targetUser, event, elementType, elementID","asid = ?","")
	
	return nil
}

func write_left_joins(adapter DB_Adapter) error {
	adapter.simple_left_join("get_topic_list","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	
	adapter.simple_left_join("get_topic_user","topics","users","topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level","topics.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_topic_by_reply","replies","topics","topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, topics.data","replies.tid = topics.tid","rid = ?","")
	
	adapter.simple_left_join("get_topic_replies","replies","users","replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress","replies.createdBy = users.uid","tid = ?","")
	
	adapter.simple_left_join("get_forum_topics","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","topics.parentID = ?","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc")
	
	adapter.simple_left_join("get_profile_replies","users_replies","users","users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group","users_replies.createdBy = users.uid","users_replies.uid = ?","")
	
	return nil
}

func write_inner_joins(adapter DB_Adapter) error {
	adapter.simple_inner_join("get_watchers","activity_stream","activity_subscriptions","activity_subscriptions.user","activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor","asid = ?","")
	
	return nil
}

func write_inserts(adapter DB_Adapter) error {
	adapter.simple_insert("create_topic","topics","parentID,title,content,parsed_content,createdAt,lastReplyAt,ipaddress,words,createdBy","?,?,?,?,NOW(),NOW(),?,?,?")
	
	adapter.simple_insert("create_report","topics","title,content,parsed_content,createdAt,lastReplyAt,createdBy,data,parentID,css_class","?,?,?,NOW(),NOW(),?,?,1,'report'")

	adapter.simple_insert("create_reply","replies","tid,content,parsed_content,createdAt,ipaddress,words,createdBy","?,?,?,NOW(),?,?,?")
	
	adapter.simple_insert("create_action_reply","replies","tid,actionType,ipaddress,createdBy","?,?,?,?")
	
	adapter.simple_insert("create_like","likes","weight, targetItem, targetType, sentBy","?,?,?,?")
	
	adapter.simple_insert("add_activity","activity_stream","actor,targetUser,event,elementType,elementID","?,?,?,?,?")
	
	adapter.simple_insert("notify_one","activity_stream_matches","watcher,asid","?,?")
	
	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	adapter.simple_insert("register","users","name, email, password, salt, group, is_super_admin, session, active, message","?,?,?,?,?,0,?,?,''")
	
	adapter.simple_insert("add_email","emails","email, uid, validated, token","?,?,?,?")
	
	adapter.simple_insert("create_profile_reply","users_replies","uid,content,parsed_content,createdAt,createdBy","?,?,?,NOW(),?")
	
	adapter.simple_insert("add_subscription","activity_subscriptions","user,targetID,targetType,level","?,?,?,2")
	
	adapter.simple_insert("create_forum","forums","name, desc, active, preset","?,?,?,?")
	
	adapter.simple_insert("add_forum_perms_to_forum","forums_permissions","gid,fid,preset,permissions","?,?,?,?")
	
	adapter.simple_insert("add_plugin","plugins","uname,active","?,?")
	
	adapter.simple_insert("add_theme","themes","uname,default","?,?")

	
	adapter.simple_insert("create_group","users_groups","name, tag, is_admin, is_mod, is_banned, permissions","?,?,?,?,?,?")
	
	adapter.simple_insert("add_modlog_entry","moderation_logs","action, elementID, elementType, ipaddress, actorID, doneAt","?,?,?,?,?,NOW()")
	
	adapter.simple_insert("add_adminlog_entry","administration_logs","action, elementID, elementType, ipaddress, actorID, doneAt","?,?,?,?,?,NOW()")
	
	return nil
}

func write_replaces(adapter DB_Adapter) error {
	adapter.simple_replace("add_forum_perms_to_group","forums_permissions","gid,fid,preset,permissions","?,?,?,?")
	
	return nil
}

func write_updates(adapter DB_Adapter) error {
	adapter.simple_update("add_replies_to_topic","topics","postCount = postCount + ?, lastReplyAt = NOW()","tid = ?")
	
	adapter.simple_update("remove_replies_from_topic","topics","postCount = postCount - ?","tid = ?")
	
	adapter.simple_update("add_topics_to_forum","forums","topicCount = topicCount + ?","fid = ?")
	
	adapter.simple_update("remove_topics_from_forum","forums","topicCount = topicCount - ?","fid = ?")
	
	adapter.simple_update("update_forum_cache","forums","lastTopic = ?, lastTopicID = ?, lastReplyer = ?, lastReplyerID = ?, lastTopicTime = NOW()","fid = ?")

	adapter.simple_update("add_likes_to_topic","topics","likeCount = likeCount + ?","tid = ?")
	
	adapter.simple_update("add_likes_to_reply","replies","likeCount = likeCount + ?","rid = ?")
	
	adapter.simple_update("edit_topic","topics","title = ?, content = ?, parsed_content = ?, is_closed = ?","tid = ?")
	
	adapter.simple_update("edit_reply","replies","content = ?, parsed_content = ?","rid = ?")
	
	adapter.simple_update("stick_topic","topics","sticky = 1","tid = ?")
	
	adapter.simple_update("unstick_topic","topics","sticky = 0","tid = ?")
	
	adapter.simple_update("update_last_ip","users","last_ip = ?","uid = ?")

	adapter.simple_update("update_session","users","session = ?","uid = ?")
	
	adapter.simple_update("logout","users","session = ''","uid = ?")

	adapter.simple_update("set_password","users","password = ?, salt = ?","uid = ?")
	
	adapter.simple_update("set_avatar","users","avatar = ?","uid = ?")
	
	adapter.simple_update("set_username","users","name = ?","uid = ?")
	
	adapter.simple_update("change_group","users","group = ?","uid = ?")
	
	adapter.simple_update("activate_user","users","active = 1","uid = ?")
	
	adapter.simple_update("update_user_level","users","level = ?","uid = ?")
	
	adapter.simple_update("increment_user_score","users","score = score + ?","uid = ?")
	
	adapter.simple_update("increment_user_posts","users","posts = posts + ?","uid = ?")
	
	adapter.simple_update("increment_user_bigposts","users","posts = posts + ?, bigposts = bigposts + ?","uid = ?")
	
	adapter.simple_update("increment_user_megaposts","users","posts = posts + ?, bigposts = bigposts + ?, megaposts = megaposts + ?","uid = ?")
	
	adapter.simple_update("increment_user_topics","users","topics =  topics + ?","uid = ?")

	adapter.simple_update("edit_profile_reply","users_replies","content = ?, parsed_content = ?","rid = ?")
	
	//delete_forum_stmt, err = db.Prepare("delete from forums where fid = ?")
	adapter.simple_update("delete_forum","forums","name= '', active = 0","fid = ?")
	
	adapter.simple_update("update_forum","forums","name = ?, desc = ?, active = ?, preset = ?","fid = ?")
	
	adapter.simple_update("update_setting","settings","content = ?","name = ?")
	
	adapter.simple_update("update_plugin","plugins","active = ?","uname = ?")
	
	adapter.simple_update("update_theme","themes","default = ?","uname = ?")
	
	adapter.simple_update("update_user","users","name = ?, email = ?, group = ?","uid = ?")

	adapter.simple_update("update_group_perms","users_groups","permissions = ?","gid = ?")
	
	adapter.simple_update("update_group_rank","users_groups","is_admin = ?, is_mod = ?, is_banned = ?","gid = ?")
	
	adapter.simple_update("update_group","users_groups","name = ?, tag = ?","gid = ?")
	
	return nil
}

func write_deletes(adapter DB_Adapter) error {
	adapter.simple_delete("delete_reply","replies","rid = ?")
	
	adapter.simple_delete("delete_topic","topics","tid = ?")
	
	
	adapter.simple_delete("delete_profile_reply","users_replies","rid = ?")
	
	adapter.simple_delete("delete_forum_perms_by_forum","forums_permissions","fid = ?")
	
	return nil
}