package common

import (
	"database/sql"
	"encoding/json"

	"../query_gen/lib"
)

var Polls PollStore

type Poll struct {
	ID   int
	Type int // 0: Single choice, 1: Multiple choice, 2: Multiple choice w/ points
	//AntiCheat bool // Apply various mitigations for cheating
	// GroupPower map[gid]points // The number of points a group can spend in this poll, defaults to 1

	Options      map[int]string
	Results      map[int]int  // map[optionIndex]points
	QuickOptions []PollOption // TODO: Fix up the template transpiler so we don't need to use this hack anymore
	VoteCount    int
}

func (poll *Poll) Copy() Poll {
	return *poll
}

type PollOption struct {
	ID    int
	Value string
}

type Pollable interface {
	SetPoll(pollID int) error
}

type PollStore interface {
	Get(id int) (*Poll, error)
	Exists(id int) bool
	Create(parent Pollable, pollType int, pollOptions map[int]string) (int, error)
	Reload(id int) error
	//GlobalCount() int

	SetCache(cache PollCache)
	GetCache() PollCache
}

type DefaultPollStore struct {
	cache PollCache

	get    *sql.Stmt
	exists *sql.Stmt
	create *sql.Stmt
	delete *sql.Stmt
	//pollCount      *sql.Stmt
}

func NewDefaultPollStore(cache PollCache) (*DefaultPollStore, error) {
	acc := qgen.Builder.Accumulator()
	if cache == nil {
		cache = NewNullPollCache()
	}
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &DefaultPollStore{
		cache:  cache,
		get:    acc.Select("polls").Columns("type, options, votes").Where("pollID = ?").Prepare(),
		exists: acc.Select("polls").Columns("pollID").Where("pollID = ?").Prepare(),
		create: acc.Insert("polls").Columns("type, options").Fields("?,?").Prepare(),
		//pollCount: acc.SimpleCount("polls", "", ""),
	}, acc.FirstError()
}

func (store *DefaultPollStore) Exists(id int) bool {
	err := store.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

func (store *DefaultPollStore) Get(id int) (*Poll, error) {
	poll, err := store.cache.Get(id)
	if err == nil {
		return poll, nil
	}

	poll = &Poll{ID: id}
	var optionTxt []byte
	err = store.get.QueryRow(id).Scan(&poll.Type, &optionTxt, &poll.VoteCount)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(optionTxt, &poll.Options)
	if err == nil {
		poll.QuickOptions = store.unpackOptionsMap(poll.Options)
		store.cache.Set(poll)
	}
	return poll, err
}

func (store *DefaultPollStore) Reload(id int) error {
	poll := &Poll{ID: id}
	var optionTxt []byte
	err := store.get.QueryRow(id).Scan(&poll.Type, &optionTxt, &poll.VoteCount)
	if err != nil {
		store.cache.Remove(id)
		return err
	}

	err = json.Unmarshal(optionTxt, &poll.Options)
	if err != nil {
		store.cache.Remove(id)
		return err
	}

	poll.QuickOptions = store.unpackOptionsMap(poll.Options)
	_ = store.cache.Set(poll)
	return nil
}

func (store *DefaultPollStore) unpackOptionsMap(rawOptions map[int]string) []PollOption {
	options := make([]PollOption, len(rawOptions))
	for id, option := range rawOptions {
		options[id] = PollOption{id, option}
	}
	return options
}

func (store *DefaultPollStore) Create(parent Pollable, pollType int, pollOptions map[int]string) (id int, err error) {
	pollOptionsTxt, err := json.Marshal(pollOptions)
	if err != nil {
		return 0, err
	}

	res, err := store.create.Exec(pollType, pollOptionsTxt) //pollOptionsTxt
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), parent.SetPoll(int(lastID)) // TODO: Delete the poll if SetPoll fails
}

func (store *DefaultPollStore) SetCache(cache PollCache) {
	store.cache = cache
}

// TODO: We're temporarily doing this so that you can do ucache != nil in getTopicUser. Refactor it.
func (store *DefaultPollStore) GetCache() PollCache {
	_, ok := store.cache.(*NullPollCache)
	if ok {
		return nil
	}
	return store.cache
}
