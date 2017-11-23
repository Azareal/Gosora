package common

import (
	"sync"
	"sync/atomic"
)

type TopicCache interface {
	Get(id int) (*Topic, error)
	GetUnsafe(id int) (*Topic, error)
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

func (mts *MemoryTopicCache) Get(id int) (*Topic, error) {
	mts.RLock()
	item, ok := mts.items[id]
	mts.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mts *MemoryTopicCache) GetUnsafe(id int) (*Topic, error) {
	item, ok := mts.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mts *MemoryTopicCache) Set(item *Topic) error {
	mts.Lock()
	_, ok := mts.items[item.ID]
	if ok {
		mts.items[item.ID] = item
	} else if int(mts.length) >= mts.capacity {
		mts.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		mts.items[item.ID] = item
		atomic.AddInt64(&mts.length, 1)
	}
	mts.Unlock()
	return nil
}

func (mts *MemoryTopicCache) Add(item *Topic) error {
	if int(mts.length) >= mts.capacity {
		return ErrStoreCapacityOverflow
	}
	mts.Lock()
	mts.items[item.ID] = item
	mts.Unlock()
	atomic.AddInt64(&mts.length, 1)
	return nil
}

// TODO: Make these length increments thread-safe. Ditto for the other DataStores
func (mts *MemoryTopicCache) AddUnsafe(item *Topic) error {
	if int(mts.length) >= mts.capacity {
		return ErrStoreCapacityOverflow
	}
	mts.items[item.ID] = item
	atomic.AddInt64(&mts.length, 1)
	return nil
}

// TODO: Make these length decrements thread-safe. Ditto for the other DataStores
func (mts *MemoryTopicCache) Remove(id int) error {
	mts.Lock()
	delete(mts.items, id)
	mts.Unlock()
	atomic.AddInt64(&mts.length, -1)
	return nil
}

func (mts *MemoryTopicCache) RemoveUnsafe(id int) error {
	delete(mts.items, id)
	atomic.AddInt64(&mts.length, -1)
	return nil
}

func (mts *MemoryTopicCache) Flush() {
	mts.Lock()
	mts.items = make(map[int]*Topic)
	mts.length = 0
	mts.Unlock()
}

// ! Is this concurrent?
// Length returns the number of topics in the memory cache
func (mts *MemoryTopicCache) Length() int {
	return int(mts.length)
}

func (mts *MemoryTopicCache) SetCapacity(capacity int) {
	mts.capacity = capacity
}

func (mts *MemoryTopicCache) GetCapacity() int {
	return mts.capacity
}
