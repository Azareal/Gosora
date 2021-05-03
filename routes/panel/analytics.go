package panel

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

func analyticsTimeRange(rawTimeRange string) (*c.AnalyticsTimeRange, error) {
	tr := &c.AnalyticsTimeRange{
		Quantity:   6,
		Unit:       "hour",
		Slices:     12,
		SliceWidth: 60 * 30,
		Range:      "six-hours",
	}

	switch rawTimeRange {
	// This might be pushing it, we might want to come up with a more efficient scheme for dealing with large timeframes like this
	case "one-year":
		tr.Quantity = 12
		tr.Unit = "month"
		tr.Slices = 12
		tr.SliceWidth = 60 * 60 * 24 * 30
	case "three-months":
		tr.Quantity = 90
		tr.Unit = "day"
		tr.Slices = 30
		tr.SliceWidth = 60 * 60 * 24 * 3
	case "one-month":
		tr.Quantity = 30
		tr.Unit = "day"
		tr.Slices = 30
		tr.SliceWidth = 60 * 60 * 24
	case "one-week":
		tr.Quantity = 7
		tr.Unit = "day"
		tr.Slices = 14
		tr.SliceWidth = 60 * 60 * 12
	case "two-days": // Two days is experimental
		tr.Quantity = 2
		tr.Unit = "day"
		tr.Slices = 24
		tr.SliceWidth = 60 * 60 * 2
	case "one-day":
		tr.Quantity = 1
		tr.Unit = "day"
		tr.Slices = 24
		tr.SliceWidth = 60 * 60
	case "twelve-hours":
		tr.Quantity = 12
		tr.Slices = 24
	case "six-hours", "":
		return tr, nil
	default:
		return tr, errors.New("Unknown time range")
	}
	tr.Range = rawTimeRange
	return tr, nil
}

type pAvg struct {
	Avg int64
	Tot int64
}

func analyticsRowsToAverageMap(rows *sql.Rows, labelList []int64, avgMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		e := rows.Scan(&count, &createdAt)
		if e != nil {
			return avgMap, e
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}
		pAvgMap := make(map[int64]pAvg)
		for _, value := range labelList {
			if unixCreatedAt > value {
				prev := pAvgMap[value]
				prev.Avg += count
				prev.Tot++
				pAvgMap[value] = prev
				break
			}
		}
		for key, pAvg := range pAvgMap {
			avgMap[key] = pAvg.Avg / pAvg.Tot
		}
	}
	return avgMap, rows.Err()
}

func analyticsRowsToAverageMap2(rows *sql.Rows, labelList []int64, avgMap map[int64]int64, typ int) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var stack, heap int64
		var createdAt time.Time
		e := rows.Scan(&stack, &heap, &createdAt)
		if e != nil {
			return avgMap, e
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("stack: ", stack)
			log.Print("heap: ", heap)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}
		if typ == 1 {
			heap = 0
		} else if typ == 2 {
			stack = 0
		}
		pAvgMap := make(map[int64]pAvg)
		for _, value := range labelList {
			if unixCreatedAt > value {
				prev := pAvgMap[value]
				prev.Avg += stack + heap
				prev.Tot++
				pAvgMap[value] = prev
				break
			}
		}
		for key, pAvg := range pAvgMap {
			avgMap[key] = pAvg.Avg / pAvg.Tot
		}
	}
	return avgMap, rows.Err()
}

func analyticsRowsToAverageMap3(rows *sql.Rows, labelList []int64, avgMap map[int64]int64, typ int) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var low, high, avg int64
		var createdAt time.Time
		e := rows.Scan(&low, &high, &avg, &createdAt)
		if e != nil {
			return avgMap, e
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("low: ", low)
			log.Print("high: ", high)
			log.Print("avg: ", avg)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}
		var dat int64
		switch typ {
		case 0:
			dat = low
		case 1:
			dat = high
		default:
			dat = avg
		}
		pAvgMap := make(map[int64]pAvg)
		for _, val := range labelList {
			if unixCreatedAt > val {
				prev := pAvgMap[val]
				prev.Avg += dat
				prev.Tot++
				pAvgMap[val] = prev
				break
			}
		}
		for key, pAvg := range pAvgMap {
			avgMap[key] = pAvg.Avg / pAvg.Tot
		}
	}
	return avgMap, rows.Err()
}

