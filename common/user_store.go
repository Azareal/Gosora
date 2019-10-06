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
	Count() int

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
	count      *sql.Stmt
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
		get:            acc.SimpleSelect("users", "name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, posts, liked, last_ip, temp_group", "uid = ?", "", ""),
		getByName:      acc.Select("users").Columns("uid, name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, posts, liked, last_ip, temp_group").Where("name = ?").Prepare(),
		getOffset:      acc.Select("users").Columns("uid, name, group, active, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, posts, liked, last_ip, temp_group").Orderby("uid ASC").Limit("?,?").Prepare(),
		exists:         acc.SimpleSelect("users", "uid", "uid = ?", "", ""),
		register:       acc.Insert("users").Columns("name, email, password, salt, group, is_super_admin, session, active, message, createdAt, lastActiveAt, lastLiked, oldestItemLikedCreatedAt").Fields("?,?,?,?,?,0,'',?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(), // TODO: Implement user_count on users_groups here
		usernameExists: acc.SimpleSelect("users", "name", "name = ?", "", ""),
		count:      acc.Count("users").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultUserStore) DirtyGet(id int) *User {
	user, err := s.Get(id)
	if err == nil {
		return user
	}
	/*if s.OutOfBounds(id) {
		return BlankUser()
	}*/
	return BlankUser()
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
func (s *DefaultUserStore) Get(id int) (*User, error) {
	u, err := s.cache.Get(id)
	if err == nil {
		//log.Print("cached user")
		//log.Print(string(debug.Stack()))
		//log.Println("")
		return u, nil
	}
	//log.Print("uncached user")

	u = &User{ID: id, Loggedin: true}
	err = s.get.QueryRow(id).Scan(&u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.URLPrefix, &u.URLName, &u.Level, &u.Score, &u.Posts,&u.Liked, &u.LastIP, &u.TempGroup)
	if err == nil {
		u.Init()
		s.cache.Set(u)
	}
	return u, err
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
// ! This bypasses the cache, use frugally
func (s *DefaultUserStore) GetByName(name string) (*User, error) {
	u := &User{Loggedin: true}
	err := s.getByName.QueryRow(name).Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.URLPrefix, &u.URLName, &u.Level, &u.Score, &u.Posts,&u.Liked, &u.LastIP, &u.TempGroup)
	if err == nil {
		u.Init()
		s.cache.Set(u)
	}
	return u, err
}

// TODO: Optimise this, so we don't wind up hitting the database every-time for small gaps
// TODO: Make this a little more consistent with DefaultGroupStore's GetRange method
func (s *DefaultUserStore) GetOffset(offset int, perPage int) (users []*User, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.URLPrefix, &u.URLName, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup)
		if err != nil {
			return nil, err
		}
		u.Init()
		s.cache.Set(u)
		users = append(users, u)
	}
	return users, rows.Err()
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (s *DefaultUserStore) BulkGetMap(ids []int) (list map[int]*User, err error) {
	idCount := len(ids)
	list = make(map[int]*User)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := s.cache.BulkGet(ids)
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
		user, err := s.Get(ids[0])
		if err != nil {
			return list, err
		}
		list[user.ID] = user
		return list, nil
	}

	// TODO: Add a function for the q stuff
	var q string
	idList := make([]interface{},len(ids))
	for i, id := range ids {
		idList[i] = strconv.Itoa(id)
		q += "?,"
	}
	q = q[0 : len(q)-1]

	rows, err := qgen.NewAcc().Select("users").Columns("uid,name,group,active,is_super_admin,session,email,avatar,message,url_prefix,url_name,level,score,posts,liked,last_ip,temp_group").Where("uid IN(" + q + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.URLPrefix, &u.URLName, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup)
		if err != nil {
			return list, err
		}
		u.Init()
		s.cache.Set(u)
		list[u.ID] = u
	}
	err = rows.Err()
	if err != nil {
		return list, err
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

func (s *DefaultUserStore) BypassGet(id int) (*User, error) {
	u := &User{ID: id, Loggedin: true}
	err := s.get.QueryRow(id).Scan(&u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.URLPrefix, &u.URLName, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup)
	if err == nil {
		u.Init()
	}
	return u, err
}

func (s *DefaultUserStore) Reload(id int) error {
	u, err := s.BypassGet(id)
	if err != nil {
		s.cache.Remove(id)
		return err
	}
	_ = s.cache.Set(u)
	TopicListThaw.Thaw()
	return nil
}

func (s *DefaultUserStore) Exists(id int) bool {
	err := s.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

// TODO: Change active to a bool?
// TODO: Use unique keys for the usernames
func (s *DefaultUserStore) Create(username string, password string, email string, group int, active bool) (int, error) {
	// TODO: Strip spaces?

	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(username) > Config.MaxUsernameLength {
		return 0, ErrLongUsername
	}

	// Is this username already taken..?
	err := s.usernameExists.QueryRow(username).Scan(&username)
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

	res, err := s.register.Exec(username, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}

// Count returns the total number of users registered on the forums
func (s *DefaultUserStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultUserStore) SetCache(cache UserCache) {
	s.cache = cache
}

// TODO: We're temporarily doing this so that you can do ucache != nil in getTopicUser. Refactor it.
func (s *DefaultUserStore) GetCache() UserCache {
	_, ok := s.cache.(*NullUserCache)
	if ok {
		return nil
	}
	return s.cache
}
