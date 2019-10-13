package common

import qgen "github.com/Azareal/Gosora/query_gen"

type BlockStore interface {
	IsBlockedBy(blocker int, blockee int) (bool, error)
}

type DefaultBlockStore struct {
}

func NewDefaultBlockStore(acc *qgen.Accumulator) (*DefaultBlockStore, error) {
	return &DefaultBlockStore{}, acc.FirstError()
}

func (s *DefaultBlockStore) IsBlockedBy(blocker int, blockee int) (bool, error) {
	return false, nil
}

type FriendStore interface {
}

type DefaultFriendStore struct {
}

func NewDefaultFriendStore(acc *qgen.Accumulator) (*DefaultFriendStore, error) {
	return &DefaultFriendStore{}, acc.FirstError()
}
