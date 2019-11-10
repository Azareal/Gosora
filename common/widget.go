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
	
	//qgen.SimpleModel
}

var widgetStmts WidgetStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		w := "widgets"
		widgetStmts = WidgetStmts{
			//getList: acc.Select(w).Columns("wid, position, side, type, active, location, data").Orderby("position ASC").Prepare(),
			getDockList: acc.Select(w).Columns("wid, position, type, active, location, data").Where("side = ?").Orderby("position ASC").Prepare(),
			//model: acc.SimpleModel(w,"position,type,active,location,data","wid"),
			delete:      acc.Delete(w).Where("wid = ?").Prepare(),
			create:      acc.Insert(w).Columns("position, side, type, active, location, data").Fields("?,?,?,?,?,?").Prepare(),
			update:      acc.Update(w).Set("position = ?, side = ?, type = ?, active = ?, location = ?, data = ?").Where("wid = ?").Prepare(),
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
	InitFunc     func(w *Widget, schedule *WidgetScheduler) error
	ShutdownFunc func(w *Widget) error
	BuildFunc    func(w *Widget, hvars interface{}) (string, error)
	TickFunc     func(w *Widget) error
}

func (w *Widget) Delete() error {
	_, err := widgetStmts.delete.Exec(w.ID)
	if err != nil {
		return err
	}

	// Reload the dock
	// TODO: Better synchronisation
	Widgets.delete(w.ID)
	widgets, err := getDockWidgets(w.Side)
	if err != nil {
		return err
	}
	setDock(w.Side, widgets)
	return nil
}

func (w *Widget) Copy() (owidget *Widget) {
	owidget = &Widget{}
	*owidget = *w
	return owidget
}

// TODO: Test this
// TODO: Add support for zone:id. Perhaps, carry a ZoneID property around in *Header? It might allow some weirdness like frontend[5] which matches any zone with an ID of 5 but it would be a tad faster than verifying each zone, although it might be problematic if users end up relying on this behaviour for areas which don't pass IDs to the widgets system but *probably* should
// TODO: Add a selector which also matches topics inside a specific forum?
func (w *Widget) Allowed(zone string, zoneid int) bool {
	for _, loc := range strings.Split(w.Location, "|") {
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
func (w *Widget) Build(hvars interface{}) (string, error) {
	if w.Literal {
		return w.Body, nil
	}
	if w.BuildFunc != nil {
		return w.BuildFunc(w, hvars)
	}
	header := hvars.(*Header)
	err := header.Theme.RunTmpl(w.Body, hvars, header.Writer)
	return "", err
}

type WidgetEdit struct {
	*Widget
	Data map[string]string
}

func (w *WidgetEdit) Create() (int, error) {
	data, err := json.Marshal(w.Data)
	if err != nil {
		return 0, err
	}
	res, err := widgetStmts.create.Exec(w.Position, w.Side, w.Type, w.Enabled, w.Location, data)
	if err != nil {
		return 0, err
	}

	// Reload the dock
	widgets, err := getDockWidgets(w.Side)
	if err != nil {
		return 0, err
	}
	setDock(w.Side, widgets)

	wid64, err := res.LastInsertId()
	return int(wid64), err
}

func (w *WidgetEdit) Commit() error {
	data, err := json.Marshal(w.Data)
	if err != nil {
		return err
	}
	_, err = widgetStmts.update.Exec(w.Position, w.Side, w.Type, w.Enabled, w.Location, data, w.ID)
	if err != nil {
		return err
	}

	// Reload the dock
	widgets, err := getDockWidgets(w.Side)
	if err != nil {
		return err
	}
	setDock(w.Side, widgets)
	return nil
}
