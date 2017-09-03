// +build no_ws

package main

import "errors"
import "net/http"

var wsHub WS_Hub
var wsNouser error = errors.New("This user isn't connected via WebSockets")

type WS_Hub struct {
}

func (_ *WS_Hub) guest_count() int {
	return 0
}

func (_ *WS_Hub) user_count() int {
	return 0
}

func (hub *WS_Hub) broadcast_message(_ string) error {
	return nil
}

func (hub *WS_Hub) push_message(_ int, _ string) error {
	return wsNouser
}

func (hub *WS_Hub) push_alert(_ int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return wsNouser
}

func (hub *WS_Hub) push_alerts(_ []int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return wsNouser
}

func route_websockets(_ http.ResponseWriter, _ *http.Request, _ User) {
}
