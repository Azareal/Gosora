package common

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TODO: Rename this to WebSockets?
var WsHub WsHubImpl

// TODO: Make this an interface?
// TODO: Write tests for this
type WsHubImpl struct {
	// TODO: Implement some form of generics so we don't write as much odd-even sharding code
	evenOnlineUsers map[int]*WSUser
	oddOnlineUsers  map[int]*WSUser
	evenUserLock    sync.RWMutex
	oddUserLock     sync.RWMutex

	// TODO: Add sharding for this too?
	OnlineGuests map[*WSUser]bool
	GuestLock    sync.RWMutex

	lastTick      time.Time
	lastTopicList []*TopicsRow
}

func init() {
	// TODO: Do we really want to initialise this here instead of in main.go / general_test.go like the other things?
	WsHub = WsHubImpl{
		evenOnlineUsers: make(map[int]*WSUser),
		oddOnlineUsers:  make(map[int]*WSUser),
		OnlineGuests:    make(map[*WSUser]bool),
	}
}

func (h *WsHubImpl) Start() {
	log.Print("Setting up the WebSocket ticks")
	ticker := time.NewTicker(time.Minute * 5)
	defer func() {
		ticker.Stop()
	}()

	go func() {
		for {
			item := func(l *sync.RWMutex, userMap map[int]*WSUser) {
				l.RLock()
				defer l.RUnlock()
				// TODO: Copy to temporary slice for less contention?
				for _, user := range userMap {
					user.Ping()
				}
			}
			select {
			case <-ticker.C:
				item(&h.evenUserLock, h.evenOnlineUsers)
				item(&h.oddUserLock, h.oddOnlineUsers)
			}
		}
	}()
	if Config.DisableLiveTopicList {
		return
	}
	h.lastTick = time.Now()
	AddScheduledSecondTask(h.Tick)
}

// This Tick is separate from the admin one, as we want to process that in parallel with this due to the blocking calls to gopsutil
func (h *WsHubImpl) Tick() error {
	return wsTopicListTick(h)
}

