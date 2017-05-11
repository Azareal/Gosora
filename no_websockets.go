// +build no_ws

package main

import "net/http"

var ws_hub WS_Hub

type WS_Hub struct
{
}

func (_ *WS_Hub) GuestCount() int {
	return 0
}

func (_ *WS_Hub) UserCount() int {
	return 0
}

func route_websockets(_ http.ResponseWriter, _ *http.Request) {
}
