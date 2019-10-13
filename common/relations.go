package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type BlockStore interface {
	IsBlockedBy(blocker int, blockee int) (bool, error)
}

type DefaultBlockStore struct {
	isBlocked *sql.Stmt
}

func NewDefaultBlockStore(acc *qgen.Accumulator) (*DefaultBlockStore, error) {
	return &DefaultBlockStore{
		isBlocked: acc.Select("users_blocks").Cols("blocker").Where("blocker = ? AND blockedUser = ?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultBlockStore) IsBlockedBy(blocker int, blockee int) (bool, error) {
	err := s.isBlocked.QueryRow(blocker, blockee).Scan(&blocker)
	if err != nil && err != ErrNoRows {
		return false, err
	}
	return err != ErrNoRows, nil
}

type FriendStore interface {
}

type DefaultFriendStore struct {
}

func NewDefaultFriendStore(acc *qgen.Accumulator) (*DefaultFriendStore, error) {
	return &DefaultFriendStore{}, acc.FirstError()
}
