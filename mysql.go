/* Copyright Azareal 2016 - 2017 */
package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "log"
import "fmt"
import "strconv"
import "encoding/json"

var db *sql.DB
var db_version string
var db_collation string = "utf8mb4_general_ci"

var get_user_stmt *sql.Stmt
var get_full_user_stmt *sql.Stmt
var get_topic_list_stmt *sql.Stmt
var get_topic_user_stmt *sql.Stmt
var get_topic_stmt *sql.Stmt
var get_topic_by_reply_stmt *sql.Stmt
var get_topic_replies_stmt *sql.Stmt
var get_topic_replies_offset_stmt *sql.Stmt
var get_reply_stmt *sql.Stmt
var get_forum_topics_stmt *sql.Stmt
var get_forum_topics_offset_stmt *sql.Stmt
var create_topic_stmt *sql.Stmt
var create_report_stmt *sql.Stmt
var create_reply_stmt *sql.Stmt
var create_action_reply_stmt *sql.Stmt
var add_replies_to_topic_stmt *sql.Stmt
var remove_replies_from_topic_stmt *sql.Stmt
var add_topics_to_forum_stmt *sql.Stmt
var remove_topics_from_forum_stmt *sql.Stmt
var update_forum_cache_stmt *sql.Stmt
var create_like_stmt *sql.Stmt
var add_likes_to_topic_stmt *sql.Stmt
var add_likes_to_reply_stmt *sql.Stmt
var add_activity_stmt *sql.Stmt
var notify_watchers_stmt *sql.Stmt
var notify_one_stmt *sql.Stmt
var add_subscription_stmt *sql.Stmt
var edit_topic_stmt *sql.Stmt
var edit_reply_stmt *sql.Stmt
var delete_reply_stmt *sql.Stmt
var delete_topic_stmt *sql.Stmt
var stick_topic_stmt *sql.Stmt
var unstick_topic_stmt *sql.Stmt
var get_activity_feed_by_watcher_stmt *sql.Stmt
var update_last_ip_stmt *sql.Stmt
var login_stmt *sql.Stmt
var update_session_stmt *sql.Stmt
var logout_stmt *sql.Stmt
var set_password_stmt *sql.Stmt
var get_password_stmt *sql.Stmt
var set_avatar_stmt *sql.Stmt
var set_username_stmt *sql.Stmt
var add_email_stmt *sql.Stmt
var update_email_stmt *sql.Stmt
var verify_email_stmt *sql.Stmt
var register_stmt *sql.Stmt
var username_exists_stmt *sql.Stmt
var change_group_stmt *sql.Stmt
var activate_user_stmt *sql.Stmt
var update_user_level_stmt *sql.Stmt
var increment_user_score_stmt *sql.Stmt
var increment_user_posts_stmt *sql.Stmt
var increment_user_bigposts_stmt *sql.Stmt
var increment_user_megaposts_stmt *sql.Stmt
var increment_user_topics_stmt *sql.Stmt
var create_profile_reply_stmt *sql.Stmt
var edit_profile_reply_stmt *sql.Stmt
var delete_profile_reply_stmt *sql.Stmt

var create_forum_stmt *sql.Stmt
var delete_forum_stmt *sql.Stmt
var update_forum_stmt *sql.Stmt
var forum_entry_exists_stmt *sql.Stmt
var group_entry_exists_stmt *sql.Stmt
var delete_forum_perms_by_forum_stmt *sql.Stmt
var add_forum_perms_to_forum_stmt *sql.Stmt
var add_forum_perms_to_forum_admins_stmt *sql.Stmt
var add_forum_perms_to_forum_staff_stmt *sql.Stmt
var add_forum_perms_to_forum_members_stmt *sql.Stmt
var add_forum_perms_to_group_stmt *sql.Stmt
var update_setting_stmt *sql.Stmt
var add_plugin_stmt *sql.Stmt
var update_plugin_stmt *sql.Stmt
var update_user_stmt *sql.Stmt
var update_group_perms_stmt *sql.Stmt
var update_group_rank_stmt *sql.Stmt
var update_group_stmt *sql.Stmt
var create_group_stmt *sql.Stmt
var add_theme_stmt *sql.Stmt
var update_theme_stmt *sql.Stmt
var add_modlog_entry_stmt *sql.Stmt
var add_adminlog_entry_stmt *sql.Stmt

