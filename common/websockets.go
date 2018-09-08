// +build !no_ws

/*
*
*	Gosora WebSocket Subsystem
*	Copyright Azareal 2017 - 2019
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
	"sync"
	"time"

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
	topicListWatchers = make(map[*WSUser]bool)
}

type WsTopicList struct {
	Topics []*WsTopicsRow
}

// TODO: How should we handle errors for this?
// TODO: Move this out of common?
func RouteWebsockets(w http.ResponseWriter, r *http.Request, user User) RouteError {
	// TODO: Spit out a 500 instead of nil?
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
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

		messages := bytes.Split(message, []byte("\r"))
		for _, msg := range messages {
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
			}
			/*if bytes.Equal(message,[]byte(`start-view`)) {
			} else if bytes.Equal(message,[]byte(`end-view`)) {
			}*/
		}
	}
	conn.Close()
	return nil
}

// TODO: Use a map instead of a switch to make this more modular?
func wsPageResponses(wsUser *WSUser, conn *websocket.Conn, page string) {
	if page == "/" {
		page = Config.DefaultPath
	}

	switch page {
	// Live Topic List is an experimental feature
	// TODO: Optimise this to reduce the amount of contention
	case "/topics/":
		topicListMutex.Lock()
		topicListWatchers[wsUser] = true
		topicListMutex.Unlock()
	case "/panel/":
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
	wsUser.SetPageForSocket(conn, page)
}

// TODO: Use a map instead of a switch to make this more modular?
func wsLeavePage(wsUser *WSUser, conn *websocket.Conn, page string) {
	if page == "/" {
		page = Config.DefaultPath
	}

	switch page {
	case "/topics/":
		wsUser.FinalizePage("/topics/", func() {
			topicListMutex.Lock()
			delete(topicListWatchers, wsUser)
			topicListMutex.Unlock()
		})
	case "/panel/":
		adminStatsMutex.Lock()
		delete(adminStatsWatchers, conn)
		adminStatsMutex.Unlock()
	}
	wsUser.SetPageForSocket(conn, "")
}

// TODO: Abstract this
// TODO: Use odd-even sharding
var topicListWatchers map[*WSUser]bool
var topicListMutex sync.RWMutex
var adminStatsWatchers map[*websocket.Conn]*WSUser
var adminStatsMutex sync.RWMutex

func adminStatsTicker() {
	time.Sleep(time.Second)

	var lastUonline = -1
	var lastGonline = -1
	var lastTotonline = -1
	var lastCPUPerc = -1
	var lastAvailableRAM int64 = -1
	var noStatUpdates, noRAMUpdates bool

	var onlineColour, onlineGuestsColour, onlineUsersColour, cpustr, cpuColour, ramstr, ramColour string
	var cpuerr, ramerr error
	var memres *mem.VirtualMemoryStat
	var cpuPerc []float64

	var totunit, uunit, gunit string

	lessThanSwitch := func(number int, lowerBound int, midBound int) string {
		switch {
		case number < lowerBound:
			return "stat_green"
		case number < midBound:
			return "stat_orange"
		}
		return "stat_red"
	}

	greaterThanSwitch := func(number int, lowerBound int, midBound int) string {
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
			if !noStatUpdates {
				w.Write([]byte("set #dash-totonline <span>" + strconv.Itoa(totonline) + totunit + " online</span>\r"))
				w.Write([]byte("set #dash-gonline <span>" + strconv.Itoa(gonline) + gunit + " guests online</span>\r"))
				w.Write([]byte("set #dash-uonline <span>" + strconv.Itoa(uonline) + uunit + " users online</span>\r"))
				w.Write([]byte("set #dash-reqs <span>" + strconv.Itoa(reqCount) + " reqs / second</span>\r"))

				w.Write([]byte("set-class #dash-totonline grid_item grid_stat " + onlineColour + "\r"))
				w.Write([]byte("set-class #dash-gonline grid_item grid_stat " + onlineGuestsColour + "\r"))
				w.Write([]byte("set-class #dash-uonline grid_item grid_stat " + onlineUsersColour + "\r"))
				//w.Write([]byte("set-class #dash-reqs grid_item grid_stat grid_end_group \r"))
			}

			w.Write([]byte("set #dash-cpu <span>CPU: " + cpustr + "%</span>\r"))
			w.Write([]byte("set-class #dash-cpu grid_item grid_istat " + cpuColour + "\r"))

			if !noRAMUpdates {
				w.Write([]byte("set #dash-ram <span>RAM: " + ramstr + "</span>\r"))
				w.Write([]byte("set-class #dash-ram grid_item grid_istat " + ramColour + "\r"))
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
