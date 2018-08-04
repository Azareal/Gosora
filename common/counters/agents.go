package counters

import (
	"database/sql"

	".."
	"../../query_gen/lib"
)

var AgentViewCounter *DefaultAgentViewCounter

type DefaultAgentViewCounter struct {
	agentBuckets []*RWMutexCounterBucket //[AgentID]count
	insert       *sql.Stmt
}

func NewDefaultAgentViewCounter() (*DefaultAgentViewCounter, error) {
	acc := qgen.NewAcc()
	var agentBuckets = make([]*RWMutexCounterBucket, len(agentMapEnum))
	for bucketID, _ := range agentBuckets {
		agentBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	counter := &DefaultAgentViewCounter{
		agentBuckets: agentBuckets,
		insert:       acc.Insert("viewchunks_agents").Columns("count, createdAt, browser").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick)
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultAgentViewCounter) Tick() error {
	for agentID, agentBucket := range counter.agentBuckets {
		var count int
		agentBucket.RLock()
		count = agentBucket.counter
		agentBucket.counter = 0
		agentBucket.RUnlock()

		err := counter.insertChunk(count, agentID) // TODO: Bulk insert for speed?
		if err != nil {
			return err
		}
	}
	return nil
}

func (counter *DefaultAgentViewCounter) insertChunk(count int, agent int) error {
	if count == 0 {
		return nil
	}
	var agentName = reverseAgentMapEnum[agent]
	common.DebugLogf("Inserting a viewchunk with a count of %d for agent %s (%d)", count, agentName, agent)
	_, err := counter.insert.Exec(count, agentName)
	return err
}

func (counter *DefaultAgentViewCounter) Bump(agent int) {
	// TODO: Test this check
	common.DebugDetail("counter.agentBuckets[", agent, "]: ", counter.agentBuckets[agent])
	if len(counter.agentBuckets) <= agent || agent < 0 {
		return
	}
	counter.agentBuckets[agent].Lock()
	counter.agentBuckets[agent].counter++
	counter.agentBuckets[agent].Unlock()
}
