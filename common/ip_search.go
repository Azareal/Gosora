package common

import (
	"database/sql"

	"../query_gen/lib"
)

var IPSearch IPSearcher

type IPSearcher interface {
	Lookup(ip string) (uids []int, err error)
}

type DefaultIPSearcher struct {
	searchUsers        *sql.Stmt
	searchTopics       *sql.Stmt
	searchReplies      *sql.Stmt
	searchUsersReplies *sql.Stmt
}

// NewDefaultIPSearcher gives you a new instance of DefaultIPSearcher
func NewDefaultIPSearcher() (*DefaultIPSearcher, error) {
	acc := qgen.Builder.Accumulator()
	return &DefaultIPSearcher{
		searchUsers:        acc.Select("users").Columns("uid").Where("last_ip = ?").Prepare(),
		searchTopics:       acc.Select("users").Columns("uid").InQ("uid", acc.Select("topics").Columns("createdBy").Where("ipaddress = ?")).Prepare(),
		searchReplies:      acc.Select("users").Columns("uid").InQ("uid", acc.Select("replies").Columns("createdBy").Where("ipaddress = ?")).Prepare(),
		searchUsersReplies: acc.Select("users").Columns("uid").InQ("uid", acc.Select("users_replies").Columns("createdBy").Where("ipaddress = ?")).Prepare(),
	}, acc.FirstError()
}

func (searcher *DefaultIPSearcher) Lookup(ip string) (uids []int, err error) {
	var uid int
	var reqUserList = make(map[int]bool)

	var runQuery = func(stmt *sql.Stmt) error {
		rows, err := stmt.Query(ip)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&uid)
			if err != nil {
				return err
			}
			reqUserList[uid] = true
		}
		return rows.Err()
	}

	err = runQuery(searcher.searchUsers)
	if err != nil {
		return uids, err
	}
	err = runQuery(searcher.searchTopics)
	if err != nil {
		return uids, err
	}
	err = runQuery(searcher.searchReplies)
	if err != nil {
		return uids, err
	}
	err = runQuery(searcher.searchUsersReplies)
	if err != nil {
		return uids, err
	}

	// Convert the user ID map to a slice, then bulk load the users
	uids = make([]int, len(reqUserList))
	var i int
	for userID := range reqUserList {
		uids[i] = userID
		i++
	}

	return uids, nil
}
