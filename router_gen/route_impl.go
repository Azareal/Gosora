package main

import "strings"

type RouteImpl struct {
	Name      string
	Path      string
	Action    bool
	NoHead    bool
	Vars      []string
	RunBefore []Runnable

	Parent *RouteGroup
}

type Runnable struct {
	Contents string
	Literal  bool
}

func (r *RouteImpl) Before(items ...string) *RouteImpl {
	for _, item := range items {
		r.RunBefore = append(r.RunBefore, Runnable{item, false})
	}
	return r
}

func (r *RouteImpl) LitBefore(items ...string) *RouteImpl {
	for _, item := range items {
		r.RunBefore = append(r.RunBefore, Runnable{item, true})
	}
	return r
}

func (r *RouteImpl) LitBeforeMultiline(items ...string) *RouteImpl {
	for _, item := range items {
		for _, line := range strings.Split(item, "\n") {
			r.LitBefore(strings.TrimSpace(line))
		}
	}
	return r
}

func (r *RouteImpl) hasBefore(items ...string) bool {
	for _, item := range items {
		if r.hasBeforeItem(item) {
			return true
		}
	}
	return false
}

func (r *RouteImpl) hasBeforeItem(item string) bool {
	for _, before := range r.RunBefore {
		if before.Contents == item {
			return true
		}
	}
	return false
}

func (r *RouteImpl) NoGzip() *RouteImpl {
	return r.LitBeforeMultiline(`gzw, ok := w.(c.GzipResponseWriter)
	if ok {
		w = gzw.ResponseWriter
		w.Header().Del("Content-Type")
		w.Header().Del("Content-Encoding")
	}`)
}

func (r *RouteImpl) NoHeader() *RouteImpl {
	r.NoHead = true
	return r
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", false, false, []string{}, []Runnable{}, nil}
}

func route(fname string, path string, action bool, special bool, args ...string) *RouteImpl {
	return &RouteImpl{fname, path, action, special, args, []Runnable{}, nil}
}

func View(fname string, path string, args ...string) *RouteImpl {
	return route(fname, path, false, false, args...)
}

func MView(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, false, false, args...)
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func MemberView(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, false, false, args...)
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func ModView(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, false, false, args...)
	if !route.hasBefore("AdminOnly") {
		route.Before("SuperModOnly")
	}
	return route
}

func Action(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, true, false, args...)
	route.Before("NoSessionMismatch")
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func AnonAction(fname string, path string, args ...string) *RouteImpl {
	return route(fname, path, true, false, args...).Before("ParseForm")
}

func Special(fname string, path string, args ...string) *RouteImpl {
	return route(fname, path, false, true, args...).LitBefore("req.URL.Path += extraData")
}

// Make this it's own type to force the user to manipulate methods on it to set parameters
type uploadAction struct {
	Route *RouteImpl
}

func UploadAction(fname string, path string, args ...string) *uploadAction {
	route := route(fname, path, true, false, args...)
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return &uploadAction{route}
}

func (action *uploadAction) MaxSizeVar(varName string) *RouteImpl {
	action.Route.LitBeforeMultiline(`err = c.HandleUploadRoute(w,req,user,` + varName + `)
			if err != nil {
				return err
			}`)
	action.Route.Before("NoUploadSessionMismatch")
	return action.Route
}

type RouteSet struct {
	Name  string
	Path  string
	Items []*RouteImpl
}

func Set(name string, path string, routes ...*RouteImpl) RouteSet {
	return RouteSet{name, path, routes}
}
