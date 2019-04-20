// Highly experimental plugin for caching rendered pages for guests
package main

import (
	"log"
	"bytes"
	"sync/atomic"
	"net/http"
	"net/http/httptest"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/routes"
)

var hyperspace *Hyperspace

func init() {
	c.Plugins.Add(&c.Plugin{UName: "hyperdrive", Name: "Hyperdrive", Author: "Azareal", Init: initHdrive, Deactivate: deactivateHdrive})
}

func initHdrive(plugin *c.Plugin) error {
	hyperspace = newHyperspace()
	plugin.AddHook("tasks_tick_topic_list",tickHdrive)
	plugin.AddHook("route_topic_list_start",jumpHdrive)
	return nil
}

func deactivateHdrive(plugin *c.Plugin) {
	plugin.RemoveHook("tasks_tick_topic_list",tickHdrive)
	plugin.RemoveHook("route_topic_list_start",jumpHdrive)
	hyperspace = nil
}

type Hyperspace struct {
	topicList atomic.Value
}

func newHyperspace() *Hyperspace {
	pageCache := new(Hyperspace)
	pageCache.topicList.Store([]byte(""))
	return pageCache
}

// TODO: Find a better way of doing this
func tickHdrive(args ...interface{}) (skip bool, rerr c.RouteError) {
	log.Print("Refueling...")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("get", "/topics/", bytes.NewReader(nil))
	user := c.GuestUser

	head, err := c.UserCheck(w, req, &user)
	if err != nil {
		c.LogWarning(err)
		return true, rerr
	}
			
	rerr = routes.TopicList(w, req, user, head)
	if rerr != nil {
		c.LogWarning(err)
		return true, rerr
	}
	if w.Code != 200 {
		c.LogWarning(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(w.Result().Body)
	hyperspace.topicList.Store(buf.Bytes())

	return false, nil
}

func jumpHdrive(args ...interface{}) (skip bool, rerr c.RouteError) {
	tList := hyperspace.topicList.Load().([]byte)
	if len(tList) == 0 {
		log.Print("no topiclist in hyperspace")
		return false, nil
	}

	// Avoid intercepting user requests as we only have guests in cache right now
	user := args[2].(*c.User)
	if user.ID != 0 {
		log.Print("not guest")
		return false, nil
	}
	
	// Avoid intercepting search requests and filters as we don't have those in cache
	r := args[1].(*http.Request)
	//log.Print("r.URL.Path:",r.URL.Path)
	log.Print("r.URL.RawQuery:",r.URL.RawQuery)
	if r.URL.RawQuery != "" {
		return false, nil
	}
	log.Print("Successful jump")

	w := args[0].(http.ResponseWriter)
	header := args[3].(*c.Header)
	routes.FootHeaders(w, header)
	w.Write(tList)
	
	return true, nil
}