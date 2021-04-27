package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

// MetaStore is a simple key-value store for the system to stash things in when needed
type MetaStore interface {
	Get(name string) (val string, err error)
	Set(name, val string) error
	SetInt(name string, val int) error
	SetInt64(name string, val int64) error
}

type DefaultMetaStore struct {
	get *sql.Stmt
	set *sql.Stmt
	add *sql.Stmt
}

func NewDefaultMetaStore(acc *qgen.Accumulator) (*DefaultMetaStore, error) {
	t := "meta"
	m := &DefaultMetaStore{
		get: acc.Select(t).Columns("value").Where("name=?").Prepare(),
		set: acc.Update(t).Set("value=?").Where("name=?").Prepare(),
		add: acc.Insert(t).Columns("name,value").Fields("?,''").Prepare(),
	}
	return m, acc.FirstError()
}

func (s *DefaultMetaStore) Get(name string) (val string, e error) {
	e = s.get.QueryRow(name).Scan(&val)
	return val, e
}

// TODO: Use timestamped rows as a more robust method of ensuring data integrity
func (s *DefaultMetaStore) setVal(name string, val interface{}) error {
	_, e := s.Get(name)
	if e == sql.ErrNoRows {
		_, e := s.add.Exec(name)
		if e != nil {
			return e
		}
	}
	_, e = s.set.Exec(val, name)
	return e
}

func (s *DefaultMetaStore) Set(name, val string) error {
	return s.setVal(name, val)
}

func (s *DefaultMetaStore) SetInt(name string, val int) error {
	return s.setVal(name, val)
}

func (s *DefaultMetaStore) SetInt64(name string, val int64) error {
	return s.setVal(name, val)
}
