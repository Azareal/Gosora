package common

import (
	"database/sql"
	"sync/atomic"

	qgen "github.com/Azareal/Gosora/query_gen"
)

// TODO: Move some features into methods on this?
type WordFilter struct {
	ID      int
	Find    string
	Replace string
}
type WordFilterDiff struct {
	BeforeFind    string
	BeforeReplace string
	AfterFind     string
	AfterReplace  string
}

var WordFilters WordFilterStore

type WordFilterStore interface {
	ReloadAll() error
	GetAll() (filters map[int]*WordFilter, err error)
	Get(id int) (*WordFilter, error)
	Create(find string, replacement string) (int, error)
	Delete(id int) error
	Update(id int, find string, replacement string) error
	Length() int
	EstCount() int
	Count() (count int)
}

type DefaultWordFilterStore struct {
	box atomic.Value // An atomic value holding a WordFilterMap

	getAll *sql.Stmt
	get    *sql.Stmt
	create *sql.Stmt
	delete *sql.Stmt
	update *sql.Stmt
	count  *sql.Stmt
}

func NewDefaultWordFilterStore(acc *qgen.Accumulator) (*DefaultWordFilterStore, error) {
	wf := "word_filters"
	store := &DefaultWordFilterStore{
		getAll: acc.Select(wf).Columns("wfid,find,replacement").Prepare(),
		get:    acc.Select(wf).Columns("find,replacement").Where("wfid = ?").Prepare(),
		create: acc.Insert(wf).Columns("find,replacement").Fields("?,?").Prepare(),
		delete: acc.Delete(wf).Where("wfid = ?").Prepare(),
		update: acc.Update(wf).Set("find = ?, replacement = ?").Where("wfid = ?").Prepare(),
		count:  acc.Count(wf).Prepare(),
	}
	// TODO: Should we initialise this elsewhere?
	if acc.FirstError() == nil {
		acc.RecordError(store.ReloadAll())
	}
	return store, acc.FirstError()
}

// ReloadAll drops all the items in the memory cache and replaces them with fresh copies from the database
func (s *DefaultWordFilterStore) ReloadAll() error {
	wordFilters := make(map[int]*WordFilter)
	filters, err := s.bypassGetAll()
	if err != nil {
		return err
	}

	for _, filter := range filters {
		wordFilters[filter.ID] = filter
	}

	s.box.Store(wordFilters)
	return nil
}

// ? - Return pointers to word filters intead to save memory? -- A map is a pointer.
func (s *DefaultWordFilterStore) bypassGetAll() (filters []*WordFilter, err error) {
	rows, err := s.getAll.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		f := &WordFilter{ID: 0}
		err := rows.Scan(&f.ID, &f.Find, &f.Replace)
		if err != nil {
			return filters, err
		}
		filters = append(filters, f)
	}
	return filters, rows.Err()
}

// GetAll returns all of the word filters in a map. Do note mutate this map (or maps returned from any store not explicitly noted as copies) as multiple threads may be accessing it at once
func (s *DefaultWordFilterStore) GetAll() (filters map[int]*WordFilter, err error) {
	return s.box.Load().(map[int]*WordFilter), nil
}

func (s *DefaultWordFilterStore) Get(id int) (*WordFilter, error) {
	wf := &WordFilter{ID: id}
	err := s.get.QueryRow(id).Scan(&wf.Find, &wf.Replace)
	return wf, err
}

// Create adds a new word filter to the database and refreshes the memory cache
func (s *DefaultWordFilterStore) Create(find string, replace string) (int, error) {
	res, err := s.create.Exec(find, replace)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id64), s.ReloadAll()
}

// Delete removes a word filter from the database and refreshes the memory cache
func (s *DefaultWordFilterStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	if err != nil {
		return err
	}
	return s.ReloadAll()
}

func (s *DefaultWordFilterStore) Update(id int, find string, replace string) error {
	_, err := s.update.Exec(find, replace, id)
	if err != nil {
		return err
	}
	return s.ReloadAll()
}

// Length gets the number of word filters currently in memory, for the DefaultWordFilterStore, this should be all of them
func (s *DefaultWordFilterStore) Length() int {
	return len(s.box.Load().(map[int]*WordFilter))
}

// EstCount provides the same result as Length(), intended for alternate implementations of WordFilterStore, so that Length is the number of items in cache, if only a subset is held there and EstCount is the total count
func (s *DefaultWordFilterStore) EstCount() int {
	return len(s.box.Load().(map[int]*WordFilter))
}

// Count gets the total number of word filters directly from the database
func (s *DefaultWordFilterStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
