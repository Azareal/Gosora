package common

import "sync"

type DefaultMenuItemStore struct {
	items    map[int]MenuItem
	itemLock sync.RWMutex
}

func NewDefaultMenuItemStore() *DefaultMenuItemStore {
	return &DefaultMenuItemStore{
		items: make(map[int]MenuItem),
	}
}

func (store *DefaultMenuItemStore) Add(item MenuItem) {
	store.itemLock.Lock()
	defer store.itemLock.Unlock()
	store.items[item.ID] = item
}

func (store *DefaultMenuItemStore) Get(id int) (MenuItem, error) {
	store.itemLock.RLock()
	item, ok := store.items[id]
	store.itemLock.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}
