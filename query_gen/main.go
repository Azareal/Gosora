/* WIP Under Construction */
package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"./lib"
)

// TODO: Make sure all the errors in this file propagate upwards properly
func main() {
	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			return
		}
	}()

	log.Println("Running the query generator")
	for _, adapter := range qgen.DB_Registry {
		log.Println("Building the queries for the " + adapter.GetName() + " adapter")
		qgen.Install.SetAdapterInstance(adapter)
		qgen.Install.RegisterPlugin(NewPrimaryKeySpitter()) // TODO: Do we really need to fill the spitter for every adapter?

		err := writeStatements(adapter)
		if err != nil {
			log.Print(err)
		}
		err = qgen.Install.Write()
		if err != nil {
			log.Print(err)
		}
		err = adapter.Write()
		if err != nil {
			log.Print(err)
		}
	}
}

// nolint
func writeStatements(adapter qgen.DB_Adapter) error {
	err := createTables(adapter)
	if err != nil {
		return err
	}

	err = seedTables(adapter)
	if err != nil {
		return err
	}

	err = writeSelects(adapter)
	if err != nil {
		return err
	}

	err = writeLeftJoins(adapter)
	if err != nil {
		return err
	}

	err = writeInnerJoins(adapter)
	if err != nil {
		return err
	}

	err = writeInserts(adapter)
	if err != nil {
		return err
	}

	/*err = writeReplaces(adapter)
	if err != nil {
		return err
	}

	err = writeUpserts(adapter)
	if err != nil {
		return err
	}*/

	err = writeUpdates(adapter)
	if err != nil {
		return err
	}

	err = writeDeletes(adapter)
	if err != nil {
		return err
	}

	err = writeSimpleCounts(adapter)
	if err != nil {
		return err
	}

	err = writeInsertSelects(adapter)
	if err != nil {
		return err
	}

	err = writeInsertLeftJoins(adapter)
	if err != nil {
		return err
	}

	err = writeInsertInnerJoins(adapter)
	if err != nil {
		return err
	}

	return nil
}

func seedTables(adapter qgen.DB_Adapter) error {
	qgen.Install.SimpleInsert("sync", "last_update", "UTC_TIMESTAMP()")

	qgen.Install.SimpleInsert("settings", "name, content, type", "'url_tags','1','bool'")
	qgen.Install.SimpleInsert("settings", "name, content, type, constraints", "'activation_type','1','list','1-3'")
	qgen.Install.SimpleInsert("settings", "name, content, type", "'bigpost_min_words','250','int'")
	qgen.Install.SimpleInsert("settings", "name, content, type", "'megapost_min_words','1000','int'")
	qgen.Install.SimpleInsert("themes", "uname, default", "'tempra-simple',1")
	qgen.Install.SimpleInsert("emails", "email, uid, validated", "'admin@localhost',1,1") // ? - Use a different default email or let the admin input it during installation?

	/*
		The Permissions:

		Global Permissions:
		BanUsers
		ActivateUsers
		EditUser
		EditUserEmail
		EditUserPassword
		EditUserGroup
		EditUserGroupSuperMod
		EditUserGroupAdmin
		EditGroup
		EditGroupLocalPerms
		EditGroupGlobalPerms
		EditGroupSuperMod
		EditGroupAdmin
		ManageForums
		EditSettings
		ManageThemes
		ManagePlugins
		ViewAdminLogs
		ViewIPs

		Non-staff Global Permissions:
		UploadFiles

		Forum Permissions:
		ViewTopic
		LikeItem
		CreateTopic
		EditTopic
		DeleteTopic
		CreateReply
		EditReply
		DeleteReply
		PinTopic
		CloseTopic
	*/

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, is_mod, is_admin, tag", `'Administrator','{"BanUsers":true,"ActivateUsers":true,"EditUser":true,"EditUserEmail":true,"EditUserPassword":true,"EditUserGroup":true,"EditUserGroupSuperMod":true,"EditUserGroupAdmin":false,"EditGroup":true,"EditGroupLocalPerms":true,"EditGroupGlobalPerms":true,"EditGroupSuperMod":true,"EditGroupAdmin":false,"ManageForums":true,"EditSettings":true,"ManageThemes":true,"ManagePlugins":true,"ViewAdminLogs":true,"ViewIPs":true,"UploadFiles":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}','{}',1,1,"Admin"`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, is_mod, tag", `'Moderator','{"BanUsers":true,"ActivateUsers":false,"EditUser":true,"EditUserEmail":false,"EditUserGroup":true,"ViewIPs":true,"UploadFiles":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}','{}',1,"Mod"`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms", `'Member','{"UploadFiles":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"CreateReply":true}','{}'`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, is_banned", `'Banned','{"ViewTopic":true}','{}',1`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms", `'Awaiting Activation','{"ViewTopic":true}','{}'`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, tag", `'Not Loggedin','{"ViewTopic":true}','{}','Guest'`)

	//
	// TODO: Stop processFields() from stripping the spaces in the descriptions in the next commit

	qgen.Install.SimpleInsert("forums", "name, active, desc", "'Reports',0,'All the reports go here'")

	qgen.Install.SimpleInsert("forums", "name, lastTopicID, lastReplyerID, desc", "'General',1,1,'A place for general discussions which don't fit elsewhere'")

	//

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `1,1,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"PinTopic":true,"CloseTopic":true}'`)
	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `2,1,'{"ViewTopic":true,"CreateReply":true,"CloseTopic":true}'`)
	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", "3,1,'{}'")
	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", "4,1,'{}'")
	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", "5,1,'{}'")
	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", "6,1,'{}'")

	//

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `1,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `2,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `3,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `4,2,'{"ViewTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `5,2,'{"ViewTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `6,2,'{"ViewTopic":true}'`)

	//

	qgen.Install.SimpleInsert("topics", "title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, createdBy, parentID", "'Test Topic','A topic automatically generated by the software.','A topic automatically generated by the software.',UTC_TIMESTAMP(),UTC_TIMESTAMP(),1,1,2")

	qgen.Install.SimpleInsert("replies", "tid, content, parsed_content, createdAt, createdBy, lastUpdated, lastEdit, lastEditBy", "1,'A reply!','A reply!',UTC_TIMESTAMP(),1,UTC_TIMESTAMP(),0,0")

	return nil
}

