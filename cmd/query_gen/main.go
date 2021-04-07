/* WIP Under Construction */
package main // import "github.com/Azareal/Gosora/query_gen"

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
)

// TODO: Make sure all the errors in this file propagate upwards properly
func main() {
	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
			return
		}
	}()

	log.Println("Running the query generator")
	for _, a := range qgen.Registry {
		log.Printf("Building the queries for the %s adapter", a.GetName())
		qgen.Install.SetAdapterInstance(a)
		qgen.Install.AddPlugins(NewPrimaryKeySpitter()) // TODO: Do we really need to fill the spitter for every adapter?

		e := writeStatements(a)
		if e != nil {
			log.Print(e)
		}
		e = qgen.Install.Write()
		if e != nil {
			log.Print(e)
		}
		e = a.Write()
		if e != nil {
			log.Print(e)
		}
	}
}

// nolint
func writeStatements(a qgen.Adapter) (err error) {
	e := func(f func(qgen.Adapter) error) {
		if err != nil {
			return
		}
		err = f(a)
	}
	e(createTables)
	e(seedTables)
	e(writeSelects)
	e(writeLeftJoins)
	e(writeInnerJoins)
	e(writeInserts)
	e(writeUpdates)
	e(writeDeletes)
	e(writeSimpleCounts)
	e(writeInsertSelects)
	e(writeInsertLeftJoins)
	e(writeInsertInnerJoins)
	return err
}

type si = map[string]interface{}
type tK = tblKey

