package common

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azareal/Gosora/common/alerts"
	p "github.com/Azareal/Gosora/common/phrases"
	tmpl "github.com/Azareal/Gosora/common/templates"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/uutils"
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
		theme := Themes[DefaultThemeBox.Load().(string)]
		mapping, ok := theme.TemplatesMap[name]
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

func tmplInitUsers() (*User, *User, *User) {
	avatar, microAvatar := BuildAvatar(62, "")
	u := User{62, BuildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", avatar, microAvatar, "", "", 0, 0, 0, 0, StartTime, "0.0.0.0.0", 0, 0, nil, UserPrivacy{}}

	// TODO: Do a more accurate level calculation for this?
	avatar, microAvatar = BuildAvatar(1, "")
	u2 := User{1, BuildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", 58, 1000, 0, 1000, StartTime, "127.0.0.1", 0, 0, nil, UserPrivacy{}}

	avatar, microAvatar = BuildAvatar(2, "")
	u3 := User{2, BuildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", avatar, microAvatar, "", "", 42, 900, 0, 900, StartTime, "::1", 0, 0, nil, UserPrivacy{}}
	return &u, &u2, &u3
}

func tmplInitHeaders(u, u2, u3 *User) (*Header, *Header, *Header) {
	header := &Header{
		Site:            Site,
		Settings:        SettingBox.Load().(SettingMap),
		Themes:          Themes,
		Theme:           Themes[DefaultThemeBox.Load().(string)],
		CurrentUser:     u,
		NoticeList:      []string{"test"},
		Stylesheets:     []HScript{{"panel.css", ""}},
		Scripts:         []HScript{{"whatever.js", ""}},
		PreScriptsAsync: []HScript{{"whatever.js", ""}},
		ScriptsAsync:    []HScript{{"whatever.js", ""}},
		Widgets: PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	buildHeader := func(u *User) *Header {
		head := &Header{Site: Site}
		*head = *header
		head.CurrentUser = u
		return head
	}

	return header, buildHeader(u2), buildHeader(u3)
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

func (h TItemHold) Add(name, expects string, expectsInt interface{}) {
	h[name] = TItem{expects, expectsInt, true}
}

func (h TItemHold) AddStd(name, expects string, expectsInt interface{}) {
	h[name] = TItem{expects, expectsInt, false}
}

// ? - Add template hooks?
func CompileTemplates() error {
	log.Print("Compiling the templates")
	// TODO: Implement per-theme template overrides here too
	overriden := make(map[string]map[string]bool)
	for _, th := range Themes {
		overriden[th.Name] = make(map[string]bool)
		log.Printf("th.OverridenTemplates: %+v\n", th.OverridenTemplates)
		for _, override := range th.OverridenTemplates {
			overriden[th.Name][override] = true
		}
	}
	log.Printf("overriden: %+v\n", overriden)

	config := tmpl.CTemplateConfig{
		Minify:     Config.MinifyTemplates,
		Debug:      Dev.DebugMode,
		SuperDebug: Dev.TemplateDebug,
		DockToID:   DockToID,
	}
	c := tmpl.NewCTemplateSet("normal", "./logs/")
	c.SetConfig(config)
	c.SetBaseImportMap(map[string]string{
		"io":                               "io",
		"github.com/Azareal/Gosora/common": "c github.com/Azareal/Gosora/common",
	})
	c.SetBuildTags("!no_templategen")
	c.SetOverrideTrack(overriden)
	c.SetPerThemeTmpls(make(map[string]bool))

	log.Print("Compiling the default templates")
	var wg sync.WaitGroup
	if err := compileTemplates(&wg, c, ""); err != nil {
		return err
	}
	oroots := c.GetOverridenRoots()
	log.Printf("oroots: %+v\n", oroots)

	log.Print("Compiling the per-theme templates")
	for th, tmpls := range oroots {
		c.ResetLogs("normal-" + th)
		c.SetThemeName(th)
		c.SetPerThemeTmpls(tmpls)
		log.Print("th: ", th)
		log.Printf("perThemeTmpls: %+v\n", tmpls)
		err := compileTemplates(&wg, c, th)
		if err != nil {
			return err
		}
	}
	writeTemplateList(c, &wg, "./")
	return nil
}

func compileCommons(c *tmpl.CTemplateSet, head, head2 *Header, forumList []Forum, o TItemHold) error {
	// TODO: Add support for interface{}s
	_, user2, user3 := tmplInitUsers()
	now := time.Now()

	// Convienience function to save a line here and there
	htitle := func(name string) *Header {
		head.Title = name
		return head
	}
	/*htitle2 := func(name string) *Header {
		head2.Title = name
		return head2
	}*/

	var topicsList []TopicsRowMut
	topic := Topic{1, "/topic/topic-title.1", "Topic Title", "The topic content.", 1, false, false, now, now, user3.ID, 1, 1, "", "::1", 1, 0, 1, 1, 1, "classname", 0, "", nil}
	topicsList = append(topicsList, TopicsRowMut{&TopicsRow{topic, 1, user2, "", 0, user3, "General", "/forum/general.2"}, false})
	topicListPage := TopicListPage{htitle("Topic List"), topicsList, forumList, Config.DefaultForum, TopicListSort{"lastupdated", false}, []int{1}, QuickTools{false, false, false}, Paginator{[]int{1}, 1, 1}}
	o.Add("topics", "c.TopicListPage", topicListPage)
	o.Add("topics_mini", "c.TopicListPage", topicListPage)

	forumItem := BlankForum(1, "general-forum.1", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0)
	forumPage := ForumPage{htitle("General Forum"), topicsList, forumItem, false, false, Paginator{[]int{1}, 1, 1}}
	o.Add("forum", "c.ForumPage", forumPage)
	o.Add("forums", "c.ForumsPage", ForumsPage{htitle("Forum List"), forumList})

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		{0, "Nothing"},
		{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{{Path: "/"}}
	tu := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, now, 1, 1, 0, "", "127.0.0.1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", 58, false, miniAttach, nil, false}

	var replyList []*ReplyUser
	reply := Reply{1, 1, "Yo!", 1 /*, Config.DefaultGroup*/, now, 0, 0, 1, "::1", true, 1, 1, ""}
	ru := &ReplyUser{ClassName: "", Reply: reply, CreatedByName: "Alice", Avatar: avatar, Group: Config.DefaultGroup, Level: 0, Attachments: miniAttach}
	_, err := ru.Init(user2)
	if err != nil {
		return err
	}
	replyList = append(replyList, ru)
	tpage := TopicPage{htitle("Topic Name"), replyList, tu, &Forum{ID: 1, Name: "Hahaha"}, &poll, Paginator{[]int{1}, 1, 1}}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	o.Add("topic", "c.TopicPage", tpage)
	o.Add("topic_mini", "c.TopicPage", tpage)
	o.Add("topic_alt", "c.TopicPage", tpage)
	o.Add("topic_alt_mini", "c.TopicPage", tpage)
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
	//avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{{Path: "/"}}
	var replyList []*ReplyUser
	//topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, now, 1, 1, 0, "", "127.0.0.1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", "", "", 58, false, miniAttach, nil}
	// TODO: Do we want the UID on this to be 0?
	//avatar, microAvatar = BuildAvatar(0, "")
	reply := Reply{1, 1, "Yo!", 1 /*, Config.DefaultGroup*/, now, 0, 0, 1, "::1", true, 1, 1, ""}
	ru := &ReplyUser{ClassName: "", Reply: reply, CreatedByName: "Alice", Avatar: "", Group: Config.DefaultGroup, Level: 0, Attachments: miniAttach}
	_, err := ru.Init(user)
	if err != nil {
		return err
	}
	replyList = append(replyList, ru)

	forum := BlankForum(1, "/forum/d.1", "d", "d desc", true, "", 0, "", 1)
	forum.LastTopic = BlankTopic()
	forum.LastReplyer = BlankUser()
	forumList := []Forum{*forum}

	// Convienience function to save a line here and there
	htitle := func(name string) *Header {
		header.Title = name
		return header
	}
	t := TItemHold(make(map[string]TItem))
	err = compileCommons(c, header, header2, forumList, t)
	if err != nil {
		return err
	}

	ppage := ProfilePage{htitle("User 526"), replyList, *user, 0, 0, false, false, false, false} // TODO: Use the score from user to generate the currentScore and nextScore
	t.Add("profile", "c.ProfilePage", ppage)

	var topicsList []TopicsRowMut
	topic := Topic{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, user3.ID, 1, 1, "", "::1", 1, 0, 1, 1, 1, "classname", 0, "", nil}
	topicsList = append(topicsList, TopicsRowMut{&TopicsRow{topic, 0, user2, "", 0, user3, "General", "/forum/general.2"}, false})
	topicListPage := TopicListPage{htitle("Topic List"), topicsList, forumList, Config.DefaultForum, TopicListSort{"lastupdated", false}, []int{1}, QuickTools{false, false, false}, Paginator{[]int{1}, 1, 1}}

	forumItem := BlankForum(1, "general-forum.1", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0)
	forumPage := ForumPage{htitle("General Forum"), topicsList, forumItem, false, false, Paginator{[]int{1}, 1, 1}}

	// Experimental!
	for _, tmpl := range strings.Split(Dev.ExtraTmpls, ",") {
		sp := strings.Split(tmpl, ":")
		if len(sp) < 2 {
			continue
		}
		typ := "0"
		if len(sp) == 3 {
			typ = sp[2]
		}

		var pi interface{}
		switch sp[1] {
		case "c.TopicListPage":
			pi = topicListPage
		case "c.ForumPage":
			pi = forumPage
		case "c.ProfilePage":
			pi = ppage
		case "c.Page":
			pi = Page{htitle("Something"), tList, nil}
		default:
			continue
		}

		if typ == "1" {
			t.Add(sp[0], sp[1], pi)
		} else {
			t.AddStd(sp[0], sp[1], pi)
		}
	}

	t.AddStd("login", "c.Page", Page{htitle("Login Page"), tList, nil})
	t.AddStd("register", "c.RegisterPage", RegisterPage{htitle("Registration Page"), false, "", []RegisterVerify{{true, &RegisterVerifyImageGrid{"What?", []RegisterVerifyImageGridImage{{"something.png"}}}}}})
	t.AddStd("error", "c.ErrorPage", ErrorPage{htitle("Error"), "A problem has occurred in the system."})

	ipSearchPage := IPSearchPage{htitle("IP Search"), map[int]*User{1: user2}, "::1"}
	t.AddStd("ip_search", "c.IPSearchPage", ipSearchPage)

	var inter nobreak
	accountPage := Account{header, "dashboard", "account_own_edit", inter}
	t.AddStd("account", "c.Account", accountPage)

	parti := []*User{user}
	convo := &Conversation{1, BuildConvoURL(1), user.ID, time.Now(), 0, time.Now()}
	convoItems := []ConvoViewRow{{&ConversationPost{1, 1, "hey", "", user.ID}, user, "", 4, true}}
	convoPage := ConvoViewPage{header, convo, convoItems, parti, true, Paginator{[]int{1}, 1, 1}}
	t.AddStd("convo", "c.ConvoViewPage", convoPage)

	convos := []*ConversationExtra{{&Conversation{}, []*User{user}}}
	var cRows []ConvoListRow
	for _, convo := range convos {
		cRows = append(cRows, ConvoListRow{convo, convo.Users, false})
	}
	convoListPage := ConvoListPage{header, cRows, Paginator{[]int{1}, 1, 1}}
	t.AddStd("convos", "c.ConvoListPage", convoListPage)

	basePage := &BasePanelPage{header, PanelStats{}, "dashboard", ReportForumID, true}
	t.AddStd("panel", "c.Panel", Panel{basePage, "panel_dashboard_right", "", "panel_dashboard", inter})
	ges := []GridElement{{"", "", "", 1, "grid_istat", "", "", ""}}
	t.AddStd("panel_dashboard", "c.DashGrids", DashGrids{ges, ges})

	goVersion := runtime.Version()
	dbVersion := qgen.Builder.DbVersion()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	debugTasks := DebugPageTasks{0, 0, 0, 0, 0}
	debugCache := DebugPageCache{1, 1, 1, 2, 2, 2, true}
	debugDatabase := DebugPageDatabase{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	debugDisk := DebugPageDisk{1, 1, 1, 1, 1, 1}
	dpage := PanelDebugPage{basePage, goVersion, dbVersion, "0s", 1, qgen.Builder.GetAdapter().GetName(), 1, 1, 1, debugTasks, memStats, debugCache, debugDatabase, debugDisk}
	t.AddStd("panel_debug", "c.PanelDebugPage", dpage)
	//t.AddStd("panel_analytics", "c.PanelAnalytics", Panel{basePage, "panel_dashboard_right","panel_dashboard", inter})

	writeTemplate := func(name string, content interface{}) {
		log.Print("Writing template '" + name + "'")
		writeTmpl := func(name, content string) {
			if content == "" {
				return //log.Fatal("No content body for " + name)
			}
			e := writeFile("./tmpl_"+name+".go", content)
			if e != nil {
				log.Fatal(e)
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
		tmplItem := tmplfunc(*user, header)
		varList := make(map[string]tmpl.VarItem)
		compiledTmpl, err := c.Compile(tmplItem.Filename, tmplItem.Path, tmplItem.StructName, tmplItem.Data, varList, tmplItem.Imports...)
		if err != nil {
			return err
		}
		writeTemplate(tmplItem.Name, compiledTmpl)
	}

	log.Print("Writing the templates")
	for name, titem := range t {
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

	return nil
}

// ? - Add template hooks?
func CompileJSTemplates() error {
	log.Print("Compiling the JS templates")
	// TODO: Implement per-theme template overrides here too
	overriden := make(map[string]map[string]bool)
	for _, theme := range Themes {
		overriden[theme.Name] = make(map[string]bool)
		log.Printf("theme.OverridenTemplates: %+v\n", theme.OverridenTemplates)
		for _, override := range theme.OverridenTemplates {
			overriden[theme.Name][override] = true
		}
	}
	log.Printf("overriden: %+v\n", overriden)

	config := tmpl.CTemplateConfig{
		Minify:         Config.MinifyTemplates,
		Debug:          Dev.DebugMode,
		SuperDebug:     Dev.TemplateDebug,
		SkipHandles:    true,
		SkipTmplPtrMap: true,
		SkipInitBlock:  false,
		PackageName:    "tmpl",
		DockToID:       DockToID,
	}
	c := tmpl.NewCTemplateSet("js", "./logs/")
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
	dirPrefix := "./tmpl_client/"
	writeTemplateList(c, &wg, dirPrefix)
	return nil
}

func compileJSTemplates(wg *sync.WaitGroup, c *tmpl.CTemplateSet, themeName string) error {
	user, user2, user3 := tmplInitUsers()
	header, _, _ := tmplInitHeaders(user, user2, user3)
	now := time.Now()
	varList := make(map[string]tmpl.VarItem)

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
		"github.com/Azareal/Gosora/common": "c github.com/Azareal/Gosora/common",
	})
	// TODO: Fix the import loop so we don't have to use this hack anymore
	c.SetBuildTags("!no_templategen,tmplgentopic")

	t := TItemHold(make(map[string]TItem))

	topic := Topic{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, now, now, user3.ID, 1, 1, "", "::1", 1, 0, 1, 0, 1, "classname", 1, "", nil}
	topicsRow := TopicsRowMut{&TopicsRow{topic, 0, user2, "", 0, user3, "General", "/forum/general.2"}, false}
	t.AddStd("topics_topic", "c.TopicsRowMut", topicsRow)

	poll := Poll{ID: 1, Type: 0, Options: map[int]string{0: "Nothing", 1: "Something"}, Results: map[int]int{0: 5, 1: 2}, QuickOptions: []PollOption{
		{0, "Nothing"},
		{1, "Something"},
	}, VoteCount: 7}
	avatar, microAvatar := BuildAvatar(62, "")
	miniAttach := []*MiniAttachment{{Path: "/"}}
	tu := TopicUser{1, "blah", "Blah", "Hey there!", 62, false, false, now, now, 1, 1, 0, "", "::1", 1, 0, 1, 0, "classname", poll.ID, "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, avatar, microAvatar, 0, "", "", "", 58, false, miniAttach, nil, false}
	var replyList []*ReplyUser
	// TODO: Do we really want the UID here to be zero?
	avatar, microAvatar = BuildAvatar(0, "")
	reply := Reply{1, 1, "Yo!", 1 /*, Config.DefaultGroup*/, now, 0, 0, 1, "::1", true, 1, 1, ""}
	ru := &ReplyUser{ClassName: "", Reply: reply, CreatedByName: "Alice", Avatar: avatar, Group: Config.DefaultGroup, Level: 0, Attachments: miniAttach}
	_, err = ru.Init(user)
	if err != nil {
		return err
	}
	replyList = append(replyList, ru)

	varList = make(map[string]tmpl.VarItem)
	header.Title = "Topic Name"
	tpage := TopicPage{header, replyList, tu, &Forum{ID: 1, Name: "Hahaha"}, &poll, Paginator{[]int{1}, 1, 1}}
	tpage.Forum.Link = BuildForumURL(NameToSlug(tpage.Forum.Name), tpage.Forum.ID)
	t.AddStd("topic_posts", "c.TopicPage", tpage)
	t.AddStd("topic_alt_posts", "c.TopicPage", tpage)

	itemsPerPage := 25
	_, page, lastPage := PageOffset(20, 1, itemsPerPage)
	pageList := Paginate(page, lastPage, 5)
	t.AddStd("paginator", "c.Paginator", Paginator{pageList, page, lastPage})

	t.AddStd("topic_c_edit_post", "c.TopicCEditPost", TopicCEditPost{ID: 0, Source: "", Ref: ""})
	t.AddStd("topic_c_attach_item", "c.TopicCAttachItem", TopicCAttachItem{ID: 1, ImgSrc: "", Path: "", FullPath: ""})
	t.AddStd("topic_c_poll_input", "c.TopicCPollInput", TopicCPollInput{Index: 0})

	parti := []*User{user}
	convo := &Conversation{1, BuildConvoURL(1), user.ID, time.Now(), 0, time.Now()}
	convoItems := []ConvoViewRow{{&ConversationPost{1, 1, "hey", "", user.ID}, user, "", 4, true}}
	convoPage := ConvoViewPage{header, convo, convoItems, parti, true, Paginator{[]int{1}, 1, 1}}
	t.AddStd("convo", "c.ConvoViewPage", convoPage)

	t.AddStd("notice", "string", "nonono")

	dirPrefix := "./tmpl_client/"
	writeTemplate := func(name, content string) {
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
			e := writeFile(dirPrefix+"tmpl_"+name+tname+".jgo", content)
			if e != nil {
				log.Fatal(e)
			}
			wg.Done()
		}()
	}

	log.Print("Writing the templates")
	for name, titem := range t {
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

var poutlen = len("\n// nolint\nfunc init() {\n")
var poutlooplen = len("__frags[0]=arr_0[:]\n")

func getTemplateList(c *tmpl.CTemplateSet, wg *sync.WaitGroup, prefix string) string {
	DebugLog("in getTemplateList")
	//pout := "\n// nolint\nfunc init() {\n"
	tFragCount := make(map[string]int)
	bodyMap := make(map[string]string) //map[body]fragmentPrefix
	//tmplMap := make(map[string]map[string]string) // map[tmpl]map[body]fragmentPrefix
	tmpCount := 0
	var bsb strings.Builder
	var poutsb strings.Builder
	poutsb.Grow(poutlen + (poutlooplen * len(c.FragOut)))
	poutsb.WriteString("\n// nolint\nfunc init() {\n")
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
			//var bits string
			bsb.Reset()
			DebugLog("encoding f.Body")
			for _, char := range []byte(frag.Body) {
				if char == '\'' {
					//bits += "'\\" + string(char) + "',"
					bsb.WriteString("'\\'',")
				} else if char < 32 {
					//bits += strconv.Itoa(int(char)) + ","
					bsb.WriteString(strconv.Itoa(int(char)))
					bsb.WriteByte(',')
				} else {
					//bits += "'" + string(char) + "',"
					bsb.WriteByte('\'')
					bsb.WriteString(string(char))
					bsb.WriteString("',")
				}
			}
			tmpStr := strconv.Itoa(tmpCount)
			//"arr_" + tmpStr + ":=[...]byte{" + /*bits*/ bsb.String() + "}\n"
			poutsb.WriteString("arr_")
			poutsb.WriteString(tmpStr)
			poutsb.WriteString(":=[...]byte{")
			poutsb.WriteString(bsb.String())
			poutsb.WriteString("}\n")

			//front + "=arr_" + tmpStr + "[:]\n"
			poutsb.WriteString(front)
			poutsb.WriteString("=arr_")
			poutsb.WriteString(tmpStr)
			poutsb.WriteString("[:]\n")
			tmpCount++
			//pout += front + "=[]byte(`" + frag.Body + "`)\n"
		} else {
			DebugLog("encoding cached index " + fp)
			poutsb.WriteString(front + "=" + fp + "\n")
		}

		_, ok = tFragCount[frag.TmplName]
		if !ok {
			tFragCount[frag.TmplName] = 0
		}
		tFragCount[frag.TmplName]++
	}

	//out := "package " + c.GetConfig().PackageName + "\n\n"
	bsb.Reset()
	sb := bsb
	pkgName := c.GetConfig().PackageName
	sb.Grow(tllenhint + ((looplenhint + 2) + (looplenhint2+2)*len(tFragCount)) + len(pkgName))
	sb.WriteString("package ")
	sb.WriteString(pkgName)
	sb.WriteString("\n\n")
	for templateName, count := range tFragCount {
		//out += "var " + templateName + "_frags = make([][]byte," + strconv.Itoa(count) + ")\n"
		//out += "var " + templateName + "_frags [" + strconv.Itoa(count) + "][]byte\n"
		sb.WriteString("var ")
		sb.WriteString(templateName)
		sb.WriteString("_frags [")
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString("][]byte\n")
	}
	sb.WriteString(poutsb.String())
	sb.WriteString("\n\n// nolint\nGetFrag = func(name string) [][]byte {\nswitch(name) {\n")
	//getterstr := "\n// nolint\nGetFrag = func(name string) [][]byte {\nswitch(name) {\n"
	for templateName, _ := range tFragCount {
		//getterstr += "\tcase \"" + templateName + "\":\n"
		///getterstr += "\treturn " + templateName + "_frags\n"
		//getterstr += "\treturn " + templateName + "_frags[:]\n"
		sb.WriteString("\tcase \"")
		sb.WriteString(templateName)
		sb.WriteString("\":\n\treturn ")
		sb.WriteString(templateName)
		sb.WriteString("_frags[:]\n")
	}
	sb.WriteString("}\nreturn nil\n}\n}\n")
	//getterstr += "}\nreturn nil\n}\n"
	//out += pout + "\n" + getterstr + "}\n"

	return sb.String()
}

var looplenhint = len("var _frags [][]byte\n")
var looplenhint2 = len("\tcase \"\":\n\treturn _frags[:]\n")
var tllenhint = len("package \n\n\n// nolint\nGetFrag = func(name string) [][]byte {\nswitch(name) {\nvar _frags [][]byte\n\tcase \"\":\n\treturn _frags[:]\n}\nreturn nil\n}\n\n}\n")

func writeTemplateList(c *tmpl.CTemplateSet, wg *sync.WaitGroup, prefix string) {
	log.Print("Writing template list")
	wg.Add(1)
	go func() {
		e := writeFile(prefix+"tmpl_list.go", getTemplateList(c, wg, prefix))
		if e != nil {
			log.Fatal(e)
		}
		wg.Done()
	}()
	wg.Wait()
}

func arithToInt64(in interface{}) (o int64) {
	switch in := in.(type) {
	case int64:
		o = in
	case int32:
		o = int64(in)
	case int:
		o = int64(in)
	case uint32:
		o = int64(in)
	case uint16:
		o = int64(in)
	case uint8:
		o = int64(in)
	case uint:
		o = int64(in)
	}
	return o
}

func arithDuoToInt64(left, right interface{}) (leftInt, rightInt int64) {
	return arithToInt64(left), arithToInt64(right)
}

func initDefaultTmplFuncMap() {
	// TODO: Add support for floats
	fmap := make(map[string]interface{})
	fmap["add"] = func(left, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt + rightInt
	}

	fmap["subtract"] = func(left, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt - rightInt
	}

	fmap["multiply"] = func(left, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		return leftInt * rightInt
	}

	fmap["divide"] = func(left, right interface{}) interface{} {
		leftInt, rightInt := arithDuoToInt64(left, right)
		if leftInt == 0 || rightInt == 0 {
			return 0
		}
		return leftInt / rightInt
	}

	fmap["dock"] = func(dock, headerInt interface{}) interface{} {
		return template.HTML(BuildWidget(dock.(string), headerInt.(*Header)))
	}

	fmap["hasWidgets"] = func(dock, headerInt interface{}) interface{} {
		return HasWidgets(dock.(string), headerInt.(*Header))
	}

	fmap["elapsed"] = func(startedAtInt interface{}) interface{} {
		//return time.Since(startedAtInt.(time.Time)).String()
		return time.Duration(uutils.Nanotime() - startedAtInt.(int64)).String()
	}

	fmap["lang"] = func(phraseNameInt interface{}) interface{} {
		phraseName, ok := phraseNameInt.(string)
		if !ok {
			panic("phraseNameInt is not a string")
		}
		// TODO: Log non-existent phrases?
		return template.HTML(p.GetTmplPhrase(phraseName))
	}

	// TODO: Implement this in the template generator too
	fmap["langf"] = func(phraseNameInt interface{}, args ...interface{}) interface{} {
		phraseName, ok := phraseNameInt.(string)
		if !ok {
			panic("phraseNameInt is not a string")
		}
		// TODO: Log non-existent phrases?
		// TODO: Optimise TmplPhrasef so we don't use slow Sprintf there
		return template.HTML(p.GetTmplPhrasef(phraseName, args...))
	}

	fmap["level"] = func(levelInt interface{}) interface{} {
		level, ok := levelInt.(int)
		if !ok {
			panic("levelInt is not an integer")
		}
		return template.HTML(p.GetLevelPhrase(level))
	}

	fmap["bunit"] = func(byteInt interface{}) interface{} {
		var byteFloat float64
		var unit string
		switch bytes := byteInt.(type) {
		case int:
			byteFloat, unit = ConvertByteUnit(float64(bytes))
		case int64:
			byteFloat, unit = ConvertByteUnit(float64(bytes))
		case uint64:
			byteFloat, unit = ConvertByteUnit(float64(bytes))
		case float64:
			byteFloat, unit = ConvertByteUnit(bytes)
		default:
			panic("bytes is not an int, int64 or uint64")
		}
		return fmt.Sprintf("%.1f", byteFloat) + unit
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

	fmap["dyntmpl"] = func(nameInt, pageInt, headerInt interface{}) interface{} {
		header := headerInt.(*Header)
		err := header.Theme.RunTmpl(nameInt.(string), pageInt, header.Writer)
		if err != nil {
			return err
		}
		return ""
	}

	fmap["ptmpl"] = func(nameInt, pageInt, headerInt interface{}) interface{} {
		header := headerInt.(*Header)
		err := header.Theme.RunTmpl(nameInt.(string), pageInt, header.Writer)
		if err != nil {
			return err
		}
		return ""
	}

	fmap["js"] = func() interface{} {
		return false
	}

	fmap["flush"] = func() interface{} {
		return nil
	}

	fmap["res"] = func(nameInt interface{}) interface{} {
		n := nameInt.(string)
		if n[0] == '/' && n[1] == '/' {
		} else {
			if f, ok := StaticFiles.GetShort(n); ok {
				n = f.OName
			}
		}
		return n
	}

	DefaultTemplateFuncMap = fmap
}

func loadTemplates(t *template.Template, themeName string) error {
	t.Funcs(DefaultTemplateFuncMap)
	tFiles, err := filepath.Glob("templates/*.html")
	if err != nil {
		return err
	}

	tFileMap := make(map[string]int)
	for index, path := range tFiles {
		path = strings.Replace(path, "\\", "/", -1)
		log.Print("templateFile: ", path)
		if skipCTmpl(path) {
			log.Print("skipping")
			continue
		}
		tFileMap[path] = index
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
		index, ok := tFileMap["templates/"+strings.TrimPrefix(path, "templates/overrides/")]
		if !ok {
			log.Print("not ok: templates/" + strings.TrimPrefix(path, "templates/overrides/"))
			tFiles = append(tFiles, path)
			continue
		}
		tFiles[index] = path
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
			index, ok := tFileMap["templates/"+strings.TrimPrefix(path, "themes/"+themeName+"/overrides/")]
			if !ok {
				log.Print("not ok: templates/" + strings.TrimPrefix(path, "themes/"+themeName+"/overrides/"))
				tFiles = append(tFiles, path)
				continue
			}
			tFiles[index] = path
		}
	}

	// TODO: Minify these
	/*err = t.ParseFiles(tFiles...)
	if err != nil {
		return err
	}*/
	for _, fname := range tFiles {
		b, err := ioutil.ReadFile(fname)
		if err != nil {
			return err
		}
		s := tmpl.Minify(string(b))
		name := filepath.Base(fname)
		var tmpl *template.Template
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return err
		}
	}
	_, err = t.ParseGlob("pages/*")
	return err
}

func InitTemplates() error {
	DebugLog("Initialising the template system")
	initDefaultTmplFuncMap()

	// The interpreted templates...
	DebugLog("Loading the template files...")
	return loadTemplates(DefaultTemplates, "")
}
