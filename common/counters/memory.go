package counters

import (
	"time"
	"sync"
	"runtime"
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

var MemoryCounter *DefaultMemoryCounter

type DefaultMemoryCounter struct {
	insert *sql.Stmt
	totMem uint64
	totCount uint64
	sync.Mutex
}

func NewMemoryCounter(acc *qgen.Accumulator) (*DefaultMemoryCounter, error) {
	co := &DefaultMemoryCounter{
		insert:        acc.Insert("memchunks").Columns("count, createdAt").Fields("?,UTC_TIMESTAMP()").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				co.Lock()
				co.totCount++
				co.totMem += m.Sys
				co.Unlock()
			}
		}
	}()
	return co, acc.FirstError()
}

func (co *DefaultMemoryCounter) Tick() (err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	var avgMem uint64
	co.Lock()
	co.totCount++
	co.totMem += m.Sys
	avgMem = co.totMem / co.totCount
	co.totMem = 0
	co.totCount = 0
	co.Unlock()
	c.DebugLogf("Inserting a memchunk with a value of %d", avgMem)
	_, err = co.insert.Exec(avgMem)
	return err
}