func seedTables(a qgen.Adapter) error {
	qgen.Install.AddIndex("topics", "parentID", "parentID")
	qgen.Install.AddIndex("replies", "tid", "tid")
	qgen.Install.AddIndex("polls", "parentID", "parentID")
	qgen.Install.AddIndex("likes", "targetItem", "targetItem")
	qgen.Install.AddIndex("emails", "uid", "uid")
	qgen.Install.AddIndex("attachments", "originID", "originID")
	qgen.Install.AddIndex("attachments", "path", "path")
	qgen.Install.AddIndex("activity_stream_matches", "watcher", "watcher")
	// TODO: Remove these keys to save space when Elasticsearch is active?
	//qgen.Install.AddKey("topics", "title", tK{"title", "fulltext", "", false})
	//qgen.Install.AddKey("topics", "content", tK{"content", "fulltext", "", false})
	//qgen.Install.AddKey("topics", "title,content", tK{"title,content", "fulltext", "", false})
	//qgen.Install.AddKey("replies", "content", tK{"content", "fulltext", "", false})

	insert := func(tbl, cols, vals string) {
		qgen.Install.SimpleInsert(tbl, cols, vals)
	}
	insert("sync", "last_update", "UTC_TIMESTAMP()")
	addSetting := func(name, content, stype string, constraints ...string) {
		if strings.Contains(name, "'") {
			panic("name contains '")
		}
		if strings.Contains(stype, "'") {
			panic("stype contains '")
		}
		// TODO: Add more field validators
		cols := "name,content,type"
		if len(constraints) > 0 {
			cols += ",constraints"
		}
		q := func(s string) string {
			return "'" + s + "'"
		}
		c := func() string {
			if len(constraints) == 0 {
				return ""
			}
			return "," + q(constraints[0])
		}
		insert("settings", cols, q(name)+","+q(content)+","+q(stype)+c())
	}
	addSetting("activation_type", "1", "list", "1-3")
	addSetting("bigpost_min_words", "250", "int")
	addSetting("megapost_min_words", "1000", "int")
	addSetting("meta_desc", "", "html-attribute")
	addSetting("rapid_loading", "1", "bool")
	addSetting("google_site_verify", "", "html-attribute")
	addSetting("avatar_visibility", "0", "list", "0-1")
	insert("themes", "uname, default", "'cosora',1")
	insert("emails", "email, uid, validated", "'admin@localhost',1,1") // ? - Use a different default email or let the admin input it during installation?

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
		UploadAvatars
		UseConvos
		UseConvosOnlyWithMod
		CreateProfileReply
		AutoEmbed
		AutoLink
		// CreateConvo ?
		// CreateConvoReply ?

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

	p := func(perms *c.Perms) string {
		jBytes, err := json.Marshal(perms)
		if err != nil {
			panic(err)
		}
		return string(jBytes)
	}
	addGroup := func(name string, perms c.Perms, mod, admin, banned bool, tag string) {
		mi, ai, bi := "0", "0", "0"
		if mod {
			mi = "1"
		}
		if admin {
			ai = "1"
		}
		if banned {
			bi = "1"
		}
		insert("users_groups", "name, permissions, plugin_perms, is_mod, is_admin, is_banned, tag", `'`+name+`','`+p(&perms)+`','{}',`+mi+`,`+ai+`,`+bi+`,"`+tag+`"`)
	}

	perms := c.AllPerms
	perms.EditUserGroupAdmin = false
	perms.EditGroupAdmin = false
	addGroup("Administrator", perms, true, true, false, "Admin")

	perms = c.Perms{BanUsers: true, ActivateUsers: true, EditUser: true, EditUserEmail: false, EditUserGroup: true, ViewIPs: true, UploadFiles: true, UploadAvatars: true, UseConvos: true, UseConvosOnlyWithMod: true, CreateProfileReply: true, AutoEmbed: true, AutoLink: true, ViewTopic: true, LikeItem: true, CreateTopic: true, EditTopic: true, DeleteTopic: true, CreateReply: true, EditReply: true, DeleteReply: true, PinTopic: true, CloseTopic: true, MoveTopic: true}
	addGroup("Moderator", perms, true, false, false, "Mod")

	perms = c.Perms{UploadFiles: true, UploadAvatars: true, UseConvos: true, UseConvosOnlyWithMod: true, CreateProfileReply: true, AutoEmbed: true, AutoLink: true, ViewTopic: true, LikeItem: true, CreateTopic: true, CreateReply: true}
	addGroup("Member", perms, false, false, false, "")

	perms = c.Perms{ViewTopic: true}
	addGroup("Banned", perms, false, false, true, "")
	addGroup("Awaiting Activation", c.Perms{ViewTopic: true, UseConvosOnlyWithMod: true}, false, false, false, "")
	addGroup("Not Loggedin", perms, false, false, false, "Guest")

	//
	// TODO: Stop processFields() from stripping the spaces in the descriptions in the next commit

	insert("forums", "name, active, desc, tmpl", "'Reports',0,'All the reports go here',''")
	insert("forums", "name, lastTopicID, lastReplyerID, desc, tmpl", "'General',1,1,'A place for general discussions which don't fit elsewhere',''")

	//

	/*var addForumPerm = func(gid, fid int, permStr string) {
		insert("forums_permissions", "gid, fid, permissions", strconv.Itoa(gid)+`,`+strconv.Itoa(fid)+`,'`+permStr+`'`)
	}*/

	insert("forums_permissions", "gid, fid, permissions", `1,1,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"PinTopic":true,"CloseTopic":true}'`)
	insert("forums_permissions", "gid, fid, permissions", `2,1,'{"ViewTopic":true,"CreateReply":true,"CloseTopic":true}'`)
	insert("forums_permissions", "gid, fid, permissions", "3,1,'{}'")
	insert("forums_permissions", "gid, fid, permissions", "4,1,'{}'")
	insert("forums_permissions", "gid, fid, permissions", "5,1,'{}'")
	insert("forums_permissions", "gid, fid, permissions", "6,1,'{}'")

	//

	insert("forums_permissions", "gid, fid, permissions", `1,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}'`)

	insert("forums_permissions", "gid, fid, permissions", `2,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true,"EditTopic":true,"DeleteTopic":true,"EditReply":true,"DeleteReply":true,"PinTopic":true,"CloseTopic":true,"MoveTopic":true}'`)

	insert("forums_permissions", "gid, fid, permissions", `3,2,'{"ViewTopic":true,"CreateReply":true,"CreateTopic":true,"LikeItem":true}'`)

	insert("forums_permissions", "gid, fid, permissions", `4,2,'{"ViewTopic":true}'`)

	insert("forums_permissions", "gid, fid, permissions", `5,2,'{"ViewTopic":true}'`)

	insert("forums_permissions", "gid, fid, permissions", `6,2,'{"ViewTopic":true}'`)

	//

	insert("topics", "title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, createdBy, parentID, ip", "'Test Topic','A topic automatically generated by the software.','A topic automatically generated by the software.',UTC_TIMESTAMP(),UTC_TIMESTAMP(),1,1,2,''")

	insert("replies", "tid, content, parsed_content, createdAt, createdBy, lastUpdated, lastEdit, lastEditBy, ip", "1,'A reply!','A reply!',UTC_TIMESTAMP(),1,UTC_TIMESTAMP(),0,0,''")

	insert("menus", "", "")

	// Go maps have a random iteration order, so we have to do this, otherwise the schema files will become unstable and harder to audit
	order := 0
	mOrder := "mid, name, htmlID, cssClass, position, path, aria, tooltip, guestOnly, memberOnly, staffOnly, adminOnly"
	addMenuItem := func(data map[string]interface{}) {
		if data["mid"] == nil {
			data["mid"] = 1
		}
		if data["position"] == nil {
			data["position"] = "left"
		}
		cols, values := qgen.InterfaceMapToInsertStrings(data, mOrder)
		insert("menu_items", cols+", order", values+","+strconv.Itoa(order))
		order++
	}

	addMenuItem(si{"name": "{lang.menu_forums}", "htmlID": "menu_forums", "path": "/forums/", "aria": "{lang.menu_forums_aria}", "tooltip": "{lang.menu_forums_tooltip}"})

	addMenuItem(si{"name": "{lang.menu_topics}", "htmlID": "menu_topics", "cssClass": "menu_topics", "path": "/topics/", "aria": "{lang.menu_topics_aria}", "tooltip": "{lang.menu_topics_tooltip}"})

	addMenuItem(si{"htmlID": "general_alerts", "cssClass": "menu_alerts", "position": "right", "tmplName": "menu_alerts"})

	addMenuItem(si{"name": "{lang.menu_account}", "cssClass": "menu_account", "path": "/user/edit/", "aria": "{lang.menu_account_aria}", "tooltip": "{lang.menu_account_tooltip}", "memberOnly": true})

	addMenuItem(si{"name": "{lang.menu_profile}", "cssClass": "menu_profile", "path": "{me.Link}", "aria": "{lang.menu_profile_aria}", "tooltip": "{lang.menu_profile_tooltip}", "memberOnly": true})

	addMenuItem(si{"name": "{lang.menu_panel}", "cssClass": "menu_panel menu_account", "path": "/panel/", "aria": "{lang.menu_panel_aria}", "tooltip": "{lang.menu_panel_tooltip}", "memberOnly": true, "staffOnly": true})

	addMenuItem(si{"name": "{lang.menu_logout}", "cssClass": "menu_logout", "path": "/accounts/logout/?s={me.Session}", "aria": "{lang.menu_logout_aria}", "tooltip": "{lang.menu_logout_tooltip}", "memberOnly": true})

	addMenuItem(si{"name": "{lang.menu_register}", "cssClass": "menu_register", "path": "/accounts/create/", "aria": "{lang.menu_register_aria}", "tooltip": "{lang.menu_register_tooltip}", "guestOnly": true})

	addMenuItem(si{"name": "{lang.menu_login}", "cssClass": "menu_login", "path": "/accounts/login/", "aria": "{lang.menu_login_aria}", "tooltip": "{lang.menu_login_tooltip}", "guestOnly": true})

	/*var fSet []string
	for _, table := range tables {
		fSet = append(fSet, "'"+table+"'")
	}
	qgen.Install.SimpleBulkInsert("tables", "name", fSet)*/

	return nil
}

