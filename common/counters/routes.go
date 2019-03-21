package counters

import "database/sql"
import "github.com/Azareal/Gosora/common"
import "github.com/Azareal/Gosora/query_gen"

var RouteViewCounter *DefaultRouteViewCounter

// TODO: Make this lockless?
type DefaultRouteViewCounter struct {
	buckets []*RWMutexCounterBucket //[RouteID]count
	insert  *sql.Stmt
}

func NewDefaultRouteViewCounter(acc *qgen.Accumulator) (*DefaultRouteViewCounter, error) {
	var routeBuckets = make([]*RWMutexCounterBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	counter := &DefaultRouteViewCounter{
		buckets: routeBuckets,
		insert:  acc.Insert("viewchunks").Columns("count, createdAt, route").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultRouteViewCounter) Tick() error {
	for routeID, routeBucket := range counter.buckets {
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
	common.DebugDetail("counter.buckets[", route, "]: ", counter.buckets[route])
	if len(counter.buckets) <= route || route < 0 {
		return
	}
	counter.buckets[route].Lock()
	counter.buckets[route].counter++
	counter.buckets[route].Unlock()
}
