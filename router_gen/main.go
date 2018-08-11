/* WIP Under Construction */
package main

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"text/template"
)

var routeList []*RouteImpl
var routeGroups []*RouteGroup

type TmplVars struct {
	RouteList     []*RouteImpl
	RouteGroups   []*RouteGroup
	AllRouteNames []string
	AllRouteMap   map[string]int
	AllAgentNames []string
	AllAgentMap   map[string]int
	AllOSNames    []string
	AllOSMap      map[string]int
}

func main() {
	log.Println("Generating the router...")

	// Load all the routes...
	routes()

	var tmplVars = TmplVars{
		RouteList:   routeList,
		RouteGroups: routeGroups,
	}
	var allRouteNames []string
	var allRouteMap = make(map[string]int)

	var out string
	var mapIt = func(name string) {
		allRouteNames = append(allRouteNames, name)
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
					out += "\n" + indentor + "err = common." + runnable.Contents + "(w,req,user)\n" +
						indentor + "if err != nil {\n" +
						indentor + "\trouter.handleError(err,w,req,user)\n" +
						indentor + "\treturn\n" +
						indentor + "}\n" + indentor
				}
			}
		}
		return out
	}

	for _, route := range routeList {
		mapIt(route.Name)
		var end = len(route.Path) - 1
		out += "\n\t\tcase \"" + route.Path[0:end] + "\":"
		out += runBefore(route.RunBefore, 4)
		out += "\n\t\t\tcounters.RouteViewCounter.Bump(" + strconv.Itoa(allRouteMap[route.Name]) + ")"
		out += "\n\t\t\terr = " + route.Name + "(w,req,user"
		for _, item := range route.Vars {
			out += "," + item
		}
		out += `)
			if err != nil {
				router.handleError(err,w,req,user)
			}`
	}

	for _, group := range routeGroups {
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
					err = common.` + runnable.Contents + `(w,req,user)
					if err != nil {
						router.handleError(err,w,req,user)
						return
					}
					`
					}
				}
			}
			out += "\n\t\t\t\t\tcounters.RouteViewCounter.Bump(" + strconv.Itoa(allRouteMap[route.Name]) + ")"
			out += "\n\t\t\t\t\terr = " + route.Name + "(w,req,user"
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
			out += "\n\t\t\t\t\terr = " + defaultRoute.Name + "(w,req,user"
			for _, item := range defaultRoute.Vars {
				out += ", " + item
			}
			out += ")"
		}
		out += `
			}
			if err != nil {
				router.handleError(err,w,req,user)
			}`
	}

	// Stubs for us to refer to these routes through
	mapIt("routes.DynamicRoute")
	mapIt("routes.UploadedFile")
	mapIt("routes.StaticFile")
	mapIt("routes.RobotsTxt")
	mapIt("routes.SitemapXml")
	mapIt("routes.BadRoute")
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
		"cloudflare",
		"uptimebot",
		"slackbot",
		"discourse",
		"lynx",
		"blank",
		"malformed",
		"suspicious",
		"zgrab",
	}

	tmplVars.AllAgentMap = make(map[string]int)
	for id, agent := range tmplVars.AllAgentNames {
		tmplVars.AllAgentMap[agent] = id
	}

	var fileData = `// Code generated by Gosora's Router Generator. DO NOT EDIT.
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main

import (
	"log"
	"strings"
	"strconv"
	"sync"
	"sync/atomic"
	"errors"
	"os"
	"net/http"

	"./common"
	"./common/counters"
	"./routes"
	"./routes/panel"
)

var ErrNoRoute = errors.New("That route doesn't exist.")
// TODO: What about the /uploads/ route? x.x
var RouteMap = map[string]interface{}{ {{range .AllRouteNames}}
	"{{.}}": {{.}},{{end}}
}

