package main

import "log"
import "html/template"
import "net/http"

var templates = template.New("")

func interpreted_topic_template(pi TopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["topic"]
	if !ok {
		mapping = "topic"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

var template_topic_handle func(TopicPage, http.ResponseWriter) = interpreted_topic_template
var template_topic_alt_handle func(TopicPage, http.ResponseWriter) = interpreted_topic_template

var template_topics_handle func(TopicsPage, http.ResponseWriter) = func(pi TopicsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["topics"]
	if !ok {
		mapping = "topics"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

var template_forum_handle func(ForumPage, http.ResponseWriter) = func(pi ForumPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["forum"]
	if !ok {
		mapping = "forum"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

var template_forums_handle func(ForumsPage, http.ResponseWriter) = func(pi ForumsPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["forums"]
	if !ok {
		mapping = "forums"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

var template_profile_handle func(ProfilePage, http.ResponseWriter) = func(pi ProfilePage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["profile"]
	if !ok {
		mapping = "profile"
	}
	err := templates.ExecuteTemplate(w, mapping+".html", pi)
	if err != nil {
		InternalError(err, w)
	}
}

var template_create_topic_handle func(CreateTopicPage, http.ResponseWriter) = func(pi CreateTopicPage, w http.ResponseWriter) {
	mapping, ok := themes[defaultTheme].TemplatesMap["create-topic"]
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
	user := User{62, buildProfileURL("fake-user", 62), "Fake User", "compiler@localhost", 0, false, false, false, false, false, false, GuestPerms, make(map[string]bool), "", false, "", "", "", "", "", 0, 0, "0.0.0.0.0", 0}
	// TO-DO: Do a more accurate level calculation for this?
	user2 := User{1, buildProfileURL("admin-alice", 1), "Admin Alice", "alice@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 58, 1000, "127.0.0.1", 0}
	user3 := User{2, buildProfileURL("admin-fred", 62), "Admin Fred", "fred@localhost", 1, true, true, true, true, false, false, AllPerms, make(map[string]bool), "", true, "", "", "", "", "", 42, 900, "::1", 0}
	headerVars := &HeaderVars{
		Site:        site,
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

	var varList map[string]VarItem = make(map[string]VarItem)
	tpage := TopicPage{"Title", user, headerVars, replyList, topic, 1, 1}
	topic_id_tmpl, err := c.compileTemplate("topic.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}
	topic_id_alt_tmpl, err := c.compileTemplate("topic_alt.html", "templates/", "TopicPage", tpage, varList)
	if err != nil {
		return err
	}

	varList = make(map[string]VarItem)
	ppage := ProfilePage{"User 526", user, headerVars, replyList, user}
	profile_tmpl, err := c.compileTemplate("profile.html", "templates/", "ProfilePage", ppage, varList)
	if err != nil {
		return err
	}

	var forumList []Forum
	forums, err := fstore.GetAll()
	if err != nil {
		return err
	}

	for _, forum := range forums {
		if forum.Active {
			forumList = append(forumList, *forum)
		}
	}
	varList = make(map[string]VarItem)
	forums_page := ForumsPage{"Forum List", user, headerVars, forumList}
	forums_tmpl, err := c.compileTemplate("forums.html", "templates/", "ForumsPage", forums_page, varList)
	if err != nil {
		return err
	}

	var topicsList []*TopicsRow
	topicsList = append(topicsList, &TopicsRow{1, "topic-title", "Topic Title", "The topic content.", 1, false, false, "Date", "Date", user3.ID, 1, "", "127.0.0.1", 0, 1, "classname", "", &user2, "", 0, &user3, "General", "/forum/general.2"})
	topics_page := TopicsPage{"Topic List", user, headerVars, topicsList}
	topics_tmpl, err := c.compileTemplate("topics.html", "templates/", "TopicsPage", topics_page, varList)
	if err != nil {
		return err
	}

	//var topicList []TopicUser
	//topicList = append(topicList,TopicUser{1,"topic-title","Topic Title","The topic content.",1,false,false,"Date","Date",1,"","127.0.0.1",0,1,"classname","","admin-fred","Admin Fred",config.DefaultGroup,"",0,"","","","",58,false})
	forum_item := Forum{1, "general", "General Forum", "Where the general stuff happens", true, "all", 0, "", 0, "", "", 0, "", 0, ""}
	forum_page := ForumPage{"General Forum", user, headerVars, topicsList, forum_item, 1, 1}
	forum_tmpl, err := c.compileTemplate("forum.html", "templates/", "ForumPage", forum_page, varList)
	if err != nil {
		return err
	}

	log.Print("Writing the templates")
	go writeTemplate("topic", topic_id_tmpl)
	go writeTemplate("topic_alt", topic_id_alt_tmpl)
	go writeTemplate("profile", profile_tmpl)
	go writeTemplate("forums", forums_tmpl)
	go writeTemplate("topics", topics_tmpl)
	go writeTemplate("forum", forum_tmpl)
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

	// TO-DO: Add support for 64-bit integers
	// TO-DO: Add support for floats
	fmap := make(map[string]interface{})
	fmap["add"] = func(left interface{}, right interface{}) interface{} {
		var left_int int
		var right_int int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			left_int = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			right_int = right.(int)
		}
		return left_int + right_int
	}

	fmap["subtract"] = func(left interface{}, right interface{}) interface{} {
		var left_int int
		var right_int int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			left_int = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			right_int = right.(int)
		}
		return left_int - right_int
	}

	fmap["multiply"] = func(left interface{}, right interface{}) interface{} {
		var left_int int
		var right_int int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			left_int = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			right_int = right.(int)
		}
		return left_int * right_int
	}

	fmap["divide"] = func(left interface{}, right interface{}) interface{} {
		var left_int int
		var right_int int
		switch left := left.(type) {
		case uint, uint8, uint16, int, int32:
			left_int = left.(int)
		}
		switch right := right.(type) {
		case uint, uint8, uint16, int, int32:
			right_int = right.(int)
		}
		if left_int == 0 || right_int == 0 {
			return 0
		}
		return left_int / right_int
	}

	// The interpreted templates...
	if dev.DebugMode {
		log.Print("Loading the template files...")
	}
	templates.Funcs(fmap)
	template.Must(templates.ParseGlob("templates/*"))
	template.Must(templates.ParseGlob("pages/*"))
}
