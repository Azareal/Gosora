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
	insert5 *sql.Stmt
}

func NewDefaultRouteViewCounter(acc *qgen.Accumulator) (*DefaultRouteViewCounter, error) {
	routeBuckets := make([]*RWMutexCounterBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}

	fields := "?,UTC_TIMESTAMP(),?"
	co := &DefaultRouteViewCounter{
		buckets: routeBuckets,
		insert:  acc.Insert("viewchunks").Columns("count,createdAt,route").Fields(fields).Prepare(),
		insert5:  acc.BulkInsert("viewchunks").Columns("count,createdAt,route").Fields(fields,fields,fields,fields,fields).Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

type RVCount struct {
	RouteID int
	Count int
}

func (co *DefaultRouteViewCounter) Tick() error {
	var tb []RVCount
	for routeID, b := range co.buckets {
		var count int
		b.RLock()
		count = b.counter
		b.counter = 0
		b.RUnlock()

		if count == 0 {
			continue
		}
		tb = append(tb, RVCount{routeID,count})
	}

	// TODO: Expand on this?
	var i int
	if len(tb) >= 5 {
		for ; len(tb) > (i+5); i += 5 {
			err := co.insert5Chunk(tb[i:i+5])
			if err != nil {
				c.DebugLogf("tb: %+v\n", tb)
				c.DebugLog("i: ", i)
				return errors.Wrap(errors.WithStack(err), "route counter x 5")
			}
		}
	}

	for ; len(tb) > i; i++ {
		err := co.insertChunk(tb[i].Count, tb[i].RouteID)
		if err != nil {
			c.DebugLogf("tb: %+v\n", tb)
			c.DebugLog("i: ", i)
			return errors.Wrap(errors.WithStack(err), "route counter")
		}
	}

	return nil
}

func (co *DefaultRouteViewCounter) insertChunk(count, route int) error {
	routeName := reverseRouteMapEnum[route]
	c.DebugLogf("Inserting a vchunk with a count of %d for route %s (%d)", count, routeName, route)
	_, err := co.insert.Exec(count, routeName)
	return err
}

func (co *DefaultRouteViewCounter) insert5Chunk(rvs []RVCount) error {
	args := make([]interface{}, len(rvs) * 2)
	i := 0
	for _, rv := range rvs {
		routeName := reverseRouteMapEnum[rv.RouteID]
		c.DebugLogf("Queueing a vchunk with a count of %d for routes %s (%d)", rv.Count, routeName, rv.RouteID)
		args[i] = rv.Count
		args[i+1] = routeName
		i += 2
	}
	c.DebugLogf("args: %+v\n", args)
	_, err := co.insert5.Exec(args...)
	return err
}

func (co *DefaultRouteViewCounter) Bump(route int) {
	// TODO: Test this check
	b := co.buckets[route]
	c.DebugDetail("co.buckets[", route, "]: ", b)
	if len(co.buckets) <= route || route < 0 {
		return
	}
	// TODO: Avoid lock by using atomic increment?
	b.Lock()
	b.counter++
	b.Unlock()
}
