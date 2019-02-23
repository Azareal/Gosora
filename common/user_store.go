package common

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/Azareal/Gosora/query_gen"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Add the watchdog goroutine
// TODO: Add some sort of update method
var Users UserStore
var ErrAccountExists = errors.New("this username is already in use")
var ErrLongUsername = errors.New("this username is too long")

type UserStore interface {
	DirtyGet(id int) *User
	Get(id int) (*User, error)
	GetByName(name string) (*User, error)
	Exists(id int) bool
	GetOffset(offset int, perPage int) (users []*User, err error)
	//BulkGet(ids []int) ([]*User, error)
	BulkGetMap(ids []int) (map[int]*User, error)
	BypassGet(id int) (*User, error)
	Create(username string, password string, email string, group int, active bool) (int, error)
	Reload(id int) error
	GlobalCount() int

	SetCache(cache UserCache)
	GetCache() UserCache
}

type DefaultUserStore struct {
	cache UserCache

	get            *sql.Stmt
	getByName      *sql.Stmt
	getOffset      *sql.Stmt
	exists         *sql.Stmt
	register       *sql.Stmt
	usernameExists *sql.Stmt
	userCount      *sql.Stmt
}

// NewDefaultUserStore gives you a new instance of DefaultUserStore
func NewDefaultUserStore(cache UserCache) (*DefaultUserStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullUserCache()
	}
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &DefaultUserStore{
		cache:          cache,
		get:            acc.SimpleSelect("users", "name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, liked, last_ip, temp_group", "uid = ?", "", ""),
		getByName:      acc.Select("users").Columns("uid, name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, liked, last_ip, temp_group").Where("name = ?").Prepare(),
		getOffset:      acc.Select("users").Columns("uid, name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, liked, last_ip, temp_group").Orderby("uid ASC").Limit("?,?").Prepare(),
		exists:         acc.SimpleSelect("users", "uid", "uid = ?", "", ""),
		register:       acc.Insert("users").Columns("name, email, password, salt, group, is_super_admin, session, active, message, createdAt, lastActiveAt, lastLiked, oldestItemLikedCreatedAt").Fields("?,?,?,?,?,0,'',?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(), // TODO: Implement user_count on users_groups here
		usernameExists: acc.SimpleSelect("users", "name", "name = ?", "", ""),
		userCount:      acc.Count("users").Prepare(),
	}, acc.FirstError()
}

func (mus *DefaultUserStore) DirtyGet(id int) *User {
	user, err := mus.cache.Get(id)
	if err == nil {
		return user
	}
	/*if mus.OutOfBounds(id) {
		return BlankUser()
	}*/

	user = &User{ID: id, Loggedin: true}
	err = mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)

	user.Init()
	if err == nil {
		mus.cache.Set(user)
		return user
	}
	return BlankUser()
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
func (mus *DefaultUserStore) Get(id int) (*User, error) {
	user, err := mus.cache.Get(id)
	if err == nil {
		//log.Print("cached user")
		//log.Print(string(debug.Stack()))
		//log.Println("")
		return user, nil
	}
	//log.Print("uncached user")

	user = &User{ID: id, Loggedin: true}
	err = mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)

	user.Init()
	if err == nil {
		mus.cache.Set(user)
	}
	return user, err
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
// ! This bypasses the cache, use frugally
func (mus *DefaultUserStore) GetByName(name string) (*User, error) {
	user := &User{Loggedin: true}
	err := mus.getByName.QueryRow(name).Scan(&user.ID, &user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)

	user.Init()
	if err == nil {
		mus.cache.Set(user)
	}
	return user, err
}

// TODO: Optimise this, so we don't wind up hitting the database every-time for small gaps
// TODO: Make this a little more consistent with DefaultGroupStore's GetRange method
func (store *DefaultUserStore) GetOffset(offset int, perPage int) (users []*User, err error) {
	rows, err := store.getOffset.Query(offset, perPage)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{Loggedin: true}
		err := rows.Scan(&user.ID, &user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)
		if err != nil {
			return nil, err
		}
		user.Init()
		store.cache.Set(user)
		users = append(users, user)
	}
	return users, rows.Err()
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (mus *DefaultUserStore) BulkGetMap(ids []int) (list map[int]*User, err error) {
	var idCount = len(ids)
	list = make(map[int]*User)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := mus.cache.BulkGet(ids)
	if len(sliceList) > 0 {
		for i, sliceItem := range sliceList {
			if sliceItem != nil {
				list[sliceItem.ID] = sliceItem
			} else {
				stillHere = append(stillHere, ids[i])
			}
		}
		ids = stillHere
	}

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	} else if len(ids) == 1 {
		topic, err := mus.Get(ids[0])
		if err != nil {
			return list, err
		}
		list[topic.ID] = topic
		return list, nil
	}

	// TODO: Add a function for the qlist stuff
	var qlist string
	var idList []interface{}
	for _, id := range ids {
		idList = append(idList, strconv.Itoa(id))
		qlist += "?,"
	}
	qlist = qlist[0 : len(qlist)-1]

	rows, err := qgen.NewAcc().Select("users").Columns("uid, name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, liked, last_ip, temp_group").Where("uid IN(" + qlist + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	for rows.Next() {
		user := &User{Loggedin: true}
		err := rows.Scan(&user.ID, &user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)
		if err != nil {
			return list, err
		}
		user.Init()
		mus.cache.Set(user)
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
		if sidList != "" {
			sidList = sidList[0 : len(sidList)-1]
			err = errors.New("Unable to find users with the following IDs: " + sidList)
		}
	}

	return list, err
}

func (mus *DefaultUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)

	user.Init()
	return user, err
}

func (mus *DefaultUserStore) Reload(id int) error {
	user := &User{ID: id, Loggedin: true}
	err := mus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Active, &user.IsSuperAdmin, &user.Session, &user.Email, &user.RawAvatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Liked, &user.LastIP, &user.TempGroup)
	if err != nil {
		mus.cache.Remove(id)
		return err
	}
	user.Init()
	_ = mus.cache.Set(user)
	TopicListThaw.Thaw()
	return nil
}

func (mus *DefaultUserStore) Exists(id int) bool {
	err := mus.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

// TODO: Change active to a bool?
// TODO: Use unique keys for the usernames
func (mus *DefaultUserStore) Create(username string, password string, email string, group int, active bool) (int, error) {
	// TODO: Strip spaces?

	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(username) > Config.MaxUsernameLength {
		return 0, ErrLongUsername
	}

	// Is this username already taken..?
	err := mus.usernameExists.QueryRow(username).Scan(&username)
	if err != ErrNoRows {
		return 0, ErrAccountExists
	}
	salt, err := GenerateSafeString(SaltLength)
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
func (mus *DefaultUserStore) GlobalCount() (ucount int) {
	err := mus.userCount.QueryRow().Scan(&ucount)
	if err != nil {
		LogError(err)
	}
	return ucount
}

func (mus *DefaultUserStore) SetCache(cache UserCache) {
	mus.cache = cache
}

// TODO: We're temporarily doing this so that you can do ucache != nil in getTopicUser. Refactor it.
func (mus *DefaultUserStore) GetCache() UserCache {
	_, ok := mus.cache.(*NullUserCache)
	if ok {
		return nil
	}
	return mus.cache
}
