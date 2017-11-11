package common

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"./templates"
)

var Templates = template.New("")
var PrebuildTmplList []func(User, *HeaderVars) CTmpl

type CTmpl struct {
	Name       string
	Filename   string
	Path       string
	StructName string
	Data       interface{}
}

// nolint
func interpreted_topic_template(pi TopicPage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["topic"]
	if !ok {
		mapping = "topic"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var template_topic_handle func(TopicPage, http.ResponseWriter) error = interpreted_topic_template
var template_topic_alt_handle func(TopicPage, http.ResponseWriter) error = interpreted_topic_template

// nolint
var template_topics_handle func(TopicsPage, http.ResponseWriter) error = func(pi TopicsPage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["topics"]
	if !ok {
		mapping = "topics"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var template_forum_handle func(ForumPage, http.ResponseWriter) error = func(pi ForumPage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["forum"]
	if !ok {
		mapping = "forum"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var template_forums_handle func(ForumsPage, http.ResponseWriter) error = func(pi ForumsPage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["forums"]
	if !ok {
		mapping = "forums"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var template_profile_handle func(ProfilePage, http.ResponseWriter) error = func(pi ProfilePage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["profile"]
	if !ok {
		mapping = "profile"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// nolint
var template_create_topic_handle func(CreateTopicPage, http.ResponseWriter) error = func(pi CreateTopicPage, w http.ResponseWriter) error {
	mapping, ok := Themes[DefaultThemeBox.Load().(string)].TemplatesMap["create-topic"]
	if !ok {
		mapping = "create-topic"
	}
	return Templates.ExecuteTemplate(w, mapping+".html", pi)
}

// ? - Add template hooks?
func compileTemplates() error {
	var c tmpl.CTemplateSet
	c.Minify(Config.MinifyTemplates)
	c.SuperDebug(Dev.TemplateDebug)

	// Schemas to train the template compiler on what to expect
	// TODO: Add support for interface{}s
	user := User{62, BuildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	// TODO: Do a more accurate level calculation for this?
	user2 := User{1, BuildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 58, 1000, "127.0.0.1", 0}
	user3 := User{2, BuildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 42, 900, "::1", 0}
	headerVars := &HeaderVars{
		Site:        Site,
		Settings:    SettingBox.Load().(SettingMap),
		Themes:      Themes,
		ThemeName:   DefaultThemeBox.Load().(string),
		NoticeList:  []string{"test"},
		Stylesheets: []string{"panel"},
		Scripts:     []string{"whatever"},
		Widgets: PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	log.Print("Compiling the templates")

	var now = time.Now()
	topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, now, RelativeTime(now), now, RelativeTime(now), 0, "", "127.0.0.1", 0, 1, "classname", "weird-data", BuildProfileURL("fake-user", 62), "Fake User", Config.DefaultGroup, "", 0, "", "", "", "", "", 58, false}
	var replyList []ReplyUser
	replyList = append(replyList, ReplyUser{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", Config.DefaultGroup, now, RelativeTime(now), 0, 0, "", "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, "", ""})

	var varList = make(map[string]tmpl.VarItem)
	tpage := TopicPage{"Title", user, headerVars, replyList, topic, 1, 1}
	topicIDTmpl, err := c.Compile("topic.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}
	topicIDAltTmpl, err := c.Compile("topic_alt.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}

	varList = make(map[string]tmpl.VarItem)
	ppage := ProfilePage{"User 526", user, headerVars, replyList, user}
	profileTmpl, err := c.Compile("profile.html", "templates/", "ProfilePage", ppage, varList)
	if err != nil {
		return err
	}

	// TODO: Use a dummy forum list to avoid o(n) problems
	var forumList []Forum
	forums, err := Fstore.GetAll()
	if err != nil {
		return err
	}

	for _, forum := range forums {
		//log.Printf("*forum %+v\n", *forum)
		forumList = append(forumList, *forum)
	}
	varList = make(map[string]tmpl.VarItem)
	forumsPage := ForumsPage{"Forum List", user, headerVars, forumList}
	forumsTmpl, err := c.Compile("forums.html", "templates/", "ForumsPage", forumsPage, varList)
	if err != nil {
		return err
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList, &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, "Date", time.Now(), "Date", user3.ID, 1, "", "127.0.0.1", 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"})
	topicsPage := TopicsPage{"Topic List", user, headerVars, topicsList, forumList, Config.DefaultForum}
	topicsTmpl, err := c.Compile("topics.html", "templates/", "TopicsPage", topicsPage, varList)
	if err != nil {
		return err
	}

	//var topicList []TopicUser
	//topicList = append(topicList,TopicUser{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","","admin-fred","Admin Fred",config.DefaultGroup,"",0,"","","","",58,false})
	forumItem := makeDummyForum(1, "general-forum.1", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0)
	forumPage := ForumPage{"General Forum", user, headerVars, topicsList, forumItem, 1, 1}
	forumTmpl, err := c.Compile("forum.html", "templates/", "ForumPage", forumPage, varList)
	if err != nil {
		return err
	}

	// Let plugins register their own templates
	for _, tmplfunc := range PrebuildTmplList {
		tmplItem := tmplfunc(user, headerVars)
		varList = make(map[string]tmpl.VarItem)
		compiledTmpl, err := c.Compile(tmplItem.Filename, tmplItem.Path, tmplItem.StructName, tmplItem.Data, varList)
		if err != nil {
			return err
		}
		go writeTemplate(tmplItem.Name, compiledTmpl)
	}

	log.Print("Writing the templates")
	go writeTemplate("topic", topicIDTmpl)
	go writeTemplate("topic_alt", topicIDAltTmpl)
	go writeTemplate("profile", profileTmpl)
	go writeTemplate("forums", forumsTmpl)
	go writeTemplate("topics", topicsTmpl)
	go writeTemplate("forum", forumTmpl)
	go func() {
		err := writeFile("./template_list.go", "package main\n\n// nolint\n"+c.FragOut)
		if err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func writeTemplate(name string, content string) {
	err := writeFile("./template_"+name+".go", content)
	if err != nil {
		log.Fatal(err)
	}
}

func InitTemplates() {
	if Dev.DebugMode {
		log.Print("Initialising the template system")
	}
	compileTemplates()

	// TODO: Add support for 64-bit integers
	// TODO: Add support for floats
	fmap := make(map[string]interface{})
	fmap["add"] = func(left interface{}, right interface{}) interface{} {
		var leftInt, rightInt int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			leftInt = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			rightInt = right.(int)
		}
		return leftInt + rightInt
	}

	fmap["subtract"] = func(left interface{}, right interface{}) interface{} {
		var leftInt, rightInt int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			leftInt = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			rightInt = right.(int)
		}
		return leftInt - rightInt
	}

	fmap["multiply"] = func(left interface{}, right interface{}) interface{} {
		var leftInt, rightInt int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			leftInt = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			rightInt = right.(int)
		}
		return leftInt * rightInt
	}

	fmap["divide"] = func(left interface{}, right interface{}) interface{} {
		var leftInt, rightInt int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			leftInt = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			rightInt = right.(int)
		}
		if leftInt == 0 || rightInt == 0 {
			return 0
		}
		return leftInt / rightInt
	}

	// The interpreted templates...
	if Dev.DebugMode {
		log.Print("Loading the template files...")
	}
	Templates.Funcs(fmap)
	template.Must(Templates.ParseGlob("templates/*"))
	template.Must(Templates.ParseGlob("pages/*"))
}
