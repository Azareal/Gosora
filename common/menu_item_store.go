package common

import "sync"

type DefaultMenuItemStore struct {
	items map[int]MenuItem
	lock  sync.RWMutex
}

func NewDefaultMenuItemStore() *DefaultMenuItemStore {
	return &DefaultMenuItemStore{
		items: make(map[int]MenuItem),
	}
}

func (s *DefaultMenuItemStore) Add(item MenuItem) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items[item.ID] = item
}

func (s *DefaultMenuItemStore) Get(id int) (MenuItem, error) {
	s.lock.RLock()
	item, ok := s.items[id]
	s.lock.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}
