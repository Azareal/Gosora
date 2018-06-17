package common

import (
	"sync"
	"sync/atomic"
)

// UserCache is an interface which spits out users from a fast cache rather than the database, whether from memory or from an application like Redis. Users may not be present in the cache but may be in the database
type UserCache interface {
	Get(id int) (*User, error)
	GetUnsafe(id int) (*User, error)
	BulkGet(ids []int) (list []*User)
	Set(item *User) error
	Add(item *User) error
	AddUnsafe(item *User) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	Flush()
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}

// MemoryUserCache stores and pulls users out of the current process' memory
type MemoryUserCache struct {
	items    map[int]*User
	length   int64
	capacity int

	sync.RWMutex
}

// NewMemoryUserCache gives you a new instance of MemoryUserCache
func NewMemoryUserCache(capacity int) *MemoryUserCache {
	return &MemoryUserCache{
		items:    make(map[int]*User),
		capacity: capacity,
	}
}

// Get fetches a user by ID. Returns ErrNoRows if not present.
func (mus *MemoryUserCache) Get(id int) (*User, error) {
	mus.RLock()
	item, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple users by their IDs. Indices without users will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (mus *MemoryUserCache) BulkGet(ids []int) (list []*User) {
	list = make([]*User, len(ids))
	mus.RLock()
	for i, id := range ids {
		list[i] = mus.items[id]
	}
	mus.RUnlock()
	return list
}

// GetUnsafe fetches a user by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (mus *MemoryUserCache) GetUnsafe(id int) (*User, error) {
	item, ok := mus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// Set overwrites the value of a user in the cache, whether it's present or not. May return a capacity overflow error.
func (mus *MemoryUserCache) Set(item *User) error {
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

// Add adds a user to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (mus *MemoryUserCache) Add(item *User) error {
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
func (mus *MemoryUserCache) AddUnsafe(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	return nil
}

// Remove removes a user from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (mus *MemoryUserCache) Remove(id int) error {
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
func (mus *MemoryUserCache) RemoveUnsafe(id int) error {
	_, ok := mus.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

// Flush removes all the users from the cache, useful for tests.
func (mus *MemoryUserCache) Flush() {
	mus.Lock()
	mus.items = make(map[int]*User)
	mus.length = 0
	mus.Unlock()
}

// ! Is this concurrent?
// Length returns the number of users in the memory cache
func (mus *MemoryUserCache) Length() int {
	return int(mus.length)
}

// SetCapacity sets the maximum number of users which this cache can hold
func (mus *MemoryUserCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	mus.capacity = capacity
}

// GetCapacity returns the maximum number of users this cache can hold
func (mus *MemoryUserCache) GetCapacity() int {
	return mus.capacity
}
