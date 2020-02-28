package counters

import (
	"database/sql"
	"math"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var PerfCounter *DefaultPerfCounter

type PerfCounterBucket struct {
	low  *MutexCounter64Bucket
	high *MutexCounter64Bucket
	avg  *MutexCounter64Bucket
}

// TODO: Track perf on a per route basis
type DefaultPerfCounter struct {
	buckets []*PerfCounterBucket

	insert *sql.Stmt
}

func NewDefaultPerfCounter(acc *qgen.Accumulator) (*DefaultPerfCounter, error) {
	co := &DefaultPerfCounter{
		buckets: []*PerfCounterBucket{
			&PerfCounterBucket{
				low:  &MutexCounter64Bucket{counter: math.MaxInt64},
				high: &MutexCounter64Bucket{counter: 0},
				avg:  &MutexCounter64Bucket{counter: 0},
			},
		},
		insert: acc.Insert("perfchunks").Columns("low,high,avg,createdAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
	}

	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultPerfCounter) Tick() error {
	getCounter := func(b *MutexCounter64Bucket) (c int64) {
		b.Lock()
		c = b.counter
		b.counter = 0
		b.Unlock()
		return c
	}
	for _, b := range co.buckets {
		var low int64
		b.low.Lock()
		low = b.low.counter
		b.low.counter = math.MaxInt64
		b.low.Unlock()
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

func (co *DefaultPerfCounter) Push(dur time.Duration /*,_ bool*/) {
	id := 0
	b := co.buckets[id]
	//c.DebugDetail("buckets[", id, "]: ", b)
	micro := dur.Microseconds()
	if micro >= math.MaxInt32 {
		c.LogWarning(errors.New("dur should not be int32 max or higher"))
	}

	low := b.low
	low.Lock()
	if micro < low.counter {
		low.counter = micro
	}
	low.Unlock()

	high := b.high
	high.Lock()
	if micro > high.counter {
		high.counter = micro
	}
	high.Unlock()

	avg := b.avg
	avg.Lock()
	if micro != avg.counter {
		if avg.counter == 0 {
			avg.counter = micro
		} else {
			avg.counter = (micro + avg.counter) / 2
		}
	}
	avg.Unlock()
}
