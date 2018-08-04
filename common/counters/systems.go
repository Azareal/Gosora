package counters

import "database/sql"
import ".."
import "../../query_gen/lib"

var OSViewCounter *DefaultOSViewCounter

type DefaultOSViewCounter struct {
	buckets []*RWMutexCounterBucket //[OSID]count
	insert  *sql.Stmt
}

func NewDefaultOSViewCounter() (*DefaultOSViewCounter, error) {
	acc := qgen.NewAcc()
	var osBuckets = make([]*RWMutexCounterBucket, len(osMapEnum))
	for bucketID, _ := range osBuckets {
		osBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	counter := &DefaultOSViewCounter{
		buckets: osBuckets,
		insert:  acc.Insert("viewchunks_systems").Columns("count, createdAt, system").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick)
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultOSViewCounter) Tick() error {
	for id, bucket := range counter.buckets {
		var count int
		bucket.RLock()
		count = bucket.counter
		bucket.counter = 0 // TODO: Add a SetZero method to reduce the amount of duplicate code between the OS and agent counters?
		bucket.RUnlock()

		err := counter.insertChunk(count, id) // TODO: Bulk insert for speed?
		if err != nil {
			return err
		}
	}
	return nil
}

func (counter *DefaultOSViewCounter) insertChunk(count int, os int) error {
	if count == 0 {
		return nil
	}
	var osName = reverseOSMapEnum[os]
	common.DebugLogf("Inserting a viewchunk with a count of %d for OS %s (%d)", count, osName, os)
	_, err := counter.insert.Exec(count, osName)
	return err
}

func (counter *DefaultOSViewCounter) Bump(id int) {
	// TODO: Test this check
	common.DebugDetail("counter.buckets[", id, "]: ", counter.buckets[id])
	if len(counter.buckets) <= id || id < 0 {
		return
	}
	counter.buckets[id].Lock()
	counter.buckets[id].counter++
	counter.buckets[id].Unlock()
}
