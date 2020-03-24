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

// TODO: Move this to another file, probably common/pages.go
type AnalyticsTimeRange struct {
	Quantity   int
	Unit       string
	Slices     int
	SliceWidth int
	Range      string
}

func analyticsTimeRange(rawTimeRange string) (*AnalyticsTimeRange, error) {
	tRange := &AnalyticsTimeRange{
		Quantity:   6,
		Unit:       "hour",
		Slices:     12,
		SliceWidth: 60 * 30,
		Range:      "six-hours",
	}

	switch rawTimeRange {
	// This might be pushing it, we might want to come up with a more efficient scheme for dealing with large timeframes like this
	case "one-year":
		tRange.Quantity = 12
		tRange.Unit = "month"
		tRange.Slices = 12
		tRange.SliceWidth = 60 * 60 * 24 * 30
		tRange.Range = "one-year"
	case "three-months":
		tRange.Quantity = 90
		tRange.Unit = "day"
		tRange.Slices = 30
		tRange.SliceWidth = 60 * 60 * 24 * 3
		tRange.Range = "three-months"
	case "one-month":
		tRange.Quantity = 30
		tRange.Unit = "day"
		tRange.Slices = 30
		tRange.SliceWidth = 60 * 60 * 24
		tRange.Range = "one-month"
	case "one-week":
		tRange.Quantity = 7
		tRange.Unit = "day"
		tRange.Slices = 14
		tRange.SliceWidth = 60 * 60 * 12
		tRange.Range = "one-week"
	case "two-days": // Two days is experimental
		tRange.Quantity = 2
		tRange.Unit = "day"
		tRange.Slices = 24
		tRange.SliceWidth = 60 * 60 * 2
		tRange.Range = "two-days"
	case "one-day":
		tRange.Quantity = 1
		tRange.Unit = "day"
		tRange.Slices = 24
		tRange.SliceWidth = 60 * 60
		tRange.Range = "one-day"
	case "twelve-hours":
		tRange.Quantity = 12
		tRange.Slices = 24
		tRange.Range = "twelve-hours"
	case "six-hours", "":
	default:
		return tRange, errors.New("Unknown time range")
	}
	return tRange, nil
}

// TODO: Clamp it rather than using an offset off the current time to avoid chaotic changes in stats as adjacent sets converge and diverge?
func analyticsTimeRangeToLabelList(timeRange *AnalyticsTimeRange) (revLabelList []int64, labelList []int64, viewMap map[int64]int64) {
	viewMap = make(map[int64]int64)
	currentTime := time.Now().Unix()
	for i := 1; i <= timeRange.Slices; i++ {
		label := currentTime - int64(i*timeRange.SliceWidth)
		revLabelList = append(revLabelList, label)
		viewMap[label] = 0
	}
	for _, value := range revLabelList {
		labelList = append(labelList, value)
	}
	return revLabelList, labelList, viewMap
}

