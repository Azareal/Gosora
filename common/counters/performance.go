package counters

import (
	"database/sql"
	"sync/atomic"
	"time"
	"math"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var PerfCounter *DefaultPerfCounter

type PerfCounterBucket struct {
	low *MutexCounter64Bucket
	high *MutexCounter64Bucket
	avg *MutexCounter64Bucket
}

// TODO: Track perf on a per route basis
type DefaultPerfCounter struct {
	buckets []PerfCounterBucket

	insert *sql.Stmt
}

func NewDefaultPerfCounter(acc *qgen.Accumulator) (*DefaultPerfCounter, error) {
	co := &DefaultPerfCounter{
		buckets:       	[]PerfCounterBucket{
			PerfCounterBucket{
				low: &MutexCounter64Bucket{counter: 0},
				high: &MutexCounter64Bucket{counter: 0},
				avg: &MutexCounter64Bucket{counter: 0},
			},
		},
		insert:         acc.Insert("perfchunks").Columns("low,high,avg,createdAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
	}

	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultPerfCounter) Tick() error {
	getCounter := func(b *MutexCounter64Bucket) int64 {
		return atomic.SwapInt64(&b.counter, 0)
	}
	for _, b := range co.buckets {
		low := atomic.SwapInt64(&b.low.counter, math.MaxInt64)
		if low == math.MaxInt64 {
			low = 0
		}
		high := getCounter(b.high)
		avg := getCounter(b.avg)

		err := co.insertChunk(low, high, avg) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "perf counter")
		}
	}
	return nil
}

func (co *DefaultPerfCounter) insertChunk(low, high, avg int64) error {
	if low == 0 && high == 0 && avg == 0 {
		return nil
	}
	c.DebugLogf("Inserting a pchunk with low %d, high %d, avg %d", low, high, avg)
	_, err := co.insert.Exec(low, high, avg)
	return err
}

func (co *DefaultPerfCounter) Push(dur time.Duration) {
	id := 0
	b := co.buckets[id]
	//c.DebugDetail("co.buckets[", id, "]: ", b)
	micro := dur.Microseconds()

	low := b.low
	if micro < low.counter {
		low.Lock()
		if micro < low.counter {
			atomic.StoreInt64(&low.counter,micro)
		}
		low.Unlock()
	}

	high := b.high
	if micro > high.counter {
		high.Lock()
		if micro > high.counter {
			atomic.StoreInt64(&high.counter,micro)
		}
		high.Unlock()
	}

	avg := b.avg
	// TODO: Sync semantics are slightly loose but it should be close enough for our purposes here
	if micro != avg.counter {
		t := false
		avg.Lock()
		if avg.counter == 0 {
			t = atomic.CompareAndSwapInt64(&avg.counter, 0, micro)
		}
		if !t && micro != avg.counter {
			atomic.StoreInt64(&avg.counter,(micro+avg.counter) / 2)
		}
		avg.Unlock()
	}
}
