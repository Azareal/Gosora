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

func (w *DefaultWidgetStore) Get(id int) (*Widget, error) {
	w.RLock()
	defer w.RUnlock()
	widget, ok := w.widgets[id]
	if !ok {
		return widget, sql.ErrNoRows
	}
	return widget, nil
}

func (w *DefaultWidgetStore) set(widget *Widget) {
	w.Lock()
	defer w.Unlock()
	w.widgets[widget.ID] = widget
}

func (w *DefaultWidgetStore) delete(id int) {
	w.Lock()
	defer w.Unlock()
	delete(w.widgets, id)
}
