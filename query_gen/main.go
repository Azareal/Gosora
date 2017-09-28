/* WIP Under Construction */
package main

import "log"
import "./lib"

func main() {
	log.Println("Running the query generator")
	for _, adapter := range qgen.DB_Registry {
		log.Println("Building the queries for the " + adapter.GetName() + " adapter")
		qgen.Install.SetAdapterInstance(adapter)
		write_statements(adapter)
		qgen.Install.Write()
		adapter.Write()
	}
}

// nolint
func write_statements(adapter qgen.DB_Adapter) error {
	err := create_tables(adapter)
	if err != nil {
		return err
	}

	err = seed_tables(adapter)
	if err != nil {
		return err
	}

	err = write_selects(adapter)
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

	err = write_simple_counts(adapter)
	if err != nil {
		return err
	}

	err = write_insert_selects(adapter)
	if err != nil {
		return err
	}

	err = write_insert_left_joins(adapter)
	if err != nil {
		return err
	}

	err = write_insert_inner_joins(adapter)
	if err != nil {
		return err
	}

	return nil
}

// nolint
func create_tables(adapter qgen.DB_Adapter) error {
	qgen.Install.CreateTable("users", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"password", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"salt", "varchar", 80, false, false, "''"},
			qgen.DB_Table_Column{"group", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"is_super_admin", "boolean", 0, false, false, "0"},
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"lastActiveAt", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"session", "varchar", 200, false, false, "''"},
			qgen.DB_Table_Column{"last_ip", "varchar", 200, false, false, "0.0.0.0.0"},
			qgen.DB_Table_Column{"email", "varchar", 200, false, false, "''"},
			qgen.DB_Table_Column{"avatar", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"message", "text", 0, false, false, "''"},
			qgen.DB_Table_Column{"url_prefix", "varchar", 20, false, false, "''"},
			qgen.DB_Table_Column{"url_name", "varchar", 100, false, false, "''"},
			qgen.DB_Table_Column{"level", "smallint", 0, false, false, "0"},
			qgen.DB_Table_Column{"score", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"posts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"bigposts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"megaposts", "int", 0, false, false, "0"},
			qgen.DB_Table_Column{"topics", "int", 0, false, false, "0"},
			//qgen.DB_Table_Column{"penalty_count","int",0,false,false,"0"},
			qgen.DB_Table_Column{"temp_group", "int", 0, false, false, "0"}, // For temporary groups, set this to zero when a temporary group isn't in effect
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uid", "primary"},
			qgen.DB_Table_Key{"name", "unique"},
		},
	)

	// What should we do about global penalties? Put them on the users table for speed? Or keep them here?
	// Should we add IP Penalties? No, that's a stupid idea, just implement IP Bans properly. What about shadowbans?
	// TODO: Perm overrides
	// TODO: Add a mod-queue and other basic auto-mod features. This is needed for awaiting activation and the mod_queue penalty flag
	// TODO: Add a penalty type where a user is stopped from creating plugin_socialgroups social groups
	// TODO: Shadow bans. We will probably have a CanShadowBan permission for this, as we *really* don't want people using this lightly.
	/*qgen.Install.CreateTable("users_penalties","","",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid","int",0,false,false,""},
			qgen.DB_Table_Column{"element_id","int",0,false,false,""},
			qgen.DB_Table_Column{"element_type","varchar",50,false,false,""}, //forum, profile?, and social_group. Leave blank for global.
			qgen.DB_Table_Column{"overrides","text",0,false,false,"{}"},

			qgen.DB_Table_Column{"mod_queue","boolean",0,false,false,"0"},
			qgen.DB_Table_Column{"shadow_ban","boolean",0,false,false,"0"},
			qgen.DB_Table_Column{"no_avatar","boolean",0,false,false,"0"}, // Coming Soon. Should this be a perm override instead?

			// Do we *really* need rate-limit penalty types? Are we going to be allowing bots or something?
			//qgen.DB_Table_Column{"posts_per_hour","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"topics_per_hour","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"posts_count","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"topic_count","int",0,false,false,"0"},
			//qgen.DB_Table_Column{"last_hour","int",0,false,false,"0"}, // UNIX Time, as we don't need to do anything too fancy here. When an hour has elapsed since that time, reset the hourly penalty counters.

			qgen.DB_Table_Column{"issued_by","int",0,false,false,""},
			qgen.DB_Table_Column{"issued_at","createdAt",0,false,false,""},
			qgen.DB_Table_Column{"expires_at","datetime",0,false,false,""},
		},
		[]qgen.DB_Table_Key{},
	)*/

	qgen.Install.CreateTable("users_groups_scheduler", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"set_group", "int", 0, false, false, ""},

			qgen.DB_Table_Column{"issued_by", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"issued_at", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"revert_at", "datetime", 0, false, false, ""},
			qgen.DB_Table_Column{"temporary", "boolean", 0, false, false, ""}, // special case for permanent bans to do the necessary bookkeeping, might be removed in the future
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"uid", "primary"},
		},
	)

	qgen.Install.CreateTable("word_filters", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"wfid", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"find", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"replacement", "varchar", 200, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"wfid", "primary"},
		},
	)

	qgen.Install.CreateTable("sync", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"last_update", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)

	return nil
}

