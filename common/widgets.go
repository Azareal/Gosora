/* Copyright Azareal 2017 - 2019 */
package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"sync/atomic"
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

func preparseWidget(widget *Widget, wdata string) (err error) {
	prebuildWidget := func(name string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := Templates.ExecuteTemplate(&b, name+".html", data)
		return string(b.Bytes()), err
	}

	sbytes := []byte(wdata)
	widget.Literal = true
	// TODO: Split these hard-coded items out of this file and into the files for the individual widget types
	switch widget.Type {
	case "simple", "about":
		var tmp NameTextPair
		err = json.Unmarshal(sbytes, &tmp)
		if err != nil {
			return err
		}
		widget.Body, err = prebuildWidget("widget_"+widget.Type, tmp)
	case "search_and_filter":
		widget.Literal = false
		widget.BuildFunc = widgetSearchAndFilter
	case "wol":
		widget.Literal = false
		widget.InitFunc = wolInit
		widget.BuildFunc = wolRender
		widget.TickFunc = wolTick
	case "wol_context":
		widget.Literal = false
		widget.BuildFunc = wolContextRender
	default:
		widget.Body = wdata
	}

	// TODO: Test this
	// TODO: Should we toss this through a proper parser rather than crudely replacing it?
	widget.Location = strings.Replace(widget.Location, " ", "", -1)
	widget.Location = strings.Replace(widget.Location, "frontend", "!panel", -1)
	widget.Location = strings.Replace(widget.Location, "!!", "", -1)

	// Skip blank zones
	var locs = strings.Split(widget.Location, "|")
	if len(locs) > 0 {
		widget.Location = ""
		for _, loc := range locs {
			if loc == "" {
				continue
			}
			widget.Location += loc + "|"
		}
		widget.Location = widget.Location[:len(widget.Location)-1]
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

func BuildWidget(dock string, header *Header) (sbody string) {
	var widgets []*Widget
	if !header.Theme.HasDock(dock) {
		return ""
	}

	// Let themes forcibly override this slot
	sbody = header.Theme.BuildDock(dock)
	if sbody != "" {
		return sbody
	}

	switch dock {
	case "leftOfNav":
		widgets = Docks.LeftOfNav
	case "rightOfNav":
		widgets = Docks.RightOfNav
	case "topMenu":
		// 1 = id for the default menu
		mhold, err := Menus.Get(1)
		if err == nil {
			err := mhold.Build(header.Writer, &header.CurrentUser, header.Path)
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
		if widget.Allowed(header.Zone) {
			item, err := widget.Build(header)
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
		var widget = &Widget{Position: 0, Side: dock}
		err = rows.Scan(&widget.ID, &widget.Position, &widget.Type, &widget.Enabled, &widget.Location, &widget.RawBody)
		if err != nil {
			return nil, err
		}

		err = preparseWidget(widget, widget.RawBody)
		if err != nil {
			return nil, err
		}
		Widgets.set(widget)
		widgets = append(widgets, widget)
	}
	return widgets, rows.Err()
}

// TODO: Make a store for this?
func InitWidgets() error {
	leftOfNavWidgets, err := getDockWidgets("leftOfNav")
	if err != nil {
		return err
	}
	rightOfNavWidgets, err := getDockWidgets("rightOfNav")
	if err != nil {
		return err
	}
	leftSidebarWidgets, err := getDockWidgets("leftSidebar")
	if err != nil {
		return err
	}
	rightSidebarWidgets, err := getDockWidgets("rightSidebar")
	if err != nil {
		return err
	}
	footerWidgets, err := getDockWidgets("footer")
	if err != nil {
		return err
	}

	// TODO: Let themes set default values for widget docks, and let them lock in particular places with their stuff, e.g. leftOfNav and rightOfNav

	setDock("leftOfNav", leftOfNavWidgets)
	setDock("rightOfNav", rightOfNavWidgets)
	setDock("leftSidebar", leftSidebarWidgets)
	setDock("rightSidebar", rightSidebarWidgets)
	setDock("footer", footerWidgets)
	AddScheduledSecondTask(Docks.LeftSidebar.Scheduler.Tick)
	AddScheduledSecondTask(Docks.RightSidebar.Scheduler.Tick)
	AddScheduledSecondTask(Docks.Footer.Scheduler.Tick)

	return nil
}

func releaseWidgets(widgets []*Widget) {
	for _, widget := range widgets {
		if widget.ShutdownFunc != nil {
			widget.ShutdownFunc(widget)
		}
	}
}

// TODO: Use atomics
func setDock(dock string, widgets []*Widget) {
	var dockHandle = func(dockWidgets []*Widget) {
		widgetUpdateMutex.Lock()
		DebugLog(dock, widgets)
		releaseWidgets(dockWidgets)
	}
	var dockHandle2 = func(dockWidgets WidgetDock) WidgetDock {
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
	widgetUpdateMutex.Unlock()
}

type WidgetScheduler struct {
	widgets []*Widget
	store   atomic.Value
}

func (schedule *WidgetScheduler) Add(widget *Widget) {
	schedule.widgets = append(schedule.widgets, widget)
}

func (schedule *WidgetScheduler) Store() {
	schedule.store.Store(schedule.widgets)
}

func (schedule *WidgetScheduler) Tick() error {
	widgets := schedule.store.Load().([]*Widget)
	for _, widget := range widgets {
		if widget.TickFunc == nil {
			continue
		}
		err := widget.TickFunc(widget.Copy())
		if err != nil {
			return err
		}
	}
	return nil
}