func PreAnalyticsDetail(w http.ResponseWriter, r *http.Request, u *c.User) (*c.BasePanelPage, c.RouteError) {
	bp, fe := buildBasePage(w, r, u, "analytics", "analytics")
	if fe != nil {
		return nil, fe
	}
	bp.AddSheet("chartist/chartist.min.css")
	bp.AddScript("chartist/chartist.min.js")
	bp.AddScriptAsync("analytics.js")
	bp.LooseCSP = true
	return bp, nil
}

func createTimeGraph(series [][]int64, labelList []int64, legends ...[]string) c.PanelTimeGraph {
	var llegends []string
	if len(legends) > 0 {
		llegends = legends[0]
	}
	graph := c.PanelTimeGraph{Series: series, Labels: labelList, Legends: llegends}
	c.DebugLogf("graph: %+v\n", graph)
	return graph
}

func CreateViewListItems(revLabelList []int64, viewMap map[int64]int64) ([]int64, []c.PanelAnalyticsItem) {
	viewList := make([]int64, len(revLabelList))
	viewItems := make([]c.PanelAnalyticsItem, len(revLabelList))
	for i, val := range revLabelList {
		viewList[i] = viewMap[val]
		viewItems[i] = c.PanelAnalyticsItem{Time: val, Count: viewMap[val]}
	}
	return viewList, viewItems
}

func AnalyticsViews(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, fe := PreAnalyticsDetail(w, r, u)
	if fe != nil {
		return fe
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsViews")
	// TODO: Add some sort of analytics store / iterator?
	viewMap, e = c.Analytics.FillViewMap("viewchunks", tr, labelList, viewMap, "route", "")
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList, viewItems := CreateViewListItems(revLabelList, viewMap)

	graph := createTimeGraph([][]int64{viewList}, labelList)
	var ttime string
	if tr.Range == "six-hours" || tr.Range == "twelve-hours" || tr.Range == "one-day" {
		ttime = "time"
	}

	pi := c.PanelAnalyticsStd{graph, viewItems, tr.Range, tr.Unit, ttime}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_views", pi})
}

func AnalyticsRouteViews(w http.ResponseWriter, r *http.Request, u *c.User, route string) c.RouteError {
	bp, fe := PreAnalyticsDetail(w, r, u)
	if fe != nil {
		return fe
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsRouteViews")
	// TODO: Validate the route is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks", tr, labelList, viewMap, "route", route)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList, viewItems := CreateViewListItems(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	pi := c.PanelAnalyticsRoutePage{bp, c.SanitiseSingleLine(route), graph, viewItems, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_route_views", pi})
}

func AnalyticsAgentViews(w http.ResponseWriter, r *http.Request, u *c.User, agent string) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)
	// ? Only allow valid agents? The problem with this is that agents wind up getting renamed and it would take a migration to get them all up to snuff
	agent = c.SanitiseSingleLine(agent)

	c.DebugLog("in panel.AnalyticsAgentViews")
	// TODO: Verify the agent is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks_agents", tr, labelList, viewMap, "browser", agent)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList := CreateViewList(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	friendlyAgent, ok := p.GetUserAgentPhrase(agent)
	if !ok {
		friendlyAgent = agent
	}

	pi := c.PanelAnalyticsAgentPage{bp, agent, friendlyAgent, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_agent_views", pi})
}

func AnalyticsForumViews(w http.ResponseWriter, r *http.Request, u *c.User, sfid string) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	fid, e := strconv.Atoi(sfid)
	if e != nil {
		return c.LocalError("Invalid integer", w, r, u)
	}

	c.DebugLog("in panel.AnalyticsForumViews")
	// TODO: Verify the agent is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks_forums", tr, labelList, viewMap, "forum", fid)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList := CreateViewList(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	forum, e := c.Forums.Get(fid)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	pi := c.PanelAnalyticsAgentPage{bp, sfid, forum.Name, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_forum_views", pi})
}