// nolint
func seed_tables(adapter qgen.DB_Adapter) error {
	return nil
}

// nolint
func write_selects(adapter qgen.DB_Adapter) error {
	// url_prefix and url_name will be removed from this query in a later commit
	adapter.SimpleSelect("getUser", "users", "name, group, is_super_admin, avatar, message, url_prefix, url_name, level", "uid = ?", "", "")

	// Looking for getTopic? Your statement is in another castle

	adapter.SimpleSelect("getReply", "replies", "tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount", "rid = ?", "", "")

	adapter.SimpleSelect("getUserReply", "users_replies", "uid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress", "rid = ?", "", "")

	adapter.SimpleSelect("getPassword", "users", "password,salt", "uid = ?", "", "")

	adapter.SimpleSelect("getSettings", "settings", "name, content, type", "", "", "")

	adapter.SimpleSelect("getSetting", "settings", "content, type", "name = ?", "", "")

	adapter.SimpleSelect("getFullSetting", "settings", "name, type, constraints", "name = ?", "", "")

	adapter.SimpleSelect("getFullSettings", "settings", "name, content, type, constraints", "", "", "")

	adapter.SimpleSelect("getGroups", "users_groups", "gid, name, permissions, plugin_perms, is_mod, is_admin, is_banned, tag", "", "", "")

	adapter.SimpleSelect("getForums", "forums", "fid, name, desc, active, preset, parentID, parentType, topicCount, lastTopicID, lastReplyerID", "", "fid ASC", "")

	adapter.SimpleSelect("getForumsPermissions", "forums_permissions", "gid, fid, permissions", "", "gid ASC, fid ASC", "")

	adapter.SimpleSelect("getPlugins", "plugins", "uname, active, installed", "", "", "")

	adapter.SimpleSelect("getThemes", "themes", "uname, default", "", "", "")

	adapter.SimpleSelect("getWidgets", "widgets", "position, side, type, active,  location, data", "", "position ASC", "")

	adapter.SimpleSelect("isPluginActive", "plugins", "active", "uname = ?", "", "")

	//adapter.SimpleSelect("isPluginInstalled","plugins","installed","uname = ?","","")

	adapter.SimpleSelect("getUsers", "users", "uid, name, group, active, is_super_admin, avatar", "", "", "")

	adapter.SimpleSelect("getUsersOffset", "users", "uid, name, group, active, is_super_admin, avatar", "", "", "?,?")

	adapter.SimpleSelect("getWordFilters", "word_filters", "wfid, find, replacement", "", "", "")

	adapter.SimpleSelect("isThemeDefault", "themes", "default", "uname = ?", "", "")

	adapter.SimpleSelect("getModlogs", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "", "", "")

	adapter.SimpleSelect("getModlogsOffset", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "", "", "?,?")

	adapter.SimpleSelect("getReplyTID", "replies", "tid", "rid = ?", "", "")

	adapter.SimpleSelect("getTopicFID", "topics", "parentID", "tid = ?", "", "")

	adapter.SimpleSelect("getUserReplyUID", "users_replies", "uid", "rid = ?", "", "")

	adapter.SimpleSelect("hasLikedTopic", "likes", "targetItem", "sentBy = ? and targetItem = ? and targetType = 'topics'", "", "")

	adapter.SimpleSelect("hasLikedReply", "likes", "targetItem", "sentBy = ? and targetItem = ? and targetType = 'replies'", "", "")

	adapter.SimpleSelect("getUserName", "users", "name", "uid = ?", "", "")

	adapter.SimpleSelect("getEmailsByUser", "emails", "email, validated, token", "uid = ?", "", "")

	adapter.SimpleSelect("getTopicBasic", "topics", "title, content", "tid = ?", "", "")

	adapter.SimpleSelect("getActivityEntry", "activity_stream", "actor, targetUser, event, elementType, elementID", "asid = ?", "", "")

	adapter.SimpleSelect("forumEntryExists", "forums", "fid", "name = ''", "fid ASC", "0,1")

	adapter.SimpleSelect("groupEntryExists", "users_groups", "gid", "name = ''", "gid ASC", "0,1")

	adapter.SimpleSelect("getForumTopicsOffset", "topics", "tid, title, content, createdBy, is_closed, sticky, createdAt, lastReplyAt, lastReplyBy, parentID, postCount, likeCount", "parentID = ?", "sticky DESC, lastReplyAt DESC, createdBy DESC", "?,?")

	adapter.SimpleSelect("getExpiredScheduledGroups", "users_groups_scheduler", "uid", "UTC_TIMESTAMP() > revert_at AND temporary = 1", "", "")

	adapter.SimpleSelect("getSync", "sync", "last_update", "", "", "")

	return nil
}

