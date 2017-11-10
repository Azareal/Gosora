package main

type RouteImpl struct {
	Name      string
	Path      string
	Vars      []string
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

func addRouteGroup(routeGroup *RouteGroup) {
	routeGroups = append(routeGroups, routeGroup)
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", []string{}, []Runnable{}}
}

func Route(fname string, path string, args ...string) *RouteImpl {
	return &RouteImpl{fname, path, args, []Runnable{}}
}
