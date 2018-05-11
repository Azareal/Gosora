/* Copyright Azareal 2017 - 2018 */
package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"strings"
	"sync"

	"../query_gen/lib"
)

var Docks WidgetDocks
var widgetUpdateMutex sync.RWMutex

type WidgetDocks struct {
	LeftOfNav    []*Widget
	RightOfNav   []*Widget
	LeftSidebar  []*Widget
	RightSidebar []*Widget
	//PanelLeft []Menus
	Footer []*Widget
}

type Widget struct {
	Enabled  bool
	Location string // Coming Soon: overview, topics, topic / topic_view, forums, forum, global
	Position int
	Body     string
	Side     string
	Type     string
	Literal  bool
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

type WidgetStmts struct {
	getWidgets *sql.Stmt
}

var widgetStmts WidgetStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		widgetStmts = WidgetStmts{
			getWidgets: acc.Select("widgets").Columns("position, side, type, active,  location, data").Orderby("position ASC").Prepare(),
		}
		return acc.FirstError()
	})
}

func preparseWidget(widget *Widget, wdata string) (err error) {
	prebuildWidget := func(name string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := Templates.ExecuteTemplate(&b, name+".html", data)
		return string(b.Bytes()), err
	}

	sbytes := []byte(wdata)
	switch widget.Type {
	case "simple":
		var tmp NameTextPair
		err = json.Unmarshal(sbytes, &tmp)
		if err != nil {
			return err
		}
		widget.Body, err = prebuildWidget("widget_simple", tmp)
	case "about":
		var tmp NameTextPair
		err = json.Unmarshal(sbytes, &tmp)
		if err != nil {
			return err
		}
		widget.Body, err = prebuildWidget("widget_about", tmp)
	default:
		widget.Body = wdata
	}
	widget.Literal = true

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
			err := mhold.Build(header.Writer, &header.CurrentUser)
			if err != nil {
				LogError(err)
			}
		}
		return ""
	case "rightSidebar":
		widgets = Docks.RightSidebar
	case "footer":
		widgets = Docks.Footer
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

// TODO: Test this
// TODO: Add support for zone:id. Perhaps, carry a ZoneID property around in *Header? It might allow some weirdness like frontend[5] which matches any zone with an ID of 5 but it would be a tad faster than verifying each zone, although it might be problematic if users end up relying on this behaviour for areas which don't pass IDs to the widgets system but *probably* should
func (widget *Widget) Allowed(zone string) bool {
	for _, loc := range strings.Split(widget.Location, "|") {
		if loc == "global" || loc == zone {
			return true
		} else if len(loc) > 0 && loc[0] == '!' {
			loc = loc[1:]
			if loc != "global" && loc != zone {
				return true
			}
		}
	}
	return false
}

// TODO: Refactor
func (widget *Widget) Build(hvars interface{}) (string, error) {
	if widget.Literal {
		return widget.Body, nil
	}

	var header = hvars.(*Header)
	err := RunThemeTemplate(header.Theme.Name, widget.Body, hvars, header.Writer)
	return "", err
}

// TODO: Make a store for this?
func InitWidgets() error {
	rows, err := widgetStmts.getWidgets.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var data string
	var leftOfNavWidgets []*Widget
	var rightOfNavWidgets []*Widget
	var leftSidebarWidgets []*Widget
	var rightSidebarWidgets []*Widget
	var footerWidgets []*Widget

	for rows.Next() {
		var widget = &Widget{Position: 0}
		err = rows.Scan(&widget.Position, &widget.Side, &widget.Type, &widget.Enabled, &widget.Location, &data)
		if err != nil {
			return err
		}

		err = preparseWidget(widget, data)
		if err != nil {
			return err
		}

		switch widget.Side {
		case "leftOfNav":
			leftOfNavWidgets = append(leftOfNavWidgets, widget)
		case "rightOfNav":
			rightOfNavWidgets = append(rightOfNavWidgets, widget)
		case "left":
			leftSidebarWidgets = append(leftSidebarWidgets, widget)
		case "right":
			rightSidebarWidgets = append(rightSidebarWidgets, widget)
		case "footer":
			footerWidgets = append(footerWidgets, widget)
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	// TODO: Let themes set default values for widget docks, and let them lock in particular places with their stuff, e.g. leftOfNav and rightOfNav

	widgetUpdateMutex.Lock()
	Docks.LeftOfNav = leftOfNavWidgets
	Docks.RightOfNav = rightOfNavWidgets
	Docks.LeftSidebar = leftSidebarWidgets
	Docks.RightSidebar = rightSidebarWidgets
	Docks.Footer = footerWidgets
	widgetUpdateMutex.Unlock()

	DebugLog("Docks.LeftOfNav", Docks.LeftOfNav)
	DebugLog("Docks.RightOfNav", Docks.RightOfNav)
	DebugLog("Docks.LeftSidebar", Docks.LeftSidebar)
	DebugLog("Docks.RightSidebar", Docks.RightSidebar)
	DebugLog("Docks.Footer", Docks.Footer)

	return nil
}