// ? - What is this for?
/*func copyInsertMap(in map[string]interface{}) (out map[string]interface{}) {
	out = make(map[string]interface{})
	for col, value := range in {
		out[col] = value
	}
	return out
}*/

type LitStr string

func writeSelects(a qgen.Adapter) error {
	b := a.Builder()

	// Looking for getTopic? Your statement is in another castle

	//b.Select("isPluginInstalled").Table("plugins").Columns("installed").Where("uname = ?").Parse()

	b.Select("forumEntryExists").Table("forums").Columns("fid").Where("name = ''").Orderby("fid ASC").Limit("0,1").Parse()

	b.Select("groupEntryExists").Table("users_groups").Columns("gid").Where("name = ''").Orderby("gid ASC").Limit("0,1").Parse()

	return nil
}

func writeLeftJoins(a qgen.Adapter) error {
	a.SimpleLeftJoin("getForumTopics", "topics", "users", "topics.tid, topics.title, topics.content, topics.createdBy, topics.is_closed, topics.sticky, topics.createdAt, topics.lastReplyAt, topics.parentID, users.name, users.avatar", "topics.createdBy = users.uid", "topics.parentID = ?", "topics.sticky DESC, topics.lastReplyAt DESC, topics.createdBy desc", "")

	return nil
}

