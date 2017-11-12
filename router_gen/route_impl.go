package main

type RouteImpl struct {
	Name         string
	Path         string
	Vars         []string
	MemberAction bool
	RunBefore    []Runnable

	Parent *RouteGroup
}

type Runnable struct {
	Contents string
	Literal  bool
}

func addRoute(route *RouteImpl) {
	routeList = append(routeList, route)
}

func (route *RouteImpl) Before(items ...string) *RouteImpl {
	for _, item := range items {
		route.RunBefore = append(route.RunBefore, Runnable{item, false})
	}
	return route
}

func (route *RouteImpl) LitBefore(items ...string) *RouteImpl {
	for _, item := range items {
		route.RunBefore = append(route.RunBefore, Runnable{item, true})
	}
	return route
}

func (route *RouteImpl) hasBefore(items ...string) bool {
	for _, item := range items {
		for _, before := range route.RunBefore {
			if before.Contents == item {
				return true
			}
		}
	}
	return false
}

func addRouteGroup(routeGroup *RouteGroup) {
	routeGroups = append(routeGroups, routeGroup)
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", []string{}, false, []Runnable{}, nil}
}

func View(fname string, path string, args ...string) *RouteImpl {
	return &RouteImpl{fname, path, args, false, []Runnable{}, nil}
}

func MemberView(fname string, path string, args ...string) *RouteImpl {
	route := &RouteImpl{fname, path, args, false, []Runnable{}, nil}
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func ModView(fname string, path string, args ...string) *RouteImpl {
	route := &RouteImpl{fname, path, args, false, []Runnable{}, nil}
	if !route.hasBefore("AdminOnly") {
		route.Before("SuperModOnly")
	}
	return route
}

func Action(fname string, path string, args ...string) *RouteImpl {
	route := &RouteImpl{fname, path, args, true, []Runnable{}, nil}
	route.Before("NoSessionMismatch")
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func AnonAction(fname string, path string, args ...string) *RouteImpl {
	route := &RouteImpl{fname, path, args, false, []Runnable{}, nil}
	route.Before("ParseForm")
	return route
}

func UploadAction(fname string, path string, args ...string) *RouteImpl {
	route := &RouteImpl{fname, path, args, true, []Runnable{}, nil}
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}
