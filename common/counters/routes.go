package counters

import (
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/Azareal/Gosora/uutils"
	"github.com/pkg/errors"
)

var RouteViewCounter *DefaultRouteViewCounter

type RVBucket struct {
	counter int64
	avg     int

	sync.Mutex
}

// TODO: Make this lockless?
type DefaultRouteViewCounter struct {
	buckets []*RVBucket //[RouteID]count
	insert  *sql.Stmt
	insert5 *sql.Stmt
}

func NewDefaultRouteViewCounter(acc *qgen.Accumulator) (*DefaultRouteViewCounter, error) {
	routeBuckets := make([]*RVBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RVBucket{counter: 0, avg: 0}
	}

	fields := "?,?,UTC_TIMESTAMP(),?"
	co := &DefaultRouteViewCounter{
		buckets: routeBuckets,
		insert:  acc.Insert("viewchunks").Columns("count,avg,createdAt,route").Fields(fields).Prepare(),
		insert5: acc.BulkInsert("viewchunks").Columns("count,avg,createdAt,route").Fields(fields, fields, fields, fields, fields).Prepare(),
	}
	if !c.Config.DisableAnalytics {
		c.Tasks.FifteenMin.Add(co.Tick) // There could be a lot of routes, so we don't want to be running this every second
		//c.Tasks.Sec.Add(co.Tick)
		c.Tasks.Shutdown.Add(co.Tick)
	}
	return co, acc.FirstError()
}

type RVCount struct {
	RouteID int
	Count   int64
	Avg     int
}

func (co *DefaultRouteViewCounter) Tick() (err error) {
	var tb []RVCount
	for routeID, b := range co.buckets {
		var avg int
		count := atomic.SwapInt64(&b.counter, 0)
		b.Lock()
		avg = b.avg
		b.avg = 0
		b.Unlock()

		if count == 0 {
			continue
		}
		tb = append(tb, RVCount{routeID, count, avg})
	}

	// TODO: Expand on this?
	var i int
	if len(tb) >= 5 {
		for ; len(tb) > (i + 5); i += 5 {
			err := co.insert5Chunk(tb[i : i+5])
			if err != nil {
				c.DebugLogf("tb: %+v\n", tb)
				c.DebugLog("i: ", i)
				return errors.Wrap(errors.WithStack(err), "route counter x 5")
			}
		}
	}

	for ; len(tb) > i; i++ {
		it := tb[i]
		err = co.insertChunk(it.Count, it.Avg, it.RouteID)
		if err != nil {
			c.DebugLogf("tb: %+v\n", tb)
			c.DebugLog("i: ", i)
			return errors.Wrap(errors.WithStack(err), "route counter")
		}
	}

	return nil
}

func (co *DefaultRouteViewCounter) insertChunk(count int64, avg, route int) error {
	routeName := reverseRouteMapEnum[route]
	c.DebugLogf("Inserting vchunk with count %d, avg %d for route %s (%d)", count, avg, routeName, route)
	_, err := co.insert.Exec(count, avg, routeName)
	return err
}

func (co *DefaultRouteViewCounter) insert5Chunk(rvs []RVCount) error {
	args := make([]interface{}, len(rvs)*3)
	i := 0
	for _, rv := range rvs {
		routeName := reverseRouteMapEnum[rv.RouteID]
		if rv.Avg == 0 {
			c.DebugLogf("Queueing vchunk with count %d for routes %s (%d)", rv.Count, routeName, rv.RouteID)
		} else {
			c.DebugLogf("Queueing vchunk with count %d, avg %d for routes %s (%d)", rv.Count, rv.Avg, routeName, rv.RouteID)
		}
		args[i] = rv.Count
		args[i+1] = rv.Avg
		args[i+2] = routeName
		i += 3
	}
	c.DebugDetailf("args: %+v\n", args)
	_, err := co.insert5.Exec(args...)
	return err
}

func (co *DefaultRouteViewCounter) Bump(route int) {
	if c.Config.DisableAnalytics {
		return
	}
	// TODO: Test this check
	b := co.buckets[route]
	c.DebugDetail("bucket ", route, ": ", b)
	if len(co.buckets) <= route || route < 0 {
		return
	}
	atomic.AddInt64(&b.counter, 1)
}

// TODO: Eliminate the lock?
func (co *DefaultRouteViewCounter) Bump2(route int, t time.Time) {
	if c.Config.DisableAnalytics {
		return
	}
	// TODO: Test this check
	b := co.buckets[route]
	c.DebugDetail("bucket ", route, ": ", b)
	if len(co.buckets) <= route || route < 0 {
		return
	}
	micro := int(time.Since(t).Microseconds())
	//co.PerfCounter.Push(since, true)
	atomic.AddInt64(&b.counter, 1)
	b.Lock()
	if micro != b.avg {
		if b.avg == 0 {
			b.avg = micro
		} else {
			b.avg = (micro + b.avg) / 2
		}
	}
	b.Unlock()
}

// TODO: Eliminate the lock?
func (co *DefaultRouteViewCounter) Bump3(route int, nano int64) {
	if c.Config.DisableAnalytics {
		return
	}
	// TODO: Test this check
	b := co.buckets[route]
	c.DebugDetail("bucket ", route, ": ", b)
	if len(co.buckets) <= route || route < 0 {
		return
	}
	micro := int((uutils.Nanotime() - nano) / 1000)
	//co.PerfCounter.Push(since, true)
	atomic.AddInt64(&b.counter, 1)
	b.Lock()
	if micro != b.avg {
		if b.avg == 0 {
			b.avg = micro
		} else {
			b.avg = (micro + b.avg) / 2
		}
	}
	b.Unlock()
}
