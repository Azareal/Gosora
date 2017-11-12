/* Copyright Azareal 2017 - 2018 */
package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	"../query_gen/lib"
)

var Docks WidgetDocks
var widgetUpdateMutex sync.RWMutex

type WidgetDocks struct {
	LeftSidebar  []Widget
	RightSidebar []Widget
	//PanelLeft []Menus
}

type Widget struct {
	Enabled  bool
	Location string // Coming Soon: overview, topics, topic / topic_view, forums, forum, global
	Position int
	Body     string
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
	Text string
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

// TODO: Make a store for this?
func InitWidgets() error {
	rows, err := widgetStmts.getWidgets.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var sbytes []byte
	var side, wtype, data string

	var leftWidgets []Widget
	var rightWidgets []Widget

	for rows.Next() {
		var widget Widget
		err = rows.Scan(&widget.Position, &side, &wtype, &widget.Enabled, &widget.Location, &data)
		if err != nil {
			return err
		}

		sbytes = []byte(data)
		switch wtype {
		case "simple":
			var tmp NameTextPair
			err = json.Unmarshal(sbytes, &tmp)
			if err != nil {
				return err
			}

			var b bytes.Buffer
			err = Templates.ExecuteTemplate(&b, "widget_simple.html", tmp)
			if err != nil {
				return err
			}
			widget.Body = string(b.Bytes())
		default:
			widget.Body = data
		}

		if side == "left" {
			leftWidgets = append(leftWidgets, widget)
		} else if side == "right" {
			rightWidgets = append(rightWidgets, widget)
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	widgetUpdateMutex.Lock()
	Docks.LeftSidebar = leftWidgets
	Docks.RightSidebar = rightWidgets
	widgetUpdateMutex.Unlock()

	if Dev.SuperDebug {
		log.Print("Docks.LeftSidebar", Docks.LeftSidebar)
		log.Print("Docks.RightSidebar", Docks.RightSidebar)
	}

	return nil
}
