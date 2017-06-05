package main

//import "fmt"
import "sync"
import "strconv"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type ForumAdmin struct
{
	ID int
	Name string
	Desc string
	Active bool
	Preset string
	TopicCount int
	PresetLang string
}

type Forum struct
{
	ID int
	Name string
	Desc string
	Active bool
	Preset string
	TopicCount int
	LastTopic string
	LastTopicID int
	LastReplyer string
	LastReplyerID int
	LastTopicTime string
}

type ForumSimple struct
{
	ID int
	Name string
	Active bool
	Preset string
}

var forum_update_mutex sync.Mutex
func create_forum(forum_name string, forum_desc string, active bool, preset string) (int, error) {
	var fid int
	err := forum_entry_exists_stmt.QueryRow().Scan(&fid)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err != sql.ErrNoRows {
		forum_update_mutex.Lock()
		_, err = update_forum_stmt.Exec(forum_name, forum_desc, active, preset, fid)
		if err != nil {
			return fid, err
		}
		forums[fid].Name = forum_name
		forums[fid].Desc = forum_desc
		forums[fid].Active = active
		forums[fid].Preset = preset
		forum_update_mutex.Unlock()
		return fid, nil
	}

	res, err := create_forum_stmt.Exec(forum_name, forum_desc, active, preset)
	if err != nil {
		return 0, err
	}

	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid = int(fid64)

	forums = append(forums, Forum{fid,forum_name,forum_desc,active,preset,0,"",0,"",0,""})
	return fid, nil
}

func delete_forum(fid int) error {
	_, err := delete_forum_stmt.Exec(fid)
	if err != nil {
		return err
	}
	forums[fid].Name = ""
	return nil
}

func get_forum(fid int) (forum *Forum, res bool) {
	if !((fid <= forumCapCount) && (fid >= 0) && forums[fid].Name!="") {
		return forum, false
	}
	return &forums[fid], true
}

func get_forum_copy(fid int) (forum Forum, res bool) {
	if !((fid <= forumCapCount) && (fid >= 0) && forums[fid].Name != "") {
		return forum, false
	}
	return forums[fid], true
}

func forum_exists(fid int) bool {
	return (fid <= forumCapCount) && (fid >= 0) && forums[fid].Name != ""
}

func build_forum_url(fid int) string {
	return "/forum/" + strconv.Itoa(fid)
}
