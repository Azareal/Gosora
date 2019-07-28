package counters

import (
	"database/sql"
	"sync"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var ForumViewCounter *DefaultForumViewCounter

// TODO: Unload forum counters without any views over the past 15 minutes, if the admin has configured the forumstore with a cap and it's been hit?
// Forums can be reloaded from the database at any time, so we want to keep the counters separate from them
type DefaultForumViewCounter struct {
	oddMap   map[int]*RWMutexCounterBucket // map[fid]struct{counter,sync.RWMutex}
	evenMap  map[int]*RWMutexCounterBucket
	oddLock  sync.RWMutex
	evenLock sync.RWMutex

	insert *sql.Stmt
}

func NewDefaultForumViewCounter() (*DefaultForumViewCounter, error) {
	acc := qgen.NewAcc()
	co := &DefaultForumViewCounter{
		oddMap:  make(map[int]*RWMutexCounterBucket),
		evenMap: make(map[int]*RWMutexCounterBucket),
		insert:  acc.Insert("viewchunks_forums").Columns("count, createdAt, forum").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultForumViewCounter) Tick() error {
	cLoop := func(l *sync.RWMutex, m map[int]*RWMutexCounterBucket) error {
		l.RLock()
		for forumID, forum := range m {
			l.RUnlock()
			var count int
			forum.RLock()
			count = forum.counter
			forum.RUnlock()
			// TODO: Only delete the bucket when it's zero to avoid hitting popular forums?
			l.Lock()
			delete(m, forumID)
			l.Unlock()
			err := co.insertChunk(count, forumID)
			if err != nil {
				return errors.Wrap(errors.WithStack(err),"forum counter")
			}
			l.RLock()
		}
		l.RUnlock()
		return nil
	}
	err := cLoop(&co.oddLock,co.oddMap)
	if err != nil {
		return err
	}
	return cLoop(&co.evenLock,co.evenMap)
}

func (co *DefaultForumViewCounter) insertChunk(count int, forum int) error {
	if count == 0 {
		return nil
	}
	c.DebugLogf("Inserting a vchunk with a count of %d for forum %d", count, forum)
	_, err := co.insert.Exec(count, forum)
	return err
}

func (co *DefaultForumViewCounter) Bump(forumID int) {
	// Is the ID even?
	if forumID%2 == 0 {
		co.evenLock.RLock()
		forum, ok := co.evenMap[forumID]
		co.evenLock.RUnlock()
		if ok {
			forum.Lock()
			forum.counter++
			forum.Unlock()
		} else {
			co.evenLock.Lock()
			co.evenMap[forumID] = &RWMutexCounterBucket{counter: 1}
			co.evenLock.Unlock()
		}
		return
	}

	co.oddLock.RLock()
	forum, ok := co.oddMap[forumID]
	co.oddLock.RUnlock()
	if ok {
		forum.Lock()
		forum.counter++
		forum.Unlock()
	} else {
		co.oddLock.Lock()
		co.oddMap[forumID] = &RWMutexCounterBucket{counter: 1}
		co.oddLock.Unlock()
	}
}

// TODO: Add a forum counter backed by two maps which grow as forums are created but never shrinks
