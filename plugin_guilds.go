package main

import (
	//"fmt"

	"./common"
	"./extend/guilds/lib"
	"./query_gen/lib"
)

// TODO: Add a better way of splitting up giant plugins like this

// TODO: Add a plugin interface instead of having a bunch of argument to AddPlugin?
func init() {
	common.Plugins["guilds"] = common.NewPlugin("guilds", "Guilds", "Azareal", "http://github.com/Azareal", "", "", "", initGuilds, nil, deactivateGuilds, installGuilds, nil)

	// TODO: Is it possible to avoid doing this when the plugin isn't activated?
	common.PrebuildTmplList = append(common.PrebuildTmplList, guilds.PrebuildTmplList)
}

func initGuilds() (err error) {
	common.Plugins["guilds"].AddHook("intercept_build_widgets", guilds.Widgets)
	common.Plugins["guilds"].AddHook("trow_assign", guilds.TrowAssign)
	common.Plugins["guilds"].AddHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	common.Plugins["guilds"].AddHook("pre_render_view_forum", guilds.PreRenderViewForum)
	common.Plugins["guilds"].AddHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	common.Plugins["guilds"].AddHook("forum_check_pre_perms", guilds.ForumCheck)
	// TODO: Auto-grant this perm to admins upon installation?
	common.RegisterPluginPerm("CreateGuild")
	router.HandleFunc("/guilds/", guilds.RouteGuildList)
	router.HandleFunc("/guild/", guilds.MiddleViewGuild)
	router.HandleFunc("/guild/create/", guilds.RouteCreateGuild)
	router.HandleFunc("/guild/create/submit/", guilds.RouteCreateGuildSubmit)
	router.HandleFunc("/guild/members/", guilds.RouteMemberList)

	acc := qgen.Builder.Accumulator()

	guilds.ListStmt = acc.Select("guilds").Columns("guildID, name, desc, active, privacy, joinable, owner, memberCount, createdAt, lastUpdateTime").Prepare()

	guilds.GetGuildStmt = acc.Select("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Where("guildID = ?").Prepare()

	guilds.MemberListStmt = acc.Select("guilds_members").Columns("guildID, uid, rank, posts, joinedAt").Prepare()

	guilds.MemberListJoinStmt = acc.SimpleLeftJoin("guilds_members", "users", "users.uid, guilds_members.rank, guilds_members.posts, guilds_members.joinedAt, users.name, users.avatar", "guilds_members.uid = users.uid", "guilds_members.guildID = ?", "guilds_members.rank DESC, guilds_members.joinedat ASC", "")

	guilds.GetMemberStmt = acc.Select("guilds_members").Columns("rank, posts, joinedAt").Where("guildID = ? AND uid = ?").Prepare()

	guilds.CreateGuildStmt = acc.Insert("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Fields("?,?,?,?,1,?,1,?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare()

	guilds.AttachForumStmt = acc.Update("forums").Set("parentID = ?, parentType = 'guild'").Where("fid = ?").Prepare()

	guilds.UnattachForumStmt = acc.Update("forums").Set("parentID = 0, parentType = ''").Where("fid = ?").Prepare()

	guilds.AddMemberStmt = acc.Insert("guilds_members").Columns("guildID, uid, rank, posts, joinedAt").Fields("?,?,?,0,UTC_TIMESTAMP()").Prepare()

	return acc.FirstError()
}

func deactivateGuilds() {
	common.Plugins["guilds"].RemoveHook("intercept_build_widgets", guilds.Widgets)
	common.Plugins["guilds"].RemoveHook("trow_assign", guilds.TrowAssign)
	common.Plugins["guilds"].RemoveHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	common.Plugins["guilds"].RemoveHook("pre_render_view_forum", guilds.PreRenderViewForum)
	common.Plugins["guilds"].RemoveHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	common.Plugins["guilds"].RemoveHook("forum_check_pre_perms", guilds.ForumCheck)
	common.DeregisterPluginPerm("CreateGuild")
	_ = router.RemoveFunc("/guilds/")
	_ = router.RemoveFunc("/guild/")
	_ = router.RemoveFunc("/guild/create/")
	_ = router.RemoveFunc("/guild/create/submit/")
	_ = guilds.ListStmt.Close()
	_ = guilds.MemberListStmt.Close()
	_ = guilds.MemberListJoinStmt.Close()
	_ = guilds.GetMemberStmt.Close()
	_ = guilds.GetGuildStmt.Close()
	_ = guilds.CreateGuildStmt.Close()
	_ = guilds.AttachForumStmt.Close()
	_ = guilds.UnattachForumStmt.Close()
	_ = guilds.AddMemberStmt.Close()
}

// TODO: Stop accessing the query builder directly and add a feature in Gosora which is more easily reversed, if an error comes up during the installation process
func installGuilds() error {
	guildTableStmt, err := qgen.Builder.CreateTable("guilds", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"guildID", "int", 0, false, true, ""},
			qgen.DBTableColumn{"name", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"desc", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"active", "boolean", 1, false, false, ""},
			qgen.DBTableColumn{"privacy", "smallint", 0, false, false, ""},
			qgen.DBTableColumn{"joinable", "smallint", 0, false, false, "0"},
			qgen.DBTableColumn{"owner", "int", 0, false, false, ""},
			qgen.DBTableColumn{"memberCount", "int", 0, false, false, ""},
			qgen.DBTableColumn{"mainForum", "int", 0, false, false, "0"}, // The board the user lands on when they click on a group, we'll make it possible for group admins to change what users land on
			//qgen.DBTableColumn{"boards","varchar",255,false,false,""}, // Cap the max number of boards at 8 to avoid overflowing the confines of a 64-bit integer?
			qgen.DBTableColumn{"backdrop", "varchar", 200, false, false, ""}, // File extension for the uploaded file, or an external link
			qgen.DBTableColumn{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DBTableColumn{"lastUpdateTime", "datetime", 0, false, false, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"guildID", "primary"},
		},
	)
	if err != nil {
		return err
	}

	_, err = guildTableStmt.Exec()
	if err != nil {
		return err
	}

	guildMembersTableStmt, err := qgen.Builder.CreateTable("guilds_members", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"guildID", "int", 0, false, false, ""},
			qgen.DBTableColumn{"uid", "int", 0, false, false, ""},
			qgen.DBTableColumn{"rank", "int", 0, false, false, "0"},  /* 0: Member. 1: Mod. 2: Admin. */
			qgen.DBTableColumn{"posts", "int", 0, false, false, "0"}, /* Per-Group post count. Should we do some sort of score system? */
			qgen.DBTableColumn{"joinedAt", "datetime", 0, false, false, ""},
		},
		[]qgen.DBTableKey{},
	)
	if err != nil {
		return err
	}

	_, err = guildMembersTableStmt.Exec()
	return err
}

// TO-DO; Implement an uninstallation system into Gosora. And a better installation system.
func uninstallGuilds() error {
	return nil
}