func wsTopicListTick(h *WsHubImpl) error {
	// Avoid hitting GetList when the topic list hasn't changed
	if !TopicListThaw.Thawed() && h.lastTopicList != nil {
		return nil
	}
	tickStart := time.Now()

	// Don't waste CPU time if nothing has happened
	// TODO: Get a topic list method which strips stickies?
	tList, _, _, err := TopicList.GetList(1, "", nil)
	if err != nil {
		h.lastTick = tickStart
		return err // TODO: Do we get ErrNoRows here?
	}
	defer func() {
		h.lastTick = tickStart
		h.lastTopicList = tList
	}()
	if len(tList) == 0 {
		return nil
	}

	// TODO: Optimise this by only sniffing the top non-sticky
	// TODO: Optimise this by getting back an unsorted list so we don't have to hop around the stickies
	// TODO: Add support for new stickies / replies to them
	if len(tList) == len(h.lastTopicList) {
		hasItem := false
		for j, tItem := range tList {
			if !tItem.Sticky {
				if tItem.ID != h.lastTopicList[j].ID || !tItem.LastReplyAt.Equal(h.lastTopicList[j].LastReplyAt) {
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
		topicListMutex.RUnlock()
		return nil
	}

	// Copy these over so we close this loop as fast as possible so we can release the read lock, especially if the group gets are backed by calls to the database
	groupIDs := make(map[int]bool)
	currentWatchers := make([]*WSUser, len(topicListWatchers))
	i := 0
	for wsUser, _ := range topicListWatchers {
		currentWatchers[i] = wsUser
		groupIDs[wsUser.User.Group] = true
		i++
	}
	topicListMutex.RUnlock()

	groups := make(map[int]*Group)
	canSeeMap := make(map[string][]int)
	for groupID, _ := range groupIDs {
		group, err := Groups.Get(groupID)
		if err != nil {
			// TODO: Do we really want to halt all pushes for what is possibly just one user?
			return err
		}
		groups[group.ID] = group

		canSee := make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}
		canSeeMap[string(canSee)] = group.CanSee
	}

	canSeeRenders := make(map[string][]byte)
	for name, canSee := range canSeeMap {
		topicList, forumList, _, err := TopicList.GetListByCanSee(canSee, 1, "", nil)
		if err != nil {
			return err // TODO: Do we get ErrNoRows here?
		}
		if len(topicList) == 0 {
			continue
		}
		_ = forumList // Might use this later after we get the base feature working

		if topicList[0].Sticky {
			lastSticky := 0
			for i, row := range topicList {
				if !row.Sticky {
					lastSticky = i
					break
				}
			}
			if lastSticky == 0 {
				continue
			}
			topicList = topicList[lastSticky:]
		}

		// TODO: Compare to previous tick to eliminate unnecessary work and data
		wsTopicList := make([]*WsTopicsRow, len(topicList))
		for i, topicRow := range topicList {
			wsTopicList[i] = topicRow.WebSockets()
		}

		outBytes, err := json.Marshal(&WsTopicList{wsTopicList, 0, tickStart.Unix()})
		if err != nil {
			return err
		}
		canSeeRenders[name] = outBytes
	}

	// TODO: Use MessagePack for additional speed?
	//fmt.Println("writing to the clients")
	for _, wsUser := range currentWatchers {
		group := groups[wsUser.User.Group]
		canSee := make([]byte, len(group.CanSee))
		for i, item := range group.CanSee {
			canSee[i] = byte(item)
		}

		//fmt.Println("writing to user #", wsUser.User.ID)
		outBytes := canSeeRenders[string(canSee)]
		//fmt.Println("outBytes: ", string(outBytes))
		err := wsUser.WriteToPageBytes(outBytes, "/topics/")
		if err == ErrNoneOnPage {
			//fmt.Printf("werr for #%d: %s\n", wsUser.User.ID, err)
			wsUser.FinalizePage("/topics/", func() {
				topicListMutex.Lock()
				delete(topicListWatchers, wsUser)
				topicListMutex.Unlock()
			})
			continue
		}
	}
	return nil
}

func (h *WsHubImpl) GuestCount() int {
	h.GuestLock.RLock()
	defer h.GuestLock.RUnlock()
	return len(h.OnlineGuests)
}

func (h *WsHubImpl) UserCount() (count int) {
	h.evenUserLock.RLock()
	count += len(h.evenOnlineUsers)
	h.evenUserLock.RUnlock()

	h.oddUserLock.RLock()
	count += len(h.oddOnlineUsers)
	h.oddUserLock.RUnlock()
	return count
}

func (h *WsHubImpl) HasUser(uid int) (exists bool) {
	h.evenUserLock.RLock()
	_, exists = h.evenOnlineUsers[uid]
	h.evenUserLock.RUnlock()
	if exists {
		return exists
	}
	
	h.oddUserLock.RLock()
	_, exists = h.oddOnlineUsers[uid]
	h.oddUserLock.RUnlock()
	return exists
}

func (h *WsHubImpl) broadcastMessage(msg string) error {
	userLoop := func(users map[int]*WSUser, m *sync.RWMutex) error {
		m.RLock()
		defer m.RUnlock()
		for _, wsUser := range users {
			err := wsUser.WriteAll(msg)
			if err != nil {
				return err
			}
		}
		return nil
	}
	// TODO: Can we move this RLock inside the closure safely?
	err := userLoop(h.evenOnlineUsers, &h.evenUserLock)
	if err != nil {
		return err
	}
	return userLoop(h.oddOnlineUsers, &h.oddUserLock)
}

func (h *WsHubImpl) getUser(uid int) (wsUser *WSUser, err error) {
	var ok bool
	if uid%2 == 0 {
		h.evenUserLock.RLock()
		wsUser, ok = h.evenOnlineUsers[uid]
		h.evenUserLock.RUnlock()
	} else {
		h.oddUserLock.RLock()
		wsUser, ok = h.oddOnlineUsers[uid]
		h.oddUserLock.RUnlock()
	}
	if !ok {
		return nil, errWsNouser
	}
	return wsUser, nil
}

// Warning: For efficiency, some of the *WSUsers may be nil pointers, DO NOT EXPORT
func (h *WsHubImpl) getUsers(uids []int) (wsUsers []*WSUser, err error) {
	if len(uids) == 0 {
		return nil, errWsNouser
	}
	appender := func(l *sync.RWMutex, users map[int]*WSUser) {
		l.RLock()
		defer l.RUnlock()
		// We don't want to keep a lock on this for too long, so we'll accept some nil pointers
		for _, uid := range uids {
			wsUsers = append(wsUsers, users[uid])
		}
	}
	appender(&h.evenUserLock, h.evenOnlineUsers)
	appender(&h.oddUserLock, h.oddOnlineUsers)
	if len(wsUsers) == 0 {
		return nil, errWsNouser
	}
	return wsUsers, nil
}

// For Widget WOL, please avoid using this as it might wind up being really long and slow without the right safeguards
func (h *WsHubImpl) AllUsers() (users []*User) {
	appender := func(l *sync.RWMutex, userMap map[int]*WSUser) {
		l.RLock()
		defer l.RUnlock()
		for _, user := range userMap {
			users = append(users, user.User)
		}
	}
	appender(&h.evenUserLock, h.evenOnlineUsers)
	appender(&h.oddUserLock, h.oddOnlineUsers)
	return users
}

func (h *WsHubImpl) removeUser(uid int) {
	if uid%2 == 0 {
		h.evenUserLock.Lock()
		delete(h.evenOnlineUsers, uid)
		h.evenUserLock.Unlock()
	} else {
		h.oddUserLock.Lock()
		delete(h.oddOnlineUsers, uid)
		h.oddUserLock.Unlock()
	}
}

func (h *WsHubImpl) AddConn(user User, conn *websocket.Conn) (*WSUser, error) {
	if user.ID == 0 {
		wsUser := new(WSUser)
		wsUser.User = new(User)
		*wsUser.User = user
		wsUser.AddSocket(conn, "")
		WsHub.GuestLock.Lock()
		WsHub.OnlineGuests[wsUser] = true
		WsHub.GuestLock.Unlock()
		return wsUser, nil
	}

	// TODO: How should we handle user state changes if we're holding a pointer which never changes?
	userptr, err := Users.Get(user.ID)
	if err != nil && err != ErrStoreCapacityOverflow {
		return nil, err
	}

	var mutex *sync.RWMutex
	var theMap map[int]*WSUser
	if user.ID%2 == 0 {
		mutex = &h.evenUserLock
		theMap = h.evenOnlineUsers
	} else {
		mutex = &h.oddUserLock
		theMap = h.oddOnlineUsers
	}

	mutex.Lock()
	wsUser, ok := theMap[user.ID]
	if !ok {
		wsUser = new(WSUser)
		wsUser.User = userptr
		wsUser.Sockets = []*WSUserSocket{&WSUserSocket{conn, ""}}
		theMap[user.ID] = wsUser
		mutex.Unlock()
		return wsUser, nil
	}
	mutex.Unlock()
	wsUser.AddSocket(conn, "")
	return wsUser, nil
}

func (h *WsHubImpl) RemoveConn(wsUser *WSUser, conn *websocket.Conn) {
	wsUser.RemoveSocket(conn)
	wsUser.Lock()
	if len(wsUser.Sockets) == 0 {
		h.removeUser(wsUser.User.ID)
	}
	wsUser.Unlock()
}

func (h *WsHubImpl) PushMessage(targetUser int, msg string) error {
	wsUser, err := h.getUser(targetUser)
	if err != nil {
		return err
	}
	return wsUser.WriteAll(msg)
}

func (h *WsHubImpl) pushAlert(targetUser int, alert Alert) error {
	wsUser, err := h.getUser(targetUser)
	if err != nil {
		return err
	}
	astr, err := BuildAlert(alert, *wsUser.User)
	if err != nil {
		return err
	}
	return wsUser.WriteAll(astr)
}

func (h *WsHubImpl) pushAlerts(users []int, alert Alert) error {
	wsUsers, err := h.getUsers(users)
	if err != nil {
		return err
	}

	var errs []error
	for _, wsUser := range wsUsers {
		if wsUser == nil {
			continue
		}
		alert, err := BuildAlert(alert, *wsUser.User)
		if err != nil {
			errs = append(errs, err)
		}
		err = wsUser.WriteAll(alert)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Return the first error
	if len(errs) != 0 {
		for _, err := range errs {
			return err
		}
	}
	return nil
}
