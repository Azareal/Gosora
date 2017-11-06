package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	"./query_gen/lib"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Add the watchdog goroutine
// TODO: Add some sort of update method
var users UserStore
var errAccountExists = errors.New("this username is already in use")

type UserStore interface {
	Get(id int) (*User, error)
	Exists(id int) bool
	//BulkGet(ids []int) ([]*User, error)
	BulkGetMap(ids []int) (map[int]*User, error)
	BypassGet(id int) (*User, error)
	Create(username string, password string, email string, group int, active bool) (int, error)
	GlobalCount() int
}

type UserCache interface {
	CacheGet(id int) (*User, error)
	CacheGetUnsafe(id int) (*User, error)
	CacheSet(item *User) error
	CacheAdd(item *User) error
	CacheAddUnsafe(item *User) error
	CacheRemove(id int) error
	CacheRemoveUnsafe(id int) error
	Flush()
	Reload(id int) error
	Length() int
	SetCapacity(capacity int)
	GetCapacity() int
}

type MemoryUserStore struct {
	items          map[int]*User
	length         int64
	capacity       int
	get            *sql.Stmt
	exists         *sql.Stmt
	register       *sql.Stmt
	usernameExists *sql.Stmt
	userCount      *sql.Stmt
	sync.RWMutex
}

// NewMemoryUserStore gives you a new instance of MemoryUserStore
func NewMemoryUserStore(capacity int) (*MemoryUserStore, error) {
	acc := qgen.Builder.Accumulator()
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &MemoryUserStore{
		items:          make(map[int]*User),
		capacity:       capacity,
		get:            acc.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", ""),
		exists:         acc.SimpleSelect("users", "uid", "uid = ?", "", ""),
		register:       acc.SimpleInsert("users", "name, email, password, salt, group, is_super_admin, session, active, message, createdAt, lastActiveAt", "?,?,?,?,?,0,'',?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()"),
		usernameExists: acc.SimpleSelect("users", "name", "name = ?", "", ""),
		userCount:      acc.SimpleCount("users", "", ""),
	}, acc.FirstError()
}

func (mus *MemoryUserStore) CacheGet(id int) (*User, error) {
	mus.RLock()
	item, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mus *MemoryUserStore) CacheGetUnsafe(id int) (*User, error) {
	item, ok := mus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (mus *MemoryUserStore) Get(id int) (*User, error) {
	mus.RLock()
	user, ok := mus.items[id]
	mus.RUnlock()
	if ok {
		return user, nil
	}

	user = &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	user.Init()
	if err == nil {
		mus.CacheSet(user)
	}
	return user, err
}

// WARNING: We did a little hack to make this as thin and quick as possible to reduce lock contention, use the * Cascade* methods instead for normal use
func (mus *MemoryUserStore) bulkGet(ids []int) (list []*User) {
	list = make([]*User, len(ids))
	mus.RLock()
	for i, id := range ids {
		list[i] = mus.items[id]
	}
	mus.RUnlock()
	return list
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (mus *MemoryUserStore) BulkGetMap(ids []int) (list map[int]*User, err error) {
	var idCount = len(ids)
	list = make(map[int]*User)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := mus.bulkGet(ids)
	for i, sliceItem := range sliceList {
		if sliceItem != nil {
			list[sliceItem.ID] = sliceItem
		} else {
			stillHere = append(stillHere, ids[i])
		}
	}
	ids = stillHere

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	}

	// TODO: Add a function for the qlist stuff
	var qlist string
	var uidList []interface{}
	for _, id := range ids {
		uidList = append(uidList, strconv.Itoa(id))
		qlist += "?,"
	}
	qlist = qlist[0 : len(qlist)-1]

	stmt, err := qgen.Builder.SimpleSelect("users", "uid, name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid IN("+qlist+")", "", "")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(uidList...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		user := &User{Loggedin: true}
		err := rows.Scan(&user.ID, &user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
		if err != nil {
			return nil, err
		}

		// Initialise the user
		user.Init()

		// Add it to the cache...
		_ = mus.CacheSet(user)

		// Add it to the list to be returned
		list[user.ID] = user
	}

	// Did we miss any users?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			_, ok := list[id]
			if !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}

		// We probably don't need this, but it might be useful in case of bugs in BulkCascadeGetMap
		if sidList == "" {
			if dev.DebugMode {
				log.Print("This data is sampled later in the BulkCascadeGetMap function, so it might miss the cached IDs")
				log.Print("idCount", idCount)
				log.Print("ids", ids)
				log.Print("list", list)
			}
			return list, errors.New("We weren't able to find a user, but we don't know which one")
		}
		sidList = sidList[0 : len(sidList)-1]

		return list, errors.New("Unable to find the users with the following IDs: " + sidList)
	}

	return list, nil
}

func (mus *MemoryUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	user.Init()
	return user, err
}

func (mus *MemoryUserStore) Reload(id int) error {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
	if err != nil {
		mus.CacheRemove(id)
		return err
	}

	user.Init()
	_ = mus.CacheSet(user)
	return nil
}

