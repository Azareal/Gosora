package main

// TODO: How should we handle *HeaderLite and *Header?
func routes() {
	addRoute(View("routes.Overview", "/overview/"))
	addRoute(View("routes.CustomPage", "/pages/", "extraData"))
	addRoute(View("routes.ForumList", "/forums/" /*,"&forums"*/))
	addRoute(View("routes.ViewForum", "/forum/", "extraData"))
	addRoute(AnonAction("routes.ChangeTheme", "/theme/"))
	addRoute(
		View("routes.ShowAttachment", "/attachs/", "extraData").Before("ParseForm").NoGzip(),
	)

	apiGroup := newRouteGroup("/api/",
		View("routeAPI", "/api/"),
		View("routeAPIPhrases", "/api/phrases/"), // TODO: Be careful with exposing the panel phrases here
		View("routes.APIMe", "/api/me/"),
		View("routeJSAntispam", "/api/watches/"),
	)
	addRouteGroup(apiGroup)

	// TODO: Reduce the number of Befores. With a new method, perhaps?
	reportGroup := newRouteGroup("/report/",
		Action("routes.ReportSubmit", "/report/submit/", "extraData"),
	).Before("NoBanned")
	addRouteGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		View("routes.TopicList", "/topics/"),
		View("routes.TopicListMostViewed", "/topics/most-viewed/"),
		MemberView("routes.CreateTopic", "/topics/create/", "extraData"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
	buildUserRoutes()
	buildTopicRoutes()
	buildReplyRoutes()
	buildProfileReplyRoutes()
	buildPollRoutes()
	buildAccountRoutes()

	addRoute(Special("common.RouteWebsockets", "/ws/"))
}

// TODO: Test the email token route
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/")
	userGroup.Routes(
		View("routes.ViewProfile", "/user/").LitBefore("req.URL.Path += extraData"),
		MemberView("routes.AccountEdit", "/user/edit/"),
		MemberView("routes.AccountEditPassword", "/user/edit/password/"),
		Action("routes.AccountEditPasswordSubmit", "/user/edit/password/submit/"), // TODO: Full test this
		UploadAction("routes.AccountEditAvatarSubmit", "/user/edit/avatar/submit/").MaxSizeVar("int(common.Config.MaxRequestSize)"),
		Action("routes.AccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		MemberView("routes.AccountEditMFA", "/user/edit/mfa/"),
		MemberView("routes.AccountEditMFASetup", "/user/edit/mfa/setup/"),
		Action("routes.AccountEditMFASetupSubmit", "/user/edit/mfa/setup/submit/"),
		Action("routes.AccountEditMFADisableSubmit", "/user/edit/mfa/disable/submit/"),
		MemberView("routes.AccountEditEmail", "/user/edit/email/"),
		Action("routes.AccountEditEmailTokenSubmit", "/user/edit/token/", "extraData"),
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
		View("routes.ViewTopic", "/topic/", "extraData"),
		UploadAction("routes.CreateTopicSubmit", "/topic/create/submit/").MaxSizeVar("int(common.Config.MaxRequestSize)"),
		Action("routes.EditTopicSubmit", "/topic/edit/submit/", "extraData"),
		Action("routes.DeleteTopicSubmit", "/topic/delete/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.StickTopicSubmit", "/topic/stick/submit/", "extraData"),
		Action("routes.UnstickTopicSubmit", "/topic/unstick/submit/", "extraData"),
		Action("routes.LockTopicSubmit", "/topic/lock/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.UnlockTopicSubmit", "/topic/unlock/submit/", "extraData"),
		Action("routes.MoveTopicSubmit", "/topic/move/submit/", "extraData"),
		Action("routes.LikeTopicSubmit", "/topic/like/submit/", "extraData").Before("ParseForm"),
	)
	addRouteGroup(topicGroup)
}

func buildReplyRoutes() {
	//router.HandleFunc("/reply/edit/", routeReplyEdit) // No js fallback
	//router.HandleFunc("/reply/delete/", routeReplyDelete) // No js confirmation page? We could have a confirmation modal for the JS case
	replyGroup := newRouteGroup("/reply/")
	replyGroup.Routes(
		// TODO: Reduce this to 1MB for attachments for each file?
		UploadAction("routes.CreateReplySubmit", "/reply/create/").MaxSizeVar("int(common.Config.MaxRequestSize)"), // TODO: Rename the route so it's /reply/create/submit/
		Action("routes.ReplyEditSubmit", "/reply/edit/submit/", "extraData"),
		Action("routes.ReplyDeleteSubmit", "/reply/delete/submit/", "extraData"),
		Action("routes.ReplyLikeSubmit", "/reply/like/submit/", "extraData").Before("ParseForm"),
	)
	addRouteGroup(replyGroup)
}

// TODO: Move these into /user/?
func buildProfileReplyRoutes() {
	pReplyGroup := newRouteGroup("/profile/")
	pReplyGroup.Routes(
		Action("routes.ProfileReplyCreateSubmit", "/profile/reply/create/"), // TODO: Add /submit/ to the end
		Action("routes.ProfileReplyEditSubmit", "/profile/reply/edit/submit/", "extraData"),
		Action("routes.ProfileReplyDeleteSubmit", "/profile/reply/delete/submit/", "extraData"),
	)
	addRouteGroup(pReplyGroup)
}

func buildPollRoutes() {
	pollGroup := newRouteGroup("/poll/")
	pollGroup.Routes(
		Action("routes.PollVote", "/poll/vote/", "extraData"),
		View("routes.PollResults", "/poll/results/", "extraData"),
	)
	addRouteGroup(pollGroup)
}

func buildAccountRoutes() {
	//router.HandleFunc("/accounts/list/", routeLogin) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	accReplyGroup := newRouteGroup("/accounts/")
	accReplyGroup.Routes(
		View("routes.AccountLogin", "/accounts/login/"),
		View("routes.AccountRegister", "/accounts/create/"),
		Action("routes.AccountLogout", "/accounts/logout/"),
		AnonAction("routes.AccountLoginSubmit", "/accounts/login/submit/"), // TODO: Guard this with a token, maybe the IP hashed with a rotated key?
		View("routes.AccountLoginMFAVerify", "/accounts/mfa_verify/"),
		AnonAction("routes.AccountLoginMFAVerifySubmit", "/accounts/mfa_verify/submit/"), // We have logic in here which filters out regular guests
		AnonAction("routes.AccountRegisterSubmit", "/accounts/create/submit/"),
	)
	addRouteGroup(accReplyGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		View("routePanelDashboard", "/panel/"),
		View("panel.Forums", "/panel/forums/"),
		Action("panel.ForumsCreateSubmit", "/panel/forums/create/"),
		Action("panel.ForumsDelete", "/panel/forums/delete/", "extraData"),
		Action("panel.ForumsDeleteSubmit", "/panel/forums/delete/submit/", "extraData"),
		View("panel.ForumsEdit", "/panel/forums/edit/", "extraData"),
		Action("panel.ForumsEditSubmit", "/panel/forums/edit/submit/", "extraData"),
		Action("panel.ForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extraData"),
		View("panel.ForumsEditPermsAdvance", "/panel/forums/edit/perms/", "extraData"),
		Action("panel.ForumsEditPermsAdvanceSubmit", "/panel/forums/edit/perms/adv/submit/", "extraData"),

		View("panel.Settings", "/panel/settings/"),
		View("panel.SettingEdit", "/panel/settings/edit/", "extraData"),
		Action("panel.SettingEditSubmit", "/panel/settings/edit/submit/", "extraData"),

		View("panel.WordFilters", "/panel/settings/word-filters/"),
		Action("panel.WordFiltersCreateSubmit", "/panel/settings/word-filters/create/"),
		View("panel.WordFiltersEdit", "/panel/settings/word-filters/edit/", "extraData"),
		Action("panel.WordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "extraData"),
		Action("panel.WordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "extraData"),

		View("panel.Pages", "/panel/pages/").Before("AdminOnly"),
		Action("panel.PagesCreateSubmit", "/panel/pages/create/submit/").Before("AdminOnly"),
		View("panel.PagesEdit", "/panel/pages/edit/", "extraData").Before("AdminOnly"),
		Action("panel.PagesEditSubmit", "/panel/pages/edit/submit/", "extraData").Before("AdminOnly"),
		Action("panel.PagesDeleteSubmit", "/panel/pages/delete/submit/", "extraData").Before("AdminOnly"),

		View("routePanelThemes", "/panel/themes/"),
		Action("routePanelThemesSetDefault", "/panel/themes/default/", "extraData"),
		View("routePanelThemesMenus", "/panel/themes/menus/"),
		View("routePanelThemesMenusEdit", "/panel/themes/menus/edit/", "extraData"),
		View("routePanelThemesMenuItemEdit", "/panel/themes/menus/item/edit/", "extraData"),
		Action("routePanelThemesMenuItemEditSubmit", "/panel/themes/menus/item/edit/submit/", "extraData"),
		Action("routePanelThemesMenuItemCreateSubmit", "/panel/themes/menus/item/create/submit/"),
		Action("routePanelThemesMenuItemDeleteSubmit", "/panel/themes/menus/item/delete/submit/", "extraData"),
		Action("routePanelThemesMenuItemOrderSubmit", "/panel/themes/menus/item/order/edit/submit/", "extraData"),

		View("panel.Plugins", "/panel/plugins/"),
		Action("panel.PluginsActivate", "/panel/plugins/activate/", "extraData"),
		Action("panel.PluginsDeactivate", "/panel/plugins/deactivate/", "extraData"),
		Action("panel.PluginsInstall", "/panel/plugins/install/", "extraData"),

		View("panel.Users", "/panel/users/"),
		View("panel.UsersEdit", "/panel/users/edit/", "extraData"),
		Action("panel.UsersEditSubmit", "/panel/users/edit/submit/", "extraData"),

		View("panel.AnalyticsViews", "/panel/analytics/views/").Before("ParseForm"),
		View("panel.AnalyticsRoutes", "/panel/analytics/routes/").Before("ParseForm"),
		View("panel.AnalyticsAgents", "/panel/analytics/agents/").Before("ParseForm"),
		View("panel.AnalyticsSystems", "/panel/analytics/systems/").Before("ParseForm"),
		View("panel.AnalyticsLanguages", "/panel/analytics/langs/").Before("ParseForm"),
		View("panel.AnalyticsReferrers", "/panel/analytics/referrers/").Before("ParseForm"),
		View("panel.AnalyticsRouteViews", "/panel/analytics/route/", "extraData"),
		View("panel.AnalyticsAgentViews", "/panel/analytics/agent/", "extraData"),
		View("panel.AnalyticsForumViews", "/panel/analytics/forum/", "extraData"),
		View("panel.AnalyticsSystemViews", "/panel/analytics/system/", "extraData"),
		View("panel.AnalyticsLanguageViews", "/panel/analytics/lang/", "extraData"),
		View("panel.AnalyticsReferrerViews", "/panel/analytics/referrer/", "extraData"),
		View("panel.AnalyticsPosts", "/panel/analytics/posts/").Before("ParseForm"),
		View("panel.AnalyticsTopics", "/panel/analytics/topics/").Before("ParseForm"),
		View("panel.AnalyticsForums", "/panel/analytics/forums/").Before("ParseForm"),

		View("panel.Groups", "/panel/groups/"),
		View("panel.GroupsEdit", "/panel/groups/edit/", "extraData"),
		View("panel.GroupsEditPerms", "/panel/groups/edit/perms/", "extraData"),
		Action("panel.GroupsEditSubmit", "/panel/groups/edit/submit/", "extraData"),
		Action("panel.GroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "extraData"),
		Action("panel.GroupsCreateSubmit", "/panel/groups/create/"),

		View("panel.Backups", "/panel/backups/", "extraData").Before("SuperAdminOnly").NoGzip(), // TODO: Tests for this
		View("panel.LogsRegs", "/panel/logs/regs/"),
		View("panel.LogsMod", "/panel/logs/mod/"),
		View("panel.Debug", "/panel/debug/").Before("AdminOnly"),
	)
	addRouteGroup(panelGroup)
}
