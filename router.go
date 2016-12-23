package main
/*import "sync"
import "net/http"

type Router struct {
	mu sync.RWMutex
	routes map[string]http.Handler
}

func (route *Router) ServeHTTP() {
	route.mu.RLock()
	defer route.mu.RUnlock()
	
	if path[0] != "/" {
		return route.routes["/"]
	}
	
	// Do something on the path to turn slashes facing the wrong way "\" into "/" slashes. If it's bytes, then alter the bytes in place for the maximum speed
	
	handle := route.routes[path]
	if !ok {
		if path[-1] != "/" {
			handle = route.routes[path + "/"]
			if !ok {
				return route.routes["/"]
			}
			return handle
		}
	}
	return handle
}*/