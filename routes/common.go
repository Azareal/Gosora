package routes

import (
	//"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	c "github.com/Azareal/Gosora/common"
)

var successJSONBytes = []byte(`{"success":"1"}`)

func ParseSEOURL(urlBit string) (slug string, id int, err error) {
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	tid, err := strconv.Atoi(halves[1])
	return halves[0], tid, err
}

var slen1 = len("</static/>; rel=preload; as=script,")
var slen2 = len("</static/>; rel=preload; as=style,")

func doPush(w http.ResponseWriter, header *c.Header) {
	//fmt.Println("in doPush")
	if c.Config.EnableCDNPush {
		// TODO: Cache these in a sync.Pool?
		var sb strings.Builder
		var push = func(in []string) {
			sb.Grow((slen1 + 5) * len(in))
			for _, path := range in {
				sb.WriteString("</static/")
				sb.WriteString(path)
				sb.WriteString(">; rel=preload; as=script,")
			}
		}
		push(header.Scripts)
		//push(header.PreScriptsAsync)
		push(header.ScriptsAsync)

		if len(header.Stylesheets) > 0 {
			sb.Grow((slen2 + 6) * len(header.Stylesheets))
			for _, path := range header.Stylesheets {
				sb.WriteString("</static/")
				sb.WriteString(path)
				sb.WriteString(">; rel=preload; as=style,")
			}
		}
		// TODO: Push avatars?

		if sb.Len() > 0 {
			sbuf := sb.String()
			w.Header().Set("Link", sbuf[:len(sbuf)-1])
		}
	} else if !c.Config.DisableServerPush {
		//fmt.Println("push enabled")
		gzw, ok := w.(c.GzipResponseWriter)
		if ok {
			w = gzw.ResponseWriter
		}
		pusher, ok := w.(http.Pusher)
		if !ok {
			return
		}
		//fmt.Println("has pusher")

		var push = func(in []string) {
			for _, path := range in {
				//fmt.Println("pushing /static/" + path)
				err := pusher.Push("/static/"+path, nil)
				if err != nil {
					break
				}
			}
		}
		push(header.Scripts)
		//push(header.PreScriptsAsync)
		push(header.ScriptsAsync)
		push(header.Stylesheets)
		// TODO: Push avatars?
	}
}

func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) c.RouteError {
	return renderTemplate2(tmplName, tmplName, w, r, header, pi)
}

func renderTemplate2(tmplName string, hookName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) c.RouteError {
	err := renderTemplate3(tmplName, tmplName, w, r, header, pi)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}

func renderTemplate3(tmplName string, hookName string, w http.ResponseWriter, r *http.Request, header *c.Header, pi interface{}) error {
	c.PrepResources(&header.CurrentUser, header, header.Theme)

	if header.CurrentUser.Loggedin {
		header.MetaDesc = ""
		header.OGDesc = ""
	} else if header.MetaDesc != "" && header.OGDesc == "" {
		header.OGDesc = header.MetaDesc
	}
	// TODO: Expand this to non-HTTPS requests too
	if !header.LooseCSP && c.Site.EnableSsl {
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self' www.youtube-nocookie.com;upgrade-insecure-requests")
	}
	header.AddScript("global.js")

	// Server pushes can backfire on certain browsers, so we want to make sure it's only triggered for ones where it'll help
	lastAgent := header.CurrentUser.LastAgent
	//fmt.Println("lastAgent:", lastAgent)
	if lastAgent == "chrome" || lastAgent == "firefox" {
		doPush(w, header)
	}

	if header.CurrentUser.IsAdmin {
		header.Elapsed1 = time.Since(header.StartedAt).String()
	}
	if c.RunPreRenderHook("pre_render_"+hookName, w, r, &header.CurrentUser, pi) {
		return nil
	}
	return header.Theme.RunTmpl(tmplName, pi, w)
}

// TODO: Rename renderTemplate to RenderTemplate instead of using this hack to avoid breaking things
var RenderTemplate = renderTemplate3
