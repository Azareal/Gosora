// +build !no_ws

package main

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

type WS_User struct {
	conn *websocket.Conn
	User *User
}

type WS_Hub struct {
	onlineUsers  map[int]*WS_User
	onlineGuests map[*WS_User]bool
	guests       sync.RWMutex
	users        sync.RWMutex
}

var wsHub WS_Hub
var wsUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var errWsNouser = errors.New("This user isn't connected via WebSockets")

func init() {
	enableWebsockets = true
	adminStatsWatchers = make(map[*WS_User]bool)
	wsHub = WS_Hub{
		onlineUsers:  make(map[int]*WS_User),
		onlineGuests: make(map[*WS_User]bool),
	}
}

func (hub *WS_Hub) guestCount() int {
	defer hub.guests.RUnlock()
	hub.guests.RLock()
	return len(hub.onlineGuests)
}

func (hub *WS_Hub) userCount() int {
	defer hub.users.RUnlock()
	hub.users.RLock()
	return len(hub.onlineUsers)
}

func (hub *WS_Hub) broadcastMessage(msg string) error {
	hub.users.RLock()
	for _, wsUser := range hub.onlineUsers {
		w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		w.Write([]byte(msg))
	}
	hub.users.RUnlock()
	return nil
}

func (hub *WS_Hub) pushMessage(targetUser int, msg string) error {
	hub.users.RLock()
	wsUser, ok := hub.onlineUsers[targetUser]
	hub.users.RUnlock()
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

func (hub *WS_Hub) pushAlert(targetUser int, asid int, event string, elementType string, actorID int, targetUser_id int, elementID int) error {
	//log.Print("In push_alert")
	hub.users.RLock()
	wsUser, ok := hub.onlineUsers[targetUser]
	hub.users.RUnlock()
	if !ok {
		return errWsNouser
	}

	//log.Print("Building alert")
	alert, err := buildAlert(asid, event, elementType, actorID, targetUser_id, elementID, *wsUser.User)
	if err != nil {
		return err
	}

	//log.Print("Getting WS Writer")
	w, err := wsUser.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	w.Write([]byte(alert))
	w.Close()
	return nil
}

func (hub *WS_Hub) pushAlerts(users []int, asid int, event string, elementType string, actorID int, targetUserID int, elementID int) error {
	//log.Print("In pushAlerts")
	var wsUsers []*WS_User
	hub.users.RLock()
	// We don't want to keep a lock on this for too long, so we'll accept some nil pointers
	for _, uid := range users {
		wsUsers = append(wsUsers, hub.onlineUsers[uid])
	}
	hub.users.RUnlock()
	if len(wsUsers) == 0 {
		return errWsNouser
	}

	var errs []error
	for _, wsUser := range wsUsers {
		if wsUser == nil {
			continue
		}

		//log.Print("Building alert")
		alert, err := buildAlert(asid, event, elementType, actorID, targetUserID, elementID, *wsUser.User)
		if err != nil {
			errs = append(errs, err)
		}

		//log.Print("Getting WS Writer")
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

func route_websockets(w http.ResponseWriter, r *http.Request, user User) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	userptr, err := users.CascadeGet(user.ID)
	if err != nil && err != ErrStoreCapacityOverflow {
		return
	}

	wsUser := &WS_User{conn, userptr}
	if user.ID == 0 {
		wsHub.guests.Lock()
		wsHub.onlineGuests[wsUser] = true
		wsHub.guests.Unlock()
	} else {
		wsHub.users.Lock()
		wsHub.onlineUsers[user.ID] = wsUser
		wsHub.users.Unlock()
	}

	//conn.SetReadLimit(/* put the max request size from earlier here? */)
	//conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var currentPage []byte
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if user.ID == 0 {
				wsHub.guests.Lock()
				delete(wsHub.onlineGuests, wsUser)
				wsHub.guests.Unlock()
			} else {
				wsHub.users.Lock()
				delete(wsHub.onlineUsers, user.ID)
				wsHub.users.Unlock()
			}
			break
		}

		//log.Print("Message",message)
		//log.Print("string(Message)",string(message))
		messages := bytes.Split(message, []byte("\r"))
		for _, msg := range messages {
			//log.Print("Submessage",msg)
			//log.Print("Submessage",string(msg))
			if bytes.HasPrefix(msg, []byte("page ")) {
				msgblocks := bytes.SplitN(msg, []byte(" "), 2)
				if len(msgblocks) < 2 {
					continue
				}

				if !bytes.Equal(msgblocks[1], currentPage) {
					wsLeavePage(wsUser, currentPage)
					currentPage = msgblocks[1]
					//log.Print("Current Page:",currentPage)
					//log.Print("Current Page:",string(currentPage))
					wsPageResponses(wsUser, currentPage)
				}
			}
			/*if bytes.Equal(message,[]byte(`start-view`)) {

			} else if bytes.Equal(message,[]byte(`end-view`)) {

			}*/
		}
	}
	conn.Close()
}

