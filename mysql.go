/* Copyright Azareal 2016 - 2017 */
package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "log"
import "fmt"
import "encoding/json"

var db *sql.DB
var get_session_stmt *sql.Stmt
var get_topic_list_stmt *sql.Stmt
var get_topic_user_stmt *sql.Stmt
var get_topic_replies_stmt *sql.Stmt
var get_forum_topics_stmt *sql.Stmt
var create_topic_stmt *sql.Stmt
var create_report_stmt *sql.Stmt
var create_reply_stmt *sql.Stmt
var update_forum_cache_stmt *sql.Stmt
var edit_topic_stmt *sql.Stmt
var edit_reply_stmt *sql.Stmt
var delete_reply_stmt *sql.Stmt
var delete_topic_stmt *sql.Stmt
var stick_topic_stmt *sql.Stmt
var unstick_topic_stmt *sql.Stmt
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
var update_setting_stmt *sql.Stmt
var add_plugin_stmt *sql.Stmt
var update_plugin_stmt *sql.Stmt
var update_user_stmt *sql.Stmt
var add_theme_stmt *sql.Stmt
var update_theme_stmt *sql.Stmt

func init_database(err error) {
	if(dbpassword != ""){
		dbpassword = ":" + dbpassword
	}
	db, err = sql.Open("mysql",dbuser + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?collation=utf8mb4_general_ci")
	if err != nil {
		log.Fatal(err)
	}
	
	// Make sure that the connection is alive..
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_session statement.")
	get_session_stmt, err = db.Prepare("select `uid`,`name`,`group`,`is_super_admin`,`session`,`email`,`avatar`,`message`,`url_prefix`,`url_name`,`level`,`score`,`last_ip` from `users` where `uid` = ? and `session` = ? AND `session` <> ''")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_topic_list statement.")
	get_topic_list_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_topic_user statement.")
	get_topic_user_stmt, err = db.Prepare("select topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, users.name, users.avatar, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, topics.ipaddress from topics left join users ON topics.createdBy = users.uid where tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_topic_replies statement.")
	get_topic_replies_stmt, err = db.Prepare("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.is_super_admin, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress from replies left join users ON replies.createdBy = users.uid where tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_forum_topics statement.")
	get_forum_topics_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid WHERE topics.parentID = ? order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_topic statement.")
	create_topic_stmt, err = db.Prepare("insert into topics(title,content,parsed_content,createdAt,ipaddress,createdBy) VALUES(?,?,?,NOW(),?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_report statement.")
	create_report_stmt, err = db.Prepare("INSERT INTO topics(title,content,parsed_content,createdAt,createdBy,data,parentID) VALUES(?,?,?,NOW(),?,?,-1)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_reply statement.")
	create_reply_stmt, err = db.Prepare("INSERT INTO replies(tid,content,parsed_content,createdAt,ipaddress,createdBy) VALUES(?,?,?,NOW(),?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_forum_cache statement.")
	update_forum_cache_stmt, err = db.Prepare("UPDATE forums SET lastTopic = ?, lastTopicID = ?, lastReplyer = ?, lastReplyerID = ?, lastTopicTime = NOW() WHERE fid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing edit_topic statement.")
	edit_topic_stmt, err = db.Prepare("UPDATE topics SET title = ?, content = ?, parsed_content = ?, is_closed = ? WHERE tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing edit_reply statement.")
	edit_reply_stmt, err = db.Prepare("UPDATE replies SET content = ?, parsed_content = ? WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing delete_reply statement.")
	delete_reply_stmt, err = db.Prepare("DELETE FROM replies WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing delete_topic statement.")
	delete_topic_stmt, err = db.Prepare("DELETE FROM topics WHERE tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing stick_topic statement.")
	stick_topic_stmt, err = db.Prepare("UPDATE topics SET sticky = 1 WHERE tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing unstick_topic statement.")
	unstick_topic_stmt, err = db.Prepare("UPDATE topics SET sticky = 0 WHERE tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_last_ip statement.")
	update_last_ip_stmt, err = db.Prepare("UPDATE users SET last_ip = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing login statement.")
	login_stmt, err = db.Prepare("SELECT `uid`, `name`, `password`, `salt` FROM `users` WHERE `name` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_session statement.")
	update_session_stmt, err = db.Prepare("UPDATE users SET session = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing logout statement.")
	logout_stmt, err = db.Prepare("UPDATE users SET session = '' WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing set_password statement.")
	set_password_stmt, err = db.Prepare("UPDATE users SET password = ?, salt = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing get_password statement.")
	get_password_stmt, err = db.Prepare("SELECT `password`, `salt` FROM `users` WHERE `uid` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing set_avatar statement.")
	set_avatar_stmt, err = db.Prepare("UPDATE users SET avatar = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing set_username statement.")
	set_username_stmt, err = db.Prepare("UPDATE users SET name = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	// Add an admin version of register_stmt with more flexibility
	// create_account_stmt, err = db.Prepare("INSERT INTO 
	
	log.Print("Preparing register statement.")
	register_stmt, err = db.Prepare("INSERT INTO users(`name`,`email`,`password`,`salt`,`group`,`is_super_admin`,`session`,`active`,`message`) VALUES(?,?,?,?,?,0,?,?,'')")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing username_exists statement.")
	username_exists_stmt, err = db.Prepare("SELECT `name` FROM `users` WHERE `name` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing change_group statement.")
	change_group_stmt, err = db.Prepare("UPDATE `users` SET `group` = ? WHERE `uid` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing add_email statement.")
	add_email_stmt, err = db.Prepare("INSERT INTO emails(`email`,`uid`,`validated`,`token`) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_email statement.")
	update_email_stmt, err = db.Prepare("UPDATE emails SET email = ?, uid = ?, validated = ?, token = ? WHERE email = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing verify_email statement.")
	verify_email_stmt, err = db.Prepare("UPDATE emails SET validated = 1, token = '' WHERE email = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing activate_user statement.")
	activate_user_stmt, err = db.Prepare("UPDATE users SET active = 1 WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_user_level statement.")
	update_user_level_stmt, err = db.Prepare("UPDATE users SET level = ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing increment_user_score statement.")
	increment_user_score_stmt, err = db.Prepare("UPDATE users SET score = score + ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing increment_user_posts statement.")
	increment_user_posts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing increment_user_bigposts statement.")
	increment_user_bigposts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ?, bigposts =  bigposts + ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing increment_user_megaposts statement.")
	increment_user_megaposts_stmt, err = db.Prepare("UPDATE users SET posts = posts + ?, bigposts = bigposts + ?, megaposts =  megaposts + ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing increment_user_topics statement.")
	increment_user_topics_stmt, err = db.Prepare("UPDATE users SET topics =  topics + ? WHERE uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_profile_reply statement.")
	create_profile_reply_stmt, err = db.Prepare("INSERT INTO users_replies(uid,content,parsed_content,createdAt,createdBy) VALUES(?,?,?,NOW(),?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing edit_profile_reply statement.")
	edit_profile_reply_stmt, err = db.Prepare("UPDATE users_replies SET content = ?, parsed_content = ? WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing delete_profile_reply statement.")
	delete_profile_reply_stmt, err = db.Prepare("DELETE FROM users_replies WHERE rid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing create_forum statement.")
	create_forum_stmt, err = db.Prepare("INSERT INTO forums(name) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing delete_forum statement.")
	delete_forum_stmt, err = db.Prepare("DELETE FROM forums WHERE fid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_forum statement.")
	update_forum_stmt, err = db.Prepare("UPDATE forums SET name = ? WHERE fid = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_setting statement.")
	update_setting_stmt, err = db.Prepare("UPDATE settings SET content = ? WHERE name = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing add_plugin statement.")
	add_plugin_stmt, err = db.Prepare("INSERT INTO plugins(uname,active) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_plugin statement.")
	update_plugin_stmt, err = db.Prepare("UPDATE plugins SET active = ? WHERE uname = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing add_theme statement.")
	add_theme_stmt, err = db.Prepare("INSERT INTO `themes`(`uname`,`default`) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_theme statement.")
	update_theme_stmt, err = db.Prepare("UPDATE `themes` SET `default` = ? WHERE `uname` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Preparing update_user statement.")
	update_user_stmt, err = db.Prepare("UPDATE `users` SET `name` = ?, `email` = ?, `group` = ? WHERE `uid` = ?")
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Loading the usergroups.")
	rows, err := db.Query("SELECT gid,name,permissions,is_mod,is_admin,is_banned,tag FROM users_groups")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	for rows.Next() {
		group := Group{ID: 0,}
		err := rows.Scan(&group.ID, &group.Name, &group.PermissionsText, &group.Is_Mod, &group.Is_Admin, &group.Is_Banned, &group.Tag)
		if err != nil {
			log.Fatal(err)
		}
		
		err = json.Unmarshal(group.PermissionsText, &group.Perms)
		if err != nil {
			log.Fatal(err)
		}
		if !nogrouplog {
			fmt.Println(group.Name + ": ")
			fmt.Printf("%+v\n", group.Perms)
		}
		
		group.Perms.ExtData = make(map[string]bool)
		groups[group.ID] = group
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Loading the forums.")
	rows, err = db.Query("SELECT fid, name, active, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime FROM forums")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	for rows.Next() {
		forum := Forum{0,"",true,"",0,"",0,""}
		err := rows.Scan(&forum.ID, &forum.Name, &forum.Active, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			log.Fatal(err)
		}
		
		if forum.LastTopicID != 0 {
			forum.LastTopicTime, err = relative_time(forum.LastTopicTime)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			forum.LastTopic = "None"
			forum.LastTopicTime = ""
		}
		forums[forum.ID] = forum
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Adding the uncategorised forum")
	forums[0] = Forum{0,"Uncategorised",uncategorised_forum_visible,"",0,"",0,""}
	log.Print("Adding the reports forum")
	forums[-1] = Forum{-1,"Reports",false,"",0,"",0,""}
	
	log.Print("Loading the settings.")
	rows, err = db.Query("SELECT name, content, type, constraints FROM settings")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var sname string
	var scontent string
	var stype string
	var sconstraints string
	for rows.Next() {
		err := rows.Scan(&sname, &scontent, &stype, &sconstraints)
		if err != nil {
			log.Fatal(err)
		}
		errmsg := parseSetting(sname, scontent, stype, sconstraints)
		if errmsg != "" {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Loading the plugins.")
	rows, err = db.Query("SELECT uname, active FROM plugins")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var uname string
	var active bool
	for rows.Next() {
		err := rows.Scan(&uname, &active)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}
	
	log.Print("Loading the themes.")
	rows, err = db.Query("SELECT `uname`, `default` FROM `themes`")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	
	var defaultThemeSwitch bool
	for rows.Next() {
		err := rows.Scan(&uname, &defaultThemeSwitch)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}
	
	
}
