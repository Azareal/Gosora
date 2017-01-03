package main
//import "fmt"
import "strings"
import "sync"
import "net/http"

type Router struct {
	mu sync.RWMutex
	routes map[string]func(http.ResponseWriter, *http.Request)
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(http.ResponseWriter, *http.Request)),
	}
}

func (router *Router) Handle(pattern string, handle http.Handler) {
	router.routes[pattern] = handle.ServeHTTP
}

func (router *Router) HandleFunc(pattern string, handle func(http.ResponseWriter, *http.Request)) {
	router.routes[pattern] = handle
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.mu.RLock()
	defer router.mu.RUnlock()
	
	if req.URL.Path[0] != '/' {
		w.WriteHeader(405)
		w.Write([]byte(""))
		return
	}
	
	// Do something on the path to turn slashes facing the wrong way "\" into "/" slashes. If it's bytes, then alter the bytes in place for the maximum speed
	
	handle, ok := router.routes[req.URL.Path]
	if ok {
		handle(w,req)
		return
	}
	
	if req.URL.Path[len(req.URL.Path) - 1] == '/' {
		w.WriteHeader(404)
		w.Write([]byte(""))
		return
	}
	
	handle, ok = router.routes[req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/') + 1]]
	if ok {
		handle(w,req)
		return
	}
	//fmt.Println(req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/')])
	
	handle, ok = router.routes[req.URL.Path + "/"]
	if ok {
		handle(w,req)
		return
	}
	
	w.WriteHeader(404)
	w.Write([]byte(""))
	return
}