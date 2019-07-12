package main

// TODO: How should we handle *HeaderLite and *Header?
func routes(r *Router) {
	r.Add(View("routes.Overview", "/overview/"))
	r.Add(View("routes.CustomPage", "/pages/", "extraData"))
	r.Add(View("routes.ForumList", "/forums/" /*,"&forums"*/))
	r.Add(View("routes.ViewForum", "/forum/", "extraData"))
	r.Add(AnonAction("routes.ChangeTheme", "/theme/"))
	r.Add(
		View("routes.ShowAttachment", "/attachs/", "extraData").Before("ParseForm").NoGzip().NoHeader(),
	)

	apiGroup := newRouteGroup("/api/",
		View("routeAPI", "/api/"),
		View("routeAPIPhrases", "/api/phrases/"), // TODO: Be careful with exposing the panel phrases here
		View("routes.APIMe", "/api/me/"),
		View("routeJSAntispam", "/api/watches/"),
	).NoHeader()
	r.AddGroup(apiGroup)

	// TODO: Reduce the number of Befores. With a new method, perhaps?
	reportGroup := newRouteGroup("/report/",
		Action("routes.ReportSubmit", "/report/submit/", "extraData"),
	).Before("NoBanned")
	r.AddGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		View("routes.TopicList", "/topics/"),
		View("routes.TopicListMostViewed", "/topics/most-viewed/"),
		MView("routes.CreateTopic", "/topics/create/", "extraData"),
	)
	r.AddGroup(topicGroup)

	r.AddGroup(panelRoutes())
	r.AddGroup(userRoutes())
	r.AddGroup(usersRoutes())
	r.AddGroup(topicRoutes())
	r.AddGroup(replyRoutes())
	r.AddGroup(profileReplyRoutes())
	r.AddGroup(pollRoutes())
	r.AddGroup(accountRoutes())

	r.Add(Special("common.RouteWebsockets", "/ws/"))
}

// TODO: Test the email token route
func userRoutes() *RouteGroup {
	return newRouteGroup("/user/").Routes(
		View("routes.ViewProfile", "/user/").LitBefore("req.URL.Path += extraData"),

		Set("routes.AccountEdit","/user/edit",
			MView("", "/"),
			MView("Password", "/password/"),
			Action("PasswordSubmit", "/password/submit/"), // TODO: Full test this
			UploadAction("AvatarSubmit", "/avatar/submit/").MaxSizeVar("int(c.Config.MaxRequestSize)"),
			Action("RevokeAvatarSubmit", "/avatar/revoke/submit/"),
			Action("UsernameSubmit", "/username/submit/"), // TODO: Full test this
			MView("MFA", "/mfa/"),
			MView("MFASetup", "/mfa/setup/"),
			Action("MFASetupSubmit", "/mfa/setup/submit/"),
			Action("MFADisableSubmit", "/mfa/disable/submit/"),
			MView("Email", "/email/"),
			View("EmailTokenSubmit", "/token/", "extraData").NoHeader(),
		),

		/*MView("routes.AccountEdit", "/user/edit/"),
		MView("routes.AccountEditPassword", "/user/edit/password/"),
		Action("routes.AccountEditPasswordSubmit", "/user/edit/password/submit/"), // TODO: Full test this
		UploadAction("routes.AccountEditAvatarSubmit", "/user/edit/avatar/submit/").MaxSizeVar("int(c.Config.MaxRequestSize)"),
		Action("routes.AccountEditRevokeAvatarSubmit", "/user/edit/avatar/revoke/submit/"),
		Action("routes.AccountEditUsernameSubmit", "/user/edit/username/submit/"), // TODO: Full test this
		MView("routes.AccountEditMFA", "/user/edit/mfa/"),
		MView("routes.AccountEditMFASetup", "/user/edit/mfa/setup/"),
		Action("routes.AccountEditMFASetupSubmit", "/user/edit/mfa/setup/submit/"),
		Action("routes.AccountEditMFADisableSubmit", "/user/edit/mfa/disable/submit/"),
		MView("routes.AccountEditEmail", "/user/edit/email/"),
		View("routes.AccountEditEmailTokenSubmit", "/user/edit/token/", "extraData").NoHeader(),*/

		MView("routes.AccountLogins", "/user/edit/logins/"),

		MView("routes.LevelList", "/user/levels/"),
		//MView("routes.LevelRankings", "/user/rankings/"),
		//MView("routes.Alerts", "/user/alerts/"),
		
		/*MView("routes.Convos", "/user/convos/"),
		MView("routes.ConvosCreate", "/user/convos/create/"),
		MView("routes.Convo", "/user/convo/","extraData"),
		Action("routes.ConvosCreateSubmit", "/user/convos/create/submit/"),
		Action("routes.ConvosDeleteSubmit", "/user/convos/delete/submit/","extraData"),
		Action("routes.ConvosCreateReplySubmit", "/user/convo/create/submit/"),
		Action("routes.ConvosDeleteReplySubmit", "/user/convo/delete/submit/"),*/
	)
}

