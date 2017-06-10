// +build !no_ws

package main

import(
	"fmt"
	"sync"
	"time"
	"bytes"
	"strconv"
	"errors"
	"runtime"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type WS_User struct
{
	conn *websocket.Conn
	User *User
}

type WS_Hub struct
{
	online_users map[int]*WS_User
	online_guests map[*WS_User]bool
	guests sync.RWMutex
	users sync.RWMutex
}

var ws_hub WS_Hub
var ws_upgrader = websocket.Upgrader{ReadBufferSize:1024,WriteBufferSize:1024}
var ws_nouser error = errors.New("This user isn't connected via WebSockets")

func init() {
	enable_websockets = true
	admin_stats_watchers = make(map[*WS_User]bool)
	ws_hub = WS_Hub{
		online_users: make(map[int]*WS_User),
		online_guests: make(map[*WS_User]bool),
	}
}

func (hub *WS_Hub) guest_count() int {
	defer hub.guests.RUnlock()
	hub.guests.RLock()
	return len(hub.online_guests)
}

func (hub *WS_Hub) user_count() int {
	defer hub.users.RUnlock()
	hub.users.RLock()
	return len(hub.online_users)
}

func (hub *WS_Hub) broadcast_message(msg string) error {
	hub.users.RLock()
	for _, ws_user := range hub.online_users {
		w, err := ws_user.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		w.Write([]byte(msg))
	}
	hub.users.RUnlock()
	return nil
}

func (hub *WS_Hub) push_message(targetUser int, msg string) error {
	hub.users.RLock()
	ws_user, ok := hub.online_users[targetUser]
	hub.users.RUnlock()
	if !ok {
		return ws_nouser
	}

	w, err := ws_user.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	w.Write([]byte(msg))
	w.Close()
	return nil
}

func(hub *WS_Hub) push_alert(targetUser int, event string, elementType string, actor_id int, targetUser_id int, elementID int) error {
	//fmt.Println("In push_alert")
	hub.users.RLock()
	ws_user, ok := hub.online_users[targetUser]
	hub.users.RUnlock()
	if !ok {
		return ws_nouser
	}

	//fmt.Println("Building alert")
	alert, err := build_alert(event, elementType, actor_id, targetUser_id, elementID, *ws_user.User)
	if err != nil {
		return err
	}

	//fmt.Println("Getting WS Writer")
	w, err := ws_user.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	//fmt.Println("Writing to the client")
	w.Write([]byte(alert))
	w.Close()
	return nil
}

func route_websockets(w http.ResponseWriter, r *http.Request) {
	user, ok := SimpleSessionCheck(w,r)
	if !ok {
		return
	}
	conn, err := ws_upgrader.Upgrade(w,r,nil)
	if err != nil {
		return
	}
	userptr, err := users.CascadeGet(user.ID)
	if err != nil && err != ErrStoreCapacityOverflow {
		return
	}

	ws_user := &WS_User{conn,userptr}
	if user.ID == 0 {
		ws_hub.guests.Lock()
		ws_hub.online_guests[ws_user] = true
		ws_hub.guests.Unlock()
	} else {
		ws_hub.users.Lock()
		ws_hub.online_users[user.ID] = ws_user
		ws_hub.users.Unlock()
	}

	//conn.SetReadLimit(/* put the max request size from earlier here? */)
	//conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var current_page []byte
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if user.ID == 0 {
				ws_hub.guests.Lock()
				delete(ws_hub.online_guests,ws_user)
				ws_hub.guests.Unlock()
			} else {
				ws_hub.users.Lock()
				delete(ws_hub.online_users,user.ID)
				ws_hub.users.Unlock()
			}
			break
		}

		//fmt.Println("Message",message)
		//fmt.Println("Message",string(message))
		messages := bytes.Split(message,[]byte("\r"))
		for _, msg := range messages {
			//fmt.Println("Submessage",msg)
			//fmt.Println("Submessage",string(msg))
			if bytes.HasPrefix(msg,[]byte("page ")) {
				msgblocks := bytes.SplitN(msg,[]byte(" "),2)
				if len(msgblocks) < 2 {
					continue
				}

				if !bytes.Equal(msgblocks[1],current_page) {
					ws_leave_page(ws_user, current_page)
					current_page = msgblocks[1]
					//fmt.Println("Current Page: ",current_page)
					//fmt.Println("Current Page: ",string(current_page))
					ws_page_responses(ws_user, current_page)
				}
			}
			/*if bytes.Equal(message,[]byte(`start-view`)) {

			} else if bytes.Equal(message,[]byte(`end-view`)) {

			}*/
		}
	}
	conn.Close()
}

