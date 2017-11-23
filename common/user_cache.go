package common

import (
	"sync"
	"sync/atomic"
)

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

func (mus *MemoryUserCache) Get(id int) (*User, error) {
	mus.RLock()
	item, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mus *MemoryUserCache) BulkGet(ids []int) (list []*User) {
	list = make([]*User, len(ids))
	mus.RLock()
	for i, id := range ids {
		list[i] = mus.items[id]
	}
	mus.RUnlock()
	return list
}

func (mus *MemoryUserCache) GetUnsafe(id int) (*User, error) {
	item, ok := mus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

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

func (mus *MemoryUserCache) Add(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.Lock()
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	mus.Unlock()
	return nil
}

func (mus *MemoryUserCache) AddUnsafe(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	return nil
}

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

func (mus *MemoryUserCache) RemoveUnsafe(id int) error {
	_, ok := mus.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

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

func (mus *MemoryUserCache) SetCapacity(capacity int) {
	mus.capacity = capacity
}

func (mus *MemoryUserCache) GetCapacity() int {
	return mus.capacity
}