func usersRoutes() *RouteGroup {
	// TODO: Auto test and manual test these routes
	return newRouteGroup("/users/").Routes(
		Action("routes.BanUserSubmit", "/users/ban/submit/", "extraData"),
		Action("routes.UnbanUser", "/users/unban/", "extraData"),
		Action("routes.ActivateUser", "/users/activate/", "extraData"),
		MView("routes.IPSearch", "/users/ips/"), // TODO: .Perms("ViewIPs")?
	)
}

func topicRoutes() *RouteGroup {
	return newRouteGroup("/topic/").Routes(
		View("routes.ViewTopic", "/topic/", "extraData"),
		UploadAction("routes.CreateTopicSubmit", "/topic/create/submit/").MaxSizeVar("int(c.Config.MaxRequestSize)"),
		Action("routes.EditTopicSubmit", "/topic/edit/submit/", "extraData"),
		Action("routes.DeleteTopicSubmit", "/topic/delete/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.StickTopicSubmit", "/topic/stick/submit/", "extraData"),
		Action("routes.UnstickTopicSubmit", "/topic/unstick/submit/", "extraData"),
		Action("routes.LockTopicSubmit", "/topic/lock/submit/").LitBefore("req.URL.Path += extraData"),
		Action("routes.UnlockTopicSubmit", "/topic/unlock/submit/", "extraData"),
		Action("routes.MoveTopicSubmit", "/topic/move/submit/", "extraData"),
		Action("routes.LikeTopicSubmit", "/topic/like/submit/", "extraData"),
		UploadAction("routes.AddAttachToTopicSubmit", "/topic/attach/add/submit/", "extraData").MaxSizeVar("int(c.Config.MaxRequestSize)"),
		Action("routes.RemoveAttachFromTopicSubmit", "/topic/attach/remove/submit/", "extraData"),
	)
}

func replyRoutes() *RouteGroup {
	return newRouteGroup("/reply/").Routes(
		// TODO: Reduce this to 1MB for attachments for each file?
		UploadAction("routes.CreateReplySubmit", "/reply/create/").MaxSizeVar("int(c.Config.MaxRequestSize)"), // TODO: Rename the route so it's /reply/create/submit/
		Action("routes.ReplyEditSubmit", "/reply/edit/submit/", "extraData"),
		Action("routes.ReplyDeleteSubmit", "/reply/delete/submit/", "extraData"),
		Action("routes.ReplyLikeSubmit", "/reply/like/submit/", "extraData"),
		//MemberView("routes.ReplyEdit","/reply/edit/","extraData"), // No js fallback
		//MemberView("routes.ReplyDelete","/reply/delete/","extraData"), // No js confirmation page? We could have a confirmation modal for the JS case
		UploadAction("routes.AddAttachToReplySubmit", "/reply/attach/add/submit/", "extraData").MaxSizeVar("int(c.Config.MaxRequestSize)"),
		Action("routes.RemoveAttachFromReplySubmit", "/reply/attach/remove/submit/", "extraData"),
	)
}

// TODO: Move these into /user/?
func profileReplyRoutes() *RouteGroup {
	return newRouteGroup("/profile/").Routes(
		Action("routes.ProfileReplyCreateSubmit", "/profile/reply/create/"), // TODO: Add /submit/ to the end
		Action("routes.ProfileReplyEditSubmit", "/profile/reply/edit/submit/", "extraData"),
		Action("routes.ProfileReplyDeleteSubmit", "/profile/reply/delete/submit/", "extraData"),
	)
}

func pollRoutes() *RouteGroup {
	return newRouteGroup("/poll/").Routes(
		Action("routes.PollVote", "/poll/vote/", "extraData"),
		View("routes.PollResults", "/poll/results/", "extraData").NoHeader(),
	)
}

