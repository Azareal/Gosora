package counters

import "database/sql"
import ".."
import "../../query_gen/lib"

var RouteViewCounter *DefaultRouteViewCounter

// TODO: Make this lockless?
type DefaultRouteViewCounter struct {
	routeBuckets []*RWMutexCounterBucket //[RouteID]count
	insert       *sql.Stmt
}

func NewDefaultRouteViewCounter() (*DefaultRouteViewCounter, error) {
	acc := qgen.Builder.Accumulator()
	var routeBuckets = make([]*RWMutexCounterBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	counter := &DefaultRouteViewCounter{
		routeBuckets: routeBuckets,
		insert:       acc.Insert("viewchunks").Columns("count, createdAt, route").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultRouteViewCounter) Tick() error {
	for routeID, routeBucket := range counter.routeBuckets {
		var count int
		routeBucket.RLock()
		count = routeBucket.counter
		routeBucket.counter = 0
		routeBucket.RUnlock()

		err := counter.insertChunk(count, routeID) // TODO: Bulk insert for speed?
		if err != nil {
			return err
		}
	}
	return nil
}

func (counter *DefaultRouteViewCounter) insertChunk(count int, route int) error {
	if count == 0 {
		return nil
	}
	var routeName = reverseRouteMapEnum[route]
	common.DebugLogf("Inserting a viewchunk with a count of %d for route %s (%d)", count, routeName, route)
	_, err := counter.insert.Exec(count, routeName)
	return err
}

func (counter *DefaultRouteViewCounter) Bump(route int) {
	// TODO: Test this check
	common.DebugDetail("counter.routeBuckets[", route, "]: ", counter.routeBuckets[route])
	if len(counter.routeBuckets) <= route || route < 0 {
		return
	}
	counter.routeBuckets[route].Lock()
	counter.routeBuckets[route].counter++
	counter.routeBuckets[route].Unlock()
}
