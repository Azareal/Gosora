package counters

import (
	"runtime"
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

var MemoryCounter *DefaultMemoryCounter

type DefaultMemoryCounter struct {
	insert *sql.Stmt
}

func NewMemoryCounter(acc *qgen.Accumulator) (*DefaultMemoryCounter, error) {
	co := &DefaultMemoryCounter{
		insert:        acc.Insert("memchunks").Columns("count, createdAt").Fields("?,UTC_TIMESTAMP()").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultMemoryCounter) Tick() (err error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.DebugLogf("Inserting a memchunk with a value of %d", m.Sys)
	_, err = co.insert.Exec(m.Sys)
	return err
}