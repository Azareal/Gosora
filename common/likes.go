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
	count  *sql.Stmt
	delete *sql.Stmt
}

func NewDefaultLikeStore(acc *qgen.Accumulator) (*DefaultLikeStore, error) {
	return &DefaultLikeStore{
		count:  acc.Count("likes").Prepare(),
		delete: acc.Delete("likes").Where("targetItem=? AND targetType=?").Prepare(),
	}, acc.FirstError()
}

// TODO: Write a test for this
func (s *DefaultLikeStore) BulkExists(ids []int, sentBy int, targetType string) (eids []int, err error) {
	rows, err := qgen.NewAcc().Select("likes").Columns("targetItem").Where("sentBy=? AND targetType=?").In("targetItem", ids).Query(sentBy, targetType)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return nil, err
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
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
