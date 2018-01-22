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
		MemberView("routes.CreateTopic", "/topics/create/", "extraData"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
	buildUserRoutes()
	buildTopicRoutes()
	buildReplyRoutes()
	buildProfileReplyRoutes()
	buildAccountRoutes()

	addRoute(Special("routeWebsockets", "/ws/"))
}

// TODO: Test the email token route
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/")
	userGroup.Routes(
		View("routeProfile", "/user/").LitBefore("req.URL.Path += extraData"),
		MemberView("routeAccountEditCritical", "/user/edit/critical/"),
		Action("routeAccountEditCriticalSubmit", "/user/edit/critical/submit/"), // TODO: Full test this
		MemberView("routeAccountEditAvatar", "/user/edit/avatar/"),
		UploadAction("routeAccountEditAvatarSubmit", "/user/edit/avatar/submit/").MaxSizeVar("common.Config.MaxRequestSize"),
		MemberView("routeAccountEditUsername", "/user/edit/username/"),
		Action("routeAccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		MemberView("routeAccountEditEmail", "/user/edit/email/"),
		Action("routeAccountEditEmailTokenSubmit", "/user/edit/token/", "extraData"),
	)
	addRouteGroup(userGroup)

	// TODO: Auto test and manual test these routes
	userGroup = newRouteGroup("/users/")
	userGroup.Routes(
		Action("routes.BanUserSubmit", "/users/ban/submit/", "extraData"),
		Action("routes.UnbanUser", "/users/unban/", "extraData"),
		Action("routes.ActivateUser", "/users/activate/", "extraData"),
		MemberView("routes.IPSearch", "/users/ips/"), // TODO: .Perms("ViewIPs")?
	)
	addRouteGroup(userGroup)
}

func buildTopicRoutes() {
	topicGroup := newRouteGroup("/topic/")
	topicGroup.Routes(
		View("routeTopicID", "/topic/", "extraData"),
		UploadAction("routes.CreateTopicSubmit", "/topic/create/submit/").MaxSizeVar("common.Config.MaxRequestSize"),
		Action("routes.EditTopicSubmit", "/topic/edit/submit/", "extraData"),
		Action("routes.DeleteTopicSubmit", "/topic/delete/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.StickTopicSubmit", "/topic/stick/submit/", "extraData"),
		Action("routes.UnstickTopicSubmit", "/topic/unstick/submit/", "extraData"),
		Action("routes.LockTopicSubmit", "/topic/lock/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.UnlockTopicSubmit", "/topic/unlock/submit/", "extraData"),
		Action("routes.MoveTopicSubmit", "/topic/move/submit/", "extraData"),
		Action("routeLikeTopicSubmit", "/topic/like/submit/", "extraData"),
	)
	addRouteGroup(topicGroup)
}

func buildReplyRoutes() {
	//router.HandleFunc("/reply/edit/", routeReplyEdit) // No js fallback
	//router.HandleFunc("/reply/delete/", routeReplyDelete) // No js confirmation page? We could have a confirmation modal for the JS case
	replyGroup := newRouteGroup("/reply/")
	replyGroup.Routes(
		// TODO: Reduce this to 1MB for attachments for each file?
		UploadAction("routeCreateReplySubmit", "/reply/create/").MaxSizeVar("common.Config.MaxRequestSize"), // TODO: Rename the route so it's /reply/create/submit/
		Action("routes.ReplyEditSubmit", "/reply/edit/submit/", "extraData"),
		Action("routes.ReplyDeleteSubmit", "/reply/delete/submit/", "extraData"),
		Action("routeReplyLikeSubmit", "/reply/like/submit/", "extraData"),
	)
	addRouteGroup(replyGroup)
}

// TODO: Move these into /user/?
func buildProfileReplyRoutes() {
	//router.HandleFunc("/user/edit/submit/", routeLogout) // routeLogout? what on earth? o.o
	pReplyGroup := newRouteGroup("/profile/")
	pReplyGroup.Routes(
		Action("routeProfileReplyCreateSubmit", "/profile/reply/create/"), // TODO: Add /submit/ to the end
		Action("routes.ProfileReplyEditSubmit", "/profile/reply/edit/submit/", "extraData"),
		Action("routes.ProfileReplyDeleteSubmit", "/profile/reply/delete/submit/", "extraData"),
	)
	addRouteGroup(pReplyGroup)
}

func buildAccountRoutes() {
	//router.HandleFunc("/accounts/list/", routeLogin) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	accReplyGroup := newRouteGroup("/accounts/")
	accReplyGroup.Routes(
		View("routeLogin", "/accounts/login/"),
		View("routeRegister", "/accounts/create/"),
		Action("routeLogout", "/accounts/logout/"),
		AnonAction("routeLoginSubmit", "/accounts/login/submit/"), // TODO: Guard this with a token, maybe the IP hashed with a rotated key?
		AnonAction("routeRegisterSubmit", "/accounts/create/submit/"),
	)
	addRouteGroup(accReplyGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		View("routePanelDashboard", "/panel/"),
		View("routePanelForums", "/panel/forums/"),
		Action("routePanelForumsCreateSubmit", "/panel/forums/create/"),
		Action("routePanelForumsDelete", "/panel/forums/delete/", "extraData"),
		Action("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "extraData"),
		View("routePanelForumsEdit", "/panel/forums/edit/", "extraData"),
		Action("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "extraData"),
		Action("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extraData"),
		View("routePanelForumsEditPermsAdvance", "/panel/forums/edit/perms/", "extraData"),
		Action("routePanelForumsEditPermsAdvanceSubmit", "/panel/forums/edit/perms/adv/submit/", "extraData"),

		View("routePanelSettings", "/panel/settings/"),
		View("routePanelSettingEdit", "/panel/settings/edit/", "extraData"),
		Action("routePanelSettingEditSubmit", "/panel/settings/edit/submit/", "extraData"),

		View("routePanelWordFilters", "/panel/settings/word-filters/"),
		Action("routePanelWordFiltersCreateSubmit", "/panel/settings/word-filters/create/"),
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

		View("routePanelAnalyticsViews", "/panel/analytics/views/").Before("ParseForm"),
		View("routePanelAnalyticsRoutes", "/panel/analytics/routes/").Before("ParseForm"),
		View("routePanelAnalyticsAgents", "/panel/analytics/agents/").Before("ParseForm"),
		View("routePanelAnalyticsRouteViews", "/panel/analytics/route/", "extraData"),
		View("routePanelAnalyticsAgentViews", "/panel/analytics/agent/", "extraData"),
		View("routePanelAnalyticsPosts", "/panel/analytics/posts/").Before("ParseForm"),
		View("routePanelAnalyticsTopics", "/panel/analytics/topics/").Before("ParseForm"),

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
