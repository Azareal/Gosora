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
		insert:  acc.Insert("viewchunks_forums").Columns("count,createdAt,forum").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	c.AddScheduledFifteenMinuteTask(co.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultForumViewCounter) Tick() error {
	cLoop := func(l *sync.RWMutex, m map[int]*RWMutexCounterBucket) error {
		l.RLock()
		for fid, f := range m {
			l.RUnlock()
			var count int
			f.RLock()
			count = f.counter
			f.RUnlock()
			// TODO: Only delete the bucket when it's zero to avoid hitting popular forums?
			l.Lock()
			delete(m, fid)
			l.Unlock()
			e := co.insertChunk(count, fid)
			if e != nil {
				return errors.Wrap(errors.WithStack(e),"forum counter")
			}
			l.RLock()
		}
		l.RUnlock()
		return nil
	}
	e := cLoop(&co.oddLock,co.oddMap)
	if e != nil {
		return e
	}
	return cLoop(&co.evenLock,co.evenMap)
}

func (co *DefaultForumViewCounter) insertChunk(count, forum int) error {
	if count == 0 {
		return nil
	}
	c.DebugLogf("Inserting a vchunk with a count of %d for forum %d", count, forum)
	_, e := co.insert.Exec(count, forum)
	return e
}

func (co *DefaultForumViewCounter) Bump(fid int) {
	// Is the ID even?
	if fid%2 == 0 {
		co.evenLock.RLock()
		f, ok := co.evenMap[fid]
		co.evenLock.RUnlock()
		if ok {
			f.Lock()
			f.counter++
			f.Unlock()
		} else {
			co.evenLock.Lock()
			co.evenMap[fid] = &RWMutexCounterBucket{counter: 1}
			co.evenLock.Unlock()
		}
		return
	}

	co.oddLock.RLock()
	f, ok := co.oddMap[fid]
	co.oddLock.RUnlock()
	if ok {
		f.Lock()
		f.counter++
		f.Unlock()
	} else {
		co.oddLock.Lock()
		co.oddMap[fid] = &RWMutexCounterBucket{counter: 1}
		co.oddLock.Unlock()
	}
}

// TODO: Add a forum counter backed by two maps which grow as forums are created but never shrinks
