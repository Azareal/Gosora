package main

type RouteGroup struct {
	Path      string
	RouteList []*RouteImpl
	RunBefore []Runnable
}

func newRouteGroup(path string, routes ...*RouteImpl) *RouteGroup {
	group := &RouteGroup{Path: path}
	for _, route := range routes {
		route.Parent = group
		group.RouteList = append(group.RouteList, route)
	}
	return group
}

func (group *RouteGroup) Not(path ...string) *RouteSubset {
	routes := make([]*RouteImpl, len(group.RouteList))
	copy(routes, group.RouteList)
	for i, route := range routes {
		if inStringList(route.Path, path) {
			routes = append(routes[:i], routes[i+1:]...)
		}
	}
	return &RouteSubset{routes}
}

func inStringList(needle string, list []string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}
	return false
}

func (group *RouteGroup) Before(lines ...string) *RouteGroup {
	for _, line := range lines {
		group.RunBefore = append(group.RunBefore, Runnable{line, false})
	}
	return group
}

func (group *RouteGroup) LitBefore(lines ...string) *RouteGroup {
	for _, line := range lines {
		group.RunBefore = append(group.RunBefore, Runnable{line, true})
	}
	return group
}

func (group *RouteGroup) Routes(routes ...*RouteImpl) *RouteGroup {
	for _, route := range routes {
		route.Parent = group
		group.RouteList = append(group.RouteList, route)
	}
	return group
}
