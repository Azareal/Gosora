/* Copyright Azareal 2017 - 2018 */
package main

import "fmt"
import "bytes"
import "sync"
import "encoding/json"
//import "html/template"

var docks WidgetDocks
var widget_update_mutex sync.RWMutex

type WidgetDocks struct
{
	LeftSidebar []Widget
	RightSidebar []Widget
	//PanelLeft []Menus
}

type Widget struct
{
	Enabled bool
	Location string // Coming Soon: overview, topics, topic / topic_view, forums, forum, global
	Position int
	Body string
}

type NameTextPair struct
{
	Name string
	Text string
}

func init_widgets() error {
	rows, err := get_widgets_stmt.Query()
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
		switch(wtype) {
		case "simple":
				var tmp NameTextPair
				err = json.Unmarshal(sbytes, &tmp)
				if err != nil {
					return err
				}

				var b bytes.Buffer
				err = templates.ExecuteTemplate(&b,"widget_simple.html",tmp)
				if err != nil {
					return err
				}
				widget.Body = string(b.Bytes())
			default:
				widget.Body = data
		}

		if side == "left" {
			leftWidgets = append(leftWidgets,widget)
		} else if side == "right" {
			rightWidgets = append(rightWidgets,widget)
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	widget_update_mutex.Lock()
	docks.LeftSidebar = leftWidgets
	docks.RightSidebar = rightWidgets
	widget_update_mutex.Unlock()

	if super_debug {
		fmt.Println("docks.LeftSidebar",docks.LeftSidebar)
		fmt.Println("docks.RightSidebar",docks.RightSidebar)
	}

	return nil
}