func wsPageResponses(wsUser *WS_User, page []byte) {
	switch string(page) {
	case "/panel/":
		//log.Print("/panel/ WS Route")
		/*w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			//log.Print(err.Error())
			return
		}

		log.Print(wsHub.online_users)
		uonline := wsHub.userCount()
		gonline := wsHub.guestCount()
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

func wsLeavePage(wsUser *WS_User, page []byte) {
	switch string(page) {
	case "/panel/":
		adminStatsMutex.Lock()
		delete(adminStatsWatchers, wsUser)
		adminStatsMutex.Unlock()
	}
}

var adminStatsWatchers map[*WS_User]bool
var adminStatsMutex sync.RWMutex

func adminStatsTicker() {
	time.Sleep(time.Second)

	var last_uonline int = -1
	var last_gonline int = -1
	var last_totonline int = -1
	var last_cpu_perc int = -1
	var last_available_ram int64 = -1
	var no_stat_updates bool = false
	var no_ram_updates bool = false

	var onlineColour, onlineGuestsColour, onlineUsersColour, cpustr, cpuColour, ramstr, ramColour string
	var cpuerr, ramerr error
	var memres *mem.VirtualMemoryStat
	var cpu_perc []float64

	var totunit, uunit, gunit string

AdminStatLoop:
	for {
		adminStatsMutex.RLock()
		watch_count := len(adminStatsWatchers)
		adminStatsMutex.RUnlock()
		if watch_count == 0 {
			break AdminStatLoop
		}

		cpu_perc, cpuerr = cpu.Percent(time.Duration(time.Second), true)
		memres, ramerr = mem.VirtualMemory()
		uonline := wsHub.userCount()
		gonline := wsHub.guestCount()
		totonline := uonline + gonline

		// It's far more likely that the CPU Usage will change than the other stats, so we'll optimise them seperately...
		no_stat_updates = (uonline == last_uonline && gonline == last_gonline && totonline == last_totonline)
		no_ram_updates = (last_available_ram == int64(memres.Available))
		if int(cpu_perc[0]) == last_cpu_perc && no_stat_updates && no_ram_updates {
			time.Sleep(time.Second)
			continue
		}

		if !no_stat_updates {
			if totonline > 10 {
				onlineColour = "stat_green"
			} else if totonline > 3 {
				onlineColour = "stat_orange"
			} else {
				onlineColour = "stat_red"
			}

			if gonline > 10 {
				onlineGuestsColour = "stat_green"
			} else if gonline > 1 {
				onlineGuestsColour = "stat_orange"
			} else {
				onlineGuestsColour = "stat_red"
			}

			if uonline > 5 {
				onlineUsersColour = "stat_green"
			} else if uonline > 1 {
				onlineUsersColour = "stat_orange"
			} else {
				onlineUsersColour = "stat_red"
			}

			totonline, totunit = convertFriendlyUnit(totonline)
			uonline, uunit = convertFriendlyUnit(uonline)
			gonline, gunit = convertFriendlyUnit(gonline)
		}

		if cpuerr != nil {
			cpustr = "Unknown"
		} else {
			calcperc := int(cpu_perc[0]) / runtime.NumCPU()
			cpustr = strconv.Itoa(calcperc)
			if calcperc < 30 {
				cpuColour = "stat_green"
			} else if calcperc < 75 {
				cpuColour = "stat_orange"
			} else {
				cpuColour = "stat_red"
			}
		}

		if !no_ram_updates {
			if ramerr != nil {
				ramstr = "Unknown"
			} else {
				total_count, total_unit := convertByteUnit(float64(memres.Total))
				used_count := convertByteInUnit(float64(memres.Total-memres.Available), total_unit)

				// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
				var totstr string
				if (total_count - float64(int(total_count))) > 0.85 {
					used_count += 1.0 - (total_count - float64(int(total_count)))
					totstr = strconv.Itoa(int(total_count) + 1)
				} else {
					totstr = fmt.Sprintf("%.1f", total_count)
				}

				if used_count > total_count {
					used_count = total_count
				}
				ramstr = fmt.Sprintf("%.1f", used_count) + " / " + totstr + total_unit

				ramperc := ((memres.Total - memres.Available) * 100) / memres.Total
				if ramperc < 50 {
					ramColour = "stat_green"
				} else if ramperc < 75 {
					ramColour = "stat_orange"
				} else {
					ramColour = "stat_red"
				}
			}
		}

		adminStatsMutex.RLock()
		watchers := adminStatsWatchers
		adminStatsMutex.RUnlock()

		for watcher, _ := range watchers {
			w, err := watcher.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				//log.Print(err.Error())
				adminStatsMutex.Lock()
				delete(adminStatsWatchers, watcher)
				adminStatsMutex.Unlock()
				continue
			}

			if !no_stat_updates {
				w.Write([]byte("set #dash-totonline " + strconv.Itoa(totonline) + totunit + " online\r"))
				w.Write([]byte("set #dash-gonline " + strconv.Itoa(gonline) + gunit + " guests online\r"))
				w.Write([]byte("set #dash-uonline " + strconv.Itoa(uonline) + uunit + " users online\r"))

				w.Write([]byte("set-class #dash-totonline grid_item grid_stat " + onlineColour + "\r"))
				w.Write([]byte("set-class #dash-gonline grid_item grid_stat " + onlineGuestsColour + "\r"))
				w.Write([]byte("set-class #dash-uonline grid_item grid_stat " + onlineUsersColour + "\r"))
			}

			w.Write([]byte("set #dash-cpu CPU: " + cpustr + "%\r"))
			w.Write([]byte("set-class #dash-cpu grid_item grid_istat " + cpuColour + "\r"))

			if !no_ram_updates {
				w.Write([]byte("set #dash-ram RAM: " + ramstr + "\r"))
				w.Write([]byte("set-class #dash-ram grid_item grid_istat " + ramColour + "\r"))
			}

			w.Close()
		}

		last_uonline = uonline
		last_gonline = gonline
		last_totonline = totonline
		last_cpu_perc = int(cpu_perc[0])
		last_available_ram = int64(memres.Available)

		//time.Sleep(time.Second)
	}
}
