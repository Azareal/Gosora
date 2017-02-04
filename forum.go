package main
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type ForumAdmin struct
{
	ID int
	Name string
	Active bool
	Preset string
	TopicCount int
	
	PresetLang string
	PresetEmoji string
}

type Forum struct
{
	ID int
	Name string
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

func create_forum(forum_name string, active bool, preset string) (int, error) {
	var fid int
	err := forum_entry_exists_stmt.QueryRow().Scan(&fid)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err != sql.ErrNoRows {
		_, err = update_forum_stmt.Exec(forum_name, active, preset, fid)
		if err != nil {
			return fid, err
		}
		forums[fid].Name = forum_name
		forums[fid].Active = active
		forums[fid].Preset = preset
		return fid, nil
	}
	
	res, err := create_forum_stmt.Exec(forum_name, active, preset)
	if err != nil {
		return 0, err
	}
	
	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid = int(fid64)
	
	forums = append(forums, Forum{fid,forum_name,active,preset,0,"",0,"",0,""})
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
