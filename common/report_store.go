package common

import (
	"database/sql"
	"errors"
	"strconv"

	"../query_gen/lib"
)

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
	return &DefaultReportStore{
		create: acc.Insert("topics").Columns("title, content, parsed_content, ipaddress, createdAt, lastReplyAt, createdBy, lastReplyBy, data, parentID, css_class").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,1,'report'").Prepare(),
		exists: acc.Count("topics").Where("data = ? AND data != '' AND parentID = 1").Prepare(),
	}, acc.FirstError()
}

// ! There's a data race in this. If two users report one item at the exact same time, then both reports will go through
func (store *DefaultReportStore) Create(title string, content string, user *User, itemType string, itemID int) (int, error) {
	var count int
	err := store.exists.QueryRow(itemType + "_" + strconv.Itoa(itemID)).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if count != 0 {
		return 0, ErrAlreadyReported
	}

	res, err := store.create.Exec(title, content, ParseMessage(content, 0, ""), user.LastIP, user.ID, user.ID, itemType+"_"+strconv.Itoa(itemID))
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), Forums.AddTopic(int(lastID), user.ID, 1)
}
