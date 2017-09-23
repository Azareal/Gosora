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
	addRoute("routeAPI", "/api/", "")
	///addRoute("routeStatic","/static/","req.URL.Path += extra_data")
	addRoute("routeOverview", "/overview/", "")
	//addRoute("routeCustomPage","/pages/",""/*,"&extra_data"*/)
	addRoute("routeForums", "/forums/", "" /*,"&forums"*/)
	addRoute("routeForum", "/forum/", "", "extra_data")
	//addRoute("routeTopicCreate","/topics/create/","","extra_data")
	//addRoute("routeTopics","/topics/",""/*,"&groups","&forums"*/)
	addRoute("routeChangeTheme", "/theme/", "")

	addRouteGroup("/report/",
		Route{"routeReportSubmit", "/report/submit/", "", []string{"extra_data"}},
	)

	addRouteGroup("/topics/",
		Route{"routeTopics", "/topics/", "", []string{}},
		Route{"routeTopicCreate", "/topics/create/", "", []string{"extra_data"}},
	)

	// The Control Panel
	addRouteGroup("/panel/",
		Route{"routePanel", "/panel/", "", []string{}},
		Route{"routePanelForums", "/panel/forums/", "", []string{}},
		Route{"routePanelForumsCreateSubmit", "/panel/forums/create/", "", []string{}},
		Route{"routePanelForumsDelete", "/panel/forums/delete/", "", []string{"extra_data"}},
		Route{"routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "", []string{"extra_data"}},
		Route{"routePanelForumsEdit", "/panel/forums/edit/", "", []string{"extra_data"}},
		Route{"routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "", []string{"extra_data"}},
		Route{"routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "", []string{"extra_data"}},

		Route{"routePanelSettings", "/panel/settings/", "", []string{}},
		Route{"routePanelSetting", "/panel/settings/edit/", "", []string{"extra_data"}},
		Route{"routePanelSettingEdit", "/panel/settings/edit/submit/", "", []string{"extra_data"}},

		Route{"routePanelWordFilters", "/panel/settings/word-filters/", "", []string{}},
		Route{"routePanelWordFiltersCreate", "/panel/settings/word-filters/create/", "", []string{}},
		Route{"routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "", []string{"extra_data"}},
		Route{"routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "", []string{"extra_data"}},
		Route{"routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "", []string{"extra_data"}},

		Route{"routePanelThemes", "/panel/themes/", "", []string{}},
		Route{"routePanelThemesSetDefault", "/panel/themes/default/", "", []string{"extra_data"}},

		Route{"routePanelPlugins", "/panel/plugins/", "", []string{}},
		Route{"routePanelPluginsActivate", "/panel/plugins/activate/", "", []string{"extra_data"}},
		Route{"routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "", []string{"extra_data"}},
		Route{"routePanelPluginsInstall", "/panel/plugins/install/", "", []string{"extra_data"}},

		Route{"routePanelUsers", "/panel/users/", "", []string{}},
		Route{"routePanelUsersEdit", "/panel/users/edit/", "", []string{"extra_data"}},
		Route{"routePanelUsersEditSubmit", "/panel/users/edit/submit/", "", []string{"extra_data"}},

		Route{"routePanelGroups", "/panel/groups/", "", []string{}},
		Route{"routePanelGroupsEdit", "/panel/groups/edit/", "", []string{"extra_data"}},
		Route{"routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "", []string{"extra_data"}},
		Route{"routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "", []string{"extra_data"}},
		Route{"routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "", []string{"extra_data"}},
		Route{"routePanelGroupsCreateSubmit", "/panel/groups/create/", "", []string{}},

		Route{"routePanelBackups", "/panel/backups/", "", []string{"extra_data"}},
		Route{"routePanelLogsMod", "/panel/logs/mod/", "", []string{}},
		Route{"routePanelDebug", "/panel/debug/", "", []string{}},
	)
}
