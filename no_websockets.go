// +build no_ws

package main

import "errors"
import "net/http"

var wsHub WS_Hub
var errWsNouser = errors.New("This user isn't connected via WebSockets")

type WS_Hub struct {
}

func (_ *WS_Hub) guestCount() int {
	return 0
}

func (_ *WS_Hub) userCount() int {
	return 0
}

func (hub *WS_Hub) broadcastMessage(_ string) error {
	return nil
}

func (hub *WS_Hub) pushMessage(_ int, _ string) error {
	return errWsNouser
}

func (hub *WS_Hub) pushAlert(_ int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return errWsNouser
}

func (hub *WS_Hub) pushAlerts(_ []int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return errWsNouser
}

func route_websockets(_ http.ResponseWriter, _ *http.Request, _ User) {
}
