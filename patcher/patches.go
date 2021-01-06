package main

import (
	"bufio"
	"database/sql"
	"strconv"
	"strings"
	"unicode"

	meta "github.com/Azareal/Gosora/common/meta"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type tblColumn = qgen.DBTableColumn
type tC = tblColumn
type tblKey = qgen.DBTableKey
type tK = tblKey

func init() {
	addPatch(0, patch0)
	addPatch(1, patch1)
	addPatch(2, patch2)
	addPatch(3, patch3)
	addPatch(4, patch4)
	addPatch(5, patch5)
	addPatch(6, patch6)
	addPatch(7, patch7)
	addPatch(8, patch8)
	addPatch(9, patch9)
	addPatch(10, patch10)
	addPatch(11, patch11)
	addPatch(12, patch12)
	addPatch(13, patch13)
	addPatch(14, patch14)
	addPatch(15, patch15)
	addPatch(16, patch16)
	addPatch(17, patch17)
	addPatch(18, patch18)
	addPatch(19, patch19)
	addPatch(20, patch20)
	addPatch(21, patch21)
	addPatch(22, patch22)
	addPatch(23, patch23)
	addPatch(24, patch24)
	addPatch(25, patch25)
	addPatch(26, patch26)
	addPatch(27, patch27)
	addPatch(28, patch28)
	addPatch(29, patch29)
	addPatch(30, patch30)
	addPatch(31, patch31)
	addPatch(32, patch32)
	addPatch(33, patch33)
	addPatch(34, patch34)
	addPatch(35, patch35)
}

func bcol(col string, val bool) qgen.DBTableColumn {
	if val {
		return tC{col, "boolean", 0, false, false, "1"}
	}
	return tC{col, "boolean", 0, false, false, "0"}
}
func ccol(col string, size int, sdefault string) qgen.DBTableColumn {
	return tC{col, "varchar", size, false, false, sdefault}
}

func patch0(scanner *bufio.Scanner) (err error) {
	err = createTable("menus", "", "",
		[]tC{
			{"mid", "int", 0, false, true, ""},
		},
		[]tK{
			{"mid", "primary", "", false},
		},
	)
	if err != nil {
		return err
	}

	err = createTable("menu_items", "", "",
		[]tC{
			{"miid", "int", 0, false, true, ""},
			{"mid", "int", 0, false, false, ""},
			ccol("name", 200, ""),
			ccol("htmlID", 200, "''"),
			ccol("cssClass", 200, "''"),
			ccol("position", 100, ""),
			ccol("path", 200, "''"),
			ccol("aria", 200, "''"),
			ccol("tooltip", 200, "''"),
			ccol("tmplName", 200, "''"),
			{"order", "int", 0, false, false, "0"},

			bcol("guestOnly", false),
			bcol("memberOnly", false),
			bcol("staffOnly", false),
			bcol("adminOnly", false),
		},
		[]tK{
			{"miid", "primary", "", false},
		},
	)
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleInsert("menus", "", ""))
	if err != nil {
		return err
	}

	var order int
	mOrder := "mid, name, htmlID, cssClass, position, path, aria, tooltip, guestOnly, memberOnly, staffOnly, adminOnly"
	addMenuItem := func(data map[string]interface{}) error {
		cols, values := qgen.InterfaceMapToInsertStrings(data, mOrder)
		err := execStmt(qgen.Builder.SimpleInsert("menu_items", cols+", order", values+","+strconv.Itoa(order)))
		order++
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_forums}", "htmlID": "menu_forums", "position": "left", "path": "/forums/", "aria": "{lang.menu_forums_aria}", "tooltip": "{lang.menu_forums_tooltip}"})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_topics}", "htmlID": "menu_topics", "cssClass": "menu_topics", "position": "left", "path": "/topics/", "aria": "{lang.menu_topics_aria}", "tooltip": "{lang.menu_topics_tooltip}"})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "htmlID": "general_alerts", "cssClass": "menu_alerts", "position": "right", "tmplName": "menu_alerts"})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_account}", "cssClass": "menu_account", "position": "left", "path": "/user/edit/critical/", "aria": "{lang.menu_account_aria}", "tooltip": "{lang.menu_account_tooltip}", "memberOnly": true})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_profile}", "cssClass": "menu_profile", "position": "left", "path": "{me.Link}", "aria": "{lang.menu_profile_aria}", "tooltip": "{lang.menu_profile_tooltip}", "memberOnly": true})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_panel}", "cssClass": "menu_panel menu_account", "position": "left", "path": "/panel/", "aria": "{lang.menu_panel_aria}", "tooltip": "{lang.menu_panel_tooltip}", "memberOnly": true, "staffOnly": true})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_logout}", "cssClass": "menu_logout", "position": "left", "path": "/accounts/logout/?session={me.Session}", "aria": "{lang.menu_logout_aria}", "tooltip": "{lang.menu_logout_tooltip}", "memberOnly": true})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_register}", "cssClass": "menu_register", "position": "left", "path": "/accounts/create/", "aria": "{lang.menu_register_aria}", "tooltip": "{lang.menu_register_tooltip}", "guestOnly": true})
	if err != nil {
		return err
	}

	err = addMenuItem(map[string]interface{}{"mid": 1, "name": "{lang.menu_login}", "cssClass": "menu_login", "position": "left", "path": "/accounts/login/", "aria": "{lang.menu_login_aria}", "tooltip": "{lang.menu_login_tooltip}", "guestOnly": true})
	if err != nil {
		return err
	}

	return nil
}

