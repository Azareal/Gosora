package common

import (
	"sync"
	"sync/atomic"
)

// TopicCache is an interface which spits out topics from a fast cache rather than the database, whether from memory or from an application like Redis. Topics may not be present in the cache but may be in the database
type TopicCache interface {
	Get(id int) (*Topic, error)
	GetUnsafe(id int) (*Topic, error)
	BulkGet(ids []int) (list []*Topic)
	Set(item *Topic) error
	Add(item *Topic) error
	AddUnsafe(item *Topic) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	Flush()
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}

// MemoryTopicCache stores and pulls topics out of the current process' memory
type MemoryTopicCache struct {
	items    map[int]*Topic
	length   int64 // sync/atomic only lets us operate on int32s and int64s
	capacity int

	sync.RWMutex
}

// NewMemoryTopicCache gives you a new instance of MemoryTopicCache
func NewMemoryTopicCache(capacity int) *MemoryTopicCache {
	return &MemoryTopicCache{
		items:    make(map[int]*Topic),
		capacity: capacity,
	}
}

// Get fetches a topic by ID. Returns ErrNoRows if not present.
func (s *MemoryTopicCache) Get(id int) (*Topic, error) {
	s.RLock()
	item, ok := s.items[id]
	s.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// GetUnsafe fetches a topic by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryTopicCache) GetUnsafe(id int) (*Topic, error) {
	item, ok := s.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple topics by their IDs. Indices without topics will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (s *MemoryTopicCache) BulkGet(ids []int) (list []*Topic) {
	list = make([]*Topic, len(ids))
	s.RLock()
	for i, id := range ids {
		list[i] = s.items[id]
	}
	s.RUnlock()
	return list
}

// Set overwrites the value of a topic in the cache, whether it's present or not. May return a capacity overflow error.
func (s *MemoryTopicCache) Set(item *Topic) error {
	s.Lock()
	_, ok := s.items[item.ID]
	if ok {
		s.items[item.ID] = item
	} else if int(s.length) >= s.capacity {
		s.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		s.items[item.ID] = item
		atomic.AddInt64(&s.length, 1)
	}
	s.Unlock()
	return nil
}

// Add adds a topic to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (s *MemoryTopicCache) Add(item *Topic) error {
	s.Lock()
	if int(s.length) >= s.capacity {
		s.Unlock()
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.Unlock()
	atomic.AddInt64(&s.length, 1)
	return nil
}

// AddUnsafe is the unsafe version of Add. May return a capacity overflow error. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryTopicCache) AddUnsafe(item *Topic) error {
	if int(s.length) >= s.capacity {
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.length = int64(len(s.items))
	return nil
}

// Remove removes a topic from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (s *MemoryTopicCache) Remove(id int) error {
	s.Lock()
	_, ok := s.items[id]
	if !ok {
		s.Unlock()
		return ErrNoRows
	}
	delete(s.items, id)
	s.Unlock()
	atomic.AddInt64(&s.length, -1)
	return nil
}

// RemoveUnsafe is the unsafe version of Remove. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryTopicCache) RemoveUnsafe(id int) error {
	_, ok := s.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(s.items, id)
	atomic.AddInt64(&s.length, -1)
	return nil
}

// Flush removes all the topics from the cache, useful for tests.
func (s *MemoryTopicCache) Flush() {
	s.Lock()
	s.items = make(map[int]*Topic)
	s.length = 0
	s.Unlock()
}

// ! Is this concurrent?
// Length returns the number of topics in the memory cache
func (s *MemoryTopicCache) Length() int {
	return int(s.length)
}

// SetCapacity sets the maximum number of topics which this cache can hold
func (s *MemoryTopicCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	s.capacity = capacity
}

// GetCapacity returns the maximum number of topics this cache can hold
func (s *MemoryTopicCache) GetCapacity() int {
	return s.capacity
}
