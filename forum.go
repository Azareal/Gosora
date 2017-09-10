package main

//import "fmt"
import "strconv"
import _ "github.com/go-sql-driver/mysql"

type ForumAdmin struct {
	ID         int
	Name       string
	Desc       string
	Active     bool
	Preset     string
	TopicCount int
	PresetLang string
}

type Forum struct {
	ID            int
	Link          string
	Name          string
	Desc          string
	Active        bool
	Preset        string
	ParentID      int
	ParentType    string
	TopicCount    int
	LastTopicLink string
	LastTopic     string
	LastTopicID   int
	LastReplyer   string
	LastReplyerID int
	LastTopicTime string
}

type ForumSimple struct {
	ID     int
	Name   string
	Active bool
	Preset string
}

func buildForumURL(slug string, fid int) string {
	if slug == "" {
		return "/forum/" + strconv.Itoa(fid)
	}
	return "/forum/" + slug + "." + strconv.Itoa(fid)
}

func getForumURLPrefix() string {
	return "/forum/"
}
