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
			update: acc.Update(rl).Set("username=?,email=?,failureReason=?,success=?").Where("rlid=?").Prepare(),
			create: acc.Insert(rl).Columns("username,email,failureReason,success,ipaddress,doneAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
func (l *RegLogItem) Commit() error {
	_, err := regLogStmts.update.Exec(l.Username, l.Email, l.FailureReason, l.Success, l.ID)
	return err
}

func (l *RegLogItem) Create() (id int, err error) {
	res, err := regLogStmts.create.Exec(l.Username, l.Email, l.FailureReason, l.Success, l.IP)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	l.ID = int(id64)
	return l.ID, err
}

type RegLogStore interface {
	Count() (count int)
	GetOffset(offset, perPage int) (logs []RegLogItem, err error)
}

type SQLRegLogStore struct {
	count     *sql.Stmt
	getOffset *sql.Stmt
}

func NewRegLogStore(acc *qgen.Accumulator) (*SQLRegLogStore, error) {
	rl := "registration_logs"
	return &SQLRegLogStore{
		count:     acc.Count(rl).Prepare(),
		getOffset: acc.Select(rl).Columns("rlid,username,email,failureReason,success,ipaddress,doneAt").Orderby("doneAt DESC").Limit("?,?").Prepare(),
	}, acc.FirstError()
}

func (s *SQLRegLogStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *SQLRegLogStore) GetOffset(offset, perPage int) (logs []RegLogItem, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		var l RegLogItem
		var doneAt time.Time
		err := rows.Scan(&l.ID, &l.Username, &l.Email, &l.FailureReason, &l.Success, &l.IP, &doneAt)
		if err != nil {
			return logs, err
		}
		l.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, l)
	}
	return logs, rows.Err()
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
			update: acc.Update(ll).Set("uid=?,success=?").Where("lid=?").Prepare(),
			create: acc.Insert(ll).Columns("uid,success,ipaddress,doneAt").Fields("?,?,?,UTC_TIMESTAMP()").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: Reload this item in the store, probably doesn't matter right now, but it might when we start caching this stuff in memory
// ! Retroactive updates of date are not permitted for integrity reasons
func (l *LoginLogItem) Commit() error {
	_, err := loginLogStmts.update.Exec(l.UID, l.Success, l.ID)
	return err
}

func (l *LoginLogItem) Create() (id int, err error) {
	res, err := loginLogStmts.create.Exec(l.UID, l.Success, l.IP)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	l.ID = int(id64)
	return l.ID, err
}

type LoginLogStore interface {
	Count() (count int)
	CountUser(uid int) (count int)
	GetOffset(uid, offset, perPage int) (logs []LoginLogItem, err error)
}

type SQLLoginLogStore struct {
	count           *sql.Stmt
	countForUser    *sql.Stmt
	getOffsetByUser *sql.Stmt
}

func NewLoginLogStore(acc *qgen.Accumulator) (*SQLLoginLogStore, error) {
	ll := "login_logs"
	return &SQLLoginLogStore{
		count:           acc.Count(ll).Prepare(),
		countForUser:    acc.Count(ll).Where("uid=?").Prepare(),
		getOffsetByUser: acc.Select(ll).Columns("lid,success,ipaddress,doneAt").Where("uid=?").Orderby("doneAt DESC").Limit("?,?").Prepare(),
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

func (s *SQLLoginLogStore) GetOffset(uid, offset, perPage int) (logs []LoginLogItem, err error) {
	rows, err := s.getOffsetByUser.Query(uid, offset, perPage)
	if err != nil {
		return logs, err
	}
	defer rows.Close()

	for rows.Next() {
		l := LoginLogItem{UID: uid}
		var doneAt time.Time
		err := rows.Scan(&l.ID, &l.Success, &l.IP, &doneAt)
		if err != nil {
			return logs, err
		}
		l.DoneAt = doneAt.Format("2006-01-02 15:04:05")
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
