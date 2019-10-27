package common

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var ErrNoneOnPage = errors.New("This user isn't on that page")
var ErrInvalidSocket = errors.New("That's not a valid WebSocket Connection")

type WSUser struct {
	User    *User
	Sockets []*WSUserSocket
	sync.Mutex
}

type WSUserSocket struct {
	conn *websocket.Conn
	Page string
}

func (u *WSUser) Ping() error {
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		socket.conn.SetWriteDeadline(time.Now().Add(time.Minute))
		err := socket.conn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			socket.conn.Close()
		}
	}
	return nil
}

func (u *WSUser) WriteAll(msg string) error {
	msgbytes := []byte(msg)
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		w, err := socket.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		_, _ = w.Write(msgbytes)
		w.Close()
	}
	return nil
}

func (u *WSUser) WriteToPage(msg string, page string) error {
	return u.WriteToPageBytes([]byte(msg), page)
}

// Inefficient as it looks for sockets for a page even if there are none
func (u *WSUser) WriteToPageBytes(msg []byte, page string) error {
	var success bool
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		if socket.Page != page {
			continue
		}
		w, err := socket.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			continue // Skip dead sockets, a dedicated goroutine handles those
		}
		_, _ = w.Write(msg)
		w.Close()
		success = true
	}
	if !success {
		return ErrNoneOnPage
	}
	return nil
}

func (u *WSUser) AddSocket(conn *websocket.Conn, page string) {
	u.Lock()
	// If the number of the sockets is small, then we can keep the size of the slice mostly static and just walk through it looking for empty slots
	if len(u.Sockets) < 6 {
		for i, socket := range u.Sockets {
			if socket == nil {
				u.Sockets[i] = &WSUserSocket{conn, page}
				u.Unlock()
				//fmt.Printf("%+v\n", u.Sockets)
				return
			}
		}
	}
	u.Sockets = append(u.Sockets, &WSUserSocket{conn, page})
	//fmt.Printf("%+v\n", u.Sockets)
	u.Unlock()
}

func (u *WSUser) RemoveSocket(conn *websocket.Conn) {
	u.Lock()
	defer u.Unlock()
	if len(u.Sockets) < 6 {
		for i, socket := range u.Sockets {
			if socket == nil {
				continue
			}
			if socket.conn == conn {
				u.Sockets[i] = nil
				//fmt.Printf("%+v\n", wsUser.Sockets)
				return
			}
		}
	}

	var key int
	for i, socket := range u.Sockets {
		if socket.conn == conn {
			key = i
			break
		}
	}
	u.Sockets = append(u.Sockets[:key], u.Sockets[key+1:]...)
	//fmt.Printf("%+v\n", u.Sockets)
}

func (u *WSUser) SetPageForSocket(conn *websocket.Conn, page string) error {
	if conn == nil {
		return ErrInvalidSocket
	}

	u.Lock()
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		if socket.conn == conn {
			socket.Page = page
		}
	}
	u.Unlock()

	return nil
}

func (u *WSUser) InPage(page string) bool {
	u.Lock()
	defer u.Unlock()
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		if socket.Page == page {
			return true
		}
	}
	return false
}

func (u *WSUser) FinalizePage(page string, handle func()) {
	u.Lock()
	defer u.Unlock()
	for _, socket := range u.Sockets {
		if socket == nil {
			continue
		}
		if socket.Page == page {
			return
		}
	}
	handle()
}