func writeSelects(adapter qgen.DB_Adapter) error {
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

	adapter.SimpleSelect("getUsersOffset", "users", "uid, name, group, active, is_super_admin, avatar", "", "uid ASC", "?,?")

	adapter.SimpleSelect("getWordFilters", "word_filters", "wfid, find, replacement", "", "", "")

	adapter.SimpleSelect("isThemeDefault", "themes", "default", "uname = ?", "", "")

	adapter.SimpleSelect("getModlogs", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "", "", "")

	adapter.SimpleSelect("getModlogsOffset", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "", "doneAt DESC", "?,?")

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

	adapter.SimpleSelect("getAttachment", "attachments", "sectionID, sectionTable, originID, originTable, uploadedBy, path", "path = ? AND sectionID = ? AND sectionTable = ?", "", "")

	return nil
}

func writeLeftJoins(adapter qgen.DB_Adapter) error {
	adapter.SimpleLeftJoin("getTopicRepliesOffset", "replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.actionType", "replies.createdBy = users.uid", "replies.tid = ?", "replies.rid ASC", "?,?")

	adapter.SimpleLeftJoin("getTopicList", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy DESC", "")

	adapter.SimpleLeftJoin("getTopicUser", "topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level", "topics.createdBy = users.uid", "tid = ?", "", "")

	adapter.SimpleLeftJoin("getTopicByReply", "replies", "topics", "topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, topics.data", "replies.tid = topics.tid", "rid = ?", "", "")

	adapter.SimpleLeftJoin("getTopicReplies", "replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress", "replies.createdBy = users.uid", "tid = ?", "", "")

	adapter.SimpleLeftJoin("getForumTopics", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "topics.parentID = ?", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc", "")

	adapter.SimpleLeftJoin("getProfileReplies", "users_replies", "users", "users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group", "users_replies.createdBy = users.uid", "users_replies.uid = ?", "", "")

	return nil
}

func writeInnerJoins(adapter qgen.DB_Adapter) (err error) {
	_, err = adapter.SimpleInnerJoin("getWatchers", "activity_stream", "activity_subscriptions", "activity_subscriptions.user", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", "")
	if err != nil {
		return err
	}

	return nil
}

func writeInserts(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsert("createTopic", "topics", "parentID, title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, ipaddress, words, createdBy", "?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?")

	adapter.SimpleInsert("createReport", "topics", "title, content, parsed_content, createdAt, lastReplyAt, createdBy, lastReplyBy, data, parentID, css_class", "?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,1,'report'")

	adapter.SimpleInsert("createReply", "replies", "tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy", "?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?")

	adapter.SimpleInsert("createActionReply", "replies", "tid, actionType, ipaddress, createdBy, createdAt, lastUpdated, content, parsed_content", "?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),'',''")

	adapter.SimpleInsert("createLike", "likes", "weight, targetItem, targetType, sentBy", "?,?,?,?")

	adapter.SimpleInsert("addActivity", "activity_stream", "actor, targetUser, event, elementType, elementID", "?,?,?,?,?")

	adapter.SimpleInsert("notifyOne", "activity_stream_matches", "watcher, asid", "?,?")

	adapter.SimpleInsert("addEmail", "emails", "email, uid, validated, token", "?,?,?,?")

	adapter.SimpleInsert("createProfileReply", "users_replies", "uid, content, parsed_content, createdAt, createdBy, ipaddress", "?,?,?,UTC_TIMESTAMP(),?,?")

	adapter.SimpleInsert("addSubscription", "activity_subscriptions", "user, targetID, targetType, level", "?,?,?,2")

	adapter.SimpleInsert("createForum", "forums", "name, desc, active, preset", "?,?,?,?")

	adapter.SimpleInsert("addForumPermsToForum", "forums_permissions", "gid,fid,preset,permissions", "?,?,?,?")

	adapter.SimpleInsert("addPlugin", "plugins", "uname, active, installed", "?,?,?")

	adapter.SimpleInsert("addTheme", "themes", "uname, default", "?,?")

	adapter.SimpleInsert("addModlogEntry", "moderation_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "?,?,?,?,?,UTC_TIMESTAMP()")

	adapter.SimpleInsert("addAdminlogEntry", "administration_logs", "action, elementID, elementType, ipaddress, actorID, doneAt", "?,?,?,?,?,UTC_TIMESTAMP()")

	adapter.SimpleInsert("addAttachment", "attachments", "sectionID, sectionTable, originID, originTable, uploadedBy, path", "?,?,?,?,?,?")

	adapter.SimpleInsert("createWordFilter", "word_filters", "find, replacement", "?,?")

	return nil
}

func writeReplaces(adapter qgen.DB_Adapter) (err error) {
	return nil
}

// ! Upserts are broken atm
/*func writeUpserts(adapter qgen.DB_Adapter) (err error) {
	_, err = adapter.SimpleUpsert("addForumPermsToGroup", "forums_permissions", "gid, fid, preset, permissions", "?,?,?,?", "gid = ? AND fid = ?")
	if err != nil {
		return err
	}

	_, err = adapter.SimpleUpsert("replaceScheduleGroup", "users_groups_scheduler", "uid, set_group, issued_by, issued_at, revert_at, temporary", "?,?,?,UTC_TIMESTAMP(),?,?", "uid = ?")
	if err != nil {
		return err
	}

	return nil
}*/

func writeUpdates(adapter qgen.DB_Adapter) error {
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

	adapter.SimpleUpdate("updateUserGroup", "users", "group = ?", "uid = ?")

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

func writeDeletes(adapter qgen.DB_Adapter) error {
	adapter.SimpleDelete("deleteUser", "users", "uid = ?")

	adapter.SimpleDelete("deleteTopic", "topics", "tid = ?")

	adapter.SimpleDelete("deleteReply", "replies", "rid = ?")

	adapter.SimpleDelete("deleteProfileReply", "users_replies", "rid = ?")

	//adapter.SimpleDelete("deleteForumPermsByForum", "forums_permissions", "fid = ?")

	adapter.SimpleDelete("deleteActivityStreamMatch", "activity_stream_matches", "watcher = ? AND asid = ?")
	//adapter.SimpleDelete("deleteActivityStreamMatchesByWatcher","activity_stream_matches","watcher = ?")

	adapter.SimpleDelete("deleteWordFilter", "word_filters", "wfid = ?")

	return nil
}

func writeSimpleCounts(adapter qgen.DB_Adapter) error {
	adapter.SimpleCount("reportExists", "topics", "data = ? AND data != '' AND parentID = 1", "")

	adapter.SimpleCount("groupCount", "users_groups", "", "")

	adapter.SimpleCount("modlogCount", "moderation_logs", "", "")

	return nil
}

func writeInsertSelects(adapter qgen.DB_Adapter) error {
	/*adapter.SimpleInsertSelect("addForumPermsToForumAdmins",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 1", "", ""},
	)*/

	/*adapter.SimpleInsertSelect("addForumPermsToForumStaff",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 1", "", ""},
	)*/

	/*adapter.SimpleInsertSelect("addForumPermsToForumMembers",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 0 AND is_banned = 0", "", ""},
	)*/

	return nil
}

// nolint
func writeInsertLeftJoins(adapter qgen.DB_Adapter) error {
	return nil
}

func writeInsertInnerJoins(adapter qgen.DB_Adapter) error {
	adapter.SimpleInsertInnerJoin("notifyWatchers",
		qgen.DB_Insert{"activity_stream_matches", "watcher, asid", ""},
		qgen.DB_Join{"activity_stream", "activity_subscriptions", "activity_subscriptions.user, activity_stream.asid", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""},
	)

	return nil
}

func writeFile(name string, content string) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return f.Close()
}
