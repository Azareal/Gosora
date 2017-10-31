package main

type RouteImpl struct {
	Name   string
	Path   string
	Before string
	Vars   []string
}

type RouteGroup struct {
	Path      string
	RouteList []*RouteImpl
	Before    []string
}

func addRoute(fname string, path string, before string, vars ...string) {
	routeList = append(routeList, &RouteImpl{fname, path, before, vars})
}

func newRouteGroup(path string, routes ...*RouteImpl) *RouteGroup {
	return &RouteGroup{path, routes, []string{}}
}

func addRouteGroup(routeGroup *RouteGroup) {
	routeGroups = append(routeGroups, routeGroup)
}

func (group *RouteGroup) RunBefore(line string) {
	group.Before = append(group.Before, line)
}

func (group *RouteGroup) Routes(routes ...*RouteImpl) {
	group.RouteList = append(group.RouteList, routes...)
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", "", []string{}}
}

func Route(fname string, path string, args ...string) *RouteImpl {
	var before = ""
	if len(args) > 0 {
		before = args[0]
		args = args[1:]
	}
	return &RouteImpl{fname, path, before, args}
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
	addRoute("routeShowAttachment", "/attachs/", "", "extra_data")

	reportGroup := newRouteGroup("/report/",
		Route("routeReportSubmit", "/report/submit/", "", "extra_data"),
	)
	addRouteGroup(reportGroup)

	topicGroup := newRouteGroup("/topics/",
		Route("routeTopics", "/topics/"),
		Route("routeTopicCreate", "/topics/create/", "", "extra_data"),
	)
	addRouteGroup(topicGroup)

	buildPanelRoutes()
}

func buildPanelRoutes() {
	panelGroup := newRouteGroup("/panel/")
	panelGroup.RunBefore("SuperModOnly")
	panelGroup.Routes(
		Route("routePanel", "/panel/"),
		Route("routePanelForums", "/panel/forums/"),
		Route("routePanelForumsCreateSubmit", "/panel/forums/create/"),
		Route("routePanelForumsDelete", "/panel/forums/delete/", "", "extra_data"),
		Route("routePanelForumsDeleteSubmit", "/panel/forums/delete/submit/", "", "extra_data"),
		Route("routePanelForumsEdit", "/panel/forums/edit/", "", "extra_data"),
		Route("routePanelForumsEditSubmit", "/panel/forums/edit/submit/", "", "extra_data"),
		Route("routePanelForumsEditPermsSubmit", "/panel/forums/edit/perms/submit/", "", "extra_data"),

		Route("routePanelSettings", "/panel/settings/"),
		Route("routePanelSetting", "/panel/settings/edit/", "", "extra_data"),
		Route("routePanelSettingEdit", "/panel/settings/edit/submit/", "", "extra_data"),

		Route("routePanelWordFilters", "/panel/settings/word-filters/"),
		Route("routePanelWordFiltersCreate", "/panel/settings/word-filters/create/"),
		Route("routePanelWordFiltersEdit", "/panel/settings/word-filters/edit/", "", "extra_data"),
		Route("routePanelWordFiltersEditSubmit", "/panel/settings/word-filters/edit/submit/", "", "extra_data"),
		Route("routePanelWordFiltersDeleteSubmit", "/panel/settings/word-filters/delete/submit/", "", "extra_data"),

		Route("routePanelThemes", "/panel/themes/"),
		Route("routePanelThemesSetDefault", "/panel/themes/default/", "", "extra_data"),

		Route("routePanelPlugins", "/panel/plugins/"),
		Route("routePanelPluginsActivate", "/panel/plugins/activate/", "", "extra_data"),
		Route("routePanelPluginsDeactivate", "/panel/plugins/deactivate/", "", "extra_data"),
		Route("routePanelPluginsInstall", "/panel/plugins/install/", "", "extra_data"),

		Route("routePanelUsers", "/panel/users/"),
		Route("routePanelUsersEdit", "/panel/users/edit/", "", "extra_data"),
		Route("routePanelUsersEditSubmit", "/panel/users/edit/submit/", "", "extra_data"),

		Route("routePanelGroups", "/panel/groups/"),
		Route("routePanelGroupsEdit", "/panel/groups/edit/", "", "extra_data"),
		Route("routePanelGroupsEditPerms", "/panel/groups/edit/perms/", "", "extra_data"),
		Route("routePanelGroupsEditSubmit", "/panel/groups/edit/submit/", "", "extra_data"),
		Route("routePanelGroupsEditPermsSubmit", "/panel/groups/edit/perms/submit/", "", "extra_data"),
		Route("routePanelGroupsCreateSubmit", "/panel/groups/create/"),

		Route("routePanelBackups", "/panel/backups/", "", "extra_data"),
		Route("routePanelLogsMod", "/panel/logs/mod/"),
		Route("routePanelDebug", "/panel/debug/"),
	)
	addRouteGroup(panelGroup)
}
