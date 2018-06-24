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
	"encoding/json"
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

// TODO: Make this an interface?
type WsHubImpl struct {
	// TODO: Shard this map
	OnlineUsers  map[int]*WSUser
	OnlineGuests map[*WSUser]bool
	GuestLock    sync.RWMutex
	UserLock     sync.RWMutex

	lastTick      time.Time
	lastTopicList []*TopicsRow
}

// TODO: Disable WebSockets on high load? Add a Control Panel interface for disabling it?
var EnableWebsockets = true // Put this in caps for consistency with the other constants?

// TODO: Rename this to WebSockets?
var WsHub WsHubImpl
var wsUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var errWsNouser = errors.New("This user isn't connected via WebSockets")

func init() {
	adminStatsWatchers = make(map[*WSUser]bool)
	topicListWatchers = make(map[*WSUser]bool)
	// TODO: Do we really want to initialise this here instead of in main.go / general_test.go like the other things?
	WsHub = WsHubImpl{
		OnlineUsers:  make(map[int]*WSUser),
		OnlineGuests: make(map[*WSUser]bool),
	}
}

func (hub *WsHubImpl) Start() {
	//fmt.Println("running hub.Start")
	if Config.DisableLiveTopicList {
		return
	}
	hub.lastTick = time.Now()
	AddScheduledSecondTask(hub.Tick)
}

type WsTopicList struct {
	Topics []*WsTopicsRow
}

// This Tick is seperate from the admin one, as we want to process that in parallel with this due to the blocking calls to gopsutil
func (hub *WsHubImpl) Tick() error {
	//fmt.Println("running hub.Tick")

	// Don't waste CPU time if nothing has happened
	// TODO: Get a topic list method which strips stickies?
	tList, _, _, err := TopicList.GetList(1)
	if err != nil {
		hub.lastTick = time.Now()
		return err // TODO: Do we get ErrNoRows here?
	}
	defer func() {
		hub.lastTick = time.Now()
		hub.lastTopicList = tList
	}()
	if len(tList) == 0 {
		return nil
	}

	//fmt.Println("checking for changes")
	// TODO: Optimise this by only sniffing the top non-sticky
	if len(tList) == len(hub.lastTopicList) {
		var hasItem = false
		for j, tItem := range tList {
			if !tItem.Sticky {
				if tItem.ID != hub.lastTopicList[j].ID {
					hasItem = true
				}
			}
		}
		if !hasItem {
			return nil
		}
	}

	// TODO: Implement this for guests too? Should be able to optimise it far better there due to them sharing the same permission set
	// TODO: Be less aggressive with the locking, maybe use an array of sorts instead of hitting the main map every-time
	topicListMutex.RLock()
	if len(topicListWatchers) == 0 {
		//fmt.Println("no watchers")
		topicListMutex.RUnlock()
		return nil
	}
	//fmt.Println("found changes")

	// Copy these over so we close this loop as fast as possible so we can release the read lock, especially if the group gets are backed by calls to the database
	var groupIDs = make(map[int]bool)
	var currentWatchers = make([]*WSUser, len(topicListWatchers))
	var i = 0
	for wsUser, _ := range topicListWatchers {
		currentWatchers[i] = wsUser
		groupIDs[wsUser.User.Group] = true
		i++
	}
	topicListMutex.RUnlock()

	var groups = make(map[int]*Group)
	var canSeeMap = make(map[string][]int)
	for groupID, _ := range groupIDs {
		group, err := Groups.Get(groupID)
		if err != nil {
			// TODO: Do we really want to halt all pushes for what is possibly just one user?
			return err
		}
		groups[group.ID] = group

		var canSee = make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}
		canSeeMap[string(canSee)] = group.CanSee
	}

	var canSeeRenders = make(map[string][]byte)
	for name, canSee := range canSeeMap {
		topicList, forumList, _, err := TopicList.GetListByCanSee(canSee, 1)
		if err != nil {
			return err // TODO: Do we get ErrNoRows here?
		}
		if len(topicList) == 0 {
			continue
		}
		_ = forumList // Might use this later after we get the base feature working

		//fmt.Println("canSeeItem")
		if topicList[0].Sticky {
			var lastSticky = 0
			for i, row := range topicList {
				if !row.Sticky {
					lastSticky = i
					break
				}
			}
			if lastSticky == 0 {
				continue
			}
			//fmt.Println("lastSticky: ", lastSticky)
			//fmt.Println("before topicList: ", topicList)
			topicList = topicList[lastSticky:]
			//fmt.Println("after topicList: ", topicList)
		}

		// TODO: Compare to previous tick to eliminate unnecessary work and data
		var wsTopicList = make([]*WsTopicsRow, len(topicList))
		for i, topicRow := range topicList {
			wsTopicList[i] = topicRow.WebSockets()
		}

		outBytes, err := json.Marshal(&WsTopicList{wsTopicList})
		if err != nil {
			return err
		}
		canSeeRenders[name] = outBytes
	}

	// TODO: Use MessagePack for additional speed?
	//fmt.Println("writing to the clients")
	for _, wsUser := range currentWatchers {
		group := groups[wsUser.User.Group]
		var canSee = make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}

		w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			//fmt.Printf("werr for #%d: %s\n", wsUser.User.ID, err)
			topicListMutex.Lock()
			delete(topicListWatchers, wsUser)
			topicListMutex.Unlock()
			continue
		}

		//fmt.Println("writing to user #", wsUser.User.ID)
		outBytes := canSeeRenders[string(canSee)]
		//fmt.Println("outBytes: ", string(outBytes))
		w.Write(outBytes)
		w.Close()
	}
	return nil
}

