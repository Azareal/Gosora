// +build !no_ws

/*
*
*	Gosora WebSocket Subsystem
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	p "github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/gopsutil/cpu"
	"github.com/Azareal/gopsutil/mem"
	"github.com/gorilla/websocket"
)

// TODO: Disable WebSockets on high load? Add a Control Panel interface for disabling it?
var EnableWebsockets = true // Put this in caps for consistency with the other constants?

var wsUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var errWsNouser = errors.New("This user isn't connected via WebSockets")

func init() {
	adminStatsWatchers = make(map[*websocket.Conn]*WSUser)
	topicListWatchers = make(map[*WSUser]struct{})
	topicWatchers = make(map[int]map[*WSUser]struct{})
}

//easyjson:json
type WsTopicList struct {
	Topics     []*WsTopicsRow
	LastPage   int // Not for WebSockets, but for the JSON endpoint for /topics/ to keep the paginator functional
	LastUpdate int64
}

// TODO: How should we handle errors for this?
// TODO: Move this out of common?
func RouteWebsockets(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Spit out a 500 instead of nil?
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return LocalError("unable to upgrade", w, r, user)
	}
	defer conn.Close()

	wsUser, err := WsHub.AddConn(user, conn)
	if err != nil {
		return nil
	}

	//conn.SetReadLimit(/* put the max request size from earlier here? */)
	//conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var currentPage string
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if user.ID == 0 {
				WsHub.GuestLock.Lock()
				delete(WsHub.OnlineGuests, wsUser)
				WsHub.GuestLock.Unlock()
			} else {
				// TODO: Make sure the admin is removed from the admin stats list in the case that an error happens
				WsHub.RemoveConn(wsUser, conn)
			}
			break
		}
		if conn == nil {
			panic("conn must not be nil")
		}

		for _, msg := range bytes.Split(message, []byte("\r")) {
			//StoppedServer("Profile end") // A bit of code for me to profile the software
			if bytes.HasPrefix(msg, []byte("page ")) {
				msgblocks := bytes.SplitN(msg, []byte(" "), 2)
				if len(msgblocks) < 2 {
					continue
				}

				if !bytes.Equal(msgblocks[1], []byte(currentPage)) {
					wsLeavePage(wsUser, conn, currentPage)
					currentPage = string(msgblocks[1])
					wsPageResponses(wsUser, conn, currentPage)
				}
			} else if bytes.HasPrefix(msg, []byte("resume ")) {
				msgblocks := bytes.SplitN(msg, []byte(" "), 3)
				if len(msgblocks) < 3 {
					continue
				}
				//log.Print("resuming on " + string(msgblocks[1]) + " at " + string(msgblocks[2]))

				if !bytes.Equal(msgblocks[1], []byte(currentPage)) {
					wsLeavePage(wsUser, conn, currentPage) // Avoid clients abusing late resumes
					currentPage = string(msgblocks[1])
					// TODO: Synchronise this better?
					resume, err := strconv.ParseInt(string(msgblocks[2]), 10, 64)
					wsPageResponses(wsUser, conn, currentPage)
					if err != nil {
						wsPageResume(wsUser, conn, currentPage, resume)
					}
				}
			}
			/*if bytes.Equal(message,[]byte(`start-view`)) {
			} else if bytes.Equal(message,[]byte(`end-view`)) {
			}*/
		}
	}
	DebugLog("Closing connection for user " + strconv.Itoa(user.ID))
	return nil
}

// TODO: Copied from routes package for use in wsPageResponse, find a more elegant solution.
func ParseSEOURL(urlBit string) (slug string, id int, err error) {
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	tid, err := strconv.Atoi(halves[1])
	return halves[0], tid, err
}

