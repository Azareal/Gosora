package main

// TODO: How should we handle headerLite and headerVar?
func routes() {
	addRoute(View("routeAPI", "/api/"))
	addRoute(View("routeOverview", "/overview/"))
	//addRoute("routeCustomPage","/pages/",""/*,"&extra_data"*/)
	addRoute(View("routeForums", "/forums/" /*,"&forums"*/))
	addRoute(View("routeForum", "/forum/", "extra_data"))
	addRoute(AnonAction("routeChangeTheme", "/theme/"))
	addRoute(
		View("routeShowAttachment", "/attachs/", "extra_data").Before("ParseForm"),
	)

	// TODO: Reduce the number of Befores. With a new method, perhaps?
	reportGroup := newRouteGroup("/report/",
		Action("routeReportSubmit", "/report/submit/", "extra_data"),
	).Before("NoBanned")
	addRouteGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		View("routeTopics", "/topics/"),
		MemberView("routeTopicCreate", "/topics/create/", "extra_data"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
	buildUserRoutes()
}

// TODO: Test the email token route
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/")
	userGroup.Routes(
		View("routeProfile", "/user/").LitBefore("req.URL.Path += extra_data"),
		MemberView("routeAccountEditCritical", "/user/edit/critical/"),
		Action("routeAccountEditCriticalSubmit", "/user/edit/critical/submit/"), // TODO: Full test this
		MemberView("routeAccountEditAvatar", "/user/edit/avatar/"),
		UploadAction("routeAccountEditAvatarSubmit", "/user/edit/avatar/submit/"),
		MemberView("routeAccountEditUsername", "/user/edit/username/"),
		Action("routeAccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		MemberView("routeAccountEditEmail", "/user/edit/email/"),
		Action("routeAccountEditEmailTokenSubmit", "/user/edit/token/", "extra_data"),
	)
	addRouteGroup(userGroup)

	// TODO: Auto test and manual test these routes
	userGroup = newRouteGroup("/users/")
	userGroup.Routes(
		Action("routeBanSubmit", "/users/ban/submit/"),
		Action("routeUnban", "/users/unban/"),
		Action("routeActivate", "/users/activate/"),
		MemberView("routeIps", "/users/ips/"),
	)
	addRouteGroup(userGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		View("routePanel", "/panel/"),
		View("routePanelForums", "/panel/forums/"),
		Action("routePanelForumsCreateSubmit", "/panel/forums/create/"),
		Action("routePanelForumsDelete", "/panel/forums/delete/", "extra_data"),
		Action("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "extra_data"),
		View("routePanelForumsEdit", "/panel/forums/edit/", "extra_data"),
		Action("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "extra_data"),
		Action("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extra_data"),

		View("routePanelSettings", "/panel/settings/"),
		View("routePanelSettingEdit", "/panel/settings/edit/", "extra_data"),
		Action("routePanelSettingEditSubmit", "/panel/settings/edit/submit/", "extra_data"),

		View("routePanelWordFilters", "/panel/settings/word-filters/"),
		Action("routePanelWordFiltersCreate", "/panel/settings/word-filters/create/"),
		View("routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "extra_data"),
		Action("routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "extra_data"),
		Action("routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "extra_data"),

		View("routePanelThemes", "/panel/themes/"),
		Action("routePanelThemesSetDefault", "/panel/themes/default/", "extra_data"),

		View("routePanelPlugins", "/panel/plugins/"),
		Action("routePanelPluginsActivate", "/panel/plugins/activate/", "extra_data"),
		Action("routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "extra_data"),
		Action("routePanelPluginsInstall", "/panel/plugins/install/", "extra_data"),

		View("routePanelUsers", "/panel/users/"),
		View("routePanelUsersEdit", "/panel/users/edit/", "extra_data"),
		Action("routePanelUsersEditSubmit", "/panel/users/edit/submit/", "extra_data"),

		View("routePanelGroups", "/panel/groups/"),
		View("routePanelGroupsEdit", "/panel/groups/edit/", "extra_data"),
		View("routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "extra_data"),
		Action("routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "extra_data"),
		Action("routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "extra_data"),
		Action("routePanelGroupsCreateSubmit", "/panel/groups/create/"),

		View("routePanelBackups", "/panel/backups/", "extra_data"),
		View("routePanelLogsMod", "/panel/logs/mod/"),
		View("routePanelDebug", "/panel/debug/").Before("AdminOnly"),
	)
	addRouteGroup(panelGroup)
}
