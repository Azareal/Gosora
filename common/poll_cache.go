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
func (mus *MemoryPollCache) Get(id int) (*Poll, error) {
	mus.RLock()
	item, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple polls by their IDs. Indices without polls will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (mus *MemoryPollCache) BulkGet(ids []int) (list []*Poll) {
	list = make([]*Poll, len(ids))
	mus.RLock()
	for i, id := range ids {
		list[i] = mus.items[id]
	}
	mus.RUnlock()
	return list
}

// GetUnsafe fetches a poll by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (mus *MemoryPollCache) GetUnsafe(id int) (*Poll, error) {
	item, ok := mus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// Set overwrites the value of a poll in the cache, whether it's present or not. May return a capacity overflow error.
func (mus *MemoryPollCache) Set(item *Poll) error {
	mus.Lock()
	user, ok := mus.items[item.ID]
	if ok {
		mus.Unlock()
		*user = *item
	} else if int(mus.length) >= mus.capacity {
		mus.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		mus.items[item.ID] = item
		mus.Unlock()
		atomic.AddInt64(&mus.length, 1)
	}
	return nil
}

// Add adds a poll to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (mus *MemoryPollCache) Add(item *Poll) error {
	mus.Lock()
	if int(mus.length) >= mus.capacity {
		mus.Unlock()
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	mus.Unlock()
	return nil
}

// AddUnsafe is the unsafe version of Add. May return a capacity overflow error. THIS METHOD IS NOT THREAD-SAFE.
func (mus *MemoryPollCache) AddUnsafe(item *Poll) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	return nil
}

// Remove removes a poll from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (mus *MemoryPollCache) Remove(id int) error {
	mus.Lock()
	_, ok := mus.items[id]
	if !ok {
		mus.Unlock()
		return ErrNoRows
	}
	delete(mus.items, id)
	mus.Unlock()
	atomic.AddInt64(&mus.length, -1)
	return nil
}

// RemoveUnsafe is the unsafe version of Remove. THIS METHOD IS NOT THREAD-SAFE.
func (mus *MemoryPollCache) RemoveUnsafe(id int) error {
	_, ok := mus.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

// Flush removes all the polls from the cache, useful for tests.
func (mus *MemoryPollCache) Flush() {
	mus.Lock()
	mus.items = make(map[int]*Poll)
	mus.length = 0
	mus.Unlock()
}

// ! Is this concurrent?
// Length returns the number of polls in the memory cache
func (mus *MemoryPollCache) Length() int {
	return int(mus.length)
}

// SetCapacity sets the maximum number of polls which this cache can hold
func (mus *MemoryPollCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	mus.capacity = capacity
}

// GetCapacity returns the maximum number of polls this cache can hold
func (mus *MemoryPollCache) GetCapacity() int {
	return mus.capacity
}

// NullPollCache is a poll cache to be used when you don't want a cache and just want queries to passthrough to the database
type NullPollCache struct {
}

// NewNullPollCache gives you a new instance of NullPollCache
func NewNullPollCache() *NullPollCache {
	return &NullPollCache{}
}

// nolint
func (mus *NullPollCache) Get(id int) (*Poll, error) {
	return nil, ErrNoRows
}
func (mus *NullPollCache) BulkGet(ids []int) (list []*Poll) {
	return make([]*Poll, len(ids))
}
func (mus *NullPollCache) GetUnsafe(id int) (*Poll, error) {
	return nil, ErrNoRows
}
func (mus *NullPollCache) Set(_ *Poll) error {
	return nil
}
func (mus *NullPollCache) Add(_ *Poll) error {
	return nil
}
func (mus *NullPollCache) AddUnsafe(_ *Poll) error {
	return nil
}
func (mus *NullPollCache) Remove(id int) error {
	return nil
}
func (mus *NullPollCache) RemoveUnsafe(id int) error {
	return nil
}
func (mus *NullPollCache) Flush() {
}
func (mus *NullPollCache) Length() int {
	return 0
}
func (mus *NullPollCache) SetCapacity(_ int) {
}
func (mus *NullPollCache) GetCapacity() int {
	return 0
}