// ! NEVER RELY ON THESE REMAINING THE SAME BETWEEN COMMITS
var routeMapEnum = map[string]int{ {{range $index, $element := .AllRouteNames}}
	"{{$element}}": {{$index}},{{end}}
}
var reverseRouteMapEnum = map[int]string{ {{range $index, $element := .AllRouteNames}}
	{{$index}}: "{{$element}}",{{end}}
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
var markToAgent = map[string]string{
	"OPR":"opera",
	"Chrome":"chrome",
	"Firefox":"firefox",
	"MSIE":"internetexplorer",
	"Trident":"trident", // Hack to support IE11
	"Edge":"edge",
	"Lynx":"lynx", // There's a rare android variant of lynx which isn't covered by this
	"SamsungBrowser":"samsung",
	"UCBrowser":"ucbrowser",

	"Google":"googlebot",
	"Googlebot":"googlebot",
	"yandex": "yandex", // from the URL
	"DuckDuckBot":"duckduckgo",
	"Baiduspider":"baidu",
	"bingbot":"bing",
	"BingPreview":"bing",
	"SeznamBot":"seznambot",
	"CloudFlare":"cloudflare", // Track alwayson specifically in case there are other bots?
	"Uptimebot":"uptimebot",
	"Slackbot":"slackbot",
	"Discordbot":"discord",
	"Twitterbot":"twitter",
	"Discourse":"discourse",

	"zgrab":"zgrab",
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
	w http.ResponseWriter
	code int
}

func NewWriterIntercept(w http.ResponseWriter) *WriterIntercept {
	return &WriterIntercept{w:w,code:200}
}

func (writ *WriterIntercept) Header() http.Header {
	return writ.w.Header()
}

func (writ *WriterIntercept) Write(pieces []byte) (int, error) {
	return writ.w.Write(pieces)
}

func (writ *WriterIntercept) WriteHeader(code int) {
	writ.w.WriteHeader(code)
	writ.code = code
}

func (writ *WriterIntercept) GetCode() int {
	return writ.code
}

type GenRouter struct {
	UploadHandler func(http.ResponseWriter, *http.Request)
	extraRoutes map[string]func(http.ResponseWriter, *http.Request, common.User) common.RouteError
	requestLogger *log.Logger
	
	sync.RWMutex
}

func NewGenRouter(uploads http.Handler) (*GenRouter, error) {
	f, err := os.OpenFile("./logs/requests.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	return &GenRouter{
		UploadHandler: func(w http.ResponseWriter, req *http.Request) {
			writ := NewWriterIntercept(w)
			http.StripPrefix("/uploads/",uploads).ServeHTTP(writ,req)
			if writ.GetCode() == 200 {
				w.Header().Set("Cache-Control", "max-age=" + strconv.Itoa(int(common.Day)))
				w.Header().Set("Vary", "Accept-Encoding")
			} 
		},
		extraRoutes: make(map[string]func(http.ResponseWriter, *http.Request, common.User) common.RouteError),
		requestLogger: log.New(f, "", log.LstdFlags),
	}, nil
}

func (router *GenRouter) handleError(err common.RouteError, w http.ResponseWriter, r *http.Request, user common.User) {
	if err.Handled() {
		return
	}
	
	if err.Type() == "system" {
		common.InternalErrorJSQ(err, w, r, err.JSON())
		return
	}
	common.LocalErrorJSQ(err.Error(), w, r, user,err.JSON())
}

func (router *GenRouter) Handle(_ string, _ http.Handler) {
}

func (router *GenRouter) HandleFunc(pattern string, handle func(http.ResponseWriter, *http.Request, common.User) common.RouteError) {
	router.Lock()
	defer router.Unlock()
	router.extraRoutes[pattern] = handle
}

func (router *GenRouter) RemoveFunc(pattern string) error {
	router.Lock()
	defer router.Unlock()
	_, ok := router.extraRoutes[pattern]
	if !ok {
		return ErrNoRoute
	}
	delete(router.extraRoutes, pattern)
	return nil
}

func (router *GenRouter) DumpRequest(req *http.Request, prepend string) {
	var heads string
	for key, value := range req.Header {
		for _, vvalue := range value {
			heads += "Header '" + common.SanitiseSingleLine(key) + "': " + common.SanitiseSingleLine(vvalue) + "!!\n"
		}
	}

	router.requestLogger.Print(prepend + 
		"\nUA: " + common.SanitiseSingleLine(req.UserAgent()) + "\n" +
		"Method: " + common.SanitiseSingleLine(req.Method) + "\n" + heads + 
		"req.Host: " + common.SanitiseSingleLine(req.Host) + "\n" + 
		"req.URL.Path: " + common.SanitiseSingleLine(req.URL.Path) + "\n" + 
		"req.URL.RawQuery: " + common.SanitiseSingleLine(req.URL.RawQuery) + "\n" + 
		"req.Referer(): " + common.SanitiseSingleLine(req.Referer()) + "\n" + 
		"req.RemoteAddr: " + req.RemoteAddr + "\n")
}

func (router *GenRouter) SuspiciousRequest(req *http.Request, prepend string) {
	if prepend != "" {
		prepend += "\n"
	}
	router.DumpRequest(req,prepend+"Suspicious Request")
	counters.AgentViewCounter.Bump({{.AllAgentMap.suspicious}})
}

// TODO: Pass the default path or config struct to the router rather than accessing it via a package global
// TODO: SetDefaultPath
// TODO: GetDefaultPath
func (router *GenRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Redirect www. requests to the right place
	if req.Host == "www." + common.Site.Host {
		w.Header().Set("Connection", "close")
		var s string
		if common.Site.EnableSsl {
			s = "s"
		}
		dest := "http"+s+"://" + req.Host + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			dest += "?" + req.URL.RawQuery
		}
		http.Redirect(w, req, dest, http.StatusMovedPermanently)
		return
	}

	// Deflect malformed requests
	if len(req.URL.Path) == 0 || req.URL.Path[0] != '/' || req.Host != common.Site.Host {
		w.WriteHeader(200) // 400
		w.Write([]byte(""))
		router.DumpRequest(req,"Malformed Request")
		counters.AgentViewCounter.Bump({{.AllAgentMap.malformed}})
		return
	}

	// TODO: Cover more suspicious strings and at a lower layer than this
	for _, char := range req.URL.Path {
		if char != '&' && !(char > 44 && char < 58) && char != '=' && char != '?' && !(char > 64 && char < 91) && char != '\\' && char != '_' && !(char > 96 && char < 123) {
			router.SuspiciousRequest(req,"Bad char in path")
			break
		}
	}
	lowerPath := strings.ToLower(req.URL.Path)
	// TODO: Flag any requests which has a dot with anything but a number after that
	if strings.Contains(req.URL.Path,"..") || strings.Contains(req.URL.Path,"--") || strings.Contains(lowerPath,".php") || strings.Contains(lowerPath,".asp") || strings.Contains(lowerPath,".cgi") || strings.Contains(lowerPath,".py") || strings.Contains(lowerPath,".sql") || strings.Contains(lowerPath,".action") {
		router.SuspiciousRequest(req,"Bad snippet in path")
	}

	// Indirect the default route onto a different one
	if req.URL.Path == "/" {
		req.URL.Path = common.Config.DefaultPath
	}
	
	var prefix, extraData string
	prefix = req.URL.Path[0:strings.IndexByte(req.URL.Path[1:],'/') + 1]
	if req.URL.Path[len(req.URL.Path) - 1] != '/' {
		extraData = req.URL.Path[strings.LastIndexByte(req.URL.Path,'/') + 1:]
		req.URL.Path = req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/') + 1]
	}
	
	if common.Dev.SuperDebug {
		router.DumpRequest(req,"before routes.StaticFile")
	}
	// Increment the request counter
	counters.GlobalViewCounter.Bump()
	
	if prefix == "/static" {
		counters.RouteViewCounter.Bump({{ index .AllRouteMap "routes.StaticFile" }})
		req.URL.Path += extraData
		routes.StaticFile(w, req)
		return
	}
	if atomic.LoadInt32(&common.IsDBDown) == 1 {
		common.DatabaseError(w, req)
		return
	}
	if common.Dev.SuperDebug {
		router.requestLogger.Print("before PreRoute")
	}

	// Track the user agents. Unfortunately, everyone pretends to be Mozilla, so this'll be a little less efficient than I would like.
	// TODO: Add a setting to disable this?
	// TODO: Use a more efficient detector instead of smashing every possible combination in
	ua := strings.TrimSpace(strings.Replace(strings.TrimPrefix(req.UserAgent(),"Mozilla/5.0 ")," Safari/537.36","",-1)) // Noise, no one's going to be running this and it would require some sort of agent ranking system to determine which identifier should be prioritised over another
	if ua == "" {
		counters.AgentViewCounter.Bump({{.AllAgentMap.blank}})
		if common.Dev.DebugMode {
			var prepend string
			for _, char := range req.UserAgent() {
				prepend += strconv.Itoa(int(char)) + " "
			}
			router.DumpRequest(req,"Blank UA: " + prepend)
		}
	} else {
		var runeEquals = func(a []rune, b []rune) bool {
			if len(a) != len(b) {
				return false
			}
			for i, item := range a {
				if item != b[i] {
					return false
				}
			}
			return true
		}
		
		// WIP UA Parser
		var indices []int
		var items []string
		var buffer []rune
		for index, item := range ua {
			if (item > 64 && item < 91) || (item > 96 && item < 123) {
				buffer = append(buffer, item)
			} else if item == ' ' || item == '(' || item == ')' || item == '-' || (item > 47 && item < 58) || item == '_' || item == ';' || item == '.' || item == '+' || (item == ':' && (runeEquals(buffer,[]rune("http")) || runeEquals(buffer,[]rune("rv")))) || item == ',' || item == '/' {
				if len(buffer) != 0 {
					items = append(items, string(buffer))
					indices = append(indices, index - 1)
					buffer = buffer[:0]
				}
			} else {
				// TODO: Test this
				items = items[:0]
				indices = indices[:0]
				router.SuspiciousRequest(req,"Illegal char in UA")
				router.requestLogger.Print("UA Buffer: ", buffer)
				router.requestLogger.Print("UA Buffer String: ", string(buffer))
				break
			}
		}

		// Iterate over this in reverse as the real UA tends to be on the right side
		var agent string
		for i := len(items) - 1; i >= 0; i-- {
			fAgent, ok := markToAgent[items[i]]
			if ok {
				agent = fAgent
				if agent != "safari" {
					break
				}
			}
		}
		if common.Dev.SuperDebug {
			router.requestLogger.Print("parsed agent: ", agent)
		}

		var os string
		for _, mark := range items {
			switch(mark) {
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
			}
		}
		if os == "" {
			os = "unknown"
		}
		if common.Dev.SuperDebug {
			router.requestLogger.Print("os: ", os)
			router.requestLogger.Printf("items: %+v\n",items)
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
			router.SuspiciousRequest(req,"Vulnerability Scanner")
		}
		
		if agent == "" {
			counters.AgentViewCounter.Bump({{.AllAgentMap.unknown}})
			if common.Dev.DebugMode {
				var prepend string
				for _, char := range req.UserAgent() {
					prepend += strconv.Itoa(int(char)) + " "
				}
				router.DumpRequest(req,"Blank UA: " + prepend)
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
		common.DebugDetail("lLang:", lLang)
		counters.LangViewCounter.Bump(lLang[0])
	} else {
		counters.LangViewCounter.Bump("none")
	}

	referrer := req.Header.Get("Referer") // Check the 'referrer' header too? :P
	if referrer != "" {
		// ? Optimise this a little?
		referrer = strings.TrimPrefix(strings.TrimPrefix(referrer,"http://"),"https://")
		referrer = strings.Split(referrer,"/")[0]
		portless := strings.Split(referrer,":")[0]
		if portless != "localhost" && portless != "127.0.0.1" && portless != common.Site.Host {
			counters.ReferrerTracker.Bump(referrer)
		}
	}
	
	// Deal with the session stuff, etc.
	user, ok := common.PreRoute(w, req)
	if !ok {
		return
	}
	if common.Dev.SuperDebug {
		router.requestLogger.Print(
			"after PreRoute\n" +
			"routeMapEnum: ", routeMapEnum)
	}

	router.routeSwitch(w, req, user, prefix, extraData)
}
	
func (router *GenRouter) routeSwitch(w http.ResponseWriter, req *http.Request, user common.User, prefix string, extraData string) {
	var err common.RouteError
	switch(prefix) {` + out + `
		/*case "/sitemaps": // TODO: Count these views
			req.URL.Path += extraData
			err = sitemapSwitch(w,req)
			if err != nil {
				router.handleError(err,w,req,user)
			}*/
		case "/uploads":
			if extraData == "" {
				common.NotFound(w,req,nil)
				return
			}
			counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.UploadedFile" }})
			req.URL.Path += extraData
			// TODO: Find a way to propagate errors up from this?
			router.UploadHandler(w,req) // TODO: Count these views
		case "":
			// Stop the favicons, robots.txt file, etc. resolving to the topics list
			// TODO: Add support for favicons and robots.txt files
			switch(extraData) {
				case "robots.txt":
					counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.RobotsTxt"}})
					err = routes.RobotsTxt(w,req)
					if err != nil {
						router.handleError(err,w,req,user)
					}
					return
				/*case "sitemap.xml":
					counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.SitemapXml"}})
					err = routes.SitemapXml(w,req)
					if err != nil {
						router.handleError(err,w,req,user)
					}
					return*/
			}
			common.NotFound(w,req,nil)
			return
		default:
			// A fallback for the routes which haven't been converted to the new router yet or plugins
			router.RLock()
			handle, ok := router.extraRoutes[req.URL.Path]
			router.RUnlock()
			
			if ok {
				counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.DynamicRoute" }}) // TODO: Be more specific about *which* dynamic route it is
				req.URL.Path += extraData
				err = handle(w,req,user)
				if err != nil {
					router.handleError(err,w,req,user)
				}
				return
			}

			lowerPath := strings.ToLower(req.URL.Path)
			if strings.Contains(lowerPath,"admin") || strings.Contains(lowerPath,"sql") || strings.Contains(lowerPath,"manage") || strings.Contains(lowerPath,"//") || strings.Contains(lowerPath,"\\\\") || strings.Contains(lowerPath,"wp") || strings.Contains(lowerPath,"wordpress") || strings.Contains(lowerPath,"config") || strings.Contains(lowerPath,"setup") || strings.Contains(lowerPath,"install") || strings.Contains(lowerPath,"update") || strings.Contains(lowerPath,"php") {
				router.SuspiciousRequest(req,"Bad Route")
			} else {
				router.DumpRequest(req,"Bad Route")
			}
			counters.RouteViewCounter.Bump({{index .AllRouteMap "routes.BadRoute" }})
			common.NotFound(w,req,nil)
	}
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
