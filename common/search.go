package common

import (
	"database/sql"
	"errors"

	"github.com/Azareal/Gosora/query_gen"
)

//var RepliesSearch Searcher

type Searcher interface {
	Query(q string) ([]int, error)
}

type ZoneSearcher interface {
	QueryZone(q string, zoneID int) ([]int, error)
}

// TODO: Implement this
// Note: This is slow compared to something like ElasticSearch and very limited
type SQLSearcher struct {
	queryReplies     *sql.Stmt
	queryTopics      *sql.Stmt
	queryZoneReplies *sql.Stmt
	queryZoneTopics  *sql.Stmt
}

// TODO: Support things other than MySQL
func NewSQLSearcher(acc *qgen.Accumulator) (*SQLSearcher, error) {
	if acc.GetAdapter().GetName() != "mysql" {
		return nil, errors.New("SQLSearcher only supports MySQL at this time")
	}
	return &SQLSearcher{
		queryReplies:     acc.RawPrepare("SELECT `rid` FROM `replies` WHERE MATCH(content) AGAINST (? IN NATURAL LANGUAGE MODE);"),
		queryTopics:      acc.RawPrepare("SELECT `tid` FROM `topics` WHERE MATCH(title,content) AGAINST (? IN NATURAL LANGUAGE MODE);"),
		queryZoneReplies: acc.RawPrepare("SELECT `rid` FROM `replies` WHERE MATCH(content) AGAINST (? IN NATURAL LANGUAGE MODE) AND `parentID` = ?;"),
		queryZoneTopics:  acc.RawPrepare("SELECT `tid` FROM `topics` WHERE MATCH(title,content) AGAINST (? IN NATURAL LANGUAGE MODE) AND `parentID` = ?;"),
	}, acc.FirstError()
}

func (searcher *SQLSearcher) Query(q string) ([]int, error) {
	return nil, nil

	/*
		rows, err := stmt.Query(q)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	*/
}

func (searcher *SQLSearcher) QueryZone(q string, zoneID int) ([]int, error) {
	return nil, nil
}

// TODO: Implement this
type ElasticSearchSearcher struct {
}

func NewElasticSearchSearcher() *ElasticSearchSearcher {
	return &ElasticSearchSearcher{}
}

func (searcher *ElasticSearchSearcher) Query(q string) ([]int, error) {
	return nil, nil
}

func (searcher *ElasticSearchSearcher) QueryZone(q string, zoneID int) ([]int, error) {
	return nil, nil
}
