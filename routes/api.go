package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
)

// TODO: Make this a static file somehow? Is it possible for us to put this file somewhere else?
// TODO: Add an API so that plugins can register disallowed areas. E.g. /guilds/join for plugin_guilds
func RobotsTxt(w http.ResponseWriter, r *http.Request) c.RouteError {
	// TODO: Do we have to put * or something at the end of the paths?
	_, _ = w.Write([]byte(`User-agent: *
Disallow: /panel/*
Disallow: /topics/create/
Disallow: /user/edit/*
Disallow: /accounts/*
Disallow: /report/*
`))
	return nil
}

var sitemapPageCap = 40000 // 40k, bump it up to 50k once we gzip this? Does brotli work on sitemaps?

func writeXMLHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
}

// TODO: Keep track of when a sitemap was last modifed and add a lastmod element for it
func SitemapXml(w http.ResponseWriter, r *http.Request) c.RouteError {
	var s string
	if c.Config.SslSchema {
		s = "s"
	}
	sitemapItem := func(path string) {
		w.Write([]byte(`<sitemap>
	<loc>http` + s + `://` + c.Site.URL + "/" + path + `</loc>
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
	Handle func(http.ResponseWriter, *http.Request, int) c.RouteError
}

// TODO: Add a sitemap API and clean things up
// TODO: ^-- Make sure that the API is concurrent
// TODO: Add a social group sitemap
var sitemapRoutes = map[string]func(http.ResponseWriter, *http.Request) c.RouteError{
	"forums.xml": SitemapForums,
	"topics.xml": SitemapTopics,
}

// TODO: Use a router capable of parsing this rather than hard-coding the logic in
var fuzzySitemapRoutes = map[string]FuzzyRoute{
	"topics_page_": FuzzyRoute{"topics_page_(%d).xml", SitemapTopic},
}

func sitemapSwitch(w http.ResponseWriter, r *http.Request) c.RouteError {
	path := r.URL.Path[len("/sitemaps/"):]
	for name, fuzzy := range fuzzySitemapRoutes {
		if strings.HasPrefix(path, name) && strings.HasSuffix(path, ".xml") {
			spath := strings.TrimPrefix(path, name)
			spath = strings.TrimSuffix(spath, ".xml")
			page, err := strconv.Atoi(spath)
			if err != nil {
				// ? What's this? Do we need it? Was it just a quick trace?
				c.DebugLogf("Unable to convert string '%s' to integer in fuzzy route", spath)
				return c.NotFound(w, r, nil)
			}
			return fuzzy.Handle(w, r, page)
		}
	}

	route, ok := sitemapRoutes[path]
	if !ok {
		return c.NotFound(w, r, nil)
	}
	return route(w, r)
}

func SitemapForums(w http.ResponseWriter, r *http.Request) c.RouteError {
	var s string
	if c.Config.SslSchema {
		s = "s"
	}
	sitemapItem := func(path string) {
		w.Write([]byte(`<url>
	<loc>http` + s + `://` + c.Site.URL + path + `</loc>
</url>
`))
	}

	group, err := c.Groups.Get(c.GuestUser.Group)
	if err != nil {
		return c.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	writeXMLHeader(w, r)
	w.Write([]byte("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))

	for _, fid := range group.CanSee {
		// Avoid data races by copying the struct into something we can freely mold without worrying about breaking something somewhere else
		f := c.Forums.DirtyGet(fid).Copy()
		if f.ParentID == 0 && f.Name != "" && f.Active {
			sitemapItem(c.BuildForumURL(c.NameToSlug(f.Name), f.ID))
		}
	}

	w.Write([]byte("</urlset>"))
	return nil
}

// TODO: Add a global ratelimit. 10 50MB files (smaller if compressed better) per minute?
// ? We might have problems with banned users, if they have fewer ViewTopic permissions than guests as they'll be able to see this list. Then again, a banned user could just logout to see it
func SitemapTopics(w http.ResponseWriter, r *http.Request) c.RouteError {
	var s string
	if c.Config.SslSchema {
		s = "s"
	}
	sitemapItem := func(path string) {
		w.Write([]byte(`<sitemap>
	<loc>http` + s + `://` + c.Site.URL + "/" + path + `</loc>
</sitemap>
`))
	}

	group, err := c.Groups.Get(c.GuestUser.Group)
	if err != nil {
		return c.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	var visibleForums []c.Forum
	for _, fid := range group.CanSee {
		forum := c.Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active {
			visibleForums = append(visibleForums, forum.Copy())
		}
	}

	topicCount, err := c.TopicCountInForums(visibleForums)
	if err != nil {
		return c.InternalErrorXML(err, w, r)
	}

	pageCount := topicCount / sitemapPageCap
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

func SitemapTopic(w http.ResponseWriter, r *http.Request, page int) c.RouteError {
	/*var s string
	if c.Config.SslSchema {
		s = "s"
	}
	var sitemapItem = func(path string) {
			w.Write([]byte(`<url>
		<loc>http` + s + `://` + c.Site.URL + "/" + path + `</loc>
	</url>
	`))
		}*/

	group, err := c.Groups.Get(c.GuestUser.Group)
	if err != nil {
		return c.SilentInternalErrorXML(errors.New("The guest group doesn't exist for some reason"), w, r)
	}

	var visibleForums []c.Forum
	for _, fid := range group.CanSee {
		forum := c.Forums.DirtyGet(fid)
		if forum.Name != "" && forum.Active {
			visibleForums = append(visibleForums, forum.Copy())
		}
	}

	argList, qlist := c.ForumListToArgQ(visibleForums)
	topicCount, err := c.ArgQToTopicCount(argList, qlist)
	if err != nil {
		return c.InternalErrorXML(err, w, r)
	}

	pageCount := topicCount / sitemapPageCap
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

func SitemapUsers(w http.ResponseWriter, r *http.Request) c.RouteError {
	writeXMLHeader(w, r)
	w.Write([]byte("<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"))
	return nil
}

type JsonMe struct {
	User *c.MeUser
	Site MeSite
}

// We don't want to expose too much information about the site, so we'll make this a small subset of c.site
type MeSite struct {
	MaxRequestSize int
}

// APIMe returns information about the current logged-in user
// TODO: Find some way to stop intermediaries from doing compression to avoid the BREACH attack
// TODO: Decouple site settings into a different API? I'd like to avoid having too many requests, if possible, maybe we can use a different name for this?
func APIMe(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	h := w.Header()
	h.Set("Content-Type", "application/json")
	// We don't want an intermediary accidentally caching this
	// TODO: Use this header anywhere with a user check?
	h.Set("Cache-Control", "private")

	me := JsonMe{u.Me(), MeSite{c.Site.MaxRequestSize}}
	jsonBytes, err := json.Marshal(me)
	if err != nil {
		return c.InternalErrorJS(err, w, r)
	}
	w.Write(jsonBytes)

	return nil
}

func OpenSearchXml(w http.ResponseWriter, r *http.Request) c.RouteError {
	w.Header().Set("Content-Type", "application/xml")
	furl := "http"
	if c.Config.SslSchema {
		furl += "s"
	}
	furl += "://" + c.Site.URL
	w.Write([]byte(`<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/" xmlns:moz="http://www.mozilla.org/2006/browser/search/">
	<ShortName>` + c.Site.Name + `</ShortName>
	<InputEncoding>UTF-8</InputEncoding>
	<Url type="text/html" template="` + furl + `/topics/?q={searchTerms}"/>
	<Url type="application/opensearchdescription+xml" rel="self" template="` + furl + `/opensearch.xml"/>
	<moz:SearchForm>` + furl + `</moz:SearchForm>
</OpenSearchDescription>`))
	return nil
}
