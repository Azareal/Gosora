/* Copyright Azareal 2016 - 2017 */
// +build !pgsql !sqlite !mssql
package main

import "log"
import "fmt"
import "strconv"
import "encoding/json"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

var db *sql.DB
var db_version string
var db_collation string = "utf8mb4_general_ci"

var get_topic_replies_offset_stmt *sql.Stmt // I'll need to rewrite this one to stop it hard-coding the per page setting before moving it to the query generator
var get_forum_topics_offset_stmt *sql.Stmt
var notify_watchers_stmt *sql.Stmt
var get_activity_feed_by_watcher_stmt *sql.Stmt
var get_activity_count_by_watcher_stmt *sql.Stmt
var update_email_stmt, verify_email_stmt *sql.Stmt

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
var report_exists_stmt *sql.Stmt

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

	// Build the generated prepared statements, we are going to slowly move the queries over to the query generator rather than writing them all by hand, this'll make it easier for us to implement database adapters for other databases like PostgreSQL, MSSQL, SQlite, etc.
	err = gen_mysql()
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

	log.Print("Preparing report_exists statement.")
	report_exists_stmt, err = db.Prepare("select count(*) as count from topics where data = ? and data != '' and parentID = 1")
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
	forums = append(forums, Forum{0,"Uncategorised","",uncategorised_forum_visible,"all",0,"",0,"",0,""})

	//rows, err = db.Query("SELECT fid, name, active, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime FROM forums")
	rows, err = db.Query("select `fid`, `name`, `desc`, `active`, `preset`, `topicCount`, `lastTopic`, `lastTopicID`, `lastReplyer`, `lastReplyerID`, `lastTopicTime` from forums order by fid asc")
	if err != nil {
		return err
	}
	defer rows.Close()

	i = 1
	for ;rows.Next();i++ {
		forum := Forum{ID:0,Name:"",Active:true,Preset:"all"}
		err := rows.Scan(&forum.ID, &forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			return err
		}

		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if forum.ID != i {
			log.Print("Stop physically deleting forums. You are messing up the IDs. Use the Forum Manager or delete_forum() instead x.x")
			fill_forum_id_gap(i, forum.ID)
		}

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

	var sname, scontent, stype, sconstraints string
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
