package counters

import (
	"database/sql"
	"runtime"
	"sync"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var MemoryCounter *DefaultMemoryCounter

type DefaultMemoryCounter struct {
	insert *sql.Stmt

	totMem     uint64
	totCount   uint64
	stackMem   uint64
	stackCount uint64
	heapMem    uint64
	heapCount  uint64

	sync.Mutex
}

func NewMemoryCounter(acc *qgen.Accumulator) (*DefaultMemoryCounter, error) {
	co := &DefaultMemoryCounter{
		insert: acc.Insert("memchunks").Columns("count, stack, heap, createdAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
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
				co.stackCount++
				co.stackMem += m.StackInuse
				co.heapCount++
				co.heapMem += m.HeapAlloc
				co.Unlock()
			}
		}
	}()
	return co, acc.FirstError()
}

func (co *DefaultMemoryCounter) Tick() (err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	var avgMem, avgStack, avgHeap uint64
	co.Lock()

	co.totCount++
	co.totMem += m.Sys
	co.stackCount++
	co.stackMem += m.StackInuse
	co.heapCount++
	co.heapMem += m.HeapAlloc

	avgMem = co.totMem / co.totCount
	avgStack = co.stackMem / co.stackCount
	avgHeap = co.heapMem / co.heapCount

	co.totMem = 0
	co.totCount = 0
	co.stackMem = 0
	co.stackCount = 0
	co.heapMem = 0
	co.heapCount = 0

	co.Unlock()

	c.DebugLogf("Inserting a memchunk with a value of %d - %d - %d", avgMem, avgStack, avgHeap)
	_, err = co.insert.Exec(avgMem, avgStack, avgHeap)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "mem counter")
	}
	return nil
}
