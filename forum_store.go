/* Work in progress. Check back later! */
package main

import (
	"database/sql"
	"log"
	"sync"
	"sync/atomic"

	"./query_gen/lib"
)

var forumUpdateMutex sync.Mutex
var forumCreateMutex sync.Mutex
var forumPerms map[int]map[int]ForumPerms // [gid][fid]Perms
var fstore ForumStore

type ForumStore interface {
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
	Delete(id int)
	CascadeDelete(id int) error
	IncrementTopicCount(id int) error
	DecrementTopicCount(id int) error
	UpdateLastTopic(topicName string, tid int, username string, uid int, time string, fid int) error
	Exists(id int) bool
	GetAll() ([]*Forum, error)
	GetAllIDs() ([]int, error)
	GetAllVisible() ([]*Forum, error)
	GetAllVisibleIDs() ([]int, error)
	//GetChildren(parentID int, parentType string) ([]*Forum,error)
	//GetFirstChild(parentID int, parentType string) (*Forum,error)
	CreateForum(forumName string, forumDesc string, active bool, preset string) (int, error)

	GetGlobalCount() int
}

type MemoryForumStore struct {
	//forums    map[int]*Forum
	forums    sync.Map
	forumView atomic.Value // []*Forum
	//fids []int
	forumCount int

	get           *sql.Stmt
	getAll        *sql.Stmt
	delete        *sql.Stmt
	getForumCount *sql.Stmt
}

func NewMemoryForumStore() *MemoryForumStore {
	getStmt, err := qgen.Builder.SimpleSelect("forums", "name, desc, active, preset, parentID, parentType, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime", "fid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	getAllStmt, err := qgen.Builder.SimpleSelect("forums", "fid, name, desc, active, preset, parentID, parentType, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime", "", "fid ASC", "")
	if err != nil {
		log.Fatal(err)
	}
	deleteStmt, err := qgen.Builder.SimpleUpdate("forums", "name= '', active = 0", "fid = ?")
	if err != nil {
		log.Fatal(err)
	}
	forumCountStmt, err := qgen.Builder.SimpleCount("forums", "name != ''", "")
	if err != nil {
		log.Fatal(err)
	}
	return &MemoryForumStore{
		get:           getStmt,
		getAll:        getAllStmt,
		delete:        deleteStmt,
		getForumCount: forumCountStmt,
	}
}

func (mfs *MemoryForumStore) LoadForums() error {
	log.Print("Adding the uncategorised forum")
	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()

	var forumView []*Forum
	addForum := func(forum *Forum) {
		mfs.forums.Store(forum.ID, forum)
		if forum.Active && forum.Name != "" {
			forumView = append(forumView, forum)
		}
	}

	addForum(&Forum{0, buildForumURL(nameToSlug("Uncategorised"), 0), "Uncategorised", "", config.UncategorisedForumVisible, "all", 0, "", 0, "", "", 0, "", 0, ""})

	rows, err := get_forums_stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var i = 0
	for ; rows.Next(); i++ {
		forum := Forum{ID: 0, Active: true, Preset: "all"}
		err = rows.Scan(&forum.ID, &forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.ParentID, &forum.ParentType, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
		if err != nil {
			return err
		}

		if forum.Name == "" {
			if dev.DebugMode {
				log.Print("Adding a placeholder forum")
			}
		} else {
			log.Print("Adding the " + forum.Name + " forum")
		}

		forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
		forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
		addForum(&forum)
	}
	mfs.forumCount = i
	mfs.forumView.Store(forumView)
	return rows.Err()
}

// TODO: Hide social groups too
func (mfs *MemoryForumStore) rebuildView() {
	var forumView []*Forum
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		forum := value.(*Forum)
		if forum.Active && forum.Name != "" {
			forumView = append(forumView, forum)
		}
		return true
	})
	mfs.forumView.Store(forumView)
}

func (mfs *MemoryForumStore) DirtyGet(id int) *Forum {
	fint, ok := mfs.forums.Load(id)
	forum := fint.(*Forum)
	if !ok || forum.Name == "" {
		return &Forum{ID: -1, Name: ""}
	}
	return forum
}

func (mfs *MemoryForumStore) Get(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	forum := fint.(*Forum)
	if !ok || forum.Name == "" {
		return nil, ErrNoRows
	}
	return forum, nil
}

func (mfs *MemoryForumStore) CascadeGet(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	forum := fint.(*Forum)
	if !ok || forum.Name == "" {
		err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

		forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
		forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
		return forum, err
	}
	return forum, nil
}

func (mfs *MemoryForumStore) CascadeGetCopy(id int) (Forum, error) {
	fint, ok := mfs.forums.Load(id)
	forum := fint.(*Forum)
	if !ok || forum.Name == "" {
		err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

		forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
		forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
		return *forum, err
	}
	return *forum, nil
}

