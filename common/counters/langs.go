package counters

import (
	"database/sql"
	"sync/atomic"

	c "github.com/Azareal/Gosora/common"
	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/pkg/errors"
)

var LangViewCounter *DefaultLangViewCounter

var langCodes = []string{
	"unknown",
	"",
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
	"kw",
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
	//buckets        []*MutexCounterBucket //[OSID]count
	buckets        []int64 //[OSID]count
	codesToIndices map[string]int

	insert *sql.Stmt
}

func NewDefaultLangViewCounter(acc *qgen.Accumulator) (*DefaultLangViewCounter, error) {
	codesToIndices := make(map[string]int, len(langCodes))
	for index, code := range langCodes {
		codesToIndices[code] = index
	}
	co := &DefaultLangViewCounter{
		buckets:        make([]int64, len(langCodes)),
		codesToIndices: codesToIndices,
		insert:         acc.Insert("viewchunks_langs").Columns("count,createdAt,lang").Fields("?,UTC_TIMESTAMP(),?").Prepare(),
	}

	c.AddScheduledFifteenMinuteTask(co.Tick)
	//c.AddScheduledSecondTask(co.Tick)
	c.AddShutdownTask(co.Tick)
	return co, acc.FirstError()
}

func (co *DefaultLangViewCounter) Tick() error {
	for id := 0; id < len(co.buckets); id++ {
		count := atomic.SwapInt64(&co.buckets[id], 0)
		err := co.insertChunk(count, id) // TODO: Bulk insert for speed?
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "langview counter")
		}
	}
	return nil
}

func (co *DefaultLangViewCounter) insertChunk(count int64, id int) error {
	if count == 0 {
		return nil
	}
	langCode := langCodes[id]
	if langCode == "" {
		langCode = "none"
	}
	c.DebugLogf("Inserting a vchunk with a count of %d for lang %s (%d)", count, langCode, id)
	_, err := co.insert.Exec(count, langCode)
	return err
}

func (co *DefaultLangViewCounter) Bump(langCode string) (validCode bool) {
	validCode = true
	id, ok := co.codesToIndices[langCode]
	if !ok {
		// TODO: Tell the caller that the code's invalid
		id = 0 // Unknown
		validCode = false
	}

	// TODO: Test this check
	c.DebugDetail("buckets ", id, ": ", co.buckets[id])
	if len(co.buckets) <= id || id < 0 {
		return validCode
	}
	atomic.AddInt64(&co.buckets[id], 1)

	return validCode
}

func (co *DefaultLangViewCounter) Bump2(id int) {
	// TODO: Test this check
	c.DebugDetail("bucket ", id, ": ", co.buckets[id])
	if len(co.buckets) <= id || id < 0 {
		return
	}
	atomic.AddInt64(&co.buckets[id], 1)
}