// nolint
func write_left_joins(adapter qgen.DB_Adapter) error {
	adapter.SimpleLeftJoin("getTopicRepliesOffset", "replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.actionType", "replies.createdBy = users.uid", "tid = ?", "", "?,?")

	adapter.SimpleLeftJoin("getTopicList", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC", "")

	adapter.SimpleLeftJoin("getTopicUser", "topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level", "topics.createdBy = users.uid", "tid = ?", "", "")

	adapter.SimpleLeftJoin("getTopicByReply", "replies", "topics", "topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, topics.data", "replies.tid = topics.tid", "rid = ?", "", "")

	adapter.SimpleLeftJoin("getTopicReplies", "replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress", "replies.createdBy = users.uid", "tid = ?", "", "")

	adapter.SimpleLeftJoin("getForumTopics", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "topics.parentID = ?", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc", "")

	adapter.SimpleLeftJoin("getProfileReplies", "users_replies", "users", "users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group", "users_replies.createdBy = users.uid", "users_replies.uid = ?", "", "")

	return nil
}

// nolint
func write_inner_joins(adapter qgen.DB_Adapter) error {
	adapter.SimpleInnerJoin("getWatchers", "activity_stream", "activity_subscriptions", "activity_subscriptions.user", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", "")

	return nil
}

// nolint
func write_inserts(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsert("createTopic", "topics", "parentID,title,content,parsed_content,createdAt,lastReplyAt,lastReplyBy,ipaddress,words,createdBy", "?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?")

	adapter.SimpleInsert("createReport", "topics", "title,content,parsed_content,createdAt,lastReplyAt,createdBy,data,parentID,css_class", "?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,1,'report'")

	adapter.SimpleInsert("createReply", "replies", "tid,content,parsed_content,createdAt,ipaddress,words,createdBy", "?,?,?,UTC_TIMESTAMP(),?,?,?")

	adapter.SimpleInsert("createActionReply", "replies", "tid,actionType,ipaddress,createdBy", "?,?,?,?")

	adapter.SimpleInsert("createLike", "likes", "weight, targetItem, targetType, sentBy", "?,?,?,?")

	adapter.SimpleInsert("addActivity", "activity_stream", "actor,targetUser,event,elementType,elementID", "?,?,?,?,?")

	adapter.SimpleInsert("notifyOne", "activity_stream_matches", "watcher,asid", "?,?")

	adapter.SimpleInsert("addEmail", "emails", "email, uid, validated, token", "?,?,?,?")

	adapter.SimpleInsert("createProfileReply", "users_replies", "uid, content, parsed_content, createdAt, createdBy, ipaddress", "?,?,?,UTC_TIMESTAMP(),?,?")

	adapter.SimpleInsert("addSubscription", "activity_subscriptions", "user,targetID,targetType,level", "?,?,?,2")

	adapter.SimpleInsert("createForum", "forums", "name, desc, active, preset", "?,?,?,?")

	adapter.SimpleInsert("addForumPermsToForum", "forums_permissions", "gid,fid,preset,permissions", "?,?,?,?")

	adapter.SimpleInsert("addPlugin", "plugins", "uname, active, installed", "?,?,?")

	adapter.SimpleInsert("addTheme", "themes", "uname,default", "?,?")

	adapter.SimpleInsert("createGroup", "users_groups", "name, tag, is_admin, is_mod, is_banned, permissions", "?,?,?,?,?,?")

	adapter.SimpleInsert("addModlogEntry", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "?,?,?,?,?,UTC_TIMESTAMP()")

	adapter.SimpleInsert("addAdminlogEntry", "administration_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "?,?,?,?,?,UTC_TIMESTAMP()")

	adapter.SimpleInsert("createWordFilter", "word_filters", "find, replacement", "?,?")

	return nil
}

// nolint
func write_replaces(adapter qgen.DB_Adapter) error {
	adapter.SimpleReplace("addForumPermsToGroup", "forums_permissions", "gid,fid,preset,permissions", "?,?,?,?")

	adapter.SimpleReplace("replaceScheduleGroup", "users_groups_scheduler", "uid, set_group, issued_by, issued_at, revert_at, temporary", "?,?,?,UTC_TIMESTAMP(),?,?")

	return nil
}

// nolint
func write_updates(adapter qgen.DB_Adapter) error {
	adapter.SimpleUpdate("addRepliesToTopic", "topics", "postCount = postCount + ?, lastReplyBy = ?, lastReplyAt = UTC_TIMESTAMP()", "tid = ?")

	adapter.SimpleUpdate("removeRepliesFromTopic", "topics", "postCount = postCount - ?", "tid = ?")

	adapter.SimpleUpdate("addTopicsToForum", "forums", "topicCount = topicCount + ?", "fid = ?")

	adapter.SimpleUpdate("removeTopicsFromForum", "forums", "topicCount = topicCount - ?", "fid = ?")

	adapter.SimpleUpdate("updateForumCache", "forums", "lastTopicID = ?, lastReplyerID = ?", "fid = ?")

	adapter.SimpleUpdate("addLikesToTopic", "topics", "likeCount = likeCount + ?", "tid = ?")

	adapter.SimpleUpdate("addLikesToReply", "replies", "likeCount = likeCount + ?", "rid = ?")

	adapter.SimpleUpdate("editTopic", "topics", "title = ?, content = ?, parsed_content = ?", "tid = ?")

	adapter.SimpleUpdate("editReply", "replies", "content = ?, parsed_content = ?", "rid = ?")

	adapter.SimpleUpdate("stickTopic", "topics", "sticky = 1", "tid = ?")

	adapter.SimpleUpdate("unstickTopic", "topics", "sticky = 0", "tid = ?")

	adapter.SimpleUpdate("lockTopic", "topics", "is_closed = 1", "tid = ?")

	adapter.SimpleUpdate("unlockTopic", "topics", "is_closed = 0", "tid = ?")

	adapter.SimpleUpdate("updateLastIP", "users", "last_ip = ?", "uid = ?")

	adapter.SimpleUpdate("updateSession", "users", "session = ?", "uid = ?")

	adapter.SimpleUpdate("setPassword", "users", "password = ?, salt = ?", "uid = ?")

	adapter.SimpleUpdate("setAvatar", "users", "avatar = ?", "uid = ?")

	adapter.SimpleUpdate("setUsername", "users", "name = ?", "uid = ?")

	adapter.SimpleUpdate("changeGroup", "users", "group = ?", "uid = ?")

	adapter.SimpleUpdate("activateUser", "users", "active = 1", "uid = ?")

	adapter.SimpleUpdate("updateUserLevel", "users", "level = ?", "uid = ?")

	adapter.SimpleUpdate("incrementUserScore", "users", "score = score + ?", "uid = ?")

	adapter.SimpleUpdate("incrementUserPosts", "users", "posts = posts + ?", "uid = ?")

	adapter.SimpleUpdate("incrementUserBigposts", "users", "posts = posts + ?, bigposts = bigposts + ?", "uid = ?")

	adapter.SimpleUpdate("incrementUserMegaposts", "users", "posts = posts + ?, bigposts = bigposts + ?, megaposts = megaposts + ?", "uid = ?")

	adapter.SimpleUpdate("incrementUserTopics", "users", "topics =  topics + ?", "uid = ?")

	adapter.SimpleUpdate("editProfileReply", "users_replies", "content = ?, parsed_content = ?", "rid = ?")

	adapter.SimpleUpdate("updateForum", "forums", "name = ?, desc = ?, active = ?, preset = ?", "fid = ?")

	adapter.SimpleUpdate("updateSetting", "settings", "content = ?", "name = ?")

	adapter.SimpleUpdate("updatePlugin", "plugins", "active = ?", "uname = ?")

	adapter.SimpleUpdate("updatePluginInstall", "plugins", "installed = ?", "uname = ?")

	adapter.SimpleUpdate("updateTheme", "themes", "default = ?", "uname = ?")

	adapter.SimpleUpdate("updateUser", "users", "name = ?, email = ?, group = ?", "uid = ?")

	adapter.SimpleUpdate("updateGroupPerms", "users_groups", "permissions = ?", "gid = ?")

	adapter.SimpleUpdate("updateGroupRank", "users_groups", "is_admin = ?, is_mod = ?, is_banned = ?", "gid = ?")

	adapter.SimpleUpdate("updateGroup", "users_groups", "name = ?, tag = ?", "gid = ?")

	adapter.SimpleUpdate("updateEmail", "emails", "email = ?, uid = ?, validated = ?, token = ?", "email = ?")

	adapter.SimpleUpdate("verifyEmail", "emails", "validated = 1, token = ''", "email = ?") // Need to fix this: Empty string isn't working, it gets set to 1 instead x.x -- Has this been fixed?

	adapter.SimpleUpdate("setTempGroup", "users", "temp_group = ?", "uid = ?")

	adapter.SimpleUpdate("updateWordFilter", "word_filters", "find = ?, replacement = ?", "wfid = ?")

	adapter.SimpleUpdate("bumpSync", "sync", "last_update = UTC_TIMESTAMP()", "")

	return nil
}

// nolint
func write_deletes(adapter qgen.DB_Adapter) error {
	adapter.SimpleDelete("deleteReply", "replies", "rid = ?")

	adapter.SimpleDelete("deleteProfileReply", "users_replies", "rid = ?")

	adapter.SimpleDelete("deleteForumPermsByForum", "forums_permissions", "fid = ?")

	adapter.SimpleDelete("deleteActivityStreamMatch", "activity_stream_matches", "watcher = ? AND asid = ?")
	//adapter.SimpleDelete("delete_activity_stream_matches_by_watcher","activity_stream_matches","watcher = ?")

	adapter.SimpleDelete("deleteWordFilter", "word_filters", "wfid = ?")

	return nil
}

// nolint
func write_simple_counts(adapter qgen.DB_Adapter) error {
	adapter.SimpleCount("reportExists", "topics", "data = ? AND data != '' AND parentID = 1", "")

	adapter.SimpleCount("groupCount", "users_groups", "", "")

	adapter.SimpleCount("modlogCount", "moderation_logs", "", "")

	return nil
}

// nolint
func write_insert_selects(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsertSelect("addForumPermsToForumAdmins",
		qgen.DB_Insert{"forums_permissions", "gid,fid,preset,permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 1", "", ""},
	)

	adapter.SimpleInsertSelect("addForumPermsToForumStaff",
		qgen.DB_Insert{"forums_permissions", "gid,fid,preset,permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 1", "", ""},
	)

	adapter.SimpleInsertSelect("addForumPermsToForumMembers",
		qgen.DB_Insert{"forums_permissions", "gid,fid,preset,permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 0 AND is_banned = 0", "", ""},
	)

	return nil
}

// nolint
func write_insert_left_joins(adapter qgen.DB_Adapter) error {
	return nil
}

// nolint
func write_insert_inner_joins(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsertInnerJoin("notifyWatchers",
		qgen.DB_Insert{"activity_stream_matches", "watcher, asid", ""},
		qgen.DB_Join{"activity_stream", "activity_subscriptions", "activity_subscriptions.user, activity_stream.asid", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""},
	)

	return nil
}