func analyticsRowsToViewMap(rows *sql.Rows, labelList []int64, viewMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		err := rows.Scan(&count, &createdAt)
		if err != nil {
			return viewMap, err
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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

type pAvg struct {
	Avg int64
	Tot int64
}

func analyticsRowsToAverageMap(rows *sql.Rows, labelList []int64, avgMap map[int64]int64) (map[int64]int64, error) {
	defer rows.Close()
	for rows.Next() {
		var count int64
		var createdAt time.Time
		err := rows.Scan(&count, &createdAt)
		if err != nil {
			return avgMap, err
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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
		err := rows.Scan(&stack, &heap, &createdAt)
		if err != nil {
			return avgMap, err
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("stack: ", stack)
			log.Print("heap: ", heap)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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
		err := rows.Scan(&low, &high, &avg, &createdAt)
		if err != nil {
			return avgMap, err
		}
		unixCreatedAt := createdAt.Unix()
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("low: ", low)
			log.Print("high: ", high)
			log.Print("avg: ", avg)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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
		for _, value := range labelList {
			if unixCreatedAt > value {
				prev := pAvgMap[value]
				prev.Avg += dat
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

func PreAnalyticsDetail(w http.ResponseWriter, r *http.Request, user *c.User) (*c.BasePanelPage, c.RouteError) {
	bp, ferr := buildBasePage(w, r, user, "analytics", "analytics")
	if ferr != nil {
		return nil, ferr
	}
	bp.AddSheet("chartist/chartist.min.css")
	bp.AddScript("chartist/chartist.min.js")
	bp.AddScriptAsync("analytics.js")
	bp.LooseCSP = true
	return bp, nil
}

func AnalyticsViews(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsViews")
	// TODO: Add some sort of analytics store / iterator?
	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count,createdAt").Where("route=''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	viewList := make([]int64, len(revLabelList))
	viewItems := make([]c.PanelAnalyticsItem, len(revLabelList))
	for i, value := range revLabelList {
		viewList[i] = viewMap[value]
		viewItems[i] = c.PanelAnalyticsItem{Time: value, Count: viewMap[value]}
	}

	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	var ttime string
	if timeRange.Range == "six-hours" || timeRange.Range == "twelve-hours" || timeRange.Range == "one-day" {
		ttime = "time"
	}

	pi := c.PanelAnalyticsStd{graph, viewItems, timeRange.Range, timeRange.Unit, ttime}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_views", pi})
}

func AnalyticsRouteViews(w http.ResponseWriter, r *http.Request, user *c.User, route string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsRouteViews")
	// TODO: Validate the route is valid
	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count,createdAt").Where("route=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(route)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []c.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, c.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)

	pi := c.PanelAnalyticsRoutePage{basePage, c.SanitiseSingleLine(route), graph, viewItems, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_route_views", pi})
}

func AnalyticsAgentViews(w http.ResponseWriter, r *http.Request, user *c.User, agent string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	// ? Only allow valid agents? The problem with this is that agents wind up getting renamed and it would take a migration to get them all up to snuff
	agent = c.SanitiseSingleLine(agent)

	c.DebugLog("in panel.AnalyticsAgentViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_agents").Columns("count,createdAt").Where("browser=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(agent)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)

	friendlyAgent, ok := p.GetUserAgentPhrase(agent)
	if !ok {
		friendlyAgent = agent
	}

	pi := c.PanelAnalyticsAgentPage{basePage, agent, friendlyAgent, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_agent_views", pi})
}

func AnalyticsForumViews(w http.ResponseWriter, r *http.Request, user *c.User, sfid string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	fid, err := strconv.Atoi(sfid)
	if err != nil {
		return c.LocalError("Invalid integer", w, r, user)
	}

	c.DebugLog("in panel.AnalyticsForumViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_forums").Columns("count,createdAt").Where("forum=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(fid)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)

	forum, err := c.Forums.Get(fid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.PanelAnalyticsAgentPage{basePage, sfid, forum.Name, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_forum_views", pi})
}

func AnalyticsSystemViews(w http.ResponseWriter, r *http.Request, user *c.User, system string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	system = c.SanitiseSingleLine(system)

	c.DebugLog("in panel.AnalyticsSystemViews")
	// TODO: Verify the OS name is valid
	rows, err := qgen.NewAcc().Select("viewchunks_systems").Columns("count,createdAt").Where("system=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(system)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)

	friendlySystem, ok := p.GetOSPhrase(system)
	if !ok {
		friendlySystem = system
	}

	pi := c.PanelAnalyticsAgentPage{basePage, system, friendlySystem, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_system_views", pi})
}

func AnalyticsLanguageViews(w http.ResponseWriter, r *http.Request, user *c.User, lang string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)
	lang = c.SanitiseSingleLine(lang)

	c.DebugLog("in panel.AnalyticsLanguageViews")
	// TODO: Verify the language code is valid
	rows, err := qgen.NewAcc().Select("viewchunks_langs").Columns("count,createdAt").Where("lang=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(lang)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}

	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)

	friendlyLang, ok := p.GetHumanLangPhrase(lang)
	if !ok {
		friendlyLang = lang
	}

	pi := c.PanelAnalyticsAgentPage{basePage, lang, friendlyLang, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_lang_views", pi})
}

func AnalyticsReferrerViews(w http.ResponseWriter, r *http.Request, user *c.User, domain string) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsReferrerViews")
	// TODO: Verify the agent is valid
	rows, err := qgen.NewAcc().Select("viewchunks_referrers").Columns("count,createdAt").Where("domain=?").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query(domain)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsAgentPage{basePage, c.SanitiseSingleLine(domain), "", graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_referrer_views", pi})
}

func AnalyticsTopics(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsTopics")
	rows, err := qgen.NewAcc().Select("topicchunks").Columns("count,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []c.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, c.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsStd{graph, viewItems, timeRange.Range, timeRange.Unit, "time"}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_topics", pi})
}

func AnalyticsPosts(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsPosts")
	rows, err := qgen.NewAcc().Select("postchunks").Columns("count,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	viewMap, err = analyticsRowsToViewMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var viewList []int64
	var viewItems []c.PanelAnalyticsItem
	for _, value := range revLabelList {
		viewList = append(viewList, viewMap[value])
		viewItems = append(viewItems, c.PanelAnalyticsItem{Time: value, Count: viewMap[value]})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{viewList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsStd{graph, viewItems, timeRange.Range, timeRange.Unit, "time"}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_posts", pi})
}

func AnalyticsMemory(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, avgMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsMemory")
	rows, err := qgen.NewAcc().Select("memchunks").Columns("count,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	avgMap, err = analyticsRowsToAverageMap(rows, labelList, avgMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	var avgList []int64
	var avgItems []c.PanelAnalyticsItemUnit
	for _, value := range revLabelList {
		avgList = append(avgList, avgMap[value])
		cv, cu := c.ConvertByteUnit(float64(avgMap[value]))
		avgItems = append(avgItems, c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{avgList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsStdUnit{graph, avgItems, timeRange.Range, timeRange.Unit, "time"}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_memory", pi})
}

// TODO: Show stack and heap memory separately on the chart
func AnalyticsActiveMemory(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, avgMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsActiveMemory")
	rows, err := qgen.NewAcc().Select("memchunks").Columns("stack,heap,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
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
	avgMap, err = analyticsRowsToAverageMap2(rows, labelList, avgMap, typ)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	var avgList []int64
	var avgItems []c.PanelAnalyticsItemUnit
	for _, value := range revLabelList {
		avgList = append(avgList, avgMap[value])
		cv, cu := c.ConvertByteUnit(float64(avgMap[value]))
		avgItems = append(avgItems, c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{avgList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsActiveMemory{graph, avgItems, timeRange.Range, timeRange.Unit, "time", typ}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_active_memory", pi})
}

func AnalyticsPerf(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, avgMap := analyticsTimeRangeToLabelList(timeRange)

	c.DebugLog("in panel.AnalyticsPerf")
	rows, err := qgen.NewAcc().Select("perfchunks").Columns("low,high,avg,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
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
	avgMap, err = analyticsRowsToAverageMap3(rows, labelList, avgMap, typ)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Adjust for the missing chunks in week and month
	var avgList []int64
	var avgItems []c.PanelAnalyticsItemUnit
	for _, value := range revLabelList {
		avgList = append(avgList, avgMap[value])
		cv, cu := c.ConvertPerfUnit(float64(avgMap[value]))
		avgItems = append(avgItems, c.PanelAnalyticsItemUnit{Time: value, Unit: cu, Count: int64(cv)})
	}
	graph := c.PanelTimeGraph{Series: [][]int64{avgList}, Labels: labelList}
	c.DebugLogf("graph: %+v\n", graph)
	pi := c.PanelAnalyticsPerf{graph, avgItems, timeRange.Range, timeRange.Unit, "time", typ}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_performance", pi})
}

func analyticsRowsToAvgDuoMap(rows *sql.Rows, labelList []int64, avgMap map[int64]int64) (map[string]map[int64]int64, map[string]int, error) {
	aMap := make(map[string]map[int64]int64)
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int64
		var name string
		var createdAt time.Time
		err := rows.Scan(&count, &name, &createdAt)
		if err != nil {
			return aMap, nameMap, err
		}

		// TODO: Bulk log this
		unixCreatedAt := createdAt.Unix()
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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

func sortOVList(ovList []OVItem) (tOVList []OVItem) {
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
	for i := len(ovList) - 1; i >= 0; i-- {
		tOVList = append(tOVList, ovList[i])
	}
	return tOVList
}

func analyticsAMapToOVList(aMap map[string]map[int64]int64) (ovList []OVItem) {
	// Order the map
	for name, avgMap := range aMap {
		var totcount int
		for _, count := range avgMap {
			totcount = (totcount + int(count)) / 2
		}
		ovList = append(ovList, OVItem{name, totcount, avgMap})
	}

	return sortOVList(ovList)
}

func AnalyticsRoutesPerf(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	// avgMap contains timestamps but not the averages for those stamps
	revLabelList, labelList, avgMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks").Columns("avg,route,createdAt").Where("count!=0 AND route!=''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	aMap, routeMap, err := analyticsRowsToAvgDuoMap(rows, labelList, avgMap)
	if err != nil {
		return c.InternalError(err, w, r)
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
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		shortName := strings.Replace(ovitem.name, "routes.", "r.", -1)
		legendList = append(legendList, shortName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

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

	pi := c.PanelAnalyticsRoutesPerfPage{basePage, routeItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_routes_perf", pi})
}

func analyticsRowsToRefMap(rows *sql.Rows) (map[string]int, error) {
	nameMap := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var count int
		var name string
		err := rows.Scan(&count, &name)
		if err != nil {
			return nameMap, err
		}
		// TODO: Bulk log this
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
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
		err := rows.Scan(&count, &name, &createdAt)
		if err != nil {
			return vMap, nameMap, err
		}

		// TODO: Bulk log this
		unixCreatedAt := createdAt.Unix()
		if c.Dev.SuperDebug {
			log.Print("count: ", count)
			log.Print("name: ", name)
			log.Print("createdAt: ", createdAt)
			log.Print("unixCreatedAt: ", unixCreatedAt)
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
	for name, viewMap := range vMap {
		var totcount int
		for _, count := range viewMap {
			totcount += int(count)
		}
		ovList = append(ovList, OVItem{name, totcount, viewMap})
	}

	return sortOVList(ovList)
}

func AnalyticsForums(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_forums").Columns("count,forum,createdAt").Where("forum!=''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	vMap, forumMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		fid, err := strconv.Atoi(ovitem.name)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		var lName string
		forum, err := c.Forums.Get(fid)
		if err == sql.ErrNoRows {
			// TODO: Localise this
			lName = "Deleted Forum"
		} else if err != nil {
			return c.InternalError(err, w, r)
		} else {
			lName = forum.Name
		}
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var forumItems []c.PanelAnalyticsAgentsItem
	for sfid, count := range forumMap {
		fid, err := strconv.Atoi(sfid)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		var lName string
		forum, err := c.Forums.Get(fid)
		if err == sql.ErrNoRows {
			// TODO: Localise this
			lName = "Deleted Forum"
		} else if err != nil {
			return c.InternalError(err, w, r)
		} else {
			lName = forum.Name
		}
		forumItems = append(forumItems, c.PanelAnalyticsAgentsItem{
			Agent:         sfid,
			FriendlyAgent: lName,
			Count:         count,
		})
	}

	pi := c.PanelAnalyticsDuoPage{basePage, forumItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_forums", pi})
}

func AnalyticsRoutes(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks").Columns("count,route,createdAt").Where("route!=''").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	vMap, routeMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
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
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		shortName := strings.Replace(ovitem.name, "routes.", "r.", -1)
		legendList = append(legendList, shortName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

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

	pi := c.PanelAnalyticsRoutesPage{basePage, routeItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_routes", pi})
}

// Trialling multi-series charts
func AnalyticsAgents(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_agents").Columns("count,browser,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	vMap, agentMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
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
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		legendList = append(legendList, lName)
		if i >= 7 {
			break
		}
		i++
	}
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

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

	pi := c.PanelAnalyticsDuoPage{basePage, agentItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_agents", pi})
}

func AnalyticsSystems(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_systems").Columns("count,system,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	vMap, osMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	ovList := analyticsVMapToOVList(vMap)

	var vList [][]int64
	var legendList []string
	var i int
	for _, ovitem := range ovList {
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
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
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

	// TODO: Sort this slice
	var systemItems []c.PanelAnalyticsAgentsItem
	for system, count := range osMap {
		sSystem, ok := p.GetOSPhrase(system)
		if !ok {
			sSystem = system
		}
		systemItems = append(systemItems, c.PanelAnalyticsAgentsItem{
			Agent:         system,
			FriendlyAgent: sSystem,
			Count:         count,
		})
	}

	pi := c.PanelAnalyticsDuoPage{basePage, systemItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_systems", pi})
}

func AnalyticsLanguages(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := PreAnalyticsDetail(w, r, user)
	if ferr != nil {
		return ferr
	}
	basePage.AddScript("chartist/chartist-plugin-legend.min.js")
	basePage.AddSheet("chartist/chartist-plugin-legend.css")

	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	revLabelList, labelList, viewMap := analyticsTimeRangeToLabelList(timeRange)

	rows, err := qgen.NewAcc().Select("viewchunks_langs").Columns("count,lang,createdAt").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	vMap, langMap, err := analyticsRowsToDuoMap(rows, labelList, viewMap)
	if err != nil {
		return c.InternalError(err, w, r)
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
		var viewList []int64
		for _, value := range revLabelList {
			viewList = append(viewList, ovitem.viewMap[value])
		}
		vList = append(vList, viewList)
		legendList = append(legendList, lName)
		if i >= 6 {
			break
		}
		i++
	}
	graph := c.PanelTimeGraph{Series: vList, Labels: labelList, Legends: legendList}
	c.DebugLogf("graph: %+v\n", graph)

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

	pi := c.PanelAnalyticsDuoPage{basePage, langItems, graph, timeRange.Range}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_langs", pi})
}

func AnalyticsReferrers(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, user, "analytics", "analytics")
	if ferr != nil {
		return ferr
	}
	timeRange, err := analyticsTimeRange(r.FormValue("timeRange"))
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}

	rows, err := qgen.NewAcc().Select("viewchunks_referrers").Columns("count,domain").DateCutoff("createdAt", timeRange.Quantity, timeRange.Unit).Query()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	refMap, err := analyticsRowsToRefMap(rows)
	if err != nil {
		return c.InternalError(err, w, r)
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

	pi := c.PanelAnalyticsReferrersPage{basePage, refItems, timeRange.Range, showSpam}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_analytics_right", "analytics", "panel_analytics_referrers", pi})
}
