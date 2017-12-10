package common

import (
	"database/sql"
	"sync/atomic"

	"../query_gen/lib"
)

var GlobalViewCounter *BufferedViewCounter

type BufferedViewCounter struct {
	buckets       [2]int64
	currentBucket int64

	insert *sql.Stmt
}

func NewGlobalViewCounter() (*BufferedViewCounter, error) {
	acc := qgen.Builder.Accumulator()
	counter := &BufferedViewCounter{
		currentBucket: 0,
		insert:        acc.SimpleInsert("viewchunks", "count, createdAt", "?,UTC_TIMESTAMP()"),
	}
	//AddScheduledFifteenMinuteTask(counter.Tick)
	AddScheduledSecondTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *BufferedViewCounter) Tick() (err error) {
	var oldBucket = counter.currentBucket
	var nextBucket int64
	if counter.currentBucket == 1 {
		nextBucket = 0
	} else {
		nextBucket = 1
	}
	atomic.AddInt64(&counter.buckets[oldBucket], counter.buckets[nextBucket])
	atomic.StoreInt64(&counter.buckets[nextBucket], 0)
	atomic.StoreInt64(&counter.currentBucket, nextBucket)
	/*debugLog("counter.buckets[nextBucket]: ", counter.buckets[nextBucket])
	debugLog("counter.buckets[oldBucket]: ", counter.buckets[oldBucket])
	debugLog("counter.currentBucket:", counter.currentBucket)
	debugLog("oldBucket:", oldBucket)
	debugLog("nextBucket:", nextBucket)*/

	var previousViewChunk = counter.buckets[oldBucket]
	atomic.AddInt64(&counter.buckets[oldBucket], -previousViewChunk)
	return counter.insertChunk(previousViewChunk)
}

func (counter *BufferedViewCounter) Bump() {
	atomic.AddInt64(&counter.buckets[counter.currentBucket], 1)
}

func (counter *BufferedViewCounter) insertChunk(count int64) error {
	if count == 0 {
		return nil
	}
	debugLogf("Inserting a viewchunk with a count of %d", count)
	_, err := counter.insert.Exec(count)
	return err
}
