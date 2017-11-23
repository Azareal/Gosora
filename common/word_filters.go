package common

import (
	"database/sql"
	"sync/atomic"

	"../query_gen/lib"
)

type WordFilter struct {
	ID          int
	Find        string
	Replacement string
}
type WordFilterMap map[int]WordFilter

var WordFilterBox atomic.Value // An atomic value holding a WordFilterBox

type FilterStmts struct {
	getWordFilters *sql.Stmt
}

var filterStmts FilterStmts

func init() {
	WordFilterBox.Store(WordFilterMap(make(map[int]WordFilter)))
	DbInits.Add(func(acc *qgen.Accumulator) error {
		filterStmts = FilterStmts{
			getWordFilters: acc.Select("word_filters").Columns("wfid, find, replacement").Prepare(),
		}
		return acc.FirstError()
	})
}

func LoadWordFilters() error {
	var wordFilters = WordFilterMap(make(map[int]WordFilter))
	filters, err := wordFilters.BypassGetAll()
	if err != nil {
		return err
	}

	for _, filter := range filters {
		wordFilters[filter.ID] = filter
	}

	WordFilterBox.Store(wordFilters)
	return nil
}

// TODO: Return pointers to word filters intead to save memory?
func (wBox WordFilterMap) BypassGetAll() (filters []WordFilter, err error) {
	rows, err := filterStmts.getWordFilters.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		filter := WordFilter{ID: 0}
		err := rows.Scan(&filter.ID, &filter.Find, &filter.Replacement)
		if err != nil {
			return filters, err
		}
		filters = append(filters, filter)
	}
	return filters, rows.Err()
}

func AddWordFilter(id int, find string, replacement string) {
	wordFilters := WordFilterBox.Load().(WordFilterMap)
	wordFilters[id] = WordFilter{ID: id, Find: find, Replacement: replacement}
	WordFilterBox.Store(wordFilters)
}