func init_database() (err error) {
	if(dbpassword != ""){
		dbpassword = ":" + dbpassword
	}
	
	// Open the database connection
	db, err = sql.Open("mysql",dbuser + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?collation=" + db_collation)
	if err != nil {
		return err
	}
	
	// Make sure that the connection is alive
	err = db.Ping()
	if err != nil {
		return err
	}
	
	// Fetch the database version
	db.QueryRow("SELECT VERSION()").Scan(&db_version)
	
	// Set the number of max open connections
	db.SetMaxOpenConns(64)
	
	/*log.Print("Preparing get_session statement.")
	get_session_stmt, err = db.Prepare("select `uid`,`name`,`group`,`is_super_admin`,`session`,`email`,`avatar`,`message`,`url_prefix`,`url_name`,`level`,`score`,`last_ip` from `users` where `uid` = ? and `session` = ? AND `session` <> ''")
	if err != nil {
		return err
	}*/
	
	log.Print("Preparing get_user statement.")
	get_user_stmt, err = db.Prepare("select `name`,`group`,`is_super_admin`,`avatar`,`message`,`url_prefix`,`url_name`,`level` from `users` where `uid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_full_user statement.")
	get_full_user_stmt, err = db.Prepare("select `name`,`group`,`is_super_admin`,`session`,`email`,`avatar`,`message`,`url_prefix`,`url_name`,`level`,`score`,`last_ip` from `users` where `uid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic_list statement.")
	get_topic_list_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic_user statement.")
	get_topic_user_stmt, err = db.Prepare("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level from topics left join users ON topics.createdBy = users.uid where tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic statement.")
	get_topic_stmt, err = db.Prepare("select title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount from topics where tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic_by_reply statement.")
	get_topic_by_reply_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount from replies left join topics on replies.tid = topics.tid where rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic_replies statement.")
	get_topic_replies_stmt, err = db.Prepare("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_topic_replies_offset statement.")
	get_topic_replies_offset_stmt, err = db.Prepare("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.actionType from replies left join users on replies.createdBy = users.uid where tid = ? limit ?, " + strconv.Itoa(items_per_page))
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_reply statement.")
	get_reply_stmt, err = db.Prepare("select content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount from replies where rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_forum_topics statement.")
	get_forum_topics_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid where topics.parentID = ? order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_forum_topics_offset statement.")
	get_forum_topics_offset_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, topics.likeCount, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid WHERE topics.parentID = ? order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC limit ?, " + strconv.Itoa(items_per_page))
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_topic statement.")
	create_topic_stmt, err = db.Prepare("insert into topics(parentID,title,content,parsed_content,createdAt,lastReplyAt,ipaddress,words,createdBy) VALUES(?,?,?,?,NOW(),NOW(),?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_report statement.")
	create_report_stmt, err = db.Prepare("INSERT INTO topics(title,content,parsed_content,createdAt,lastReplyAt,createdBy,data,parentID) VALUES(?,?,?,NOW(),NOW(),?,?,1)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_reply statement.")
	create_reply_stmt, err = db.Prepare("INSERT INTO replies(tid,content,parsed_content,createdAt,ipaddress,words,createdBy) VALUES(?,?,?,NOW(),?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_action_reply statement.")
	create_action_reply_stmt, err = db.Prepare("INSERT INTO replies(tid,actionType,ipaddress,createdBy) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_replies_to_topic statement.")
	add_replies_to_topic_stmt, err = db.Prepare("UPDATE topics SET postCount = postCount + ?, lastReplyAt = NOW() WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing remove_replies_from_topic statement.")
	remove_replies_from_topic_stmt, err = db.Prepare("UPDATE topics SET postCount = postCount - ? WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_topics_to_forum statement.")
	add_topics_to_forum_stmt, err = db.Prepare("UPDATE forums SET topicCount = topicCount + ? WHERE fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing remove_topics_from_forum statement.")
	remove_topics_from_forum_stmt, err = db.Prepare("UPDATE forums SET topicCount = topicCount - ? WHERE fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_forum_cache statement.")
	update_forum_cache_stmt, err = db.Prepare("UPDATE forums SET lastTopic = ?, lastTopicID = ?, lastReplyer = ?, lastReplyerID = ?, lastTopicTime = NOW() WHERE fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_like statement.")
	create_like_stmt, err = db.Prepare("INSERT INTO likes(weight, targetItem, targetType, sentBy) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_likes_to_topic statement.")
	add_likes_to_topic_stmt, err = db.Prepare("UPDATE topics SET likeCount = likeCount + ? WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_likes_to_reply statement.")
	add_likes_to_reply_stmt, err = db.Prepare("UPDATE replies SET likeCount = likeCount + ? WHERE rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_activity statement.")
	add_activity_stmt, err = db.Prepare("INSERT INTO activity_stream(actor,targetUser,event,elementType,elementID) VALUES(?,?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing notify_watchers statement.")
	notify_watchers_stmt, err = db.Prepare("INSERT INTO activity_stream_matches(watcher, asid) SELECT activity_subscriptions.user, activity_stream.asid FROM activity_stream INNER JOIN activity_subscriptions ON activity_subscriptions.targetType = activity_stream.elementType and activity_subscriptions.targetID = activity_stream.elementID and activity_subscriptions.user != activity_stream.actor where asid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing notify_one statement.")
	notify_one_stmt, err = db.Prepare("INSERT INTO activity_stream_matches(watcher,asid) VALUES(?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_subscription statement.")
	add_subscription_stmt, err = db.Prepare("INSERT INTO activity_subscriptions(user,targetID,targetType,level) VALUES(?,?,?,2)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing edit_topic statement.")
	edit_topic_stmt, err = db.Prepare("UPDATE topics SET title = ?, content = ?, parsed_content = ?, is_closed = ? WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing edit_reply statement.")
	edit_reply_stmt, err = db.Prepare("UPDATE replies SET content = ?, parsed_content = ? WHERE rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing delete_reply statement.")
	delete_reply_stmt, err = db.Prepare("DELETE FROM replies WHERE rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing delete_topic statement.")
	delete_topic_stmt, err = db.Prepare("DELETE FROM topics WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing stick_topic statement.")
	stick_topic_stmt, err = db.Prepare("UPDATE topics SET sticky = 1 WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing unstick_topic statement.")
	unstick_topic_stmt, err = db.Prepare("UPDATE topics SET sticky = 0 WHERE tid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_activity_feed_by_watcher statement.")
	get_activity_feed_by_watcher_stmt, err = db.Prepare("SELECT activity_stream_matches.asid, activity_stream.actor, activity_stream.targetUser, activity_stream.event, activity_stream.elementType, activity_stream.elementID FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_last_ip statement.")
	update_last_ip_stmt, err = db.Prepare("UPDATE users SET last_ip = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing login statement.")
	login_stmt, err = db.Prepare("SELECT `uid`,`name`,`password`,`salt` FROM `users` WHERE `name` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_session statement.")
	update_session_stmt, err = db.Prepare("UPDATE users SET session = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing logout statement.")
	logout_stmt, err = db.Prepare("UPDATE users SET session = '' WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing set_password statement.")
	set_password_stmt, err = db.Prepare("UPDATE users SET password = ?, salt = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing get_password statement.")
	get_password_stmt, err = db.Prepare("SELECT `password`,`salt` FROM `users` WHERE `uid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing set_avatar statement.")
	set_avatar_stmt, err = db.Prepare("UPDATE users SET avatar = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing set_username statement.")
	set_username_stmt, err = db.Prepare("UPDATE users SET name = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	// Add an admin version of register_stmt with more flexibility
	// create_account_stmt, err = db.Prepare("INSERT INTO 
	
	log.Print("Preparing register statement.")
	register_stmt, err = db.Prepare("INSERT INTO users(`name`,`email`,`password`,`salt`,`group`,`is_super_admin`,`session`,`active`,`message`) VALUES(?,?,?,?,?,0,?,?,'')")
	if err != nil {
		return err
	}
	
	log.Print("Preparing username_exists statement.")
	username_exists_stmt, err = db.Prepare("SELECT `name` FROM `users` WHERE `name` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing change_group statement.")
	change_group_stmt, err = db.Prepare("update `users` set `group` = ? where `uid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_email statement.")
	add_email_stmt, err = db.Prepare("INSERT INTO emails(`email`,`uid`,`validated`,`token`) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_email statement.")
	update_email_stmt, err = db.Prepare("UPDATE emails SET email = ?, uid = ?, validated = ?, token = ? WHERE email = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing verify_email statement.")
	verify_email_stmt, err = db.Prepare("UPDATE emails SET validated = 1, token = '' WHERE email = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing activate_user statement.")
	activate_user_stmt, err = db.Prepare("UPDATE users SET active = 1 WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_user_level statement.")
	update_user_level_stmt, err = db.Prepare("UPDATE users SET level = ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing increment_user_score statement.")
	increment_user_score_stmt, err = db.Prepare("UPDATE users SET score = score + ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing increment_user_posts statement.")
	increment_user_posts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing increment_user_bigposts statement.")
	increment_user_bigposts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ?, bigposts =  bigposts + ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing increment_user_megaposts statement.")
	increment_user_megaposts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ?, bigposts = bigposts + ?, megaposts =  megaposts + ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing increment_user_topics statement.")
	increment_user_topics_stmt, err = db.Prepare("UPDATE users SET topics =  topics + ? WHERE uid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_profile_reply statement.")
	create_profile_reply_stmt, err = db.Prepare("INSERT INTO users_replies(uid,content,parsed_content,createdAt,createdBy) VALUES(?,?,?,NOW(),?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing edit_profile_reply statement.")
	edit_profile_reply_stmt, err = db.Prepare("UPDATE users_replies SET content = ?, parsed_content = ? WHERE rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing delete_profile_reply statement.")
	delete_profile_reply_stmt, err = db.Prepare("DELETE FROM users_replies WHERE rid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_forum statement.")
	create_forum_stmt, err = db.Prepare("INSERT INTO forums(name,active,preset) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing delete_forum statement.")
	//delete_forum_stmt, err = db.Prepare("DELETE FROM forums WHERE fid = ?")
	delete_forum_stmt, err = db.Prepare("update forums set name= '', active = 0 where fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_forum statement.")
	update_forum_stmt, err = db.Prepare("update forums set name = ?, active = ?, preset = ? where fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing forum_entry_exists statement.")
	forum_entry_exists_stmt, err = db.Prepare("SELECT `fid` FROM `forums` WHERE `name` = '' order by fid asc limit 1")
	if err != nil {
		return err
	}
	
	log.Print("Preparing group_entry_exists statement.")
	group_entry_exists_stmt, err = db.Prepare("SELECT `gid` FROM `users_groups` WHERE `name` = '' order by gid asc limit 1")
	if err != nil {
		return err
	}
	
	log.Print("Preparing delete_forum_perms_by_forum statement.")
	delete_forum_perms_by_forum_stmt, err = db.Prepare("DELETE FROM forums_permissions WHERE fid = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_forum_perms_to_forum statement.")
	add_forum_perms_to_forum_stmt, err = db.Prepare("INSERT INTO forums_permissions(gid,fid,preset,permissions) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_forum_perms_to_forum_admins statement.")
	add_forum_perms_to_forum_admins_stmt, err = db.Prepare("INSERT INTO forums_permissions(gid,fid,preset,permissions) SELECT `gid`,? AS fid,? AS preset,? AS permissions FROM users_groups WHERE is_admin = 1")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_forum_perms_to_forum_staff statement.")
	add_forum_perms_to_forum_staff_stmt, err = db.Prepare("INSERT INTO forums_permissions(gid,fid,preset,permissions) SELECT `gid`,? AS fid,? AS preset,? AS permissions FROM users_groups WHERE is_admin = 0 AND is_mod = 1")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_forum_perms_to_forum_members statement.")
	add_forum_perms_to_forum_members_stmt, err = db.Prepare("INSERT INTO forums_permissions(gid,fid,preset,permissions) SELECT `gid`,? AS fid,? AS preset,? AS permissions FROM users_groups WHERE is_admin = 0 AND is_mod = 0 AND is_banned = 0")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_forum_perms_to_group statement.")
	add_forum_perms_to_group_stmt, err = db.Prepare("INSERT INTO forums_permissions(gid,fid,preset,permissions) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_setting statement.")
	update_setting_stmt, err = db.Prepare("UPDATE settings SET content = ? WHERE name = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_plugin statement.")
	add_plugin_stmt, err = db.Prepare("INSERT INTO plugins(uname,active) VALUES(?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_plugin statement.")
	update_plugin_stmt, err = db.Prepare("UPDATE plugins SET active = ? WHERE uname = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_theme statement.")
	add_theme_stmt, err = db.Prepare("INSERT INTO `themes`(`uname`,`default`) VALUES(?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_theme statement.")
	update_theme_stmt, err = db.Prepare("update `themes` set `default` = ? where `uname` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_user statement.")
	update_user_stmt, err = db.Prepare("update `users` set `name` = ?,`email` = ?,`group` = ? where `uid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_group_rank statement.")
	update_group_perms_stmt, err = db.Prepare("update `users_groups` set `permissions` = ? where `gid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_group_rank statement.")
	update_group_rank_stmt, err = db.Prepare("update `users_groups` set `is_admin` = ?, `is_mod` = ?, `is_banned` = ? where `gid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing update_group statement.")
	update_group_stmt, err = db.Prepare("update `users_groups` set `name` = ?, `tag` = ? where `gid` = ?")
	if err != nil {
		return err
	}
	
	log.Print("Preparing create_group statement.")
	create_group_stmt, err = db.Prepare("INSERT INTO users_groups(name,tag,is_admin,is_mod,is_banned,permissions) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_modlog_entry statement.")
	add_modlog_entry_stmt, err = db.Prepare("INSERT INTO moderation_logs(action,elementID,elementType,ipaddress,actorID,doneAt) VALUES(?,?,?,?,?,NOW())")
	if err != nil {
		return err
	}
	
	log.Print("Preparing add_adminlog_entry statement.")
	add_adminlog_entry_stmt, err = db.Prepare("INSERT INTO moderation_logs(action,elementID,elementType,ipaddress,actorID,doneAt) VALUES(?,?,?,?,?,NOW())")
	if err != nil {
		return err
	}
	
	log.Print("Loading the usergroups.")
	groups = append(groups, Group{ID:0,Name:"System"})
	
	rows, err := db.Query("select gid,name,permissions,is_mod,is_admin,is_banned,tag from users_groups")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	i := 1
	for ;rows.Next();i++ {
		group := Group{ID: 0,}
		err := rows.Scan(&group.ID, &group.Name, &group.PermissionsText, &group.Is_Mod, &group.Is_Admin, &group.Is_Banned, &group.Tag)
		if err != nil {
			return err
		}
		
		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if group.ID != i {
			log.Print("Stop physically deleting groups. You are messing up the IDs. Use the Group Manager or delete_group() instead x.x")
			fill_group_id_gap(i, group.ID)
		}
		
		err = json.Unmarshal(group.PermissionsText, &group.Perms)
		if err != nil {
			return err
		}
		if debug {
			log.Print(group.Name + ": ")
			fmt.Printf("%+v\n", group.Perms)
		}
		
		group.Perms.ExtData = make(map[string]bool)
		groups = append(groups, group)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	groupCapCount = i
	
	log.Print("Binding the Not Loggedin Group")
	GuestPerms = groups[6].Perms
	
	log.Print("Loading the forums.")
	log.Print("Adding the uncategorised forum")
	forums = append(forums, Forum{0,"Uncategorised",uncategorised_forum_visible,"all",0,"",0,"",0,""})
	
	//rows, err = db.Query("SELECT fid, name, active, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime FROM forums")
	rows, err = db.Query("select fid, name, active, preset, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime from forums order by fid asc")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	i = 1
	for ;rows.Next();i++ {
		forum := Forum{ID:0,Name:"",Active:true,Preset:"all"}
		err := rows.Scan(&forum.ID, &forum.Name, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			return err
		}
		
		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if forum.ID != i {
			log.Print("Stop physically deleting forums. You are messing up the IDs. Use the Forum Manager or delete_forum() instead x.x")
			fill_forum_id_gap(i, forum.ID)
		}
		
		/*if forum.LastTopicID != 0 {
			forum.LastTopicTime, err = relative_time(forum.LastTopicTime)
			if err != nil {
				return err
			}
		} else {
			forum.LastTopic = "None"
			forum.LastTopicTime = ""
		}*/
		
		if forum.Name == "" {
			if debug {
				log.Print("Adding a placeholder forum")
			}
		} else {
			log.Print("Adding the " + forum.Name + " forum")
		}
		forums = append(forums,forum)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	forumCapCount = i
	
	//log.Print("Adding the reports forum")
	//forums[-1] = Forum{-1,"Reports",false,0,"",0,"",0,""}
	
	log.Print("Loading the forum permissions")
	rows, err = db.Query("select gid, fid, permissions from forums_permissions order by gid asc, fid asc")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if debug {
		log.Print("Adding the forum permissions")
	}
	// Temporarily store the forum perms in a map before transferring it to a much faster slice
	forum_perms = make(map[int]map[int]ForumPerms)
	for rows.Next() {
		var gid, fid int
		var perms []byte
		var pperms ForumPerms
		err := rows.Scan(&gid, &fid, &perms)
		if err != nil {
			return err
		}
		err = json.Unmarshal(perms, &pperms)
		if err != nil {
			return err
		}
		pperms.ExtData = make(map[string]bool)
		pperms.Overrides = true
		_, ok := forum_perms[gid]
		if !ok {
			forum_perms[gid] = make(map[int]ForumPerms)
		}
		forum_perms[gid][fid] = pperms
	}
	for gid, _ := range groups {
		if debug {
			log.Print("Adding the forum permissions for Group #" + strconv.Itoa(gid) + " - " + groups[gid].Name)
		}
		//groups[gid].Forums = append(groups[gid].Forums,BlankForumPerms) // GID 0. I sometimes wish MySQL's AUTO_INCREMENT would start at zero
		for fid, _ := range forums {
			forum_perm, ok := forum_perms[gid][fid]
			if ok {
				// Override group perms
				//log.Print("Overriding permissions for forum #" + strconv.Itoa(fid))
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			} else {
				// Inherit from Group
				//log.Print("Inheriting from default for forum #" + strconv.Itoa(fid))
				forum_perm = BlankForumPerms
				groups[gid].Forums = append(groups[gid].Forums,forum_perm)
			}
			
			if forum_perm.Overrides {
				if forum_perm.ViewTopic {
					groups[gid].CanSee = append(groups[gid].CanSee, fid)
				}
			} else if groups[gid].Perms.ViewTopic {
				groups[gid].CanSee = append(groups[gid].CanSee, fid)
			}
		}
		//fmt.Printf("%+v\n", groups[gid].CanSee)
		//fmt.Printf("%+v\n", groups[gid].Forums)
        //fmt.Println(len(groups[gid].CanSee))
		//fmt.Println(len(groups[gid].Forums))
	}
	
	log.Print("Loading the settings.")
	rows, err = db.Query("select name, content, type, constraints from settings")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var sname string
	var scontent string
	var stype string
	var sconstraints string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype, &sconstraints)
		if err != nil {
			return err
		}
		errmsg := parseSetting(sname, scontent, stype, sconstraints)
		if errmsg != "" {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	
	log.Print("Loading the plugins.")
	rows, err = db.Query("select uname, active from plugins")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var uname string
	var active bool
	for rows.Next() {
		err := rows.Scan(&uname, &active)
		if err != nil {
			return err
		}
		
		// Was the plugin deleted at some point?
		plugin, ok := plugins[uname]
		if !ok {
			continue
		}
		plugin.Active = active
		plugins[uname] = plugin
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	
	log.Print("Loading the themes.")
	rows, err = db.Query("select `uname`, `default` from `themes`")
	if err != nil {
		return err
	}
	defer rows.Close()
	
	var defaultThemeSwitch bool
	for rows.Next() {
		err := rows.Scan(&uname, &defaultThemeSwitch)
		if err != nil {
			return err
		}
		
		// Was the theme deleted at some point?
		theme, ok := themes[uname]
		if !ok {
			continue
		}
		
		if defaultThemeSwitch {
			log.Print("Loading the theme '" + theme.Name + "'")
			theme.Active = true
			defaultTheme = uname
			add_theme_static_files(uname)
			map_theme_templates(theme)
		} else {
			theme.Active = false
		}
		themes[uname] = theme
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
