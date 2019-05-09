package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Meta MetaStore

// MetaStore is a simple key-value store for the system to stash things in when needed
type MetaStore interface {
	Get(name string) (val string, err error)
	Set(name string, val string) error
}

type DefaultMetaStore struct {
	get *sql.Stmt
	set *sql.Stmt
	add *sql.Stmt
}

func NewDefaultMetaStore(acc *qgen.Accumulator) (*DefaultMetaStore, error) {
	m := &DefaultMetaStore{
		get: acc.Select("meta").Columns("value").Where("name = ?").Prepare(),
		set: acc.Update("meta").Set("value = ?").Where("name = ?").Prepare(),
		add: acc.Insert("meta").Columns("name,value").Fields("?,''").Prepare(),
	}
	return m, acc.FirstError()
}

func (s *DefaultMetaStore) Get(name string) (val string, err error) {
	err = s.get.QueryRow(name).Scan(&val)
	return val, err
}

// TODO: Use timestamped rows as a more robust method of ensuring data integrity
func (s *DefaultMetaStore) Set(name string, val string) error {
	_, err := s.Get(name)
	if err == sql.ErrNoRows {
		_, err := s.add.Exec(name)
		if err != nil {
			return err
		}
	}
	_, err = s.set.Exec(val, name)
	return err
}
