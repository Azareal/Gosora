package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var IPSearch IPSearcher

type IPSearcher interface {
	Lookup(ip string) (uids []int, e error)
}

type DefaultIPSearcher struct {
	searchUsers        *sql.Stmt
	searchTopics       *sql.Stmt
	searchReplies      *sql.Stmt
	searchUsersReplies *sql.Stmt
}

// NewDefaultIPSearcher gives you a new instance of DefaultIPSearcher
func NewDefaultIPSearcher() (*DefaultIPSearcher, error) {
	acc := qgen.NewAcc()
	uu := "users"
	q := func(tbl string) *sql.Stmt {
		return acc.Select(uu).Columns("uid").InQ("uid", acc.Select(tbl).Columns("createdBy").Where("ip=?")).Prepare()
	}
	return &DefaultIPSearcher{
		searchUsers:        acc.Select(uu).Columns("uid").Where("last_ip=? OR last_ip LIKE CONCAT('%-',?)").Prepare(),
		searchTopics:       q("topics"),
		searchReplies:      q("replies"),
		searchUsersReplies: q("users_replies"),
	}, acc.FirstError()
}

func (s *DefaultIPSearcher) Lookup(ip string) (uids []int, e error) {
	var uid int
	reqUserList := make(map[int]bool)
	runQuery2 := func(rows *sql.Rows, e error) error {
		if e != nil {
			return e
		}
		defer rows.Close()

		for rows.Next() {
			if e := rows.Scan(&uid); e != nil {
				return e
			}
			reqUserList[uid] = true
		}
		return rows.Err()
	}
	runQuery := func(stmt *sql.Stmt) error {
		return runQuery2(stmt.Query(ip))
	}

	e = runQuery2(s.searchUsers.Query(ip, ip))
	if e != nil {
		return uids, e
	}
	e = runQuery(s.searchTopics)
	if e != nil {
		return uids, e
	}
	e = runQuery(s.searchReplies)
	if e != nil {
		return uids, e
	}
	e = runQuery(s.searchUsersReplies)
	if e != nil {
		return uids, e
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
