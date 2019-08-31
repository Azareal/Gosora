package common

import (
	"sync"
	"sync/atomic"
)

// UserCache is an interface which spits out users from a fast cache rather than the database, whether from memory or from an application like Redis. Users may not be present in the cache but may be in the database
type UserCache interface {
	DeallocOverflow(evictPriority bool) (evicted int) // May cause thread contention, looks for items to evict
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
	items    map[int]*User // TODO: Shard this into two?
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

// TODO: Avoid deallocating topic list users
func (s *MemoryUserCache) DeallocOverflow(evictPriority bool) (evicted int) {
	toEvict := make([]int, 10)
	evIndex := 0
	s.RLock()
	for _, user := range s.items {
		if /*user.LastActiveAt < lastActiveCutoff && */ user.Score == 0 && !user.IsMod {
			if EnableWebsockets && WsHub.HasUser(user.ID) {
				continue
			}
			toEvict[evIndex] = user.ID
			evIndex++
			if evIndex == 10 {
				break
			}
		}
	}
	s.RUnlock()

	// Clear some of the less active users now with a bit more aggressiveness
	if evIndex == 0 && evictPriority {
		toEvict = make([]int, 20)
		s.RLock()
		for _, user := range s.items {
			if user.Score < 100 && !user.IsMod {
				if EnableWebsockets && WsHub.HasUser(user.ID) {
					continue
				}
				toEvict[evIndex] = user.ID
				evIndex++
				if evIndex == 20 {
					break
				}
			}
		}
		s.RUnlock()
	}

	// Remove zero IDs from the evictable list, so we don't waste precious cycles locked for those
	lastZero := -1
	for i, uid := range toEvict {
		if uid == 0 {
			lastZero = i
		}
	}
	if lastZero != -1 {
		toEvict = toEvict[:lastZero]
	}

	s.BulkRemove(toEvict)
	return len(toEvict)
}

// Get fetches a user by ID. Returns ErrNoRows if not present.
func (s *MemoryUserCache) Get(id int) (*User, error) {
	s.RLock()
	item, ok := s.items[id]
	s.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// BulkGet fetches multiple users by their IDs. Indices without users will be set to nil, so make sure you check for those, we might want to change this behaviour to make it less confusing.
func (s *MemoryUserCache) BulkGet(ids []int) (list []*User) {
	list = make([]*User, len(ids))
	s.RLock()
	for i, id := range ids {
		list[i] = s.items[id]
	}
	s.RUnlock()
	return list
}

// GetUnsafe fetches a user by ID. Returns ErrNoRows if not present. THIS METHOD IS NOT THREAD-SAFE.
func (s *MemoryUserCache) GetUnsafe(id int) (*User, error) {
	item, ok := s.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

// Set overwrites the value of a user in the cache, whether it's present or not. May return a capacity overflow error.
func (s *MemoryUserCache) Set(item *User) error {
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

// Add adds a user to the cache, similar to Set, but it's only intended for new items. This method might be deprecated in the near future, use Set. May return a capacity overflow error.
// ? Is this redundant if we have Set? Are the efficiency wins worth this? Is this even used?
func (s *MemoryUserCache) Add(item *User) error {
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
func (s *MemoryUserCache) AddUnsafe(item *User) error {
	if int(s.length) >= s.capacity {
		return ErrStoreCapacityOverflow
	}
	s.items[item.ID] = item
	s.length = int64(len(s.items))
	return nil
}

// Remove removes a user from the cache by ID, if they exist. Returns ErrNoRows if no items exist.
func (s *MemoryUserCache) Remove(id int) error {
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
func (s *MemoryUserCache) RemoveUnsafe(id int) error {
	_, ok := s.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(s.items, id)
	atomic.AddInt64(&s.length, -1)
	return nil
}

func (s *MemoryUserCache) BulkRemove(ids []int) {
	var rCount int64
	s.Lock()
	for _, id := range ids {
		_, ok := s.items[id]
		if ok {
			delete(s.items, id)
			rCount++
		}
	}
	s.Unlock()
	atomic.AddInt64(&s.length, -rCount)
}

// Flush removes all the users from the cache, useful for tests.
func (s *MemoryUserCache) Flush() {
	s.Lock()
	s.items = make(map[int]*User)
	s.length = 0
	s.Unlock()
}

// ! Is this concurrent?
// Length returns the number of users in the memory cache
func (s *MemoryUserCache) Length() int {
	return int(s.length)
}

// SetCapacity sets the maximum number of users which this cache can hold
func (s *MemoryUserCache) SetCapacity(capacity int) {
	// Ints are moved in a single instruction, so this should be thread-safe
	s.capacity = capacity
}

// GetCapacity returns the maximum number of users this cache can hold
func (s *MemoryUserCache) GetCapacity() int {
	return s.capacity
}
