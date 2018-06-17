package main

import (
	"bufio"
	"strconv"

	"../query_gen/lib"
)

func init() {
	addPatch(0, patch0)
	addPatch(1, patch1)
	addPatch(2, patch2)
	addPatch(3, patch3)
	addPatch(4, patch4)
	addPatch(5, patch5)
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
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	var replaceTextWhere = func(replaceThis string, withThis string) error {
		return execStmt(qgen.Builder.SimpleUpdate("viewchunks", "route = '"+withThis+"'", "route = '"+replaceThis+"'"))
	}

	err := replaceTextWhere("routeAccountEditCriticalSubmit", "routes.AccountEditCriticalSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditAvatar", "routes.AccountEditAvatar")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditAvatarSubmit", "routes.AccountEditAvatarSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditUsername", "routes.AccountEditUsername")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditUsernameSubmit", "routes.AccountEditUsernameSubmit")
	if err != nil {
		return err
	}

	return nil
}

func patch2(scanner *bufio.Scanner) error {
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	var replaceTextWhere = func(replaceThis string, withThis string) error {
		return execStmt(qgen.Builder.SimpleUpdate("viewchunks", "route = '"+withThis+"'", "route = '"+replaceThis+"'"))
	}

	err := replaceTextWhere("routeLogout", "routes.AccountLogout")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeShowAttachment", "routes.ShowAttachment")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeChangeTheme", "routes.ChangeTheme")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeProfileReplyCreateSubmit", "routes.ProfileReplyCreateSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeLikeTopicSubmit", "routes.LikeTopicSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeReplyLikeSubmit", "routes.ReplyLikeSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeDynamic", "routes.DynamicRoute")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeUploads", "routes.UploadedFile")
	if err != nil {
		return err
	}

	err = replaceTextWhere("BadRoute", "routes.BadRoute")
	if err != nil {
		return err
	}

	return nil
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
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	var replaceTextWhere = func(replaceThis string, withThis string) error {
		return execStmt(qgen.Builder.SimpleUpdate("viewchunks", "route = '"+withThis+"'", "route = '"+replaceThis+"'"))
	}

	err := replaceTextWhere("routeReportSubmit", "routes.ReportSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditEmail", "routes.AccountEditEmail")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routeAccountEditEmailTokenSubmit", "routes.AccountEditEmailTokenSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelLogsRegs", "panel.LogsRegs")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelLogsMod", "panel.LogsMod")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelLogsAdmin", "panel.LogsAdmin")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelDebug", "panel.Debug")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsViews", "panel.AnalyticsViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsRouteViews", "panel.AnalyticsRouteViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsRouteViews", "panel.AnalyticsRouteViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsAgentViews", "panel.AnalyticsAgentViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsForumViews", "panel.AnalyticsForumViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsSystemViews", "panel.AnalyticsSystemViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsLanguageViews", "panel.AnalyticsLanguageViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsReferrerViews", "panel.AnalyticsReferrerViews")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsTopics", "panel.AnalyticsTopics")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsPosts", "panel.AnalyticsPosts")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsForums", "panel.AnalyticsForums")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsRoutes", "panel.AnalyticsRoutes")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsAgents", "panel.AnalyticsAgents")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsSystems", "panel.AnalyticsSystems")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsLanguages", "panel.AnalyticsLanguages")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelAnalyticsReferrers", "panel.AnalyticsReferrers")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelSettings", "panel.Settings")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelSettingEdit", "panel.SettingEdit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelSettingEditSubmit", "panel.SettingEditSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForums", "panel.Forums")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsCreateSubmit", "panel.ForumsCreateSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsDelete", "panel.ForumsDelete")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsDeleteSubmit", "panel.ForumsDeleteSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsEdit", "panel.ForumsEdit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsEditSubmit", "panel.ForumsEditSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsEditPermsSubmit", "panel.ForumsEditPermsSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsEditPermsAdvance", "panel.ForumsEditPermsAdvance")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelForumsEditPermsAdvanceSubmit", "panel.ForumsEditPermsAdvanceSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelBackups", "panel.Backups")
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
	// ! Don't reuse this function blindly, it doesn't escape apostrophes
	var replaceTextWhere = func(replaceThis string, withThis string) error {
		return execStmt(qgen.Builder.SimpleUpdate("viewchunks", "route = '"+withThis+"'", "route = '"+replaceThis+"'"))
	}

	err := replaceTextWhere("routePanelUsers", "panel.Users")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelUsersEdit", "panel.UsersEdit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routePanelUsersEditSubmit", "panel.UsersEditSubmit")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routes.AccountEditCritical", "routes.AccountEditPassword")
	if err != nil {
		return err
	}

	err = replaceTextWhere("routes.AccountEditCriticalSubmit", "routes.AccountEditPasswordSubmit")
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
