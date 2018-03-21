// +build no_ws

package common

import "errors"
import "net/http"

// TODO: Disable WebSockets on high load? Add a Control Panel interface for disabling it?
var EnableWebsockets = false // Put this in caps for consistency with the other constants?

var wsHub WSHub
var errWsNouser = errors.New("This user isn't connected via WebSockets")

type WSHub struct{}

func (_ *WSHub) guestCount() int { return 0 }

func (_ *WSHub) userCount() int { return 0 }

func (hub *WSHub) broadcastMessage(_ string) error { return nil }

func (hub *WSHub) pushMessage(_ int, _ string) error {
	return errWsNouser
}

func (hub *WSHub) pushAlert(_ int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return errWsNouser
}

func (hub *WSHub) pushAlerts(_ []int, _ int, _ string, _ string, _ int, _ int, _ int) error {
	return errWsNouser
}

func RouteWebsockets(_ http.ResponseWriter, _ *http.Request, _ User) {}
