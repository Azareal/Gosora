package main

type Router struct {
	routeList   []*RouteImpl
	routeGroups []*RouteGroup
}

func (r *Router) Add(route ...*RouteImpl) {
	r.routeList = append(r.routeList, route...)
}

func (r *Router) AddGroup(routeGroup ...*RouteGroup) {
	r.routeGroups = append(r.routeGroups, routeGroup...)
}