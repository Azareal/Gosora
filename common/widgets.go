/* Copyright Azareal 2017 - 2020 */
package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"sync/atomic"

	min "github.com/Azareal/Gosora/common/templates"
	"github.com/pkg/errors"
)

// TODO: Clean this file up
var Docks WidgetDocks
var widgetUpdateMutex sync.RWMutex

type WidgetDock struct {
	Items     []*Widget
	Scheduler *WidgetScheduler
}

type WidgetDocks struct {
	LeftOfNav    []*Widget
	RightOfNav   []*Widget
	LeftSidebar  WidgetDock
	RightSidebar WidgetDock
	//PanelLeft []Menus
	Footer WidgetDock
}

type WidgetMenu struct {
	Name     string
	MenuList []WidgetMenuItem
}

type WidgetMenuItem struct {
	Text     string
	Location string
	Compact  bool
}

type NameTextPair struct {
	Name string
	Text template.HTML
}

func preparseWidget(w *Widget, wdata string) (err error) {
	prebuildWidget := func(name string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := DefaultTemplates.ExecuteTemplate(&b, name+".html", data)
		content := string(b.Bytes())
		if Config.MinifyTemplates {
			content = min.Minify(content)
		}
		return content, err
	}

	sbytes := []byte(wdata)
	w.Literal = true
	// TODO: Split these hard-coded items out of this file and into the files for the individual widget types
	switch w.Type {
	case "simple", "about":
		var tmp NameTextPair
		err = json.Unmarshal(sbytes, &tmp)
		if err != nil {
			return err
		}
		w.Body, err = prebuildWidget("widget_"+w.Type, tmp)
	case "search_and_filter":
		w.Literal = false
		w.BuildFunc = widgetSearchAndFilter
	case "wol":
		w.Literal = false
		w.InitFunc = wolInit
		w.BuildFunc = wolRender
		w.TickFunc = wolTick
	case "wol_context":
		w.Literal = false
		w.BuildFunc = wolContextRender
	default:
		w.Body = wdata
	}

	// TODO: Test this
	// TODO: Should we toss this through a proper parser rather than crudely replacing it?
	w.Location = strings.Replace(w.Location, " ", "", -1)
	w.Location = strings.Replace(w.Location, "frontend", "!panel", -1)
	w.Location = strings.Replace(w.Location, "!!", "", -1)

	// Skip blank zones
	locs := strings.Split(w.Location, "|")
	if len(locs) > 0 {
		w.Location = ""
		for _, loc := range locs {
			if loc == "" {
				continue
			}
			w.Location += loc + "|"
		}
		w.Location = w.Location[:len(w.Location)-1]
	}

	return err
}

func GetDockList() []string {
	return []string{
		"leftOfNav",
		"rightOfNav",
		"rightSidebar",
		"footer",
	}
}

func GetDock(dock string) []*Widget {
	switch dock {
	case "leftOfNav":
		return Docks.LeftOfNav
	case "rightOfNav":
		return Docks.RightOfNav
	case "rightSidebar":
		return Docks.RightSidebar.Items
	case "footer":
		return Docks.Footer.Items
	}
	return nil
}

func HasDock(dock string) bool {
	switch dock {
	case "leftOfNav", "rightOfNav", "rightSidebar", "footer":
		return true
	}
	return false
}

// TODO: Find a more optimimal way of doing this...
func HasWidgets(dock string, h *Header) bool {
	if !h.Theme.HasDock(dock) {
		return false
	}

	// Let themes forcibly override this slot
	sbody := h.Theme.BuildDock(dock)
	if sbody != "" {
		return true
	}

	var widgets []*Widget
	switch dock {
	case "leftOfNav":
		widgets = Docks.LeftOfNav
	case "rightOfNav":
		widgets = Docks.RightOfNav
	case "rightSidebar":
		widgets = Docks.RightSidebar.Items
	case "footer":
		widgets = Docks.Footer.Items
	}

	wcount := 0
	for _, widget := range widgets {
		if !widget.Enabled {
			continue
		}
		if widget.Allowed(h.Zone, h.ZoneID) {
			wcount++
		}
	}
	return wcount > 0
}

