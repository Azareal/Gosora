package common

import (
	"database/sql"
	"errors"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

// TODO: Make the default report forum ID configurable
// TODO: Make sure this constant is used everywhere for the report forum ID
const ReportForumID = 1

var Reports ReportStore
var ErrAlreadyReported = errors.New("This item has already been reported")

// The report system mostly wraps around the topic system for simplicty
type ReportStore interface {
	Create(title string, content string, user *User, itemType string, itemID int) (int, error)
}

type DefaultReportStore struct {
	create *sql.Stmt
	exists *sql.Stmt
}

func NewDefaultReportStore(acc *qgen.Accumulator) (*DefaultReportStore, error) {
	t := "topics"
	return &DefaultReportStore{
		create: acc.Insert(t).Columns("title, content, parsed_content, ipaddress, createdAt, lastReplyAt, createdBy, lastReplyBy, data, parentID, css_class").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?,'report'").Prepare(),
		exists: acc.Count(t).Where("data = ? AND data != '' AND parentID = ?").Prepare(),
	}, acc.FirstError()
}

// ! There's a data race in this. If two users report one item at the exact same time, then both reports will go through
func (s *DefaultReportStore) Create(title string, content string, user *User, itemType string, itemID int) (tid int, err error) {
	var count int
	err = s.exists.QueryRow(itemType+"_"+strconv.Itoa(itemID), ReportForumID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if count != 0 {
		return 0, ErrAlreadyReported
	}

	res, err := s.create.Exec(title, content, ParseMessage(content, 0, ""), user.LastIP, user.ID, user.ID, itemType+"_"+strconv.Itoa(itemID), ReportForumID)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	tid = int(lastID)
	return tid, Forums.AddTopic(tid, user.ID, ReportForumID)
}
