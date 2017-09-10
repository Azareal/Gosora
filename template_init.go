package main

import "log"
import "html/template"
import "net/http"

var templates = template.New("")

// nolint
func interpreted_topic_template(pi TopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["topic"]
	if !ok {
		mapping = "topic"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// nolint
var template_topic_handle func(TopicPage, http.ResponseWriter) = interpreted_topic_template
var template_topic_alt_handle func(TopicPage, http.ResponseWriter) = interpreted_topic_template

// nolint
var template_topics_handle func(TopicsPage, http.ResponseWriter) = func(pi TopicsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["topics"]
	if !ok {
		mapping = "topics"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// nolint
var template_forum_handle func(ForumPage, http.ResponseWriter) = func(pi ForumPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["forum"]
	if !ok {
		mapping = "forum"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// nolint
var template_forums_handle func(ForumsPage, http.ResponseWriter) = func(pi ForumsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["forums"]
	if !ok {
		mapping = "forums"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// nolint
var template_profile_handle func(ProfilePage, http.ResponseWriter) = func(pi ProfilePage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["profile"]
	if !ok {
		mapping = "profile"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

// nolint
var template_create_topic_handle func(CreateTopicPage, http.ResponseWriter) = func(pi CreateTopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultThemeBox.Load().(string)].TemplatesMap["create-topic"]
	if !ok {
		mapping = "create-topic"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

func compileTemplates() error {
	var c CTemplateSet

	// Schemas to train the template compiler on what to expect
	// TODO: Add support for interface{}s
	user := User{62, buildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	// TODO: Do a more accurate level calculation for this?
	user2 := User{1, buildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 58, 1000, "127.0.0.1", 0}
	user3 := User{2, buildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 42, 900, "::1", 0}
	headerVars := &HeaderVars{
		Site:        site,
		Settings:    settingBox.Load().(SettingBox),
		Themes:      themes,
		ThemeName:   defaultThemeBox.Load().(string),
		NoticeList:  []string{"test"},
		Stylesheets: []string{"panel"},
		Scripts:     []string{"whatever"},
		Widgets: PageWidgets{
			LeftSidebar: template.HTML("lalala"),
		},
	}

	log.Print("Compiling the templates")

	topic := TopicUser{1, "blah", "Blah", "Hey there!", 0, false, false, "Date", "Date", 0, "", "127.0.0.1", 0, 1, "classname", "weird-data", buildProfileURL("fake-user", 62), "Fake User", config.DefaultGroup, "", 0, "", "", "", "", 58, false}
	var replyList []Reply
	replyList = append(replyList, Reply{0, 0, "Yo!", "Yo!", 0, "alice", "Alice", config.DefaultGroup, "", 0, 0, "", "", 0, "", "", "", "", 0, "127.0.0.1", false, 1, "", ""})

	var varList = make(map[string]VarItem)
	tpage := TopicPage{"Title", user, headerVars, replyList, topic, 1, 1}
	topicIDTmpl, err := c.compileTemplate("topic.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}
	topicIDAltTmpl, err := c.compileTemplate("topic_alt.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}

	varList = make(map[string]VarItem)
	ppage := ProfilePage{"User 526", user, headerVars, replyList, user}
	profileTmpl, err := c.compileTemplate("profile.html", "templates/", "ProfilePage", ppage, varList)
	if err != nil {
		return err
	}

	var forumList []Forum
	forums, err := fstore.GetAllVisible()
	if err != nil {
		return err
	}

	for _, forum := range forums {
		forumList = append(forumList, *forum)
	}
	varList = make(map[string]VarItem)
	forumsPage := ForumsPage{"Forum List", user, headerVars, forumList}
	forumsTmpl, err := c.compileTemplate("forums.html", "templates/", "ForumsPage", forumsPage, varList)
	if err != nil {
		return err
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList, &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, "Date", "Date", user3.ID, 1, "", "127.0.0.1", 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"})
	topicsPage := TopicsPage{"Topic List", user, headerVars, topicsList}
	topicsTmpl, err := c.compileTemplate("topics.html", "templates/", "TopicsPage", topicsPage, varList)
	if err != nil {
		return err
	}

	//var topicList []TopicUser
	//topicList = append(topicList,TopicUser{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","","admin-fred","Admin Fred",config.DefaultGroup,"",0,"","","","",58,false})
	forumItem := Forum{1, "general", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0, "", "", 0, "", 0, ""}
	forumPage := ForumPage{"General Forum", user, headerVars, topicsList, forumItem, 1, 1}
	forumTmpl, err := c.compileTemplate("forum.html", "templates/", "ForumPage", forumPage, varList)
	if err != nil {
		return err
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

func initTemplates() {
	if dev.DebugMode {
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
	if dev.DebugMode {
		log.Print("Loading the template files...")
	}
	templates.Funcs(fmap)
	template.Must(templates.ParseGlob("templates/*"))
	template.Must(templates.ParseGlob("pages/*"))
}
