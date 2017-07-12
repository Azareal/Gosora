/* Work in progress. Check back later! */
package main

import "log"
import "sync"
//import "sync/atomic"
import "database/sql"
import "./query_gen/lib"

var forum_update_mutex sync.Mutex
var forum_create_mutex sync.Mutex
var forum_perms map[int]map[int]ForumPerms // [gid][fid]Perms
var fstore ForumStore

type ForumStore interface
{
	LoadForums() error
	DirtyGet(id int) *Forum
	Get(id int) (*Forum, error)
	CascadeGet(id int) (*Forum, error)
	CascadeGetCopy(id int) (Forum, error)
	BypassGet(id int) (*Forum, error)
	Load(id int) error
	Set(forum *Forum) error
	//Update(Forum) error
	//CascadeUpdate(Forum) error
	Delete(id int) error
	CascadeDelete(id int) error
	IncrementTopicCount(id int) error
	DecrementTopicCount(id int) error
	UpdateLastTopic(topic_name string, tid int, username string, uid int, time string, fid int) error
	Exists(id int) bool
	GetAll() ([]*Forum,error)
	GetAllIDs() ([]int,error)
	//GetChildren(parentID int, parentType string) ([]*Forum,error)
	//GetFirstChild(parentID int, parentType string) (*Forum,error)
	CreateForum(forum_name string, forum_desc string, active bool, preset string) (int, error)
}

type StaticForumStore struct
{
	forums []*Forum // The IDs for a forum tend to be low and sequential for the most part, so we can get more performance out of using a slice instead of a map AND it has better concurrency
	//fids []int
	forumCapCount int

	get *sql.Stmt
	get_all *sql.Stmt
}

func NewStaticForumStore() *StaticForumStore {
	get_stmt, err := qgen.Builder.SimpleSelect("forums","name, desc, active, preset, parentID, parentType, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime","fid = ?","","")
	if err != nil {
		log.Fatal(err)
	}
	get_all_stmt, err := qgen.Builder.SimpleSelect("forums","fid, name, desc, active, preset, parentID, parentType, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime","","fid ASC","")
	if err != nil {
		log.Fatal(err)
	}
	return &StaticForumStore{
		get: get_stmt,
		get_all: get_all_stmt,
	}
}

