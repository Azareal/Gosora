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
	t := &SingleServerThaw{}
	if Config.ServerCount == 1 {
		AddScheduledSecondTask(t.Tick)
	}
	return t
}

func (t *SingleServerThaw) Thawed() bool {
	if Config.ServerCount == 1 {
		return t.DefaultThaw.Thawed()
	}
	return true
}

func (t *SingleServerThaw) Thaw() {
	if Config.ServerCount == 1 {
		t.DefaultThaw.Thaw()
	}
}

type DefaultThaw struct {
	thawed int64
}

func NewDefaultThaw() *DefaultThaw {
	t := &DefaultThaw{}
	AddScheduledSecondTask(t.Tick)
	return t
}

// Decrement the thawed counter once a second until it goes cold
func (t *DefaultThaw) Tick() error {
	prior := t.thawed
	if prior > 0 {
		atomic.StoreInt64(&t.thawed, prior-1)
	}
	return nil
}

func (t *DefaultThaw) Thawed() bool {
	return t.thawed > 0
}

func (t *DefaultThaw) Thaw() {
	atomic.StoreInt64(&t.thawed, 4)
}
