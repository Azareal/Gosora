package common

import (
	"html/template"
	"io"
	"log"
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
var DefaultTemplates = template.New("")
var DefaultTemplateFuncMap map[string]interface{}

//var Templates = template.New("")
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

func genIntTmpl(name string) func(pi interface{}, w io.Writer) error {
	return func(pi interface{}, w io.Writer) error {
		mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap[name]
		if !ok {
			mapping = name
		}
		return DefaultTemplates.ExecuteTemplate(w, mapping+".html", pi)
	}
}

// TODO: Refactor the template trees to not need these
// nolint
var Template_topic_handle = genIntTmpl("topic")
var Template_topic_guest_handle = Template_topic_handle
var Template_topic_member_handle = Template_topic_handle
var Template_topic_alt_handle = genIntTmpl("topic")
var Template_topic_alt_guest_handle = Template_topic_alt_handle
var Template_topic_alt_member_handle = Template_topic_alt_handle

// nolint
var Template_topics_handle = genIntTmpl("topics")
var Template_topics_guest_handle = Template_topics_handle
var Template_topics_member_handle = Template_topics_handle

// nolint
var Template_forum_handle = genIntTmpl("forum")
var Template_forum_guest_handle = Template_forum_handle
var Template_forum_member_handle = Template_forum_handle

// nolint
var Template_forums_handle = genIntTmpl("forums")
var Template_forums_guest_handle = Template_forums_handle
var Template_forums_member_handle = Template_forums_handle

// nolint
var Template_profile_handle = genIntTmpl("profile")
var Template_profile_guest_handle = Template_profile_handle
var Template_profile_member_handle = Template_profile_handle

// nolint
var Template_create_topic_handle = genIntTmpl("create_topic")
var Template_login_handle = genIntTmpl("login")
var Template_register_handle = genIntTmpl("register")
var Template_error_handle = genIntTmpl("error")
var Template_ip_search_handle = genIntTmpl("ip_search")
var Template_account_handle = genIntTmpl("account")

func tmplInitUsers() (User, User, User) {
	avatar, microAvatar := BuildAvatar(62, "")
	user := User{62, BuildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", avatar, microAvatar, "", "", "", "", 0, 0, 0, "0.0.0.0.0", "", 0}

	// TODO: Do a more accurate level calculation for this?
	avatar, microAvatar = BuildAvatar(1, "")
	user2 := User{1, BuildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", "", "", 58, 1000, 0, "127.0.0.1", "", 0}

	avatar, microAvatar = BuildAvatar(2, "")
	user3 := User{2, BuildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", "", "", 42, 900, 0, "::1", "", 0}
	return user, user2, user3
}

func tmplInitHeaders(user User, user2 User, user3 User) (*Header, *Header, *Header) {
	header := &Header{
		Site:            Site,
		Settings:        SettingBox.Load().(SettingMap),
		Themes:          Themes,
		Theme:           Themes[DefaultThemeBox.Load().(string)],
		CurrentUser:     user,
		NoticeList:      []string{"test"},
		Stylesheets:     []string{"panel.css"},
		Scripts:         []string{"whatever.js"},
		PreScriptsAsync: []string{"whatever.js"},
		ScriptsAsync:    []string{"whatever.js"},
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

type nobreak interface{}

type TItem struct {
	Expects    string
	ExpectsInt interface{}
	LoggedIn   bool
}

type TItemHold map[string]TItem

func (hold TItemHold) Add(name string, expects string, expectsInt interface{}) {
	hold[name] = TItem{expects, expectsInt, true}
}

func (hold TItemHold) AddStd(name string, expects string, expectsInt interface{}) {
	hold[name] = TItem{expects, expectsInt, false}
}

// ? - Add template hooks?
func CompileTemplates() error {
	log.Print("Compiling the templates")
	// TODO: Implement per-theme template overrides here too
	var overriden = make(map[string]map[string]bool)
	for _, theme := range Themes {
		overriden[theme.Name] = make(map[string]bool)
		log.Printf("theme.OverridenTemplates: %+v\n", theme.OverridenTemplates)
		for _, override := range theme.OverridenTemplates {
			overriden[theme.Name][override] = true
		}
	}
	log.Printf("overriden: %+v\n", overriden)

	var config tmpl.CTemplateConfig
	config.Minify = Config.MinifyTemplates
	config.Debug = Dev.DebugMode
	config.SuperDebug = Dev.TemplateDebug

	c := tmpl.NewCTemplateSet("normal")
	c.SetConfig(config)
	c.SetBaseImportMap(map[string]string{
		"io":                               "io",
		"github.com/Azareal/Gosora/common": "github.com/Azareal/Gosora/common",
	})
	c.SetBuildTags("!no_templategen")
	c.SetOverrideTrack(overriden)
	c.SetPerThemeTmpls(make(map[string]bool))

	log.Print("Compiling the default templates")
	var wg sync.WaitGroup
	err := compileTemplates(&wg, c, "")
	if err != nil {
		return err
	}
	oroots := c.GetOverridenRoots()
	log.Printf("oroots: %+v\n", oroots)

	log.Print("Compiling the per-theme templates")
	for theme, tmpls := range oroots {
		c.ResetLogs("normal-" + theme)
		c.SetThemeName(theme)
		c.SetPerThemeTmpls(tmpls)
		log.Print("theme: ", theme)
		log.Printf("perThemeTmpls: %+v\n", tmpls)
		err = compileTemplates(&wg, c, theme)
		if err != nil {
			return err
		}
	}
	writeTemplateList(c, &wg, "./")
	return nil
}

func compileCommons(c *tmpl.CTemplateSet, header *Header, header2 *Header, out TItemHold) error {
	// TODO: Add support for interface{}s
	_, user2, user3 := tmplInitUsers()
	now := time.Now()

	// Convienience function to save a line here and there
	var htitle = func(name string) *Header {
		header.Title = name
		return header
	}
	/*var htitle2 = func(name string) *Header {
		header2.Title = name
		return header2
	}*/

	// TODO: Use a dummy forum list to avoid o(n) problems
	var forumList []Forum
	forums, err := Forums.GetAll()
	if err != nil {
		return err
	}
	for _, forum := range forums {
		forumList = append(forumList, *forum)
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList, &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, user3.ID, 1, 1, "", "127.0.0.1", 1, 0, 1, 1, 0, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"})
	topicListPage := TopicListPage{htitle("Topic List"), topicsList, forumList, Config.DefaultForum, TopicListSort{"lastupdated", false}, Paginator{[]int{1}, 1, 1}}
	out.Add("topics", "common.TopicListPage", topicListPage)

	forumItem := BlankForum(1, "general-forum.1", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0)
	forumPage := ForumPage{htitle("General Forum"), topicsList, forumItem, Paginator{[]int{1}, 1, 1}}
	out.Add("forum", "common.ForumPage", forumPage)
	out.Add("forums", "common.ForumsPage", ForumsPage{htitle("Forum List"), forumList})

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		PollOption{0, "Nothing"},
		PollOption{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{&MiniAttachment{Path: "/"}}
	topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, now, 1, 1, 0, "", "127.0.0.1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false, miniAttach}
	var replyList []ReplyUser
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, 0, 0, avatar, microAvatar, "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, 1, "", "", miniAttach})
	tpage := TopicPage{htitle("Topic Name"), replyList, topic, &Forum{ID: 1, Name: "Hahaha"}, poll, Paginator{[]int{1}, 1, 1}}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	out.Add("topic", "common.TopicPage", tpage)
	out.Add("topic_alt", "common.TopicPage", tpage)
	return nil
}

func compileTemplates(wg *sync.WaitGroup, c *tmpl.CTemplateSet, themeName string) error {
	// Schemas to train the template compiler on what to expect
	// TODO: Add support for interface{}s
	user, user2, user3 := tmplInitUsers()
	header, header2, _ := tmplInitHeaders(user, user2, user3)
	now := time.Now()

	/*poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		PollOption{0, "Nothing"},
		PollOption{1, "Something"},
	}, VoteCount: 7}*/
	avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{&MiniAttachment{Path: "/"}}
	var replyList []ReplyUser
	//topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, now, 1, 1, 0, "", "127.0.0.1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false, miniAttach}
	// TODO: Do we want the UID on this to be 0?
	avatar, microAvatar = BuildAvatar(0, "")
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, 0, 0, avatar, microAvatar, "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, 1, "", "", miniAttach})

	// Convienience function to save a line here and there
	var htitle = func(name string) *Header {
		header.Title = name
		return header
	}
	tmpls := TItemHold(make(map[string]TItem))
	err := compileCommons(c, header, header2, tmpls)
	if err != nil {
		return err
	}

	ppage := ProfilePage{htitle("User 526"), replyList, user, 0, 0} // TODO: Use the score from user to generate the currentScore and nextScore
	tmpls.Add("profile", "common.ProfilePage", ppage)

	tmpls.AddStd("login", "common.Page", Page{htitle("Login Page"), tList, nil})
	tmpls.AddStd("register", "common.Page", Page{htitle("Registration Page"), tList, "nananana"})
	tmpls.AddStd("error", "common.ErrorPage", ErrorPage{htitle("Error"), "A problem has occurred in the system."})

	ipSearchPage := IPSearchPage{htitle("IP Search"), map[int]*User{1: &user2}, "::1"}
	tmpls.AddStd("ip_search", "common.IPSearchPage", ipSearchPage)

	var inter nobreak
	accountPage := Account{header, "dashboard", "account_own_edit", inter}
	tmpls.AddStd("account", "common.Account", accountPage)

	basePage := &BasePanelPage{header, PanelStats{}, "dashboard", ReportForumID}
	tmpls.AddStd("panel", "common.Panel", Panel{basePage, "panel_dashboard_right","panel_dashboard", inter})

	var writeTemplate = func(name string, content interface{}) {
		log.Print("Writing template '" + name + "'")
		var writeTmpl = func(name string, content string) {
			if content == "" {
				return //log.Fatal("No content body for " + name)
			}
			err := writeFile("./template_"+name+".go", content)
			if err != nil {
				log.Fatal(err)
			}
		}
		wg.Add(1)
		go func() {
			tname := themeName
			if tname != "" {
				tname = "_" + tname
			}
			switch content := content.(type) {
			case string:
				writeTmpl(name+tname, content)
			case TmplLoggedin:
				writeTmpl(name+tname, content.Stub)
				writeTmpl(name+tname+"_guest", content.Guest)
				writeTmpl(name+tname+"_member", content.Member)
			}
			wg.Done()
		}()
	}

	// Let plugins register their own templates
	DebugLog("Registering the templates for the plugins")
	config := c.GetConfig()
	config.SkipHandles = true
	c.SetConfig(config)
	for _, tmplfunc := range PrebuildTmplList {
		tmplItem := tmplfunc(user, header)
		varList := make(map[string]tmpl.VarItem)
		compiledTmpl, err := c.Compile(tmplItem.Filename, tmplItem.Path, tmplItem.StructName, tmplItem.Data, varList, tmplItem.Imports...)
		if err != nil {
			return err
		}
		writeTemplate(tmplItem.Name, compiledTmpl)
	}

	log.Print("Writing the templates")
	for name, titem := range tmpls {
		log.Print("Writing " + name)
		varList := make(map[string]tmpl.VarItem)
		if titem.LoggedIn {
			stub, guest, member, err := c.CompileByLoggedin(name+".html", "templates/", titem.Expects, titem.ExpectsInt, varList)
			if err != nil {
				return err
			}
			writeTemplate(name, TmplLoggedin{stub, guest, member})
		} else {
			tmpl, err := c.Compile(name+".html", "templates/", titem.Expects, titem.ExpectsInt, varList)
			if err != nil {
				return err
			}
			writeTemplate(name, tmpl)
		}
	}
	/*writeTemplate("login", loginTmpl)
	writeTemplate("register", registerTmpl)
	writeTemplate("ip_search", ipSearchTmpl)
	writeTemplate("error", errorTmpl)*/
	return nil
}

// ? - Add template hooks?
func CompileJSTemplates() error {
	log.Print("Compiling the JS templates")
	// TODO: Implement per-theme template overrides here too
	var overriden = make(map[string]map[string]bool)
	for _, theme := range Themes {
		overriden[theme.Name] = make(map[string]bool)
		log.Printf("theme.OverridenTemplates: %+v\n", theme.OverridenTemplates)
		for _, override := range theme.OverridenTemplates {
			overriden[theme.Name][override] = true
		}
	}
	log.Printf("overriden: %+v\n", overriden)

	var config tmpl.CTemplateConfig
	config.Minify = Config.MinifyTemplates
	config.Debug = Dev.DebugMode
	config.SuperDebug = Dev.TemplateDebug
	config.SkipHandles = true
	config.SkipTmplPtrMap = true
	config.SkipInitBlock = false
	config.PackageName = "tmpl"

	c := tmpl.NewCTemplateSet("js")
	c.SetConfig(config)
	c.SetBuildTags("!no_templategen")
	c.SetOverrideTrack(overriden)
	c.SetPerThemeTmpls(make(map[string]bool))

	log.Print("Compiling the default templates")
	var wg sync.WaitGroup
	err := compileJSTemplates(&wg, c, "")
	if err != nil {
		return err
	}
	oroots := c.GetOverridenRoots()
	log.Printf("oroots: %+v\n", oroots)

	log.Print("Compiling the per-theme templates")
	for theme, tmpls := range oroots {
		c.SetThemeName(theme)
		c.SetPerThemeTmpls(tmpls)
		log.Print("theme: ", theme)
		log.Printf("perThemeTmpls: %+v\n", tmpls)
		err = compileJSTemplates(&wg, c, theme)
		if err != nil {
			return err
		}
	}
	var dirPrefix = "./tmpl_client/"
	writeTemplateList(c, &wg, dirPrefix)
	return nil
}

func compileJSTemplates(wg *sync.WaitGroup, c *tmpl.CTemplateSet, themeName string) error {
	user, user2, user3 := tmplInitUsers()
	header, _, _ := tmplInitHeaders(user, user2, user3)
	now := time.Now()
	var varList = make(map[string]tmpl.VarItem)

	c.SetBaseImportMap(map[string]string{
		"io": "io",
		"github.com/Azareal/Gosora/common/alerts": "github.com/Azareal/Gosora/common/alerts",
	})

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

	tmpls := TItemHold(make(map[string]TItem))

	var topicsRow = &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, user3.ID, 1, 1, "", "127.0.0.1", 1, 0, 1, 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"}
	tmpls.AddStd("topics_topic", "common.TopicsRow", topicsRow)

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		PollOption{0, "Nothing"},
		PollOption{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{&MiniAttachment{Path: "/"}}
	topic := TopicUser{1, "blah", "Blah", "Hey there!", 62, false, false, now, now, 1, 1, 0, "", "127.0.0.1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false, miniAttach}
	var replyList []ReplyUser
	// TODO: Do we really want the UID here to be zero?
	avatar, microAvatar = BuildAvatar(0, "")
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, 0, 0, avatar, microAvatar, "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, 1, "", "", miniAttach})

	varList = make(map[string]tmpl.VarItem)
	header.Title = "Topic Name"
	tpage := TopicPage{header, replyList, topic, &Forum{ID: 1, Name: "Hahaha"}, poll, Paginator{[]int{1}, 1, 1}}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	tmpls.AddStd("topic_posts", "common.TopicPage", tpage)
	tmpls.AddStd("topic_alt_posts", "common.TopicPage", tpage)

	itemsPerPage := 25
	_, page, lastPage := PageOffset(20, 1, itemsPerPage)
	pageList := Paginate(20, itemsPerPage, 5)
	tmpls.AddStd("paginator", "common.Paginator", Paginator{pageList, page, lastPage})

	tmpls.AddStd("topic_c_edit_post", "common.TopicCEditPost", TopicCEditPost{ID: 0, Source: "", Ref: ""})

	tmpls.AddStd("topic_c_attach_item", "common.TopicCAttachItem", TopicCAttachItem{ID: 1, ImgSrc: "", Path: "", FullPath: ""})

	tmpls.AddStd("notice", "string", "nonono")

	var dirPrefix = "./tmpl_client/"
	var writeTemplate = func(name string, content string) {
		log.Print("Writing template '" + name + "'")
		if content == "" {
			return //log.Fatal("No content body")
		}
		wg.Add(1)
		go func() {
			tname := themeName
			if tname != "" {
				tname = "_" + tname
			}
			err := writeFile(dirPrefix+"template_"+name+tname+".jgo", content)
			if err != nil {
				log.Fatal(err)
			}
			wg.Done()
		}()
	}

	log.Print("Writing the templates")
	for name, titem := range tmpls {
		log.Print("Writing " + name)
		varList := make(map[string]tmpl.VarItem)
		tmpl, err := c.Compile(name+".html", "templates/", titem.Expects, titem.ExpectsInt, varList)
		if err != nil {
			return err
		}
		writeTemplate(name, tmpl)
	}
	writeTemplate("alert", alertTmpl)
	/*//writeTemplate("forum", forumTmpl)
	writeTemplate("topic_posts", topicPostsTmpl)
	writeTemplate("topic_alt_posts", topicAltPostsTmpl)
	writeTemplateList(c, &wg, dirPrefix)*/
	return nil
}

func getTemplateList(c *tmpl.CTemplateSet, wg *sync.WaitGroup, prefix string) string {
	DebugLog("in getTemplateList")
	pout := "\n// nolint\nfunc init() {\n"
	var tFragCount = make(map[string]int)
	var bodyMap = make(map[string]string) //map[body]fragmentPrefix
	//var tmplMap = make(map[string]map[string]string) // map[tmpl]map[body]fragmentPrefix
	var tmpCount = 0
	for _, frag := range c.FragOut {
		front := frag.TmplName + "_frags[" + strconv.Itoa(frag.Index) + "]"
		DebugLog("front: ", front)
		DebugLog("frag.Body: ", frag.Body)
		/*bodyMap, tok := tmplMap[frag.TmplName]
		if !tok {
			tmplMap[frag.TmplName] = make(map[string]string)
			bodyMap = tmplMap[frag.TmplName]
		}*/
		fp, ok := bodyMap[frag.Body]
		if !ok {
			bodyMap[frag.Body] = front
			var bits string
			DebugLog("encoding frag.Body")
			for _, char := range []byte(frag.Body) {
				if char == '\'' {
					bits += "'\\" + string(char) + "',"
				} else if char < 32 {
					bits += strconv.Itoa(int(char)) + ","
				} else {
					bits += "'" + string(char) + "',"
				}
			}
			tmpStr := strconv.Itoa(tmpCount)
			pout += "arr_" + tmpStr + " := [...]byte{" + bits + "}\n"
			pout += front + " = arr_" + tmpStr + "[:]\n"
			tmpCount++
			//pout += front + " = []byte(`" + frag.Body + "`)\n"
		} else {
			DebugLog("encoding cached index " + fp)
			pout += front + " = " + fp + "\n"
		}

		_, ok = tFragCount[frag.TmplName]
		if !ok {
			tFragCount[frag.TmplName] = 0
		}
		tFragCount[frag.TmplName]++
	}

	out := "package " + c.GetConfig().PackageName + "\n\n"
	var getterstr = "\n// nolint\nGetFrag = func(name string) [][]byte {\nswitch(name) {\n"
	for templateName, count := range tFragCount {
		out += "var " + templateName + "_frags = make([][]byte," + strconv.Itoa(count) + ")\n"
		getterstr += "\tcase \"" + templateName + "\":\n"
		getterstr += "\treturn " + templateName + "_frags\n"
	}
	getterstr += "}\nreturn nil\n}\n"
	out += pout + "\n" + getterstr + "}\n"

	return out
}

func writeTemplateList(c *tmpl.CTemplateSet, wg *sync.WaitGroup, prefix string) {
	log.Print("Writing template list")
	wg.Add(1)
	go func() {
		err := writeFile(prefix+"template_list.go", getTemplateList(c, wg, prefix))
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

func initDefaultTmplFuncMap() {
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

	fmap["hasWidgets"] = func(dock interface{}, headerInt interface{}) interface{} {
		return HasWidgets(dock.(string), headerInt.(*Header))
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
		return time.Format("2006-01-02 15:04:05")
	}

	fmap["reltime"] = func(timeInt interface{}) interface{} {
		time, ok := timeInt.(time.Time)
		if !ok {
			panic("timeInt is not a time.Time")
		}
		return RelativeTime(time)
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

	fmap["flush"] = func() interface{} {
		return nil
	}

	DefaultTemplateFuncMap = fmap
}

func loadTemplates(tmpls *template.Template, themeName string) error {
	tmpls.Funcs(DefaultTemplateFuncMap)
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

	if themeName != "" {
		overrideFiles, err := filepath.Glob("./themes/" + themeName + "/overrides/*.html")
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
			index, ok := templateFileMap["templates/"+strings.TrimPrefix(path, "themes/"+themeName+"/overrides/")]
			if !ok {
				log.Print("not ok: templates/" + strings.TrimPrefix(path, "themes/"+themeName+"/overrides/"))
				templateFiles = append(templateFiles, path)
				continue
			}
			templateFiles[index] = path
		}
	}

	template.Must(tmpls.ParseFiles(templateFiles...))
	template.Must(tmpls.ParseGlob("pages/*"))
	return nil
}

func InitTemplates() error {
	DebugLog("Initialising the template system")
	initDefaultTmplFuncMap()

	// The interpreted templates...
	DebugLog("Loading the template files...")
	return loadTemplates(DefaultTemplates, "")
}
