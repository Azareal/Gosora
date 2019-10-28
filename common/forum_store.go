/*
*
*	Gosora Forum Store
* 	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"errors"
	"log"
	//"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
)

var forumCreateMutex sync.Mutex
var forumPerms map[int]map[int]*ForumPerms // [gid][fid]*ForumPerms // TODO: Add an abstraction around this and make it more thread-safe
var Forums ForumStore
var ErrBlankName = errors.New("The name must not be blank")
var ErrNoDeleteReports = errors.New("You cannot delete the Reports forum")

// ForumStore is an interface for accessing the forums and the metadata stored on them
type ForumStore interface {
	LoadForums() error
	DirtyGet(id int) *Forum
	Get(id int) (*Forum, error)
	BypassGet(id int) (*Forum, error)
	BulkGetCopy(ids []int) (forums []Forum, err error)
	Reload(id int) error // ? - Should we move this to ForumCache? It might require us to do some unnecessary casting though
	//Update(Forum) error
	Delete(id int) error
	AddTopic(tid int, uid int, fid int) error
	RemoveTopic(fid int) error
	UpdateLastTopic(tid int, uid int, fid int) error
	Exists(id int) bool
	GetAll() ([]*Forum, error)
	GetAllIDs() ([]int, error)
	GetAllVisible() ([]*Forum, error)
	GetAllVisibleIDs() ([]int, error)
	//GetChildren(parentID int, parentType string) ([]*Forum,error)
	//GetFirstChild(parentID int, parentType string) (*Forum,error)
	Create(forumName string, forumDesc string, active bool, preset string) (int, error)
	UpdateOrder(updateMap map[int]int) error

	Count() int
}

type ForumCache interface {
	CacheGet(id int) (*Forum, error)
	CacheSet(forum *Forum) error
	CacheDelete(id int)
	Length() int
}

// MemoryForumStore is a struct which holds an arbitrary number of forums in memory, usually all of them, although we might introduce functionality to hold a smaller subset in memory for sites with an extremely large number of forums
type MemoryForumStore struct {
	forums    sync.Map     // map[int]*Forum
	forumView atomic.Value // []*Forum

	get          *sql.Stmt
	getAll       *sql.Stmt
	delete       *sql.Stmt
	create       *sql.Stmt
	count        *sql.Stmt
	updateCache  *sql.Stmt
	addTopics    *sql.Stmt
	removeTopics *sql.Stmt
	updateOrder  *sql.Stmt
}

// NewMemoryForumStore gives you a new instance of MemoryForumStore
func NewMemoryForumStore() (*MemoryForumStore, error) {
	acc := qgen.NewAcc()
	f := "forums"
	// TODO: Do a proper delete
	return &MemoryForumStore{
		get:          acc.Select(f).Columns("name, desc, tmpl, active, order, preset, parentID, parentType, topicCount, lastTopicID, lastReplyerID").Where("fid = ?").Prepare(),
		getAll:       acc.Select(f).Columns("fid, name, desc, tmpl, active, order, preset, parentID, parentType, topicCount, lastTopicID, lastReplyerID").Orderby("order ASC, fid ASC").Prepare(),
		delete:       acc.Update(f).Set("name= '', active = 0").Where("fid = ?").Prepare(),
		create:       acc.Insert(f).Columns("name, desc, tmpl, active, preset").Fields("?,?,'',?,?").Prepare(),
		count:        acc.Count(f).Where("name != ''").Prepare(),
		updateCache:  acc.Update(f).Set("lastTopicID = ?, lastReplyerID = ?").Where("fid = ?").Prepare(),
		addTopics:    acc.Update(f).Set("topicCount = topicCount + ?").Where("fid = ?").Prepare(),
		removeTopics: acc.Update(f).Set("topicCount = topicCount - ?").Where("fid = ?").Prepare(),
		updateOrder:  acc.Update(f).Set("order = ?").Where("fid = ?").Prepare(),
	}, acc.FirstError()
}

// TODO: Rename to ReloadAll?
// TODO: Add support for subforums
func (s *MemoryForumStore) LoadForums() error {
	var forumView []*Forum
	addForum := func(forum *Forum) {
		s.forums.Store(forum.ID, forum)
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
			forumView = append(forumView, forum)
		}
	}

	rows, err := s.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	for ; rows.Next(); i++ {
		f := &Forum{ID: 0, Active: true, Preset: "all"}
		err = rows.Scan(&f.ID, &f.Name, &f.Desc, &f.Tmpl, &f.Active, &f.Order, &f.Preset, &f.ParentID, &f.ParentType, &f.TopicCount, &f.LastTopicID, &f.LastReplyerID)
		if err != nil {
			return err
		}

		if f.Name == "" {
			DebugLog("Adding a placeholder forum")
		} else {
			log.Printf("Adding the '%s' forum", f.Name)
		}

		f.Link = BuildForumURL(NameToSlug(f.Name), f.ID)
		f.LastTopic = Topics.DirtyGet(f.LastTopicID)
		f.LastReplyer = Users.DirtyGet(f.LastReplyerID)
		addForum(f)
	}
	s.forumView.Store(forumView)
	TopicListThaw.Thaw()
	return rows.Err()
}

// TODO: Hide social groups too
// ? - Will this be hit a lot by plugin_guilds?
func (s *MemoryForumStore) rebuildView() {
	var forumView []*Forum
	s.forums.Range(func(_ interface{}, value interface{}) bool {
		forum := value.(*Forum)
		// ? - ParentType blank means that it doesn't have a parent
		if forum.Active && forum.Name != "" && forum.ParentType == "" {
			forumView = append(forumView, forum)
		}
		return true
	})
	sort.Sort(SortForum(forumView))
	s.forumView.Store(forumView)
	TopicListThaw.Thaw()
}

func (s *MemoryForumStore) DirtyGet(id int) *Forum {
	fint, ok := s.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		return &Forum{ID: -1, Name: ""}
	}
	return fint.(*Forum)
}

func (s *MemoryForumStore) CacheGet(id int) (*Forum, error) {
	fint, ok := s.forums.Load(id)
	if !ok || fint.(*Forum).Name == "" {
		return nil, ErrNoRows
	}
	return fint.(*Forum), nil
}

func (s *MemoryForumStore) Get(id int) (*Forum, error) {
	fint, ok := s.forums.Load(id)
	if ok {
		forum := fint.(*Forum)
		if forum.Name == "" {
			return nil, ErrNoRows
		}
		return forum, nil
	}

	forum, err := s.BypassGet(id)
	if err != nil {
		return nil, err
	}
	s.CacheSet(forum)
	return forum, err
}

func (s *MemoryForumStore) BypassGet(id int) (*Forum, error) {
	var f = &Forum{ID: id}
	err := s.get.QueryRow(id).Scan(&f.Name, &f.Desc, &f.Tmpl,&f.Active, &f.Order, &f.Preset, &f.ParentID, &f.ParentType, &f.TopicCount, &f.LastTopicID, &f.LastReplyerID)
	if err != nil {
		return nil, err
	}
	if f.Name == "" {
		return nil, ErrNoRows
	}
	f.Link = BuildForumURL(NameToSlug(f.Name), f.ID)
	f.LastTopic = Topics.DirtyGet(f.LastTopicID)
	f.LastReplyer = Users.DirtyGet(f.LastReplyerID)
	//TopicListThaw.Thaw()

	return f, err
}

// TODO: Optimise this
func (s *MemoryForumStore) BulkGetCopy(ids []int) (forums []Forum, err error) {
	forums = make([]Forum, len(ids))
	for i, id := range ids {
		forum, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		forums[i] = forum.Copy()
	}
	return forums, nil
}

func (s *MemoryForumStore) Reload(id int) error {
	forum, err := s.BypassGet(id)
	if err != nil {
		return err
	}
	s.CacheSet(forum)
	return nil
}

func (s *MemoryForumStore) CacheSet(forum *Forum) error {
	s.forums.Store(forum.ID, forum)
	s.rebuildView()
	return nil
}

// ! Has a randomised order
func (s *MemoryForumStore) GetAll() (forumView []*Forum, err error) {
	s.forums.Range(func(_ interface{}, value interface{}) bool {
		forumView = append(forumView, value.(*Forum))
		return true
	})
	sort.Sort(SortForum(forumView))
	return forumView, nil
}

// ? - Can we optimise the sorting?
func (s *MemoryForumStore) GetAllIDs() (ids []int, err error) {
	s.forums.Range(func(_ interface{}, value interface{}) bool {
		ids = append(ids, value.(*Forum).ID)
		return true
	})
	sort.Ints(ids)
	return ids, nil
}

func (s *MemoryForumStore) GetAllVisible() (forumView []*Forum, err error) {
	forumView = s.forumView.Load().([]*Forum)
	return forumView, nil
}

func (s *MemoryForumStore) GetAllVisibleIDs() ([]int, error) {
	forumView := s.forumView.Load().([]*Forum)
	ids := make([]int, len(forumView))
	for i := 0; i < len(forumView); i++ {
		ids[i] = forumView[i].ID
	}
	return ids, nil
}

// TODO: Implement sub-forums.
/*func (s *MemoryForumStore) GetChildren(parentID int, parentType string) ([]*Forum,error) {
	return nil, nil
}
func (s *MemoryForumStore) GetFirstChild(parentID int, parentType string) (*Forum,error) {
	return nil, nil
}*/

