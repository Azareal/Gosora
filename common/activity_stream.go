package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Activity ActivityStream

type ActivityStream interface {
	Add(a Alert) (int, error)
	Get(id int) (Alert, error)
	Count() (count int)
}

type DefaultActivityStream struct {
	add *sql.Stmt
	get *sql.Stmt
	count *sql.Stmt
}

func NewDefaultActivityStream(acc *qgen.Accumulator) (*DefaultActivityStream, error) {
	as := "activity_stream"
	return &DefaultActivityStream{
		add: acc.Insert(as).Columns("actor, targetUser, event, elementType, elementID, createdAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		get: acc.Select(as).Columns("actor, targetUser, event, elementType, elementID, createdAt").Where("asid = ?").Prepare(),
		count: acc.Count(as).Prepare(),
	}, acc.FirstError()
}

func (s *DefaultActivityStream) Add(a Alert) (int, error) {
	res, err := s.add.Exec(a.ActorID, a.TargetUserID, a.Event, a.ElementType, a.ElementID)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (s *DefaultActivityStream) Get(id int) (Alert, error) {
	a := Alert{ASID: id}
	err := s.get.QueryRow(id).Scan(&a.ActorID, &a.TargetUserID, &a.Event, &a.ElementType, &a.ElementID, &a.CreatedAt)
	return a, err
}

// TODO: Write a test for this
// Count returns the total number of activity stream items
func (s *DefaultActivityStream) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}