package common

import (
	"database/sql"
	"sync"
)

var Widgets *DefaultWidgetStore

type DefaultWidgetStore struct {
	widgets map[int]*Widget
	sync.RWMutex
}

func NewDefaultWidgetStore() *DefaultWidgetStore {
	return &DefaultWidgetStore{widgets: make(map[int]*Widget)}
}

func (s *DefaultWidgetStore) Get(id int) (*Widget, error) {
	s.RLock()
	defer s.RUnlock()
	w, ok := s.widgets[id]
	if !ok {
		return w, sql.ErrNoRows
	}
	return w, nil
}

func (s *DefaultWidgetStore) set(w *Widget) {
	s.Lock()
	defer s.Unlock()
	s.widgets[w.ID] = w
}

func (s *DefaultWidgetStore) delete(id int) {
	s.Lock()
	defer s.Unlock()
	delete(s.widgets, id)
}
