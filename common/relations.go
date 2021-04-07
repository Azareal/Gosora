package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var UserBlocks BlockStore

//var UserFriends FriendStore

type BlockStore interface {
	IsBlockedBy(blocker, blockee int) (bool, error)
	BulkIsBlockedBy(blockers []int, blockee int) (bool, error)
	Add(blocker, blockee int) error
	Remove(blocker, blockee int) error
	BlockedByOffset(blocker, offset, perPage int) ([]int, error)
	BlockedByCount(blocker int) int
}

type DefaultBlockStore struct {
	isBlocked      *sql.Stmt
	add            *sql.Stmt
	remove         *sql.Stmt
	blockedBy      *sql.Stmt
	blockedByCount *sql.Stmt
}

func NewDefaultBlockStore(acc *qgen.Accumulator) (*DefaultBlockStore, error) {
	ub := "users_blocks"
	return &DefaultBlockStore{
		isBlocked:      acc.Select(ub).Cols("blocker").Where("blocker=? AND blockedUser=?").Prepare(),
		add:            acc.Insert(ub).Columns("blocker,blockedUser").Fields("?,?").Prepare(),
		remove:         acc.Delete(ub).Where("blocker=? AND blockedUser=?").Prepare(),
		blockedBy:      acc.Select(ub).Columns("blockedUser").Where("blocker=?").Limit("?,?").Prepare(),
		blockedByCount: acc.Count(ub).Where("blocker=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultBlockStore) IsBlockedBy(blocker, blockee int) (bool, error) {
	e := s.isBlocked.QueryRow(blocker, blockee).Scan(&blocker)
	if e == ErrNoRows {
		return false, nil
	}
	return e == nil, e
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
func (s *DefaultBlockStore) BulkIsBlockedBy(blockers []int, blockee int) (bool, error) {
	if len(blockers) == 0 {
		return false, nil
	}
	if len(blockers) == 1 {
		return s.IsBlockedBy(blockers[0], blockee)
	}
	idList, q := inqbuild(blockers)
	count, e := qgen.NewAcc().Count("users_blocks").Where("blocker IN(" + q + ") AND blockedUser=?").TotalP(idList...)
	if e == ErrNoRows {
		return false, nil
	}
	return count == 0, e
}

func (s *DefaultBlockStore) Add(blocker, blockee int) error {
	_, e := s.add.Exec(blocker, blockee)
	return e
}

func (s *DefaultBlockStore) Remove(blocker, blockee int) error {
	_, e := s.remove.Exec(blocker, blockee)
	return e
}

func (s *DefaultBlockStore) BlockedByOffset(blocker, offset, perPage int) (uids []int, err error) {
	rows, e := s.blockedBy.Query(blocker, offset, perPage)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var uid int
		if e := rows.Scan(&uid); e != nil {
			return nil, e
		}
		uids = append(uids, uid)
	}
	return uids, rows.Err()
}

func (s *DefaultBlockStore) BlockedByCount(blocker int) (count int) {
	e := s.blockedByCount.QueryRow(blocker).Scan(&count)
	if e != nil {
		LogError(e)
	}
	return count
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
	ufi := "users_friends_invites"
	return &DefaultFriendStore{
		addInvite:         acc.Insert(ufi).Columns("requester,target").Fields("?,?").Prepare(),
		confirm:           acc.Insert("users_friends").Columns("uid,uid2").Fields("?,?").Prepare(),
		getOwnSentInvites: acc.Select(ufi).Cols("requester,target").Where("requester=?").Prepare(),
		getOwnRecvInvites: acc.Select(ufi).Cols("requester,target").Where("target=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultFriendStore) AddInvite(requester, target int) error {
	_, e := s.addInvite.Exec(requester, target)
	return e
}

func (s *DefaultFriendStore) Confirm(requester, target int) error {
	_, e := s.confirm.Exec(requester, target)
	return e
}

func (s *DefaultFriendStore) GetOwnSentInvites(uid int) ([]FriendInvite, error) {
	return nil, nil
}
func (s *DefaultFriendStore) GetOwnRecvInvites(uid int) ([]FriendInvite, error) {
	return nil, nil
}
