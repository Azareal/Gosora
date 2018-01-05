package main

// TODO: How should we handle headerLite and headerVar?
func routes() {
	addRoute(View("routeAPI", "/api/"))
	addRoute(View("routeOverview", "/overview/"))
	addRoute(View("routeCustomPage", "/pages/", "extraData"))
	addRoute(View("routeForums", "/forums/" /*,"&forums"*/))
	addRoute(View("routeForum", "/forum/", "extraData"))
	addRoute(AnonAction("routeChangeTheme", "/theme/"))
	addRoute(
		View("routeShowAttachment", "/attachs/", "extraData").Before("ParseForm"),
	)

	// TODO: Reduce the number of Befores. With a new method, perhaps?
	reportGroup := newRouteGroup("/report/",
		Action("routeReportSubmit", "/report/submit/", "extraData"),
	).Before("NoBanned")
	addRouteGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		View("routeTopics", "/topics/"),
		MemberView("routeTopicCreate", "/topics/create/", "extraData"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
	buildUserRoutes()
}

// TODO: Test the email token route
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/")
	userGroup.Routes(
		View("routeProfile", "/user/").LitBefore("req.URL.Path += extraData"),
		MemberView("routeAccountEditCritical", "/user/edit/critical/"),
		Action("routeAccountEditCriticalSubmit", "/user/edit/critical/submit/"), // TODO: Full test this
		MemberView("routeAccountEditAvatar", "/user/edit/avatar/"),
		UploadAction("routeAccountEditAvatarSubmit", "/user/edit/avatar/submit/"),
		MemberView("routeAccountEditUsername", "/user/edit/username/"),
		Action("routeAccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		MemberView("routeAccountEditEmail", "/user/edit/email/"),
		Action("routeAccountEditEmailTokenSubmit", "/user/edit/token/", "extraData"),
	)
	addRouteGroup(userGroup)

	// TODO: Auto test and manual test these routes
	userGroup = newRouteGroup("/users/")
	userGroup.Routes(
		Action("routeBanSubmit", "/users/ban/submit/", "extraData"),
		Action("routeUnban", "/users/unban/", "extraData"),
		Action("routeActivate", "/users/activate/", "extraData"),
		MemberView("routeIps", "/users/ips/"), // TODO: .Perms("ViewIPs")?
	)
	addRouteGroup(userGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		View("routePanel", "/panel/"),
		View("routePanelForums", "/panel/forums/"),
		Action("routePanelForumsCreateSubmit", "/panel/forums/create/"),
		Action("routePanelForumsDelete", "/panel/forums/delete/", "extraData"),
		Action("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "extraData"),
		View("routePanelForumsEdit", "/panel/forums/edit/", "extraData"),
		Action("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "extraData"),
		Action("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extraData"),

		View("routePanelSettings", "/panel/settings/"),
		View("routePanelSettingEdit", "/panel/settings/edit/", "extraData"),
		Action("routePanelSettingEditSubmit", "/panel/settings/edit/submit/", "extraData"),

		View("routePanelWordFilters", "/panel/settings/word-filters/"),
		Action("routePanelWordFiltersCreate", "/panel/settings/word-filters/create/"),
		View("routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "extraData"),
		Action("routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "extraData"),
		Action("routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "extraData"),

		View("routePanelThemes", "/panel/themes/"),
		Action("routePanelThemesSetDefault", "/panel/themes/default/", "extraData"),

		View("routePanelPlugins", "/panel/plugins/"),
		Action("routePanelPluginsActivate", "/panel/plugins/activate/", "extraData"),
		Action("routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "extraData"),
		Action("routePanelPluginsInstall", "/panel/plugins/install/", "extraData"),

		View("routePanelUsers", "/panel/users/"),
		View("routePanelUsersEdit", "/panel/users/edit/", "extraData"),
		Action("routePanelUsersEditSubmit", "/panel/users/edit/submit/", "extraData"),

		View("routePanelAnalyticsViews", "/panel/analytics/views/"),
		View("routePanelAnalyticsRoutes", "/panel/analytics/routes/"),
		View("routePanelAnalyticsRouteViews", "/panel/analytics/route/", "extraData"),

		View("routePanelGroups", "/panel/groups/"),
		View("routePanelGroupsEdit", "/panel/groups/edit/", "extraData"),
		View("routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "extraData"),
		Action("routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "extraData"),
		Action("routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "extraData"),
		Action("routePanelGroupsCreateSubmit", "/panel/groups/create/"),

		View("routePanelBackups", "/panel/backups/", "extraData").Before("SuperAdminOnly"), // TODO: Test
		View("routePanelLogsMod", "/panel/logs/mod/"),
		View("routePanelDebug", "/panel/debug/").Before("AdminOnly"),
	)
	addRouteGroup(panelGroup)
}
