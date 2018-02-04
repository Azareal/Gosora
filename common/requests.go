package common

import (
	"sync"
	"sync/atomic"
)

// Add ReferrerItems here after they've had zero views for a while
var referrersToDelete = make(map[string]ReferrerDeletable)

type ReferrerDeletable struct {
	item        *ReferrerItem
	scheduledAt int64 //unixtime
}

type ReferrerItem struct {
	Counter int64
}

// ? We'll track referrer domains here rather than the exact URL they arrived from for now, we'll think about expanding later
// ? Referrers are fluid and ever-changing so we have to use string keys rather than 'enum' ints
type DefaultReferrerTracker struct {
	odd      map[string]*ReferrerItem
	even     map[string]*ReferrerItem
	oddLock  sync.RWMutex
	evenLock sync.RWMutex
}

func NewDefaultReferrerTracker() *DefaultReferrerTracker {
	return &DefaultReferrerTracker{
		odd:  make(map[string]*ReferrerItem),
		even: make(map[string]*ReferrerItem),
	}
}

func (ref *DefaultReferrerTracker) Tick() (err error) {
	for _, del := range referrersToDelete {
		_ = del
		// TODO: Calculate the gap between now and the times they were scheduled
	}
	// TODO: Run the queries and schedule zero view refs for deletion from memory
	return nil
}

func (ref *DefaultReferrerTracker) Bump(referrer string) {
	if referrer == "" {
		return
	}
	var refItem *ReferrerItem

	// Slightly crude and rudimentary, but it should give a basic degree of sharding
	if referrer[0]%2 == 0 {
		ref.evenLock.RLock()
		refItem = ref.even[referrer]
		ref.evenLock.RUnlock()
		if ref != nil {
			atomic.AddInt64(&refItem.Counter, 1)
		} else {
			ref.evenLock.Lock()
			ref.even[referrer] = &ReferrerItem{Counter: 1}
			ref.evenLock.Unlock()
		}
	} else {
		ref.oddLock.RLock()
		refItem = ref.odd[referrer]
		ref.oddLock.RUnlock()
		if ref != nil {
			atomic.AddInt64(&refItem.Counter, 1)
		} else {
			ref.oddLock.Lock()
			ref.odd[referrer] = &ReferrerItem{Counter: 1}
			ref.oddLock.Unlock()
		}
	}
}
