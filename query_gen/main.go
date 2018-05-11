/* WIP Under Construction */
package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"

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
	for _, adapter := range qgen.Registry {
		log.Printf("Building the queries for the %s adapter", adapter.GetName())
		qgen.Install.SetAdapterInstance(adapter)
		qgen.Install.AddPlugins(NewPrimaryKeySpitter()) // TODO: Do we really need to fill the spitter for every adapter?

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
func writeStatements(adapter qgen.Adapter) error {
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

func seedTables(adapter qgen.Adapter) error {
	qgen.Install.SimpleInsert("sync", "last_update", "UTC_TIMESTAMP()")

	qgen.Install.SimpleInsert("settings", "name, content, type", "'url_tags','1','bool'")
	qgen.Install.SimpleInsert("settings", "name, content, type, constraints", "'activation_type','1','list','1-3'")
	qgen.Install.SimpleInsert("settings", "name, content, type", "'bigpost_min_words','250','int'")
	qgen.Install.SimpleInsert("settings", "name, content, type", "'megapost_min_words','1000','int'")
	qgen.Install.SimpleInsert("settings", "name, content, type", "'meta_desc','','html-attribute'")
	qgen.Install.SimpleInsert("themes", "uname, default", "'cosora',1")
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
		MoveTopic
	*/

	// TODO: Set the permissions on a struct and then serialize the struct and insert that instead of writing raw JSON
	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, is_mod, is_admin, tag", `'Administrator','{"BanUsers":true,"ActivateUsers":true,"EditUser":true,"EditUserEmail":true,"EditUserPassword":true,"EditUserGroup":true,"EditUserGroupSuperMod":true,"EditUserGroupAdmin":false,"EditGroup":true,"EditGroupLocalPerms":true,"EditGroupGlobalPerms":true,"EditGroupSuperMod":true,"EditGroupAdmin":false,"ManageForums":true,"EditSettings":true,"ManageThemes":true,"ManagePlugins":true,"ViewAdminLogs":true,"ViewIPs":true,"UploadFiles":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}','{}',1,1,"Admin"`)

	qgen.Install.SimpleInsert("users_groups", "name, permissions, plugin_perms, is_mod, tag", `'Moderator','{"BanUsers":true,"ActivateUsers":false,"EditUser":true,"EditUserEmail":false,"EditUserGroup":true,"ViewIPs":true,"UploadFiles":true,"ViewTopic":true,"LikeItem":true,"CreateTopic":true,"EditTopic":true,"DeleteTopic":true,"CreateReply":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}','{}',1,"Mod"`)

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

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `1,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `2,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `3,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `4,2,'{"ViewTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `5,2,'{"ViewTopic":true}'`)

	qgen.Install.SimpleInsert("forums_permissions", "gid, fid, permissions", `6,2,'{"ViewTopic":true}'`)

	//

	qgen.Install.SimpleInsert("topics", "title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, createdBy, parentID, ipaddress", "'Test Topic','A topic automatically generated by the software.','A topic automatically generated by the software.',UTC_TIMESTAMP(),UTC_TIMESTAMP(),1,1,2,'::1'")

	qgen.Install.SimpleInsert("replies", "tid, content, parsed_content, createdAt, createdBy, lastUpdated, lastEdit, lastEditBy, ipaddress", "1,'A reply!','A reply!',UTC_TIMESTAMP(),1,UTC_TIMESTAMP(),0,0,'::1'")

	qgen.Install.SimpleInsert("menus", "", "")

	// Go maps have a random iteration order, so we have to do this, otherwise the schema files will become unstable and harder to audit
	var order = 0
	var mOrder = "mid, name, htmlID, cssClass, position, path, aria, tooltip, guestOnly, memberOnly, staffOnly, adminOnly"
	var addMenuItem = func(data map[string]interface{}) {
		cols, values := qgen.InterfaceMapToInsertStrings(data, mOrder)
		qgen.Install.SimpleInsert("menu_items", cols+", order", values+","+strconv.Itoa(order))
		order++
	}

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_forums}", "htmlID": "menu_forums", "position": "left", "path": "/forums/", "aria": "{lang.menu_forums_aria}", "tooltip": "{lang.menu_forums_tooltip}"})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_topics}", "htmlID": "menu_topics", "cssClass": "menu_topics", "position": "left", "path": "/topics/", "aria": "{lang.menu_topics_aria}", "tooltip": "{lang.menu_topics_tooltip}"})

	addMenuItem(map[string]interface{}{"mid": 1, "htmlID": "general_alerts", "cssClass": "menu_alerts", "position": "right", "tmplName": "menu_alerts"})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_account}", "cssClass": "menu_account", "position": "left", "path": "/user/edit/critical/", "aria": "{lang.menu_account_aria}", "tooltip": "{lang.menu_account_tooltip}", "memberOnly": true})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_profile}", "cssClass": "menu_profile", "position": "left", "path": "{me.Link}", "aria": "{lang.menu_profile_aria}", "tooltip": "{lang.menu_profile_tooltip}", "memberOnly": true})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_panel}", "cssClass": "menu_panel menu_account", "position": "left", "path": "/panel/", "aria": "{lang.menu_panel_aria}", "tooltip": "{lang.menu_panel_tooltip}", "memberOnly": true, "staffOnly": true})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_logout}", "cssClass": "menu_logout", "position": "left", "path": "/accounts/logout/?session={me.Session}", "aria": "{lang.menu_logout_aria}", "tooltip": "{lang.menu_logout_tooltip}", "memberOnly": true})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_register}", "cssClass": "menu_register", "position": "left", "path": "/accounts/create/", "aria": "{lang.menu_register_aria}", "tooltip": "{lang.menu_register_tooltip}", "guestOnly": true})

	addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_login}", "cssClass": "menu_login", "position": "left", "path": "/accounts/login/", "aria": "{lang.menu_login_aria}", "tooltip": "{lang.menu_login_tooltip}", "guestOnly": true})

	return nil
}

