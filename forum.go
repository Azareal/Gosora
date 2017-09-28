package main

//import "fmt"
import (
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

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
	ID         int
	Link       string
	Name       string
	Desc       string
	Active     bool
	Preset     string
	ParentID   int
	ParentType string
	TopicCount int

	LastTopic     *Topic
	LastTopicID   int
	LastReplyer   *User
	LastReplyerID int
	LastTopicTime string // So that we can re-calculate the relative time on the spot in /forums/

	//LastLock sync.RWMutex // ? - Is this safe to copy? Use a pointer to it? Should we do an fstore.Reload() instead?
}

// ? - What is this for?
type ForumSimple struct {
	ID     int
	Name   string
	Active bool
	Preset string
}

func (forum *Forum) Copy() (fcopy Forum) {
	//forum.LastLock.RLock()
	fcopy = *forum
	//forum.LastLock.RUnlock()
	return fcopy
}

/*func (forum *Forum) GetLast() (topic *Topic, user *User) {
	forum.LastLock.RLock()
	topic = forum.LastTopic
	if topic == nil {
		topic = &Topic{ID: 0}
	}

	user = forum.LastReplyer
	if user == nil {
		user = &User{ID: 0}
	}
	forum.LastLock.RUnlock()
	return topic, user
}

func (forum *Forum) SetLast(topic *Topic, user *User) {
	forum.LastLock.Lock()
	forum.LastTopic = topic
	forum.LastReplyer = user
	forum.LastLock.Unlock()
}*/

// TODO: Write tests for this
func (forum *Forum) Update(name string, desc string, active bool, preset string) error {
	if name == "" {
		name = forum.Name
	}
	preset = strings.TrimSpace(preset)
	_, err := updateForumStmt.Exec(name, desc, active, preset, forum.ID)
	if err != nil {
		return err
	}
	if forum.Preset != preset || preset == "custom" || preset == "" {
		err = permmapToQuery(presetToPermmap(preset), forum.ID)
		if err != nil {
			return err
		}
	}
	_ = fstore.Reload(forum.ID)
	return nil
}

// TODO: Replace this sorting mechanism with something a lot more efficient
// ? - Use sort.Slice instead?
type SortForum []*Forum

func (sf SortForum) Len() int {
	return len(sf)
}
func (sf SortForum) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}
func (sf SortForum) Less(i, j int) bool {
	return sf[i].ID < sf[j].ID
}

// ! Don't use this outside of tests and possibly template_init.go
func makeDummyForum(fid int, link string, name string, desc string, active bool, preset string, parentID int, parentType string, topicCount int) *Forum {
	return &Forum{ID: fid, Link: link, Name: name, Desc: desc, Active: active, Preset: preset, ParentID: parentID, ParentType: parentType, TopicCount: topicCount}
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
