package main

import (
	"bufio"
	"database/sql"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
)

type tblColumn = qgen.DBTableColumn
type tblKey = qgen.DBTableKey

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
}

func patch0(scanner *bufio.Scanner) (err error) {
	err = execStmt(qgen.Builder.DropTable("menus"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.DropTable("menu_items"))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("menus", "", "",
		[]tblColumn{
			tblColumn{"mid", "int", 0, false, true, ""},
		},
		[]tblKey{
			tblKey{"mid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("menu_items", "", "",
		[]tblColumn{
			tblColumn{"miid", "int", 0, false, true, ""},
			tblColumn{"mid", "int", 0, false, false, ""},
			tblColumn{"name", "varchar", 200, false, false, ""},
			tblColumn{"htmlID", "varchar", 200, false, false, "''"},
			tblColumn{"cssClass", "varchar", 200, false, false, "''"},
			tblColumn{"position", "varchar", 100, false, false, ""},
			tblColumn{"path", "varchar", 200, false, false, "''"},
			tblColumn{"aria", "varchar", 200, false, false, "''"},
			tblColumn{"tooltip", "varchar", 200, false, false, "''"},
			tblColumn{"tmplName", "varchar", 200, false, false, "''"},
			tblColumn{"order", "int", 0, false, false, "0"},

			tblColumn{"guestOnly", "boolean", 0, false, false, "0"},
			tblColumn{"memberOnly", "boolean", 0, false, false, "0"},
			tblColumn{"staffOnly", "boolean", 0, false, false, "0"},
			tblColumn{"adminOnly", "boolean", 0, false, false, "0"},
		},
		[]tblKey{
			tblKey{"miid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleInsert("menus", "", ""))
	if err != nil {
		return err
	}

	var order int
	var mOrder = "mid, name, htmlID, cssClass, position, path, aria, tooltip, guestOnly, memberOnly, staffOnly, adminOnly"
	var addMenuItem = func(data map[string]interface{}) error {
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
	var routes = map[string]string{
		"routeAccountEditCriticalSubmit": "routes.AccountEditCriticalSubmit",
		"routeAccountEditAvatar":         "routes.AccountEditAvatar",
		"routeAccountEditAvatarSubmit":   "routes.AccountEditAvatarSubmit",
		"routeAccountEditUsername":       "routes.AccountEditUsername",
		"routeAccountEditUsernameSubmit": "routes.AccountEditUsernameSubmit",
	}
	return renameRoutes(routes)
}

func patch2(scanner *bufio.Scanner) error {
	var routes = map[string]string{
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
	return execStmt(qgen.Builder.CreateTable("registration_logs", "", "",
		[]tblColumn{
			tblColumn{"rlid", "int", 0, false, true, ""},
			tblColumn{"username", "varchar", 100, false, false, ""},
			tblColumn{"email", "varchar", 100, false, false, ""},
			tblColumn{"failureReason", "varchar", 100, false, false, ""},
			tblColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"rlid", "primary"},
		},
	))
}

func patch4(scanner *bufio.Scanner) error {
	var routes = map[string]string{
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

	err = execStmt(qgen.Builder.CreateTable("pages", "utf8mb4", "utf8mb4_general_ci",
		[]tblColumn{
			tblColumn{"pid", "int", 0, false, true, ""},
			tblColumn{"name", "varchar", 200, false, false, ""},
			tblColumn{"title", "varchar", 200, false, false, ""},
			tblColumn{"body", "text", 0, false, false, ""},
			tblColumn{"allowedGroups", "text", 0, false, false, ""},
			tblColumn{"menuID", "int", 0, false, false, "-1"},
		},
		[]tblKey{
			tblKey{"pid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
}

func patch5(scanner *bufio.Scanner) error {
	var routes = map[string]string{
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

	err = execStmt(qgen.Builder.SimpleUpdate("menu_items", "path = '/user/edit/'", "path = '/user/edit/critical/'"))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("users_2fa_keys", "utf8mb4", "utf8mb4_general_ci",
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, false, ""},
			tblColumn{"secret", "varchar", 100, false, false, ""},
			tblColumn{"scratch1", "varchar", 50, false, false, ""},
			tblColumn{"scratch2", "varchar", 50, false, false, ""},
			tblColumn{"scratch3", "varchar", 50, false, false, ""},
			tblColumn{"scratch4", "varchar", 50, false, false, ""},
			tblColumn{"scratch5", "varchar", 50, false, false, ""},
			tblColumn{"scratch6", "varchar", 50, false, false, ""},
			tblColumn{"scratch7", "varchar", 50, false, false, ""},
			tblColumn{"scratch8", "varchar", 50, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"uid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
}

func patch6(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.SimpleInsert("settings", "name, content, type", "'rapid_loading','1','bool'"))
}

func patch7(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.CreateTable("users_avatar_queue", "", "",
		[]tblColumn{
			tblColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
		},
		[]tblKey{
			tblKey{"uid", "primary"},
		},
	))
}

func renameRoutes(routes map[string]string) error {
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	var replaceTextWhere = func(replaceThis string, withThis string) error {
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
	var routes = map[string]string{
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

	err = execStmt(qgen.Builder.DropTable("updates"))
	if err != nil {
		return err
	}
	return execStmt(qgen.Builder.CreateTable("updates", "", "",
		[]tblColumn{
			tblColumn{"dbVersion", "int", 0, false, false, "0"},
		},
		[]tblKey{},
	))
}

func patch9(scanner *bufio.Scanner) error {
	// Table "updates" might not exist due to the installer, so drop it and remake it if so
	err := patch8(scanner)
	if err != nil {
		return err
	}

	return execStmt(qgen.Builder.CreateTable("login_logs", "", "",
		[]tblColumn{
			tblColumn{"lid", "int", 0, false, true, ""},
			tblColumn{"uid", "int", 0, false, false, ""},
			tblColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			tblColumn{"ipaddress", "varchar", 200, false, false, ""},
			tblColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]tblKey{
			tblKey{"lid", "primary"},
		},
	))
}

var acc = qgen.NewAcc
var itoa = strconv.Itoa

func patch10(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("topics", tblColumn{"attachCount", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddColumn("topics", tblColumn{"lastReplyID", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	err = acc().Select("topics").Cols("tid").EachInt(func(tid int) error {
		stid := itoa(tid)

		count, err := acc().Count("attachments").Where("originTable = 'topics' and originID = " + stid).Total()
		if err != nil {
			return err
		}

		var hasReply = false
		err = acc().Select("replies").Cols("rid").Where("tid = " + stid).Orderby("rid DESC").Limit("1").EachInt(func(rid int) error {
			hasReply = true
			_, err := acc().Update("topics").Set("lastReplyID = ?, attachCount = ?").Where("tid = "+stid).Exec(rid, count)
			return err
		})
		if err != nil {
			return err
		}
		if !hasReply {
			_, err = acc().Update("topics").Set("attachCount = ?").Where("tid = " + stid).Exec(count)
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
	err := execStmt(qgen.Builder.AddColumn("replies", tblColumn{"attachCount", "int", 0, false, false, "0"}, nil))
	if err != nil {
		return err
	}

	// Attachments for replies got the topicID rather than the replyID for a while in error, so we want to separate these out
	_, err = acc().Update("attachments").Set("originTable = 'freplies'").Where("originTable = 'replies'").Exec()
	if err != nil {
		return err
	}

	// We could probably do something more efficient, but as there shouldn't be too many sites right now, we can probably cheat a little, otherwise it'll take forever to get things done
	return acc().Select("topics").Cols("tid").EachInt(func(tid int) error {
		stid := itoa(tid)

		count, err := acc().Count("attachments").Where("originTable = 'topics' and originID = " + stid).Total()
		if err != nil {
			return err
		}

		_, err = acc().Update("topics").Set("attachCount = ?").Where("tid = " + stid).Exec(count)
		return err
	})

	/*return acc().Select("replies").Cols("rid").EachInt(func(rid int) error {
		srid := itoa(rid)

		count, err := acc().Count("attachments").Where("originTable = 'replies' and originID = " + srid).Total()
		if err != nil {
			return err
		}

		_, err = acc().Update("replies").Set("attachCount = ?").Where("rid = " + srid).Exec(count)
		return err
	})*/
}

func patch12(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddIndex("topics", "parentID", "parentID"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("replies", "tid", "tid"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("polls", "parentID", "parentID"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("likes", "targetItem", "targetItem"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("emails", "uid", "uid"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("attachments", "originID", "originID"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("attachments", "path", "path"))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddIndex("activity_stream_matches", "watcher", "watcher"))
	if err != nil {
		return err
	}
	return nil
}

func patch13(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("widgets", tblColumn{"wid", "int", 0, false, true, ""}, &tblKey{"wid", "primary"}))
	if err != nil {
		return err
	}

	return nil
}

func patch14(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddKey("topics", "title", tblKey{"title", "fulltext"}))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddKey("topics", "content", tblKey{"content", "fulltext"}))
	if err != nil {
		return err
	}
	err = execStmt(qgen.Builder.AddKey("replies", "content", tblKey{"content", "fulltext"}))
	if err != nil {
		return err
	}

	return nil
}

func patch15(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.SimpleInsert("settings", "name, content, type", "'google_site_verify','','html-attribute'"))
}

func patch16(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.CreateTable("password_resets", "", "",
		[]tblColumn{
			tblColumn{"email", "varchar", 200, false, false, ""},
			tblColumn{"uid", "int", 0, false, false, ""},             // TODO: Make this a foreign key
			tblColumn{"validated", "varchar", 200, false, false, ""}, // Token given once the one-use token is consumed, used to prevent multiple people consuming the same one-use token
			tblColumn{"token", "varchar", 200, false, false, ""},
			tblColumn{"createdAt", "createdAt", 0, false, false, ""},
		}, nil,
	))
}

func patch17(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.AddColumn("attachments", tblColumn{"extra", "varchar", 200, false, false, ""}, nil))
	if err != nil {
		return err
	}

	err = acc().Select("topics").Cols("tid, parentID").Where("attachCount > 0").Each(func(rows *sql.Rows) error {
		var tid, parentID int
		err := rows.Scan(&tid, &parentID)
		if err != nil {
			return err
		}
		_, err = acc().Update("attachments").Set("sectionID = ?").Where("originTable = 'topics' AND originID = ?").Exec(parentID, tid)
		return err
	})
	if err != nil {
		return err
	}

	return acc().Select("replies").Cols("rid, tid").Where("attachCount > 0").Each(func(rows *sql.Rows) error {
		var rid, tid, sectionID int
		err := rows.Scan(&rid, &tid)
		if err != nil {
			return err
		}

		err = acc().Select("topics").Cols("parentID").Where("tid = ?").QueryRow(tid).Scan(&sectionID)
		if err != nil {
			return err
		}

		_, err = acc().Update("attachments").Set("sectionID = ?, extra = ?").Where("originTable = 'replies' AND originID = ?").Exec(sectionID, tid, rid)
		return err
	})
}

func patch18(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.AddColumn("forums", tblColumn{"order", "int", 0, false, false, "0"}, nil))
}

func patch19(scanner *bufio.Scanner) error {
	return execStmt(qgen.Builder.CreateTable("memchunks", "", "",
		[]tblColumn{
			tblColumn{"count", "int", 0, false, false, "0"},
			tblColumn{"createdAt", "datetime", 0, false, false, ""},
		}, nil,
	))
}