func (hub *WsHubImpl) GuestCount() int {
	defer hub.GuestLock.RUnlock()
	hub.GuestLock.RLock()
	return len(hub.OnlineGuests)
}

func (hub *WsHubImpl) UserCount() int {
	defer hub.UserLock.RUnlock()
	hub.UserLock.RLock()
	return len(hub.OnlineUsers)
}

func (hub *WsHubImpl) broadcastMessage(msg string) error {
	hub.UserLock.RLock()
	defer hub.UserLock.RUnlock()
	for _, wsUser := range hub.OnlineUsers {
		w, err := wsUser.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte(msg))
		w.Close()
	}
	return nil
}

func (hub *WsHubImpl) pushMessage(targetUser int, msg string) error {
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

func (hub *WsHubImpl) pushAlert(targetUser int, asid int, event string, elementType string, actorID int, targetUserID int, elementID int) error {
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

func (hub *WsHubImpl) pushAlerts(users []int, asid int, event string, elementType string, actorID int, targetUserID int, elementID int) error {
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
				// TODO: Make sure the admin is removed from the admin stats list in the case that an error happens
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

// TODO: Use a map instead of a switch to make this more modular?
func wsPageResponses(wsUser *WSUser, page []byte) {
	//fmt.Println("entering page: ", string(page))
	switch string(page) {
	// Live Topic List is an experimental feature
	// TODO: Optimise this to reduce the amount of contention
	case "/topics/":
		topicListMutex.Lock()
		topicListWatchers[wsUser] = true
		topicListMutex.Unlock()
	case "/panel/":
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

// TODO: Use a map instead of a switch to make this more modular?
func wsLeavePage(wsUser *WSUser, page []byte) {
	//fmt.Println("leaving page: ", string(page))
	switch string(page) {
	// Live Topic List is an experimental feature
	case "/topics/":
		topicListMutex.Lock()
		delete(topicListWatchers, wsUser)
		topicListMutex.Unlock()
	case "/panel/":
		adminStatsMutex.Lock()
		delete(adminStatsWatchers, wsUser)
		adminStatsMutex.Unlock()
	}
}

// TODO: Abstract this
// TODO: Use odd-even sharding
var topicListWatchers map[*WSUser]bool
var topicListMutex sync.RWMutex
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

		// Acquire a write lock for now, so we can handle the delete() case below and the read one simultaneously
		// TODO: Stop taking a write lock here if it isn't necessary
		adminStatsMutex.Lock()
		for watcher := range adminStatsWatchers {
			w, err := watcher.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				delete(adminStatsWatchers, watcher)
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

		//time.Sleep(time.Second)
	}
}
