package main

import "sync"
import "database/sql"

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

type StaticTopicStore struct {
	items map[int]*Topic
	length int
	capacity int
	get *sql.Stmt
	sync.RWMutex
}

func NewStaticTopicStore(capacity int) *StaticTopicStore {
	return &StaticTopicStore{items:make(map[int]*Topic),capacity:capacity,get:get_topic_stmt}
}

func (sts *StaticTopicStore) Get(id int) (*Topic, error) {
	sts.RLock()
	item, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sts *StaticTopicStore) GetUnsafe(id int) (*Topic, error) {
	item, ok := sts.items[id]
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sts *StaticTopicStore) CascadeGet(id int) (*Topic, error) {
	sts.RLock()
	topic, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return topic, nil
	}

	topic = &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		sts.Add(topic)
	}
	return topic, err
}

func (sts *StaticTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	return topic, err
}

func (sts *StaticTopicStore) Load(id int) error {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	if err == nil {
		sts.Set(topic)
	} else {
		sts.Remove(id)
	}
	return err
}

func (sts *StaticTopicStore) Set(item *Topic) error {
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

func (sts *StaticTopicStore) Add(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.Lock()
	sts.items[item.ID] = item
	sts.Unlock()
	sts.length++
	return nil
}

func (sts *StaticTopicStore) AddUnsafe(item *Topic) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.items[item.ID] = item
	sts.length++
	return nil
}

func (sts *StaticTopicStore) Remove(id int) error {
	sts.Lock()
	delete(sts.items,id)
	sts.Unlock()
	sts.length--
	return nil
}

func (sts *StaticTopicStore) RemoveUnsafe(id int) error {
	delete(sts.items,id)
	sts.length--
	return nil
}

func (sts *StaticTopicStore) AddLastTopic(item *Topic, fid int) error {
	// Coming Soon...
	return nil
}

func (sts *StaticTopicStore) GetLength() int {
	return sts.length
}

func (sts *StaticTopicStore) SetCapacity(capacity int) {
	sts.capacity = capacity
}

func (sts *StaticTopicStore) GetCapacity() int {
	return sts.capacity
}

//type DynamicTopicStore struct {
//	items_expiries list.List
//	items map[int]*Topic
//}

type SqlTopicStore struct {
		get *sql.Stmt
}

func NewSqlTopicStore() *SqlTopicStore {
	return &SqlTopicStore{get_topic_stmt}
}

func (sts *SqlTopicStore) Get(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	return &topic, err
}

func (sts *SqlTopicStore) GetUnsafe(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	return &topic, err
}

func (sts *SqlTopicStore) CascadeGet(id int) (*Topic, error) {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	return &topic, err
}

func (sts *SqlTopicStore) BypassGet(id int) (*Topic, error) {
	topic := &Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	return topic, err
}

func (sts *SqlTopicStore) Load(id int) error {
	topic := Topic{ID:id}
	err := sts.get.QueryRow(id).Scan(&topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.Is_Closed, &topic.Sticky, &topic.ParentID, &topic.IpAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
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
