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

func (widgets *DefaultWidgetStore) Get(id int) (*Widget, error) {
	widgets.RLock()
	defer widgets.RUnlock()
	widget, ok := widgets.widgets[id]
	if !ok {
		return widget, sql.ErrNoRows
	}
	return widget, nil
}

func (widgets *DefaultWidgetStore) set(widget *Widget) {
	widgets.Lock()
	defer widgets.Unlock()
	widgets.widgets[widget.ID] = widget
}

func (widgets *DefaultWidgetStore) delete(id int) {
	widgets.Lock()
	defer widgets.Unlock()
	delete(widgets.widgets, id)
}
