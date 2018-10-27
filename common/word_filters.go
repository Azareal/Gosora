package common

import (
	"database/sql"
	"sync/atomic"

	"github.com/Azareal/Gosora/query_gen"
)

// TODO: Move some features into methods on this?
type WordFilter struct {
	ID          int
	Find        string
	Replacement string
}

var WordFilters WordFilterStore

type WordFilterStore interface {
	ReloadAll() error
	GetAll() (filters map[int]*WordFilter, err error)
	Create(find string, replacement string) error
	Delete(id int) error
	Update(id int, find string, replacement string) error
	Length() int
	EstCount() int
	GlobalCount() (count int)
}

type DefaultWordFilterStore struct {
	box atomic.Value // An atomic value holding a WordFilterMap

	getAll *sql.Stmt
	create *sql.Stmt
	delete *sql.Stmt
	update *sql.Stmt
	count  *sql.Stmt
}

func NewDefaultWordFilterStore(acc *qgen.Accumulator) (*DefaultWordFilterStore, error) {
	store := &DefaultWordFilterStore{
		getAll: acc.Select("word_filters").Columns("wfid, find, replacement").Prepare(),
		create: acc.Insert("word_filters").Columns("find, replacement").Fields("?,?").Prepare(),
		delete: acc.Delete("word_filters").Where("wfid = ?").Prepare(),
		update: acc.Update("word_filters").Set("find = ?, replacement = ?").Where("wfid = ?").Prepare(),
		count:  acc.Count("word_filters").Prepare(),
	}
	// TODO: Should we initialise this elsewhere?
	if acc.FirstError() == nil {
		acc.RecordError(store.ReloadAll())
	}
	return store, acc.FirstError()
}

// ReloadAll drops all the items in the memory cache and replaces them with fresh copies from the database
func (store *DefaultWordFilterStore) ReloadAll() error {
	var wordFilters = make(map[int]*WordFilter)
	filters, err := store.bypassGetAll()
	if err != nil {
		return err
	}

	for _, filter := range filters {
		wordFilters[filter.ID] = filter
	}

	store.box.Store(wordFilters)
	return nil
}

// ? - Return pointers to word filters intead to save memory? -- A map is a pointer.
func (store *DefaultWordFilterStore) bypassGetAll() (filters []*WordFilter, err error) {
	rows, err := store.getAll.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		filter := &WordFilter{ID: 0}
		err := rows.Scan(&filter.ID, &filter.Find, &filter.Replacement)
		if err != nil {
			return filters, err
		}
		filters = append(filters, filter)
	}
	return filters, rows.Err()
}

// GetAll returns all of the word filters in a map. Do note mutate this map (or maps returned from any store not explicitly noted as copies) as multiple threads may be accessing it at once
func (store *DefaultWordFilterStore) GetAll() (filters map[int]*WordFilter, err error) {
	return store.box.Load().(map[int]*WordFilter), nil
}

// Create adds a new word filter to the database and refreshes the memory cache
func (store *DefaultWordFilterStore) Create(find string, replacement string) error {
	_, err := store.create.Exec(find, replacement)
	if err != nil {
		return err
	}
	return store.ReloadAll()
}

// Delete removes a word filter from the database and refreshes the memory cache
func (store *DefaultWordFilterStore) Delete(id int) error {
	_, err := store.delete.Exec(id)
	if err != nil {
		return err
	}
	return store.ReloadAll()
}

func (store *DefaultWordFilterStore) Update(id int, find string, replacement string) error {
	_, err := store.update.Exec(find, replacement, id)
	if err != nil {
		return err
	}
	return store.ReloadAll()
}

// Length gets the number of word filters currently in memory, for the DefaultWordFilterStore, this should be all of them
func (store *DefaultWordFilterStore) Length() int {
	return len(store.box.Load().(map[int]*WordFilter))
}

// EstCount provides the same result as Length(), intended for alternate implementations of WordFilterStore, so that Length is the number of items in cache, if only a subset is held there and EstCount is the total count
func (store *DefaultWordFilterStore) EstCount() int {
	return len(store.box.Load().(map[int]*WordFilter))
}

// GlobalCount gets the total number of word filters directly from the database
func (store *DefaultWordFilterStore) GlobalCount() (count int) {
	err := store.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
