// Highly experimental plugin for caching rendered pages for guests
package main

import (
	//"log"
	"bytes"
	"errors"
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
	plugin.AddHook("tasks_tick_widget_wol",tickHdriveWol)
	plugin.AddHook("route_topic_list_start",jumpHdrive)
	return nil
}

func deactivateHdrive(plugin *c.Plugin) {
	plugin.RemoveHook("tasks_tick_topic_list",tickHdrive)
	plugin.RemoveHook("tasks_tick_widget_wol",tickHdriveWol)
	plugin.RemoveHook("route_topic_list_start",jumpHdrive)
	hyperspace = nil
}

type Hyperspace struct {
	topicList atomic.Value
	gzipTopicList atomic.Value
}

func newHyperspace() *Hyperspace {
	pageCache := new(Hyperspace)
	pageCache.topicList.Store([]byte(""))
	return pageCache
}

func tickHdriveWol(args ...interface{}) (skip bool, rerr c.RouteError) {
	c.DebugLog("docking at wol")
	return tickHdrive(args)
}

// TODO: Find a better way of doing this
func tickHdrive(args ...interface{}) (skip bool, rerr c.RouteError) {
	c.DebugLog("Refueling...")

	// Avoid accidentally caching already cached content
	hyperspace.topicList.Store([]byte(""))
	hyperspace.gzipTopicList.Store([]byte(""))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("get", "/topics/", bytes.NewReader(nil))
	user := c.GuestUser

	head, rerr := c.UserCheck(w, req, &user)
	if rerr != nil {
		return true, rerr
	}
			
	rerr = routes.TopicList(w, req, user, head)
	if rerr != nil {
		return true, rerr
	}
	if w.Code != 200 {
		c.LogWarning(errors.New("not 200 for topic list in hyperdrive"))
		return false, nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(w.Result().Body)
	hyperspace.topicList.Store(buf.Bytes())

	gbuf, err := c.CompressBytesGzip(buf.Bytes())
	if err != nil {
		c.LogWarning(err)
		return false, nil
	}
	hyperspace.gzipTopicList.Store(gbuf)
	
	return false, nil
}

func jumpHdrive(args ...interface{}) (skip bool, rerr c.RouteError) {
	var tList []byte
	w := args[0].(http.ResponseWriter)
	var iw http.ResponseWriter
	gzw, ok := w.(c.GzipResponseWriter)
	if ok {
		tList = hyperspace.gzipTopicList.Load().([]byte)
		iw = gzw.ResponseWriter
	} else {
		tList = hyperspace.topicList.Load().([]byte)
		iw = w
	}
	if len(tList) == 0 {
		c.DebugLog("no topiclist in hyperspace")
		return false, nil
	}
	//c.DebugLog("tList: ", tList)

	// Avoid intercepting user requests as we only have guests in cache right now
	user := args[2].(*c.User)
	if user.ID != 0 {
		c.DebugLog("not guest")
		return false, nil
	}
	
	// Avoid intercepting search requests and filters as we don't have those in cache
	r := args[1].(*http.Request)
	//c.DebugLog("r.URL.Path:",r.URL.Path)
	//c.DebugLog("r.URL.RawQuery:",r.URL.RawQuery)
	if r.URL.RawQuery != "" {
		return false, nil
	}
	if r.FormValue("js") == "1" {
		return false, nil
	}
	//c.DebugLog
	c.DebugLog("Successful jump")

	header := args[3].(*c.Header)
	routes.FootHeaders(w, header)
	iw.Write(tList)
	if ok {
		w.Header().Set("X-I","1")
	}

	return true, nil
}