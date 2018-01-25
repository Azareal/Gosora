package common

import (
	"sync"
	"sync/atomic"
)

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

func (mus *MemoryPollCache) Get(id int) (*Poll, error) {
	mus.RLock()
	item, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mus *MemoryPollCache) BulkGet(ids []int) (list []*Poll) {
	list = make([]*Poll, len(ids))
	mus.RLock()
	for i, id := range ids {
		list[i] = mus.items[id]
	}
	mus.RUnlock()
	return list
}

func (mus *MemoryPollCache) GetUnsafe(id int) (*Poll, error) {
	item, ok := mus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

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

func (mus *MemoryPollCache) Add(item *Poll) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.Lock()
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	mus.Unlock()
	return nil
}

func (mus *MemoryPollCache) AddUnsafe(item *Poll) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	return nil
}

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

func (mus *MemoryPollCache) RemoveUnsafe(id int) error {
	_, ok := mus.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

func (mus *MemoryPollCache) Flush() {
	mus.Lock()
	mus.items = make(map[int]*Poll)
	mus.length = 0
	mus.Unlock()
}

// ! Is this concurrent?
// Length returns the number of users in the memory cache
func (mus *MemoryPollCache) Length() int {
	return int(mus.length)
}

func (mus *MemoryPollCache) SetCapacity(capacity int) {
	mus.capacity = capacity
}

func (mus *MemoryPollCache) GetCapacity() int {
	return mus.capacity
}

type NullPollCache struct {
}

// NewNullPollCache gives you a new instance of NullPollCache
func NewNullPollCache() *NullPollCache {
	return &NullPollCache{}
}

func (mus *NullPollCache) Get(id int) (*Poll, error) {
	return nil, ErrNoRows
}

func (mus *NullPollCache) BulkGet(ids []int) (list []*Poll) {
	return list
}

func (mus *NullPollCache) GetUnsafe(id int) (*Poll, error) {
	return nil, ErrNoRows
}

func (mus *NullPollCache) Set(_ *Poll) error {
	return nil
}

func (mus *NullPollCache) Add(item *Poll) error {
	_ = item
	return nil
}

func (mus *NullPollCache) AddUnsafe(item *Poll) error {
	_ = item
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
