package counters

import (
	"database/sql"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var AgentViewCounter *DefaultAgentViewCounter

type DefaultAgentViewCounter struct {
	agentBuckets []*RWMutexCounterBucket //[AgentID]count
	insert       *sql.Stmt
}

func NewDefaultAgentViewCounter(acc *qgen.Accumulator) (*DefaultAgentViewCounter, error) {
	var agentBuckets = make([]*RWMutexCounterBucket, len(agentMapEnum))
	for bucketID, _ := range agentBuckets {
		agentBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	co := &DefaultAgentViewCounter{
		agentBuckets: agentBuckets,
		insert:       acc.Insert("viewchunks_agents").Columns("count, createdAt, browser").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultAgentViewCounter) Tick() error {
	for agentID, agentBucket := range co.agentBuckets {
		var count int
		agentBucket.RLock()
		count = agentBucket.counter
		agentBucket.counter = 0
		agentBucket.RUnlock()

		err := co.insertChunk(count, agentID) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "agent counter")
		}
	}
	return nil
}

func (co *DefaultAgentViewCounter) insertChunk(count int, agent int) error {
	if count == 0 {
		return nil
	}
	agentName := reverseAgentMapEnum[agent]
	c.DebugLogf("Inserting a vchunk with a count of %d for agent %s (%d)", count, agentName, agent)
	_, err := co.insert.Exec(count, agentName)
	return err
}

func (co *DefaultAgentViewCounter) Bump(agent int) {
	// TODO: Test this check
	c.DebugDetail("co.agentBuckets[", agent, "]: ", co.agentBuckets[agent])
	if len(co.agentBuckets) <= agent || agent < 0 {
		return
	}
	co.agentBuckets[agent].Lock()
	co.agentBuckets[agent].counter++
	co.agentBuckets[agent].Unlock()
}
