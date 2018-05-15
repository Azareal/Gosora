package main

import (
	"bufio"
	"strconv"

	"../query_gen/lib"
)

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
