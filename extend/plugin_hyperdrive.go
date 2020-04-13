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
	forumList           atomic.Value
	lastTopicListUpdate atomic.Value
}

func newHyperspace() *Hyperspace {
	pageCache := new(Hyperspace)
	blank := make(map[string][3][]byte, len(c.Themes))
	pageCache.topicList.Store(blank)
	pageCache.forumList.Store(blank)
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
	blank := make(map[string][3][]byte, len(c.Themes))
	hyperspace.topicList.Store(blank)
	hyperspace.forumList.Store(blank)

	tListMap := make(map[string][3][]byte)
	fListMap := make(map[string][3][]byte)

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

		gbuf, err := c.CompressBytesGzip(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}

		bbuf, err := c.CompressBytesBrotli(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}
		tListMap[tname] = [3][]byte{buf.Bytes(), gbuf, bbuf}

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

		gbuf, err = c.CompressBytesGzip(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}

		bbuf, err = c.CompressBytesBrotli(buf.Bytes())
		if err != nil {
			c.LogWarning(err)
			return false, true, nil
		}
		fListMap[tname] = [3][]byte{buf.Bytes(), gbuf, bbuf}
		return false, false, nil
	}

	for tname, _ := range c.Themes {
		skip, fail, rerr := cacheTheme(tname)
		if fail || rerr != nil {
			return skip, rerr
		}
	}

	hyperspace.topicList.Store(tListMap)
	hyperspace.forumList.Store(fListMap)
	hyperspace.lastTopicListUpdate.Store(time.Now().Unix())

	return false, nil
}

func jumpHdriveTopicList(args ...interface{}) (skip bool, rerr c.RouteError) {
	theme := c.GetThemeByReq(args[1].(*http.Request))
	p := hyperspace.topicList.Load().(map[string][3][]byte)
	return jumpHdrive(p[theme.Name], args)
}

func jumpHdriveForumList(args ...interface{}) (skip bool, rerr c.RouteError) {
	theme := c.GetThemeByReq(args[1].(*http.Request))
	p := hyperspace.forumList.Load().(map[string][3][]byte)
	return jumpHdrive(p[theme.Name], args)
}

func jumpHdrive( /*pg, */ p [3][]byte, args []interface{}) (skip bool, rerr c.RouteError) {
	var tList []byte
	w := args[0].(http.ResponseWriter)
	r := args[1].(*http.Request)
	var iw http.ResponseWriter
	gzw, ok := w.(c.GzipResponseWriter)
	//bzw, ok2 := w.(c.BrResponseWriter)
	// !temp until global brotli
	br := strings.Contains(r.Header.Get("Accept-Encoding"), "br")
	if br && ok {
		tList = p[2]
		iw = gzw.ResponseWriter
	} else if br {
		tList = p[2]
		iw = w
	} else if ok {
		tList = p[1]
		iw = gzw.ResponseWriter
		/*} else if ok2 {
		tList = p[2]
		iw = bzw.ResponseWriter
		*/
	} else {
		tList = p[0]
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
	//c.DebugLog("r.URL.Path:",r.URL.Path)
	//c.DebugLog("r.URL.RawQuery:",r.URL.RawQuery)
	if r.URL.RawQuery != "" {
		return false, nil
	}
	if r.FormValue("js") == "1" || r.FormValue("i") == "1" {
		return false, nil
	}
	c.DebugLog("Successful jump")

	var etag string
	lastUpdate := hyperspace.lastTopicListUpdate.Load().(int64)
	c.DebugLog("lastUpdate:", lastUpdate)
	if br {
		h := iw.Header()
		h.Set("X-I", "1")
		h.Set("Content-Encoding", "br")
		etag = "\"" + strconv.FormatInt(lastUpdate, 10) + "-b\""
	} else if ok {
		iw.Header().Set("X-I", "1")
		etag = "\"" + strconv.FormatInt(lastUpdate, 10) + "-g\""
		/*} else if ok2 {
		iw.Header().Set("X-I", "1")
		etag = "\"" + strconv.FormatInt(lastUpdate, 10) + "-b\""
		*/
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
	if br || ok /*ok2*/ {
		iw.Header().Set("Content-Type", "text/html;charset=utf-8")
	}
	routes.FootHeaders(w, header)
	iw.Write(tList)

	return true, nil
}
