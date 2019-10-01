package common

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
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
	queryZone    *sql.Stmt
}

// TODO: Support things other than MySQL
// TODO: Use LIMIT?
func NewSQLSearcher(acc *qgen.Accumulator) (*SQLSearcher, error) {
	if acc.GetAdapter().GetName() != "mysql" {
		return nil, errors.New("SQLSearcher only supports MySQL at this time")
	}
	return &SQLSearcher{
		queryReplies: acc.RawPrepare("SELECT `tid` FROM `replies` WHERE MATCH(content) AGAINST (? IN NATURAL LANGUAGE MODE);"),
		queryTopics:  acc.RawPrepare("SELECT `tid` FROM `topics` WHERE MATCH(title) AGAINST (? IN NATURAL LANGUAGE MODE) OR MATCH(content) AGAINST (? IN NATURAL LANGUAGE MODE);"),
		queryZone:    acc.RawPrepare("SELECT `topics`.`tid` FROM `topics` INNER JOIN `replies` ON `topics`.`tid` = `replies`.`tid` WHERE (MATCH(`topics`.`title`) AGAINST (? IN NATURAL LANGUAGE MODE) OR MATCH(`topics`.`content`) AGAINST (? IN NATURAL LANGUAGE MODE) OR MATCH(`replies`.`content`) AGAINST (? IN NATURAL LANGUAGE MODE)) AND `topics`.`parentID` = ?;"),
	}, acc.FirstError()
}

func (s *SQLSearcher) queryAll(q string) ([]int, error) {
	var ids []int
	run := func(stmt *sql.Stmt, q ...interface{}) error {
		rows, err := stmt.Query(q...)
		if err == sql.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			err := rows.Scan(&id)
			if err != nil {
				return err
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
	run := func(rows *sql.Rows, err error) error {
		if err == sql.ErrNoRows {
			return nil
		} else if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			err := rows.Scan(&id)
			if err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rows.Err()
	}

	if len(zones) == 1 {
		err = run(s.queryZone.Query(q, q, q, zones[0]))
	} else {
		var zList string
		for _, zone := range zones {
			zList += strconv.Itoa(zone) + ","
		}
		zList = zList[:len(zList)-1]

		acc := qgen.NewAcc()
		stmt := acc.RawPrepare("SELECT `topics`.`tid` FROM `topics` INNER JOIN `replies` ON `topics`.`tid` = `replies`.`tid` WHERE (MATCH(`topics`.`title`) AGAINST (? IN NATURAL LANGUAGE MODE) OR MATCH(`topics`.`content`) AGAINST (? IN NATURAL LANGUAGE MODE) OR MATCH(`replies`.`content`) AGAINST (? IN NATURAL LANGUAGE MODE)) AND `topics`.`parentID` IN(" + zList + ");")
		err := acc.FirstError()
		if err != nil {
			return nil, err
		}
		err = run(stmt.Query(q, q, q))
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
