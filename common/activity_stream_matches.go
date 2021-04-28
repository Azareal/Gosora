package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var ActivityMatches ActivityStreamMatches

type ActivityStreamMatches interface {
	Add(watcher, asid int) error
	Delete(watcher, asid int) error
	DeleteAndCountChanged(watcher, asid int) (int, error)
	CountAsid(asid int) int
}

type DefaultActivityStreamMatches struct {
	add       *sql.Stmt
	delete    *sql.Stmt
	countAsid *sql.Stmt
}

func NewDefaultActivityStreamMatches(acc *qgen.Accumulator) (*DefaultActivityStreamMatches, error) {
	asm := "activity_stream_matches"
	return &DefaultActivityStreamMatches{
		add:       acc.Insert(asm).Columns("watcher,asid").Fields("?,?").Prepare(),
		delete:    acc.Delete(asm).Where("watcher=? AND asid=?").Prepare(),
		countAsid: acc.Count(asm).Where("asid=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultActivityStreamMatches) Add(watcher, asid int) error {
	_, e := s.add.Exec(watcher, asid)
	return e
}

func (s *DefaultActivityStreamMatches) Delete(watcher, asid int) error {
	_, e := s.delete.Exec(watcher, asid)
	return e
}

func (s *DefaultActivityStreamMatches) DeleteAndCountChanged(watcher, asid int) (int, error) {
	res, e := s.delete.Exec(watcher, asid)
	if e != nil {
		return 0, e
	}
	c64, e := res.RowsAffected()
	return int(c64), e
}

func (s *DefaultActivityStreamMatches) CountAsid(asid int) int {
	return Countf(s.countAsid, asid)
}