func copyInsertMap(in map[string]interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{})
	for col, value := range in {
		out[col] = value
	}
	return out
}

type LitStr string

func writeSelects(adapter qgen.Adapter) error {
	build := adapter.Builder()

	// Looking for getTopic? Your statement is in another castle

	build.Select("getPassword").Table("users").Columns("password, salt").Where("uid = ?").Parse()

	build.Select("isPluginActive").Table("plugins").Columns("active").Where("uname = ?").Parse()

	//build.Select("isPluginInstalled").Table("plugins").Columns("installed").Where("uname = ?").Parse()

	build.Select("getUsersOffset").Table("users").Columns("uid, name, group, active, is_super_admin, avatar").Orderby("uid ASC").Limit("?,?").Parse()

	build.Select("isThemeDefault").Table("themes").Columns("default").Where("uname = ?").Parse()

	build.Select("getModlogs").Table("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Parse()

	build.Select("getModlogsOffset").Table("moderation_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Orderby("doneAt DESC").Limit("?,?").Parse()

	build.Select("getAdminlogsOffset").Table("administration_logs").Columns("action, elementID, elementType, ipaddress, actorID, doneAt").Orderby("doneAt DESC").Limit("?,?").Parse()

	build.Select("getTopicFID").Table("topics").Columns("parentID").Where("tid = ?").Parse()

	build.Select("getUserName").Table("users").Columns("name").Where("uid = ?").Parse()

	build.Select("getEmailsByUser").Table("emails").Columns("email, validated, token").Where("uid = ?").Parse()

	build.Select("getTopicBasic").Table("topics").Columns("title, content").Where("tid = ?").Parse()

	build.Select("forumEntryExists").Table("forums").Columns("fid").Where("name = ''").Orderby("fid ASC").Limit("0,1").Parse()

	build.Select("groupEntryExists").Table("users_groups").Columns("gid").Where("name = ''").Orderby("gid ASC").Limit("0,1").Parse()

	build.Select("getAttachment").Table("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path").Where("path = ? AND sectionID = ? AND sectionTable = ?").Parse()

	return nil
}

func writeLeftJoins(adapter qgen.Adapter) error {
	adapter.SimpleLeftJoin("getForumTopics", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "topics.parentID = ?", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc", "")

	return nil
}

func writeInnerJoins(adapter qgen.Adapter) (err error) {
	return nil
}

func writeInserts(adapter qgen.Adapter) error {
	build := adapter.Builder()

	build.Insert("createReport").Table("topics").Columns("title, content, parsed_content, createdAt, lastReplyAt, createdBy, lastReplyBy, data, parentID, css_class").Fields("?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,1,'report'").Parse()

	build.Insert("addForumPermsToForum").Table("forums_permissions").Columns("gid,fid,preset,permissions").Fields("?,?,?,?").Parse()

	build.Insert("addPlugin").Table("plugins").Columns("uname, active, installed").Fields("?,?,?").Parse()

	build.Insert("addTheme").Table("themes").Columns("uname, default").Fields("?,?").Parse()

	build.Insert("createWordFilter").Table("word_filters").Columns("find, replacement").Fields("?,?").Parse()

	return nil
}

func writeUpdates(adapter qgen.Adapter) error {
	build := adapter.Builder()

	build.Update("editReply").Table("replies").Set("content = ?, parsed_content = ?").Where("rid = ?").Parse()

	build.Update("updatePlugin").Table("plugins").Set("active = ?").Where("uname = ?").Parse()

	build.Update("updatePluginInstall").Table("plugins").Set("installed = ?").Where("uname = ?").Parse()

	build.Update("updateTheme").Table("themes").Set("default = ?").Where("uname = ?").Parse()

	build.Update("updateUser").Table("users").Set("name = ?, email = ?, group = ?").Where("uid = ?").Parse() // TODO: Implement user_count for users_groups on things which use this

	build.Update("updateGroupPerms").Table("users_groups").Set("permissions = ?").Where("gid = ?").Parse()

	build.Update("updateGroup").Table("users_groups").Set("name = ?, tag = ?").Where("gid = ?").Parse()

	build.Update("updateEmail").Table("emails").Set("email = ?, uid = ?, validated = ?, token = ?").Where("email = ?").Parse()

	build.Update("verifyEmail").Table("emails").Set("validated = 1, token = ''").Where("email = ?").Parse() // Need to fix this: Empty string isn't working, it gets set to 1 instead x.x -- Has this been fixed?

	build.Update("setTempGroup").Table("users").Set("temp_group = ?").Where("uid = ?").Parse()

	build.Update("updateWordFilter").Table("word_filters").Set("find = ?, replacement = ?").Where("wfid = ?").Parse()

	build.Update("bumpSync").Table("sync").Set("last_update = UTC_TIMESTAMP()").Parse()

	return nil
}

func writeDeletes(adapter qgen.Adapter) error {
	build := adapter.Builder()

	//build.Delete("deleteForumPermsByForum").Table("forums_permissions").Where("fid = ?").Parse()

	build.Delete("deleteActivityStreamMatch").Table("activity_stream_matches").Where("watcher = ? AND asid = ?").Parse()
	//build.Delete("deleteActivityStreamMatchesByWatcher").Table("activity_stream_matches").Where("watcher = ?").Parse()

	build.Delete("deleteWordFilter").Table("word_filters").Where("wfid = ?").Parse()

	return nil
}

func writeSimpleCounts(adapter qgen.Adapter) error {
	adapter.SimpleCount("reportExists", "topics", "data = ? AND data != '' AND parentID = 1", "")

	return nil
}

func writeInsertSelects(adapter qgen.Adapter) error {
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
func writeInsertLeftJoins(adapter qgen.Adapter) error {
	return nil
}

func writeInsertInnerJoins(adapter qgen.Adapter) error {
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
