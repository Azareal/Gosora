package counters

import (
	"database/sql"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var TopicViewCounter *DefaultTopicViewCounter

// TODO: Use two odd-even maps for now, and move to something more concurrent later, maybe a sharded map?
type DefaultTopicViewCounter struct {
	oddTopics  map[int]*RWMutexCounterBucket // map[tid]struct{counter,sync.RWMutex}
	evenTopics map[int]*RWMutexCounterBucket
	oddLock    sync.RWMutex
	evenLock   sync.RWMutex

	weekState byte

	update    *sql.Stmt
	resetOdd  *sql.Stmt
	resetEven *sql.Stmt
	resetBoth *sql.Stmt

	insertListBuf []TopicViewInsert
	saveTick      *SavedTick
}

func NewDefaultTopicViewCounter() (*DefaultTopicViewCounter, error) {
	acc := qgen.NewAcc()
	t := "topics"
	co := &DefaultTopicViewCounter{
		oddTopics:  make(map[int]*RWMutexCounterBucket),
		evenTopics: make(map[int]*RWMutexCounterBucket),

		//update:     acc.Update(t).Set("views=views+?").Where("tid=?").Prepare(),
		update:    acc.Update(t).Set("views=views+?,weekEvenViews=weekEvenViews+?,weekOddViews=weekOddViews+?").Where("tid=?").Prepare(),
		resetOdd:  acc.Update(t).Set("weekOddViews=0").Prepare(),
		resetEven: acc.Update(t).Set("weekEvenViews=0").Prepare(),
		resetBoth: acc.Update(t).Set("weekOddViews=0,weekEvenViews=0").Prepare(),

		//insertListBuf: make([]TopicViewInsert, 1024),
	}
	e := co.WeekResetInit()
	if e != nil {
		return co, e
	}

	tick := func(f func() error) {
		c.Tasks.FifteenMin.Add(f) // Who knows how many topics we have queued up, we probably don't want this running too frequently
		//c.Tasks.Sec.Add(f)
		c.Tasks.Shutdown.Add(f)
	}
	tick(co.Tick)
	tick(co.WeekResetTick)

	return co, acc.FirstError()
}

type TopicViewInsert struct {
	Count   int
	TopicID int
}

type SavedTick struct {
	I  int
	I2 int
}

func (co *DefaultTopicViewCounter) handleInsertListBuf(i, i2 int) error {
	ilb := co.insertListBuf
	var lastSuccess int
	for i3 := i2; i3 < i; i3++ {
		iitem := ilb[i3]
		if e := co.insertChunk(iitem.Count, iitem.TopicID); e != nil {
			co.saveTick = &SavedTick{I: i, I2: lastSuccess + 1}
			for i3 := i2; i3 < i && i3 <= lastSuccess; i3++ {
				ilb[i3].Count, ilb[i3].TopicID = 0, 0
			}
			return errors.Wrap(errors.WithStack(e), "topicview counter")
		}
		lastSuccess = i3
	}
	for i3 := i2; i3 < i; i3++ {
		ilb[i3].Count, ilb[i3].TopicID = 0, 0
	}
	return nil
}

func (co *DefaultTopicViewCounter) Tick() error {
	// TODO: Fold multiple 1 view topics into one query

	/*if co.saveTick != nil {
		e := co.handleInsertListBuf(co.saveTick.I, co.saveTick.I2)
		if e != nil {
			return e
		}
		co.saveTick = nil
	}*/

	cLoop := func(l *sync.RWMutex, m map[int]*RWMutexCounterBucket) error {
		//i := 0
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
			/*if len(co.insertListBuf) >= i {
				co.insertListBuf[i].Count = count
				co.insertListBuf[i].TopicID = topicID
				i++
			} else if i < 4096 {
				co.insertListBuf = append(co.insertListBuf, TopicViewInsert{count, topicID})
			} else */if e := co.insertChunk(count, topicID); e != nil {
				return errors.Wrap(errors.WithStack(e), "topicview counter")
			}
			l.RLock()
		}
		l.RUnlock()
		return nil //co.handleInsertListBuf(i, 0)
	}
	e := cLoop(&co.oddLock, co.oddTopics)
	if e != nil {
		return e
	}
	return cLoop(&co.evenLock, co.evenTopics)
}

func (co *DefaultTopicViewCounter) WeekResetInit() error {
	lastWeekResetStr, e := c.Meta.Get("lastWeekReset")
	if e != nil && e != sql.ErrNoRows {
		return e
	}

	spl := strings.Split(lastWeekResetStr, "-")
	if len(spl) <= 1 {
		return nil
	}
	weekState, e := strconv.Atoi(spl[1])
	if e != nil {
		return e
	}
	co.weekState = byte(weekState)

	unixLastWeekReset, e := strconv.ParseInt(spl[0], 10, 64)
	if e != nil {
		return e
	}
	resetTime := time.Unix(unixLastWeekReset, 0)
	if time.Since(resetTime).Hours() >= (24 * 7) {
		_, e = co.resetBoth.Exec()
	}
	return e
}

func (co *DefaultTopicViewCounter) WeekResetTick() (e error) {
	now := time.Now()
	_, week := now.ISOWeek()
	if week != int(co.weekState) {
		if week%2 == 0 { // is even?
			_, e = co.resetOdd.Exec()
		} else {
			_, e = co.resetEven.Exec()
		}
		co.weekState = byte(week)
	}
	// TODO: Retry?
	if e != nil {
		return e
	}
	return c.Meta.Set("lastWeekReset", strconv.FormatInt(now.Unix(), 10)+"-"+strconv.Itoa(week))
}

// TODO: Optimise this further. E.g. Using IN() on every one view topic. Rinse and repeat for two views, three views, four views and five views.
func (co *DefaultTopicViewCounter) insertChunk(count, topicID int) (err error) {
	if count == 0 {
		return nil
	}

	c.DebugLogf("Inserting %d views into topic %d", count, topicID)
	even, odd := 0, 0
	_, week := time.Now().ISOWeek()
	if week%2 == 0 { // is even?
		even += count
	} else {
		odd += count
	}

	if true {
		_, err = co.update.Exec(count, even, odd, topicID)
	} else {
		_, err = co.update.Exec(count, topicID)
	}
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	// TODO: Add a way to disable this for extra speed ;)
	tc := c.Topics.GetCache()
	if tc != nil {
		t, err := tc.Get(topicID)
		if err == sql.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		atomic.AddInt64(&t.ViewCount, int64(count))
	}

	return nil
}

func (co *DefaultTopicViewCounter) Bump(topicID int) {
	// Is the ID even?
	if topicID%2 == 0 {
		co.evenLock.RLock()
		t, ok := co.evenTopics[topicID]
		co.evenLock.RUnlock()
		if ok {
			t.Lock()
			t.counter++
			t.Unlock()
		} else {
			co.evenLock.Lock()
			co.evenTopics[topicID] = &RWMutexCounterBucket{counter: 1}
			co.evenLock.Unlock()
		}
		return
	}

	co.oddLock.RLock()
	t, ok := co.oddTopics[topicID]
	co.oddLock.RUnlock()
	if ok {
		t.Lock()
		t.counter++
		t.Unlock()
	} else {
		co.oddLock.Lock()
		co.oddTopics[topicID] = &RWMutexCounterBucket{counter: 1}
		co.oddLock.Unlock()
	}
}
