package counters

import (
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var RouteViewCounter *DefaultRouteViewCounter

// TODO: Make this lockless?
type DefaultRouteViewCounter struct {
	buckets []*RWMutexCounterBucket //[RouteID]count
	insert  *sql.Stmt
}

func NewDefaultRouteViewCounter(acc *qgen.Accumulator) (*DefaultRouteViewCounter, error) {
	routeBuckets := make([]*RWMutexCounterBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	co := &DefaultRouteViewCounter{
		buckets: routeBuckets,
		insert:  acc.Insert("viewchunks").Columns("count, createdAt, route").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultRouteViewCounter) Tick() error {
	for routeID, routeBucket := range co.buckets {
		var count int
		routeBucket.RLock()
		count = routeBucket.counter
		routeBucket.counter = 0
		routeBucket.RUnlock()

		err := co.insertChunk(count, routeID) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "route counter")
		}
	}
	return nil
}

func (co *DefaultRouteViewCounter) insertChunk(count int, route int) error {
	if count == 0 {
		return nil
	}
	routeName := reverseRouteMapEnum[route]
	c.DebugLogf("Inserting a vchunk with a count of %d for route %s (%d)", count, routeName, route)
	_, err := co.insert.Exec(count, routeName)
	return err
}

func (co *DefaultRouteViewCounter) Bump(route int) {
	// TODO: Test this check
	c.DebugDetail("co.buckets[", route, "]: ", co.buckets[route])
	if len(co.buckets) <= route || route < 0 {
		return
	}
	co.buckets[route].Lock()
	co.buckets[route].counter++
	co.buckets[route].Unlock()
}
