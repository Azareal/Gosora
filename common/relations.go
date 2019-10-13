package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

type BlockStore interface {
	IsBlockedBy(blocker, blockee int) (bool, error)
	Add(blocker, blockee int) error
}

type DefaultBlockStore struct {
	isBlocked *sql.Stmt
	add       *sql.Stmt
}

func NewDefaultBlockStore(acc *qgen.Accumulator) (*DefaultBlockStore, error) {
	return &DefaultBlockStore{
		isBlocked: acc.Select("users_blocks").Cols("blocker").Where("blocker = ? AND blockedUser = ?").Prepare(),
		add:       acc.Insert("users_blocks").Columns("blocker,blockedUser").Fields("?,?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultBlockStore) IsBlockedBy(blocker, blockee int) (bool, error) {
	err := s.isBlocked.QueryRow(blocker, blockee).Scan(&blocker)
	if err != nil && err != ErrNoRows {
		return false, err
	}
	return err != ErrNoRows, nil
}

func (s *DefaultBlockStore) Add(blocker, blockee int) error {
	return nil
}

type FriendInvite struct {
	Requester int
	Target    int
}

type FriendStore interface {
	AddInvite(requester, target int) error
	Confirm(requester, target int) error
	GetOwSentInvites(uid int) ([]FriendInvite, error)
	GetOwnRecvInvites(uid int) ([]FriendInvite, error)
}

type DefaultFriendStore struct {
	addInvite         *sql.Stmt
	confirm           *sql.Stmt
	getOwnSentInvites *sql.Stmt
	getOwnRecvInvites *sql.Stmt
}

func NewDefaultFriendStore(acc *qgen.Accumulator) (*DefaultFriendStore, error) {
	return &DefaultFriendStore{
		addInvite:         acc.Insert("users_friends_invites").Columns("requester, target").Fields("?,?").Prepare(),
		confirm:           acc.Insert("users_friends").Columns("uid,uid2").Fields("?,?").Prepare(),
		getOwnSentInvites: acc.Select("users_friends_invites").Cols("requester, target").Where("requester = ?").Prepare(),
		getOwnRecvInvites: acc.Select("users_friends_invites").Cols("requester, target").Where("target = ?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultFriendStore) AddInvite(requester, target int) error {
	return nil
}
func (s *DefaultFriendStore) Confirm(requester, target int) error {
	return nil
}
func (s *DefaultFriendStore) GetOwnSentInvites(uid int) ([]FriendInvite, error) {
	return nil, nil
}
func (s *DefaultFriendStore) GetOwnRecvInvites(uid int) ([]FriendInvite, error) {
	return nil, nil
}
