package common

import (
	"database/sql"
	"encoding/json"
	"strings"
	"strconv"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
)

type WidgetStmts struct {
	//getList *sql.Stmt
	getDockList *sql.Stmt
	delete      *sql.Stmt
	create      *sql.Stmt
	update      *sql.Stmt
}

var widgetStmts WidgetStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		widgetStmts = WidgetStmts{
			//getList: acc.Select("widgets").Columns("wid, position, side, type, active,  location, data").Orderby("position ASC").Prepare(),
			getDockList: acc.Select("widgets").Columns("wid, position, type, active,  location, data").Where("side = ?").Orderby("position ASC").Prepare(),
			delete:      acc.Delete("widgets").Where("wid = ?").Prepare(),
			create:      acc.Insert("widgets").Columns("position, side, type, active, location, data").Fields("?,?,?,?,?,?").Prepare(),
			update:      acc.Update("widgets").Set("position = ?, side = ?, type = ?, active = ?, location = ?, data = ?").Where("wid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Shrink this struct for common uses in the templates? Would that really make things go faster?
type Widget struct {
	ID       int
	Enabled  bool
	Location string // Coming Soon: overview, topics, topic / topic_view, forums, forum, global
	Position int
	RawBody  string
	Body     string
	Side     string
	Type     string

	Literal      bool
	TickMask     atomic.Value
	InitFunc     func(widget *Widget, schedule *WidgetScheduler) error
	ShutdownFunc func(widget *Widget) error
	BuildFunc    func(widget *Widget, hvars interface{}) (string, error)
	TickFunc     func(widget *Widget) error
}

func (widget *Widget) Delete() error {
	_, err := widgetStmts.delete.Exec(widget.ID)
	if err != nil {
		return err
	}

	// Reload the dock
	// TODO: Better synchronisation
	Widgets.delete(widget.ID)
	widgets, err := getDockWidgets(widget.Side)
	if err != nil {
		return err
	}
	setDock(widget.Side, widgets)
	return nil
}

func (widget *Widget) Copy() (owidget *Widget) {
	owidget = &Widget{}
	*owidget = *widget
	return owidget
}

// TODO: Test this
// TODO: Add support for zone:id. Perhaps, carry a ZoneID property around in *Header? It might allow some weirdness like frontend[5] which matches any zone with an ID of 5 but it would be a tad faster than verifying each zone, although it might be problematic if users end up relying on this behaviour for areas which don't pass IDs to the widgets system but *probably* should
// TODO: Add a selector which also matches topics inside a specific forum?
func (widget *Widget) Allowed(zone string, zoneid int) bool {
	for _, loc := range strings.Split(widget.Location, "|") {
		if len(loc) == 0 {
			continue
		}
		sloc := strings.Split(":",loc)
		if len(sloc) > 1 {
			iloc, _ := strconv.Atoi(sloc[1])
			if zoneid != 0 && iloc != zoneid {
				continue
			}
		}
		if loc == "global" || loc == zone {
			return true
		} else if loc[0] == '!' {
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
	if widget.BuildFunc != nil {
		return widget.BuildFunc(widget, hvars)
	}

	var header = hvars.(*Header)
	err := header.Theme.RunTmpl(widget.Body, hvars, header.Writer)
	return "", err
}

type WidgetEdit struct {
	*Widget
	Data map[string]string
}

func (widget *WidgetEdit) Create() error {
	data, err := json.Marshal(widget.Data)
	if err != nil {
		return err
	}
	_, err = widgetStmts.create.Exec(widget.Position, widget.Side, widget.Type, widget.Enabled, widget.Location, data)
	if err != nil {
		return err
	}

	// Reload the dock
	widgets, err := getDockWidgets(widget.Side)
	if err != nil {
		return err
	}
	setDock(widget.Side, widgets)
	return nil
}

func (widget *WidgetEdit) Commit() error {
	data, err := json.Marshal(widget.Data)
	if err != nil {
		return err
	}
	_, err = widgetStmts.update.Exec(widget.Position, widget.Side, widget.Type, widget.Enabled, widget.Location, data, widget.ID)
	if err != nil {
		return err
	}

	// Reload the dock
	widgets, err := getDockWidgets(widget.Side)
	if err != nil {
		return err
	}
	setDock(widget.Side, widgets)
	return nil
}
