package counters

import "database/sql"
import c "github.com/Azareal/Gosora/common"
import "github.com/Azareal/Gosora/query_gen"

var LangViewCounter *DefaultLangViewCounter

var langCodes = []string{
	"unknown",
	"none",
	"af",
	"ar",
	"az",
	"be",
	"bg",
	"bs",
	"ca",
	"cs",
	"cy",
	"da",
	"de",
	"dv",
	"el",
	"en",
	"eo",
	"es",
	"et",
	"eu",
	"fa",
	"fi",
	"fo",
	"fr",
	"gl",
	"gu",
	"he",
	"hi",
	"hr",
	"hu",
	"hy",
	"id",
	"is",
	"it",
	"ja",
	"ka",
	"kk",
	"kn",
	"ko",
	"kok",
	"ky",
	"lt",
	"lv",
	"mi",
	"mk",
	"mn",
	"mr",
	"ms",
	"mt",
	"nb",
	"nl",
	"nn",
	"ns",
	"pa",
	"pl",
	"ps",
	"pt",
	"qu",
	"ro",
	"ru",
	"sa",
	"se",
	"sk",
	"sl",
	"sq",
	"sr",
	"sv",
	"sw",
	"syr",
	"ta",
	"te",
	"th",
	"tl",
	"tn",
	"tr",
	"tt",
	"ts",
	"uk",
	"ur",
	"uz",
	"vi",
	"xh",
	"zh",
	"zu",
}

type DefaultLangViewCounter struct {
	buckets        []*RWMutexCounterBucket //[OSID]count
	codesToIndices map[string]int

	insert *sql.Stmt
}

func NewDefaultLangViewCounter(acc *qgen.Accumulator) (*DefaultLangViewCounter, error) {
	var langBuckets = make([]*RWMutexCounterBucket, len(langCodes))
	for bucketID, _ := range langBuckets {
		langBuckets[bucketID] = &RWMutexCounterBucket{counter: 0}
	}
	var codesToIndices = make(map[string]int)
	for index, code := range langCodes {
		codesToIndices[code] = index
	}

	counter := &DefaultLangViewCounter{
		buckets:        langBuckets,
		codesToIndices: codesToIndices,
		insert:         acc.Insert("viewchunks_langs").Columns("count, createdAt, lang").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}

	c.AddScheduledFifteenMinuteTask(counter.Tick)
	//c.AddScheduledSecondTask(counter.Tick)
	c.AddShutdownTask(counter.Tick)
	return counter, acc.FirstError()
}

func (counter *DefaultLangViewCounter) Tick() error {
	for id, bucket := range counter.buckets {
		var count int
		bucket.RLock()
		count = bucket.counter
		bucket.counter = 0 // TODO: Add a SetZero method to reduce the amount of duplicate code between the OS and agent counters?
		bucket.RUnlock()

		err := counter.insertChunk(count, id) // TODO: Bulk insert for speed?
		if err != nil {
			return err
		}
	}
	return nil
}

func (counter *DefaultLangViewCounter) insertChunk(count int, id int) error {
	if count == 0 {
		return nil
	}
	var langCode = langCodes[id]
	c.DebugLogf("Inserting a viewchunk with a count of %d for lang %s (%d)", count, langCode, id)
	_, err := counter.insert.Exec(count, langCode)
	return err
}

func (counter *DefaultLangViewCounter) Bump(langCode string) (validCode bool) {
	validCode = true
	id, ok := counter.codesToIndices[langCode]
	if !ok {
		// TODO: Tell the caller that the code's invalid
		id = 0 // Unknown
		validCode = false
	}

	// TODO: Test this check
	c.DebugDetail("counter.buckets[", id, "]: ", counter.buckets[id])
	if len(counter.buckets) <= id || id < 0 {
		return validCode
	}
	counter.buckets[id].Lock()
	counter.buckets[id].counter++
	counter.buckets[id].Unlock()

	return validCode
}
