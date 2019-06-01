package common

import (
	"database/sql"
	"time"

	"github.com/Azareal/Gosora/query_gen"
)

var RegLogs RegLogStore
var LoginLogs LoginLogStore

type RegLogItem struct {
	ID            int
	Username      string
	Email         string
	FailureReason string
	Success       bool
	IPAddress     string
	DoneAt        string
}

type RegLogStmts struct {
	update *sql.Stmt
	create *sql.Stmt
}

var regLogStmts RegLogStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		regLogStmts = RegLogStmts{
			update: acc.Update("registration_logs").Set("username = ?, email = ?, failureReason = ?, success = ?").Where("rlid = ?").Prepare(),
			create: acc.Insert("registration_logs").Columns("username, email, failureReason, success, ipaddress, doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
func (log *RegLogItem) Commit() error {
	_, err := regLogStmts.update.Exec(log.Username, log.Email, log.FailureReason, log.Success, log.ID)
	return err
}

func (log *RegLogItem) Create() (id int, err error) {
	res, err := regLogStmts.create.Exec(log.Username, log.Email, log.FailureReason, log.Success, log.IPAddress)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	log.ID = int(id64)
	return log.ID, err
}

type RegLogStore interface {
	Count() (count int)
	GetOffset(offset int, perPage int) (logs []RegLogItem, err error)
}

type SQLRegLogStore struct {
	count     *sql.Stmt
	getOffset *sql.Stmt
}

func NewRegLogStore(acc *qgen.Accumulator) (*SQLRegLogStore, error) {
	return &SQLRegLogStore{
		count:     acc.Count("registration_logs").Prepare(),
		getOffset: acc.Select("registration_logs").Columns("rlid, username, email, failureReason, success, ipaddress, doneAt").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

func (s *SQLRegLogStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (store *SQLRegLogStore) GetOffset(offset int, perPage int) (logs []RegLogItem, err error) {
	rows, err := store.getOffset.Query(offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var log RegLogItem
		var doneAt time.Time
		err := rows.Scan(&log.ID, &log.Username, &log.Email, &log.FailureReason, &log.Success, &log.IPAddress, &doneAt)
		if err != nil {
			return logs, err
		}
		log.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

type LoginLogItem struct {
	ID        int
	UID       int
	Success   bool
	IPAddress string
	DoneAt    string
}

type LoginLogStmts struct {
	update *sql.Stmt
	create *sql.Stmt
}

var loginLogStmts LoginLogStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		loginLogStmts = LoginLogStmts{
			update: acc.Update("login_logs").Set("uid = ?, success = ?").Where("lid = ?").Prepare(),
			create: acc.Insert("login_logs").Columns("uid, success, ipaddress, doneAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
func (log *LoginLogItem) Commit() error {
	_, err := loginLogStmts.update.Exec(log.UID, log.Success, log.ID)
	return err
}

func (log *LoginLogItem) Create() (id int, err error) {
	res, err := loginLogStmts.create.Exec(log.UID, log.Success, log.IPAddress)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	log.ID = int(id64)
	return log.ID, err
}

type LoginLogStore interface {
	Count() (count int)
	CountUser(uid int) (count int)
	GetOffset(uid int, offset int, perPage int) (logs []LoginLogItem, err error)
}

type SQLLoginLogStore struct {
	count           *sql.Stmt
	countForUser    *sql.Stmt
	getOffsetByUser *sql.Stmt
}

func NewLoginLogStore(acc *qgen.Accumulator) (*SQLLoginLogStore, error) {
	return &SQLLoginLogStore{
		count:           acc.Count("login_logs").Prepare(),
		countForUser:    acc.Count("login_logs").Where("uid = ?").Prepare(),
		getOffsetByUser: acc.Select("login_logs").Columns("lid, success, ipaddress, doneAt").Where("uid = ?").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

func (s *SQLLoginLogStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *SQLLoginLogStore) CountUser(uid int) (count int) {
	err := s.countForUser.QueryRow(uid).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (store *SQLLoginLogStore) GetOffset(uid int, offset int, perPage int) (logs []LoginLogItem, err error) {
	rows, err := store.getOffsetByUser.Query(uid, offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var log = LoginLogItem{UID: uid}
		var doneAt time.Time
		err := rows.Scan(&log.ID, &log.Success, &log.IPAddress, &doneAt)
		if err != nil {
			return logs, err
		}
		log.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
