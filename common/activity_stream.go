package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Activity ActivityStream

type ActivityStream interface {
	Add(a Alert) (int, error)
	Get(id int) (Alert, error)
	Delete(id int) error
	DeleteByParams(event string, targetID int, targetType string) error
	DeleteByParamsExtra(event string, targetID int, targetType, extra string) error
	AidsByParams(event string, elementID int, elementType string) (aids []int, err error)
	AidsByParamsExtra(event string, elementID int, elementType, extra string) (aids []int, err error)
	Count() (count int)
}

type DefaultActivityStream struct {
	add                 *sql.Stmt
	get                 *sql.Stmt
	delete              *sql.Stmt
	deleteByParams      *sql.Stmt
	deleteByParamsExtra *sql.Stmt
	aidsByParams        *sql.Stmt
	aidsByParamsExtra   *sql.Stmt
	count               *sql.Stmt
}

func NewDefaultActivityStream(acc *qgen.Accumulator) (*DefaultActivityStream, error) {
	as := "activity_stream"
	return &DefaultActivityStream{
		add:                 acc.Insert(as).Columns("actor,targetUser,event,elementType,elementID,createdAt,extra").Fields("?,?,?,?,?,UTC_TIMESTAMP(),?").Prepare(),
		get:                 acc.Select(as).Columns("actor,targetUser,event,elementType,elementID,createdAt,extra").Where("asid=?").Prepare(),
		delete:              acc.Delete(as).Where("asid=?").Prepare(),
		deleteByParams:      acc.Delete(as).Where("event=? AND elementID=? AND elementType=?").Prepare(),
		deleteByParamsExtra: acc.Delete(as).Where("event=? AND elementID=? AND elementType=? AND extra=?").Prepare(),
		aidsByParams:        acc.Select(as).Columns("asid").Where("event=? AND elementID=? AND elementType=?").Prepare(),
		aidsByParamsExtra:   acc.Select(as).Columns("asid").Where("event=? AND elementID=? AND elementType=? AND extra=?").Prepare(),
		count:               acc.Count(as).Prepare(),
	}, acc.FirstError()
}

func (s *DefaultActivityStream) Add(a Alert) (int, error) {
	res, err := s.add.Exec(a.ActorID, a.TargetUserID, a.Event, a.ElementType, a.ElementID, a.Extra)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (s *DefaultActivityStream) Get(id int) (Alert, error) {
	a := Alert{ASID: id}
	err := s.get.QueryRow(id).Scan(&a.ActorID, &a.TargetUserID, &a.Event, &a.ElementType, &a.ElementID, &a.CreatedAt, &a.Extra)
	return a, err
}

func (s *DefaultActivityStream) Delete(id int) error {
	_, err := s.delete.Exec(id)
	return err
}

func (s *DefaultActivityStream) DeleteByParams(event string, elementID int, elementType string) error {
	_, err := s.deleteByParams.Exec(event, elementID, elementType)
	return err
}

func (s *DefaultActivityStream) DeleteByParamsExtra(event string, elementID int, elementType, extra string) error {
	_, err := s.deleteByParamsExtra.Exec(event, elementID, elementType, extra)
	return err
}

func (s *DefaultActivityStream) AidsByParams(event string, elementID int, elementType string) (aids []int, err error) {
	rows, err := s.aidsByParams.Query(event, elementID, elementType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var aid int
		if err := rows.Scan(&aid); err != nil {
			return nil, err
		}
		aids = append(aids, aid)
	}
	return aids, rows.Err()
}

func (s *DefaultActivityStream) AidsByParamsExtra(event string, elementID int, elementType, extra string) (aids []int, e error) {
	rows, e := s.aidsByParamsExtra.Query(event, elementID, elementType, extra)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var aid int
		if e := rows.Scan(&aid); e != nil {
			return nil, e
		}
		aids = append(aids, aid)
	}
	return aids, rows.Err()
}

// TODO: Write a test for this
// Count returns the total number of activity stream items
func (s *DefaultActivityStream) Count() (count int) {
	e := s.count.QueryRow().Scan(&count)
	if e != nil {
		LogError(e)
	}
	return count
}
