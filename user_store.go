package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"./query_gen/lib"
	"golang.org/x/crypto/bcrypt"
)

// TO-DO: Add the watchdog goroutine
var users UserStore
var errAccountExists = errors.New("this username is already in use")

type UserStore interface {
	Load(id int) error
	Get(id int) (*User, error)
	GetUnsafe(id int) (*User, error)
	CascadeGet(id int) (*User, error)
	//BulkCascadeGet(ids []int) ([]*User, error)
	BulkCascadeGetMap(ids []int) (map[int]*User, error)
	BypassGet(id int) (*User, error)
	Set(item *User) error
	Add(item *User) error
	AddUnsafe(item *User) error
	Remove(id int) error
	RemoveUnsafe(id int) error
	CreateUser(username string, password string, email string, group int, active int) (int, error)
	GetLength() int
	GetCapacity() int
	GetGlobalCount() int
}

type MemoryUserStore struct {
	items           map[int]*User
	length          int
	capacity        int
	get             *sql.Stmt
	register        *sql.Stmt
	username_exists *sql.Stmt
	user_count      *sql.Stmt
	sync.RWMutex
}

func NewMemoryUserStore(capacity int) *MemoryUserStore {
	get_stmt, err := qgen.Builder.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Add an admin version of register_stmt with more flexibility?
	// create_account_stmt, err = db.Prepare("INSERT INTO
	register_stmt, err := qgen.Builder.SimpleInsert("users", "name, email, password, salt, group, is_super_admin, session, active, message", "?,?,?,?,?,0,'',?,''")
	if err != nil {
		log.Fatal(err)
	}

	username_exists_stmt, err := qgen.Builder.SimpleSelect("users", "name", "name = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}

	user_count_stmt, err := qgen.Builder.SimpleCount("users", "", "")
	if err != nil {
		log.Fatal(err)
	}

	return &MemoryUserStore{
		items:           make(map[int]*User),
		capacity:        capacity,
		get:             get_stmt,
		register:        register_stmt,
		username_exists: username_exists_stmt,
		user_count:      user_count_stmt,
	}
}

func (sus *MemoryUserStore) Get(id int) (*User, error) {
	sus.RLock()
	item, ok := sus.items[id]
	sus.RUnlock()
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sus *MemoryUserStore) GetUnsafe(id int) (*User, error) {
	item, ok := sus.items[id]
	if ok {
		return item, nil
	}
	return item, ErrNoRows
}

func (sus *MemoryUserStore) CascadeGet(id int) (*User, error) {
	sus.RLock()
	user, ok := sus.items[id]
	sus.RUnlock()
	if ok {
		return user, nil
	}

	user = &User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(user)
	if err == nil {
		sus.Set(user)
	}
	return user, err
}

// WARNING: We did a little hack to make this as thin and quick as possible to reduce lock contention, use the * Cascade* methods instead for normal use
func (sus *MemoryUserStore) bulkGet(ids []int) (list []*User) {
	list = make([]*User, len(ids))
	sus.RLock()
	for i, id := range ids {
		list[i] = sus.items[id]
	}
	sus.RUnlock()
	return list
}

// TO-DO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TO-DO: ID of 0 should always error?
func (sus *MemoryUserStore) BulkCascadeGetMap(ids []int) (list map[int]*User, err error) {
	var idCount int = len(ids)
	list = make(map[int]*User)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := sus.bulkGet(ids)
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
		user.Tag = groups[user.Group].Tag
		initUserPerms(user)

		// Add it to the cache...
		_ = sus.Set(user)

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

func (sus *MemoryUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(user)
	return user, err
}

func (sus *MemoryUserStore) Load(id int) error {
	user := &User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
	if err != nil {
		sus.Remove(id)
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
	user.Tag = groups[user.Group].Tag
	initUserPerms(user)
	_ = sus.Set(user)
	return nil
}

func (sus *MemoryUserStore) Set(item *User) error {
	sus.Lock()
	user, ok := sus.items[item.ID]
	if ok {
		sus.Unlock()
		*user = *item
	} else if sus.length >= sus.capacity {
		sus.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		sus.items[item.ID] = item
		sus.Unlock()
		sus.length++
	}
	return nil
}

func (sus *MemoryUserStore) Add(item *User) error {
	if sus.length >= sus.capacity {
		return ErrStoreCapacityOverflow
	}
	sus.Lock()
	sus.items[item.ID] = item
	sus.Unlock()
	sus.length++
	return nil
}

func (sus *MemoryUserStore) AddUnsafe(item *User) error {
	if sus.length >= sus.capacity {
		return ErrStoreCapacityOverflow
	}
	sus.items[item.ID] = item
	sus.length++
	return nil
}

func (sus *MemoryUserStore) Remove(id int) error {
	sus.Lock()
	delete(sus.items, id)
	sus.Unlock()
	sus.length--
	return nil
}

func (sus *MemoryUserStore) RemoveUnsafe(id int) error {
	delete(sus.items, id)
	sus.length--
	return nil
}

func (sus *MemoryUserStore) CreateUser(username string, password string, email string, group int, active int) (int, error) {
	// Is this username already taken..?
	err := sus.username_exists.QueryRow(username).Scan(&username)
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

	res, err := sus.register.Exec(username, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (sus *MemoryUserStore) GetLength() int {
	return sus.length
}

func (sus *MemoryUserStore) SetCapacity(capacity int) {
	sus.capacity = capacity
}

func (sus *MemoryUserStore) GetCapacity() int {
	return sus.capacity
}

// Return the total number of users registered on the forums
func (sus *MemoryUserStore) GetGlobalCount() int {
	var ucount int
	err := sus.user_count.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}

type SqlUserStore struct {
	get            *sql.Stmt
	register       *sql.Stmt
	usernameExists *sql.Stmt
	userCount      *sql.Stmt
}

func NewSqlUserStore() *SqlUserStore {
	getStmt, err := qgen.Builder.SimpleSelect("users", "name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip, temp_group", "uid = ?", "", "")
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

	return &SqlUserStore{
		get:            getStmt,
		register:       registerStmt,
		usernameExists: usernameExistsStmt,
		userCount:      userCountStmt,
	}
}

func (sus *SqlUserStore) Get(id int) (*User, error) {
	user := User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(&user)
	return &user, err
}

func (sus *SqlUserStore) GetUnsafe(id int) (*User, error) {
	user := User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(&user)
	return &user, err
}

func (sus *SqlUserStore) CascadeGet(id int) (*User, error) {
	user := User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(&user)
	return &user, err
}

// TO-DO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
func (sus *SqlUserStore) BulkCascadeGetMap(ids []int) (list map[int]*User, err error) {
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
		user.Tag = groups[user.Group].Tag
		initUserPerms(user)

		// Add it to the list to be returned
		list[user.ID] = user
	}

	return list, nil
}

func (sus *SqlUserStore) BypassGet(id int) (*User, error) {
	user := User{ID: id, Loggedin: true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(config.Noavatar, "{id}", strconv.Itoa(user.ID), 1)
	}
	user.Link = buildProfileURL(nameToSlug(user.Name), id)
	user.Tag = groups[user.Group].Tag
	initUserPerms(&user)
	return &user, err
}

func (sus *SqlUserStore) Load(id int) error {
	user := &User{ID: id}
	// Simplify this into a quick check to see whether the user exists. Add an Exists method to facilitate this?
	return sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.IsSuperAdmin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.LastIP, &user.TempGroup)
}

func (sus *SqlUserStore) CreateUser(username string, password string, email string, group int, active int) (int, error) {
	// Is this username already taken..?
	err := sus.usernameExists.QueryRow(username).Scan(&username)
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

	res, err := sus.register.Exec(username, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

// Placeholder methods, as we're not don't need to do any cache management with this implementation ofr the UserStore
func (sus *SqlUserStore) Set(item *User) error {
	return nil
}
func (sus *SqlUserStore) Add(item *User) error {
	return nil
}
func (sus *SqlUserStore) AddUnsafe(item *User) error {
	return nil
}
func (sus *SqlUserStore) Remove(id int) error {
	return nil
}
func (sus *SqlUserStore) RemoveUnsafe(id int) error {
	return nil
}
func (sus *SqlUserStore) GetCapacity() int {
	return 0
}

// Return the total number of users registered on the forums
func (sus *SqlUserStore) GetLength() int {
	var ucount int
	err := sus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}
func (sus *SqlUserStore) GetGlobalCount() int {
	var ucount int
	err := sus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}
