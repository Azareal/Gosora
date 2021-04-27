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
	AllAgentMarkIDs   map[string]int
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

	tmplVars := TmplVars{
		RouteList:   r.routeList,
		RouteGroups: r.routeGroups,
	}
	var allRouteNames []RouteName
	allRouteMap := make(map[string]int)

	var out string
	mapIt := func(name string) {
		allRouteNames = append(allRouteNames, RouteName{name, strings.Replace(name, "common.", "c.", -1)})
		allRouteMap[name] = len(allRouteNames) - 1
	}
	mapIt("routes.Error")

	var indentCache [20]string
	countToIndents := func(ind int) string {
		out := indentCache[ind]
		if out != "" {
			return out
		}
		for i := 0; i < ind; i++ {
			out += "\t"
		}
		if ind < 20 {
			indentCache[ind] = out
		}
		return out
	}
	o := func(indent int, str string) {
		out += countToIndents(indent) + str
	}
	on := func(indent int, str string) {
		out += "\n" + countToIndents(indent) + str
	}
	iferrn := func(indent int) {
		ind := countToIndents(indent)
		ind2 := countToIndents(indent + 1)
		out += "\n" + ind + "if err != nil {"
		out += "\n" + ind2 + "return err\n" + ind + "}"
	}

	runBefore := func(runnables []Runnable, ind int) {
		if len(runnables) > 0 {
			for _, runnable := range runnables {
				if runnable.Literal {
					on(ind, runnable.Contents)
				} else {
					on(ind, "err = c."+runnable.Contents+"(w,req,user)")
					iferrn(ind)
					o(ind, "\n")
				}
			}
		}
	}
	userCheckNano := func(indent int, route *RouteImpl) {
		on(indent, "h, err := c.UserCheckNano(w,req,user,cn)")
		iferrn(indent)
		vcpy := route.Vars
		route.Vars = []string{"h"}
		route.Vars = append(route.Vars, vcpy...)
	}
	writeRoute := func(indent int, r *RouteImpl) {
		on(indent, "err = "+strings.Replace(r.Name, "common.", "c.", -1)+"(w,req,user")
		for _, item := range r.Vars {
			out += "," + item
		}
		out += `)`
	}

	for _, route := range r.routeList {
		mapIt(route.Name)
		end := len(route.Path) - 1
		on(2, "case \""+route.Path[0:end]+"\":")
		//on(3,"id = " + strconv.Itoa(allRouteMap[route.Name]))
		runBefore(route.RunBefore, 3)
		if !route.Action && !route.NoHead {
			userCheckNano(3, route)
		}
		writeRoute(3, route)
		if route.Name != "common.RouteWebsockets" {
			on(3, "co.RouteViewCounter.Bump3("+strconv.Itoa(allRouteMap[route.Name])+", cn)")
		}
	}

	prec := NewPrec()
	prec.AddSet("MemberOnly", "SuperModOnly", "AdminOnly", "SuperAdminOnly")

	// Hoist runnables which appear on every route to the route group to avoid code duplication
	dupeMap := make(map[string]int)
	//skipRunnableAntiDupe:
	for _, g := range r.routeGroups {
		for _, route := range g.RouteList {
			if len(route.RunBefore) == 0 {
				continue //skipRunnableAntiDupe
			}
			// TODO: What if there are duplicates of the same runnable on this route?
			for _, runnable := range route.RunBefore {
				dupeMap[runnable.Contents] += 1
			}
		}

		// Unset entries which are already set on the route group
		for _, gRunnable := range g.RunBefore {
			delete(dupeMap, gRunnable.Contents)
			for _, item := range prec.LessThanItem(gRunnable.Contents) {
				delete(dupeMap, item)
			}
		}

		for runnable, count := range dupeMap {
			if count == len(g.RouteList) {
				g.Before(runnable)
			}
		}
		// This method is optimised in the compiler to do a bulk delete
		for name, _ := range dupeMap {
			delete(dupeMap, name)
		}
	}

	for _, group := range r.routeGroups {
		end := len(group.Path) - 1
		on(2, "case \""+group.Path[0:end]+"\":")
		runBefore(group.RunBefore, 3)
		on(3, "switch(req.URL.Path) {")

		defaultRoute := blankRoute()
		for _, route := range group.RouteList {
			if group.Path == route.Path {
				defaultRoute = route
				continue
			}
			mapIt(route.Name)

			on(4, "case \""+route.Path+"\":")
			//on(5,"id = " + strconv.Itoa(allRouteMap[route.Name]))
			if len(route.RunBefore) > 0 {
			skipRunnable:
				for _, runnable := range route.RunBefore {
					for _, gRunnable := range group.RunBefore {
						if gRunnable.Contents == runnable.Contents {
							continue skipRunnable
						}
						if prec.GreaterThan(gRunnable.Contents, runnable.Contents) {
							continue skipRunnable
						}
					}

					if runnable.Literal {
						on(5, runnable.Contents)
					} else {
						on(5, "err = c."+runnable.Contents+"(w,req,user)")
						iferrn(5)
						on(5, "")
					}
				}
			}
			if !route.Action && !route.NoHead && !group.NoHead {
				userCheckNano(5, route)
			}
			writeRoute(5, route)
			on(5, "co.RouteViewCounter.Bump3("+strconv.Itoa(allRouteMap[route.Name])+", cn)")
		}

		if defaultRoute.Name != "" {
			mapIt(defaultRoute.Name)
			on(4, "default:")
			//on(5,"id = " + strconv.Itoa(allRouteMap[defaultRoute.Name]))
			runBefore(defaultRoute.RunBefore, 4)
			if !defaultRoute.Action && !defaultRoute.NoHead && !group.NoHead {
				userCheckNano(5, defaultRoute)
			}
			writeRoute(5, defaultRoute)
			on(5, "co.RouteViewCounter.Bump3("+strconv.Itoa(allRouteMap[defaultRoute.Name])+", cn)")
		}
		on(3, "}")
	}

	// Stubs for us to refer to these routes through
	mapIt("routes.DynamicRoute")
	mapIt("routes.UploadedFile")
	mapIt("routes.StaticFile")
	mapIt("routes.RobotsTxt")
	mapIt("routes.SitemapXml")
	mapIt("routes.OpenSearchXml")
	mapIt("routes.Favicon")
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
		"opera",
		"chrome",
		"firefox",
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
		"slurp",
		"exabot",
		"mojeek",
		"cliqz",
		"qwant",
		"datenbank",
		"baidu",
		"sogou",
		"toutiao",
		"haosou",
		"duckduckgo",
		"seznambot",
		"discord",
		"telegram",
		"twitter",
		"facebook",
		"cloudflare",
		"archive_org",
		"uptimebot",
		"slackbot",
		"apple",
		"discourse",
		"xenforo",
		"mattermost",
		"alexa",
		"lynx",
		"blank",
		"malformed",
		"suspicious",
		"semrush",
		"dotbot",
		"ahrefs",
		"proximic",
		"megaindex",
		"majestic",
		"cocolyze",
		"babbar",
		"surdotly",
		"domcop",
		"netcraft",
		"seostar",
		"pandalytics",
		"blexbot",
		"wappalyzer",
		"twingly",
		"linkfluence",
		"pagething",
		"burf",
		"aspiegel",
		"mail_ru",
		"ccbot",
		"yacy",
		"zgrab",
		"cloudsystemnetworks",
		"maui",
		"curl",
		"python",
		//"go",
		"headlesschrome",
		"awesome_bot",
	}

	tmplVars.AllAgentMap = make(map[string]int)
	for id, agent := range tmplVars.AllAgentNames {
		tmplVars.AllAgentMap[agent] = id
	}

	tmplVars.AllAgentMarkNames = []string{}
	tmplVars.AllAgentMarks = map[string]string{}

	// Add agent marks
	a := func(mark, agent string) {
		tmplVars.AllAgentMarkNames = append(tmplVars.AllAgentMarkNames, mark)
		tmplVars.AllAgentMarks[mark] = agent
	}
	a("OPR", "opera")
	a("Chrome", "chrome")
	a("Firefox", "firefox")
	a("Safari", "safari")
	a("MSIE", "internetexplorer")
	a("Trident", "trident") // Hack to support IE11
	a("Edge", "edge")
	a("Lynx", "lynx") // There's a rare android variant of lynx which isn't covered by this
	a("SamsungBrowser", "samsung")
	a("UCBrowser", "ucbrowser")

	a("Google", "googlebot")
	a("Googlebot", "googlebot")
	a("yandex", "yandex") // from the URL
	a("DuckDuckBot", "duckduckgo")
	a("DuckDuckGo", "duckduckgo")
	a("Baiduspider", "baidu")
	a("Sogou", "sogou")
	a("ToutiaoSpider", "toutiao")
	a("Bytespider", "toutiao")
	a("360Spider", "haosou")
	a("bingbot", "bing")
	a("BingPreview", "bing")
	a("msnbot", "bing")
	a("Slurp", "slurp")
	a("Exabot", "exabot")
	a("MojeekBot", "mojeek")
	a("Cliqzbot", "cliqz")
	a("Qwantify", "qwant")
	a("netEstate", "datenbank")
	a("SeznamBot", "seznambot")
	a("CloudFlare", "cloudflare") // Track alwayson specifically in case there are other bots?
	a("archive", "archive_org")   //archive.org_bot
	a("Uptimebot", "uptimebot")
	a("Slackbot", "slackbot")
	a("Slack", "slackbot")
	a("Discordbot", "discord")
	a("TelegramBot", "telegram")
	a("Twitterbot", "twitter")
	a("facebookexternalhit", "facebook")
	a("Facebot", "facebook")
	a("Applebot", "apple")
	a("Discourse", "discourse")
	a("XenForo", "xenforo")
	a("mattermost", "mattermost")
	a("ia_archiver", "alexa")

	a("SemrushBot", "semrush")
	a("DotBot", "dotbot")
	a("AhrefsBot", "ahrefs")
	a("proximic", "proximic")
	a("MegaIndex", "megaindex")
	a("MJ12bot", "majestic") // TODO: This isn't matching bots out in the wild
	a("mj12bot", "majestic")
	a("Cocolyzebot", "cocolyze")
	a("Barkrowler", "babbar")
	a("SurdotlyBot", "surdotly")
	a("DomCopBot", "domcop")
	a("NetcraftSurveyAgent", "netcraft")
	a("seostar", "seostar")
	a("Pandalytics", "pandalytics")
	a("BLEXBot", "blexbot")
	a("Wappalyzer", "wappalyzer")
	a("Twingly", "twingly")
	a("linkfluence", "linkfluence")
	a("PageThing", "pagething")
	a("Burf", "burf")
	a("AspiegelBot", "aspiegel")
	a("PetalBot", "aspiegel")
	a("RU_Bot", "mail_ru") // Mail.RU_Bot
	a("CCBot", "ccbot")
	a("yacybot", "yacy")
	a("zgrab", "zgrab")
	a("Nimbostratus", "cloudsystemnetworks")
	a("MauiBot", "maui")
	a("curl", "curl")
	a("python", "python")
	//a("Go", "go") // yacy has java as part of it's UA, try to avoid hitting crawlers written in go
	a("HeadlessChrome", "headlesschrome")
	a("awesome_bot", "awesome_bot")
	// TODO: Detect Adsbot/3.1, it has a similar user agent to Google's Adsbot, but it is different. No Google fragments.

	tmplVars.AllAgentMarkIDs = make(map[string]int)
	for mark, agent := range tmplVars.AllAgentMarks {
		tmplVars.AllAgentMarkIDs[mark] = tmplVars.AllAgentMap[agent]
	}

	fileData := `// Code generated by Gosora's Router Generator. DO NOT EDIT.
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main

import (
	"strings"
	//"bytes"
	"strconv"
	"compress/gzip"
	"sync/atomic"
	"errors"
	"net/http"

	c "github.com/Azareal/Gosora/common"
	co "github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/uutils"
	"github.com/Azareal/Gosora/routes"
	"github.com/Azareal/Gosora/routes/panel"

	//"github.com/andybalholm/brotli"
)

var ErrNoRoute = errors.New("That route doesn't exist.")
// TODO: What about the /uploads/ route? x.x
var RouteMap = map[string]interface{}{ {{range .AllRouteNames}}
	"{{.Plain}}": {{.Short}},{{end}}
}

// ! NEVER RELY ON THESE REMAINING THE SAME BETWEEN COMMITS
var routeMapEnum = map[string]int{ {{range $index, $el := .AllRouteNames}}
	"{{$el.Plain}}": {{$index}},{{end}}
}
var reverseRouteMapEnum = map[int]string{ {{range $index, $el := .AllRouteNames}}
	{{$index}}: "{{$el.Plain}}",{{end}}
}
var osMapEnum = map[string]int{ {{range $index, $el := .AllOSNames}}
	"{{$el}}": {{$index}},{{end}}
}
var reverseOSMapEnum = map[int]string{ {{range $index, $el := .AllOSNames}}
	{{$index}}: "{{$el}}",{{end}}
}
var agentMapEnum = map[string]int{ {{range $index, $el := .AllAgentNames}}
	"{{$el}}": {{$index}},{{end}}
}
var reverseAgentMapEnum = map[int]string{ {{range $index, $el := .AllAgentNames}}
	{{$index}}: "{{$el}}",{{end}}
}
var markToAgent = map[string]string{ {{range $index, $el := .AllAgentMarkNames}}
	"{{$el}}": "{{index $.AllAgentMarks $el}}",{{end}}
}
var markToID = map[string]int{ {{range $index, $el := .AllAgentMarkNames}}
	"{{$el}}": {{index $.AllAgentMarkIDs $el}},{{end}}
}
/*var agentRank = map[string]int{
	"opera":9,
	"chrome":8,
	"safari":1,
}*/

// HTTPSRedirect is a connection handler which redirects all HTTP requests to HTTPS
type HTTPSRedirect struct {}

func (red *HTTPSRedirect) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	co.RouteViewCounter.Bump({{index .AllRouteMap "routes.HTTPSRedirect"}})
	dest := "https://" + req.Host + req.URL.String()
	http.Redirect(w, req, dest, http.StatusTemporaryRedirect)
}

func (r *GenRouter) SuspiciousRequest(req *http.Request, pre string) {
	if c.Config.DisableSuspLog {
		return
	}
	var sb strings.Builder
	if pre != "" {
		sb.WriteString("Suspicious Request\n")
	} else {
		pre = "Suspicious Request"
	}
	r.ddumpRequest(req,pre,r.suspLog,&sb)
	co.AgentViewCounter.Bump({{.AllAgentMap.suspicious}})
}

// TODO: Pass the default path or config struct to the router rather than accessing it via a package global
// TODO: SetDefaultPath
// TODO: GetDefaultPath
func (r *GenRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	malformedRequest := func(typ int) {
		w.WriteHeader(200) // 400
		w.Write([]byte(""))
		r.DumpRequest(req,"Malformed Request T"+strconv.Itoa(typ))
		co.AgentViewCounter.Bump({{.AllAgentMap.malformed}})
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
	} else if strings.Contains(req.Host,":") {
		spl := strings.Split(req.Host,":")
		if len(spl) > 2 {
			malformedRequest(1)
			return
		}
		shost = spl[0]
		//if len(spl)==2 {
			sport = spl[1]
		//}
	} else {
		shost = req.Host
	}
	// TODO: Reject requests from non-local IPs, if the site host is set to localhost or a localhost IP
	if !c.Config.LoosePort && c.Site.PortInt != 80 && c.Site.PortInt != 443 && sport != c.Site.Port {
		malformedRequest(2)
		return
	}
	
	// Redirect www. and local IP requests to the right place
	if strings.HasPrefix(shost, "www.") || c.Site.LocalHost {
	if shost == "www." + c.Site.Host || (c.Site.LocalHost && shost != c.Site.Host && isLocalHost(shost)) {
		// TODO: Abstract the redirect logic?
		w.Header().Set("Connection", "close")
		var s, p string
		if c.Config.SslSchema {
			s = "s"
		}
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
	}

	// Deflect malformed requests
	if len(req.URL.Path) == 0 || req.URL.Path[0] != '/' || (!c.Config.LooseHost && shost != c.Site.Host) {
		malformedRequest(3)
		return
	}
	r.suspScan(req)

	// Indirect the default route onto a different one
	if req.URL.Path == "/" {
		req.URL.Path = c.Config.DefaultPath
	}
	//log.Print("URL.Path: ", req.URL.Path)
	prefix := req.URL.Path[0:strings.IndexByte(req.URL.Path[1:],'/') + 1]

	// TODO: Use the same hook table as downstream
	hTbl := c.GetHookTable()
	skip, ferr := c.H_router_after_filters_hook(hTbl, w, req, prefix)
	if skip || ferr != nil {
		return
	}

	if prefix != "/ws" {
		h := w.Header()
		h.Set("X-Frame-Options", "deny")
		h.Set("X-XSS-Protection", "1; mode=block") // TODO: Remove when we add a CSP? CSP's are horrendously glitchy things, tread with caution before removing
		h.Set("X-Content-Type-Options", "nosniff")
		if c.Config.RefNoRef || !c.Config.SslSchema {
			h.Set("Referrer-Policy","no-referrer")
		} else {
			h.Set("Referrer-Policy","strict-origin")
		}
		h.Set("Permissions-Policy","interest-cohort=()")
	}
	
	if c.Dev.SuperDebug {
		r.DumpRequest(req,"before routes.StaticFile")
	}
	// Increment the request counter
	if !c.Config.DisableAnalytics {
		co.GlobalViewCounter.Bump()
	}
	
	if prefix == "/s" { //old prefix: /static
		if !c.Config.DisableAnalytics {
			co.RouteViewCounter.Bump({{index .AllRouteMap "routes.StaticFile"}})
		}
		routes.StaticFile(w, req)
		return
	}
	// TODO: Handle JS routes
	if atomic.LoadInt32(&c.IsDBDown) == 1 {
		c.DatabaseError(w, req)
		return
	}
	if c.Dev.SuperDebug {
		r.reqLogger.Print("before PreRoute")
	}

	/*if c.Dev.QuicPort != 0 {
		sQuicPort := strconv.Itoa(c.Dev.QuicPort)
		w.Header().Set("Alt-Svc", "quic=\":"+sQuicPort+"\"; ma=2592000; v=\"44,43,39\", h3-23=\":"+sQuicPort+"\"; ma=3600, h3-24=\":"+sQuicPort+"\"; ma=3600, h2=\":443\"; ma=3600")
	}*/

	// Track the user agents. Unfortunately, everyone pretends to be Mozilla, so this'll be a little less efficient than I would like.
	// TODO: Add a setting to disable this?
	// TODO: Use a more efficient detector instead of smashing every possible combination in
	// TODO: Make this testable
	var agent int
	if !c.Config.DisableAnalytics {
	
	ua := strings.TrimSpace(strings.Replace(strings.TrimPrefix(req.UserAgent(),"Mozilla/5.0 ")," Safari/537.36","",-1)) // Noise, no one's going to be running this and it would require some sort of agent ranking system to determine which identifier should be prioritised over another
	if ua == "" {
		co.AgentViewCounter.Bump({{.AllAgentMap.blank}})
		r.unknownUA(req)
	} else {		
		// WIP UA Parser
		//var ii = uaBufPool.Get()
		var buf []byte
		//if ii != nil {
		//	buf = ii.([]byte)
		//}
		var items []string
		var os int
		for _, it := range uutils.StringToBytes(ua) {
			if (it > 64 && it < 91) || (it > 96 && it < 123) || (it > 47 && it < 58) || it == '_' {
				// TODO: Store an index and slice that instead?
				buf = append(buf, it)
			} else if it == ' ' || it == '(' || it == ')' || it == '-' || it == ';' || it == ':' || it == '.' || it == '+' || it == '~' || it == '@' /*|| (it == ':' && bytes.Equal(buf,[]byte("http")))*/ || it == ',' || it == '/' {
				//log.Print("buf: ",string(buf))
				//log.Print("it: ",string(it))
				if len(buf) != 0 {
					if len(buf) > 2 {
						// Use an unsafe zero copy conversion here just to use the switch, it's not safe for this string to escape from here, as it will get mutated, so do a regular string conversion in append
						switch(uutils.BytesToString(buf)) {
						case "Windows":
							os = {{.AllOSMap.windows}}
						case "Linux":
							os = {{.AllOSMap.linux}}
						case "Mac":
							os = {{.AllOSMap.mac}}
						case "iPhone":
							os = {{.AllOSMap.iphone}}
						case "Android":
							os = {{.AllOSMap.android}}
						case "like","compatible","NT","X","com","KHTML":
							// Skip these words
						default:
							//log.Print("append buf")
							items = append(items, string(buf))
						}
					}
					//log.Print("reset buf")
					buf = buf[:0]
				}
			} else {
				// TODO: Test this
				items = items[:0]
				if c.Config.DisableSuspLog {
					r.reqLogger.Print("Illegal char "+strconv.Itoa(int(it))+" in UA\nUA Buf: ", buf,"\nUA Buf String: ", string(buf))
				} else {
					r.SuspiciousRequest(req,"Illegal char "+strconv.Itoa(int(it))+" in UA")
					r.reqLogger.Print("UA Buf: ", buf,"\nUA Buf String: ", string(buf))
				}
				break
			}
		}
		//uaBufPool.Put(buf)

		// Iterate over this in reverse as the real UA tends to be on the right side
		for i := len(items) - 1; i >= 0; i-- {
			//fAgent, ok := markToAgent[items[i]]
			fAgent, ok := markToID[items[i]]
			if ok {
				agent = fAgent
				if agent != {{.AllAgentMap.safari}} {
					break
				}
			}
		}
		if c.Dev.SuperDebug {
			r.reqLogger.Print("parsed agent: ", agent,"\nos: ", os)
			r.reqLogger.Printf("items: %+v\n",items)
			/*for _, it := range items {
				r.reqLogger.Printf("it: %+v\n",string(it))
			}*/
		}
		
		// Special handling
		switch(agent) {
		case {{.AllAgentMap.chrome}}:
			if os == {{.AllOSMap.android}} {
				agent = {{.AllAgentMap.androidchrome}}
			}
		case {{.AllAgentMap.safari}}:
			if os == {{.AllOSMap.iphone}} {
				agent = {{.AllAgentMap.mobilesafari}}
			}
		case {{.AllAgentMap.trident}}:
			// Hack to support IE11, change this after we start logging versions
			if strings.Contains(ua,"rv:11") {
				agent = {{.AllAgentMap.internetexplorer}}
			}
		case {{.AllAgentMap.zgrab}}:
			w.WriteHeader(200) // 400
			w.Write([]byte(""))
			r.DumpRequest(req,"Blocked Scanner")
			co.AgentViewCounter.Bump({{.AllAgentMap.zgrab}})
			return
		}
		
		if agent == 0 {
			//co.AgentViewCounter.Bump({{.AllAgentMap.unknown}})
			r.unknownUA(req)
		}// else {
			//co.AgentViewCounter.Bump(agentMapEnum[agent])
			co.AgentViewCounter.Bump(agent)
		//}
		co.OSViewCounter.Bump(os)
	}

	// TODO: Do we want to track missing language headers too? Maybe as it's own type, e.g. "noheader"?
	// TODO: Default to anything other than en, if anything else is present, to avoid over-representing it for multi-linguals?
	lang := req.Header.Get("Accept-Language")
	if lang != "" {
		// TODO: Reduce allocs here
		lLang := strings.Split(strings.TrimSpace(lang),"-")
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
		if !co.LangViewCounter.Bump(llLang) {
			r.DumpRequest(req,"Invalid ISO Code")
		}
	} else {
		co.LangViewCounter.Bump2(0)
	}

	if !c.Config.RefNoTrack {
		ae := req.Header.Get("Accept-Encoding")
		likelyBot := ae == "gzip" || ae == ""
		if !likelyBot {
			ref := req.Header.Get("Referer") // Check the 'referrer' header too? :P
			// TODO: Extend the effects of DNT elsewhere?
			if ref != "" && req.Header.Get("DNT") != "1" {
				// ? Optimise this a little?
				ref = strings.TrimPrefix(strings.TrimPrefix(ref,"http://"),"https://")
				ref = strings.Split(ref,"/")[0]
				portless := strings.Split(ref,":")[0]
				// TODO: Handle c.Site.Host in uppercase too?
				if portless != "localhost" && portless != "127.0.0.1" && portless != c.Site.Host {
					r.DumpRequest(req,"Ref Route")
					co.ReferrerTracker.Bump(ref)
				}
			}
		}
	}

	}
	
	// Deal with the session stuff, etc.
	ucpy, ok := c.PreRoute(w, req)
	if !ok {
		return
	}
	user := &ucpy
	user.LastAgent = agent
	if c.Dev.SuperDebug {
		r.reqLogger.Print(
			"after PreRoute\n" +
			"routeMapEnum: ", routeMapEnum)
	}
	//log.Println("req: ", req)

	// Disable Gzip when SSL is disabled for security reasons?
	if prefix != "/ws" {
		ae := req.Header.Get("Accept-Encoding")
		/*if strings.Contains(ae, "br") {
			h := w.Header()
			h.Set("Content-Encoding", "br")
			var ii = brPool.Get()
			var igzw *brotli.Writer
			if ii == nil {
				igzw = brotli.NewWriter(w)
			} else {
				igzw = ii.(*brotli.Writer)
				igzw.Reset(w)
			}
			gzw := c.BrResponseWriter{Writer: igzw, ResponseWriter: w}
			defer func() {
				//h := w.Header()
				if h.Get("Content-Encoding") == "br" && h.Get("X-I") == "" {
					//log.Print("push br close")
					igzw := gzw.Writer.(*brotli.Writer)
					igzw.Close()
					brPool.Put(igzw)
				}
			}()
			w = gzw
		} else */if strings.Contains(ae, "gzip") {
			h := w.Header()
			h.Set("Content-Encoding", "gzip")
			var ii = gzipPool.Get()
			var igzw *gzip.Writer
			if ii == nil {
				igzw = gzip.NewWriter(w)
			} else {
				igzw = ii.(*gzip.Writer)
				igzw.Reset(w)
			}
			gzw := c.GzipResponseWriter{Writer: igzw, ResponseWriter: w}
			defer func() {
				//h := w.Header()
				if h.Get("Content-Encoding") == "gzip" && h.Get("X-I") == "" {
					//log.Print("push gzip close")
					igzw := gzw.Writer.(*gzip.Writer)
					igzw.Close()
					gzipPool.Put(igzw)
				}
			}()
			w = gzw
		}
	}

	skip, ferr = c.H_router_pre_route_hook(hTbl, w, req, user, prefix)
	if skip || ferr != nil {
		r.handleError(ferr,w,req,user)
		return
	}
	var extraData string
	if req.URL.Path[len(req.URL.Path) - 1] != '/' {
		extraData = req.URL.Path[strings.LastIndexByte(req.URL.Path,'/') + 1:]
		req.URL.Path = req.URL.Path[:strings.LastIndexByte(req.URL.Path,'/') + 1]
	}
	ferr = r.routeSwitch(w, req, user, prefix, extraData)
	if ferr != nil {
		r.handleError(ferr,w,req,user)
		return
	}
	/*if !c.Config.DisableAnalytics {
		co.RouteViewCounter.Bump(id)
	}*/

	hTbl.VhookNoRet("router_end", w, req, user, prefix, extraData)
	//c.StoppedServer("Profile end")
}
	
func (r *GenRouter) routeSwitch(w http.ResponseWriter, req *http.Request, user *c.User, prefix, extraData string) /*(id int, orerr */c.RouteError/*)*/ {
	var err c.RouteError
	cn := uutils.Nanotime()
	switch(prefix) {` + out + `
		/*case "/sitemaps": // TODO: Count these views
			req.URL.Path += extraData
			err = sitemapSwitch(w,req)*/
		// ! Temporary fix for certain bots
		case "/static":
			w.Header().Set("Connection", "close")
			http.Redirect(w, req, "/s/"+extraData, http.StatusTemporaryRedirect)
		case "/uploads":
			if extraData == "" {
				co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.UploadedFile"}}, cn)
				return c.NotFound(w,req,nil)
			}
			w = r.responseWriter(w)
			req.URL.Path += extraData
			// TODO: Find a way to propagate errors up from this?
			r.UploadHandler(w,req) // TODO: Count these views
			co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.UploadedFile"}}, cn)
			return nil
		case "":
			// Stop the favicons, robots.txt file, etc. resolving to the topics list
			// TODO: Add support for favicons and robots.txt files
			switch(extraData) {
				case "robots.txt":
					co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.RobotsTxt"}}, cn)
					return routes.RobotsTxt(w,req)
				case "favicon.ico":
					w = r.responseWriter(w)
					req.URL.Path = "/s/favicon.ico"
					co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.Favicon"}}, cn)
					routes.StaticFile(w,req)
					return nil
				case "opensearch.xml":
					co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.OpenSearchXml"}}, cn)
					return routes.OpenSearchXml(w,req)
				/*case "sitemap.xml":
					co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.SitemapXml"}}, cn)
					return routes.SitemapXml(w,req)*/
			}
			co.RouteViewCounter.Bump({{index .AllRouteMap "routes.Error"}})
			return c.NotFound(w,req,nil)
		default:
			// A fallback for dynamic routes, e.g. ones declared by plugins
			r.RLock()
			h, ok := r.extraRoutes[req.URL.Path]
			r.RUnlock()
			req.URL.Path += extraData
			
			if ok {
				// TODO: Be more specific about *which* dynamic route it is
				co.RouteViewCounter.Bump({{index .AllRouteMap "routes.DynamicRoute"}})
				return h(w,req,user)
			}
			co.RouteViewCounter.Bump3({{index .AllRouteMap "routes.BadRoute"}}, cn)

			if !c.Config.DisableSuspLog {
			lp := strings.ToLower(req.URL.Path)
			if strings.Contains(lp,"w") {
				if strings.Contains(lp,"wp") || strings.Contains(lp,"wordpress") || strings.Contains(lp,"wget") || strings.Contains(lp,"wp-") {
					r.SuspiciousRequest(req,"Bad Route")
					return c.MicroNotFound(w,req)
				}
			}
			if strings.Contains(lp,"admin") || strings.Contains(lp,"sql") || strings.Contains(lp,"manage") || strings.Contains(lp,"//") || strings.Contains(lp,"\\\\") || strings.Contains(lp,"config") || strings.Contains(lp,"setup") || strings.Contains(lp,"install") || strings.Contains(lp,"update") || strings.Contains(lp,"php") || strings.Contains(lp,"pl") || strings.Contains(lp,"include") || strings.Contains(lp,"vendor") || strings.Contains(lp,"bin") || strings.Contains(lp,"system") || strings.Contains(lp,"eval") || strings.Contains(lp,"config") {
				r.SuspiciousRequest(req,"Bad Route")
				return c.MicroNotFound(w,req)
			}
			}

			if !c.Config.DisableBadRouteLog {
				r.DumpRequest(req,"Bad Route")
			}
			ae := req.Header.Get("Accept-Encoding")
			likelyBot := ae == "gzip" || ae == ""
			if likelyBot {
				return c.MicroNotFound(w,req)
			}
			return c.NotFound(w,req,nil)
	}
	return err
}
`
	tmpl := template.Must(template.New("router").Parse(fileData))
	var b bytes.Buffer
	if err := tmpl.Execute(&b, tmplVars); err != nil {
		log.Fatal(err)
	}

	writeFile("./gen_router.go", b.String())
	log.Println("Successfully generated the router")
}

func writeFile(name, content string) {
	f, e := os.Create(name)
	if e != nil {
		log.Fatal(e)
	}
	_, e = f.WriteString(content)
	if e != nil {
		log.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		log.Fatal(e)
	}
	if e = f.Close(); e != nil {
		log.Fatal(e)
	}
}
