package counters

import (
	"database/sql"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var OSViewCounter *DefaultOSViewCounter

type DefaultOSViewCounter struct {
	buckets []int64 //[OSID]count
	insert  *sql.Stmt
}

func NewDefaultOSViewCounter(acc *qgen.Accumulator) (*DefaultOSViewCounter, error) {
	co := &DefaultOSViewCounter{
		buckets: make([]int64, len(osMapEnum)),
		insert:  acc.Insert("viewchunks_systems").Columns("count,createdAt,system").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultOSViewCounter) Tick() error {
	for id, _ := range co.buckets {
		count := atomic.SwapInt64(&co.buckets[id], 0)
		err := co.insertChunk(count, id) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "system counter")
		}
	}
	return nil
}

func (co *DefaultOSViewCounter) insertChunk(count int64, os int) error {
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
	atomic.AddInt64(&co.buckets[id], 1)
}
