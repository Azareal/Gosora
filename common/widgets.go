/* Copyright Azareal 2017 - 2018 */
package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"strings"
	"sync"

	"../query_gen/lib"
)

var Docks WidgetDocks
var widgetUpdateMutex sync.RWMutex

type WidgetDocks struct {
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
		widget.Literal = true
		widget.Body, err = prebuildWidget("widget_simple", tmp)
	case "about":
		var tmp NameTextPair
		err = json.Unmarshal(sbytes, &tmp)
		if err != nil {
			return err
		}
		widget.Literal = true
		widget.Body, err = prebuildWidget("widget_about", tmp)
	default:
		widget.Literal = true
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

func BuildWidget(dock string, headerVars *HeaderVars) (sbody string) {
	var widgets []*Widget
	if !headerVars.Theme.HasDock(dock) {
		return ""
	}
	switch dock {
	case "rightSidebar":
		widgets = Docks.RightSidebar
	case "footer":
		widgets = Docks.Footer
	}

	for _, widget := range widgets {
		if !widget.Enabled {
			continue
		}
		if widget.Allowed(headerVars.Zone) {
			item, err := widget.Build(headerVars)
			if err != nil {
				LogError(err)
			}
			sbody += item
		}
	}
	return sbody
}

// TODO: Test this
// TODO: Add support for zone:id. Perhaps, carry a ZoneID property around in headerVars? It might allow some weirdness like frontend[5] which matches any zone with an ID of 5 but it would be a tad faster than verifying each zone, although it might be problematic if users end up relying on this behaviour for areas which don't pass IDs to the widgets system but *probably* should
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

	var b bytes.Buffer
	var headerVars = hvars.(*HeaderVars)
	err := RunThemeTemplate(headerVars.Theme.Name, widget.Body, hvars, headerVars.Writer)
	return string(b.Bytes()), err
}

// TODO: Make a store for this?
func InitWidgets() error {
	rows, err := widgetStmts.getWidgets.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var data string
	var leftWidgets []*Widget
	var rightWidgets []*Widget
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
		case "left":
			leftWidgets = append(leftWidgets, widget)
		case "right":
			rightWidgets = append(rightWidgets, widget)
		case "footer":
			footerWidgets = append(footerWidgets, widget)
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	widgetUpdateMutex.Lock()
	Docks.LeftSidebar = leftWidgets
	Docks.RightSidebar = rightWidgets
	Docks.Footer = footerWidgets
	widgetUpdateMutex.Unlock()

	if Dev.SuperDebug {
		log.Print("Docks.LeftSidebar", Docks.LeftSidebar)
		log.Print("Docks.RightSidebar", Docks.RightSidebar)
		log.Print("Docks.Footer", Docks.Footer)
	}

	return nil
}
