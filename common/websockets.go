// +build !no_ws

/*
*
*	Gosora WebSocket Subsystem
*	Copyright Azareal 2017 - 2018
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

type WSUser struct {
	conn *websocket.Conn
	User *User
}

type WSHub struct {
	OnlineUsers  map[int]*WSUser
	OnlineGuests map[*WSUser]bool
	GuestLock    sync.RWMutex
	UserLock     sync.RWMutex
}

// TODO: Disable WebSockets on high load? Add a Control Panel interface for disabling it?
var EnableWebsockets = true // Put this in caps for consistency with the other constants?

var WsHub WSHub
var wsUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var errWsNouser = errors.New("This user isn't connected via WebSockets")

func init() {
	adminStatsWatchers = make(map[*WSUser]bool)
	WsHub = WSHub{
		OnlineUsers:  make(map[int]*WSUser),
		OnlineGuests: make(map[*WSUser]bool),
	}
}

func (hub *WSHub) GuestCount() int {
	defer hub.GuestLock.RUnlock()
	hub.GuestLock.RLock()
	return len(hub.OnlineGuests)
}

func (hub *WSHub) UserCount() int {
	defer hub.UserLock.RUnlock()
	hub.UserLock.RLock()
	return len(hub.OnlineUsers)
}

func (hub *WSHub) broadcastMessage(msg string) error {
	hub.UserLock.RLock()
	for _, wsUser := range hub.OnlineUsers {
		w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte(msg))
	}
	hub.UserLock.RUnlock()
	return nil
}

func (hub *WSHub) pushMessage(targetUser int, msg string) error {
	hub.UserLock.RLock()
	wsUser, ok := hub.OnlineUsers[targetUser]
	hub.UserLock.RUnlock()
	if !ok {
		return errWsNouser
	}

	w, err := wsUser.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	w.Write([]byte(msg))
	w.Close()
	return nil
}

func (hub *WSHub) pushAlert(targetUser int, asid int, event string, elementType string, actorID int, targetUserID int, elementID int) error {
	//log.Print("In pushAlert")
	hub.UserLock.RLock()
	wsUser, ok := hub.OnlineUsers[targetUser]
	hub.UserLock.RUnlock()
	if !ok {
		return errWsNouser
	}

	//log.Print("Building alert")
	alert, err := BuildAlert(asid, event, elementType, actorID, targetUserID, elementID, *wsUser.User)
	if err != nil {
		return err
	}

	//log.Print("Getting WS Writer")
	w, err := wsUser.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	w.Write([]byte(alert))
	_ = w.Close()
	return nil
}

func (hub *WSHub) pushAlerts(users []int, asid int, event string, elementType string, actorID int, targetUserID int, elementID int) error {
	var wsUsers []*WSUser
	hub.UserLock.RLock()
	// We don't want to keep a lock on this for too long, so we'll accept some nil pointers
	for _, uid := range users {
		wsUsers = append(wsUsers, hub.OnlineUsers[uid])
	}
	hub.UserLock.RUnlock()
	if len(wsUsers) == 0 {
		return errWsNouser
	}

	var errs []error
	for _, wsUser := range wsUsers {
		if wsUser == nil {
			continue
		}

		alert, err := BuildAlert(asid, event, elementType, actorID, targetUserID, elementID, *wsUser.User)
		if err != nil {
			errs = append(errs, err)
		}

		w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			errs = append(errs, err)
		}
		w.Write([]byte(alert))
		w.Close()
	}

	// Return the first error
	if len(errs) != 0 {
		for _, err := range errs {
			return err
		}
	}
	return nil
}

// TODO: How should we handle errors for this?
// TODO: Move this out of common?
func RouteWebsockets(w http.ResponseWriter, r *http.Request, user User) RouteError {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
	userptr, err := Users.Get(user.ID)
	if err != nil && err != ErrStoreCapacityOverflow {
		return nil
	}

	wsUser := &WSUser{conn, userptr}
	if user.ID == 0 {
		WsHub.GuestLock.Lock()
		WsHub.OnlineGuests[wsUser] = true
		WsHub.GuestLock.Unlock()
	} else {
		WsHub.UserLock.Lock()
		WsHub.OnlineUsers[user.ID] = wsUser
		WsHub.UserLock.Unlock()
	}

	//conn.SetReadLimit(/* put the max request size from earlier here? */)
	//conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var currentPage []byte
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if user.ID == 0 {
				WsHub.GuestLock.Lock()
				delete(WsHub.OnlineGuests, wsUser)
				WsHub.GuestLock.Unlock()
			} else {
				WsHub.UserLock.Lock()
				delete(WsHub.OnlineUsers, user.ID)
				WsHub.UserLock.Unlock()
			}
			break
		}

		//log.Print("Message", message)
		//log.Print("string(Message)", string(message))
		messages := bytes.Split(message, []byte("\r"))
		for _, msg := range messages {
			//log.Print("Submessage", msg)
			//log.Print("Submessage", string(msg))
			if bytes.HasPrefix(msg, []byte("page ")) {
				msgblocks := bytes.SplitN(msg, []byte(" "), 2)
				if len(msgblocks) < 2 {
					continue
				}

				if !bytes.Equal(msgblocks[1], currentPage) {
					wsLeavePage(wsUser, currentPage)
					currentPage = msgblocks[1]
					//log.Print("Current Page:", currentPage)
					//log.Print("Current Page:", string(currentPage))
					wsPageResponses(wsUser, currentPage)
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

func wsPageResponses(wsUser *WSUser, page []byte) {
	switch string(page) {
	case "/panel/":
		//log.Print("/panel/ WS Route")
		/*w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			//log.Print(err.Error())
			return
		}

		log.Print(WsHub.online_users)
		uonline := WsHub.UserCount()
		gonline := WsHub.GuestCount()
		totonline := uonline + gonline

		w.Write([]byte("set #dash-totonline " + strconv.Itoa(totonline) + " online\r"))
		w.Write([]byte("set #dash-gonline " + strconv.Itoa(gonline) + " guests online\r"))
		w.Write([]byte("set #dash-uonline " + strconv.Itoa(uonline) + " users online\r"))
		w.Close()*/

		// Listen for changes and inform the admins...
		adminStatsMutex.Lock()
		watchers := len(adminStatsWatchers)
		adminStatsWatchers[wsUser] = true
		if watchers == 0 {
			go adminStatsTicker()
		}
		adminStatsMutex.Unlock()
	}
}

func wsLeavePage(wsUser *WSUser, page []byte) {
	switch string(page) {
	case "/panel/":
		adminStatsMutex.Lock()
		delete(adminStatsWatchers, wsUser)
		adminStatsMutex.Unlock()
	}
}

var adminStatsWatchers map[*WSUser]bool
var adminStatsMutex sync.RWMutex

func adminStatsTicker() {
	time.Sleep(time.Second)

	var lastUonline = -1
	var lastGonline = -1
	var lastTotonline = -1
	var lastCPUPerc = -1
	var lastAvailableRAM int64 = -1
	var noStatUpdates bool
	var noRAMUpdates bool

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

		adminStatsMutex.RLock()
		watchers := adminStatsWatchers
		adminStatsMutex.RUnlock()

		for watcher := range watchers {
			w, err := watcher.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				adminStatsMutex.Lock()
				delete(adminStatsWatchers, watcher)
				adminStatsMutex.Unlock()
				continue
			}

			// nolint
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

		lastUonline = uonline
		lastGonline = gonline
		lastTotonline = totonline
		lastCPUPerc = int(cpuPerc[0])
		lastAvailableRAM = int64(memres.Available)

		//time.Sleep(time.Second)
	}
}
