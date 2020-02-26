package counters

import (
	"database/sql"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var TopicCounter *DefaultTopicCounter

type DefaultTopicCounter struct {
	buckets       [2]int64
	currentBucket int64

	insert *sql.Stmt
}

func NewTopicCounter() (*DefaultTopicCounter, error) {
	acc := qgen.NewAcc()
	co := &DefaultTopicCounter{
		currentBucket: 0,
		insert:        acc.Insert("topicchunks").Columns("count,createdAt").Fields("?,UTC_TIMESTAMP()").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultTopicCounter) Tick() (err error) {
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
		return errors.Wrap(errors.WithStack(err), "topics counter")
	}
	return nil
}

func (co *DefaultTopicCounter) Bump() {
	atomic.AddInt64(&co.buckets[co.currentBucket], 1)
}

func (co *DefaultTopicCounter) insertChunk(count int64) error {
	if count == 0 {
		return nil
	}
	c.DebugLogf("Inserting a topicchunk with a count of %d", count)
	_, err := co.insert.Exec(count)
	return err
}
