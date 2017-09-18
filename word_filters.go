package main

import "sync/atomic"

type WordFilter struct {
	ID          int
	Find        string
	Replacement string
}
type WordFilterBox map[int]WordFilter

var wordFilterBox atomic.Value // An atomic value holding a WordFilterBox

func init() {
	wordFilterBox.Store(WordFilterBox(make(map[int]WordFilter)))
}

func LoadWordFilters() error {
	rows, err := getWordFiltersStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var wordFilters = WordFilterBox(make(map[int]WordFilter))
	var wfid int
	var find string
	var replacement string

	for rows.Next() {
		err := rows.Scan(&wfid, &find, &replacement)
		if err != nil {
			return err
		}
		wordFilters[wfid] = WordFilter{ID: wfid, Find: find, Replacement: replacement}
	}
	wordFilterBox.Store(wordFilters)
	return rows.Err()
}

func addWordFilter(id int, find string, replacement string) {
	wordFilters := wordFilterBox.Load().(WordFilterBox)
	wordFilters[id] = WordFilter{ID: id, Find: find, Replacement: replacement}
	wordFilterBox.Store(wordFilters)
}