func AnalyticsSystemViews(w http.ResponseWriter, r *http.Request, u *c.User, system string) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)
	system = c.SanitiseSingleLine(system)

	c.DebugLog("in panel.AnalyticsSystemViews")
	// TODO: Verify the OS name is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks_systems", tr, labelList, viewMap, "system", system)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList := CreateViewList(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	friendlySystem, ok := p.GetOSPhrase(system)
	if !ok {
		friendlySystem = system
	}

	pi := c.PanelAnalyticsAgentPage{bp, system, friendlySystem, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_system_views", pi})
}

func CreateViewList(revLabelList []int64, viewMap map[int64]int64) []int64 {
	viewList := make([]int64, len(revLabelList))
	for i, val := range revLabelList {
		viewList[i] = viewMap[val]
	}
	return viewList
}

func AnalyticsLanguageViews(w http.ResponseWriter, r *http.Request, u *c.User, lang string) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)
	lang = c.SanitiseSingleLine(lang)

	c.DebugLog("in panel.AnalyticsLanguageViews")
	// TODO: Verify the language code is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks_langs", tr, labelList, viewMap, "lang", lang)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList := CreateViewList(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	friendlyLang, ok := p.GetHumanLangPhrase(lang)
	if !ok {
		friendlyLang = lang
	}

	pi := c.PanelAnalyticsAgentPage{bp, lang, friendlyLang, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_lang_views", pi})
}

