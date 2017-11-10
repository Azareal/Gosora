/*
*
*	Gosora Topic Store
*	Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"../query_gen/lib"
)

// TODO: Add the watchdog goroutine
// TODO: Add BulkGetMap
// TODO: Add some sort of update method
// ? - Should we add stick, lock, unstick, and unlock methods? These might be better on the Topics not the TopicStore
var topics TopicStore
var ErrNoTitle = errors.New("This message is missing a title")
var ErrNoBody = errors.New("This message is missing a body")

type TopicStore interface {
	Get(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	Exists(id int) bool
	Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error)
	AddLastTopic(item *Topic, fid int) error // unimplemented
	// TODO: Implement these two methods
	//Replies(tid int) ([]*Reply, error)
	//RepliesRange(tid int, lower int, higher int) ([]*Reply, error)
	GlobalCount() int
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
	Reload(id int) error
	Length() int
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
	create     *sql.Stmt
	sync.RWMutex
}

// NewMemoryTopicStore gives you a new instance of MemoryTopicStore
func NewMemoryTopicStore(capacity int) (*MemoryTopicStore, error) {
	acc := qgen.Builder.Accumulator()
	return &MemoryTopicStore{
		items:      make(map[int]*Topic),
		capacity:   capacity,
		get:        acc.SimpleSelect("topics", "title, content, createdBy, createdAt, lastReplyAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data", "tid = ?", "", ""),
		exists:     acc.SimpleSelect("topics", "tid", "tid = ?", "", ""),
		topicCount: acc.SimpleCount("topics", "", ""),
		create:     acc.SimpleInsert("topics", "parentID, title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, ipaddress, words, createdBy", "?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?"),
	}, acc.FirstError()
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
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = mts.CacheAdd(topic)
	}
	return topic, err
}

// BypassGet will always bypass the cache and pull the topic directly from the database
func (mts *MemoryTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

func (mts *MemoryTopicStore) Reload(id int) error {
	topic := &Topic{ID: id}
	err := mts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
		_ = mts.CacheSet(topic)
	} else {
		_ = mts.CacheRemove(id)
	}
	return err
}

func (mts *MemoryTopicStore) Exists(id int) bool {
	return mts.exists.QueryRow(id).Scan(&id) == nil
}

func (mts *MemoryTopicStore) Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error) {
	topicName = strings.TrimSpace(topicName)
	if topicName == "" {
		return 0, ErrNoBody
	}

	content = strings.TrimSpace(content)
	parsedContent := parseMessage(content, fid, "forums")
	if strings.TrimSpace(parsedContent) == "" {
		return 0, ErrNoBody
	}

	wcount := wordCount(content)
	// TODO: Move this statement into the topic store
	res, err := mts.create.Exec(fid, topicName, content, parsedContent, uid, ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = fstore.AddTopic(int(lastID), uid, fid)
	return int(lastID), err
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

// ! Is this concurrent?
// Length returns the number of topics in the memory cache
func (mts *MemoryTopicStore) Length() int {
	return int(mts.length)
}

func (mts *MemoryTopicStore) SetCapacity(capacity int) {
	mts.capacity = capacity
}

func (mts *MemoryTopicStore) GetCapacity() int {
	return mts.capacity
}

// GlobalCount returns the total number of topics on these forums
func (mts *MemoryTopicStore) GlobalCount() int {
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
	create     *sql.Stmt
}

func NewSQLTopicStore() (*SQLTopicStore, error) {
	acc := qgen.Builder.Accumulator()
	return &SQLTopicStore{
		get:        acc.SimpleSelect("topics", "title, content, createdBy, createdAt, lastReplyAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data", "tid = ?", "", ""),
		exists:     acc.SimpleSelect("topics", "tid", "tid = ?", "", ""),
		topicCount: acc.SimpleCount("topics", "", ""),
		create:     acc.SimpleInsert("topics", "parentID, title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, ipaddress, words, createdBy", "?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?"),
	}, acc.FirstError()
}

func (sts *SQLTopicStore) Get(id int) (*Topic, error) {
	topic := Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return &topic, err
}

// BypassGet is an alias of Get(), as we don't have a cache for SQLTopicStore
func (sts *SQLTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID: id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.LastReplyAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), id)
	return topic, err
}

func (sts *SQLTopicStore) Exists(id int) bool {
	return sts.exists.QueryRow(id).Scan(&id) == nil
}

func (sts *SQLTopicStore) Create(fid int, topicName string, content string, uid int, ipaddress string) (tid int, err error) {
	topicName = strings.TrimSpace(topicName)
	if topicName == "" {
		return 0, ErrNoBody
	}

	content = strings.TrimSpace(content)
	parsedContent := parseMessage(content, fid, "forums")
	if strings.TrimSpace(parsedContent) == "" {
		return 0, ErrNoBody
	}

	wcount := wordCount(content)
	// TODO: Move this statement into the topic store
	res, err := sts.create.Exec(fid, topicName, content, parsedContent, uid, ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = fstore.AddTopic(int(lastID), uid, fid)
	return int(lastID), err
}

// ? - What're we going to do about this?
func (sts *SQLTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

// GlobalCount returns the total number of topics on these forums
func (sts *SQLTopicStore) GlobalCount() int {
	var tcount int
	err := sts.topicCount.QueryRow().Scan(&tcount)
	if err != nil {
		LogError(err)
	}
	return tcount
}
