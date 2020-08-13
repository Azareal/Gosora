package common

import (
	//"log"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	qgen "github.com/Azareal/Gosora/query_gen"
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
	Tmpl       string
	Active     bool
	Order      int
	Preset     string
	ParentID   int
	ParentType string
	TopicCount int

	LastTopic     *Topic
	LastTopicID   int
	LastReplyer   *User
	LastReplyerID int
	LastTopicTime string // So that we can re-calculate the relative time on the spot in /forums/
	LastPage int
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
			update:    acc.Update("forums").Set("name=?,desc=?,active=?,preset=?").Where("fid=?").Prepare(),
			setPreset: acc.Update("forums").Set("preset=?").Where("fid=?").Prepare(),
		}
		return acc.FirstError()
	})
}

// Copy gives you a non-pointer concurrency safe copy of the forum
func (f *Forum) Copy() (fcopy Forum) {
	fcopy = *f
	return fcopy
}

// TODO: Write tests for this
func (f *Forum) Update(name, desc string, active bool, preset string) error {
	if name == "" {
		name = f.Name
	}
	// TODO: Do a line sanitise? Does it matter?
	preset = strings.TrimSpace(preset)
	_, err := forumStmts.update.Exec(name, desc, active, preset, f.ID)
	if err != nil {
		return err
	}
	if f.Preset != preset && preset != "custom" && preset != "" {
		err = PermmapToQuery(PresetToPermmap(preset), f.ID)
		if err != nil {
			return err
		}
	}
	_ = Forums.Reload(f.ID)
	return nil
}

func (f *Forum) SetPreset(preset string, gid int) error {
	fp, changed := GroupForumPresetToForumPerms(preset)
	if changed {
		return f.SetPerms(fp, preset, gid)
	}
	return nil
}

// TODO: Refactor this
func (f *Forum) SetPerms(fperms *ForumPerms, preset string, gid int) (err error) {
	err = ReplaceForumPermsForGroup(gid, map[int]string{f.ID: preset}, map[int]*ForumPerms{f.ID: fperms})
	if err != nil {
		LogError(err)
		return errors.New("Unable to update the permissions")
	}

	// TODO: Add this and replaceForumPermsForGroup into a transaction?
	_, err = forumStmts.setPreset.Exec("", f.ID)
	if err != nil {
		LogError(err)
		return errors.New("Unable to update the forum")
	}
	err = Forums.Reload(f.ID)
	if err != nil {
		return errors.New("Unable to reload forum")
	}
	err = FPStore.Reload(f.ID)
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

/*func (sf SortForum) Less(i,j int) bool {
	l := sf.less(i,j)
	if l {
		log.Printf("%s is less than %s. order: %d. id: %d.",sf[i].Name, sf[j].Name, sf[i].Order, sf[i].ID)
	} else {
		log.Printf("%s is not less than %s. order: %d. id: %d.",sf[i].Name, sf[j].Name, sf[i].Order, sf[i].ID)
	}
	return l
}*/
func (sf SortForum) Less(i, j int) bool {
	if sf[i].Order < sf[j].Order {
		return true
	} else if sf[i].Order == sf[j].Order {
		return sf[i].ID < sf[j].ID
	}
	return false
}

// ! Don't use this outside of tests and possibly template_init.go
func BlankForum(fid int, link, name, desc string, active bool, preset string, parentID int, parentType string, topicCount int) *Forum {
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
