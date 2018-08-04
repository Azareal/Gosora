package counters

import (
	"database/sql"
	"sync"

	".."
	"../../query_gen/lib"
)

var TopicViewCounter *DefaultTopicViewCounter

// TODO: Use two odd-even maps for now, and move to something more concurrent later, maybe a sharded map?
type DefaultTopicViewCounter struct {
	oddTopics  map[int]*RWMutexCounterBucket // map[tid]struct{counter,sync.RWMutex}
	evenTopics map[int]*RWMutexCounterBucket
	oddLock    sync.RWMutex
	evenLock   sync.RWMutex

	update *sql.Stmt
}

func NewDefaultTopicViewCounter() (*DefaultTopicViewCounter, error) {
	acc := qgen.NewAcc()
	counter := &DefaultTopicViewCounter{
		oddTopics:  make(map[int]*RWMutexCounterBucket),
		evenTopics: make(map[int]*RWMutexCounterBucket),
		update:     acc.Update("topics").Set("views = views + ?").Where("tid = ?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick) // Who knows how many topics we have queued up, we probably don't want this running too frequently
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultTopicViewCounter) Tick() error {
	// TODO: Fold multiple 1 view topics into one query

	counter.oddLock.RLock()
	oddTopics := counter.oddTopics
	counter.oddLock.RUnlock()
	for topicID, topic := range oddTopics {
		var count int
		topic.RLock()
		count = topic.counter
		topic.RUnlock()
		// TODO: Only delete the bucket when it's zero to avoid hitting popular topics?
		counter.oddLock.Lock()
		delete(counter.oddTopics, topicID)
		counter.oddLock.Unlock()
		err := counter.insertChunk(count, topicID)
		if err != nil {
			return err
		}
	}

	counter.evenLock.RLock()
	evenTopics := counter.evenTopics
	counter.evenLock.RUnlock()
	for topicID, topic := range evenTopics {
		var count int
		topic.RLock()
		count = topic.counter
		topic.RUnlock()
		// TODO: Only delete the bucket when it's zero to avoid hitting popular topics?
		counter.evenLock.Lock()
		delete(counter.evenTopics, topicID)
		counter.evenLock.Unlock()
		err := counter.insertChunk(count, topicID)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Optimise this further. E.g. Using IN() on every one view topic. Rinse and repeat for two views, three views, four views and five views.
func (counter *DefaultTopicViewCounter) insertChunk(count int, topicID int) error {
	if count == 0 {
		return nil
	}
	common.DebugLogf("Inserting %d views into topic %d", count, topicID)
	_, err := counter.update.Exec(count, topicID)
	return err
}

func (counter *DefaultTopicViewCounter) Bump(topicID int) {
	// Is the ID even?
	if topicID%2 == 0 {
		counter.evenLock.RLock()
		topic, ok := counter.evenTopics[topicID]
		counter.evenLock.RUnlock()
		if ok {
			topic.Lock()
			topic.counter++
			topic.Unlock()
		} else {
			counter.evenLock.Lock()
			counter.evenTopics[topicID] = &RWMutexCounterBucket{counter: 1}
			counter.evenLock.Unlock()
		}
		return
	}

	counter.oddLock.RLock()
	topic, ok := counter.oddTopics[topicID]
	counter.oddLock.RUnlock()
	if ok {
		topic.Lock()
		topic.counter++
		topic.Unlock()
	} else {
		counter.oddLock.Lock()
		counter.oddTopics[topicID] = &RWMutexCounterBucket{counter: 1}
		counter.oddLock.Unlock()
	}
}
