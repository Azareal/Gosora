package routes

import (
	//"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	c "github.com/Azareal/Gosora/common"
	co "github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/uutils"
)

var successJSONBytes = []byte(`{"success":1}`)

func ParseSEOURL(urlBit string) (slug string, id int, err error) {
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	tid, err := strconv.Atoi(halves[1])
	return halves[0], tid, err
}

const slen1 = len("</s/>;rel=preload;as=script,") + 6
const slen2 = len("</s/>;rel=preload;as=style,") + 7

var pushStrPool = sync.Pool{}

func doPush(w http.ResponseWriter, h *c.Header) {
	//fmt.Println("in doPush")
	if len(h.Scripts) == 0 && len(h.ScriptsAsync) == 0 && len(h.Stylesheets) == 0 {
		return
	}
	if c.Config.EnableCDNPush {
		var sb *strings.Builder = &strings.Builder{}
		/*ii := pushStrPool.Get()
		if ii == nil {
			sb = &strings.Builder{}
		} else {
			sb = ii.(*strings.Builder)
			sb.Reset()
		}*/
		sb.Grow((slen1 * (len(h.Scripts) + len(h.ScriptsAsync))) + ((slen2 + 7) * len(h.Stylesheets)))
		push := func(in []c.HScript) {
			for i, s := range in {
				if i != 0 {
					sb.WriteString(",</s/")
				} else {
					sb.WriteString("</s/")
				}
				sb.WriteString(s.Name)
				sb.WriteString(">;rel=preload;as=script")
			}
		}
		push(h.Scripts)
		//push(h.PreScriptsAsync)
		push(h.ScriptsAsync)

		if len(h.Stylesheets) > 0 {
			for i, s := range h.Stylesheets {
				if i != 0 {
					sb.WriteString(",</s/")
				} else {
					sb.WriteString("</s/")
				}
				sb.WriteString(s.Name)
				sb.WriteString(">;rel=preload;as=style")
			}
		}
		// TODO: Push avatars?

		if sb.Len() > 0 {
			sbuf := sb.String()
			w.Header().Set("Link", sbuf)
			//pushStrPool.Put(sb)
		}
	} else if !c.Config.DisableServerPush {
		//fmt.Println("push enabled")
		/*if bzw, ok := w.(c.BrResponseWriter); ok {
			w = bzw.ResponseWriter
		} else */if gzw, ok := w.(c.GzipResponseWriter); ok {
			w = gzw.ResponseWriter
		}
		pusher, ok := w.(http.Pusher)
		if !ok {
			return
		}
		//panic("has pusher")
		//fmt.Println("has pusher")

		var sb *strings.Builder = &strings.Builder{}
		/*ii := pushStrPool.Get()
		if ii == nil {
			sb = &strings.Builder{}
		} else {
			sb = ii.(*strings.Builder)
			sb.Reset()
		}*/
		sb.Grow(6 * (len(h.Scripts) + len(h.ScriptsAsync) + len(h.Stylesheets)))
		push := func(in []c.HScript) {
			for _, s := range in {
				//fmt.Println("pushing /s/" + path)
				sb.WriteString("/s/")
				sb.WriteString(s.Name)
				err := pusher.Push(sb.String(), nil)
				if err != nil {
					break
				}
				sb.Reset()
			}
		}
		push(h.Scripts)
		//push(h.PreScriptsAsync)
		push(h.ScriptsAsync)
		push(h.Stylesheets)
		// TODO: Push avatars?
		//pushStrPool.Put(sb)
	}
}

func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) c.RouteError {
	return renderTemplate2(tmplName, tmplName, w, r, header, pi)
}

func renderTemplate2(tmplName, hookName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) c.RouteError {
	err := renderTemplate3(tmplName, tmplName, w, r, header, pi)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}

func FootHeaders(w http.ResponseWriter, h *c.Header) {
	// TODO: Only set video domain when there is a video on the page
	if !h.LooseCSP {
		he := w.Header()
		if c.Config.SslSchema {
			if h.ExternalMedia {
				he.Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'"+c.Config.ExtraCSPOrigins+"; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self' www.youtube-nocookie.com embed.nicovideo.jp;upgrade-insecure-requests")
			} else {
				he.Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'"+c.Config.ExtraCSPOrigins+"; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self';upgrade-insecure-requests")
			}
		} else {
			if h.ExternalMedia {
				he.Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'"+c.Config.ExtraCSPOrigins+"; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self' www.youtube-nocookie.com embed.nicovideo.jp")
			} else {
				he.Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'"+c.Config.ExtraCSPOrigins+"; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self'")
			}
		}
	}

	// Server pushes can backfire on certain browsers, so we want to make sure it's only triggered for ones where it'll help
	lastAgent := h.CurrentUser.LastAgent
	//fmt.Println("lastAgent:", lastAgent)
	if lastAgent == c.Chrome || lastAgent == c.Firefox {
		doPush(w, h)
	}
}

func renderTemplate3(tmplName, hookName string, w http.ResponseWriter, r *http.Request, h *c.Header, pi interface{}) error {
	s := h.Stylesheets
	h.Stylesheets = nil
	noDescSimpleBot := h.CurrentUser.LastAgent == c.SimpleBots[0] || h.CurrentUser.LastAgent == c.SimpleBots[1]
	var simpleBot bool
	for _, agent := range c.SimpleBots {
		if h.CurrentUser.LastAgent == agent {
			simpleBot = true
		}
	}
	inner := r.FormValue("i") == "1"
	if !inner && !simpleBot {
		c.PrepResources(h.CurrentUser, h, h.Theme)
		for _, ss := range s {
			h.Stylesheets = append(h.Stylesheets, ss)
		}
		h.AddScript("global.js")
		if h.CurrentUser.Loggedin {
			h.AddScriptAsync("member.js")
		}
	} else {
		h.CurrentUser.LastAgent = 0
	}

	if h.CurrentUser.Loggedin || inner || noDescSimpleBot {
		h.MetaDesc = ""
		h.OGDesc = ""
	} else if h.MetaDesc != "" && h.OGDesc == "" {
		h.OGDesc = h.MetaDesc
	}

	if !simpleBot {
		FootHeaders(w, h)
	} else {
		h.GoogSiteVerify = ""
	}
	if h.Zone != "error" {
		since := time.Duration(uutils.Nanotime() - h.StartedAt)
		if h.CurrentUser.IsAdmin {
			h.Elapsed1 = since.String()
		}
		co.PerfCounter.Push(since /*, false*/)
	}
	if c.RunPreRenderHook("pre_render_"+hookName, w, r, h.CurrentUser, pi) {
		return nil
	}
	/*defer func() {
		c.StrSlicePool.Put(h.Scripts)
		c.StrSlicePool.Put(h.PreScriptsAsync)
	}()*/
	return h.Theme.RunTmpl(tmplName, pi, w)
}

// TODO: Rename renderTemplate to RenderTemplate instead of using this hack to avoid breaking things
var RenderTemplate = renderTemplate3

func actionSuccess(w http.ResponseWriter, r *http.Request, dest string, js bool) c.RouteError {
	if !js {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	} else {
		_, _ = w.Write(successJSONBytes)
	}
	return nil
}