func patch1(scanner *bufio.Scanner) error {
	routes := map[string]string{
		"routeAccountEditCriticalSubmit": "routes.AccountEditCriticalSubmit",
		"routeAccountEditAvatar":         "routes.AccountEditAvatar",
		"routeAccountEditAvatarSubmit":   "routes.AccountEditAvatarSubmit",
		"routeAccountEditUsername":       "routes.AccountEditUsername",
		"routeAccountEditUsernameSubmit": "routes.AccountEditUsernameSubmit",
	}
	return renameRoutes(routes)
}

func patch2(scanner *bufio.Scanner) error {
	routes := map[string]string{
		"routeLogout":                   "routes.AccountLogout",
		"routeShowAttachment":           "routes.ShowAttachment",
		"routeChangeTheme":              "routes.ChangeTheme",
		"routeProfileReplyCreateSubmit": "routes.ProfileReplyCreateSubmit",
		"routeLikeTopicSubmit":          "routes.LikeTopicSubmit",
		"routeReplyLikeSubmit":          "routes.ReplyLikeSubmit",
		"routeDynamic":                  "routes.DynamicRoute",
		"routeUploads":                  "routes.UploadedFile",
		"BadRoute":                      "routes.BadRoute",
	}
	return renameRoutes(routes)
}

