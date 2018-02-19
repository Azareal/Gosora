package counters

import "sync"

type RWMutexCounterBucket struct {
	counter int
	sync.RWMutex
}

// TODO: Make a neater API for this
var routeMapEnum map[string]int
var reverseRouteMapEnum map[int]string

func SetRouteMapEnum(rme map[string]int) {
	routeMapEnum = rme
}

func SetReverseRouteMapEnum(rrme map[int]string) {
	reverseRouteMapEnum = rrme
}

var agentMapEnum map[string]int
var reverseAgentMapEnum map[int]string

func SetAgentMapEnum(ame map[string]int) {
	agentMapEnum = ame
}

func SetReverseAgentMapEnum(rame map[int]string) {
	reverseAgentMapEnum = rame
}

var osMapEnum map[string]int
var reverseOSMapEnum map[int]string

func SetOSMapEnum(osme map[string]int) {
	osMapEnum = osme
}

func SetReverseOSMapEnum(rosme map[int]string) {
	reverseOSMapEnum = rosme
}
