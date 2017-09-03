package main

import "errors"

// Go away, linter. We need to differentiate constants from variables somehow ;)
// nolint
const CACHE_STATIC int = 0
const CACHE_DYNAMIC int = 1
const CACHE_SQL int = 2

// ErrCacheDesync is thrown whenever a piece of data, for instance, a user is out of sync with the database. Currently unused.
var ErrCacheDesync = errors.New("The cache is out of sync with the database.") // TO-DO: A cross-server synchronisation mechanism

// ErrStoreCapacityOverflow is thrown whenever a datastore reaches it's maximum hard capacity. I'm not sure *if* this one is used, at the moment. Probably.
var ErrStoreCapacityOverflow = errors.New("This datastore has reached it's maximum capacity.")

// nolint
type DataStore interface {
	Load(id int) error
	Get(id int) (interface{}, error)
	GetUnsafe(id int) (interface{}, error)
	CascadeGet(id int) (interface{}, error)
	BypassGet(id int) (interface{}, error)
	Set(item interface{}) error
	Add(item interface{}) error
	AddUnsafe(item interface{}) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	GetLength() int
	GetCapacity() int
}