func accountRoutes() *RouteGroup {
	//router.HandleFunc("/accounts/list/", routeLogin) // Redirect /accounts/ and /user/ to here.. // Get a list of all of the accounts on the forum
	return newRouteGroup("/accounts/").Routes(
		View("routes.AccountLogin", "/accounts/login/"),
		View("routes.AccountRegister", "/accounts/create/"),
		Action("routes.AccountLogout", "/accounts/logout/"),
		AnonAction("routes.AccountLoginSubmit", "/accounts/login/submit/"), // TODO: Guard this with a token, maybe the IP hashed with a rotated key?
		View("routes.AccountLoginMFAVerify", "/accounts/mfa_verify/"),
		AnonAction("routes.AccountLoginMFAVerifySubmit", "/accounts/mfa_verify/submit/"), // We have logic in here which filters out regular guests
		AnonAction("routes.AccountRegisterSubmit", "/accounts/create/submit/"),

		View("routes.AccountPasswordReset", "/accounts/password-reset/"),
		AnonAction("routes.AccountPasswordResetSubmit", "/accounts/password-reset/submit/"),
		View("routes.AccountPasswordResetToken", "/accounts/password-reset/token/"),
		AnonAction("routes.AccountPasswordResetTokenSubmit", "/accounts/password-reset/token/submit/"),
	)
}

func panelRoutes() *RouteGroup {
	return newRouteGroup("/panel/").Before("SuperModOnly").NoHeader().Routes(
		View("panel.Dashboard", "/panel/"),
		View("panel.Forums", "/panel/forums/"),
		Action("panel.ForumsCreateSubmit", "/panel/forums/create/"),
		Action("panel.ForumsDelete", "/panel/forums/delete/", "extraData"),
		Action("panel.ForumsDeleteSubmit", "/panel/forums/delete/submit/", "extraData"),
		Action("panel.ForumsOrderSubmit", "/panel/forums/order/edit/submit/"),
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

		View("panel.Themes", "/panel/themes/"),
		Action("panel.ThemesSetDefault", "/panel/themes/default/", "extraData"),
		View("panel.ThemesMenus", "/panel/themes/menus/"),
		View("panel.ThemesMenusEdit", "/panel/themes/menus/edit/", "extraData"),
		View("panel.ThemesMenuItemEdit", "/panel/themes/menus/item/edit/", "extraData"),
		Action("panel.ThemesMenuItemEditSubmit", "/panel/themes/menus/item/edit/submit/", "extraData"),
		Action("panel.ThemesMenuItemCreateSubmit", "/panel/themes/menus/item/create/submit/"),
		Action("panel.ThemesMenuItemDeleteSubmit", "/panel/themes/menus/item/delete/submit/", "extraData"),
		Action("panel.ThemesMenuItemOrderSubmit", "/panel/themes/menus/item/order/edit/submit/", "extraData"),

		View("panel.ThemesWidgets", "/panel/themes/widgets/"),
		//View("panel.ThemesWidgetsEdit", "/panel/themes/widgets/edit/", "extraData"),
		Action("panel.ThemesWidgetsEditSubmit", "/panel/themes/widgets/edit/submit/", "extraData"),
		Action("panel.ThemesWidgetsCreateSubmit", "/panel/themes/widgets/create/submit/"),
		Action("panel.ThemesWidgetsDeleteSubmit", "/panel/themes/widgets/delete/submit/", "extraData"),

		View("panel.Plugins", "/panel/plugins/"),
		Action("panel.PluginsActivate", "/panel/plugins/activate/", "extraData"),
		Action("panel.PluginsDeactivate", "/panel/plugins/deactivate/", "extraData"),
		Action("panel.PluginsInstall", "/panel/plugins/install/", "extraData"),

		View("panel.Users", "/panel/users/"),
		View("panel.UsersEdit", "/panel/users/edit/", "extraData"),
		Action("panel.UsersEditSubmit", "/panel/users/edit/submit/", "extraData"),
		UploadAction("panel.UsersAvatarSubmit", "/panel/users/avatar/submit/", "extraData").MaxSizeVar("int(c.Config.MaxRequestSize)"),
		Action("panel.UsersAvatarRemoveSubmit", "/panel/users/avatar/remove/submit/", "extraData"),

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
		View("panel.AnalyticsMemory", "/panel/analytics/memory/").Before("ParseForm"),
		View("panel.AnalyticsActiveMemory", "/panel/analytics/active-memory/").Before("ParseForm"),
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
}
