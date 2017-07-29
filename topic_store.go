package main

import "log"
import "sync"
import "database/sql"
import "./query_gen/lib"

// TO-DO: Add the watchdog goroutine
var topics TopicStore

type TopicStore interface {
	Load(id int) error
	Get(id int) (*Topic, error)
	GetUnsafe(id int) (*Topic, error)
	CascadeGet(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	Set(item *Topic) error
	Add(item *Topic) error
	AddUnsafe(item *Topic) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	AddLastTopic(item *Topic, fid int) error
	GetLength() int
	GetCapacity() int
}

type MemoryTopicStore struct {
	items map[int]*Topic
	length int
	capacity int
	get *sql.Stmt
	sync.RWMutex
}

func NewMemoryTopicStore(capacity int) *MemoryTopicStore {
	stmt, err := qgen.Builder.SimpleSelect("topics","title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data","tid = ?","","")
	if err != nil {
		log.Fatal(err)
	}
	return &MemoryTopicStore{
		items:make(map[int]*Topic),
		capacity:capacity,
		get:stmt,
	}
}

func (sts *MemoryTopicStore) Get(id int) (*Topic, error) {
	sts.RLock()
	item, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sts *MemoryTopicStore) GetUnsafe(id int) (*Topic, error) {
	item, ok := sts.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sts *MemoryTopicStore) CascadeGet(id int) (*Topic, error) {
	sts.RLock()
	topic, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return topic, nil
	}

	topic = &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = build_topic_url(name_to_slug(topic.Title),id)
		sts.Add(topic)
	}
	return topic, err
}

func (sts *MemoryTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return topic, err
}

func (sts *MemoryTopicStore) Load(id int) error {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		topic.Link = build_topic_url(name_to_slug(topic.Title),id)
		sts.Set(topic)
	} else {
		sts.Remove(id)
	}
	return err
}

func (sts *MemoryTopicStore) Set(item *Topic) error {
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

func (sts *MemoryTopicStore) Add(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.Lock()
	sts.items[item.ID] = item
	sts.Unlock()
	sts.length++
	return nil
}

func (sts *MemoryTopicStore) AddUnsafe(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.items[item.ID] = item
	sts.length++
	return nil
}

func (sts *MemoryTopicStore) Remove(id int) error {
	sts.Lock()
	delete(sts.items,id)
	sts.Unlock()
	sts.length--
	return nil
}

func (sts *MemoryTopicStore) RemoveUnsafe(id int) error {
	delete(sts.items,id)
	sts.length--
	return nil
}

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

type SqlTopicStore struct {
		get *sql.Stmt
}

func NewSqlTopicStore() *SqlTopicStore {
	stmt, err := qgen.Builder.SimpleSelect("topics","title, content, createdBy, createdAt, is_closed, sticky, parentID, ipaddress, postCount, likeCount, data","tid = ?","","")
	if err != nil {
		log.Fatal(err)
	}
	return &SqlTopicStore{stmt}
}

func (sts *SqlTopicStore) Get(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return &topic, err
}

func (sts *SqlTopicStore) GetUnsafe(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return &topic, err
}

func (sts *SqlTopicStore) CascadeGet(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return &topic, err
}

func (sts *SqlTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return topic, err
}

func (sts *SqlTopicStore) Load(id int) error {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = build_topic_url(name_to_slug(topic.Title),id)
	return err
}

// Placeholder methods, the actual queries are done elsewhere
func (sts *SqlTopicStore) Set(item *Topic) error {
	return nil
}
func (sts *SqlTopicStore) Add(item *Topic) error {
	return nil
}
func (sts *SqlTopicStore) AddUnsafe(item *Topic) error {
	return nil
}
func (sts *SqlTopicStore) Remove(id int) error {
	return nil
}
func (sts *SqlTopicStore) RemoveUnsafe(id int) error {
	return nil
}
func (sts *SqlTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}
func (sts *SqlTopicStore) GetCapacity() int {
	return 0
}

func (sts *SqlTopicStore) GetLength() int {
	return 0 // Return the total number of topics on the forums?
}
