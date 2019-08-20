package main

import (
	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/extend/guilds/lib"
)

// TODO: Add a better way of splitting up giant plugins like this

// TODO: Add a plugin interface instead of having a bunch of argument to AddPlugin?
func init() {
	c.Plugins.Add(&c.Plugin{UName: "guilds", Name: "Guilds", Author: "Azareal", URL: "https://github.com/Azareal", Init: initGuilds, Deactivate: deactivateGuilds, Install: installGuilds})

	// TODO: Is it possible to avoid doing this when the plugin isn't activated?
	c.PrebuildTmplList = append(c.PrebuildTmplList, guilds.PrebuildTmplList)
}

func initGuilds(pl *c.Plugin) (err error) {
	pl.AddHook("intercept_build_widgets", guilds.Widgets)
	pl.AddHook("trow_assign", guilds.TrowAssign)
	pl.AddHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	pl.AddHook("pre_render_forum", guilds.PreRenderViewForum)
	pl.AddHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	pl.AddHook("forum_check_pre_perms", guilds.ForumCheck)
	// TODO: Auto-grant this perm to admins upon installation?
	c.RegisterPluginPerm("CreateGuild")
	router.HandleFunc("/guilds/", guilds.RouteGuildList)
	router.HandleFunc("/guild/", guilds.MiddleViewGuild)
	router.HandleFunc("/guild/create/", guilds.RouteCreateGuild)
	router.HandleFunc("/guild/create/submit/", guilds.RouteCreateGuildSubmit)
	router.HandleFunc("/guild/members/", guilds.RouteMemberList)

	guilds.Gstore, err = guilds.NewSQLGuildStore()
	if err != nil {
		return err
	}

	acc := qgen.NewAcc()

	guilds.ListStmt = acc.Select("guilds").Columns("guildID, name, desc, active, privacy, joinable, owner, memberCount, createdAt, lastUpdateTime").Prepare()

	guilds.MemberListStmt = acc.Select("guilds_members").Columns("guildID, uid, rank, posts, joinedAt").Prepare()

	guilds.MemberListJoinStmt = acc.SimpleLeftJoin("guilds_members", "users", "users.uid, guilds_members.rank, guilds_members.posts, guilds_members.joinedAt, users.name, users.avatar", "guilds_members.uid = users.uid", "guilds_members.guildID = ?", "guilds_members.rank DESC, guilds_members.joinedat ASC", "")

	guilds.GetMemberStmt = acc.Select("guilds_members").Columns("rank, posts, joinedAt").Where("guildID = ? AND uid = ?").Prepare()

	guilds.AttachForumStmt = acc.Update("forums").Set("parentID = ?, parentType = 'guild'").Where("fid = ?").Prepare()

	guilds.UnattachForumStmt = acc.Update("forums").Set("parentID = 0, parentType = ''").Where("fid = ?").Prepare()

	guilds.AddMemberStmt = acc.Insert("guilds_members").Columns("guildID, uid, rank, posts, joinedAt").Fields("?,?,?,0,UTC_TIMESTAMP()").Prepare()

	return acc.FirstError()
}

func deactivateGuilds(pl *common.Plugin) {
	pl.RemoveHook("intercept_build_widgets", guilds.Widgets)
	pl.RemoveHook("trow_assign", guilds.TrowAssign)
	pl.RemoveHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	pl.RemoveHook("pre_render_forum", guilds.PreRenderViewForum)
	pl.RemoveHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	pl.RemoveHook("forum_check_pre_perms", guilds.ForumCheck)
	c.DeregisterPluginPerm("CreateGuild")
	_ = router.RemoveFunc("/guilds/")
	_ = router.RemoveFunc("/guild/")
	_ = router.RemoveFunc("/guild/create/")
	_ = router.RemoveFunc("/guild/create/submit/")
	_ = guilds.ListStmt.Close()
	_ = guilds.MemberListStmt.Close()
	_ = guilds.MemberListJoinStmt.Close()
	_ = guilds.GetMemberStmt.Close()
	_ = guilds.AttachForumStmt.Close()
	_ = guilds.UnattachForumStmt.Close()
	_ = guilds.AddMemberStmt.Close()
}

// TODO: Stop accessing the query builder directly and add a feature in Gosora which is more easily reversed, if an error comes up during the installation process
type tC = qgen.DBTableColumn
func installGuilds(plugin *common.Plugin) error {
	guildTableStmt, err := qgen.Builder.CreateTable("guilds", "utf8mb4", "utf8mb4_general_ci",
		[]tC{
			tC{"guildID", "int", 0, false, true, ""},
			tC{"name", "varchar", 100, false, false, ""},
			tC{"desc", "varchar", 200, false, false, ""},
			tC{"active", "boolean", 1, false, false, ""},
			tC{"privacy", "smallint", 0, false, false, ""},
			tC{"joinable", "smallint", 0, false, false, "0"},
			tC{"owner", "int", 0, false, false, ""},
			tC{"memberCount", "int", 0, false, false, ""},
			tC{"mainForum", "int", 0, false, false, "0"}, // The board the user lands on when they click on a group, we'll make it possible for group admins to change what users land on
			//tC{"boards","varchar",255,false,false,""}, // Cap the max number of boards at 8 to avoid overflowing the confines of a 64-bit integer?
			tC{"backdrop", "varchar", 200, false, false, ""}, // File extension for the uploaded file, or an external link
			tC{"createdAt", "createdAt", 0, false, false, ""},
			tC{"lastUpdateTime", "datetime", 0, false, false, ""},
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
		[]tC{
			tC{"guildID", "int", 0, false, false, ""},
			tC{"uid", "int", 0, false, false, ""},
			tC{"rank", "int", 0, false, false, "0"},  /* 0: Member. 1: Mod. 2: Admin. */
			tC{"posts", "int", 0, false, false, "0"}, /* Per-Group post count. Should we do some sort of score system? */
			tC{"joinedAt", "datetime", 0, false, false, ""},
		}, nil,
	)
	if err != nil {
		return err
	}

	_, err = guildMembersTableStmt.Exec()
	return err
}

// TO-DO; Implement an uninstallation system into Gosora. And a better installation system.
func uninstallGuilds(plugin *c.Plugin) error {
	return nil
}
