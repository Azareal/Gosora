package common

import (
	"database/sql"
	"log"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Analytics AnalyticsStore

type AnalyticsTimeRange struct {
	Quantity   int
	Unit       string
	Slices     int
	SliceWidth int
	Range      string
}

type AnalyticsStore interface {
	FillViewMap(tbl string, tr *AnalyticsTimeRange, labelList []int64, viewMap map[int64]int64, param string, args ...interface{}) (map[int64]int64, error)
}

type DefaultAnalytics struct {
}

func NewDefaultAnalytics() *DefaultAnalytics {
	return &DefaultAnalytics{}
}

/*
	rows, e := qgen.NewAcc().Select("viewchunks_systems").Columns("count,createdAt").Where("system=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(system)
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	viewMap, e = c.AnalyticsRowsToViewMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
*/

func (s *DefaultAnalytics) FillViewMap(tbl string, tr *AnalyticsTimeRange, labelList []int64, viewMap map[int64]int64, param string, args ...interface{}) (map[int64]int64, error) {
	ac := qgen.NewAcc().Select(tbl).Columns("count,createdAt")
	if param != "" {
		ac = ac.Where(param + "=?")
	}
	rows, e := ac.DateCutoff("createdAt", tr.Quantity, tr.Unit).Query(args...)
	if e != nil && e != sql.ErrNoRows {
		return nil, e
	}
	return AnalyticsRowsToViewMap(rows, labelList, viewMap)
}

// TODO: Clamp it rather than using an offset off the current time to avoid chaotic changes in stats as adjacent sets converge and diverge?
func AnalyticsTimeRangeToLabelList(tr *AnalyticsTimeRange) (revLabelList []int64, labelList []int64, viewMap map[int64]int64) {
	viewMap = make(map[int64]int64)
	currentTime := time.Now().Unix()
	for i := 1; i <= tr.Slices; i++ {
		label := currentTime - int64(i*tr.SliceWidth)
		revLabelList = append(revLabelList, label)
		viewMap[label] = 0
	}
	labelList = append(labelList, revLabelList...)
	return revLabelList, labelList, viewMap
}

func AnalyticsRowsToViewMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		e := rows.Scan(&count, &createdAt)
		if e != nil {
			return viewMap, e
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}
		for _, value := range labelList {
			if unixCreatedAt > value {
				viewMap[value] += count
				break
			}
		}
	}
	return viewMap, rows.Err()
}
