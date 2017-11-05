package main

type RouteImpl struct {
	Name      string
	Path      string
	Vars      []string
	RunBefore []Runnable
}

type RouteGroup struct {
	Path      string
	RouteList []*RouteImpl
	RunBefore []Runnable
}

type Runnable struct {
	Contents string
	Literal  bool
}

func addRoute(route *RouteImpl) {
	routeList = append(routeList, route)
}

func (route *RouteImpl) Before(item string, literal ...bool) *RouteImpl {
	var litItem bool
	if len(literal) > 0 {
		litItem = literal[0]
	}
	route.RunBefore = append(route.RunBefore, Runnable{item, litItem})
	return route
}

func newRouteGroup(path string, routes ...*RouteImpl) *RouteGroup {
	return &RouteGroup{path, routes, []Runnable{}}
}

func addRouteGroup(routeGroup *RouteGroup) {
	routeGroups = append(routeGroups, routeGroup)
}

func (group *RouteGroup) Before(line string, literal ...bool) *RouteGroup {
	var litItem bool
	if len(literal) > 0 {
		litItem = literal[0]
	}
	group.RunBefore = append(group.RunBefore, Runnable{line, litItem})
	return group
}

func (group *RouteGroup) Routes(routes ...*RouteImpl) {
	group.RouteList = append(group.RouteList, routes...)
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", []string{}, []Runnable{}}
}

func Route(fname string, path string, args ...string) *RouteImpl {
	return &RouteImpl{fname, path, args, []Runnable{}}
}

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

	reportGroup := newRouteGroup("/report/",
		Route("routeReportSubmit", "/report/submit/", "extra_data"),
	).Before("MemberOnly")
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
// TODO: Add a BeforeExcept method?
func buildUserRoutes() {
	userGroup := newRouteGroup("/user/") //.Before("MemberOnly")
	userGroup.Routes(
		Route("routeProfile", "/user/").Before("req.URL.Path += extra_data", true),
		Route("routeAccountOwnEditCritical", "/user/edit/critical/").Before("MemberOnly"),
		Route("routeAccountOwnEditCriticalSubmit", "/user/edit/critical/submit/").Before("MemberOnly"),
		Route("routeAccountOwnEditAvatar", "/user/edit/avatar/").Before("MemberOnly"),
		Route("routeAccountOwnEditAvatarSubmit", "/user/edit/avatar/submit/").Before("MemberOnly"),
		Route("routeAccountOwnEditUsername", "/user/edit/username/").Before("MemberOnly"),
		Route("routeAccountOwnEditUsernameSubmit", "/user/edit/username/submit/").Before("MemberOnly"),
		Route("routeAccountOwnEditEmail", "/user/edit/email/").Before("MemberOnly"),
		Route("routeAccountOwnEditEmailTokenSubmit", "/user/edit/token/", "extra_data").Before("MemberOnly"),
	)
	addRouteGroup(userGroup)
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/").Before("SuperModOnly")
	panelGroup.Routes(
		Route("routePanel", "/panel/"),
		Route("routePanelForums", "/panel/forums/"),
		Route("routePanelForumsCreateSubmit", "/panel/forums/create/"),
		Route("routePanelForumsDelete", "/panel/forums/delete/", "extra_data"),
		Route("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "extra_data"),
		Route("routePanelForumsEdit", "/panel/forums/edit/", "extra_data"),
		Route("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "extra_data"),
		Route("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "extra_data"),

		Route("routePanelSettings", "/panel/settings/"),
		Route("routePanelSetting", "/panel/settings/edit/", "extra_data"),
		Route("routePanelSettingEdit", "/panel/settings/edit/submit/", "extra_data"),

		Route("routePanelWordFilters", "/panel/settings/word-filters/"),
		Route("routePanelWordFiltersCreate", "/panel/settings/word-filters/create/"),
		Route("routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "extra_data"),
		Route("routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "extra_data"),
		Route("routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "extra_data"),

		Route("routePanelThemes", "/panel/themes/"),
		Route("routePanelThemesSetDefault", "/panel/themes/default/", "extra_data"),

		Route("routePanelPlugins", "/panel/plugins/"),
		Route("routePanelPluginsActivate", "/panel/plugins/activate/", "extra_data"),
		Route("routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "extra_data"),
		Route("routePanelPluginsInstall", "/panel/plugins/install/", "extra_data"),

		Route("routePanelUsers", "/panel/users/"),
		Route("routePanelUsersEdit", "/panel/users/edit/", "extra_data"),
		Route("routePanelUsersEditSubmit", "/panel/users/edit/submit/", "extra_data"),

		Route("routePanelGroups", "/panel/groups/"),
		Route("routePanelGroupsEdit", "/panel/groups/edit/", "extra_data"),
		Route("routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "extra_data"),
		Route("routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "extra_data"),
		Route("routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "extra_data"),
		Route("routePanelGroupsCreateSubmit", "/panel/groups/create/"),

		Route("routePanelBackups", "/panel/backups/", "extra_data"),
		Route("routePanelLogsMod", "/panel/logs/mod/"),
		Route("routePanelDebug", "/panel/debug/"),
	)
	addRouteGroup(panelGroup)
}
