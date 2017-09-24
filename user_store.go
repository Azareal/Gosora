package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
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
	Create(username string, password string, email string, group int, active int) (int, error)
	GetGlobalCount() int
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
	GetLength() int
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
func NewMemoryUserStore(capacity int) *MemoryUserStore {
	getStmt, err := qgen.Builder.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	existsStmt, err := qgen.Builder.SimpleSelect("users", "uid", "uid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	registerStmt, err := qgen.Builder.SimpleInsert("users", "name, email, password, salt, group, is_super_admin, session, active, message", "?,?,?,?,?,0,'',?,''")
	if err != nil {
		log.Fatal(err)
	}

	usernameExistsStmt, err := qgen.Builder.SimpleSelect("users", "name", "name = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	userCountStmt, err := qgen.Builder.SimpleCount("users", "", "")
	if err != nil {
		log.Fatal(err)
	}

	return &MemoryUserStore{
		items:          make(map[int]*User),
		capacity:       capacity,
		get:            getStmt,
		exists:         existsStmt,
		register:       registerStmt,
		usernameExists: usernameExistsStmt,
		userCount:      userCountStmt,
	}
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

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = gstore.DirtyGet(user.Group).Tag
	user.initPerms()
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
		if user.Avatar != "" {
			if user.Avatar[0] == '.' {
				user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
			}
		} else {
			user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
		}
		user.Link = buildProfileURL(nameToSlug(user.Name), user.ID)
		user.Tag = gstore.DirtyGet(user.Group).Tag
		user.initPerms()

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

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = gstore.DirtyGet(user.Group).Tag
	user.initPerms()
	return user, err
}

func (mus *MemoryUserStore) Reload(id int) error {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
	if err != nil {
		mus.CacheRemove(id)
		return err
	}

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = gstore.DirtyGet(user.Group).Tag
	user.initPerms()
	_ = mus.CacheSet(user)
	return nil
}

func (mus *MemoryUserStore) Exists(id int) bool {
	return mus.exists.QueryRow(id).Scan(&id) == nil
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
	mus.Unlock()
	atomic.AddInt64(&mus.length, 1)
	return nil
}

func (mus *MemoryUserStore) CacheAddUnsafe(item *User) error {
	if int(mus.length) >= mus.capacity {
		return ErrStoreCapacityOverflow
	}
	mus.items[item.ID] = item
	atomic.AddInt64(&mus.length, 1)
	return nil
}

func (mus *MemoryUserStore) CacheRemove(id int) error {
	mus.Lock()
	delete(mus.items, id)
	mus.Unlock()
	atomic.AddInt64(&mus.length, -1)
	return nil
}

func (mus *MemoryUserStore) CacheRemoveUnsafe(id int) error {
	delete(mus.items, id)
	atomic.AddInt64(&mus.length, -1)
	return nil
}

func (mus *MemoryUserStore) Create(username string, password string, email string, group int, active int) (int, error) {
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

func (mus *MemoryUserStore) GetLength() int {
	return int(mus.length)
}

func (mus *MemoryUserStore) SetCapacity(capacity int) {
	mus.capacity = capacity
}

func (mus *MemoryUserStore) GetCapacity() int {
	return mus.capacity
}

// Return the total number of users registered on the forums
func (mus *MemoryUserStore) GetGlobalCount() int {
	var ucount int
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

func NewSQLUserStore() *SQLUserStore {
	getStmt, err := qgen.Builder.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	existsStmt, err := qgen.Builder.SimpleSelect("users", "uid", "uid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	registerStmt, err := qgen.Builder.SimpleInsert("users", "name, email, password, salt, group, is_super_admin, session, active, message", "?,?,?,?,?,0,'',?,''")
	if err != nil {
		log.Fatal(err)
	}

	usernameExistsStmt, err := qgen.Builder.SimpleSelect("users", "name", "name = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	userCountStmt, err := qgen.Builder.SimpleCount("users", "", "")
	if err != nil {
		log.Fatal(err)
	}

	return &SQLUserStore{
		get:            getStmt,
		exists:         existsStmt,
		register:       registerStmt,
		usernameExists: usernameExistsStmt,
		userCount:      userCountStmt,
	}
}

func (mus *SQLUserStore) Get(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = gstore.DirtyGet(user.Group).Tag
	user.initPerms()
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
		if user.Avatar != "" {
			if user.Avatar[0] == '.' {
				user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
			}
		} else {
			user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
		}
		user.Link = buildProfileURL(nameToSlug(user.Name), user.ID)
		user.Tag = gstore.DirtyGet(user.Group).Tag
		user.initPerms()

		// Add it to the list to be returned
		list[user.ID] = user
	}

	return list, nil
}

func (mus *SQLUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = gstore.DirtyGet(user.Group).Tag
	user.initPerms()
	return user, err
}

func (mus *SQLUserStore) Exists(id int) bool {
	return mus.exists.QueryRow(id).Scan(&id) == nil
}

func (mus *SQLUserStore) Create(username string, password string, email string, group int, active int) (int, error) {
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

// Return the total number of users registered on the forums
func (mus *SQLUserStore) GetGlobalCount() int {
	var ucount int
	err := mus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}