func (sfs *StaticForumStore) LoadForums() error {
	log.Print("Adding the uncategorised forum")
	var forums []*Forum = []*Forum{
		&Forum{0,"uncategorised","Uncategorised","",uncategorised_forum_visible,"all",0,"",0,"","",0,"",0,""},
	}

  rows, err := get_forums_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var i int = 1
	for ;rows.Next();i++ {
		forum := Forum{ID:0,Active:true,Preset:"all"}
		err = rows.Scan(&forum.ID, &forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			return err
		}

		// Ugh, you really shouldn't physically delete these items, it makes a big mess of things
		if forum.ID != i {
			log.Print("Stop physically deleting forums. You are messing up the IDs. Use the Forum Manager or delete_forum() instead x.x")
			sfs.fill_forum_id_gap(i, forum.ID)
		}

		if forum.Name == "" {
			if debug_mode {
				log.Print("Adding a placeholder forum")
			}
		} else {
			log.Print("Adding the " + forum.Name + " forum")
		}

		forum.Link = build_forum_url(name_to_slug(forum.Name),forum.ID)
		forum.LastTopicSlug = build_slug(name_to_slug(forum.LastTopic),forum.LastTopicID)
		forums = append(forums,&forum)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	sfs.forums = forums
	sfs.forumCapCount = i

	return nil
}

func (sfs *StaticForumStore) DirtyGet(id int) *Forum {
	if !((id <= sfs.forumCapCount) && (id >= 0) && sfs.forums[id].Name!="") {
		return &Forum{ID:-1,Name:""}
	}
	return sfs.forums[id]
}

func (sfs *StaticForumStore) Get(id int) (*Forum, error) {
	if !((id <= sfs.forumCapCount) && (id >= 0) && sfs.forums[id].Name!="") {
		return nil, ErrNoRows
	}
	return sfs.forums[id], nil
}

func (sfs *StaticForumStore) CascadeGet(id int) (*Forum, error) {
	if !((id <= sfs.forumCapCount) && (id >= 0) && sfs.forums[id].Name != "") {
		return nil, ErrNoRows
	}
	return sfs.forums[id], nil
}

func (sfs *StaticForumStore) CascadeGetCopy(id int) (forum Forum, err error) {
	if !((id <= sfs.forumCapCount) && (id >= 0) && sfs.forums[id].Name != "") {
		return forum, ErrNoRows
	}
	return *sfs.forums[id], nil
}

func (sfs *StaticForumStore) BypassGet(id int) (*Forum, error) {
	var forum Forum = Forum{ID:id}
	err := sfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (sfs *StaticForumStore) Load(id int) error {
	var forum Forum = Forum{ID:id}
	err := sfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
	if err != nil {
		return err
	}
	sfs.Set(&forum)
	return nil
}

// TO-DO: Set should be able to add new indices not just replace existing ones for consistency with UserStore and TopicStore
func (sfs *StaticForumStore) Set(forum *Forum) error {
	forum_update_mutex.Lock()
	if !sfs.Exists(forum.ID) {
		forum_update_mutex.Unlock()
		return ErrNoRows
	}
	sfs.forums[forum.ID] = forum
	forum_update_mutex.Unlock()
	return nil
}

func (sfs *StaticForumStore) GetAll() ([]*Forum,error) {
	return sfs.forums, nil
}

// TO-DO: Implement sub-forums.
/*func (sfs *StaticForumStore) GetChildren(parentID int, parentType string) ([]*Forum,error) {
	return nil, nil
}
func (sfs *StaticForumStore) GetFirstChild(parentID int, parentType string) (*Forum,error) {
	return nil, nil
}*/

// We can cheat slightly, as the StaticForumStore has all the IDs under the cap ;)
// Should we cache this? Well, it's only really used for superadmins right now.
func (sfs *StaticForumStore) GetAllIDs() ([]int,error) {
	var max int = sfs.forumCapCount
	var ids []int = make([]int,max)
	for i := 0; i < max; i++ {
		ids[i] = i
	}
	return ids, nil
}

func (sfs *StaticForumStore) Exists(id int) bool {
	return (id <= sfs.forumCapCount) && (id >= 0) && sfs.forums[id].Name != ""
}

func (sfs *StaticForumStore) Delete(id int) error {
	forum_update_mutex.Lock()
	if !sfs.Exists(id) {
		forum_update_mutex.Unlock()
		return nil
	}
	sfs.forums[id].Name = ""
	forum_update_mutex.Unlock()
	return nil
}

func (sfs *StaticForumStore) CascadeDelete(id int) error {
	forum, err := sfs.CascadeGet(id)
	if err != nil {
		return err
	}

	forum_update_mutex.Lock()
	_, err = delete_forum_stmt.Exec(id)
	if err != nil {
		return err
	}
	forum.Name = ""
	forum_update_mutex.Unlock()
	return nil
}

func (sfs *StaticForumStore) IncrementTopicCount(id int) error {
	forum, err := sfs.CascadeGet(id)
	if err != nil {
		return err
	}
	_, err = add_topics_to_forum_stmt.Exec(1,id)
	if err != nil {
		return err
	}
	forum.TopicCount += 1
	return nil
}

func (sfs *StaticForumStore) DecrementTopicCount(id int) error {
	forum, err := sfs.CascadeGet(id)
	if err != nil {
		return err
	}
	_, err = remove_topics_from_forum_stmt.Exec(1, id)
	if err != nil {
		return err
	}
	forum.TopicCount -= 1
	return nil
}

// TO-DO: Have a pointer to the last topic rather than storing it on the forum itself
func (sfs *StaticForumStore) UpdateLastTopic(topic_name string, tid int, username string, uid int, time string, fid int) error {
	forum, err := sfs.CascadeGet(fid)
	if err != nil {
		return err
	}

	_, err = update_forum_cache_stmt.Exec(topic_name,tid,username,uid,fid)
	if err != nil {
		return err
	}

	forum.LastTopic = topic_name
	forum.LastTopicID = tid
	forum.LastReplyer = username
	forum.LastReplyerID = uid
	forum.LastTopicTime = time

	return nil
}

func (sfs *StaticForumStore) CreateForum(forum_name string, forum_desc string, active bool, preset string) (int, error) {
	var fid int
	err := forum_entry_exists_stmt.QueryRow().Scan(&fid)
	if err != nil && err != ErrNoRows {
		return 0, err
	}
	if err != ErrNoRows {
		forum_update_mutex.Lock()
		_, err = update_forum_stmt.Exec(forum_name, forum_desc, active, preset, fid)
		if err != nil {
			return fid, err
		}
		forum, err := sfs.Get(fid)
		if err != nil {
			return 0, ErrCacheDesync
		}
		forum.Name = forum_name
		forum.Desc = forum_desc
		forum.Active = active
		forum.Preset = preset
		forum_update_mutex.Unlock()
		return fid, nil
	}

	forum_create_mutex.Lock()
	res, err := create_forum_stmt.Exec(forum_name, forum_desc, active, preset)
	if err != nil {
		return 0, err
	}

	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid = int(fid64)

	sfs.forums = append(sfs.forums, &Forum{fid,name_to_slug(forum_name),forum_name,forum_desc,active,preset,0,"",0,"","",0,"",0,""})
	sfs.forumCapCount++

	// TO-DO: Add a GroupStore. How would it interact with the ForumStore?
	permmap_to_query(preset_to_permmap(preset),fid)
	forum_create_mutex.Unlock()
	return fid, nil
}

func (sfs *StaticForumStore) fill_forum_id_gap(biggerID int, smallerID int) {
	dummy := Forum{ID:0,Name:"",Active:false,Preset:"all"}
	for i := smallerID; i > biggerID; i++ {
		sfs.forums = append(sfs.forums, &dummy)
	}
}

// TO-DO: Work on MapForumStore

// TO-DO: Work on SqlForumStore
