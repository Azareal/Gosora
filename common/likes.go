package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Likes LikeStore

type LikeStore interface {
	BulkExists(ids []int, sentBy int, targetType string) ([]int, error)
	Delete(targetID int, targetType string) error
	Count() (count int)
}

type DefaultLikeStore struct {
	count        *sql.Stmt
	delete       *sql.Stmt
	singleExists *sql.Stmt
}

func NewDefaultLikeStore(acc *qgen.Accumulator) (*DefaultLikeStore, error) {
	return &DefaultLikeStore{
		count:        acc.Count("likes").Prepare(),
		delete:       acc.Delete("likes").Where("targetItem=? AND targetType=?").Prepare(),
		singleExists: acc.Select("likes").Columns("targetItem").Where("sentBy=? AND targetType=? AND targetItem=?").Prepare(),
	}, acc.FirstError()
}

// TODO: Write a test for this
func (s *DefaultLikeStore) BulkExists(ids []int, sentBy int, targetType string) (eids []int, e error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows *sql.Rows
	if len(ids) == 1 {
		rows, e = s.singleExists.Query(sentBy, targetType, ids[0])
	} else {
		rows, e = qgen.NewAcc().Select("likes").Columns("targetItem").Where("sentBy=? AND targetType=?").In("targetItem", ids).Query(sentBy, targetType)
	}
	if e == sql.ErrNoRows {
		return nil, nil
	} else if e != nil {
		return nil, e
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		if e := rows.Scan(&id); e != nil {
			return nil, e
		}
		eids = append(eids, id)
	}
	return eids, rows.Err()
}

func (s *DefaultLikeStore) Delete(targetID int, targetType string) error {
	_, err := s.delete.Exec(targetID, targetType)
	return err
}

// TODO: Write a test for this
// Count returns the total number of likes globally
func (s *DefaultLikeStore) Count() (count int) {
	e := s.count.QueryRow().Scan(&count)
	if e != nil {
		LogError(e)
	}
	return count
}
