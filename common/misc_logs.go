package common

import (
	"database/sql"
	"time"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var RegLogs RegLogStore
var LoginLogs LoginLogStore

type RegLogItem struct {
	ID            int
	Username      string
	Email         string
	FailureReason string
	Success       bool
	IP            string
	DoneAt        string
}

type RegLogStmts struct {
	update *sql.Stmt
	create *sql.Stmt
}

var regLogStmts RegLogStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		rl := "registration_logs"
		regLogStmts = RegLogStmts{
			update: acc.Update(rl).Set("username=?,email=?,failureReason=?,success=?,doneAt=?").Where("rlid=?").Prepare(),
			create: acc.Insert(rl).Columns("username,email,failureReason,success,ipaddress,doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
// TODO: Do we even use this anymore or can we just make the logs immutable (except for deletes) for simplicity sake?
func (l *RegLogItem) Commit() error {
	_, e := regLogStmts.update.Exec(l.Username, l.Email, l.FailureReason, l.Success, l.DoneAt, l.ID)
	return e
}

func (l *RegLogItem) Create() (id int, e error) {
	id, e = Createf(regLogStmts.create, l.Username, l.Email, l.FailureReason, l.Success, l.IP)
	l.ID = id
	return l.ID, e
}

type RegLogStore interface {
	Count() (count int)
	GetOffset(offset, perPage int) (logs []RegLogItem, err error)
	Purge() error

	DeleteOlderThanDays(days int) error
}

type SQLRegLogStore struct {
	count     *sql.Stmt
	getOffset *sql.Stmt
	purge     *sql.Stmt

	deleteOlderThanDays *sql.Stmt
}

func NewRegLogStore(acc *qgen.Accumulator) (*SQLRegLogStore, error) {
	rl := "registration_logs"
	return &SQLRegLogStore{
		count:     acc.Count(rl).Prepare(),
		getOffset: acc.Select(rl).Columns("rlid,username,email,failureReason,success,ipaddress,doneAt").Orderby("doneAt DESC").Limit("?,?").Prepare(),
		purge:     acc.Purge(rl),

		deleteOlderThanDays: acc.Delete(rl).DateOlderThanQ("doneAt", "day").Prepare(),
	}, acc.FirstError()
}

func (s *SQLRegLogStore) Count() (count int) {
	return Count(s.count)
}

func (s *SQLRegLogStore) GetOffset(offset, perPage int) (logs []RegLogItem, e error) {
	rows, e := s.getOffset.Query(offset, perPage)
	if e != nil {
		return logs, e
	}
	defer rows.Close()

	for rows.Next() {
		var l RegLogItem
		var doneAt time.Time
		e := rows.Scan(&l.ID, &l.Username, &l.Email, &l.FailureReason, &l.Success, &l.IP, &doneAt)
		if e != nil {
			return logs, e
		}
		l.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SQLRegLogStore) DeleteOlderThanDays(days int) error {
	_, e := s.deleteOlderThanDays.Exec(days)
	return e
}

// Delete all registration logs
func (s *SQLRegLogStore) Purge() error {
	_, e := s.purge.Exec()
	return e
}

type LoginLogItem struct {
	ID      int
	UID     int
	Success bool
	IP      string
	DoneAt  string
}

type LoginLogStmts struct {
	update *sql.Stmt
	create *sql.Stmt
}

var loginLogStmts LoginLogStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		ll := "login_logs"
		loginLogStmts = LoginLogStmts{
			update: acc.Update(ll).Set("uid=?,success=?,doneAt=?").Where("lid=?").Prepare(),
			create: acc.Insert(ll).Columns("uid,success,ipaddress,doneAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
func (l *LoginLogItem) Commit() error {
	_, e := loginLogStmts.update.Exec(l.UID, l.Success, l.DoneAt, l.ID)
	return e
}

func (l *LoginLogItem) Create() (id int, e error) {
	res, e := loginLogStmts.create.Exec(l.UID, l.Success, l.IP)
	if e != nil {
		return 0, e
	}
	id64, e := res.LastInsertId()
	l.ID = int(id64)
	return l.ID, e
}

type LoginLogStore interface {
	Count() (count int)
	CountUser(uid int) (count int)
	GetOffset(uid, offset, perPage int) (logs []LoginLogItem, err error)
	Purge() error

	DeleteOlderThanDays(days int) error
}

type SQLLoginLogStore struct {
	count           *sql.Stmt
	countForUser    *sql.Stmt
	getOffsetByUser *sql.Stmt
	purge           *sql.Stmt

	deleteOlderThanDays *sql.Stmt
}

func NewLoginLogStore(acc *qgen.Accumulator) (*SQLLoginLogStore, error) {
	ll := "login_logs"
	return &SQLLoginLogStore{
		count:           acc.Count(ll).Prepare(),
		countForUser:    acc.Count(ll).Where("uid=?").Prepare(),
		getOffsetByUser: acc.Select(ll).Columns("lid,success,ipaddress,doneAt").Where("uid=?").Orderby("doneAt DESC").Limit("?,?").Prepare(),
		purge:           acc.Purge(ll),

		deleteOlderThanDays: acc.Delete(ll).DateOlderThanQ("doneAt", "day").Prepare(),
	}, acc.FirstError()
}

func (s *SQLLoginLogStore) Count() (count int) {
	return Count(s.count)
}

func (s *SQLLoginLogStore) CountUser(uid int) (count int) {
	return Countf(s.countForUser, uid)
}

func (s *SQLLoginLogStore) GetOffset(uid, offset, perPage int) (logs []LoginLogItem, e error) {
	rows, e := s.getOffsetByUser.Query(uid, offset, perPage)
	if e != nil {
		return logs, e
	}
	defer rows.Close()

	for rows.Next() {
		l := LoginLogItem{UID: uid}
		var doneAt time.Time
		e := rows.Scan(&l.ID, &l.Success, &l.IP, &doneAt)
		if e != nil {
			return logs, e
		}
		l.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SQLLoginLogStore) DeleteOlderThanDays(days int) error {
	_, e := s.deleteOlderThanDays.Exec(days)
	return e
}

// Delete all login logs
func (s *SQLLoginLogStore) Purge() error {
	_, e := s.purge.Exec()
	return e
}
