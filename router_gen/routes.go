package main

type Route struct {
	Name string
	Path string
	Before string
	Vars []string
}

type RouteGroup struct {
	Path string
	Routes []Route
}

func addRoute(fname string, path string, before string, vars ...string) {
	route_list = append(route_list,Route{fname,path,before,vars})
}

func addRouteGroup(path string, routes ...Route) {
	route_groups = append(route_groups,RouteGroup{path,routes})
}

func routes() {
	//addRoute("default_route","","")
	addRoute("route_static","/static/","req.URL.Path += extra_data")
	addRoute("route_overview","/overview/","")
	addRoute("route_custom_page","/pages/",""/*,"&extra_data"*/)
	addRoute("route_forums","/forums/",""/*,"&forums"*/)
	
	//addRoute("route_topic_create","/topics/create/","","extra_data")
	//addRoute("route_topics","/topics/",""/*,"&groups","&forums"*/)
	addRouteGroup("/topics/",
		Route{"route_topics","/topics/","",[]string{}},
		Route{"route_topic_create","/topics/create/","",[]string{"extra_data"}},
	)
}
