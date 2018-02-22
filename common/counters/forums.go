package counters

import (
	"database/sql"
	"sync"

	".."
	"../../query_gen/lib"
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
	acc := qgen.Builder.Accumulator()
	counter := &DefaultForumViewCounter{
		oddMap:  make(map[int]*RWMutexCounterBucket),
		evenMap: make(map[int]*RWMutexCounterBucket),
		insert:  acc.Insert("viewchunks_forums").Columns("count, createdAt, forum").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	common.AddScheduledFifteenMinuteTask(counter.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//common.AddScheduledSecondTask(counter.Tick)
	common.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultForumViewCounter) Tick() error {
	counter.oddLock.RLock()
	oddMap := counter.oddMap
	counter.oddLock.RUnlock()
	for forumID, forum := range oddMap {
		var count int
		forum.RLock()
		count = forum.counter
		forum.RUnlock()
		// TODO: Only delete the bucket when it's zero to avoid hitting popular forums?
		counter.oddLock.Lock()
		delete(counter.oddMap, forumID)
		counter.oddLock.Unlock()
		err := counter.insertChunk(count, forumID)
		if err != nil {
			return err
		}
	}

	counter.evenLock.RLock()
	evenMap := counter.evenMap
	counter.evenLock.RUnlock()
	for forumID, forum := range evenMap {
		var count int
		forum.RLock()
		count = forum.counter
		forum.RUnlock()
		// TODO: Only delete the bucket when it's zero to avoid hitting popular forums?
		counter.evenLock.Lock()
		delete(counter.evenMap, forumID)
		counter.evenLock.Unlock()
		err := counter.insertChunk(count, forumID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (counter *DefaultForumViewCounter) insertChunk(count int, forum int) error {
	if count == 0 {
		return nil
	}
	common.DebugLogf("Inserting a viewchunk with a count of %d for forum %d", count, forum)
	_, err := counter.insert.Exec(count, forum)
	return err
}

func (counter *DefaultForumViewCounter) Bump(forumID int) {
	// Is the ID even?
	if forumID%2 == 0 {
		counter.evenLock.RLock()
		forum, ok := counter.evenMap[forumID]
		counter.evenLock.RUnlock()
		if ok {
			forum.Lock()
			forum.counter++
			forum.Unlock()
		} else {
			counter.evenLock.Lock()
			counter.evenMap[forumID] = &RWMutexCounterBucket{counter: 1}
			counter.evenLock.Unlock()
		}
		return
	}

	counter.oddLock.RLock()
	forum, ok := counter.oddMap[forumID]
	counter.oddLock.RUnlock()
	if ok {
		forum.Lock()
		forum.counter++
		forum.Unlock()
	} else {
		counter.oddLock.Lock()
		counter.oddMap[forumID] = &RWMutexCounterBucket{counter: 1}
		counter.oddLock.Unlock()
	}
}

// TODO: Add a forum counter backed by two maps which grow as forums are created but never shrinks