func (mus *MemoryUserStore) Exists(id int) bool {
	err := mus.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

func (mus *MemoryUserStore) CacheSet(item *User) error {
	mus.Lock()
	user, ok := mus.items[item.ID]
	if ok {
		mus.Unlock()
		*user = *item
	} else if int(mus.length) >= mus.capacity {
		mus.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		mus.items[item.ID] = item
		mus.Unlock()
		atomic.AddInt64(&mus.length, 1)
	}
	return nil
}

func (mus *MemoryUserStore) CacheAdd(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.Lock()
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	mus.Unlock()
	return nil
}

func (mus *MemoryUserStore) CacheAddUnsafe(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	mus.length = int64(len(mus.items))
	return nil
}

func (mus *MemoryUserStore) CacheRemove(id int) error {
	mus.Lock()
	_, ok := mus.items[id]
	if !ok {
		mus.Unlock()
		return ErrNoRows
	}
	delete(mus.items, id)
	mus.Unlock()
	atomic.AddInt64(&mus.length, -1)
	return nil
}

func (mus *MemoryUserStore) CacheRemoveUnsafe(id int) error {
	_, ok := mus.items[id]
	if !ok {
		return ErrNoRows
	}
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

// TODO: Change active to a bool?
func (mus *MemoryUserStore) Create(username string, password string, email string, group int, active bool) (int, error) {
	// Is this username already taken..?
	err := mus.usernameExists.QueryRow(username).Scan(&username)
	if err != ErrNoRows {
		return 0, errAccountExists
	}

	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		return 0, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := mus.register.Exec(username, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (mus *MemoryUserStore) Flush() {
	mus.Lock()
	mus.items = make(map[int]*User)
	mus.length = 0
	mus.Unlock()
}

// ! Is this concurrent?
// Length returns the number of users in the memory cache
func (mus *MemoryUserStore) Length() int {
	return int(mus.length)
}

func (mus *MemoryUserStore) SetCapacity(capacity int) {
	mus.capacity = capacity
}

func (mus *MemoryUserStore) GetCapacity() int {
	return mus.capacity
}

// GlobalCount returns the total number of users registered on the forums
func (mus *MemoryUserStore) GlobalCount() (ucount int) {
	err := mus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}

type SQLUserStore struct {
	get            *sql.Stmt
	exists         *sql.Stmt
	register       *sql.Stmt
	usernameExists *sql.Stmt
	userCount      *sql.Stmt
}

func NewSQLUserStore() (*SQLUserStore, error) {
	acc := qgen.Builder.Accumulator()
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &SQLUserStore{
		get:            acc.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", ""),
		exists:         acc.SimpleSelect("users", "uid", "uid = ?", "", ""),
		register:       acc.SimpleInsert("users", "name, email, password, salt, group, is_super_admin, session, active, message, createdAt, lastActiveAt", "?,?,?,?,?,0,'',?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()"),
		usernameExists: acc.SimpleSelect("users", "name", "name = ?", "", ""),
		userCount:      acc.SimpleCount("users", "", ""),
	}, acc.FirstError()
}

func (mus *SQLUserStore) Get(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	user.Init()
	return user, err
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
func (mus *SQLUserStore) BulkGetMap(ids []int) (list map[int]*User, err error) {
	var qlist string
	var uidList []interface{}
	for _, id := range ids {
		uidList = append(uidList, strconv.Itoa(id))
		qlist += "?,"
	}
	qlist = qlist[0 : len(qlist)-1]

	stmt, err := qgen.Builder.SimpleSelect("users", "uid, name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid IN("+qlist+")", "", "")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(uidList...)
	if err != nil {
		return nil, err
	}

	list = make(map[int]*User)
	for rows.Next() {
		user := &User{Loggedin: true}
		err := rows.Scan(&user.ID, &user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
		if err != nil {
			return nil, err
		}

		// Initialise the user
		user.Init()

		// Add it to the list to be returned
		list[user.ID] = user
	}

	return list, nil
}

func (mus *SQLUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	user.Init()
	return user, err
}

func (mus *SQLUserStore) Exists(id int) bool {
	err := mus.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

func (mus *SQLUserStore) Create(username string, password string, email string, group int, active bool) (int, error) {
	// Is this username already taken..?
	err := mus.usernameExists.QueryRow(username).Scan(&username)
	if err != ErrNoRows {
		return 0, errAccountExists
	}

	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		return 0, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := mus.register.Exec(username, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

// GlobalCount returns the total number of users registered on the forums
func (mus *SQLUserStore) GlobalCount() (ucount int) {
	err := mus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}

// TODO: MockUserStore

// NullUserStore is here for tests because Go doesn't have short-circuiting
type NullUserStore struct {
}

func (nus *NullUserStore) CacheGet(_ int) (*User, error) {
	return nil, ErrNoRows
}

func (nus *NullUserStore) CacheGetUnsafe(_ int) (*User, error) {
	return nil, ErrNoRows
}

func (nus *NullUserStore) CacheSet(_ *User) error {
	return ErrStoreCapacityOverflow
}

func (nus *NullUserStore) CacheAdd(_ *User) error {
	return ErrStoreCapacityOverflow
}

func (nus *NullUserStore) CacheAddUnsafe(_ *User) error {
	return ErrStoreCapacityOverflow
}

func (nus *NullUserStore) CacheRemove(_ int) error {
	return ErrNoRows
}

func (nus *NullUserStore) CacheRemoveUnsafe(_ int) error {
	return ErrNoRows
}

func (nus *NullUserStore) Flush() {
}

func (nus *NullUserStore) Reload(_ int) error {
	return ErrNoRows
}

func (nus *NullUserStore) Length() int {
	return 0
}

func (nus *NullUserStore) SetCapacity(_ int) {
}

func (nus *NullUserStore) GetCapacity() int {
	return 0
}
