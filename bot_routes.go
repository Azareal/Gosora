package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"./common"
)

// TODO: Make this a static file somehow? Is it possible for us to put this file somewhere else?
// TODO: Add an API so that plugins can register disallowed areas. E.g. /guilds/join for plugin_guilds
func routeRobotsTxt(w http.ResponseWriter, r *http.Request) common.RouteError {
	// TODO: Do we have to put * or something at the end of the paths?
	_, _ = w.Write([]byte(`User-agent: *
Disallow: /panel/
Disallow: /topics/create/
Disallow: /user/edit/
Disallow: /accounts/
Disallow: /report/
`))
	return nil
}

var sitemapPageCap = 40000 // 40k, bump it up to 50k once we gzip this? Does brotli work on sitemaps?

func writeXMLHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
}

// TODO: Keep track of when a sitemap was last modifed and add a lastmod element for it
func routeSitemapXml(w http.ResponseWriter, r *http.Request) common.RouteError {
	var sslBit string
	if common.Site.EnableSsl {
		sslBit = "s"
	}
	var sitemapItem = func(path string) {
		w.Write([]byte(`<sitemap>
	<loc>http` + sslBit + `://` + common.Site.URL + "/" + path + `</loc>
</sitemap>
`))
	}
	writeXMLHeader(w, r)
	w.Write([]byte("<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))
	sitemapItem("sitemaps/topics.xml")
	//sitemapItem("sitemaps/forums.xml")
	//sitemapItem("sitemaps/users.xml")
	w.Write([]byte("</sitemapindex>"))

	return nil
}

type FuzzyRoute struct {
	Path   string
	Handle func(http.ResponseWriter, *http.Request, int) common.RouteError
}

// TODO: Add a sitemap API and clean things up
// TODO: ^-- Make sure that the API is concurrent
// TODO: Add a social group sitemap
var sitemapRoutes = map[string]func(http.ResponseWriter, *http.Request) common.RouteError{
	"forums.xml": routeSitemapForums,
	"topics.xml": routeSitemapTopics,
}

// TODO: Use a router capable of parsing this rather than hard-coding the logic in
var fuzzySitemapRoutes = map[string]FuzzyRoute{
	"topics_page_": FuzzyRoute{"topics_page_(%d).xml", routeSitemapTopic},
}

func sitemapSwitch(w http.ResponseWriter, r *http.Request) common.RouteError {
	var path = r.URL.Path[len("/sitemaps/"):]
	for name, fuzzy := range fuzzySitemapRoutes {
		if strings.HasPrefix(path, name) && strings.HasSuffix(path, ".xml") {
			var spath = strings.TrimPrefix(path, name)
			spath = strings.TrimSuffix(spath, ".xml")
			page, err := strconv.Atoi(spath)
			if err != nil {
				if common.Dev.DebugMode {
					log.Printf("Unable to convert string '%s' to integer in fuzzy route", spath)
				}
				return common.NotFound(w, r)
			}
			return fuzzy.Handle(w, r, page)
		}
	}

	route, ok := sitemapRoutes[path]
	if !ok {
		return common.NotFound(w, r)
	}
	return route(w, r)
}

func routeSitemapForums(w http.ResponseWriter, r *http.Request) common.RouteError {
	var sslBit string
	if common.Site.EnableSsl {
		sslBit = "s"
	}
	var sitemapItem = func(path string) {
		w.Write([]byte(`<url>
	<loc>http` + sslBit + `://` + common.Site.URL + path + `</loc>
</url>
`))
	}

	group, err := common.Groups.Get(common.GuestUser.Group)
	if err != nil {
		return common.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	writeXMLHeader(w, r)
	w.Write([]byte("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))

	for _, fid := range group.CanSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		var forum = common.Forums.DirtyGet(fid).Copy()
		if forum.ParentID == 0 && forum.Name != "" && forum.Active {
			sitemapItem(common.BuildForumURL(common.NameToSlug(forum.Name), forum.ID))
		}
	}

	w.Write([]byte("</urlset>"))
	return nil
}

// TODO: Add a global ratelimit. 10 50MB files (smaller if compressed better) per minute?
// ? We might have problems with banned users, if they have fewer ViewTopic permissions than guests as they'll be able to see this list. Then again, a banned user could just logout to see it
func routeSitemapTopics(w http.ResponseWriter, r *http.Request) common.RouteError {
	var sslBit string
	if common.Site.EnableSsl {
		sslBit = "s"
	}
	var sitemapItem = func(path string) {
		w.Write([]byte(`<sitemap>
	<loc>http` + sslBit + `://` + common.Site.URL + "/" + path + `</loc>
</sitemap>
`))
	}

	group, err := common.Groups.Get(common.GuestUser.Group)
	if err != nil {
		return common.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	var visibleForums []common.Forum
	for _, fid := range group.CanSee {
		forum := common.Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active {
			visibleForums = append(visibleForums, forum.Copy())
		}
	}

	topicCount, err := common.TopicCountInForums(visibleForums)
	if err != nil {
		return common.InternalErrorXML(err, w, r)
	}

	var pageCount = topicCount / sitemapPageCap
	//log.Print("topicCount", topicCount)
	//log.Print("pageCount", pageCount)
	writeXMLHeader(w, r)
	w.Write([]byte("<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))
	for i := 0; i <= pageCount; i++ {
		sitemapItem("sitemaps/topics_page_" + strconv.Itoa(i) + ".xml")
	}
	w.Write([]byte("</sitemapindex>"))
	return nil
}

func routeSitemapTopic(w http.ResponseWriter, r *http.Request, page int) common.RouteError {
	/*var sslBit string
	if common.Site.EnableSsl {
		sslBit = "s"
	}
	var sitemapItem = func(path string) {
			w.Write([]byte(`<url>
		<loc>http` + sslBit + `://` + common.Site.URL + "/" + path + `</loc>
	</url>
	`))
		}*/

	group, err := common.Groups.Get(common.GuestUser.Group)
	if err != nil {
		return common.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	var visibleForums []common.Forum
	for _, fid := range group.CanSee {
		forum := common.Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active {
			visibleForums = append(visibleForums, forum.Copy())
		}
	}

	argList, qlist := common.ForumListToArgQ(visibleForums)
	topicCount, err := common.ArgQToTopicCount(argList, qlist)
	if err != nil {
		return common.InternalErrorXML(err, w, r)
	}

	var pageCount = topicCount / sitemapPageCap
	//log.Print("topicCount", topicCount)
	//log.Print("pageCount", pageCount)
	//log.Print("page",page)
	if page > pageCount {
		page = pageCount
	}

	writeXMLHeader(w, r)
	w.Write([]byte("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))

	w.Write([]byte("</urlset>"))
	return nil
}

func routeSitemapUsers(w http.ResponseWriter, r *http.Request) common.RouteError {
	writeXMLHeader(w, r)
	w.Write([]byte("<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))
	return nil
}
