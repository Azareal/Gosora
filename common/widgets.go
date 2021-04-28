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
	"github.com/Azareal/Gosora/uutils"
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

func preparseWidget(w *Widget, wdata string) (e error) {
	prebuildWidget := func(name string, data interface{}) (string, error) {
		var b bytes.Buffer
		e := DefaultTemplates.ExecuteTemplate(&b, name+".html", data)
		content := b.String()
		if Config.MinifyTemplates {
			content = min.Minify(content)
		}
		return content, e
	}

	sbytes := []byte(wdata)
	w.Literal = true
	// TODO: Split these hard-coded items out of this file and into the files for the individual widget types
	switch w.Type {
	case "simple", "about":
		var tmp NameTextPair
		e = json.Unmarshal(sbytes, &tmp)
		if e != nil {
			return e
		}
		w.Body, e = prebuildWidget("widget_"+w.Type, tmp)
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
	rep := func(from, to string) {
		w.Location = strings.Replace(w.Location, from, to, -1)
	}
	rep(" ", "")
	rep("frontend", "!panel")
	rep("!!", "")

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

	return e
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
		mhold, e := Menus.Get(1)
		if e == nil {
			e := mhold.Build(h.Writer, h.CurrentUser, h.Path)
			if e != nil {
				LogError(e)
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
			item, e := widget.Build(h)
			if e != nil {
				LogError(e)
			}
			sbody += item
		}
	}
	return sbody
}

var DockToID = map[string]int{
	"leftOfNav":    0,
	"rightOfNav":   1,
	"topMenu":      2,
	"rightSidebar": 3,
	"footer":       4,
}

func BuildWidget2(dock int, h *Header) (sbody string) {
	if !h.Theme.HasDockByID(dock) {
		return ""
	}
	// Let themes forcibly override this slot
	sbody = h.Theme.BuildDockByID(dock)
	if sbody != "" {
		return sbody
	}

	var widgets []*Widget
	switch dock {
	case 0:
		widgets = Docks.LeftOfNav
	case 1:
		widgets = Docks.RightOfNav
	case 2:
		// 1 = id for the default menu
		mhold, e := Menus.Get(1)
		if e == nil {
			e := mhold.Build(h.Writer, h.CurrentUser, h.Path)
			if e != nil {
				LogError(e)
			}
		}
		return ""
	case 3:
		widgets = Docks.RightSidebar.Items
	case 4:
		widgets = Docks.Footer.Items
	}

	for _, w := range widgets {
		if !w.Enabled {
			continue
		}
		if w.Allowed(h.Zone, h.ZoneID) {
			item, e := w.Build(h)
			if e != nil {
				LogError(e)
			}
			sbody += item
		}
	}
	return sbody
}

func BuildWidget3(dock int, h *Header) {
	if !h.Theme.HasDockByID(dock) {
		return
	}
	// Let themes forcibly override this slot
	if sbody := h.Theme.BuildDockByID(dock); sbody != "" {
		h.Writer.Write(uutils.StringToBytes(sbody))
		return
	}

	var widgets []*Widget
	switch dock {
	case 0:
		widgets = Docks.LeftOfNav
	case 1:
		widgets = Docks.RightOfNav
	case 2:
		// 1 = id for the default menu
		mhold, err := Menus.Get(1)
		if err == nil {
			err := mhold.Build(h.Writer, h.CurrentUser, h.Path)
			if err != nil {
				LogError(err)
			}
		}
		return
	case 3:
		widgets = Docks.RightSidebar.Items
	case 4:
		widgets = Docks.Footer.Items
	}

	for _, w := range widgets {
		if !w.Enabled {
			continue
		}
		if w.Allowed(h.Zone, h.ZoneID) {
			item, e := w.Build(h)
			if e != nil {
				LogError(e)
			}
			if item != "" {
				h.Writer.Write(uutils.StringToBytes(item))
			}
		}
	}
}

// TODO: Find a more optimimal way of doing this...
func HasWidgets2(dock int, h *Header) bool {
	if !h.Theme.HasDockByID(dock) {
		return false
	}

	// Let themes forcibly override this slot
	// TODO: Optimise this bit
	sbody := h.Theme.BuildDockByID(dock)
	if sbody != "" {
		return true
	}

	var widgets []*Widget
	switch dock {
	case 0:
		widgets = Docks.LeftOfNav
	case 1:
		widgets = Docks.RightOfNav
	case 3:
		widgets = Docks.RightSidebar.Items
	case 4:
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

func getDockWidgets(dock string) (widgets []*Widget, e error) {
	rows, e := widgetStmts.getDockList.Query(dock)
	if e != nil {
		return nil, e
	}
	defer rows.Close()

	for rows.Next() {
		w := &Widget{Position: 0, Side: dock}
		e = rows.Scan(&w.ID, &w.Position, &w.Type, &w.Enabled, &w.Location, &w.RawBody)
		if e != nil {
			return nil, e
		}
		e = preparseWidget(w, w.RawBody)
		if e != nil {
			return nil, e
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
		dock, e := getDockWidgets(name)
		if e != nil {
			fi = e
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
		e := widget.TickFunc(widget)
		if e != nil {
			return errors.WithStack(e)
		}
	}
	return nil
}