// TODO: Use a map instead of a switch to make this more modular?
func wsPageResponses(wsUser *WSUser, conn *websocket.Conn, page string) {
	if page == "/" {
		page = Config.DefaultPath
	}

	DebugLog("Entering page " + page)
	switch {
	// Live Topic List is an experimental feature
	// TODO: Optimise this to reduce the amount of contention
	case page == "/topics/":
		topicListMutex.Lock()
		topicListWatchers[wsUser] = struct{}{}
		topicListMutex.Unlock()
		// TODO: Evict from page when permissions change? Or check user perms every-time before sending data?
	case strings.HasPrefix(page, "/topic/"):
		//fmt.Println("entering topic prefix websockets zone")
		if wsUser.User.ID == 0 {
			return
		}
		_, tid, err := ParseSEOURL(page)
		if err != nil {
			return
		}
		topic, err := Topics.Get(tid)
		if err != nil {
			return
		}
		if !Forums.Exists(topic.ParentID) {
			return
		}
		usercpy := BlankUser()
		*usercpy = *wsUser.User
		usercpy.Init()

		/*skip, rerr := header.Hooks.VhookSkippable("ws_topic_check_pre_perms", w, r, usercpy, &fid, &header)
		if skip || rerr != nil {
			return
		}*/

		fperms, err := FPStore.Get(topic.ParentID, usercpy.Group)
		if err == ErrNoRows {
			fperms = BlankForumPerms()
		} else if err != nil {
			return
		}
		cascadeForumPerms(fperms, usercpy)
		if !usercpy.Perms.ViewTopic {
			return
		}

		topicMutex.Lock()
		_, ok := topicWatchers[topic.ID]
		if !ok {
			topicWatchers[topic.ID] = make(map[*WSUser]struct{})
		}
		topicWatchers[topic.ID][wsUser] = struct{}{}
		topicMutex.Unlock()
	case page == "/panel/":
		if !wsUser.User.IsSuperMod {
			return
		}
		// Listen for changes and inform the admins...
		adminStatsMutex.Lock()
		watchers := len(adminStatsWatchers)
		adminStatsWatchers[conn] = wsUser
		if watchers == 0 {
			go adminStatsTicker()
		}
		adminStatsMutex.Unlock()
	default:
		return
	}
	err := wsUser.SetPageForSocket(conn, page)
	if err != nil {
		LogError(err)
	}
}

// TODO: Use a map instead of a switch to make this more modular?
// TODO: Implement this
func wsPageResume(wsUser *WSUser, conn *websocket.Conn, page string, resume int64) {
	if page == "/" {
		page = Config.DefaultPath
	}

	switch {
	// TODO: Synchronise this bit of resume with tick updating lastTopicList?
	case page == "/topics/":
		/*if resume >= hub.lastTick.Unix() {
			conn.Write([]byte("resume tooslow"))
		} else {
		conn.Write([]byte("resume success"))
		}*/
	default:
		return
	}
}

// TODO: Use a map instead of a switch to make this more modular?
func wsLeavePage(wsUser *WSUser, conn *websocket.Conn, page string) {
	if page == "/" {
		page = Config.DefaultPath
	} else if page != "" {
		DebugLog("Leaving page " + page)
	}
	switch {
	case page == "/topics/":
		wsUser.FinalizePage("/topics/", func() {
			topicListMutex.Lock()
			delete(topicListWatchers, wsUser)
			topicListMutex.Unlock()
		})
	case strings.HasPrefix(page, "/topic/"):
		//fmt.Println("leaving topic prefix websockets zone")
		if wsUser.User.ID == 0 {
			return
		}
		wsUser.FinalizePage(page, func() {
			_, tid, err := ParseSEOURL(page)
			if err != nil {
				return
			}
			topicMutex.Lock()
			defer topicMutex.Unlock()
			topic, ok := topicWatchers[tid]
			if !ok {
				return
			}
			_, ok = topic[wsUser]
			if !ok {
				return
			}
			delete(topic, wsUser)
			if len(topic) == 0 {
				delete(topicWatchers, tid)
			}
		})
	case page == "/panel/":
		adminStatsMutex.Lock()
		delete(adminStatsWatchers, conn)
		adminStatsMutex.Unlock()
	}
	err := wsUser.SetPageForSocket(conn, "")
	if err != nil {
		LogError(err)
	}
}

// TODO: Abstract this
// TODO: Use odd-even sharding
var topicListWatchers map[*WSUser]struct{}
var topicListMutex sync.RWMutex
var topicWatchers map[int]map[*WSUser]struct{} // map[tid]watchers
var topicMutex sync.RWMutex
var adminStatsWatchers map[*websocket.Conn]*WSUser
var adminStatsMutex sync.RWMutex

