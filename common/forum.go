package common

//import "fmt"
import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"../query_gen/lib"
	_ "github.com/go-sql-driver/mysql"
)

// TODO: Do we really need this?
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
}

// ? - What is this for?
type ForumSimple struct {
	ID     int
	Name   string
	Active bool
	Preset string
}

type ForumStmts struct {
	update    *sql.Stmt
	setPreset *sql.Stmt
}

var forumStmts ForumStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		forumStmts = ForumStmts{
			update:    acc.Update("forums").Set("name = ?, desc = ?, active = ?, preset = ?").Where("fid = ?").Prepare(),
			setPreset: acc.Update("forums").Set("preset = ?").Where("fid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// Copy gives you a non-pointer concurrency safe copy of the forum
func (forum *Forum) Copy() (fcopy Forum) {
	fcopy = *forum
	return fcopy
}

// TODO: Write tests for this
func (forum *Forum) Update(name string, desc string, active bool, preset string) error {
	if name == "" {
		name = forum.Name
	}
	// TODO: Do a line sanitise? Does it matter?
	preset = strings.TrimSpace(preset)
	_, err := forumStmts.update.Exec(name, desc, active, preset, forum.ID)
	if err != nil {
		return err
	}
	if forum.Preset != preset || preset == "custom" || preset == "" {
		err = PermmapToQuery(PresetToPermmap(preset), forum.ID)
		if err != nil {
			return err
		}
	}
	_ = Forums.Reload(forum.ID)
	return nil
}

func (forum *Forum) SetPreset(preset string, gid int) error {
	fperms, changed := GroupForumPresetToForumPerms(preset)
	if changed {
		return forum.SetPerms(fperms, preset, gid)
	}
	return nil
}

// TODO: Refactor this
func (forum *Forum) SetPerms(fperms *ForumPerms, preset string, gid int) (err error) {
	err = ReplaceForumPermsForGroup(gid, map[int]string{forum.ID: preset}, map[int]*ForumPerms{forum.ID: fperms})
	if err != nil {
		LogError(err)
		return errors.New("Unable to update the permissions")
	}

	// TODO: Add this and replaceForumPermsForGroup into a transaction?
	_, err = forumStmts.setPreset.Exec("", forum.ID)
	if err != nil {
		LogError(err)
		return errors.New("Unable to update the forum")
	}
	err = Forums.Reload(forum.ID)
	if err != nil {
		return errors.New("Unable to reload forum")
	}
	err = FPStore.Reload(forum.ID)
	if err != nil {
		return errors.New("Unable to reload the forum permissions")
	}
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
func BlankForum(fid int, link string, name string, desc string, active bool, preset string, parentID int, parentType string, topicCount int) *Forum {
	return &Forum{ID: fid, Link: link, Name: name, Desc: desc, Active: active, Preset: preset, ParentID: parentID, ParentType: parentType, TopicCount: topicCount}
}

func BuildForumURL(slug string, fid int) string {
	if slug == "" || !Config.BuildSlugs {
		return "/forum/" + strconv.Itoa(fid)
	}
	return "/forum/" + slug + "." + strconv.Itoa(fid)
}

func GetForumURLPrefix() string {
	return "/forum/"
}
