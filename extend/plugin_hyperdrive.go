// Highly experimental plugin for caching rendered pages for guests
package extend

import (
	//"log"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/routes"
)

var hyperspace *Hyperspace

func init() {
	c.Plugins.Add(&c.Plugin{UName: "hyperdrive", Name: "Hyperdrive", Author: "Azareal", Init: initHdrive, Deactivate: deactivateHdrive})
}

func initHdrive(pl *c.Plugin) error {
	hyperspace = newHyperspace()
	pl.AddHook("tasks_tick_topic_list", tickHdrive)
	pl.AddHook("tasks_tick_widget_wol", tickHdriveWol)
	pl.AddHook("route_topic_list_start", jumpHdriveTopicList)
	pl.AddHook("route_forum_list_start", jumpHdriveForumList)
	tickHdrive()
	return nil
}

func deactivateHdrive(pl *c.Plugin) {
	pl.RemoveHook("tasks_tick_topic_list", tickHdrive)
	pl.RemoveHook("tasks_tick_widget_wol", tickHdriveWol)
	pl.RemoveHook("route_topic_list_start", jumpHdriveTopicList)
	pl.RemoveHook("route_forum_list_start", jumpHdriveForumList)
	hyperspace = nil
}

type Hyperspace struct {
	topicList           atomic.Value
	gzipTopicList       atomic.Value
	forumList           atomic.Value
	gzipForumList       atomic.Value
	lastTopicListUpdate atomic.Value
}

func newHyperspace() *Hyperspace {
	pageCache := new(Hyperspace)
	blank := make(map[string][]byte, len(c.Themes))
	pageCache.topicList.Store(blank)
	pageCache.gzipTopicList.Store(blank)
	pageCache.forumList.Store(blank)
	pageCache.gzipForumList.Store(blank)
	pageCache.lastTopicListUpdate.Store(int64(0))
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
	blank := make(map[string][]byte, len(c.Themes))
	hyperspace.topicList.Store(blank)
	hyperspace.gzipTopicList.Store(blank)
	hyperspace.forumList.Store(blank)
	hyperspace.gzipForumList.Store(blank)

	tListMap := make(map[string][]byte)
	gtListMap := make(map[string][]byte)
	fListMap := make(map[string][]byte)
	gfListMap := make(map[string][]byte)

	cacheTheme := func(tname string) (skip, fail bool, rerr c.RouteError) {

		themeCookie := http.Cookie{Name: "current_theme", Value: tname, Path: "/", MaxAge: c.Year}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("get", "/topics/", bytes.NewReader(nil))
		req.AddCookie(&themeCookie)
		user := c.GuestUser

		head, rerr := c.UserCheck(w, req, &user)
		if rerr != nil {
			return true, true, rerr
		}

		rerr = routes.TopicList(w, req, &user, head)
		if rerr != nil {
			return true, true, rerr
		}
		if w.Code != 200 {
			c.LogWarning(errors.New("not 200 for topic list in hyperdrive"))
			return false, true, nil
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(w.Result().Body)
		tListMap[tname] = buf.Bytes()

		gbuf, err := c.CompressBytesGzip(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}
		gtListMap[tname] = gbuf

		w = httptest.NewRecorder()
		req = httptest.NewRequest("get", "/forums/", bytes.NewReader(nil))
		user = c.GuestUser

		head, rerr = c.UserCheck(w, req, &user)
		if rerr != nil {
			return true, true, rerr
		}

		rerr = routes.ForumList(w, req, &user, head)
		if rerr != nil {
			return true, true, rerr
		}
		if w.Code != 200 {
			c.LogWarning(errors.New("not 200 for forum list in hyperdrive"))
			return false, true, nil
		}

		buf = new(bytes.Buffer)
		buf.ReadFrom(w.Result().Body)
		fListMap[tname] = buf.Bytes()

		gbuf, err = c.CompressBytesGzip(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}
		gfListMap[tname] = gbuf
		return false, false, nil
	}

	for tname, _ := range c.Themes {
		skip, fail, rerr := cacheTheme(tname)
		if fail || rerr != nil {
			return skip, rerr
		}
	}

	hyperspace.topicList.Store(tListMap)
	hyperspace.gzipTopicList.Store(gtListMap)
	hyperspace.forumList.Store(fListMap)
	hyperspace.gzipForumList.Store(gfListMap)
	hyperspace.lastTopicListUpdate.Store(time.Now().Unix())

	return false, nil
}

func jumpHdriveTopicList(args ...interface{}) (skip bool, rerr c.RouteError) {
	theme := c.GetThemeByReq(args[1].(*http.Request))
	p := hyperspace.topicList.Load().(map[string][]byte)
	pg := hyperspace.gzipTopicList.Load().(map[string][]byte)
	return jumpHdrive(pg[theme.Name], p[theme.Name], args)
}

func jumpHdriveForumList(args ...interface{}) (skip bool, rerr c.RouteError) {
	theme := c.GetThemeByReq(args[1].(*http.Request))
	p := hyperspace.forumList.Load().(map[string][]byte)
	pg := hyperspace.gzipForumList.Load().(map[string][]byte)
	return jumpHdrive(pg[theme.Name], p[theme.Name], args)
}

func jumpHdrive(pg, p []byte, args []interface{}) (skip bool, rerr c.RouteError) {
	var tList []byte
	w := args[0].(http.ResponseWriter)
	var iw http.ResponseWriter
	gzw, ok := w.(c.GzipResponseWriter)
	if ok {
		tList = pg
		iw = gzw.ResponseWriter
	} else {
		tList = p
		iw = w
	}
	if len(tList) == 0 {
		c.DebugLog("no itemlist in hyperspace")
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

	var etag string
	lastUpdate := hyperspace.lastTopicListUpdate.Load().(int64)
	c.DebugLog("lastUpdate:", lastUpdate)
	if ok {
		iw.Header().Set("X-I", "1")
		etag = "\"" + strconv.FormatInt(lastUpdate, 10) + "-g\""
	} else {
		etag = "\"" + strconv.FormatInt(lastUpdate, 10) + "\""
	}

	if lastUpdate != 0 {
		iw.Header().Set("ETag", etag)
		if match := r.Header.Get("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				iw.WriteHeader(http.StatusNotModified)
				return true, nil
			}
		}
	}

	header := args[3].(*c.Header)
	if ok {
		gzw.Header().Set("Content-Type", "text/html;charset=utf-8")
	}
	routes.FootHeaders(w, header)
	iw.Write(tList)

	return true, nil
}