func writeInnerJoins(a qgen.Adapter) (err error) {
	return nil
}

func writeInserts(a qgen.Adapter) error {
	b := a.Builder()

	b.Insert("addForumPermsToForum").Table("forums_permissions").Columns("gid,fid,preset,permissions").Fields("?,?,?,?").Parse()

	return nil
}

func writeUpdates(a qgen.Adapter) error {
	b := a.Builder()

	b.Update("updateEmail").Table("emails").Set("email = ?, uid = ?, validated = ?, token = ?").Where("email = ?").Parse()

	b.Update("setTempGroup").Table("users").Set("temp_group = ?").Where("uid = ?").Parse()

	b.Update("bumpSync").Table("sync").Set("last_update = UTC_TIMESTAMP()").Parse()

	return nil
}

func writeDeletes(a qgen.Adapter) error {
	b := a.Builder()

	//b.Delete("deleteForumPermsByForum").Table("forums_permissions").Where("fid=?").Parse()

	b.Delete("deleteActivityStreamMatch").Table("activity_stream_matches").Where("watcher=? AND asid=?").Parse()
	//b.Delete("deleteActivityStreamMatchesByWatcher").Table("activity_stream_matches").Where("watcher=?").Parse()

	return nil
}

func writeSimpleCounts(a qgen.Adapter) error {
	return nil
}

func writeInsertSelects(a qgen.Adapter) error {
	/*a.SimpleInsertSelect("addForumPermsToForumAdmins",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 1", "", ""},
	)*/

	/*a.SimpleInsertSelect("addForumPermsToForumStaff",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 1", "", ""},
	)*/

	/*a.SimpleInsertSelect("addForumPermsToForumMembers",
		qgen.DB_Insert{"forums_permissions", "gid, fid, preset, permissions", ""},
		qgen.DB_Select{"users_groups", "gid, ? AS fid, ? AS preset, ? AS permissions", "is_admin = 0 AND is_mod = 0 AND is_banned = 0", "", ""},
	)*/

	return nil
}

// nolint
func writeInsertLeftJoins(a qgen.Adapter) error {
	return nil
}

func writeInsertInnerJoins(a qgen.Adapter) error {
	return nil
}

func writeFile(name, content string) (err error) {
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
