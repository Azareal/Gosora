package main

// TODO: How should we handle headerLite and headerVar?
func routes() {
	//addRoute("default_route","","")
	addRoute(Route("routeAPI", "/api/"))
	///addRoute("routeStatic","/static/","req.URL.Path += extra_data")
	addRoute(Route("routeOverview", "/overview/"))
	//addRoute("routeCustomPage","/pages/",""/*,"&extra_data"*/)
	addRoute(Route("routeForums", "/forums/" /*,"&forums"*/))
	addRoute(Route("routeForum", "/forum/", "extra_data"))
	//addRoute("routeTopicCreate","/topics/create/","","extra_data")
	//addRoute("routeTopics","/topics/",""/*,"&groups","&forums"*/)
	addRoute(Route("routeChangeTheme", "/theme/"))
	addRoute(Route("routeShowAttachment", "/attachs/", "extra_data"))

	// TODO: Reduce the number of Befores. With a new method, perhaps?
	reportGroup := newRouteGroup("/report/",
		Route("routeReportSubmit", "/report/submit/", "extra_data"),
	).Before("MemberOnly", "NoBanned", "NoSessionMismatch")
	addRouteGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		Route("routeTopics", "/topics/"),
		Route("routeTopicCreate", "/topics/create/", "extra_data").Before("MemberOnly"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
	buildUserRoutes()
}

// TODO: Test the email token route
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/")
	userGroup.Routes(
		Route("routeProfile", "/user/").LitBefore("req.URL.Path += extra_data"),
		Route("routeAccountEditCritical", "/user/edit/critical/"),
		Route("routeAccountEditCriticalSubmit", "/user/edit/critical/submit/").Before("NoSessionMismatch"), // TODO: Full test this
		Route("routeAccountEditAvatar", "/user/edit/avatar/"),
		Route("routeAccountEditAvatarSubmit", "/user/edit/avatar/submit/"),
		Route("routeAccountEditUsername", "/user/edit/username/"),
		Route("routeAccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		Route("routeAccountEditEmail", "/user/edit/email/"),
		Route("routeAccountEditEmailTokenSubmit", "/user/edit/token/", "extra_data"),
	).Not("/user/").Before("MemberOnly")
	addRouteGroup(userGroup)

	// TODO: Auto test and manual test these routes
	userGroup = newRouteGroup("/users/").Before("MemberOnly")
	userGroup.Routes(
		Route("routeBanSubmit", "/users/ban/submit/"),
		Route("routeUnban", "/users/unban/"),
		Route("routeActivate", "/users/activate/"),
		Route("routeIps", "/users/ips/"),
	).Not("/users/ips/").Before("NoSessionMismatch")
	addRouteGroup(userGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		Route("routePanel", "/panel/"),
		Route("routePanelForums", "/panel/forums/"),
		Route("routePanelForumsCreateSubmit", "/panel/forums/create/").Before("NoSessionMismatch"),
		Route("routePanelForumsDelete", "/panel/forums/delete/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelForumsEdit", "/panel/forums/edit/", "extra_data"),
		Route("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extra_data").Before("NoSessionMismatch"),

		Route("routePanelSettings", "/panel/settings/"),
		Route("routePanelSetting", "/panel/settings/edit/", "extra_data"),
		Route("routePanelSettingEdit", "/panel/settings/edit/submit/", "extra_data").Before("NoSessionMismatch"),

		Route("routePanelWordFilters", "/panel/settings/word-filters/"),
		Route("routePanelWordFiltersCreate", "/panel/settings/word-filters/create/").Before("ParseForm"),
		Route("routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "extra_data"),
		Route("routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "extra_data").Before("ParseForm"),
		Route("routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "extra_data").Before("ParseForm"),

		Route("routePanelThemes", "/panel/themes/"),
		Route("routePanelThemesSetDefault", "/panel/themes/default/", "extra_data").Before("NoSessionMismatch"),

		Route("routePanelPlugins", "/panel/plugins/"),
		Route("routePanelPluginsActivate", "/panel/plugins/activate/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelPluginsInstall", "/panel/plugins/install/", "extra_data").Before("NoSessionMismatch"),

		Route("routePanelUsers", "/panel/users/"),
		Route("routePanelUsersEdit", "/panel/users/edit/", "extra_data"),
		Route("routePanelUsersEditSubmit", "/panel/users/edit/submit/", "extra_data").Before("NoSessionMismatch"),

		Route("routePanelGroups", "/panel/groups/"),
		Route("routePanelGroupsEdit", "/panel/groups/edit/", "extra_data"),
		Route("routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "extra_data"),
		Route("routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "extra_data").Before("NoSessionMismatch"),
		Route("routePanelGroupsCreateSubmit", "/panel/groups/create/").Before("NoSessionMismatch"),

		Route("routePanelBackups", "/panel/backups/", "extra_data"),
		Route("routePanelLogsMod", "/panel/logs/mod/"),
		Route("routePanelDebug", "/panel/debug/").Before("AdminOnly"),
	)
	addRouteGroup(panelGroup)
}