func BuildWidget(dock string, h *Header) (sbody string) {
	if !h.Theme.HasDock(dock) {
		return ""
	}
	// Let themes forcibly override this slot
	sbody = h.Theme.BuildDock(dock)
	if sbody != "" {
		return sbody
	}

	var widgets []*Widget
	switch dock {
	case "leftOfNav":
		widgets = Docks.LeftOfNav
	case "rightOfNav":
		widgets = Docks.RightOfNav
	case "topMenu":
		// 1 = id for the default menu
		mhold, err := Menus.Get(1)
		if err == nil {
			err := mhold.Build(h.Writer, h.CurrentUser, h.Path)
			if err != nil {
				LogError(err)
			}
		}
		return ""
	case "rightSidebar":
		widgets = Docks.RightSidebar.Items
	case "footer":
		widgets = Docks.Footer.Items
	}

	for _, widget := range widgets {
		if !widget.Enabled {
			continue
		}
		if widget.Allowed(h.Zone, h.ZoneID) {
			item, err := widget.Build(h)
			if err != nil {
				LogError(err)
			}
			sbody += item
		}
	}
	return sbody
}

func getDockWidgets(dock string) (widgets []*Widget, err error) {
	rows, err := widgetStmts.getDockList.Query(dock)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		w := &Widget{Position: 0, Side: dock}
		err = rows.Scan(&w.ID, &w.Position, &w.Type, &w.Enabled, &w.Location, &w.RawBody)
		if err != nil {
			return nil, err
		}

		err = preparseWidget(w, w.RawBody)
		if err != nil {
			return nil, err
		}
		Widgets.set(w)
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// TODO: Make a store for this?
func InitWidgets() (fi error) {
	// TODO: Let themes set default values for widget docks, and let them lock in particular places with their stuff, e.g. leftOfNav and rightOfNav
	f := func(name string) {
		if fi != nil {
			return
		}
		dock, err := getDockWidgets(name)
		if err != nil {
			fi = err
			return
		}
		setDock(name, dock)
	}

	f("leftOfNav")
	f("rightOfNav")
	f("leftSidebar")
	f("rightSidebar")
	f("footer")
	if fi != nil {
		return fi
	}

	AddScheduledSecondTask(Docks.LeftSidebar.Scheduler.Tick)
	AddScheduledSecondTask(Docks.RightSidebar.Scheduler.Tick)
	AddScheduledSecondTask(Docks.Footer.Scheduler.Tick)

	return nil
}

func releaseWidgets(ws []*Widget) {
	for _, w := range ws {
		if w.ShutdownFunc != nil {
			w.ShutdownFunc(w)
		}
	}
}

// TODO: Use atomics
func setDock(dock string, widgets []*Widget) {
	dockHandle := func(dockWidgets []*Widget) {
		DebugLog(dock, widgets)
		releaseWidgets(dockWidgets)
	}
	dockHandle2 := func(dockWidgets WidgetDock) WidgetDock {
		dockHandle(dockWidgets.Items)
		if dockWidgets.Scheduler == nil {
			dockWidgets.Scheduler = &WidgetScheduler{}
		}
		for _, widget := range widgets {
			if widget.InitFunc != nil {
				widget.InitFunc(widget, dockWidgets.Scheduler)
			}
		}
		dockWidgets.Scheduler.Store()
		return WidgetDock{widgets, dockWidgets.Scheduler}
	}
	widgetUpdateMutex.Lock()
	defer widgetUpdateMutex.Unlock()
	switch dock {
	case "leftOfNav":
		dockHandle(Docks.LeftOfNav)
		Docks.LeftOfNav = widgets
	case "rightOfNav":
		dockHandle(Docks.RightOfNav)
		Docks.RightOfNav = widgets
	case "leftSidebar":
		Docks.LeftSidebar = dockHandle2(Docks.LeftSidebar)
	case "rightSidebar":
		Docks.RightSidebar = dockHandle2(Docks.RightSidebar)
	case "footer":
		Docks.Footer = dockHandle2(Docks.Footer)
	default:
		fmt.Printf("bad dock '%s'\n", dock)
		return
	}
}

type WidgetScheduler struct {
	widgets []*Widget
	store   atomic.Value
}

func (s *WidgetScheduler) Add(w *Widget) {
	s.widgets = append(s.widgets, w)
}

func (s *WidgetScheduler) Store() {
	s.store.Store(s.widgets)
}

func (s *WidgetScheduler) Tick() error {
	widgets := s.store.Load().([]*Widget)
	for _, widget := range widgets {
		if widget.TickFunc == nil {
			continue
		}
		err := widget.TickFunc(widget)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
