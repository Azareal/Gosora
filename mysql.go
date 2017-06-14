/* Copyright Azareal 2016 - 2017 */
// +build !pgsql !sqlite !mssql
package main

import "log"
import "strconv"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "./query_gen/lib"

var db *sql.DB
var db_version string
var db_collation string = "utf8mb4_general_ci"

var get_topic_replies_offset_stmt *sql.Stmt // I'll need to rewrite this one to stop it hard-coding the per page setting before moving it to the query generator
var get_forum_topics_offset_stmt *sql.Stmt
var notify_watchers_stmt *sql.Stmt
var get_activity_feed_by_watcher_stmt *sql.Stmt
var get_activity_count_by_watcher_stmt *sql.Stmt

var forum_entry_exists_stmt *sql.Stmt
var group_entry_exists_stmt *sql.Stmt
var add_forum_perms_to_forum_admins_stmt *sql.Stmt
var add_forum_perms_to_forum_staff_stmt *sql.Stmt
var add_forum_perms_to_forum_members_stmt *sql.Stmt
var update_forum_perms_for_group_stmt *sql.Stmt
var todays_post_count_stmt *sql.Stmt
var todays_topic_count_stmt *sql.Stmt
var todays_report_count_stmt *sql.Stmt
var todays_newuser_count_stmt *sql.Stmt

func _init_database() (err error) {
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

	// Build the generated prepared statements, we are going to slowly move the queries over to the query generator rather than writing them all by hand, this'll make it easier for us to implement database adapters for other databases like PostgreSQL, MSSQL, SQlite, etc.
	err = gen_mysql()
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("mysql")
	if err != nil {
		return err
	}

	log.Print("Preparing get_topic_replies_offset statement.")
	get_topic_replies_offset_stmt, err = db.Prepare("select replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.actionType from replies left join users on replies.createdBy = users.uid where tid = ? limit ?, " + strconv.Itoa(items_per_page))
	if err != nil {
		return err
	}

	log.Print("Preparing get_forum_topics_offset statement.")
	get_forum_topics_offset_stmt, err = db.Prepare("select topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, topics.postCount, topics.likeCount, users.name, users.avatar from topics left join users ON topics.createdBy = users.uid WHERE topics.parentID = ? order by topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC limit ?, " + strconv.Itoa(items_per_page))
	if err != nil {
		return err
	}

	log.Print("Preparing notify_watchers statement.")
	notify_watchers_stmt, err = db.Prepare("INSERT INTO activity_stream_matches(watcher, asid) SELECT activity_subscriptions.user, activity_stream.asid FROM activity_stream INNER JOIN activity_subscriptions ON activity_subscriptions.targetType = activity_stream.elementType and activity_subscriptions.targetID = activity_stream.elementID and activity_subscriptions.user != activity_stream.actor where asid = ?")
	if err != nil {
		return err
	}

	log.Print("Preparing get_activity_feed_by_watcher statement.")
	get_activity_feed_by_watcher_stmt, err = db.Prepare("SELECT activity_stream_matches.asid, activity_stream.actor, activity_stream.targetUser, activity_stream.event, activity_stream.elementType, activity_stream.elementID FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ? ORDER BY activity_stream.asid ASC LIMIT 8")
	if err != nil {
		return err
	}

	log.Print("Preparing get_activity_count_by_watcher statement.")
	get_activity_count_by_watcher_stmt, err = db.Prepare("SELECT count(*) FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ?")
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

	log.Print("Preparing todays_post_count statement.")
	todays_post_count_stmt, err = db.Prepare("select count(*) from replies where createdAt BETWEEN (now() - interval 1 day) and now()")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_topic_count statement.")
	todays_topic_count_stmt, err = db.Prepare("select count(*) from topics where createdAt BETWEEN (now() - interval 1 day) and now()")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_report_count statement.")
	todays_report_count_stmt, err = db.Prepare("select count(*) from topics where createdAt BETWEEN (now() - interval 1 day) and now() and parentID = 1")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_newuser_count statement.")
	todays_newuser_count_stmt, err = db.Prepare("select count(*) from users where createdAt BETWEEN (now() - interval 1 day) and now()")
	if err != nil {
		return err
	}

	return nil
}