func patch3(scanner *bufio.Scanner) error {
	return createTable("registration_logs", "", "",
		[]tC{
			{"rlid", "int", 0, false, true, ""},
			ccol("username", 100, ""),
			ccol("email", 100, ""),
			ccol("failureReason", 100, ""),
			bcol("success", false), // Did this attempt succeed?
			ccol("ipaddress", 200, ""),
			{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tK{
			{"rlid", "primary", "", false},
		},
	)
}

func patch4(scanner *bufio.Scanner) error {
	routes := map[string]string{
		"routeReportSubmit":                      "routes.ReportSubmit",
		"routeAccountEditEmail":                  "routes.AccountEditEmail",
		"routeAccountEditEmailTokenSubmit":       "routes.AccountEditEmailTokenSubmit",
		"routePanelLogsRegs":                     "panel.LogsRegs",
		"routePanelLogsMod":                      "panel.LogsMod",
		"routePanelLogsAdmin":                    "panel.LogsAdmin",
		"routePanelDebug":                        "panel.Debug",
		"routePanelAnalyticsViews":               "panel.AnalyticsViews",
		"routePanelAnalyticsRouteViews":          "panel.AnalyticsRouteViews",
		"routePanelAnalyticsAgentViews":          "panel.AnalyticsAgentViews",
		"routePanelAnalyticsForumViews":          "panel.AnalyticsForumViews",
		"routePanelAnalyticsSystemViews":         "panel.AnalyticsSystemViews",
		"routePanelAnalyticsLanguageViews":       "panel.AnalyticsLanguageViews",
		"routePanelAnalyticsReferrerViews":       "panel.AnalyticsReferrerViews",
		"routePanelAnalyticsTopics":              "panel.AnalyticsTopics",
		"routePanelAnalyticsPosts":               "panel.AnalyticsPosts",
		"routePanelAnalyticsForums":              "panel.AnalyticsForums",
		"routePanelAnalyticsRoutes":              "panel.AnalyticsRoutes",
		"routePanelAnalyticsAgents":              "panel.AnalyticsAgents",
		"routePanelAnalyticsSystems":             "panel.AnalyticsSystems",
		"routePanelAnalyticsLanguages":           "panel.AnalyticsLanguages",
		"routePanelAnalyticsReferrers":           "panel.AnalyticsReferrers",
		"routePanelSettings":                     "panel.Settings",
		"routePanelSettingEdit":                  "panel.SettingEdit",
		"routePanelSettingEditSubmit":            "panel.SettingEditSubmit",
		"routePanelForums":                       "panel.Forums",
		"routePanelForumsCreateSubmit":           "panel.ForumsCreateSubmit",
		"routePanelForumsDelete":                 "panel.ForumsDelete",
		"routePanelForumsDeleteSubmit":           "panel.ForumsDeleteSubmit",
		"routePanelForumsEdit":                   "panel.ForumsEdit",
		"routePanelForumsEditSubmit":             "panel.ForumsEditSubmit",
		"routePanelForumsEditPermsSubmit":        "panel.ForumsEditPermsSubmit",
		"routePanelForumsEditPermsAdvance":       "panel.ForumsEditPermsAdvance",
		"routePanelForumsEditPermsAdvanceSubmit": "panel.ForumsEditPermsAdvanceSubmit",
		"routePanelBackups":                      "panel.Backups",
	}
	err := renameRoutes(routes)
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleDelete("settings", "name='url_tags'"))
	if err != nil {
		return err
	}

	return createTable("pages", "utf8mb4", "utf8mb4_general_ci",
		[]tC{
			{"pid", "int", 0, false, true, ""},
			ccol("name", 200, ""),
			ccol("title", 200, ""),
			{"body", "text", 0, false, false, ""},
			{"allowedGroups", "text", 0, false, false, ""},
			{"menuID", "int", 0, false, false, "-1"},
		},
		[]tK{
			{"pid", "primary", "", false},
		},
	)
}

func patch5(scanner *bufio.Scanner) error {
	routes := map[string]string{
		"routePanelUsers":                  "panel.Users",
		"routePanelUsersEdit":              "panel.UsersEdit",
		"routePanelUsersEditSubmit":        "panel.UsersEditSubmit",
		"routes.AccountEditCritical":       "routes.AccountEditPassword",
		"routes.AccountEditCriticalSubmit": "routes.AccountEditPasswordSubmit",
	}
	err := renameRoutes(routes)
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleUpdate("menu_items", "path='/user/edit/'", "path='/user/edit/critical/'"))
	if err != nil {
		return err
	}

	return createTable("users_2fa_keys", "utf8mb4", "utf8mb4_general_ci",
		[]tC{
			{"uid", "int", 0, false, false, ""},
			ccol("secret", 100, ""),
			ccol("scratch1", 50, ""),
			ccol("scratch2", 50, ""),
			ccol("scratch3", 50, ""),
			ccol("scratch4", 50, ""),
			ccol("scratch5", 50, ""),
			ccol("scratch6", 50, ""),
			ccol("scratch7", 50, ""),
			ccol("scratch8", 50, ""),
			{"createdAt", "createdAt", 0, false, false, ""},
		},
		[]tK{
			{"uid", "primary", "", false},
		},
	)
}

