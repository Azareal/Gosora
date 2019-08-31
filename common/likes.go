package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Likes LikeStore

type LikeStore interface {
	BulkExists(ids []int, sentBy int, targetType string) ([]int, error)
	Count() (count int)
}

type DefaultLikeStore struct {
	count *sql.Stmt
}

func NewDefaultLikeStore(acc *qgen.Accumulator) (*DefaultLikeStore, error) {
	return &DefaultLikeStore{
		count: acc.Count("likes").Prepare(),
	}, acc.FirstError()
}

// TODO: Write a test for this
func (s *DefaultLikeStore) BulkExists(ids []int, sentBy int, targetType string) (eids []int, err error) {
	rows, err := qgen.NewAcc().Select("likes").Columns("targetItem").Where("sentBy = ? AND targetType = ?").In("targetItem", ids).Query(sentBy,targetType)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		eids = append(eids,id)
	}
	return eids, rows.Err()
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