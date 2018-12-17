package main

import (
	"bufio"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
)

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
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"mid", "int", 0, false, true, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"mid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("menu_items", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"miid", "int", 0, false, true, ""},
			qgen.DBTableColumn{"mid", "int", 0, false, false, ""},
			qgen.DBTableColumn{"name", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"htmlID", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"cssClass", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"position", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"path", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"aria", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"tooltip", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"tmplName", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"order", "int", 0, false, false, "0"},

			qgen.DBTableColumn{"guestOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"memberOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"staffOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"adminOnly", "boolean", 0, false, false, "0"},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"miid", "primary"},
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
	err := execStmt(qgen.Builder.CreateTable("registration_logs", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"rlid", "int", 0, false, true, ""},
			qgen.DBTableColumn{"username", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"email", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"failureReason", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			qgen.DBTableColumn{"ipaddress", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"rlid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
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
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"pid", "int", 0, false, true, ""},
			qgen.DBTableColumn{"name", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"title", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"body", "text", 0, false, false, ""},
			qgen.DBTableColumn{"allowedGroups", "text", 0, false, false, ""},
			qgen.DBTableColumn{"menuID", "int", 0, false, false, "-1"},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"pid", "primary"},
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
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"uid", "int", 0, false, false, ""},
			qgen.DBTableColumn{"secret", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"scratch1", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch2", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch3", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch4", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch5", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch6", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch7", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"scratch8", "varchar", 50, false, false, ""},
			qgen.DBTableColumn{"createdAt", "createdAt", 0, false, false, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"uid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
}

func patch6(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.SimpleInsert("settings", "name, content, type", "'rapid_loading','1','bool'"))
	if err != nil {
		return err
	}

	return nil
}

func patch7(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.CreateTable("users_avatar_queue", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"uid", "int", 0, false, false, ""}, // TODO: Make this a foreign key
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"uid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
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
	err = execStmt(qgen.Builder.CreateTable("updates", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"dbVersion", "int", 0, false, false, "0"},
		},
		[]qgen.DBTableKey{},
	))
	if err != nil {
		return err
	}

	return nil
}

func patch9(scanner *bufio.Scanner) error {
	// Table "updates" might not exist due to the installer, so drop it and remake it if so
	err := patch8(scanner)
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("login_logs", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"lid", "int", 0, false, true, ""},
			qgen.DBTableColumn{"uid", "int", 0, false, false, ""},
			qgen.DBTableColumn{"success", "bool", 0, false, false, "0"}, // Did this attempt succeed?
			qgen.DBTableColumn{"ipaddress", "varchar", 200, false, false, ""},
			qgen.DBTableColumn{"doneAt", "createdAt", 0, false, false, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"lid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	return nil
}
