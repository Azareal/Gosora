package common

import (
	"database/sql"
	"log"
	"sync"
	"sync/atomic"

	"../query_gen/lib"
)

var GlobalViewCounter *ChunkedViewCounter
var RouteViewCounter *RouteViewCounterImpl

type ChunkedViewCounter struct {
	buckets       [2]int64
	currentBucket int64

	insert *sql.Stmt
}

func NewChunkedViewCounter() (*ChunkedViewCounter, error) {
	acc := qgen.Builder.Accumulator()
	counter := &ChunkedViewCounter{
		currentBucket: 0,
		insert:        acc.Insert("viewchunks").Columns("count, createdAt").Fields("?,UTC_TIMESTAMP()").Prepare(),
	}
	AddScheduledFifteenMinuteTask(counter.Tick) // This is run once every fifteen minutes to match the frequency of the RouteViewCounter
	//AddScheduledSecondTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *ChunkedViewCounter) Tick() (err error) {
	var oldBucket = counter.currentBucket
	var nextBucket int64 // 0
	if counter.currentBucket == 0 {
		nextBucket = 1
	}
	atomic.AddInt64(&counter.buckets[oldBucket], counter.buckets[nextBucket])
	atomic.StoreInt64(&counter.buckets[nextBucket], 0)
	atomic.StoreInt64(&counter.currentBucket, nextBucket)

	var previousViewChunk = counter.buckets[oldBucket]
	atomic.AddInt64(&counter.buckets[oldBucket], -previousViewChunk)
	return counter.insertChunk(previousViewChunk)
}

func (counter *ChunkedViewCounter) Bump() {
	atomic.AddInt64(&counter.buckets[counter.currentBucket], 1)
}

func (counter *ChunkedViewCounter) insertChunk(count int64) error {
	if count == 0 {
		return nil
	}
	debugLogf("Inserting a viewchunk with a count of %d", count)
	_, err := counter.insert.Exec(count)
	return err
}

type RWMutexCounterBucket struct {
	counter int
	sync.RWMutex
}

// The name of the struct clashes with the name of the variable, so we're adding Impl to the end
type RouteViewCounterImpl struct {
	routeBuckets []*RWMutexCounterBucket //[RouteID]count
	insert       *sql.Stmt
}

func NewRouteViewCounter() (*RouteViewCounterImpl, error) {
	acc := qgen.Builder.Accumulator()
	var routeBuckets = make([]*RWMutexCounterBucket, len(routeMapEnum))
	for bucketID, _ := range routeBuckets {
		routeBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	counter := &RouteViewCounterImpl{
		routeBuckets: routeBuckets,
		insert:       acc.Insert("viewchunks").Columns("count, createdAt, route").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}
	AddScheduledFifteenMinuteTask(counter.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//AddScheduledSecondTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *RouteViewCounterImpl) Tick() error {
	for routeID, routeBucket := range counter.routeBuckets {
		var count int
		routeBucket.RLock()
		count = routeBucket.counter
		routeBucket.counter = 0
		routeBucket.RUnlock()

		err := counter.insertChunk(count, routeID) // TODO: Bulk insert for speed?
		if err != nil {
			return err
		}
	}
	return nil
}

func (counter *RouteViewCounterImpl) insertChunk(count int, route int) error {
	if count == 0 {
		return nil
	}
	var routeName = reverseRouteMapEnum[route]
	debugLogf("Inserting a viewchunk with a count of %d for route %s (%d)", count, routeName, route)
	_, err := counter.insert.Exec(count, routeName)
	return err
}

func (counter *RouteViewCounterImpl) Bump(route int) {
	// TODO: Test this check
	log.Print("counter.routeBuckets[route]: ", counter.routeBuckets[route])
	if len(counter.routeBuckets) <= route {
		return
	}
	counter.routeBuckets[route].Lock()
	counter.routeBuckets[route].counter++
	counter.routeBuckets[route].Unlock()
}

// TODO: The ForumViewCounter and TopicViewCounter

// TODO: Unload forum counters without any views over the past 15 minutes, if the admin has configured the forumstore with a cap and it's been hit?
// Forums can be reloaded from the database at any time, so we want to keep the counters separate from them
type ForumViewCounter struct {
	buckets       [2]int64
	currentBucket int64
}

/*func (counter *ForumViewCounter) insertChunk(count int, forum int) error {
	if count == 0 {
		return nil
	}
	debugLogf("Inserting a viewchunk with a count of %d for forum %d", count, forum)
	_, err := counter.insert.Exec(count, forum)
	return err
}*/

// TODO: Use two odd-even maps for now, and move to something more concurrent later, maybe a sharded map?
type TopicViewCounter struct {
	oddTopics  map[int]*RWMutexCounterBucket // map[tid]struct{counter,sync.RWMutex}
	evenTopics map[int]*RWMutexCounterBucket
	oddLock    sync.RWMutex
	evenLock   sync.RWMutex

	update *sql.Stmt
}

func NewTopicViewCounter() (*TopicViewCounter, error) {
	acc := qgen.Builder.Accumulator()
	counter := &TopicViewCounter{
		oddTopics:  make(map[int]*RWMutexCounterBucket),
		evenTopics: make(map[int]*RWMutexCounterBucket),
		update:     acc.Update("topics").Set("views = ?").Where("tid = ?").Prepare(), // TODO: Add the views column to the topics table
	}
	AddScheduledFifteenMinuteTask(counter.Tick) // There could be a lot of routes, so we don't want to be running this every second
	//AddScheduledSecondTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *TopicViewCounter) Tick() error {
	counter.oddLock.RLock()
	for topicID, topic := range counter.oddTopics {
		var count int
		topic.RLock()
		count = topic.counter
		topic.RUnlock()
		err := counter.insertChunk(count, topicID)
		if err != nil {
			return err
		}
	}
	counter.oddLock.RUnlock()

	counter.evenLock.RLock()
	for topicID, topic := range counter.evenTopics {
		var count int
		topic.RLock()
		count = topic.counter
		topic.RUnlock()
		err := counter.insertChunk(count, topicID)
		if err != nil {
			return err
		}
	}
	counter.evenLock.RUnlock()

	return nil
}

func (counter *TopicViewCounter) insertChunk(count int, topicID int) error {
	if count == 0 {
		return nil
	}
	debugLogf("Inserting %d views into topic %d", count, topicID)
	_, err := counter.update.Exec(count, topicID)
	return err
}

func (counter *TopicViewCounter) Bump(topicID int) {
	// Is the ID even?
	if topicID%2 == 0 {
		counter.evenLock.Lock()
		topic, ok := counter.evenTopics[topicID]
		counter.evenLock.Unlock()
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

	counter.oddLock.Lock()
	topic, ok := counter.oddTopics[topicID]
	counter.oddLock.Unlock()
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

type TreeCounterNode struct {
	Value  int64
	Zero   *TreeCounterNode
	One    *TreeCounterNode
	Parent *TreeCounterNode
}

// MEGA EXPERIMENTAL. Start from the right-most bits in the integer and move leftwards
type TreeTopicViewCounter struct {
	zero *TreeCounterNode
	one  *TreeCounterNode
}

func (counter *TreeTopicViewCounter) Bump(topicID int64) {

}
