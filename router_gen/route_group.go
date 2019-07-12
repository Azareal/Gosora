package main

import "strings"

type RouteGroup struct {
	Path      string
	RouteList []*RouteImpl
	RunBefore []Runnable

	NoHead bool
}

func newRouteGroup(path string, routes ...*RouteImpl) *RouteGroup {
	g := &RouteGroup{Path: path}
	for _, route := range routes {
		route.Parent = g
		g.RouteList = append(g.RouteList, route)
	}
	return g
}

func (g *RouteGroup) Not(path ...string) *RouteSubset {
	routes := make([]*RouteImpl, len(g.RouteList))
	copy(routes, g.RouteList)
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

func (g *RouteGroup) NoHeader() *RouteGroup {
	g.NoHead = true
	return g
}

func (g *RouteGroup) Before(lines ...string) *RouteGroup {
	for _, line := range lines {
		g.RunBefore = append(g.RunBefore, Runnable{line, false})
	}
	return g
}

func (g *RouteGroup) LitBefore(lines ...string) *RouteGroup {
	for _, line := range lines {
		g.RunBefore = append(g.RunBefore, Runnable{line, true})
	}
	return g
}

/*func (g *RouteGroup) Routes(routes ...*RouteImpl) *RouteGroup {
	for _, route := range routes {
		route.Parent = g
		g.RouteList = append(g.RouteList, route)
	}
	return g
}*/

func (g *RouteGroup) Routes(routes ...interface{}) *RouteGroup {
	for _, route := range routes {
		switch r := route.(type) {
		case *RouteImpl:
			r.Parent = g
			g.RouteList = append(g.RouteList, r)
		case RouteSet:
			for _, rr := range r.Items {
				rr.Name = r.Name + rr.Name
				rr.Path = strings.TrimSuffix(r.Path, "/") + "/" + strings.TrimPrefix(rr.Path, "/")
				rr.Parent = g
				g.RouteList = append(g.RouteList, rr)
			}
		}
	}
	return g
}
