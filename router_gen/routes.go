package main

type Route struct {
	Name string
	Path string
	Vars []string
}

type RouteGroup struct {
	Path string
	Routes []Route
}

func addRoute(fname string, path string, vars ...string) {
	route_list = append(route_list,Route{fname,path,vars})
}

func addRouteGroup(path string, routes ...Route) {
	route_groups = append(route_groups,RouteGroup{path,routes})
}

func routes() {
	//addRoute("route_static","/static/","&extra_data")
	addRoute("route_overview","/overview/")
	addRoute("route_custom_page","/pages/","&extra_data")
	addRoute("route_topics","/topics/","&groups","&forums")
}
