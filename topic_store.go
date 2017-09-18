/*
*
*	Gosora Topic Store
*	Copyright Azareal 2017 - 2018
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

// TODO: Add the watchdog goroutine
// TODO: Add BulkGetMap
// TODO: Add some sort of update method
// ? - Should we add stick, lock, unstick, and unlock methods? These might be better on the Topics not the TopicStore
var topics TopicStore

type TopicStore interface {
	Reload(id int) error // ? - Should we move this to TopicCache? Might require us to do a lot more casting in Gosora though...
	Get(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	Delete(id int) error
	Exists(id int) bool
	AddLastTopic(item *Topic, fid int) error // unimplemented
	// TODO: Implement these two methods
	//GetReplies() ([]*Reply, error)
	//GetRepliesRange(lower int, higher int) ([]*Reply, error)
	GetGlobalCount() int
}

type TopicCache interface {
	CacheGet(id int) (*Topic, error)
	CacheGetUnsafe(id int) (*Topic, error)
	CacheSet(item *Topic) error
	CacheAdd(item *Topic) error
	CacheAddUnsafe(item *Topic) error
	CacheRemove(id int) error
	CacheRemoveUnsafe(id int) error
	Flush()
	GetLength() int
	SetCapacity(capacity int)
	GetCapacity() int
}

type MemoryTopicStore struct {
	items      map[int]*Topic
	length     int64 // sync/atomic only lets us operate on int32s and int64s
	capacity   int
	get        *sql.Stmt
	exists     *sql.Stmt
	topicCount *sql.Stmt
	delete     *sql.Stmt
	sync.RWMutex
}

// NewMemoryTopicStore gives you a new instance of MemoryTopicStore
func NewMemoryTopicStore(capacity int) *MemoryTopicStore {
	getStmt, err := qgen.Builder.SimpleSelect("topics", "title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data", "tid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	existsStmt, err := qgen.Builder.SimpleSelect("topics", "tid", "tid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	topicCountStmt, err := qgen.Builder.SimpleCount("topics", "", "")
	if err != nil {
		log.Fatal(err)
	}
	deleteStmt, err := qgen.Builder.SimpleDelete("topics", "tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	return &MemoryTopicStore{
		items:      make(map[int]*Topic),
		capacity:   capacity,
		get:        getStmt,
		exists:     existsStmt,
		topicCount: topicCountStmt,
		delete:     deleteStmt,
	}
}

func (mts *MemoryTopicStore) CacheGet(id int) (*Topic, error) {
	mts.RLock()
	item, ok := mts.items[id]
	mts.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mts *MemoryTopicStore) CacheGetUnsafe(id int) (*Topic, error) {
	item, ok := mts.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mts *MemoryTopicStore) Get(id int) (*Topic, error) {
	mts.RLock()
	topic, ok := mts.items[id]
	mts.RUnlock()
	if ok {
		return topic, nil
	}

	topic = &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = mts.CacheAdd(topic)
	}
	return topic, err
}

// BypassGet will always bypass the cache and pull the topic directly from the database
func (mts *MemoryTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

func (mts *MemoryTopicStore) Reload(id int) error {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = mts.CacheSet(topic)
	} else {
		_ = mts.CacheRemove(id)
	}
	return err
}

// TODO: Use a transaction here
func (mts *MemoryTopicStore) Delete(id int) error {
	topic, err := mts.Get(id)
	if err != nil {
		return nil // Already gone, maybe we should check for other errors here
	}

	topicCreator, err := users.Get(topic.CreatedBy)
	if err == nil {
		wcount := wordCount(topic.Content)
		err = topicCreator.decreasePostStats(wcount, true)
		if err != nil {
			return err
		}
	} else if err != ErrNoRows {
		return err
	}

	err = fstore.DecrementTopicCount(topic.ParentID)
	if err != nil && err != ErrNoRows {
		return err
	}

	mts.Lock()
	mts.CacheRemoveUnsafe(id)
	_, err = mts.delete.Exec(id)
	mts.Unlock()
	return err
}

func (mts *MemoryTopicStore) Exists(id int) bool {
	return mts.exists.QueryRow(id).Scan(&id) == nil
}

func (mts *MemoryTopicStore) CacheSet(item *Topic) error {
	mts.Lock()
	_, ok := mts.items[item.ID]
	if ok {
		mts.items[item.ID] = item
	} else if int(mts.length) >= mts.capacity {
		mts.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		mts.items[item.ID] = item
		atomic.AddInt64(&mts.length, 1)
	}
	mts.Unlock()
	return nil
}

func (mts *MemoryTopicStore) CacheAdd(item *Topic) error {
	if int(mts.length) >= mts.capacity {
		return ErrStoreCapacityOverflow
	}
	mts.Lock()
	mts.items[item.ID] = item
	mts.Unlock()
	atomic.AddInt64(&mts.length, 1)
	return nil
}

// TODO: Make these length increments thread-safe. Ditto for the other DataStores
func (mts *MemoryTopicStore) CacheAddUnsafe(item *Topic) error {
	if int(mts.length) >= mts.capacity {
		return ErrStoreCapacityOverflow
	}
	mts.items[item.ID] = item
	atomic.AddInt64(&mts.length, 1)
	return nil
}

// TODO: Make these length decrements thread-safe. Ditto for the other DataStores
func (mts *MemoryTopicStore) CacheRemove(id int) error {
	mts.Lock()
	delete(mts.items, id)
	mts.Unlock()
	atomic.AddInt64(&mts.length, -1)
	return nil
}

func (mts *MemoryTopicStore) CacheRemoveUnsafe(id int) error {
	delete(mts.items, id)
	atomic.AddInt64(&mts.length, -1)
	return nil
}

// ? - What is this? Do we need it? Should it be in the main store interface?
func (mts *MemoryTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

func (mts *MemoryTopicStore) Flush() {
	mts.Lock()
	mts.items = make(map[int]*Topic)
	mts.length = 0
	mts.Unlock()
}

func (mts *MemoryTopicStore) GetLength() int {
	return int(mts.length)
}

func (mts *MemoryTopicStore) SetCapacity(capacity int) {
	mts.capacity = capacity
}

func (mts *MemoryTopicStore) GetCapacity() int {
	return mts.capacity
}

// Return the total number of topics on these forums
func (mts *MemoryTopicStore) GetGlobalCount() int {
	var tcount int
	err := mts.topicCount.QueryRow().Scan(&tcount)
	if err != nil {
		LogError(err)
	}
	return tcount
}

type SQLTopicStore struct {
	get        *sql.Stmt
	exists     *sql.Stmt
	topicCount *sql.Stmt
	delete     *sql.Stmt
}

func NewSQLTopicStore() *SQLTopicStore {
	getStmt, err := qgen.Builder.SimpleSelect("topics", "title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data", "tid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	existsStmt, err := qgen.Builder.SimpleSelect("topics", "tid", "tid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	topicCountStmt, err := qgen.Builder.SimpleCount("topics", "", "")
	if err != nil {
		log.Fatal(err)
	}
	deleteStmt, err := qgen.Builder.SimpleDelete("topics", "tid = ?")
	if err != nil {
		log.Fatal(err)
	}
	return &SQLTopicStore{
		get:        getStmt,
		exists:     existsStmt,
		topicCount: topicCountStmt,
		delete:     deleteStmt,
	}
}

func (sts *SQLTopicStore) Get(id int) (*Topic, error) {
	topic := Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return &topic, err
}

// BypassGet is an alias of Get(), as we don't have a cache for SQLTopicStore
func (sts *SQLTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

// Reload uses a similar query to Exists(), as we don't have any entries to reload, and the secondary benefit of calling Reload() is seeing if the item you're trying to reload exists
func (sts *SQLTopicStore) Reload(id int) error {
	return sts.exists.QueryRow(id).Scan(&id)
}

func (sts *SQLTopicStore) Exists(id int) bool {
	return sts.exists.QueryRow(id).Scan(&id) == nil
}

// TODO: Use a transaction here
func (sts *SQLTopicStore) Delete(id int) error {
	topic, err := sts.Get(id)
	if err != nil {
		return nil // Already gone, maybe we should check for other errors here
	}

	topicCreator, err := users.Get(topic.CreatedBy)
	if err == nil {
		wcount := wordCount(topic.Content)
		err = topicCreator.decreasePostStats(wcount, true)
		if err != nil {
			return err
		}
	} else if err != ErrNoRows {
		return err
	}

	err = fstore.DecrementTopicCount(topic.ParentID)
	if err != nil && err != ErrNoRows {
		return err
	}

	_, err = sts.delete.Exec(id)
	return err
}

// ? - What're we going to do about this?
func (sts *SQLTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

// Return the total number of topics on these forums
func (sts *SQLTopicStore) GetGlobalCount() int {
	var tcount int
	err := sts.topicCount.QueryRow().Scan(&tcount)
	if err != nil {
		LogError(err)
	}
	return tcount
}