func AnalyticsReferrerViews(w http.ResponseWriter, r *http.Request, u *c.User, domain string) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsReferrerViews")
	// TODO: Verify the agent is valid
	viewMap, e = c.Analytics.FillViewMap("viewchunks_referrers", tr, labelList, viewMap, "domain", domain)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList := CreateViewList(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	pi := c.PanelAnalyticsAgentPage{bp, c.SanitiseSingleLine(domain), "", graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_referrer_views", pi})
}

func AnalyticsTopics(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsTopics")
	viewMap, e = c.Analytics.FillViewMap("topicchunks", tr, labelList, viewMap, "")
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList, viewItems := CreateViewListItems(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	pi := c.PanelAnalyticsStd{graph, viewItems, tr.Range, tr.Unit, "time"}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_topics", pi})
}

func AnalyticsPosts(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, fe := PreAnalyticsDetail(w, r, u)
	if fe != nil {
		return fe
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsPosts")
	viewMap, e = c.Analytics.FillViewMap("postchunks", tr, labelList, viewMap, "")
	if e != nil {
		return c.InternalError(e, w, r)
	}
	viewList, viewItems := CreateViewListItems(revLabelList, viewMap)
	graph := createTimeGraph([][]int64{viewList}, labelList)

	pi := c.PanelAnalyticsStd{graph, viewItems, tr.Range, tr.Unit, "time"}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_posts", pi})
}

func AnalyticsMemory(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, fe := PreAnalyticsDetail(w, r, u)
	if fe != nil {
		return fe
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, avgMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsMemory")
	rows, e := qgen.NewAcc().Select("memchunks").Columns("count,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	avgMap, e = analyticsRowsToAverageMap(rows, labelList, avgMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	avgList := make([]int64, len(revLabelList))
	avgItems := make([]c.PanelAnalyticsItemUnit, len(revLabelList))
	for i, value := range revLabelList {
		avgList[i] = avgMap[value]
		cv, cu := c.ConvertByteUnit(float64(avgMap[value]))
		avgItems[i] = c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)}
	}
	graph := createTimeGraph([][]int64{avgList}, labelList)

	pi := c.PanelAnalyticsStdUnit{graph, avgItems, tr.Range, tr.Unit, "time"}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_memory", pi})
}

// TODO: Show stack and heap memory separately on the chart
func AnalyticsActiveMemory(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, avgMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsActiveMemory")
	rows, e := qgen.NewAcc().Select("memchunks").Columns("stack,heap,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}

	var typ int
	switch r.FormValue("mtype") {
	case "1":
		typ = 1
	case "2":
		typ = 2
	default:
		typ = 0
	}
	avgMap, e = analyticsRowsToAverageMap2(rows, labelList, avgMap, typ)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	avgList := make([]int64, len(revLabelList))
	avgItems := make([]c.PanelAnalyticsItemUnit, len(revLabelList))
	for i, value := range revLabelList {
		avgList[i] = avgMap[value]
		cv, cu := c.ConvertByteUnit(float64(avgMap[value]))
		avgItems[i] = c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)}
	}
	graph := createTimeGraph([][]int64{avgList}, labelList)

	pi := c.PanelAnalyticsActiveMemory{graph, avgItems, tr.Range, tr.Unit, "time", typ}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_active_memory", pi})
}

func AnalyticsPerf(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, avgMap := c.AnalyticsTimeRangeToLabelList(tr)

	c.DebugLog("in panel.AnalyticsPerf")
	rows, e := qgen.NewAcc().Select("perfchunks").Columns("low,high,avg,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}

	var typ int
	switch r.FormValue("type") {
	case "0":
		typ = 0
	case "1":
		typ = 1
	default:
		typ = 2
	}
	avgMap, e = analyticsRowsToAverageMap3(rows, labelList, avgMap, typ)
	if e != nil {
		return c.InternalError(e, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	avgList := make([]int64, len(revLabelList))
	avgItems := make([]c.PanelAnalyticsItemUnit, len(revLabelList))
	for i, value := range revLabelList {
		avgList[i] = avgMap[value]
		cv, cu := c.ConvertPerfUnit(float64(avgMap[value]))
		avgItems[i] = c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)}
	}
	graph := createTimeGraph([][]int64{avgList}, labelList)

	pi := c.PanelAnalyticsPerf{graph, avgItems, tr.Range, tr.Unit, "time", typ}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_performance", pi})
}

func analyticsRowsToAvgDuoMap(rows *sql.Rows, labelList []int64, avgMap map[int64]int64) (map[string]map[int64]int64, map[string]int, error) {
	aMap := make(map[string]map[int64]int64)
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int64
		var name string
		var createdAt time.Time
		e := rows.Scan(&count, &name, &createdAt)
		if e != nil {
			return aMap, nameMap, e
		}

		// TODO: Bulk log this
		unixCreatedAt := createdAt.Unix()
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}

		vvMap, ok := aMap[name]
		if !ok {
			vvMap = make(map[int64]int64)
			for key, val := range avgMap {
				vvMap[key] = val
			}
			aMap[name] = vvMap
		}
		for _, value := range labelList {
			if unixCreatedAt > value {
				vvMap[value] = (vvMap[value] + count) / 2
				break
			}
		}
		nameMap[name] = (nameMap[name] + int(count)) / 2
	}
	return aMap, nameMap, rows.Err()
}

func sortOVList(ovList []OVItem) []OVItem {
	// Use bubble sort for now as there shouldn't be too many items
	for i := 0; i < len(ovList)-1; i++ {
		for j := 0; j < len(ovList)-1; j++ {
			if ovList[j].count > ovList[j+1].count {
				temp := ovList[j]
				ovList[j] = ovList[j+1]
				ovList[j+1] = temp
			}
		}
	}

	// Invert the direction
	tOVList := make([]OVItem, len(ovList))
	for i, ii := len(ovList)-1, 0; i >= 0; i-- {
		tOVList[ii] = ovList[i]
		ii++
	}
	return tOVList
}

func analyticsAMapToOVList(aMap map[string]map[int64]int64) []OVItem {
	// Order the map
	ovList, i := make([]OVItem, len(aMap)), 0
	for name, avgMap := range aMap {
		var totcount int
		for _, count := range avgMap {
			totcount = (totcount + int(count)) / 2
		}
		ovList[i] = OVItem{name, totcount, avgMap}
		i++
	}
	return sortOVList(ovList)
}

func AnalyticsRoutesPerf(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	// avgMap contains timestamps but not the averages for those stamps
	revLabelList, labelList, avgMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks").Columns("avg,route,createdAt").Where("count!=0 AND route!=''").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	aMap, routeMap, e := analyticsRowsToAvgDuoMap(rows, labelList, avgMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	//c.DebugLogf("aMap: %+v\n", aMap)
	//c.DebugLogf("routeMap: %+v\n", routeMap)
	ovList := analyticsAMapToOVList(aMap)
	//c.DebugLogf("ovList: %+v\n", ovList)

	ex := strings.Split(r.FormValue("ex"), ",")
	inEx := func(name string) bool {
		for _, e := range ex {
			if e == name {
				return true
			}
		}
		return false
	}

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		if inEx(ovitem.name) {
			continue
		}
		if strings.HasPrefix(ovitem.name, "panel.") {
			continue
		}
		viewList := make([]int64, len(revLabelList))
		for i, val := range revLabelList {
			viewList[i] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		shortName := strings.Replace(ovitem.name, "routes.", "r.", -1)
		legendList = append(legendList, shortName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Sort this slice
	var routeItems []c.PanelAnalyticsRoutesPerfItem
	for route, count := range routeMap {
		if inEx(route) {
			continue
		}
		cv, cu := c.ConvertPerfUnit(float64(count))
		routeItems = append(routeItems, c.PanelAnalyticsRoutesPerfItem{
			Route: route,
			Unit:  cu,
			Count: int(cv),
		})
	}

	pi := c.PanelAnalyticsRoutesPerfPage{bp, routeItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_routes_perf", pi})
}

func analyticsRowsToRefMap(rows *sql.Rows) (map[string]int, error) {
	nameMap := make(map[string]int)
	defer rows.Close()
	c.DebugDetail("name - count")
	for rows.Next() {
		var count int
		var name string
		e := rows.Scan(&count, &name)
		if e != nil {
			return nameMap, e
		}
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print(name, " - ", count)
		}
		nameMap[name] += count
	}
	return nameMap, rows.Err()
}

func analyticsRowsToDuoMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[string]map[int64]int64, map[string]int, error) {
	vMap := make(map[string]map[int64]int64)
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int64
		var name string
		var createdAt time.Time
		e := rows.Scan(&count, &name, &createdAt)
		if e != nil {
			return vMap, nameMap, e
		}

		// TODO: Bulk log this
		unixCreatedAt := createdAt.Unix()
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
			log.Print("createdAt: ", createdAt, " - ", unixCreatedAt)
		}

		vvMap, ok := vMap[name]
		if !ok {
			vvMap = make(map[int64]int64)
			for key, val := range viewMap {
				vvMap[key] = val
			}
			vMap[name] = vvMap
		}
		for _, value := range labelList {
			if unixCreatedAt > value {
				vvMap[value] += count
				break
			}
		}
		nameMap[name] += int(count)
	}
	return vMap, nameMap, rows.Err()
}