func adminStatsTicker() {
	time.Sleep(time.Second)

	lastUonline := -1
	lastGonline := -1
	lastTotonline := -1
	lastCPUPerc := -1
	var lastAvailableRAM int64 = -1
	var noStatUpdates, noRAMUpdates bool

	var onlineColour, onlineGuestsColour, onlineUsersColour, cpustr, cpuColour, ramstr, ramColour string
	var cpuerr, ramerr error
	var memres *mem.VirtualMemoryStat
	var cpuPerc []float64

	var totunit, uunit, gunit string

	lessThanSwitch := func(number, lowerBound, midBound int) string {
		switch {
		case number < lowerBound:
			return "stat_green"
		case number < midBound:
			return "stat_orange"
		}
		return "stat_red"
	}
	greaterThanSwitch := func(number, lowerBound, midBound int) string {
		switch {
		case number > midBound:
			return "stat_green"
		case number > lowerBound:
			return "stat_orange"
		}
		return "stat_red"
	}

AdminStatLoop:
	for {
		adminStatsMutex.RLock()
		watchCount := len(adminStatsWatchers)
		adminStatsMutex.RUnlock()
		if watchCount == 0 {
			break AdminStatLoop
		}

		cpuPerc, cpuerr = cpu.Percent(time.Second, true)
		memres, ramerr = mem.VirtualMemory()
		uonline := WsHub.UserCount()
		gonline := WsHub.GuestCount()
		totonline := uonline + gonline
		reqCount := 0

		// It's far more likely that the CPU Usage will change than the other stats, so we'll optimise them separately...
		noStatUpdates = (uonline == lastUonline && gonline == lastGonline && totonline == lastTotonline)
		noRAMUpdates = (lastAvailableRAM == int64(memres.Available))
		if int(cpuPerc[0]) == lastCPUPerc && noStatUpdates && noRAMUpdates {
			time.Sleep(time.Second)
			continue
		}

		if !noStatUpdates {
			onlineColour = greaterThanSwitch(totonline, 3, 10)
			onlineGuestsColour = greaterThanSwitch(gonline, 1, 10)
			onlineUsersColour = greaterThanSwitch(uonline, 1, 5)

			totonline, totunit = ConvertFriendlyUnit(totonline)
			uonline, uunit = ConvertFriendlyUnit(uonline)
			gonline, gunit = ConvertFriendlyUnit(gonline)
		}

		if cpuerr != nil {
			cpustr = "Unknown"
		} else {
			calcperc := int(cpuPerc[0]) / runtime.NumCPU()
			cpustr = strconv.Itoa(calcperc)
			switch {
			case calcperc < 30:
				cpuColour = "stat_green"
			case calcperc < 75:
				cpuColour = "stat_orange"
			default:
				cpuColour = "stat_red"
			}
		}

		if !noRAMUpdates {
			if ramerr != nil {
				ramstr = "Unknown"
			} else {
				totalCount, totalUnit := ConvertByteUnit(float64(memres.Total))
				usedCount := ConvertByteInUnit(float64(memres.Total-memres.Available), totalUnit)

				// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
				var totstr string
				if (totalCount - float64(int(totalCount))) > 0.85 {
					usedCount += 1.0 - (totalCount - float64(int(totalCount)))
					totstr = strconv.Itoa(int(totalCount) + 1)
				} else {
					totstr = fmt.Sprintf("%.1f", totalCount)
				}

				if usedCount > totalCount {
					usedCount = totalCount
				}
				ramstr = fmt.Sprintf("%.1f", usedCount) + " / " + totstr + totalUnit

				ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
				ramColour = lessThanSwitch(int(ramperc), 50, 75)
			}
		}

		// Acquire a write lock for now, so we can handle the delete() case below and the read one simultaneously
		// TODO: Stop taking a write lock here if it isn't necessary
		adminStatsMutex.Lock()
		for conn := range adminStatsWatchers {
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				delete(adminStatsWatchers, conn)
				continue
			}

			// nolint
			// TODO: Use JSON for this to make things more portable and easier to convert to MessagePack, if need be?
			write := func(msg string) {
				w.Write([]byte(msg + "\r"))
			}
			push := func(id, msg string) {
				write("set #" + id + " <span>" + msg + "</span>")
			}
			pushc := func(id, classes string) {
				write("set-class #" + id + " " + classes)
			}
			if !noStatUpdates {
				push("dash-totonline", p.GetTmplPhrasef("panel_dashboard_online", totonline, totunit))
				push("dash-gonline", p.GetTmplPhrasef("panel_dashboard_guests_online", gonline, gunit))
				push("dash-uonline", p.GetTmplPhrasef("panel_dashboard_users_online", uonline, uunit))
				push("dash-reqs", strconv.Itoa(reqCount)+" reqs / second")
				pushc("dash-totonline", "grid_item grid_stat "+onlineColour)
				pushc("dash-gonline", "grid_item grid_stat "+onlineGuestsColour)
				pushc("dash-uonline", "grid_item grid_stat "+onlineUsersColour)
				//pushc("dash-reqs","grid_item grid_stat grid_end_group")
			}
			push("dash-cpu", p.GetTmplPhrasef("panel_dashboard_cpu", cpustr)+"%")
			pushc("dash-cpu", "grid_item grid_istat "+cpuColour)

			if !noRAMUpdates {
				push("dash-ram", p.GetTmplPhrasef("panel_dashboard_ram", ramstr))
				pushc("dash-ram", "grid_item grid_istat "+ramColour)
			}
			w.Close()
		}
		adminStatsMutex.Unlock()

		lastUonline = uonline
		lastGonline = gonline
		lastTotonline = totonline
		lastCPUPerc = int(cpuPerc[0])
		lastAvailableRAM = int64(memres.Available)
	}
}
