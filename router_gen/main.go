/* WIP Under Construction */
package main

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type TmplVars struct {
	RouteList         []*RouteImpl
	RouteGroups       []*RouteGroup
	AllRouteNames     []RouteName
	AllRouteMap       map[string]int
	AllAgentNames     []string
	AllAgentMap       map[string]int
	AllAgentMarkNames []string
	AllAgentMarks     map[string]string
	AllOSNames        []string
	AllOSMap          map[string]int
}

type RouteName struct {
	Plain string
	Short string
}

func main() {
	log.Println("Generating the router...")

	// Load all the routes...
	r := &Router{}
	routes(r)

	var tmplVars = TmplVars{
		RouteList:   r.routeList,
		RouteGroups: r.routeGroups,
	}
	var allRouteNames []RouteName
	var allRouteMap = make(map[string]int)

	var out string
	var mapIt = func(name string) {
		allRouteNames = append(allRouteNames, RouteName{name, strings.Replace(name, "common.", "c.", -1)})
		allRouteMap[name] = len(allRouteNames) - 1
	}
	var countToIndents = func(indent int) (indentor string) {
		for i := 0; i < indent; i++ {
			indentor += "\t"
		}
		return indentor
	}
	var runBefore = func(runnables []Runnable, indent int) (out string) {
		var indentor = countToIndents(indent)
		if len(runnables) > 0 {
			for _, runnable := range runnables {
				if runnable.Literal {
					out += "\n\t" + indentor + runnable.Contents
				} else {
					out += "\n" + indentor + "err = c." + runnable.Contents + "(w,req,user)\n" +
						indentor + "if err != nil {\n" +
						indentor + "\treturn err\n" +
						indentor + "}\n" + indentor
				}
			}
		}
		return out
	}

	for _, route := range r.routeList {
		mapIt(route.Name)
		var end = len(route.Path) - 1
		out += "\n\t\tcase \"" + route.Path[0:end] + "\":"
		out += runBefore(route.RunBefore, 4)
		out += "\n\t\t\tcounters.RouteViewCounter.Bump(" + strconv.Itoa(allRouteMap[route.Name]) + ")"
		if !route.Action && !route.NoHead {
			out += "\n\t\t\thead, err := c.UserCheck(w,req,&user)"
			out += "\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}"
			vcpy := route.Vars
			route.Vars = []string{"head"}
			route.Vars = append(route.Vars, vcpy...)
		}
		out += "\n\t\t\terr = " + strings.Replace(route.Name, "common.", "c.", -1) + "(w,req,user"
		for _, item := range route.Vars {
			out += "," + item
		}
		out += `)`
	}

	for _, group := range r.routeGroups {
		var end = len(group.Path) - 1
		out += "\n\t\tcase \"" + group.Path[0:end] + "\":"
		out += runBefore(group.RunBefore, 3)
		out += "\n\t\t\tswitch(req.URL.Path) {"

		var defaultRoute = blankRoute()
		for _, route := range group.RouteList {
			if group.Path == route.Path {
				defaultRoute = route
				continue
			}
			mapIt(route.Name)

			out += "\n\t\t\t\tcase \"" + route.Path + "\":"
			if len(route.RunBefore) > 0 {
			skipRunnable:
				for _, runnable := range route.RunBefore {
					for _, gRunnable := range group.RunBefore {
						if gRunnable.Contents == runnable.Contents {
							continue
						}
						// TODO: Stop hard-coding these
						if gRunnable.Contents == "AdminOnly" && runnable.Contents == "MemberOnly" {
							continue skipRunnable
						}
						if gRunnable.Contents == "AdminOnly" && runnable.Contents == "SuperModOnly" {
							continue skipRunnable
						}
						if gRunnable.Contents == "SuperModOnly" && runnable.Contents == "MemberOnly" {
							continue skipRunnable
						}
					}

					if runnable.Literal {
						out += "\n\t\t\t\t\t" + runnable.Contents
					} else {
						out += `
					err = c.` + runnable.Contents + `(w,req,user)
					if err != nil {
						return err
					}
					`
					}
				}
			}
			out += "\n\t\t\t\t\tcounters.RouteViewCounter.Bump(" + strconv.Itoa(allRouteMap[route.Name]) + ")"
			if !route.Action && !route.NoHead && !group.NoHead {
				out += "\n\t\t\t\thead, err := c.UserCheck(w,req,&user)"
				out += "\n\t\t\t\tif err != nil {\n\t\t\t\t\treturn err\n\t\t\t\t}"
				vcpy := route.Vars
				route.Vars = []string{"head"}
				route.Vars = append(route.Vars, vcpy...)
			}
			out += "\n\t\t\t\t\terr = " + strings.Replace(route.Name, "common.", "c.", -1) + "(w,req,user"
			for _, item := range route.Vars {
				out += "," + item
			}
			out += ")"
		}

		if defaultRoute.Name != "" {
			mapIt(defaultRoute.Name)
			out += "\n\t\t\t\tdefault:"
			out += runBefore(defaultRoute.RunBefore, 4)
			out += "\n\t\t\t\t\tcounters.RouteViewCounter.Bump(" + strconv.Itoa(allRouteMap[defaultRoute.Name]) + ")"
			if !defaultRoute.Action && !defaultRoute.NoHead && !group.NoHead {
				out += "\n\t\t\t\t\thead, err := c.UserCheck(w,req,&user)"
				out += "\n\t\t\t\t\tif err != nil {\n\t\t\t\t\t\treturn err\n\t\t\t\t\t}"
				vcpy := defaultRoute.Vars
				defaultRoute.Vars = []string{"head"}
				defaultRoute.Vars = append(defaultRoute.Vars, vcpy...)
			}
			out += "\n\t\t\t\t\terr = " + strings.Replace(defaultRoute.Name, "common.", "c.", -1) + "(w,req,user"
			for _, item := range defaultRoute.Vars {
				out += ", " + item
			}
			out += ")"
		}
		out += `
			}`
	}

	// Stubs for us to refer to these routes through
	mapIt("routes.DynamicRoute")
	mapIt("routes.UploadedFile")
	mapIt("routes.StaticFile")
	mapIt("routes.RobotsTxt")
	mapIt("routes.SitemapXml")
	mapIt("routes.OpenSearchXml")
	mapIt("routes.BadRoute")
	mapIt("routes.HTTPSRedirect")
	tmplVars.AllRouteNames = allRouteNames
	tmplVars.AllRouteMap = allRouteMap

	tmplVars.AllOSNames = []string{
		"unknown",
		"windows",
		"linux",
		"mac",
		"android",
		"iphone",
	}
	tmplVars.AllOSMap = make(map[string]int)
	for id, os := range tmplVars.AllOSNames {
		tmplVars.AllOSMap[os] = id
	}

	tmplVars.AllAgentNames = []string{
		"unknown",
		"firefox",
		"chrome",
		"opera",
		"safari",
		"edge",
		"internetexplorer",
		"trident", // Hack to support IE11

		"androidchrome",
		"mobilesafari",
		"samsung",
		"ucbrowser",

		"googlebot",
		"yandex",
		"bing",
		"baidu",
		"duckduckgo",
		"seznambot",
		"discord",
		"twitter",
		"facebook",
		"cloudflare",
		"uptimebot",
		"slackbot",
		"apple",
		"discourse",
		"lynx",
		"blank",
		"malformed",
		"suspicious",
		"semrush",
		"dotbot",
		"zgrab",
	}

	tmplVars.AllAgentMap = make(map[string]int)
	for id, agent := range tmplVars.AllAgentNames {
		tmplVars.AllAgentMap[agent] = id
	}

	tmplVars.AllAgentMarkNames = []string{
		"OPR",
		"Chrome",
		"Firefox",
		"MSIE",
		"Trident",
		"Edge",
		"Lynx",
		"SamsungBrowser",
		"UCBrowser",

		"Google",
		"Googlebot",
		"yandex",
		"DuckDuckBot",
		"Baiduspider",
		"bingbot",
		"BingPreview",
		"SeznamBot",
		"CloudFlare",
		"Uptimebot",
		"Slackbot",
		"Discordbot",
		"Twitterbot",
		"facebookexternalhit",
		"Facebot",
		"Applebot",
		"Discourse",

		"SemrushBot",
		"DotBot",
		"zgrab",
	}

	tmplVars.AllAgentMarks = map[string]string{
		"OPR":            "opera",
		"Chrome":         "chrome",
		"Firefox":        "firefox",
		"MSIE":           "internetexplorer",
		"Trident":        "trident", // Hack to support IE11
		"Edge":           "edge",
		"Lynx":           "lynx", // There's a rare android variant of lynx which isn't covered by this
		"SamsungBrowser": "samsung",
		"UCBrowser":      "ucbrowser",

		"Google":              "googlebot",
		"Googlebot":           "googlebot",
		"yandex":              "yandex", // from the URL
		"DuckDuckBot":         "duckduckgo",
		"Baiduspider":         "baidu",
		"bingbot":             "bing",
		"BingPreview":         "bing",
		"SeznamBot":           "seznambot",
		"CloudFlare":          "cloudflare", // Track alwayson specifically in case there are other bots?
		"Uptimebot":           "uptimebot",
		"Slackbot":            "slackbot",
		"Discordbot":          "discord",
		"Twitterbot":          "twitter",
		"facebookexternalhit": "facebook",
		"Facebot":             "facebook",
		"Applebot":"apple",
		"Discourse":           "discourse",

		"SemrushBot": "semrush",
		"DotBot":     "dotbot",
		"zgrab":      "zgrab",
	}

	var fileData = `// Code generated by Gosora's Router Generator. DO NOT EDIT.
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main

import (
	"log"
	"strings"
	"bytes"
	"strconv"
	"compress/gzip"
	"sync"
	"sync/atomic"
	"errors"
	"os"
	"net/http"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/routes"
	"github.com/Azareal/Gosora/routes/panel"
)

var ErrNoRoute = errors.New("That route doesn't exist.")
// TODO: What about the /uploads/ route? x.x
var RouteMap = map[string]interface{}{ {{range .AllRouteNames}}
	"{{.Plain}}": {{.Short}},{{end}}
}

// ! NEVER RELY ON THESE REMAINING THE SAME BETWEEN COMMITS
var routeMapEnum = map[string]int{ {{range $index, $element := .AllRouteNames}}
	"{{$element.Plain}}": {{$index}},{{end}}
}
var reverseRouteMapEnum = map[int]string{ {{range $index, $element := .AllRouteNames}}
	{{$index}}: "{{$element.Plain}}",{{end}}
}
var osMapEnum = map[string]int{ {{range $index, $element := .AllOSNames}}
	"{{$element}}": {{$index}},{{end}}
}
var reverseOSMapEnum = map[int]string{ {{range $index, $element := .AllOSNames}}
	{{$index}}: "{{$element}}",{{end}}
}
var agentMapEnum = map[string]int{ {{range $index, $element := .AllAgentNames}}
	"{{$element}}": {{$index}},{{end}}
}
var reverseAgentMapEnum = map[int]string{ {{range $index, $element := .AllAgentNames}}
	{{$index}}: "{{$element}}",{{end}}
}
var markToAgent = map[string]string{ {{range $index, $element := .AllAgentMarkNames}}
	"{{$element}}": "{{index $.AllAgentMarks $element}}",{{end}}
}
/*var agentRank = map[string]int{
	"opera":9,
	"chrome":8,
	"safari":1,
}*/

// TODO: Stop spilling these into the package scope?
func init() {
	counters.SetRouteMapEnum(routeMapEnum)
	counters.SetReverseRouteMapEnum(reverseRouteMapEnum)
	counters.SetAgentMapEnum(agentMapEnum)
	counters.SetReverseAgentMapEnum(reverseAgentMapEnum)
	counters.SetOSMapEnum(osMapEnum)
	counters.SetReverseOSMapEnum(reverseOSMapEnum)
}

type WriterIntercept struct {
	http.ResponseWriter
}

func NewWriterIntercept(w http.ResponseWriter) *WriterIntercept {
	return &WriterIntercept{w}
}

var wiMaxAge = "max-age=" + strconv.Itoa(int(c.Day))
func (writ *WriterIntercept) WriteHeader(code int) {
	if code == 200 {
		h := writ.ResponseWriter.Header()
		h.Set("Cache-Control", wiMaxAge)
		h.Set("Vary", "Accept-Encoding")
	}
	writ.ResponseWriter.WriteHeader(code)
}

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	counters.RouteViewCounter.Bump({{ index .AllRouteMap "routes.HTTPSRedirect" }})
	dest := "https://" + req.Host + req.URL.String()
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

type GenRouter struct {
	UploadHandler func(http.ResponseWriter, *http.Request)
	extraRoutes map[string]func(http.ResponseWriter, *http.Request, c.User) c.RouteError
	requestLogger *log.Logger
	
	sync.RWMutex
}

func NewGenRouter(uploads http.Handler) (*GenRouter, error) {
	f, err := os.OpenFile("./logs/reqs-"+strconv.FormatInt(c.StartTime.Unix(),10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	return &GenRouter{
		UploadHandler: func(w http.ResponseWriter, req *http.Request) {
			writ := NewWriterIntercept(w)
			http.StripPrefix("/uploads/",uploads).ServeHTTP(writ,req)
		},
		extraRoutes: make(map[string]func(http.ResponseWriter, *http.Request, c.User) c.RouteError),
		requestLogger: log.New(f, "", log.LstdFlags),
	}, nil
}

func (r *GenRouter) handleError(err c.RouteError, w http.ResponseWriter, req *http.Request, user c.User) {
	if err.Handled() {
		return
	}
	
	if err.Type() == "system" {
		c.InternalErrorJSQ(err, w, req, err.JSON())
		return
	}
	c.LocalErrorJSQ(err.Error(), w, req, user, err.JSON())
}

func (r *GenRouter) Handle(_ string, _ http.Handler) {
}

func (r *GenRouter) HandleFunc(pattern string, handle func(http.ResponseWriter, *http.Request, c.User) c.RouteError) {
	r.Lock()
	defer r.Unlock()
	r.extraRoutes[pattern] = handle
}

func (r *GenRouter) RemoveFunc(pattern string) error {
	r.Lock()
	defer r.Unlock()
	_, ok := r.extraRoutes[pattern]
	if !ok {
		return ErrNoRoute
	}
	delete(r.extraRoutes, pattern)
	return nil
}

func (r *GenRouter) DumpRequest(req *http.Request, prepend string) {
	var heads string
	for key, value := range req.Header {
		for _, vvalue := range value {
			heads += "Header '" + c.SanitiseSingleLine(key) + "': " + c.SanitiseSingleLine(vvalue) + "!\n"
		}
	}

	r.requestLogger.Print(prepend + 
		"\nUA: " + c.SanitiseSingleLine(req.UserAgent()) + "\n" +
		"Method: " + c.SanitiseSingleLine(req.Method) + "\n" + heads + 
		"req.Host: " + c.SanitiseSingleLine(req.Host) + "\n" + 
		"req.URL.Path: " + c.SanitiseSingleLine(req.URL.Path) + "\n" + 
		"req.URL.RawQuery: " + c.SanitiseSingleLine(req.URL.RawQuery) + "\n" + 
		"req.Referer(): " + c.SanitiseSingleLine(req.Referer()) + "\n" + 
		"req.RemoteAddr: " + req.RemoteAddr + "\n")
}

func (r *GenRouter) SuspiciousRequest(req *http.Request, prepend string) {
	if prepend != "" {
		prepend += "\n"
	}
	r.DumpRequest(req,prepend+"Suspicious Request")
	counters.AgentViewCounter.Bump({{.AllAgentMap.suspicious}})
}

func isLocalHost(host string) bool {
	return host=="localhost" || host=="127.0.0.1" || host=="::1"
}

// TODO: Pass the default path or config struct to the router rather than accessing it via a package global
// TODO: SetDefaultPath
// TODO: GetDefaultPath
func (r *GenRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var malformedRequest = func(typ int) {
		w.WriteHeader(200) // 400
		w.Write([]byte(""))
		r.DumpRequest(req,"Malformed Request T"+strconv.Itoa(typ))
		counters.AgentViewCounter.Bump({{.AllAgentMap.malformed}})
	}
	
	// Split the Host and Port string
	var shost, sport string
	if req.Host[0]=='[' {
		spl := strings.Split(req.Host,"]")
		if len(spl) > 2 {
			malformedRequest(0)
			return
		}
		shost = strings.TrimPrefix(spl[0],"[")
		sport = strings.TrimPrefix(spl[1],":")
	} else {
		spl := strings.Split(req.Host,":")
		if len(spl) > 2 {
			malformedRequest(1)
			return
		}
		shost = spl[0]
		if len(spl)==2 {
			sport = spl[1]
		}
	}
	// TODO: Reject requests from non-local IPs, if the site host is set to localhost or a localhost IP
	if !c.Config.LoosePort && c.Site.PortInt != 80 && c.Site.PortInt != 443 && sport != c.Site.Port {
		malformedRequest(2)
		return
	}
	
	// Redirect www. and local IP requests to the right place
	if shost == "www." + c.Site.Host || (c.Site.LocalHost && shost != c.Site.Host && isLocalHost(shost)) {
		// TODO: Abstract the redirect logic?
		w.Header().Set("Connection", "close")
		var s string
		if c.Site.EnableSsl {
			s = "s"
		}
		var p string
		if c.Site.PortInt != 80 && c.Site.PortInt != 443 {
			p = ":"+c.Site.Port
		}
		dest := "http"+s+"://" + c.Site.Host+p + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			dest += "?" + req.URL.RawQuery
		}
		http.Redirect(w, req, dest, http.StatusMovedPermanently)
		return
	}

	// Deflect malformed requests
	if len(req.URL.Path) == 0 || req.URL.Path[0] != '/' || (!c.Config.LooseHost && shost != c.Site.Host) {
		malformedRequest(3)
		return
	}
	if c.Dev.FullReqLog {
		r.DumpRequest(req,"")
	}

	// TODO: Cover more suspicious strings and at a lower layer than this
	for _, char := range req.URL.Path {
		if char != '&' && !(char > 44 && char < 58) && char != '=' && char != '?' && !(char > 64 && char < 91) && char != '\\' && char != '_' && !(char > 96 && char < 123) {
			r.SuspiciousRequest(req,"Bad char in path")
			break
		}
	}
	lowerPath := strings.ToLower(req.URL.Path)
	// TODO: Flag any requests which has a dot with anything but a number after that
	if strings.Contains(req.URL.Path,"..") || strings.Contains(req.URL.Path,"--") || strings.Contains(lowerPath,".php") || strings.Contains(lowerPath,".asp") || strings.Contains(lowerPath,".cgi") || strings.Contains(lowerPath,".py") || strings.Contains(lowerPath,".sql") || strings.Contains(lowerPath,".action") {
		r.SuspiciousRequest(req,"Bad snippet in path")
	}

	// Indirect the default route onto a different one
	if req.URL.Path == "/" {
		req.URL.Path = c.Config.DefaultPath
	}
	
	var prefix, extraData string
	prefix = req.URL.Path[0:strings.IndexByte(req.URL.Path[1:],'/') + 1]
	if req.URL.Path[len(req.URL.Path) - 1] != '/' {
		extraData = req.URL.Path[strings.LastIndexByte(req.URL.Path,'/') + 1:]
		req.URL.Path = req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/') + 1]
	}

	// TODO: Use the same hook table as downstream
	hTbl := c.GetHookTable()
	skip, ferr := hTbl.VhookSkippable("router_after_filters", w, req, prefix, extraData)
	if skip || ferr != nil {
		return
	}

	if prefix != "/ws" {
		h := w.Header()
		h.Set("X-Frame-Options", "deny")
		h.Set("X-XSS-Protection", "1; mode=block") // TODO: Remove when we add a CSP? CSP's are horrendously glitchy things, tread with caution before removing
		h.Set("X-Content-Type-Options", "nosniff")
		if c.Config.RefNoRef || !c.Site.EnableSsl {
			h.Set("Referrer-Policy","no-referrer")
		} else {
			h.Set("Referrer-Policy","strict-origin")
		}
	}
	
	if c.Dev.SuperDebug {
		r.DumpRequest(req,"before routes.StaticFile")
	}
	// Increment the request counter
	counters.GlobalViewCounter.Bump()
	
	if prefix == "/s" { //old prefix: /static
		counters.RouteViewCounter.Bump({{ index .AllRouteMap "routes.StaticFile" }})
		req.URL.Path += extraData
		routes.StaticFile(w, req)
		return
	}
	if atomic.LoadInt32(&c.IsDBDown) == 1 {
		c.DatabaseError(w, req)
		return
	}
	if c.Dev.SuperDebug {
		r.requestLogger.Print("before PreRoute")
	}

	// Track the user agents. Unfortunately, everyone pretends to be Mozilla, so this'll be a little less efficient than I would like.
	// TODO: Add a setting to disable this?
	// TODO: Use a more efficient detector instead of smashing every possible combination in
	ua := strings.TrimSpace(strings.Replace(strings.TrimPrefix(req.UserAgent(),"Mozilla/5.0 ")," Safari/537.36","",-1)) // Noise, no one's going to be running this and it would require some sort of agent ranking system to determine which identifier should be prioritised over another
	var agent string
	if ua == "" {
		counters.AgentViewCounter.Bump({{.AllAgentMap.blank}})
		if c.Dev.DebugMode {
			var prepend string
			for _, char := range req.UserAgent() {
				prepend += strconv.Itoa(int(char)) + " "
			}
			r.DumpRequest(req,"Blank UA: " + prepend)
		}
	} else {		
		// WIP UA Parser
		var items []string
		var buffer []byte
		var os string
		for _, item := range StringToBytes(ua) {
			if (item > 64 && item < 91) || (item > 96 && item < 123) {
				buffer = append(buffer, item)
			} else if item == ' ' || item == '(' || item == ')' || item == '-' || (item > 47 && item < 58) || item == '_' || item == ';' || item == ':' || item == '.' || item == '+' || item == '~' || item == '@' || (item == ':' && bytes.Equal(buffer,[]byte("http"))) || item == ',' || item == '/' {
				if len(buffer) != 0 {
					if len(buffer) > 2 {
						// Use an unsafe zero copy conversion here just to use the switch, it's not safe for this string to escape from here, as it will get mutated, so do a regular string conversion in append
						switch(BytesToString(buffer)) {
						case "Windows":
							os = "windows"
						case "Linux":
							os = "linux"
						case "Mac":
							os = "mac"
						case "iPhone":
							os = "iphone"
						case "Android":
							os = "android"
						case "like","compatible":
							// Skip these words
						default:
							items = append(items, string(buffer))
						}
					}
					buffer = buffer[:0]
				}
			} else {
				// TODO: Test this
				items = items[:0]
				r.SuspiciousRequest(req,"Illegal char "+strconv.Itoa(int(item))+" in UA")
				r.requestLogger.Print("UA Buffer: ", buffer)
				r.requestLogger.Print("UA Buffer String: ", string(buffer))
				break
			}
		}
		if os == "" {
			os = "unknown"
		}

		// Iterate over this in reverse as the real UA tends to be on the right side
		for i := len(items) - 1; i >= 0; i-- {
			fAgent, ok := markToAgent[items[i]]
			if ok {
				agent = fAgent
				if agent != "safari" {
					break
				}
			}
		}
		if c.Dev.SuperDebug {
			r.requestLogger.Print("parsed agent: ", agent)
		}
		if c.Dev.SuperDebug {
			r.requestLogger.Print("os: ", os)
			r.requestLogger.Printf("items: %+v\n",items)
		}
		
		// Special handling
		switch(agent) {
		case "chrome":
			if os == "android" {
				agent = "androidchrome"
			}
		case "safari":
			if os == "iphone" {
				agent = "mobilesafari"
			}
		case "trident":
			// Hack to support IE11, change this after we start logging versions
			if strings.Contains(ua,"rv:11") {
				agent = "internetexplorer"
			}
		case "zgrab":
			r.SuspiciousRequest(req,"Vulnerability Scanner")
		}
		
		if agent == "" {
			counters.AgentViewCounter.Bump({{.AllAgentMap.unknown}})
			if c.Dev.DebugMode {
				var prepend string
				for _, char := range req.UserAgent() {
					prepend += strconv.Itoa(int(char)) + " "
				}
				r.DumpRequest(req,"Blank UA: " + prepend)
			}
		} else {
			counters.AgentViewCounter.Bump(agentMapEnum[agent])
		}
		counters.OSViewCounter.Bump(osMapEnum[os])
	}

	// TODO: Do we want to track missing language headers too? Maybe as it's own type, e.g. "noheader"?
	lang := req.Header.Get("Accept-Language")
	if lang != "" {
		lang = strings.TrimSpace(lang)
		lLang := strings.Split(lang,"-")
		tLang := strings.Split(strings.Split(lLang[0],";")[0],",")
		c.DebugDetail("tLang:", tLang)
		var llLang string
		for _, seg := range tLang {
			if seg == "*" {
				continue
			}
			llLang = seg
			break
		}
		c.DebugDetail("llLang:", llLang)
		if llLang == "" {
			counters.LangViewCounter.Bump("none")
		} else {
			validCode := counters.LangViewCounter.Bump(llLang)
			if !validCode {
				r.DumpRequest(req,"Invalid ISO Code")
			}
		}
	} else {
		counters.LangViewCounter.Bump("none")
	}

	if !c.Config.RefNoTrack {
		referrer := req.Header.Get("Referer") // Check the 'referrer' header too? :P
		if referrer != "" {
			// ? Optimise this a little?
			referrer = strings.TrimPrefix(strings.TrimPrefix(referrer,"http://"),"https://")
			referrer = strings.Split(referrer,"/")[0]
			portless := strings.Split(referrer,":")[0]
			if portless != "localhost" && portless != "127.0.0.1" && portless != c.Site.Host {
				counters.ReferrerTracker.Bump(referrer)
			}
		}
	}
	
	// Deal with the session stuff, etc.
	user, ok := c.PreRoute(w, req)
	if !ok {
		return
	}
	user.LastAgent = agent
	if c.Dev.SuperDebug {
		r.requestLogger.Print(
			"after PreRoute\n" +
			"routeMapEnum: ", routeMapEnum)
	}

	// Disable Gzip when SSL is disabled for security reasons?
	if prefix != "/ws" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		h := w.Header()
		h.Set("Content-Encoding", "gzip")
		h.Set("Content-Type", "text/html; charset=utf-8")
		gzw := c.GzipResponseWriter{Writer: gzip.NewWriter(w), ResponseWriter: w}
		defer func() {
			if h.Get("Content-Encoding") == "gzip" && h.Get("X-I") == "" {
				gzw.Writer.(*gzip.Writer).Close()
			}
		}()
		w = gzw
	}

	skip, ferr = hTbl.VhookSkippable("router_pre_route", w, req, user, prefix, extraData)
	if skip || ferr != nil {
		r.handleError(ferr,w,req,user)
	}
	ferr = r.routeSwitch(w, req, user, prefix, extraData)
	if ferr != nil {
		r.handleError(ferr,w,req,user)
	}

	hTbl.VhookNoRet("router_end", w, req, user, prefix, extraData)
	//c.StoppedServer("Profile end")
}
	
func (r *GenRouter) routeSwitch(w http.ResponseWriter, req *http.Request, user c.User, prefix string, extraData string) c.RouteError {
	var err c.RouteError
	switch(prefix) {` + out + `
		/*case "/sitemaps": // TODO: Count these views
			req.URL.Path += extraData
			err = sitemapSwitch(w,req)*/
		case "/uploads":
			if extraData == "" {
				return c.NotFound(w,req,nil)
			}
			gzw, ok := w.(c.GzipResponseWriter)
			if ok {
				w = gzw.ResponseWriter
				h := w.Header()
				h.Del("Content-Type")
				h.Del("Content-Encoding")
			}
			counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.UploadedFile" }})
			req.URL.Path += extraData
			// TODO: Find a way to propagate errors up from this?
			r.UploadHandler(w,req) // TODO: Count these views
			return nil
		case "":
			// Stop the favicons, robots.txt file, etc. resolving to the topics list
			// TODO: Add support for favicons and robots.txt files
			switch(extraData) {
				case "robots.txt":
					counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.RobotsTxt"}})
					return routes.RobotsTxt(w,req)
				case "favicon.ico":
					gzw, ok := w.(c.GzipResponseWriter)
					if ok {
						w = gzw.ResponseWriter
						h := w.Header()
						h.Del("Content-Type")
						h.Del("Content-Encoding")
					}
					req.URL.Path = "/s/favicon.ico"
					routes.StaticFile(w,req)
					return nil
				case "opensearch.xml":
					counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.OpenSearchXml"}})
					return routes.OpenSearchXml(w,req)
				/*case "sitemap.xml":
					counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.SitemapXml"}})
					return routes.SitemapXml(w,req)*/
			}
			return c.NotFound(w,req,nil)
		default:
			// A fallback for dynamic routes, e.g. ones declared by plugins
			r.RLock()
			handle, ok := r.extraRoutes[req.URL.Path]
			r.RUnlock()
			
			if ok {
				counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.DynamicRoute" }}) // TODO: Be more specific about *which* dynamic route it is
				req.URL.Path += extraData
				return handle(w,req,user)
			}

			lowerPath := strings.ToLower(req.URL.Path)
			if strings.Contains(lowerPath,"admin") || strings.Contains(lowerPath,"sql") || strings.Contains(lowerPath,"manage") || strings.Contains(lowerPath,"//") || strings.Contains(lowerPath,"\\\\") || strings.Contains(lowerPath,"wp") || strings.Contains(lowerPath,"wordpress") || strings.Contains(lowerPath,"config") || strings.Contains(lowerPath,"setup") || strings.Contains(lowerPath,"install") || strings.Contains(lowerPath,"update") || strings.Contains(lowerPath,"php") {
				r.SuspiciousRequest(req,"Bad Route")
			} else {
				r.DumpRequest(req,"Bad Route")
			}
			counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.BadRoute" }})
			return c.NotFound(w,req,nil)
	}
	return err
}
`
	var tmpl = template.Must(template.New("router").Parse(fileData))
	var b bytes.Buffer
	err := tmpl.Execute(&b, tmplVars)
	if err != nil {
		log.Fatal(err)
	}

	writeFile("./gen_router.go", string(b.Bytes()))
	log.Println("Successfully generated the router")
}

func writeFile(name string, content string) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
	f.Close()
}
