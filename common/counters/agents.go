package counters

import (
	"database/sql"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var AgentViewCounter *DefaultAgentViewCounter

type DefaultAgentViewCounter struct {
	buckets []int64 //[AgentID]count
	insert  *sql.Stmt
}

func NewDefaultAgentViewCounter(acc *qgen.Accumulator) (*DefaultAgentViewCounter, error) {
	co := &DefaultAgentViewCounter{
		buckets: make([]int64, len(agentMapEnum)),
		insert:  acc.Insert("viewchunks_agents").Columns("count,createdAt,browser").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.Tasks.FifteenMin.Add(co.Tick)
	//c.Tasks.Sec.Add(co.Tick)
	c.Tasks.Shutdown.Add(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultAgentViewCounter) Tick() error {
	for id, _ := range co.buckets {
		count := atomic.SwapInt64(&co.buckets[id], 0)
		e := co.insertChunk(count, id) // TODO: Bulk insert for speed?
		if e != nil {
			return errors.Wrap(errors.WithStack(e), "agent counter")
		}
	}
	return nil
}

func (co *DefaultAgentViewCounter) insertChunk(count int64, agent int) error {
	if count == 0 {
		return nil
	}
	agentName := reverseAgentMapEnum[agent]
	c.DebugLogf("Inserting a vchunk with a count of %d for agent %s (%d)", count, agentName, agent)
	_, e := co.insert.Exec(count, agentName)
	return e
}

func (co *DefaultAgentViewCounter) Bump(agent int) {
	// TODO: Test this check
	c.DebugDetail("buckets ", agent, ": ", co.buckets[agent])
	if len(co.buckets) <= agent || agent < 0 {
		return
	}
	atomic.AddInt64(&co.buckets[agent], 1)
}
