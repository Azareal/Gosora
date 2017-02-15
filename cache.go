package main
import "errors"

const CACHE_STATIC int = 0
const CACHE_DYNAMIC int = 1
const CACHE_SQL int = 2

var ErrStoreCapacityOverflow = errors.New("This datastore has already reached it's max capacity")

var users UserStore
var topics TopicStore

type DataStore interface {
	Load(id int) error
	Get(id int) (interface{}, error)
	GetUnsafe(id int) (interface{}, error)
	CascadeGet(id int) (interface{}, error)
	Set(item interface{}) error
	Add(item interface{}) error
	AddUnsafe(item interface{}) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	GetLength() int
	GetCapacity() int
}
