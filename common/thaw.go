package common

import (
	"sync/atomic"
)

var TopicListThaw ThawInt

type ThawInt interface {
	Thawed() bool
	Thaw()

	Tick() error
}

type SingleServerThaw struct {
	DefaultThaw
}

func NewSingleServerThaw() *SingleServerThaw {
	thaw := &SingleServerThaw{}
	if Config.ServerCount == 1 {
		AddScheduledSecondTask(thaw.Tick)
	}
	return thaw
}

func (thaw *SingleServerThaw) Thawed() bool {
	if Config.ServerCount == 1 {
		return thaw.DefaultThaw.Thawed()
	}
	return true
}

func (thaw *SingleServerThaw) Thaw() {
	if Config.ServerCount == 1 {
		thaw.DefaultThaw.Thaw()
	}
}

type DefaultThaw struct {
	thawed int64
}

func NewDefaultThaw() *DefaultThaw {
	thaw := &DefaultThaw{}
	AddScheduledSecondTask(thaw.Tick)
	return thaw
}

// Decrement the thawed counter once a second until it goes cold
func (thaw *DefaultThaw) Tick() error {
	prior := thaw.thawed
	if prior > 0 {
		atomic.StoreInt64(&thaw.thawed, prior-1)
	}
	return nil
}

func (thaw *DefaultThaw) Thawed() bool {
	return thaw.thawed > 0
}

func (thaw *DefaultThaw) Thaw() {
	atomic.StoreInt64(&thaw.thawed, 5)
}
