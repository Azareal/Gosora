/*
*
*	Gosora Forum Store
* 	Copyright Azareal 2017 - 2018
*
 */
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
var forumPerms map[int]map[int]ForumPerms // [gid][fid]Perms // TODO: Add an abstraction around this and make it more thread-safe
var fstore ForumStore

// ForumStore is an interface for accessing the forums and the metadata stored on them
type ForumStore interface {
	LoadForums() error
	DirtyGet(id int) *Forum
	Get(id int) (*Forum, error)
	GetCopy(id int) (Forum, error)
	BypassGet(id int) (*Forum, error)
	Reload(id int) error // ? - Should we move this to TopicCache? Might require us to do a lot more casting in Gosora though...
	//Update(Forum) error
	Delete(id int) error
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
	Create(forumName string, forumDesc string, active bool, preset string) (int, error)

	GetGlobalCount() int
}

type ForumCache interface {
	CacheGet(id int) (*Forum, error)
	CacheSet(forum *Forum) error
	CacheDelete(id int)
}

// MemoryForumStore is a struct which holds an arbitrary number of forums in memory, usually all of them, although we might introduce functionality to hold a smaller subset in memory for sites with an extremely large number of forums
type MemoryForumStore struct {
	forums    sync.Map     // map[int]*Forum
	forumView atomic.Value // []*Forum
	//fids []int
	forumCount int

	get           *sql.Stmt
	getAll        *sql.Stmt
	delete        *sql.Stmt
	getForumCount *sql.Stmt
}

// NewMemoryForumStore gives you a new instance of MemoryForumStore
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

// TODO: Add support for subforums
func (mfs *MemoryForumStore) LoadForums() error {
	log.Print("Adding the uncategorised forum")
	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()

	var forumView []*Forum
	addForum := func(forum *Forum) {
		mfs.forums.Store(forum.ID, forum)
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
			forumView = append(forumView, forum)
		}
	}

	addForum(&Forum{0, buildForumURL(nameToSlug("Uncategorised"), 0), "Uncategorised", "", config.UncategorisedForumVisible, "all", 0, "", 0, "", "", 0, "", 0, ""})

	rows, err := getForumsStmt.Query()
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
		// ? - ParentType blank means that it doesn't have a parent
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
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

func (mfs *MemoryForumStore) CacheGet(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		return nil, ErrNoRows
	}
	return fint.(*Forum), nil
}

func (mfs *MemoryForumStore) Get(id int) (*Forum, error) {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		var forum = &Forum{ID: id}
		err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

		forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
		forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
		return forum, err
	}
	return fint.(*Forum), nil
}

func (mfs *MemoryForumStore) GetCopy(id int) (Forum, error) {
	fint, ok := mfs.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		var forum = Forum{ID: id}
		err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

		forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
		forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
		return forum, err
	}
	return *fint.(*Forum), nil
}

func (mfs *MemoryForumStore) BypassGet(id int) (*Forum, error) {
	var forum = Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)

	forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
	forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)
	return &forum, err
}

func (mfs *MemoryForumStore) Reload(id int) error {
	var forum = Forum{ID: id}
	err := mfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
	if err != nil {
		return err
	}
	forum.Link = buildForumURL(nameToSlug(forum.Name), forum.ID)
	forum.LastTopicLink = buildTopicURL(nameToSlug(forum.LastTopic), forum.LastTopicID)

	mfs.CacheSet(&forum)
	return nil
}

func (mfs *MemoryForumStore) CacheSet(forum *Forum) error {
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

// TODO: Add a query for this rather than hitting cache
func (mfs *MemoryForumStore) Exists(id int) bool {
	forum, ok := mfs.forums.Load(id)
	return ok && forum.(*Forum).Name != ""
}

// TODO: Batch deletions with name blanking? Is this necessary?
func (mfs *MemoryForumStore) CacheDelete(id int) {
	mfs.forums.Delete(id)
	mfs.rebuildView()
}

func (mfs *MemoryForumStore) Delete(id int) error {
	forumUpdateMutex.Lock()
	defer forumUpdateMutex.Unlock()
	_, err := mfs.delete.Exec(id)
	if err != nil {
		return err
	}
	mfs.CacheDelete(id)
	return nil
}

func (mfs *MemoryForumStore) IncrementTopicCount(id int) error {
	forum, err := mfs.Get(id)
	if err != nil {
		return err
	}
	_, err = addTopicsToForumStmt.Exec(1, id)
	if err != nil {
		return err
	}
	forum.TopicCount++
	return nil
}

func (mfs *MemoryForumStore) DecrementTopicCount(id int) error {
	forum, err := mfs.Get(id)
	if err != nil {
		return err
	}
	_, err = removeTopicsFromForumStmt.Exec(1, id)
	if err != nil {
		return err
	}
	forum.TopicCount--
	return nil
}

// TODO: Have a pointer to the last topic rather than storing it on the forum itself
func (mfs *MemoryForumStore) UpdateLastTopic(topicName string, tid int, username string, uid int, time string, fid int) error {
	forum, err := mfs.Get(fid)
	if err != nil {
		return err
	}

	_, err = updateForumCacheStmt.Exec(topicName, tid, username, uid, fid)
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

func (mfs *MemoryForumStore) Create(forumName string, forumDesc string, active bool, preset string) (int, error) {
	forumCreateMutex.Lock()
	res, err := createForumStmt.Exec(forumName, forumDesc, active, preset)
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
