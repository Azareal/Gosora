package counters

import (
	"database/sql"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

// TODO: Rename this?
var GlobalViewCounter *DefaultViewCounter

// TODO: Rename this and shard it?
type DefaultViewCounter struct {
	buckets       [2]int64
	currentBucket int64

	insert *sql.Stmt
}

func NewGlobalViewCounter(acc *qgen.Accumulator) (*DefaultViewCounter, error) {
	co := &DefaultViewCounter{
		currentBucket: 0,
		insert:        acc.Insert("viewchunks").Columns("count,createdAt,route").Fields("?,UTC_TIMESTAMP(),''").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // This is run once every fifteen minutes to match the frequency of the RouteViewCounter
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

// TODO: Simplify the atomics used here
func (co *DefaultViewCounter) Tick() (err error) {
	oldBucket := co.currentBucket
	var nextBucket int64 // 0
	if co.currentBucket == 0 {
		nextBucket = 1
	}
	atomic.AddInt64(&co.buckets[oldBucket], co.buckets[nextBucket])
	atomic.StoreInt64(&co.buckets[nextBucket], 0)
	atomic.StoreInt64(&co.currentBucket, nextBucket)

	previousViewChunk := co.buckets[oldBucket]
	atomic.AddInt64(&co.buckets[oldBucket], -previousViewChunk)
	err = co.insertChunk(previousViewChunk)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "req counter")
	}
	return nil
}

func (co *DefaultViewCounter) Bump() {
	atomic.AddInt64(&co.buckets[co.currentBucket], 1)
}

func (co *DefaultViewCounter) insertChunk(count int64) error {
	if count == 0 {
		return nil
	}
	c.DebugLogf("Inserting a vchunk with a count of %d", count)
	_, err := co.insert.Exec(count)
	return err
}
