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

	"./query_gen/lib"
)

// TODO: Add the watchdog goroutine
// TODO: Add BulkGetMap
// TODO: Add some sort of update method
var topics TopicStore

type TopicStore interface {
	Reload(id int) error // ? - Should we move this to TopicCache? Might require us to do a lot more casting in Gosora though...
	Get(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	Delete(id int) error
	Exists(id int) bool
	AddLastTopic(item *Topic, fid int) error
	GetGlobalCount() int
}

type TopicCache interface {
	CacheGet(id int) (*Topic, error)
	GetUnsafe(id int) (*Topic, error)
	CacheSet(item *Topic) error
	CacheAdd(item *Topic) error
	CacheAddUnsafe(item *Topic) error
	CacheRemove(id int) error
	CacheRemoveUnsafe(id int) error
	GetLength() int
	SetCapacity(capacity int)
	GetCapacity() int
}

type MemoryTopicStore struct {
	items      map[int]*Topic
	length     int
	capacity   int
	get        *sql.Stmt
	exists     *sql.Stmt
	topicCount *sql.Stmt
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
	return &MemoryTopicStore{
		items:      make(map[int]*Topic),
		capacity:   capacity,
		get:        getStmt,
		exists:     existsStmt,
		topicCount: topicCountStmt,
	}
}

func (sts *MemoryTopicStore) CacheGet(id int) (*Topic, error) {
	sts.RLock()
	item, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sts *MemoryTopicStore) CacheGetUnsafe(id int) (*Topic, error) {
	item, ok := sts.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sts *MemoryTopicStore) Get(id int) (*Topic, error) {
	sts.RLock()
	topic, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return topic, nil
	}

	topic = &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = sts.CacheAdd(topic)
	}
	return topic, err
}

func (sts *MemoryTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

func (sts *MemoryTopicStore) Reload(id int) error {
	topic := &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = sts.CacheSet(topic)
	} else {
		_ = sts.CacheRemove(id)
	}
	return err
}

// TODO: Use a transaction here
func (sts *MemoryTopicStore) Delete(id int) error {
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

	sts.Lock()
	sts.CacheRemoveUnsafe(id)
	_, err = delete_topic_stmt.Exec(id)
	if err != nil {
		sts.Unlock()
		return err
	}
	sts.Unlock()

	return nil
}

func (sts *MemoryTopicStore) Exists(id int) bool {
	return sts.exists.QueryRow(id).Scan(&id) == nil
}

func (sts *MemoryTopicStore) CacheSet(item *Topic) error {
	sts.Lock()
	_, ok := sts.items[item.ID]
	if ok {
		sts.items[item.ID] = item
	} else if sts.length >= sts.capacity {
		sts.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		sts.items[item.ID] = item
		sts.length++
	}
	sts.Unlock()
	return nil
}

func (sts *MemoryTopicStore) CacheAdd(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.Lock()
	sts.items[item.ID] = item
	sts.Unlock()
	sts.length++
	return nil
}

// TODO: Make these length increments thread-safe. Ditto for the other DataStores
func (sts *MemoryTopicStore) CacheAddUnsafe(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.items[item.ID] = item
	sts.length++
	return nil
}

// TODO: Make these length decrements thread-safe. Ditto for the other DataStores
func (sts *MemoryTopicStore) CacheRemove(id int) error {
	sts.Lock()
	delete(sts.items, id)
	sts.Unlock()
	sts.length--
	return nil
}

func (sts *MemoryTopicStore) CacheRemoveUnsafe(id int) error {
	delete(sts.items, id)
	sts.length--
	return nil
}

// ? - What is this? Do we need it? Should it be in the main store interface?
func (sts *MemoryTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

func (sts *MemoryTopicStore) GetLength() int {
	return sts.length
}

func (sts *MemoryTopicStore) SetCapacity(capacity int) {
	sts.capacity = capacity
}

func (sts *MemoryTopicStore) GetCapacity() int {
	return sts.capacity
}

// Return the total number of topics on these forums
func (sts *MemoryTopicStore) GetGlobalCount() int {
	var tcount int
	err := sts.topicCount.QueryRow().Scan(&tcount)
	if err != nil {
		LogError(err)
	}
	return tcount
}

type SQLTopicStore struct {
	get        *sql.Stmt
	exists     *sql.Stmt
	topicCount *sql.Stmt
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
	return &SQLTopicStore{
		get:        getStmt,
		exists:     existsStmt,
		topicCount: topicCountStmt,
	}
}

func (sts *SQLTopicStore) Get(id int) (*Topic, error) {
	topic := Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return &topic, err
}

func (sts *SQLTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

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

	_, err = delete_topic_stmt.Exec(id)
	return err
}

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
