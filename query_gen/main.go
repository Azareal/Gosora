/* WIP Under Construction */
package main

import "log"
import "./lib"

func main() {
	log.Println("Running the query generator")
	for _, adapter := range qgen.DB_Registry {
		log.Println("Building the queries for the " + adapter.GetName() + " adapter")
		write_statements(adapter)
		adapter.Write()
	}
}

func write_statements(adapter qgen.DB_Adapter) error {
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

func write_selects(adapter qgen.DB_Adapter) error {
	// url_prefix and url_name will be removed from this query in a later commit
	adapter.SimpleSelect("get_user","users","name, group, is_super_admin, avatar, message, url_prefix, url_name, level","uid = ?","")
		
	// Looking for get_topic? Your statement is in another castle
	
	adapter.SimpleSelect("get_reply","replies","content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount","rid = ?","")
		
	adapter.SimpleSelect("login","users","uid, name, password, salt","name = ?","")
		
	adapter.SimpleSelect("get_password","users","password,salt","uid = ?","")
	
	adapter.SimpleSelect("username_exists","users","name","name = ?","")
	
	
	adapter.SimpleSelect("get_settings","settings","name, content, type","","")
	
	adapter.SimpleSelect("get_setting","settings","content, type","name = ?","")
	
	adapter.SimpleSelect("get_full_setting","settings","name, type, constraints","name = ?","")
	
	adapter.SimpleSelect("is_plugin_active","plugins","active","uname = ?","")
	
	adapter.SimpleSelect("get_users","users","uid, name, group, active, is_super_admin, avatar","","")
	
	adapter.SimpleSelect("is_theme_default","themes","default","uname = ?","")
	
	adapter.SimpleSelect("get_modlogs","moderation_logs","action, elementID, elementType, ipaddress, actorID, doneAt","","")
	
	adapter.SimpleSelect("get_reply_tid","replies","tid","rid = ?","")
	
	adapter.SimpleSelect("get_topic_fid","topics","parentID","tid = ?","")
	
	adapter.SimpleSelect("get_user_reply_uid","users_replies","uid","rid = ?","")
	
	adapter.SimpleSelect("has_liked_topic","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'topics'","")
	
	adapter.SimpleSelect("has_liked_reply","likes","targetItem","sentBy = ? and targetItem = ? and targetType = 'replies'","")
	
	adapter.SimpleSelect("get_user_name","users","name","uid = ?","")
	
	adapter.SimpleSelect("get_user_rank","users","group, is_super_admin","uid = ?","")
	
	adapter.SimpleSelect("get_user_active","users","active","uid = ?","")
	
	adapter.SimpleSelect("get_user_group","users","group","uid = ?","")
	
	adapter.SimpleSelect("get_emails_by_user","emails","email, validated","uid = ?","")
	
	adapter.SimpleSelect("get_topic_basic","topics","title, content","tid = ?","")
	
	adapter.SimpleSelect("get_activity_entry","activity_stream","actor, targetUser, event, elementType, elementID","asid = ?","")
	
	return nil
}

func write_left_joins(adapter qgen.DB_Adapter) error {
	adapter.SimpleLeftJoin("get_topic_list","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	
	adapter.SimpleLeftJoin("get_topic_user","topics","users","topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level","topics.createdBy = users.uid","tid = ?","")
	
	adapter.SimpleLeftJoin("get_topic_by_reply","replies","topics","topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, topics.data","replies.tid = topics.tid","rid = ?","")
	
	adapter.SimpleLeftJoin("get_topic_replies","replies","users","replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress","replies.createdBy = users.uid","tid = ?","")
	
	adapter.SimpleLeftJoin("get_forum_topics","topics","users","topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar","topics.createdBy = users.uid","topics.parentID = ?","topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc")
	
	adapter.SimpleLeftJoin("get_profile_replies","users_replies","users","users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group","users_replies.createdBy = users.uid","users_replies.uid = ?","")
	
	return nil
}

func write_inner_joins(adapter qgen.DB_Adapter) error {
	adapter.SimpleInnerJoin("get_watchers","activity_stream","activity_subscriptions","activity_subscriptions.user","activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor","asid = ?","")
	
	return nil
}

func write_inserts(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsert("create_topic","topics","parentID,title,content,parsed_content,createdAt,lastReplyAt,ipaddress,words,createdBy","?,?,?,?,NOW(),NOW(),?,?,?")
	
	adapter.SimpleInsert("create_report","topics","title,content,parsed_content,createdAt,lastReplyAt,createdBy,data,parentID,css_class","?,?,?,NOW(),NOW(),?,?,1,'report'")

	adapter.SimpleInsert("create_reply","replies","tid,content,parsed_content,createdAt,ipaddress,words,createdBy","?,?,?,NOW(),?,?,?")
	
	adapter.SimpleInsert("create_action_reply","replies","tid,actionType,ipaddress,createdBy","?,?,?,?")
	
	adapter.SimpleInsert("create_like","likes","weight, targetItem, targetType, sentBy","?,?,?,?")
	
	adapter.SimpleInsert("add_activity","activity_stream","actor,targetUser,event,elementType,elementID","?,?,?,?,?")
	
	adapter.SimpleInsert("notify_one","activity_stream_matches","watcher,asid","?,?")
	
	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	adapter.SimpleInsert("register","users","name, email, password, salt, group, is_super_admin, session, active, message","?,?,?,?,?,0,?,?,''")
	
	adapter.SimpleInsert("add_email","emails","email, uid, validated, token","?,?,?,?")
	
	adapter.SimpleInsert("create_profile_reply","users_replies","uid,content,parsed_content,createdAt,createdBy","?,?,?,NOW(),?")
	
	adapter.SimpleInsert("add_subscription","activity_subscriptions","user,targetID,targetType,level","?,?,?,2")
	
	adapter.SimpleInsert("create_forum","forums","name, desc, active, preset","?,?,?,?")
	
	adapter.SimpleInsert("add_forum_perms_to_forum","forums_permissions","gid,fid,preset,permissions","?,?,?,?")
	
	adapter.SimpleInsert("add_plugin","plugins","uname,active","?,?")
	
	adapter.SimpleInsert("add_theme","themes","uname,default","?,?")

	
	adapter.SimpleInsert("create_group","users_groups","name, tag, is_admin, is_mod, is_banned, permissions","?,?,?,?,?,?")
	
	adapter.SimpleInsert("add_modlog_entry","moderation_logs","action, elementID, elementType, ipaddress, actorID, doneAt","?,?,?,?,?,NOW()")
	
	adapter.SimpleInsert("add_adminlog_entry","administration_logs","action, elementID, elementType, ipaddress, actorID, doneAt","?,?,?,?,?,NOW()")
	
	return nil
}

func write_replaces(adapter qgen.DB_Adapter) error {
	adapter.SimpleReplace("add_forum_perms_to_group","forums_permissions","gid,fid,preset,permissions","?,?,?,?")
	
	return nil
}

func write_updates(adapter qgen.DB_Adapter) error {
	adapter.SimpleUpdate("add_replies_to_topic","topics","postCount = postCount + ?, lastReplyAt = NOW()","tid = ?")
	
	adapter.SimpleUpdate("remove_replies_from_topic","topics","postCount = postCount - ?","tid = ?")
	
	adapter.SimpleUpdate("add_topics_to_forum","forums","topicCount = topicCount + ?","fid = ?")
	
	adapter.SimpleUpdate("remove_topics_from_forum","forums","topicCount = topicCount - ?","fid = ?")
	
	adapter.SimpleUpdate("update_forum_cache","forums","lastTopic = ?, lastTopicID = ?, lastReplyer = ?, lastReplyerID = ?, lastTopicTime = NOW()","fid = ?")

	adapter.SimpleUpdate("add_likes_to_topic","topics","likeCount = likeCount + ?","tid = ?")
	
	adapter.SimpleUpdate("add_likes_to_reply","replies","likeCount = likeCount + ?","rid = ?")
	
	adapter.SimpleUpdate("edit_topic","topics","title = ?, content = ?, parsed_content = ?, is_closed = ?","tid = ?")
	
	adapter.SimpleUpdate("edit_reply","replies","content = ?, parsed_content = ?","rid = ?")
	
	adapter.SimpleUpdate("stick_topic","topics","sticky = 1","tid = ?")
	
	adapter.SimpleUpdate("unstick_topic","topics","sticky = 0","tid = ?")
	
	adapter.SimpleUpdate("update_last_ip","users","last_ip = ?","uid = ?")

	adapter.SimpleUpdate("update_session","users","session = ?","uid = ?")
	
	adapter.SimpleUpdate("logout","users","session = ''","uid = ?")

	adapter.SimpleUpdate("set_password","users","password = ?, salt = ?","uid = ?")
	
	adapter.SimpleUpdate("set_avatar","users","avatar = ?","uid = ?")
	
	adapter.SimpleUpdate("set_username","users","name = ?","uid = ?")
	
	adapter.SimpleUpdate("change_group","users","group = ?","uid = ?")
	
	adapter.SimpleUpdate("activate_user","users","active = 1","uid = ?")
	
	adapter.SimpleUpdate("update_user_level","users","level = ?","uid = ?")
	
	adapter.SimpleUpdate("increment_user_score","users","score = score + ?","uid = ?")
	
	adapter.SimpleUpdate("increment_user_posts","users","posts = posts + ?","uid = ?")
	
	adapter.SimpleUpdate("increment_user_bigposts","users","posts = posts + ?, bigposts = bigposts + ?","uid = ?")
	
	adapter.SimpleUpdate("increment_user_megaposts","users","posts = posts + ?, bigposts = bigposts + ?, megaposts = megaposts + ?","uid = ?")
	
	adapter.SimpleUpdate("increment_user_topics","users","topics =  topics + ?","uid = ?")

	adapter.SimpleUpdate("edit_profile_reply","users_replies","content = ?, parsed_content = ?","rid = ?")
	
	//delete_forum_stmt, err = db.Prepare("delete from forums where fid = ?")
	adapter.SimpleUpdate("delete_forum","forums","name= '', active = 0","fid = ?")
	
	adapter.SimpleUpdate("update_forum","forums","name = ?, desc = ?, active = ?, preset = ?","fid = ?")
	
	adapter.SimpleUpdate("update_setting","settings","content = ?","name = ?")
	
	adapter.SimpleUpdate("update_plugin","plugins","active = ?","uname = ?")
	
	adapter.SimpleUpdate("update_theme","themes","default = ?","uname = ?")
	
	adapter.SimpleUpdate("update_user","users","name = ?, email = ?, group = ?","uid = ?")

	adapter.SimpleUpdate("update_group_perms","users_groups","permissions = ?","gid = ?")
	
	adapter.SimpleUpdate("update_group_rank","users_groups","is_admin = ?, is_mod = ?, is_banned = ?","gid = ?")
	
	adapter.SimpleUpdate("update_group","users_groups","name = ?, tag = ?","gid = ?")
	
	return nil
}

func write_deletes(adapter qgen.DB_Adapter) error {
	adapter.SimpleDelete("delete_reply","replies","rid = ?")
	
	adapter.SimpleDelete("delete_topic","topics","tid = ?")
	
	
	adapter.SimpleDelete("delete_profile_reply","users_replies","rid = ?")
	
	adapter.SimpleDelete("delete_forum_perms_by_forum","forums_permissions","fid = ?")
	
	return nil
}