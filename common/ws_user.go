package common

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrNoneOnPage = errors.New("This user isn't on that page")

type WSUser struct {
	User    *User
	Sockets []*WSUserSocket
	sync.Mutex
}

type WSUserSocket struct {
	conn *websocket.Conn
	Page string
}

func (wsUser *WSUser) WriteAll(msg string) error {
	msgbytes := []byte(msg)
	for _, socket := range wsUser.Sockets {
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

func (wsUser *WSUser) WriteToPage(msg string, page string) error {
	return wsUser.WriteToPageBytes([]byte(msg), page)
}

// Inefficient as it looks for sockets for a page even if there are none
func (wsUser *WSUser) WriteToPageBytes(msg []byte, page string) error {
	var success bool
	for _, socket := range wsUser.Sockets {
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

func (wsUser *WSUser) AddSocket(conn *websocket.Conn, page string) {
	wsUser.Lock()
	// If the number of the sockets is small, then we can keep the size of the slice mostly static and just walk through it looking for empty slots
	if len(wsUser.Sockets) < 6 {
		for i, socket := range wsUser.Sockets {
			if socket == nil {
				wsUser.Sockets[i] = &WSUserSocket{conn, page}
				wsUser.Unlock()
				return
			}
		}
	}
	wsUser.Sockets = append(wsUser.Sockets, &WSUserSocket{conn, page})
	wsUser.Unlock()
}

func (wsUser *WSUser) RemoveSocket(conn *websocket.Conn) {
	wsUser.Lock()
	if len(wsUser.Sockets) < 6 {
		for i, socket := range wsUser.Sockets {
			if socket.conn == conn {
				wsUser.Sockets[i] = nil
				wsUser.Unlock()
				return
			}
		}
	}

	var key int
	for i, socket := range wsUser.Sockets {
		if socket.conn == conn {
			key = i
			break
		}
	}
	wsUser.Sockets = append(wsUser.Sockets[:key], wsUser.Sockets[key+1:]...)

	wsUser.Unlock()
}

func (wsUser *WSUser) SetPageForSocket(conn *websocket.Conn, page string) {
	wsUser.Lock()
	for _, socket := range wsUser.Sockets {
		if socket.conn == conn {
			socket.Page = page
		}
	}
	wsUser.Unlock()
}

func (wsUser *WSUser) InPage(page string) bool {
	wsUser.Lock()
	defer wsUser.Unlock()
	for _, socket := range wsUser.Sockets {
		if socket.Page == page {
			return true
		}
	}
	return false
}

func (wsUser *WSUser) FinalizePage(page string, handle func()) {
	wsUser.Lock()
	defer wsUser.Unlock()
	for _, socket := range wsUser.Sockets {
		if socket.Page == page {
			return
		}
	}
	handle()
}