// TODO: Add a query for this rather than hitting cache
func (s *MemoryForumStore) Exists(id int) bool {
	forum, ok := s.forums.Load(id)
	if !ok {
		return false
	}
	return forum.(*Forum).Name != ""
}

// TODO: Batch deletions with name blanking? Is this necessary?
func (s *MemoryForumStore) CacheDelete(id int) {
	s.forums.Delete(id)
	s.rebuildView()
}

// TODO: Add a hook to allow plugin_guilds to detect when one of it's forums has just been deleted?
func (s *MemoryForumStore) Delete(id int) error {
	if id == ReportForumID {
		return ErrNoDeleteReports
	}
	_, err := s.delete.Exec(id)
	s.CacheDelete(id)
	return err
}

func (s *MemoryForumStore) AddTopic(tid int, uid int, fid int) error {
	_, err := s.updateCache.Exec(tid, uid, fid)
	if err != nil {
		return err
	}
	_, err = s.addTopics.Exec(1, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	return s.Reload(fid)
}

// TODO: Update the forum cache with the latest topic
func (s *MemoryForumStore) RemoveTopic(fid int) error {
	_, err := s.removeTopics.Exec(1, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	s.Reload(fid)
	return nil
}

// DEPRECATED. forum.Update() will be the way to do this in the future, once it's completed
// TODO: Have a pointer to the last topic rather than storing it on the forum itself
func (s *MemoryForumStore) UpdateLastTopic(tid int, uid int, fid int) error {
	_, err := s.updateCache.Exec(tid, uid, fid)
	if err != nil {
		return err
	}
	// TODO: Bypass the database and update this with a lock or an unsafe atomic swap
	return s.Reload(fid)
}

func (s *MemoryForumStore) Create(name string, desc string, active bool, preset string) (int, error) {
	if name == "" {
		return 0, ErrBlankName
	}
	forumCreateMutex.Lock()
	defer forumCreateMutex.Unlock()

	res, err := s.create.Exec(name, desc, active, preset)
	if err != nil {
		return 0, err
	}

	fid64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	fid := int(fid64)

	err = s.Reload(fid)
	if err != nil {
		return 0, err
	}

	PermmapToQuery(PresetToPermmap(preset), fid)
	return fid, nil
}

// TODO: Make this atomic, maybe with a transaction?
func (s *MemoryForumStore) UpdateOrder(updateMap map[int]int) error {
	for fid, order := range updateMap {
		_, err := s.updateOrder.Exec(order, fid)
		if err != nil {
			return err
		}
	}
	return s.LoadForums()
}

// ! Might be slightly inaccurate, if the sync.Map is constantly shifting and churning, but it'll stabilise eventually. Also, slow. Don't use this on every request x.x
// Length returns the number of forums in the memory cache
func (s *MemoryForumStore) Length() (length int) {
	s.forums.Range(func(_ interface{}, value interface{}) bool {
		length++
		return true
	})
	return length
}

// TODO: Get the total count of forums in the forum store rather than doing a heavy query for this?
// Count returns the total number of forums
func (s *MemoryForumStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

// TODO: Work on SqlForumStore

// TODO: Work on the NullForumStore