type OVItem struct {
	name    string
	count   int
	viewMap map[int64]int64
}

func analyticsVMapToOVList(vMap map[string]map[int64]int64) (ovList []OVItem) {
	// Order the map
	ovList, i := make([]OVItem, len(vMap)), 0
	for name, viewMap := range vMap {
		var totcount int
		for _, count := range viewMap {
			totcount += int(count)
		}
		ovList[i] = OVItem{name, totcount, viewMap}
		i++
	}
	return sortOVList(ovList)
}

func AnalyticsForums(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks_forums").Columns("count,forum,createdAt").Where("forum!=''").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	vMap, forumMap, e := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		viewList := make([]int64, len(revLabelList))
		for i, val := range revLabelList {
			viewList[i] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		fid, e := strconv.Atoi(ovitem.name)
		if e != nil {
			return c.InternalError(e, w, r)
		}
		var lName string
		forum, e := c.Forums.Get(fid)
		if e == sql.ErrNoRows {
			lName = "Deleted Forum" // TODO: Localise this
		} else if e != nil {
			return c.InternalError(e, w, r)
		} else {
			lName = forum.Name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Sort this slice
	forumItems, i := make([]c.PanelAnalyticsAgentsItem, len(forumMap)), 0
	for sfid, count := range forumMap {
		fid, e := strconv.Atoi(sfid)
		if e != nil {
			return c.InternalError(e, w, r)
		}
		var lName string
		forum, e := c.Forums.Get(fid)
		if e == sql.ErrNoRows {
			// TODO: Localise this
			lName = "Deleted Forum"
		} else if e != nil {
			return c.InternalError(e, w, r)
		} else {
			lName = forum.Name
		}
		forumItems[i] = c.PanelAnalyticsAgentsItem{
			Agent:         sfid,
			FriendlyAgent: lName,
			Count:         count,
		}
		i++
	}

	pi := c.PanelAnalyticsDuoPage{bp, forumItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_forums", pi})
}

func AnalyticsRoutes(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks").Columns("count,route,createdAt").Where("route!=''").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	vMap, routeMap, e := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	//c.DebugLogf("vMap: %+v\n", vMap)
	//c.DebugLogf("routeMap: %+v\n", routeMap)
	ovList := analyticsVMapToOVList(vMap)
	//c.DebugLogf("ovList: %+v\n", ovList)

	ex := strings.Split(r.FormValue("ex"), ",")
	inEx := func(name string) bool {
		for _, e := range ex {
			if e == name {
				return true
			}
		}
		return false
	}

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		if inEx(ovitem.name) {
			continue
		}
		viewList := make([]int64, len(revLabelList))
		for i, val := range revLabelList {
			viewList[i] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		shortName := strings.Replace(ovitem.name, "routes.", "r.", -1)
		legendList = append(legendList, shortName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Sort this slice
	var routeItems []c.PanelAnalyticsRoutesItem
	for route, count := range routeMap {
		if inEx(route) {
			continue
		}
		routeItems = append(routeItems, c.PanelAnalyticsRoutesItem{
			Route: route,
			Count: count,
		})
	}

	pi := c.PanelAnalyticsRoutesPage{bp, routeItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_routes", pi})
}

// Trialling multi-series charts
func AnalyticsAgents(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks_agents").Columns("count,browser,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	vMap, agentMap, e := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	ex := strings.Split(r.FormValue("ex"), ",")
	inEx := func(name string) bool {
		for _, e := range ex {
			if e == name {
				return true
			}
		}
		return false
	}

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		if inEx(ovitem.name) {
			continue
		}
		lName, ok := p.GetUserAgentPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		if inEx(lName) {
			continue
		}
		viewList := make([]int64, len(revLabelList))
		for i, val := range revLabelList {
			viewList[i] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		legendList = append(legendList, lName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Sort this slice
	var agentItems []c.PanelAnalyticsAgentsItem
	for agent, count := range agentMap {
		if inEx(agent) {
			continue
		}
		aAgent, ok := p.GetUserAgentPhrase(agent)
		if !ok {
			aAgent = agent
		}
		if inEx(aAgent) {
			continue
		}
		agentItems = append(agentItems, c.PanelAnalyticsAgentsItem{
			Agent:         agent,
			FriendlyAgent: aAgent,
			Count:         count,
		})
	}

	pi := c.PanelAnalyticsDuoPage{bp, agentItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_agents", pi})
}

func AnalyticsSystems(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks_systems").Columns("count,system,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	vMap, osMap, e := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		viewList := make([]int64, len(revLabelList))
		for ii, val := range revLabelList {
			viewList[ii] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		lName, ok := p.GetOSPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Sort this slice
	systemItems, i := make([]c.PanelAnalyticsAgentsItem, len(osMap)), 0
	for system, count := range osMap {
		sSystem, ok := p.GetOSPhrase(system)
		if !ok {
			sSystem = system
		}
		systemItems[i] = c.PanelAnalyticsAgentsItem{
			Agent:         system,
			FriendlyAgent: sSystem,
			Count:         count,
		}
		i++
	}

	pi := c.PanelAnalyticsDuoPage{bp, systemItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_systems", pi})
}

func AnalyticsLanguages(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := PreAnalyticsDetail(w, r, u)
	if ferr != nil {
		return ferr
	}
	bp.AddScript("chartist/chartist-plugin-legend.min.js")
	bp.AddSheet("chartist/chartist-plugin-legend.css")

	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}
	revLabelList, labelList, viewMap := c.AnalyticsTimeRangeToLabelList(tr)

	rows, e := qgen.NewAcc().Select("viewchunks_langs").Columns("count,lang,createdAt").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	vMap, langMap, e := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	ex := strings.Split(r.FormValue("ex"), ",")
	inEx := func(name string) bool {
		for _, e := range ex {
			if e == name {
				return true
			}
		}
		return false
	}

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		if inEx(ovitem.name) {
			continue
		}
		lName, ok := p.GetHumanLangPhrase(ovitem.name)
		if !ok {
			lName = ovitem.name
		}
		if inEx(lName) {
			continue
		}

		viewList := make([]int64, len(revLabelList))
		for _, val := range revLabelList {
			viewList[i] = ovitem.viewMap[val]
		}
		vList = append(vList, viewList)
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := createTimeGraph(vList, labelList, legendList)

	// TODO: Can we de-duplicate these analytics functions further?
	// TODO: Sort this slice
	var langItems []c.PanelAnalyticsAgentsItem
	for lang, count := range langMap {
		if inEx(lang) {
			continue
		}
		lLang, ok := p.GetHumanLangPhrase(lang)
		if !ok {
			lLang = lang
		}
		if inEx(lLang) {
			continue
		}
		langItems = append(langItems, c.PanelAnalyticsAgentsItem{
			Agent:         lang,
			FriendlyAgent: lLang,
			Count:         count,
		})
	}

	pi := c.PanelAnalyticsDuoPage{bp, langItems, graph, tr.Range}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_langs", pi})
}

func AnalyticsReferrers(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	bp, ferr := buildBasePage(w, r, u, "analytics", "analytics")
	if ferr != nil {
		return ferr
	}
	tr, e := analyticsTimeRange(r.FormValue("timeRange"))
	if e != nil {
		return c.LocalError(e.Error(), w, r, u)
	}

	rows, e := qgen.NewAcc().Select("viewchunks_referrers").Columns("count,domain").DateCutoff("createdAt", tr.Quantity, tr.Unit).Query()
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	refMap, e := analyticsRowsToRefMap(rows)
	if e != nil {
		return c.InternalError(e, w, r)
	}
	showSpam := r.FormValue("spam") == "1"

	isSpammy := func(domain string) bool {
		for _, substr := range c.SpammyDomainBits {
			if strings.Contains(domain, substr) {
				return true
			}
		}
		return false
	}

	// TODO: Sort this slice
	var refItems []c.PanelAnalyticsAgentsItem
	for domain, count := range refMap {
		sdomain := c.SanitiseSingleLine(domain)
		if !showSpam && isSpammy(sdomain) {
			continue
		}
		refItems = append(refItems, c.PanelAnalyticsAgentsItem{
			Agent: sdomain,
			Count: count,
		})
	}

	pi := c.PanelAnalyticsReferrersPage{bp, refItems, tr.Range, showSpam}
	return renderTemplate("panel", w, r, bp.Header, c.Panel{bp, "panel_analytics_right", "analytics", "panel_analytics_referrers", pi})
}