func (mfs *MemoryForumStore) BypassGet(id int) (*Forum, error) {
	var forum = Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

	forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
	forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
	return &forum, err
}

func (mfs *MemoryForumStore) Load(id int) error {
	var forum = Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
	if err != nil {
		return err
	}
	forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
	forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)

	mfs.Set(&forum)
	return nil
}

func (mfs *MemoryForumStore) Set(forum *Forum) error {
	if !mfs.Exists(forum.ID) {
		return ErrNoRows
	}
	mfs.forums.Store(forum.ID, forum)
	mfs.rebuildView()
	return nil
}

func (mfs *MemoryForumStore) GetAll() (forumView []*Forum, err error) {
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		forumView = append(forumView, value.(*Forum))
		return true
	})
	return forumView, nil
}

func (mfs *MemoryForumStore) GetAllIDs() (ids []int, err error) {
	mfs.forums.Range(func(_ interface{}, value interface{}) bool {
		ids = append(ids, value.(*Forum).ID)
		return true
	})
	return ids, nil
}

func (mfs *MemoryForumStore) GetAllVisible() ([]*Forum, error) {
	return mfs.forumView.Load().([]*Forum), nil
}

func (mfs *MemoryForumStore) GetAllVisibleIDs() ([]int, error) {
	forumView := mfs.forumView.Load().([]*Forum)
	var ids = make([]int, len(forumView))
	for i := 0; i < len(forumView); i++ {
		ids[i] = forumView[i].ID
	}
	return ids, nil
}

// TODO: Implement sub-forums.
/*func (mfs *MemoryForumStore) GetChildren(parentID int, parentType string) ([]*Forum,error) {
	return nil, nil
}
func (mfs *MemoryForumStore) GetFirstChild(parentID int, parentType string) (*Forum,error) {
	return nil, nil
}*/

func (mfs *MemoryForumStore) Exists(id int) bool {
	forum, ok := mfs.forums.Load(id)
	return ok && forum.(*Forum).Name != ""
}

// TODO: Batch deletions with name blanking? Is this necessary?
func (mfs *MemoryForumStore) Delete(id int) {
	mfs.forums.Delete(id)
	mfs.rebuildView()
}

func (mfs *MemoryForumStore) CascadeDelete(id int) error {
	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()
	_, err := mfs.delete.Exec(id)
	if err != nil {
		return err
	}
	mfs.Delete(id)
	return nil
}

func (mfs *MemoryForumStore) IncrementTopicCount(id int) error {
	forum, err := mfs.CascadeGet(id)
	if err != nil {
		return err
	}
	_, err = add_topics_to_forum_stmt.Exec(1, id)
	if err != nil {
		return err
	}
	forum.TopicCount++
	return nil
}

func (mfs *MemoryForumStore) DecrementTopicCount(id int) error {
	forum, err := mfs.CascadeGet(id)
	if err != nil {
		return err
	}
	_, err = remove_topics_from_forum_stmt.Exec(1, id)
	if err != nil {
		return err
	}
	forum.TopicCount--
	return nil
}

// TODO: Have a pointer to the last topic rather than storing it on the forum itself
func (mfs *MemoryForumStore) UpdateLastTopic(topicName string, tid int, username string, uid int, time string, fid int) error {
	forum, err := mfs.CascadeGet(fid)
	if err != nil {
		return err
	}

	_, err = update_forum_cache_stmt.Exec(topicName, tid, username, uid, fid)
	if err != nil {
		return err
	}

	forum.LastTopic = topicName
	forum.LastTopicID = tid
	forum.LastReplyer = username
	forum.LastReplyerID = uid
	forum.LastTopicTime = time

	return nil
}

func (mfs *MemoryForumStore) CreateForum(forumName string, forumDesc string, active bool, preset string) (int, error) {
	forumCreateMutex.Lock()
	res, err := create_forum_stmt.Exec(forumName, forumDesc, active, preset)
	if err != nil {
		return 0, err
	}

	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid := int(fid64)

	mfs.forums.Store(fid, &Forum{fid, buildForumURL(nameToSlug(forumName), fid), forumName, forumDesc, active, preset, 0, "", 0, "", "", 0, "", 0, ""})
	mfs.forumCount++

	// TODO: Add a GroupStore. How would it interact with the ForumStore?
	permmapToQuery(presetToPermmap(preset), fid)
	forumCreateMutex.Unlock()

	if active {
		mfs.rebuildView()
	}
	return fid, nil
}

// TODO: Get the total count of forums in the forum store minus the blanked forums rather than doing a heavy query for this?
// GetGlobalCount returns the total number of forums
func (mfs *MemoryForumStore) GetGlobalCount() (fcount int) {
	err := mfs.getForumCount.QueryRow().Scan(&fcount)
	if err != nil {
		LogError(err)
	}
	return fcount
}

// TODO: Work on SqlForumStore
