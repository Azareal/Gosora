package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
)

var Polls PollStore

type Poll struct {
	ID          int
	ParentID    int
	ParentTable string
	Type        int // 0: Single choice, 1: Multiple choice, 2: Multiple choice w/ points
	//AntiCheat bool // Apply various mitigations for cheating
	// GroupPower map[gid]points // The number of points a group can spend in this poll, defaults to 1

	Options      map[int]string
	Results      map[int]int  // map[optionIndex]points
	QuickOptions []PollOption // TODO: Fix up the template transpiler so we don't need to use this hack anymore
	VoteCount    int
}

func (poll *Poll) CastVote(optionIndex int, uid int, ipaddress string) error {
	return Polls.CastVote(optionIndex, poll.ID, uid, ipaddress) // TODO: Move the query into a pollStmts rather than having it in the store
}

func (poll *Poll) Copy() Poll {
	return *poll
}

type PollOption struct {
	ID    int
	Value string
}

type Pollable interface {
	GetID() int
	GetTable() string
	SetPoll(pollID int) error
}

type PollStore interface {
	Get(id int) (*Poll, error)
	Exists(id int) bool
	Create(parent Pollable, pollType int, pollOptions map[int]string) (int, error)
	CastVote(optionIndex int, pollID int, uid int, ipaddress string) error
	Reload(id int) error
	//GlobalCount() int

	SetCache(cache PollCache)
	GetCache() PollCache
}

type DefaultPollStore struct {
	cache PollCache

	get                         *sql.Stmt
	exists                      *sql.Stmt
	createPoll                  *sql.Stmt
	createPollOption            *sql.Stmt
	addVote                     *sql.Stmt
	incrementVoteCount          *sql.Stmt
	incrementVoteCountForOption *sql.Stmt
	delete                      *sql.Stmt
	//pollCount      *sql.Stmt
}

func NewDefaultPollStore(cache PollCache) (*DefaultPollStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullPollCache()
	}
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &DefaultPollStore{
		cache:                       cache,
		get:                         acc.Select("polls").Columns("parentID, parentTable, type, options, votes").Where("pollID = ?").Prepare(),
		exists:                      acc.Select("polls").Columns("pollID").Where("pollID = ?").Prepare(),
		createPoll:                  acc.Insert("polls").Columns("parentID, parentTable, type, options").Fields("?,?,?,?").Prepare(),
		createPollOption:            acc.Insert("polls_options").Columns("pollID, option, votes").Fields("?,?,0").Prepare(),
		addVote:                     acc.Insert("polls_votes").Columns("pollID, uid, option, castAt, ipaddress").Fields("?,?,?,UTC_TIMESTAMP(),?").Prepare(),
		incrementVoteCount:          acc.Update("polls").Set("votes = votes + 1").Where("pollID = ?").Prepare(),
		incrementVoteCountForOption: acc.Update("polls_options").Set("votes = votes + 1").Where("option = ? AND pollID = ?").Prepare(),
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
	err = store.get.QueryRow(id).Scan(&poll.ParentID, &poll.ParentTable, &poll.Type, &optionTxt, &poll.VoteCount)
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

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (store *DefaultPollStore) BulkGetMap(ids []int) (list map[int]*Poll, err error) {
	var idCount = len(ids)
	list = make(map[int]*Poll)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := store.cache.BulkGet(ids)
	for i, sliceItem := range sliceList {
		if sliceItem != nil {
			list[sliceItem.ID] = sliceItem
		} else {
			stillHere = append(stillHere, ids[i])
		}
	}
	ids = stillHere

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	}

	// TODO: Add a function for the qlist stuff
	var qlist string
	var pollIDList []interface{}
	for _, id := range ids {
		pollIDList = append(pollIDList, strconv.Itoa(id))
		qlist += "?,"
	}
	qlist = qlist[0 : len(qlist)-1]

	rows, err := qgen.NewAcc().Select("polls").Columns("pollID, parentID, parentTable, type, options, votes").Where("pollID IN(" + qlist + ")").Query(pollIDList...)
	if err != nil {
		return list, err
	}

	for rows.Next() {
		poll := &Poll{ID: 0}
		var optionTxt []byte
		err := rows.Scan(&poll.ID, &poll.ParentID, &poll.ParentTable, &poll.Type, &optionTxt, &poll.VoteCount)
		if err != nil {
			return list, err
		}

		err = json.Unmarshal(optionTxt, &poll.Options)
		if err != nil {
			return list, err
		}
		poll.QuickOptions = store.unpackOptionsMap(poll.Options)
		store.cache.Set(poll)

		list[poll.ID] = poll
	}

	// Did we miss any polls?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			_, ok := list[id]
			if !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}

		// We probably don't need this, but it might be useful in case of bugs in BulkCascadeGetMap
		if sidList == "" {
			// TODO: Bulk log this
			if Dev.DebugMode {
				log.Print("This data is sampled later in the BulkCascadeGetMap function, so it might miss the cached IDs")
				log.Print("idCount", idCount)
				log.Print("ids", ids)
				log.Print("list", list)
			}
			return list, errors.New("We weren't able to find a poll, but we don't know which one")
		}
		sidList = sidList[0 : len(sidList)-1]

		err = errors.New("Unable to find the polls with the following IDs: " + sidList)
	}

	return list, err
}

func (store *DefaultPollStore) Reload(id int) error {
	poll := &Poll{ID: id}
	var optionTxt []byte
	err := store.get.QueryRow(id).Scan(&poll.ParentID, &poll.ParentTable, &poll.Type, &optionTxt, &poll.VoteCount)
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

// TODO: Use a transaction for this?
func (store *DefaultPollStore) CastVote(optionIndex int, pollID int, uid int, ipaddress string) error {
	_, err := store.addVote.Exec(pollID, uid, optionIndex, ipaddress)
	if err != nil {
		return err
	}
	_, err = store.incrementVoteCount.Exec(pollID)
	if err != nil {
		return err
	}
	_, err = store.incrementVoteCountForOption.Exec(optionIndex, pollID)
	return err
}

// TODO: Use a transaction for this
func (store *DefaultPollStore) Create(parent Pollable, pollType int, pollOptions map[int]string) (id int, err error) {
	pollOptionsTxt, err := json.Marshal(pollOptions)
	if err != nil {
		return 0, err
	}

	res, err := store.createPoll.Exec(parent.GetID(), parent.GetTable(), pollType, pollOptionsTxt)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(pollOptions); i++ {
		_, err := store.createPollOption.Exec(lastID, i)
		if err != nil {
			return 0, err
		}
	}
	return int(lastID), parent.SetPoll(int(lastID)) // TODO: Delete the poll (and options) if SetPoll fails
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
