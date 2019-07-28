package counters

import (
	"database/sql"
	"sync"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var ReferrerTracker *DefaultReferrerTracker

// Add ReferrerItems here after they've had zero views for a while
var referrersToDelete = make(map[string]*ReferrerItem)

type ReferrerItem struct {
	Count int64
}

// ? We'll track referrer domains here rather than the exact URL they arrived from for now, we'll think about expanding later
// ? Referrers are fluid and ever-changing so we have to use string keys rather than 'enum' ints
type DefaultReferrerTracker struct {
	odd      map[string]*ReferrerItem
	even     map[string]*ReferrerItem
	oddLock  sync.RWMutex
	evenLock sync.RWMutex

	insert *sql.Stmt
}

func NewDefaultReferrerTracker() (*DefaultReferrerTracker, error) {
	acc := qgen.NewAcc()
	refTracker := &DefaultReferrerTracker{
		odd:    make(map[string]*ReferrerItem),
		even:   make(map[string]*ReferrerItem),
		insert: acc.Insert("viewchunks_referrers").Columns("count, createdAt, domain").Fields("?,UTC_TIMESTAMP(),?").Prepare(), // TODO: Do something more efficient than doing a query for each referrer
	}
	c.AddScheduledFifteenMinuteTask(refTracker.Tick)
	//c.AddScheduledSecondTask(refTracker.Tick)
	c.AddShutdownTask(refTracker.Tick)
	return refTracker, acc.FirstError()
}

// TODO: Move this and the other view tickers out of the main task loop to avoid blocking other tasks?
func (ref *DefaultReferrerTracker) Tick() (err error) {
	for referrer, counter := range referrersToDelete {
		// Handle views which squeezed through the gaps at the last moment
		count := counter.Count
		if count != 0 {
			err := ref.insertChunk(referrer, count) // TODO: Bulk insert for speed?
			if err != nil {
				return errors.Wrap(errors.WithStack(err),"ref counter")
			}
		}
		delete(referrersToDelete, referrer)
	}

	//  Run the queries and schedule zero view refs for deletion from memory
	refLoop := func(l *sync.RWMutex, m map[string]*ReferrerItem) error {
		l.Lock()
		defer l.Unlock()
		for referrer, counter := range m {
			if counter.Count == 0 {
				referrersToDelete[referrer] = counter
				delete(m, referrer)
			}
			count := atomic.SwapInt64(&counter.Count, 0)
			err := ref.insertChunk(referrer, count) // TODO: Bulk insert for speed?
			if err != nil {
				return errors.Wrap(errors.WithStack(err),"ref counter")
			}
		}
		return nil
	}
	err = refLoop(&ref.oddLock,ref.odd)
	if err != nil {
		return err
	}
	return refLoop(&ref.evenLock,ref.even)
}

func (ref *DefaultReferrerTracker) insertChunk(referrer string, count int64) error {
	if count == 0 {
		return nil
	}
	c.DebugDetailf("Inserting a vchunk with a count of %d for referrer %s", count, referrer)
	_, err := ref.insert.Exec(count, referrer)
	return err
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
		if refItem != nil {
			atomic.AddInt64(&refItem.Count, 1)
		} else {
			ref.evenLock.Lock()
			ref.even[referrer] = &ReferrerItem{Count: 1}
			ref.evenLock.Unlock()
		}
	} else {
		ref.oddLock.RLock()
		refItem = ref.odd[referrer]
		ref.oddLock.RUnlock()
		if refItem != nil {
			atomic.AddInt64(&refItem.Count, 1)
		} else {
			ref.oddLock.Lock()
			ref.odd[referrer] = &ReferrerItem{Count: 1}
			ref.oddLock.Unlock()
		}
	}
}
