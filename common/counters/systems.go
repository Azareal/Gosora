package counters

import (
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var OSViewCounter *DefaultOSViewCounter

type DefaultOSViewCounter struct {
	buckets []*RWMutexCounterBucket //[OSID]count
	insert  *sql.Stmt
}

func NewDefaultOSViewCounter(acc *qgen.Accumulator) (*DefaultOSViewCounter, error) {
	var osBuckets = make([]*RWMutexCounterBucket, len(osMapEnum))
	for bucketID, _ := range osBuckets {
		osBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	co := &DefaultOSViewCounter{
		buckets: osBuckets,
		insert:  acc.Insert("viewchunks_systems").Columns("count, createdAt, system").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultOSViewCounter) Tick() error {
	for id, bucket := range co.buckets {
		var count int
		bucket.RLock()
		count = bucket.counter
		bucket.counter = 0 // TODO: Add a SetZero method to reduce the amount of duplicate code between the OS and agent counters?
		bucket.RUnlock()

		err := co.insertChunk(count, id) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "system counter")
		}
	}
	return nil
}

func (co *DefaultOSViewCounter) insertChunk(count int, os int) error {
	if count == 0 {
		return nil
	}
	osName := reverseOSMapEnum[os]
	c.DebugLogf("Inserting a vchunk with a count of %d for OS %s (%d)", count, osName, os)
	_, err := co.insert.Exec(count, osName)
	return err
}

func (co *DefaultOSViewCounter) Bump(id int) {
	// TODO: Test this check
	c.DebugDetail("co.buckets[", id, "]: ", co.buckets[id])
	if len(co.buckets) <= id || id < 0 {
		return
	}
	co.buckets[id].Lock()
	co.buckets[id].counter++
	co.buckets[id].Unlock()
}