func patch6(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.SimpleInsert("settings", "name, content, type", "'rapid_loading','1','bool'"))
}

func patch7(scanner *bufio.Scanner) error {
	return createTable("users_avatar_queue", "", "",
		[]tC{
			{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
		},
		[]tK{
			{"uid", "primary", "", false},
		},
	)
}

func renameRoutes(routes map[string]string) error {
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	replaceTextWhere := func(replaceThis string, withThis string) error {
		return execStmt(qgen.Builder.SimpleUpdate("viewchunks", "route = '"+withThis+"'", "route = '"+replaceThis+"'"))
	}

	for key, value := range routes {
		err := replaceTextWhere(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func patch8(scanner *bufio.Scanner) error {
	routes := map[string]string{
		"routePanelWordFilter":                 "panel.WordFilters",
		"routePanelWordFiltersCreateSubmit":    "panel.WordFiltersCreateSubmit",
		"routePanelWordFiltersEdit":            "panel.WordFiltersEdit",
		"routePanelWordFiltersEditSubmit":      "panel.WordFiltersEditSubmit",
		"routePanelWordFiltersDeleteSubmit":    "panel.WordFiltersDeleteSubmit",
		"routePanelPlugins":                    "panel.Plugins",
		"routePanelPluginsActivate":            "panel.PluginsActivate",
		"routePanelPluginsDeactivate":          "panel.PluginsDeactivate",
		"routePanelPluginsInstall":             "panel.PluginsInstall",
		"routePanelGroups":                     "panel.Groups",
		"routePanelGroupsEdit":                 "panel.GroupsEdit",
		"routePanelGroupsEditPerms":            "panel.GroupsEditPerms",
		"routePanelGroupsEditSubmit":           "panel.GroupsEditSubmit",
		"routePanelGroupsEditPermsSubmit":      "panel.GroupsEditPermsSubmit",
		"routePanelGroupsCreateSubmit":         "panel.GroupsCreateSubmit",
		"routePanelThemes":                     "panel.Themes",
		"routePanelThemesSetDefault":           "panel.ThemesSetDefault",
		"routePanelThemesMenus":                "panel.ThemesMenus",
		"routePanelThemesMenusEdit":            "panel.ThemesMenusEdit",
		"routePanelThemesMenuItemEdit":         "panel.ThemesMenuItemEdit",
		"routePanelThemesMenuItemEditSubmit":   "panel.ThemesMenuItemEditSubmit",
		"routePanelThemesMenuItemCreateSubmit": "panel.ThemesMenuItemCreateSubmit",
		"routePanelThemesMenuItemDeleteSubmit": "panel.ThemesMenuItemDeleteSubmit",
		"routePanelThemesMenuItemOrderSubmit":  "panel.ThemesMenuItemOrderSubmit",
		"routePanelDashboard":                  "panel.Dashboard",
	}
	err := renameRoutes(routes)
	if err != nil {
		return err
	}

	return createTable("updates", "", "",
		[]tC{
			{"dbVersion", "int", 0, false, false, "0"},
		}, nil,
	)
}

func patch9(scanner *bufio.Scanner) error {
	// Table "updates" might not exist due to the installer, so drop it and remake it if so
	err := patch8(scanner)
	if err != nil {
		return err
	}

	return createTable("login_logs", "", "",
		[]tC{
			{"lid", "int", 0, false, true, ""},
			{"uid", "int", 0, false, false, ""},
			bcol("success", false), // Did this attempt succeed?
			ccol("ipaddress", 200, ""),
			{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tK{
			{"lid", "primary", "", false},
		},
	)
}

var acc = qgen.NewAcc
var itoa = strconv.Itoa

func patch10(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("topics", tC{"attachCount", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddColumn("topics", tC{"lastReplyID", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	err = acc().Select("topics").Cols("tid").EachInt(func(tid int) error {
		stid := itoa(tid)
		count, err := acc().Count("attachments").Where("originTable = 'topics' and originID=" + stid).Total()
		if err != nil {
			return err
		}

		hasReply := false
		err = acc().Select("replies").Cols("rid").Where("tid=" + stid).Orderby("rid DESC").Limit("1").EachInt(func(rid int) error {
			hasReply = true
			_, err := acc().Update("topics").Set("lastReplyID=?, attachCount=?").Where("tid="+stid).Exec(rid, count)
			return err
		})
		if err != nil {
			return err
		}
		if !hasReply {
			_, err = acc().Update("topics").Set("attachCount=?").Where("tid=" + stid).Exec(count)
		}
		return err
	})
	if err != nil {
		return err
	}

	_, err = acc().Insert("updates").Columns("dbVersion").Fields("0").Exec()
	return err
}

func patch11(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("replies", tC{"attachCount", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	// Attachments for replies got the topicID rather than the replyID for a while in error, so we want to separate these out
	_, err = acc().Update("attachments").Set("originTable='freplies'").Where("originTable='replies'").Exec()
	if err != nil {
		return err
	}

	// We could probably do something more efficient, but as there shouldn't be too many sites right now, we can probably cheat a little, otherwise it'll take forever to get things done
	return acc().Select("topics").Cols("tid").EachInt(func(tid int) error {
		stid := itoa(tid)
		count, err := acc().Count("attachments").Where("originTable='topics' and originID=" + stid).Total()
		if err != nil {
			return err
		}
		_, err = acc().Update("topics").Set("attachCount=?").Where("tid=" + stid).Exec(count)
		return err
	})

	/*return acc().Select("replies").Cols("rid").EachInt(func(rid int) error {
		srid := itoa(rid)
		count, err := acc().Count("attachments").Where("originTable='replies' and originID=" + srid).Total()
		if err != nil {
			return err
		}
		_, err = acc().Update("replies").Set("attachCount = ?").Where("rid=" + srid).Exec(count)
		return err
	})*/
}

func patch12(scanner *bufio.Scanner) error {
	var e error
	addIndex := func(tbl, iname, colname string) {
		if e != nil {
			return
		}
		/*e = execStmt(qgen.Builder.RemoveIndex(tbl, iname))
		if e != nil {
			return
		}*/
		e = execStmt(qgen.Builder.AddIndex(tbl, iname, colname))
	}
	addIndex("topics", "parentID", "parentID")
	addIndex("replies", "tid", "tid")
	addIndex("polls", "parentID", "parentID")
	addIndex("likes", "targetItem", "targetItem")
	addIndex("emails", "uid", "uid")
	addIndex("attachments", "originID", "originID")
	addIndex("attachments", "path", "path")
	addIndex("activity_stream_matches", "watcher", "watcher")
	return e
}

func patch13(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("widgets", tC{"wid", "int", 0, false, true, ""}, &tK{"wid", "primary", "", false}))
}

func patch14(scanner *bufio.Scanner) error {
	/*err := execStmt(qgen.Builder.AddKey("topics", "title", tK{"title", "fulltext", "", false}))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddKey("topics", "content", tK{"content", "fulltext", "", false}))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddKey("replies", "content", tK{"content", "fulltext", "", false}))
	if err != nil {
		return err
	}*/

	return nil
}

func patch15(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.SimpleInsert("settings", "name, content, type", "'google_site_verify','','html-attribute'"))
}

func patch16(scanner *bufio.Scanner) error {
	return createTable("password_resets", "", "",
		[]tC{
			ccol("email", 200, ""),
			{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			ccol("validated", 200, ""),          // Token given once the one-use token is consumed, used to prevent multiple people consuming the same one-use token
			ccol("token", 200, ""),
			{"createdAt", "createdAt", 0, false, false, ""},
		}, nil,
	)
}

func patch17(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("attachments", ccol("extra", 200, ""), nil))
	if err != nil {
		return err
	}

	err = acc().Select("topics").Cols("tid,parentID").Where("attachCount > 0").Each(func(rows *sql.Rows) error {
		var tid, parentID int
		err := rows.Scan(&tid, &parentID)
		if err != nil {
			return err
		}
		_, err = acc().Update("attachments").Set("sectionID=?").Where("originTable='topics' AND originID=?").Exec(parentID, tid)
		return err
	})
	if err != nil {
		return err
	}

	return acc().Select("replies").Cols("rid,tid").Where("attachCount > 0").Each(func(rows *sql.Rows) error {
		var rid, tid, sectionID int
		err := rows.Scan(&rid, &tid)
		if err != nil {
			return err
		}
		err = acc().Select("topics").Cols("parentID").Where("tid=?").QueryRow(tid).Scan(&sectionID)
		if err != nil {
			return err
		}
		_, err = acc().Update("attachments").Set("sectionID=?, extra=?").Where("originTable='replies' AND originID=?").Exec(sectionID, tid, rid)
		return err
	})
}

func patch18(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("forums", tC{"order", "int", 0, false, false, "0"}, nil))
}

func patch19(scanner *bufio.Scanner) error {
	return createTable("memchunks", "", "",
		[]tC{
			{"count", "int", 0, false, false, "0"},
			{"createdAt", "datetime", 0, false, false, ""},
		}, nil,
	)
}

func patch20(scanner *bufio.Scanner) error {
	err := acc().Select("activity_stream_matches").Cols("asid").Each(func(rows *sql.Rows) error {
		var asid int
		if e := rows.Scan(&asid); e != nil {
			return e
		}
		e := acc().Select("activity_stream").Cols("asid").Where("asid=?").QueryRow(asid).Scan(&asid)
		if e != sql.ErrNoRows {
			return e
		}
		_, e = acc().Delete("activity_stream_matches").Where("asid=?").Run(asid)
		return e
	})
	if err != nil {
		return err
	}

	return execStmt(qgen.Builder.AddForeignKey("activity_stream_matches", "asid", "activity_stream", "asid", true))
}

func patch21(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("memchunks", tC{"stack", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.AddColumn("memchunks", tC{"heap", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	err = createTable("meta", "", "",
		[]tC{
			ccol("name", 200, ""),
			ccol("value", 200, ""),
		}, nil,
	)
	if err != nil {
		return err
	}

	return execStmt(qgen.Builder.AddColumn("activity_stream", tC{"createdAt", "createdAt", 0, false, false, ""}, nil))
}

func patch22(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("forums", ccol("tmpl", 200, "''"), nil))
}

func patch23(scanner *bufio.Scanner) error {
	err := createTable("conversations", "", "",
		[]tC{
			{"cid", "int", 0, false, true, ""},
			{"createdBy", "int", 0, false, false, ""}, // TODO: Make this a foreign key
			{"createdAt", "createdAt", 0, false, false, ""},
			{"lastReplyAt", "datetime", 0, false, false, ""},
			{"lastReplyBy", "int", 0, false, false, ""},
		},
		[]tK{
			{"cid", "primary", "", false},
		},
	)
	if err != nil {
		return err
	}

	err = createTable("conversations_posts", "", "",
		[]tC{
			{"pid", "int", 0, false, true, ""},
			{"cid", "int", 0, false, false, ""},
			{"createdBy", "int", 0, false, false, ""},
			ccol("body", 50, ""),
			ccol("post", 50, "''"),
		},
		[]tK{
			{"pid", "primary", "", false},
		},
	)
	if err != nil {
		return err
	}

	return createTable("conversations_participants", "", "",
		[]tC{
			{"uid", "int", 0, false, false, ""},
			{"cid", "int", 0, false, false, ""},
		}, nil,
	)
}

func patch24(scanner *bufio.Scanner) error {
	return createTable("users_groups_promotions", "", "",
		[]tC{
			{"pid", "int", 0, false, true, ""},
			{"from_gid", "int", 0, false, false, ""},
			{"to_gid", "int", 0, false, false, ""},
			bcol("two_way", false), // If a user no longer meets the requirements for this promotion then they will be demoted if this flag is set

			// Requirements
			{"level", "int", 0, false, false, ""},
			{"minTime", "int", 0, false, false, ""}, // How long someone needs to have been in their current group before being promoted
		},
		[]tK{
			{"pid", "primary", "", false},
		},
	)
}

func patch25(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("users_groups_promotions", tC{"posts", "int", 0, false, false, "0"}, nil))
}

func patch26(scanner *bufio.Scanner) error {
	return createTable("users_blocks", "", "",
		[]tC{
			{"blocker", "int", 0, false, false, ""},
			{"blockedUser", "int", 0, false, false, ""},
		}, nil,
	)
}

func patch27(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("moderation_logs", tC{"extra", "text", 0, false, false, ""}, nil))
	if err != nil {
		return err
	}
	return execStmt(qgen.Builder.AddColumn("administration_logs", tC{"extra", "text", 0, false, false, ""}, nil))
}

func patch28(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("users", tC{"enable_embeds", "int", 0, false, false, "-1"}, nil))
}

// The word counter might run into problems with some languages where words aren't as obviously demarcated, I would advise turning it off in those cases, or if it becomes annoying in general, really.
func WordCount(input string) (count int) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0
	}

	var inSpace bool
	for _, value := range input {
		if unicode.IsSpace(value) || unicode.IsPunct(value) {
			if !inSpace {
				inSpace = true
			}
		} else if inSpace {
			count++
			inSpace = false
		}
	}

	return count + 1
}

func patch29(scanner *bufio.Scanner) error {
	f := func(tbl, idCol string) error {
		return acc().Select(tbl).Cols(idCol + ",content").Each(func(rows *sql.Rows) error {
			var id int
			var content string
			err := rows.Scan(&id, &content)
			if err != nil {
				return err
			}
			_, err = acc().Update(tbl).Set("words=?").Where(idCol+"=?").Exec(WordCount(content), id)
			return err
		})
	}
	err := f("topics", "tid")
	if err != nil {
		return err
	}
	err = f("replies", "rid")
	if err != nil {
		return err
	}

	meta, err := meta.NewDefaultMetaStore(acc())
	if err != nil {
		return err
	}
	err = meta.Set("sched", "recalc")
	if err != nil {
		return err
	}

	fixCols := func(tbls ...string) error {
		for _, tbl := range tbls {
			//err := execStmt(qgen.Builder.RenameColumn(tbl, "ipaddress","ip"))
			err := execStmt(qgen.Builder.ChangeColumn(tbl, "ipaddress", ccol("ip", 200, "''")))
			if err != nil {
				return err
			}
			err = execStmt(qgen.Builder.SetDefaultColumn(tbl, "ip", "varchar", ""))
			if err != nil {
				return err
			}
		}
		return nil
	}
	err = fixCols("topics", "replies", "polls_votes", "users_replies")
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SetDefaultColumn("replies", "lastEdit", "int", "0"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.SetDefaultColumn("replies", "lastEditBy", "int", "0"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.SetDefaultColumn("users_replies", "lastEdit", "int", "0"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.SetDefaultColumn("users_replies", "lastEditBy", "int", "0"))
	if err != nil {
		return err
	}

	return execStmt(qgen.Builder.AddColumn("activity_stream", tC{"extra", "varchar", 200, false, false, "''"}, nil))

}

func patch30(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("users_groups_promotions", tC{"registeredFor", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}
	return execStmt(qgen.Builder.SetDefaultColumn("users", "last_ip", "varchar", ""))
}

func patch31(scanner *bufio.Scanner) (e error) {
	addKey := func(tbl, col string, tk qgen.DBTableKey) error {
		/*err := execStmt(qgen.Builder.RemoveIndex(tbl, col))
		if err != nil {
			return err
		}*/
		return execStmt(qgen.Builder.AddKey(tbl, col, tk))
	}
	err := addKey("topics", "title", tK{"title", "fulltext", "", false})
	if err != nil {
		return err
	}
	err = addKey("topics", "content", tK{"content", "fulltext", "", false})
	if err != nil {
		return err
	}
	return addKey("replies", "content", tK{"content", "fulltext", "", false})
}

func createTable(tbl, charset, collation string, cols []tC, keys []tK) error {
	err := execStmt(qgen.Builder.DropTable(tbl))
	if err != nil {
		return err
	}
	return execStmt(qgen.Builder.CreateTable(tbl, charset, collation, cols, keys))
}

func patch32(scanner *bufio.Scanner) error {
	return createTable("perfchunks", "", "",
		[]tC{
			{"low", "int", 0, false, false, "0"},
			{"high", "int", 0, false, false, "0"},
			{"avg", "int", 0, false, false, "0"},
			{"createdAt", "datetime", 0, false, false, ""},
		}, nil,
	)
}

func patch33(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("viewchunks", tC{"avg", "int", 0, false, false, "0"}, nil))
}

func patch34(scanner *bufio.Scanner) error {
	/*err := createTable("tables", "", "",
		[]tC{
			{"id", "int", 0, false, true, ""},
			ccol("name", 200, ""),
		},
		[]tK{
			{"id", "primary", "", false},
			{"name", "unique", "", false},
		},
	)
	if err != nil {
		return err
	}
	insert := func(tbl, cols, fields string) {
		if err != nil {
			return
		}
		err = execStmt(qgen.Builder.SimpleInsert(tbl, cols, fields))
	}
	insert("tables", "name", "forums")
	insert("tables", "name", "topics")
	insert("tables", "name", "replies")
	// ! Hold onto freplies for a while longer
	insert("tables", "name", "freplies")*/
	/*err := execStmt(qgen.Builder.AddColumn("topics", tC{"attachCount", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}*/
	overwriteColumn := func(tbl, col string, tc qgen.DBTableColumn) error {
		/*e := execStmt(qgen.Builder.DropColumn(tbl, col))
		if e != nil {
			return e
		}*/
		return execStmt(qgen.Builder.AddColumn(tbl, tc, nil))
	}
	err := overwriteColumn("users", "profile_comments", tC{"profile_comments", "int", 0, false, false, "0"})
	if err != nil {
		return err
	}
	err = overwriteColumn("users", "who_can_convo", tC{"who_can_convo", "int", 0, false, false, "0"})
	if err != nil {
		return err
	}

	setDefault := func(tbl, col, typ, val string) {
		if err != nil {
			return
		}
		err = execStmt(qgen.Builder.SetDefaultColumn(tbl, col, typ, val))
	}
	setDefault("users_groups", "permissions", "text", "{}")
	setDefault("users_groups", "plugin_perms", "text", "{}")
	setDefault("forums_permissions", "permissions", "text", "{}")
	setDefault("topics", "content", "text", "")
	setDefault("topics", "parsed_content", "text", "")
	setDefault("replies", "content", "text", "")
	setDefault("replies", "parsed_content", "text", "")
	//setDefault("revisions", "content", "text", "")
	setDefault("users_replies", "content", "text", "")
	setDefault("users_replies", "parsed_content", "text", "")
	setDefault("pages", "body", "text", "")
	setDefault("pages", "allowedGroups", "text", "")
	setDefault("moderation_logs", "extra", "text", "")
	setDefault("administration_logs", "extra", "text", "")
	if err != nil {
		return err
	}

	return nil
}

func patch35(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("topics", tC{"weekEvenViews", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}
	return execStmt(qgen.Builder.AddColumn("topics", tC{"weekOddViews", "int", 0, false, false, "0"}, nil))
}
