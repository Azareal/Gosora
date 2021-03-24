package common

import (
	"database/sql"
	"errors"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var RepliesSearch Searcher

type Searcher interface {
	Query(q string, zones []int) ([]int, error)
}

// TODO: Implement this
// Note: This is slow compared to something like ElasticSearch and very limited
type SQLSearcher struct {
	queryReplies *sql.Stmt
	queryTopics  *sql.Stmt
	queryRepliesZone *sql.Stmt
	queryTopicsZone *sql.Stmt
	//queryZone    *sql.Stmt
	fuzzyZone *sql.Stmt
}

// TODO: Support things other than MySQL
// TODO: Use LIMIT?
func NewSQLSearcher(acc *qgen.Accumulator) (*SQLSearcher, error) {
	if acc.GetAdapter().GetName() != "mysql" {
		return nil, errors.New("SQLSearcher only supports MySQL at this time")
	}
	return &SQLSearcher{
		queryReplies: acc.RawPrepare("SELECT tid FROM replies WHERE MATCH(content) AGAINST (? IN BOOLEAN MODE)"),
		queryTopics:  acc.RawPrepare("SELECT tid FROM topics WHERE MATCH(title) AGAINST (? IN BOOLEAN MODE) OR MATCH(content) AGAINST (? IN BOOLEAN MODE)"),
		queryRepliesZone: acc.RawPrepare("SELECT tid FROM replies WHERE MATCH(content) AGAINST (? IN BOOLEAN MODE) AND tid=?"),
		queryTopicsZone:  acc.RawPrepare("SELECT tid FROM topics WHERE (MATCH(title) AGAINST (? IN BOOLEAN MODE) OR MATCH(content) AGAINST (? IN BOOLEAN MODE)) AND parentID=?"),
		//queryZone:    acc.RawPrepare("SELECT topics.tid FROM topics INNER JOIN replies ON topics.tid = replies.tid WHERE (topics.title=? OR (MATCH(topics.title) AGAINST (? IN BOOLEAN MODE) OR MATCH(topics.content) AGAINST (? IN BOOLEAN MODE) OR MATCH(replies.content) AGAINST (? IN BOOLEAN MODE)) OR topics.content=? OR replies.content=?) AND topics.parentID=?"),
		fuzzyZone:    acc.RawPrepare("SELECT topics.tid FROM topics INNER JOIN replies ON topics.tid = replies.tid WHERE (topics.title LIKE ? OR topics.content LIKE ? OR replies.content LIKE ?) AND topics.parentID=?"),
	}, acc.FirstError()
}

func (s *SQLSearcher) queryAll(q string) ([]int, error) {
	var ids []int
	run := func(stmt *sql.Stmt, q ...interface{}) error {
		rows, e := stmt.Query(q...)
		if e == sql.ErrNoRows {
			return nil
		} else if e != nil {
			return e
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			if e := rows.Scan(&id); e != nil {
				return e
			}
			ids = append(ids, id)
		}
		return rows.Err()
	}

	err := run(s.queryReplies, q)
	if err != nil {
		return nil, err
	}
	err = run(s.queryTopics, q, q)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		err = sql.ErrNoRows
	}
	return ids, err
}

func (s *SQLSearcher) Query(q string, zones []int) (ids []int, err error) {
	if len(zones) == 0 {
		return nil, nil
	}
	run := func(rows *sql.Rows, e error) error {
		/*if e == sql.ErrNoRows {
			return nil
		} else */if e != nil {
			return e
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			if e := rows.Scan(&id); e != nil {
				return e
			}
			ids = append(ids, id)
		}
		return rows.Err()
	}

	if len(zones) == 1 {
		//err = run(s.queryZone.Query(q, q, q, q, q,q, zones[0]))
		err = run(s.queryRepliesZone.Query(q, zones[0]))
		if err != nil {
			return nil, err
		}
		err = run(s.queryTopicsZone.Query(q, q,zones[0]))
	} else {
		var zList string
		for _, zone := range zones {
			zList += strconv.Itoa(zone) + ","
		}
		zList = zList[:len(zList)-1]

		acc := qgen.NewAcc()
		/*stmt := acc.RawPrepare("SELECT topics.tid FROM topics INNER JOIN replies ON topics.tid = replies.tid WHERE (MATCH(topics.title) AGAINST (? IN BOOLEAN MODE) OR MATCH(topics.content) AGAINST (? IN BOOLEAN MODE) OR MATCH(replies.content) AGAINST (? IN BOOLEAN MODE) OR topics.title=? OR topics.content=? OR replies.content=?) AND topics.parentID IN(" + zList + ")")
		if err = acc.FirstError(); err != nil {
			return nil, err
		}*/
		// TODO: Cache common IN counts
		stmt := acc.RawPrepare("SELECT tid FROM topics WHERE (MATCH(topics.title) AGAINST (? IN BOOLEAN MODE) OR MATCH(topics.content) AGAINST (? IN BOOLEAN MODE)) AND parentID IN(" + zList + ")")
		if err = acc.FirstError(); err != nil {
			return nil, err
		}
		err = run(stmt.Query(q, q))
		if err != nil {
			return nil, err
		}
		stmt = acc.RawPrepare("SELECT tid FROM replies WHERE MATCH(replies.content) AGAINST (? IN BOOLEAN MODE) AND tid IN(" + zList + ")")
		if err = acc.FirstError(); err != nil {
			return nil, err
		}
		err = run(stmt.Query(q))
		//err = run(stmt.Query(q, q, q, q, q, q))
	}
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		err = sql.ErrNoRows
	}
	return ids, err
}

// TODO: Implement this
type ElasticSearchSearcher struct {
}

func NewElasticSearchSearcher() (*ElasticSearchSearcher, error) {
	return &ElasticSearchSearcher{}, nil
}

func (s *ElasticSearchSearcher) Query(q string, zones []int) ([]int, error) {
	return nil, nil
}