func ws_page_responses(ws_user *WS_User, page []byte) {
	switch(string(page)) {
		case "/panel/":
			//fmt.Println("/panel/ WS Route")
			/*w, err := ws_user.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				//fmt.Println(err.Error())
				return
			}

			fmt.Println(ws_hub.online_users)
			uonline := ws_hub.user_count()
			gonline := ws_hub.guest_count()
			totonline := uonline + gonline

			w.Write([]byte("set #dash-totonline " + strconv.Itoa(totonline) + " online\r"))
			w.Write([]byte("set #dash-gonline " + strconv.Itoa(gonline) + " guests online\r"))
			w.Write([]byte("set #dash-uonline " + strconv.Itoa(uonline) + " users online\r"))
			w.Close()*/

			// Listen for changes and inform the admins...
			admin_stats_mutex.Lock()
			watchers := len(admin_stats_watchers)
			admin_stats_watchers[ws_user] = true
			if watchers == 0 {
				go admin_stats_ticker()
			}
			admin_stats_mutex.Unlock()
	}
}

func ws_leave_page(ws_user *WS_User, page []byte) {
	switch(string(page)) {
		case "/panel/":
			admin_stats_mutex.Lock()
			delete(admin_stats_watchers,ws_user)
			admin_stats_mutex.Unlock()
	}
}

var admin_stats_watchers map[*WS_User]bool
var admin_stats_mutex sync.RWMutex
func admin_stats_ticker() {
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
		admin_stats_mutex.RLock()
		watch_count := len(admin_stats_watchers)
		admin_stats_mutex.RUnlock()
		if watch_count == 0 {
			break AdminStatLoop
		}

		cpu_perc, cpuerr = cpu.Percent(time.Duration(time.Second),true)
		memres, ramerr = mem.VirtualMemory()
		uonline := ws_hub.user_count()
		gonline := ws_hub.guest_count()
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

			totonline, totunit = convert_friendly_unit(totonline)
			uonline, uunit = convert_friendly_unit(uonline)
			gonline, gunit = convert_friendly_unit(gonline)
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
				total_count, total_unit := convert_byte_unit(float64(memres.Total))
				used_count := convert_byte_in_unit(float64(memres.Total - memres.Available),total_unit)

				// Round totals with .9s up, it's how most people see it anyway. Floats are notoriously imprecise, so do it off 0.85
				var totstr string
				if (total_count - float64(int(total_count))) > 0.85 {
					used_count += 1.0 - (total_count - float64(int(total_count)))
					totstr = strconv.Itoa(int(total_count) + 1)
				} else {
					totstr = fmt.Sprintf("%.1f",total_count)
				}

				if used_count > total_count {
					used_count = total_count
				}
				ramstr = fmt.Sprintf("%.1f",used_count) + " / " + totstr + total_unit

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

		admin_stats_mutex.RLock()
		watchers := admin_stats_watchers
		admin_stats_mutex.RUnlock()

		for watcher, _ := range watchers {
			w, err := watcher.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				//fmt.Println(err.Error())
				admin_stats_mutex.Lock()
				delete(admin_stats_watchers,watcher)
				admin_stats_mutex.Unlock()
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
