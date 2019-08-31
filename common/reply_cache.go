package common

import (
	"log"
	"sync"
	"sync/atomic"
)

// ReplyCache is an interface which spits out replies from a fast cache rather than the database, whether from memory or from an application like Redis. Replies may not be present in the cache but may be in the database
type ReplyCache interface {
	Get(id int) (*Reply, error)
	GetUnsafe(id int) (*Reply, error)
	BulkGet(ids []int) (list []*Reply)
	Set(item *Reply) error
	Add(item *Reply) error
	AddUnsafe(item *Reply) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	Flush()
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}

// MemoryReplyCache stores and pulls replies out of the current process' memory
type MemoryReplyCache struct {
	items    map[int]*Reply
	length   int64 // sync/atomic only lets us operate on int32s and int64s
	capacity int

	sync.RWMutex
}

// NewMemoryReplyCache gives you a new instance of MemoryReplyCache
func NewMemoryReplyCache(capacity int) *MemoryReplyCache {
	return &MemoryReplyCache{
		items:    make(map[int]*Reply),
		capacity: capacity,
	}
}

// Get fetches a reply by ID. Returns ErrNoRows if not present.
func (s *MemoryReplyCache) Get(id int) (*Reply, error) {
	s.RLock()
	item, ok := s.items[id]
	s.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// GetUnsafe fetches a reply by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryReplyCache) GetUnsafe(id int) (*Reply, error) {
	item, ok := s.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple replies by their IDs. Indices without replies will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (s *MemoryReplyCache) BulkGet(ids []int) (list []*Reply) {
	list = make([]*Reply, len(ids))
	s.RLock()
	for i, id := range ids {
		list[i] = s.items[id]
	}
	s.RUnlock()
	return list
}

// Set overwrites the value of a reply in the cache, whether it's present or not. May return a capacity overflow error.
func (s *MemoryReplyCache) Set(item *Reply) error {
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

// Add adds a reply to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (s *MemoryReplyCache) Add(item *Reply) error {
	log.Print("MemoryReplyCache.Add")
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
func (s *MemoryReplyCache) AddUnsafe(item *Reply) error {
	if int(s.length) >= s.capacity {
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.length = int64(len(s.items))
	return nil
}

// Remove removes a reply from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (s *MemoryReplyCache) Remove(id int) error {
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
func (s *MemoryReplyCache) RemoveUnsafe(id int) error {
	_, ok := s.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(s.items, id)
	atomic.AddInt64(&s.length, -1)
	return nil
}

// Flush removes all the replies from the cache, useful for tests.
func (s *MemoryReplyCache) Flush() {
	s.Lock()
	s.items = make(map[int]*Reply)
	s.length = 0
	s.Unlock()
}

// ! Is this concurrent?
// Length returns the number of replies in the memory cache
func (s *MemoryReplyCache) Length() int {
	return int(s.length)
}

// SetCapacity sets the maximum number of replies which this cache can hold
func (s *MemoryReplyCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	s.capacity = capacity
}

// GetCapacity returns the maximum number of replies this cache can hold
func (s *MemoryReplyCache) GetCapacity() int {
	return s.capacity
}
