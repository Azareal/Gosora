package common

import (
	"sync"
	"sync/atomic"
)

// PollCache is an interface which spits out polls from a fast cache rather than the database, whether from memory or from an application like Redis. Polls may not be present in the cache but may be in the database
type PollCache interface {
	Get(id int) (*Poll, error)
	GetUnsafe(id int) (*Poll, error)
	BulkGet(ids []int) (list []*Poll)
	Set(item *Poll) error
	Add(item *Poll) error
	AddUnsafe(item *Poll) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	Flush()
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}

// MemoryPollCache stores and pulls polls out of the current process' memory
type MemoryPollCache struct {
	items    map[int]*Poll
	length   int64
	capacity int

	sync.RWMutex
}

// NewMemoryPollCache gives you a new instance of MemoryPollCache
func NewMemoryPollCache(capacity int) *MemoryPollCache {
	return &MemoryPollCache{
		items:    make(map[int]*Poll),
		capacity: capacity,
	}
}

// Get fetches a poll by ID. Returns ErrNoRows if not present.
func (s *MemoryPollCache) Get(id int) (*Poll, error) {
	s.RLock()
	item, ok := s.items[id]
	s.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple polls by their IDs. Indices without polls will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (s *MemoryPollCache) BulkGet(ids []int) (list []*Poll) {
	list = make([]*Poll, len(ids))
	s.RLock()
	for i, id := range ids {
		list[i] = s.items[id]
	}
	s.RUnlock()
	return list
}

// GetUnsafe fetches a poll by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryPollCache) GetUnsafe(id int) (*Poll, error) {
	item, ok := s.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// Set overwrites the value of a poll in the cache, whether it's present or not. May return a capacity overflow error.
func (s *MemoryPollCache) Set(item *Poll) error {
	s.Lock()
	user, ok := s.items[item.ID]
	if ok {
		s.Unlock()
		*user = *item
	} else if int(s.length) >= s.capacity {
		s.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		s.items[item.ID] = item
		s.Unlock()
		atomic.AddInt64(&s.length, 1)
	}
	return nil
}

// Add adds a poll to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (s *MemoryPollCache) Add(item *Poll) error {
	s.Lock()
	if int(s.length) >= s.capacity {
		s.Unlock()
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.length = int64(len(s.items))
	s.Unlock()
	return nil
}

// AddUnsafe is the unsafe version of Add. May return a capacity overflow error. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryPollCache) AddUnsafe(item *Poll) error {
	if int(s.length) >= s.capacity {
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.length = int64(len(s.items))
	return nil
}

// Remove removes a poll from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (s *MemoryPollCache) Remove(id int) error {
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
func (s *MemoryPollCache) RemoveUnsafe(id int) error {
	_, ok := s.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(s.items, id)
	atomic.AddInt64(&s.length, -1)
	return nil
}

// Flush removes all the polls from the cache, useful for tests.
func (s *MemoryPollCache) Flush() {
	m := make(map[int]*Poll)
	s.Lock()
	s.items = m
	s.length = 0
	s.Unlock()
}

// ! Is this concurrent?
// Length returns the number of polls in the memory cache
func (s *MemoryPollCache) Length() int {
	return int(s.length)
}

// SetCapacity sets the maximum number of polls which this cache can hold
func (s *MemoryPollCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	s.capacity = capacity
}

// GetCapacity returns the maximum number of polls this cache can hold
func (s *MemoryPollCache) GetCapacity() int {
	return s.capacity
}

// NullPollCache is a poll cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullPollCache struct {
}

// NewNullPollCache gives you a new instance of NullPollCache
func NewNullPollCache() *NullPollCache {
	return &NullPollCache{}
}

// nolint
func (s *NullPollCache) Get(id int) (*Poll, error) {
	return nil, ErrNoRows
}
func (s *NullPollCache) BulkGet(ids []int) (list []*Poll) {
	return make([]*Poll, len(ids))
}
func (s *NullPollCache) GetUnsafe(id int) (*Poll, error) {
	return nil, ErrNoRows
}
func (s *NullPollCache) Set(_ *Poll) error {
	return nil
}
func (s *NullPollCache) Add(_ *Poll) error {
	return nil
}
func (s *NullPollCache) AddUnsafe(_ *Poll) error {
	return nil
}
func (s *NullPollCache) Remove(id int) error {
	return nil
}
func (s *NullPollCache) RemoveUnsafe(id int) error {
	return nil
}
func (s *NullPollCache) Flush() {
}
func (s *NullPollCache) Length() int {
	return 0
}
func (s *NullPollCache) SetCapacity(_ int) {
}
func (s *NullPollCache) GetCapacity() int {
	return 0
}
