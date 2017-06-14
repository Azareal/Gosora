package main

import "log"
import "sync"
import "strings"
import "strconv"
import "database/sql"
import "./query_gen/lib"

var users UserStore

type UserStore interface {
	Load(id int) error
	Get(id int) (*User, error)
	GetUnsafe(id int) (*User, error)
	CascadeGet(id int) (*User, error)
	BypassGet(id int) (*User, error)
	Set(item *User) error
	Add(item *User) error
	AddUnsafe(item *User) error
	//SetConn(conn interface{}) error
	//GetConn() interface{}
	Remove(id int) error
	RemoveUnsafe(id int) error
	GetLength() int
	GetCapacity() int
}

type StaticUserStore struct {
	items map[int]*User
	length int
	capacity int
	get *sql.Stmt
	sync.RWMutex
}

func NewStaticUserStore(capacity int) *StaticUserStore {
	stmt, err := qgen.Builder.SimpleSelect("users","name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip","uid = ?","")
	if err != nil {
		log.Fatal(err)
	}
	return &StaticUserStore{
		items:make(map[int]*User),
		capacity:capacity,
		get:stmt,
	}
}

func (sus *StaticUserStore) Get(id int) (*User, error) {
	sus.RLock()
	item, ok := sus.items[id]
	sus.RUnlock()
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sus *StaticUserStore) GetUnsafe(id int) (*User, error) {
	item, ok := sus.items[id]
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sus *StaticUserStore) CascadeGet(id int) (*User, error) {
	sus.RLock()
	user, ok := sus.items[id]
	sus.RUnlock()
	if ok {
		return user, nil
	}

	user = &User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(user)
	if err == nil {
		sus.Set(user)
	}
	return user, err
}

func (sus *StaticUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(user)
	return user, err
}

func (sus *StaticUserStore) Load(id int) error {
	user := &User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
	if err != nil {
		sus.Remove(id)
		return err
	}

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(user)
	sus.Set(user)
	return nil
}

func (sus *StaticUserStore) Set(item *User) error {
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

func (sus *StaticUserStore) Add(item *User) error {
	if sus.length >= sus.capacity {
		return ErrStoreCapacityOverflow
	}
	sus.Lock()
	sus.items[item.ID] = item
	sus.Unlock()
	sus.length++
	return nil
}

func (sus *StaticUserStore) AddUnsafe(item *User) error {
	if sus.length >= sus.capacity {
		return ErrStoreCapacityOverflow
	}
	sus.items[item.ID] = item
	sus.length++
	return nil
}

func (sus *StaticUserStore) Remove(id int) error {
	sus.Lock()
	delete(sus.items,id)
	sus.Unlock()
	sus.length--
	return nil
}

func (sus *StaticUserStore) RemoveUnsafe(id int) error {
	delete(sus.items,id)
	sus.length--
	return nil
}

func (sus *StaticUserStore) GetLength() int {
	return sus.length
}

func (sus *StaticUserStore) SetCapacity(capacity int) {
	sus.capacity = capacity
}

func (sus *StaticUserStore) GetCapacity() int {
	return sus.capacity
}

//type DynamicUserStore struct {
//	items_expiries list.List
//	items map[int]*User
//}

type SqlUserStore struct {
	get *sql.Stmt
}

func NewSqlUserStore() *SqlUserStore {
	stmt, err := qgen.Builder.SimpleSelect("users","name, group, is_super_admin, session, email, avatar, message, url_prefix, url_name, level, score, last_ip","uid = ?","")
	if err != nil {
		log.Fatal(err)
	}
	return &SqlUserStore{stmt}
}

func (sus *SqlUserStore) Get(id int) (*User, error) {
	user := User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(&user)
	return &user, err
}

func (sus *SqlUserStore) GetUnsafe(id int) (*User, error) {
	user := User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(&user)
	return &user, err
}

func (sus *SqlUserStore) CascadeGet(id int) (*User, error) {
	user := User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(&user)
	return &user, err
}

func (sus *SqlUserStore) BypassGet(id int) (*User, error) {
	user := User{ID:id,Loggedin:true}
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	user.Tag = groups[user.Group].Tag
	init_user_perms(&user)
	return &user, err
}

func (sus *SqlUserStore) Load(id int) error {
	user := &User{ID:id}
	// Simplify this into a quick check whether the user exists
	err := sus.get.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
	return err
}

// Placeholder methods, the actual queries are done elsewhere
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

func (sus *SqlUserStore) GetLength() int {
	return 0 // Return the total number of users registered on the forums?
}
