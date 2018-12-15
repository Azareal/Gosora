package common

import (
	"html/template"
	"io"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azareal/Gosora/common/alerts"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/common/templates"
)

var Ctemplates []string // TODO: Use this to filter out top level templates we don't need
var Templates = template.New("")
var PrebuildTmplList []func(User, *Header) CTmpl

func skipCTmpl(key string) bool {
	for _, tmpl := range Ctemplates {
		if strings.HasSuffix(key, "/"+tmpl+".html") {
			return true
		}
	}
	return false
}

type CTmpl struct {
	Name       string
	Filename   string
	Path       string
	StructName string
	Data       interface{}
	Imports    []string
}

// TODO: Refactor the template trees to not need these
// TODO: Stop duplicating these bits of code
// nolint
func interpretedTopicTemplate(pi TopicPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["topic"]
	if !ok {
		mapping = "topic"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_topic_handle = interpretedTopicTemplate
var Template_topic_alt_handle = interpretedTopicTemplate
var Template_topic_alt_guest_handle = interpretedTopicTemplate
var Template_topic_alt_member_handle = interpretedTopicTemplate

// nolint
var Template_topics_handle = func(pi TopicListPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["topics"]
	if !ok {
		mapping = "topics"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}
var Template_topics_guest_handle = Template_topics_handle
var Template_topics_member_handle = Template_topics_handle

// nolint
var Template_forum_handle = func(pi ForumPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["forum"]
	if !ok {
		mapping = "forum"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}
var Template_forum_guest_handle = Template_forum_handle
var Template_forum_member_handle = Template_forum_handle

// nolint
var Template_forums_handle = func(pi ForumsPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["forums"]
	if !ok {
		mapping = "forums"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_profile_handle = func(pi ProfilePage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["profile"]
	if !ok {
		mapping = "profile"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}
var Template_profile_guest_handle = Template_profile_handle
var Template_profile_member_handle = Template_profile_handle

// nolint
var Template_create_topic_handle = func(pi CreateTopicPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["create_topic"]
	if !ok {
		mapping = "create_topic"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_login_handle = func(pi Page, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["login"]
	if !ok {
		mapping = "login"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_register_handle = func(pi Page, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["register"]
	if !ok {
		mapping = "register"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_error_handle = func(pi ErrorPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["error"]
	if !ok {
		mapping = "error"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_ip_search_handle = func(pi IPSearchPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["ip_search"]
	if !ok {
		mapping = "ip_search"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var Template_account_handle = func(pi AccountDashPage, w io.Writer) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["account"]
	if !ok {
		mapping = "account"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

func tmplInitUsers() (User, User, User) {
	avatar, microAvatar := BuildAvatar(62, "")
	user := User{62, BuildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", avatar, microAvatar, "", "", "", "", 0, 0, 0, "0.0.0.0.0", 0}

	// TODO: Do a more accurate level calculation for this?
	avatar, microAvatar = BuildAvatar(1, "")
	user2 := User{1, BuildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", "", "", 58, 1000, 0, "127.0.0.1", 0}

	avatar, microAvatar = BuildAvatar(2, "")
	user3 := User{2, BuildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", "", "", 42, 900, 0, "::1", 0}
	return user, user2, user3
}

func tmplInitHeaders(user User, user2 User, user3 User) (*Header, *Header, *Header) {
	header := &Header{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		Theme:       Themes[DefaultThemeBox.Load().(string)],
		CurrentUser: user,
		NoticeList:  []string{"test"},
		Stylesheets: []string{"panel"},
		Scripts:     []string{"whatever"},
		Widgets: PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	buildHeader := func(user User) *Header {
		var head = &Header{Site: Site}
		*head = *header
		head.CurrentUser = user
		return head
	}

	return header, buildHeader(user2), buildHeader(user3)
}

type TmplLoggedin struct {
	Stub   string
	Guest  string
	Member string
}

// ? - Add template hooks?
func CompileTemplates() error {
	var config tmpl.CTemplateConfig
	config.Minify = Config.MinifyTemplates
	config.Debug = Dev.DebugMode
	config.SuperDebug = Dev.TemplateDebug

	c := tmpl.NewCTemplateSet()
	c.SetConfig(config)
	c.SetBaseImportMap(map[string]string{
		"io":                               "io",
		"github.com/Azareal/Gosora/common": "github.com/Azareal/Gosora/common",
	})
	c.SetBuildTags("!no_templategen")

	// Schemas to train the template compiler on what to expect
	// TODO: Add support for interface{}s
	user, user2, user3 := tmplInitUsers()
	header, header2, _ := tmplInitHeaders(user, user2, user3)
	now := time.Now()

	log.Print("Compiling the templates")

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		PollOption{0, "Nothing"},
		PollOption{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, RelativeTime(now), now, RelativeTime(now), 0, "", "127.0.0.1", 1, 0, 1, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false}
	var replyList []ReplyUser
	// TODO: Do we want the UID on this to be 0?
	avatar, microAvatar = BuildAvatar(0, "")
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, RelativeTime(now), 0, 0, avatar, microAvatar, "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, "", ""})

	var varList = make(map[string]tmpl.VarItem)
	var compile = func(name string, expects string, expectsInt interface{}) (tmpl string, err error) {
		return c.Compile(name+".html", "templates/", expects, expectsInt, varList)
	}
	var compileByLoggedin = func(name string, expects string, expectsInt interface{}) (tmpl TmplLoggedin, err error) {
		stub, guest, member, err := c.CompileByLoggedin(name+".html", "templates/", expects, expectsInt, varList)
		return TmplLoggedin{stub, guest, member}, err
	}

	header.Title = "Topic Name"
	tpage := TopicPage{header, replyList, topic, &Forum{ID: 1, Name: "Hahaha"}, poll, 1, 1}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	topicTmpl, err := compile("topic", "common.TopicPage", tpage)
	if err != nil {
		return err
	}
	/*topicAltTmpl, err := compile("topic_alt", "common.TopicPage", tpage)
	if err != nil {
		return err
	}*/

	topicAltTmpl, err := compileByLoggedin("topic_alt", "common.TopicPage", tpage)
	if err != nil {
		return err
	}

	varList = make(map[string]tmpl.VarItem)
	header.Title = "User 526"
	ppage := ProfilePage{header, replyList, user, 0, 0} // TODO: Use the score from user to generate the currentScore and nextScore
	profileTmpl, err := compileByLoggedin("profile", "common.ProfilePage", ppage)
	if err != nil {
		return err
	}

	// TODO: Use a dummy forum list to avoid o(n) problems
	var forumList []Forum
	forums, err := Forums.GetAll()
	if err != nil {
		return err
	}
	for _, forum := range forums {
		forumList = append(forumList, *forum)
	}

	varList = make(map[string]tmpl.VarItem)
	header.Title = "Forum List"
	forumsPage := ForumsPage{header, forumList}
	forumsTmpl, err := compile("forums", "common.ForumsPage", forumsPage)
	if err != nil {
		return err
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList, &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, "Date", user3.ID, 1, "", "127.0.0.1", 1, 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"})
	header2.Title = "Topic List"
	topicListPage := TopicListPage{header, topicsList, forumList, Config.DefaultForum, TopicListSort{"lastupdated", false}, Paginator{[]int{1}, 1, 1}}
	/*topicListTmpl, err := compile("topics", "common.TopicListPage", topicListPage)
	if err != nil {
		return err
	}*/
	topicListTmpl, err := compileByLoggedin("topics", "common.TopicListPage", topicListPage)
	if err != nil {
		return err
	}

	forumItem := BlankForum(1, "general-forum.1", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0)
	header.Title = "General Forum"
	forumPage := ForumPage{header, topicsList, forumItem, Paginator{[]int{1}, 1, 1}}
	forumTmpl, err := compileByLoggedin("forum", "common.ForumPage", forumPage)
	if err != nil {
		return err
	}

	header.Title = "Login Page"
	loginPage := Page{header, tList, nil}
	loginTmpl, err := compile("login", "common.Page", loginPage)
	if err != nil {
		return err
	}

	header.Title = "Registration Page"
	registerPage := Page{header, tList, "nananana"}
	registerTmpl, err := compile("register", "common.Page", registerPage)
	if err != nil {
		return err
	}

	header.Title = "Error"
	errorPage := ErrorPage{header, "A problem has occurred in the system."}
	errorTmpl, err := compile("error", "common.ErrorPage", errorPage)
	if err != nil {
		return err
	}

	var ipUserList = make(map[int]*User)
	ipUserList[1] = &user2
	header.Title = "IP Search"
	ipSearchPage := IPSearchPage{header2, ipUserList, "::1"}
	ipSearchTmpl, err := compile("ip_search", "common.IPSearchPage", ipSearchPage)
	if err != nil {
		return err
	}

	mfaSetup := false
	prevScore := GetLevelScore(header.CurrentUser.Level)
	currentScore := header.CurrentUser.Score - prevScore
	nextScore := GetLevelScore(header.CurrentUser.Level+1) - prevScore
	perc := int(math.Ceil((float64(nextScore) / float64(currentScore)) * 100))

	accountPage := AccountDashPage{header, "dashboard", "account_own_edit", mfaSetup, currentScore, nextScore, user.Level + 1, perc * 2}
	accountTmpl, err := compile("account", "common.AccountDashPage", accountPage)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var writeTemplate = func(name string, content interface{}) {
		log.Print("Writing template '" + name + "'")

		var writeTmpl = func(name string, content string) {
			if content == "" {
				log.Fatal("No content body for " + name)
			}
			err := writeFile("./template_"+name+".go", content)
			if err != nil {
				log.Fatal(err)
			}
		}

		wg.Add(1)
		go func() {
			switch content := content.(type) {
			case string:
				writeTmpl(name, content)
			case TmplLoggedin:
				writeTmpl(name, content.Stub)
				writeTmpl(name+"_guest", content.Guest)
				writeTmpl(name+"_member", content.Member)
			}
			wg.Done()
		}()
	}

	// Let plugins register their own templates
	DebugLog("Registering the templates for the plugins")
	config = c.GetConfig()
	config.SkipHandles = true
	c.SetConfig(config)
	for _, tmplfunc := range PrebuildTmplList {
		tmplItem := tmplfunc(user, header)
		varList = make(map[string]tmpl.VarItem)
		compiledTmpl, err := c.Compile(tmplItem.Filename, tmplItem.Path, tmplItem.StructName, tmplItem.Data, varList, tmplItem.Imports...)
		if err != nil {
			return err
		}
		writeTemplate(tmplItem.Name, compiledTmpl)
	}

	log.Print("Writing the templates")
	writeTemplate("topic", topicTmpl)
	writeTemplate("topic_alt", topicAltTmpl)
	writeTemplate("profile", profileTmpl)
	writeTemplate("forums", forumsTmpl)
	writeTemplate("topics", topicListTmpl)
	writeTemplate("forum", forumTmpl)
	writeTemplate("login", loginTmpl)
	writeTemplate("register", registerTmpl)
	writeTemplate("ip_search", ipSearchTmpl)
	writeTemplate("account", accountTmpl)
	writeTemplate("error", errorTmpl)
	writeTemplateList(c, &wg, "./")
	return nil
}

func CompileJSTemplates() error {
	log.Print("Compiling the JS templates")
	var config tmpl.CTemplateConfig
	config.Minify = Config.MinifyTemplates
	config.Debug = Dev.DebugMode
	config.SuperDebug = Dev.TemplateDebug
	config.SkipHandles = true
	config.SkipTmplPtrMap = true
	config.SkipInitBlock = false
	config.PackageName = "tmpl"

	c := tmpl.NewCTemplateSet()
	c.SetConfig(config)
	c.SetBaseImportMap(map[string]string{
		"io": "io",
		"github.com/Azareal/Gosora/common/alerts": "github.com/Azareal/Gosora/common/alerts",
	})
	c.SetBuildTags("!no_templategen")

	user, user2, user3 := tmplInitUsers()
	header, _, _ := tmplInitHeaders(user, user2, user3)
	now := time.Now()
	var varList = make(map[string]tmpl.VarItem)

	// TODO: Check what sort of path is sent exactly and use it here
	alertItem := alerts.AlertItem{Avatar: "", ASID: 1, Path: "/", Message: "uh oh, something happened"}
	alertTmpl, err := c.Compile("alert.html", "templates/", "alerts.AlertItem", alertItem, varList)
	if err != nil {
		return err
	}

	c.SetBaseImportMap(map[string]string{
		"io":                               "io",
		"github.com/Azareal/Gosora/common": "github.com/Azareal/Gosora/common",
	})
	// TODO: Fix the import loop so we don't have to use this hack anymore
	c.SetBuildTags("!no_templategen,tmplgentopic")

	var topicsRow = &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, "Date", user3.ID, 1, "", "127.0.0.1", 1, 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"}
	topicListItemTmpl, err := c.Compile("topics_topic.html", "templates/", "*common.TopicsRow", topicsRow, varList)
	if err != nil {
		return err
	}

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		PollOption{0, "Nothing"},
		PollOption{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	topic := TopicUser{1, "blah", "Blah", "Hey there!", 62, false, false, now, RelativeTime(now), now, RelativeTime(now), 0, "", "127.0.0.1", 1, 0, 1, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false}
	var replyList []ReplyUser
	// TODO: Do we really want the UID here to be zero?
	avatar, microAvatar = BuildAvatar(0, "")
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, RelativeTime(now), 0, 0, avatar, microAvatar, "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, "", ""})

	varList = make(map[string]tmpl.VarItem)
	header.Title = "Topic Name"
	tpage := TopicPage{header, replyList, topic, &Forum{ID: 1, Name: "Hahaha"}, poll, 1, 1}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	topicPostsTmpl, err := c.Compile("topic_posts.html", "templates/", "common.TopicPage", tpage, varList)
	if err != nil {
		return err
	}
	topicAltPostsTmpl, err := c.Compile("topic_alt_posts.html", "templates/", "common.TopicPage", tpage, varList)
	if err != nil {
		return err
	}

	var dirPrefix = "./tmpl_client/"
	var wg sync.WaitGroup
	var writeTemplate = func(name string, content string) {
		log.Print("Writing template '" + name + "'")
		if content == "" {
			log.Fatal("No content body")
		}

		wg.Add(1)
		go func() {
			err := writeFile(dirPrefix+"template_"+name+".go", content)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
	}
	writeTemplate("alert", alertTmpl)
	writeTemplate("topics_topic", topicListItemTmpl)
	writeTemplate("topic_posts", topicPostsTmpl)
	writeTemplate("topic_alt_posts", topicAltPostsTmpl)
	writeTemplateList(c, &wg, dirPrefix)
	return nil
}

func writeTemplateList(c *tmpl.CTemplateSet, wg *sync.WaitGroup, prefix string) {
	log.Print("Writing template list")
	wg.Add(1)
	go func() {
		out := "package " + c.GetConfig().PackageName + "\n\n"
		var getterstr = "\n// nolint\nGetFrag = func(name string) [][]byte {\nswitch(name) {\n"
		for templateName, count := range c.TemplateFragmentCount {
			out += "var " + templateName + "_frags = make([][]byte," + strconv.Itoa(count) + ")\n"
			getterstr += "\tcase \"" + templateName + "\":\n"
			getterstr += "\treturn " + templateName + "_frags\n"
		}
		getterstr += "}\nreturn nil\n}\n"
		out += "\n// nolint\nfunc init() {\n"
		//var bodyMap = make(map[string]string) //map[body]fragmentPrefix
		var tmpCount = 0
		for _, frag := range c.FragOut {
			front := frag.TmplName + "_frags[" + strconv.Itoa(frag.Index) + "]"
			/*fp, ok := bodyMap[frag.Body]
			if !ok {
				bodyMap[frag.Body] = front*/
			var bits string
			for _, char := range []byte(frag.Body) {
				if char == '\'' {
					bits += "'\\" + string(char) + "',"
				} else {
					bits += "'" + string(char) + "',"
				}
			}
			tmpStr := strconv.Itoa(tmpCount)
			out += "arr_" + tmpStr + " := [...]byte{" + bits + "}\n"
			out += front + " = arr_" + tmpStr + "[:]\n"
			tmpCount++
			//out += front + " = []byte(`" + frag.Body + "`)\n"
			/*} else {
				out += front + " = " + fp + "\n"
			}*/
		}
		out += "\n" + getterstr + "}\n"
		err := writeFile(prefix+"template_list.go", out)
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()
	wg.Wait()
}

func arithToInt64(in interface{}) (out int64) {
	switch in := in.(type) {
	case int64:
		out = in
	case int32:
		out = int64(in)
	case int:
		out = int64(in)
	case uint32:
		out = int64(in)
	case uint16:
		out = int64(in)
	case uint8:
		out = int64(in)
	case uint:
		out = int64(in)
	}
	return out
}

func arithDuoToInt64(left interface{}, right interface{}) (leftInt int64, rightInt int64) {
	return arithToInt64(left), arithToInt64(right)
}

func InitTemplates() error {
	DebugLog("Initialising the template system")
	// TODO: Add support for floats
	fmap := make(map[string]interface{})
	fmap["add"] = func(left interface{}, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt + rightInt
	}

	fmap["subtract"] = func(left interface{}, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt - rightInt
	}

	fmap["multiply"] = func(left interface{}, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt * rightInt
	}

	fmap["divide"] = func(left interface{}, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		if leftInt == 0 || rightInt == 0 {
			return 0
		}
		return leftInt / rightInt
	}

	fmap["dock"] = func(dock interface{}, headerInt interface{}) interface{} {
		return template.HTML(BuildWidget(dock.(string), headerInt.(*Header)))
	}

	fmap["elapsed"] = func(startedAtInt interface{}) interface{} {
		return time.Since(startedAtInt.(time.Time)).String()
	}

	fmap["lang"] = func(phraseNameInt interface{}) interface{} {
		phraseName, ok := phraseNameInt.(string)
		if !ok {
			panic("phraseNameInt is not a string")
		}
		// TODO: Log non-existent phrases?
		return template.HTML(phrases.GetTmplPhrase(phraseName))
	}

	// TODO: Implement this in the template generator too
	fmap["langf"] = func(phraseNameInt interface{}, args ...interface{}) interface{} {
		phraseName, ok := phraseNameInt.(string)
		if !ok {
			panic("phraseNameInt is not a string")
		}
		// TODO: Log non-existent phrases?
		// TODO: Optimise TmplPhrasef so we don't use slow Sprintf there
		return template.HTML(phrases.GetTmplPhrasef(phraseName, args...))
	}

	fmap["level"] = func(levelInt interface{}) interface{} {
		level, ok := levelInt.(int)
		if !ok {
			panic("levelInt is not an integer")
		}
		return template.HTML(phrases.GetLevelPhrase(level))
	}

	fmap["abstime"] = func(timeInt interface{}) interface{} {
		time, ok := timeInt.(time.Time)
		if !ok {
			panic("timeInt is not a time.Time")
		}
		//return time.String()
		return time.Format("2006-01-02 15:04:05")
	}

	fmap["scope"] = func(name interface{}) interface{} {
		return ""
	}

	fmap["dyntmpl"] = func(nameInt interface{}, pageInt interface{}, headerInt interface{}) interface{} {
		header := headerInt.(*Header)
		err := header.Theme.RunTmpl(nameInt.(string), pageInt, header.Writer)
		if err != nil {
			return err
		}
		return ""
	}

	// The interpreted templates...
	DebugLog("Loading the template files...")
	Templates.Funcs(fmap)
	templateFiles, err := filepath.Glob("templates/*.html")
	if err != nil {
		return err
	}

	var templateFileMap = make(map[string]int)
	for index, path := range templateFiles {
		path = strings.Replace(path, "\\", "/", -1)
		log.Print("templateFile: ", path)
		if skipCTmpl(path) {
			log.Print("skipping")
			continue
		}
		templateFileMap[path] = index
	}

	overrideFiles, err := filepath.Glob("templates/overrides/*.html")
	if err != nil {
		return err
	}
	for _, path := range overrideFiles {
		path = strings.Replace(path, "\\", "/", -1)
		log.Print("overrideFile: ", path)
		if skipCTmpl(path) {
			log.Print("skipping")
			continue
		}
		index, ok := templateFileMap["templates/"+strings.TrimPrefix(path, "templates/overrides/")]
		if !ok {
			log.Print("not ok: templates/" + strings.TrimPrefix(path, "templates/overrides/"))
			templateFiles = append(templateFiles, path)
			continue
		}
		templateFiles[index] = path
	}
	template.Must(Templates.ParseFiles(templateFiles...))
	template.Must(Templates.ParseGlob("pages/*"))

	return nil
}
