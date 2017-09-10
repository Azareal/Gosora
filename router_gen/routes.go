package main

type Route struct {
	Name   string
	Path   string
	Before string
	Vars   []string
}

type RouteGroup struct {
	Path   string
	Routes []Route
}

func addRoute(fname string, path string, before string, vars ...string) {
	route_list = append(route_list, Route{fname, path, before, vars})
}

func addRouteGroup(path string, routes ...Route) {
	route_groups = append(route_groups, RouteGroup{path, routes})
}

func routes() {
	//addRoute("default_route","","")
	addRoute("route_api", "/api/", "")
	///addRoute("route_static","/static/","req.URL.Path += extra_data")
	addRoute("route_overview", "/overview/", "")
	//addRoute("route_custom_page","/pages/",""/*,"&extra_data"*/)
	addRoute("route_forums", "/forums/", "" /*,"&forums"*/)
	addRoute("route_forum", "/forum/", "", "extra_data")
	//addRoute("route_topic_create","/topics/create/","","extra_data")
	//addRoute("route_topics","/topics/",""/*,"&groups","&forums"*/)
	addRoute("route_change_theme", "/theme/", "")

	addRouteGroup("/report/",
		Route{"route_report_submit", "/report/submit/", "", []string{"extra_data"}},
	)

	addRouteGroup("/topics/",
		Route{"route_topics", "/topics/", "", []string{}},
		Route{"route_topic_create", "/topics/create/", "", []string{"extra_data"}},
	)

	// The Control Panel
	addRouteGroup("/panel/",
		Route{"route_panel", "/panel/", "", []string{}},
		Route{"route_panel_forums", "/panel/forums/", "", []string{}},
		Route{"route_panel_forums_create_submit", "/panel/forums/create/", "", []string{}},
		Route{"route_panel_forums_delete", "/panel/forums/delete/", "", []string{"extra_data"}},
		Route{"route_panel_forums_delete_submit", "/panel/forums/delete/submit/", "", []string{"extra_data"}},
		Route{"route_panel_forums_edit", "/panel/forums/edit/", "", []string{"extra_data"}},
		Route{"route_panel_forums_edit_submit", "/panel/forums/edit/submit/", "", []string{"extra_data"}},
		Route{"route_panel_forums_edit_perms_submit", "/panel/forums/edit/perms/submit/", "", []string{"extra_data"}},

		Route{"route_panel_settings", "/panel/settings/", "", []string{}},
		Route{"route_panel_setting", "/panel/settings/edit/", "", []string{"extra_data"}},
		Route{"route_panel_setting_edit", "/panel/settings/edit/submit/", "", []string{"extra_data"}},

		Route{"route_panel_word_filters", "/panel/settings/word-filters/", "", []string{}},
		Route{"route_panel_word_filters_create", "/panel/settings/word-filters/create/", "", []string{}},
		Route{"route_panel_word_filters_edit", "/panel/settings/word-filters/edit/", "", []string{"extra_data"}},
		Route{"route_panel_word_filters_edit_submit", "/panel/settings/word-filters/edit/submit/", "", []string{"extra_data"}},
		Route{"route_panel_word_filters_delete_submit", "/panel/settings/word-filters/delete/submit/", "", []string{"extra_data"}},

		Route{"route_panel_themes", "/panel/themes/", "", []string{}},
		Route{"route_panel_themes_set_default", "/panel/themes/default/", "", []string{"extra_data"}},

		Route{"route_panel_plugins", "/panel/plugins/", "", []string{}},
		Route{"route_panel_plugins_activate", "/panel/plugins/activate/", "", []string{"extra_data"}},
		Route{"route_panel_plugins_deactivate", "/panel/plugins/deactivate/", "", []string{"extra_data"}},
		Route{"route_panel_plugins_install", "/panel/plugins/install/", "", []string{"extra_data"}},

		Route{"route_panel_users", "/panel/users/", "", []string{}},
		Route{"route_panel_users_edit", "/panel/users/edit/", "", []string{"extra_data"}},
		Route{"route_panel_users_edit_submit", "/panel/users/edit/submit/", "", []string{"extra_data"}},

		Route{"route_panel_groups", "/panel/groups/", "", []string{}},
		Route{"route_panel_groups_edit", "/panel/groups/edit/", "", []string{"extra_data"}},
		Route{"route_panel_groups_edit_perms", "/panel/groups/edit/perms/", "", []string{"extra_data"}},
		Route{"route_panel_groups_edit_submit", "/panel/groups/edit/submit/", "", []string{"extra_data"}},
		Route{"route_panel_groups_edit_perms_submit", "/panel/groups/edit/perms/submit/", "", []string{"extra_data"}},
		Route{"route_panel_groups_create_submit", "/panel/groups/create/", "", []string{}},

		Route{"route_panel_logs_mod", "/panel/logs/mod/", "", []string{}},
		Route{"route_panel_debug", "/panel/debug/", "", []string{}},
	)
}
