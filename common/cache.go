package common

import "errors"

// nolint
// ErrCacheDesync is thrown whenever a piece of data, for instance, a user is out of sync with the database. Currently unused.
var ErrCacheDesync = errors.New("The cache is out of sync with the database.") // TODO: A cross-server synchronisation mechanism

// ErrStoreCapacityOverflow is thrown whenever a datastore reaches it's maximum hard capacity. I'm not sure if this error is actually used. It might be, we should check
var ErrStoreCapacityOverflow = errors.New("This datastore has reached it's maximum capacity.") // nolint

// nolint
type DataStore interface {
	DirtyGet(id int) interface{}
	Get(id int) (interface{}, error)
	BypassGet(id int) (interface{}, error)
	//GlobalCount()
}

// nolint
type DataCache interface {
	CacheGet(id int) (interface{}, error)
	CacheGetUnsafe(id int) (interface{}, error)
	CacheSet(item interface{}) error
	CacheAdd(item interface{}) error
	CacheAddUnsafe(item interface{}) error
	CacheRemove(id int) error
	CacheRemoveUnsafe(id int) error
	Reload(id int) error
	Flush()
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}
