package counters

import (
	"database/sql"
	"sync"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
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
	co := &DefaultTopicViewCounter{
		oddTopics:  make(map[int]*RWMutexCounterBucket),
		evenTopics: make(map[int]*RWMutexCounterBucket),
		update:     acc.Update("topics").Set("views = views + ?").Where("tid = ?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // Who knows how many topics we have queued up, we probably don't want this running too frequently
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultTopicViewCounter) Tick() error {
	// TODO: Fold multiple 1 view topics into one query

	cLoop := func(l *sync.RWMutex, m map[int]*RWMutexCounterBucket) error {
		l.RLock()
		for topicID, topic := range m {
			l.RUnlock()
			var count int
			topic.RLock()
			count = topic.counter
			topic.RUnlock()
			// TODO: Only delete the bucket when it's zero to avoid hitting popular topics?
			l.Lock()
			delete(m, topicID)
			l.Unlock()
			err := co.insertChunk(count, topicID)
			if err != nil {
				return errors.Wrap(errors.WithStack(err),"topicview counter")
			}
			l.RLock()
		}
		l.RUnlock()
		return nil
	}
	err := cLoop(&co.oddLock,co.oddTopics)
	if err != nil {
		return err
	}
	return cLoop(&co.evenLock,co.evenTopics)
}

// TODO: Optimise this further. E.g. Using IN() on every one view topic. Rinse and repeat for two views, three views, four views and five views.
func (co *DefaultTopicViewCounter) insertChunk(count int, topicID int) error {
	if count == 0 {
		return nil
	}

	c.DebugLogf("Inserting %d views into topic %d", count, topicID)
	_, err := co.update.Exec(count, topicID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	// TODO: Add a way to disable this for extra speed ;)
	tcache := c.Topics.GetCache()
	if tcache != nil {
		topic, err := tcache.Get(topicID)
		if err == sql.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		atomic.AddInt64(&topic.ViewCount, int64(count))
	}

	return nil
}

func (co *DefaultTopicViewCounter) Bump(topicID int) {
	// Is the ID even?
	if topicID%2 == 0 {
		co.evenLock.RLock()
		topic, ok := co.evenTopics[topicID]
		co.evenLock.RUnlock()
		if ok {
			topic.Lock()
			topic.counter++
			topic.Unlock()
		} else {
			co.evenLock.Lock()
			co.evenTopics[topicID] = &RWMutexCounterBucket{counter: 1}
			co.evenLock.Unlock()
		}
		return
	}

	co.oddLock.RLock()
	topic, ok := co.oddTopics[topicID]
	co.oddLock.RUnlock()
	if ok {
		topic.Lock()
		topic.counter++
		topic.Unlock()
	} else {
		co.oddLock.Lock()
		co.oddTopics[topicID] = &RWMutexCounterBucket{counter: 1}
		co.oddLock.Unlock()
	}
}
