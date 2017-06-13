package main

import "sync"
import "strings"
import "strconv"
import "database/sql"

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
	sync.RWMutex
}

func NewStaticUserStore(capacity int) *StaticUserStore {
	return &StaticUserStore{items:make(map[int]*User),capacity:capacity}
}

func (sts *StaticUserStore) Get(id int) (*User, error) {
	sts.RLock()
	item, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sts *StaticUserStore) GetUnsafe(id int) (*User, error) {
	item, ok := sts.items[id]
	if ok {
		return item, nil
	}
	return item, sql.ErrNoRows
}

func (sts *StaticUserStore) CascadeGet(id int) (*User, error) {
	sts.RLock()
	user, ok := sts.items[id]
	sts.RUnlock()
	if ok {
		return user, nil
	}

	user = &User{ID:id,Loggedin:true}
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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
		sts.Set(user)
	}
	return user, err
}

func (sts *StaticUserStore) BypassGet(id int) (*User, error) {
	user := &User{ID:id,Loggedin:true}
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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

func (sts *StaticUserStore) Load(id int) error {
	user := &User{ID:id,Loggedin:true}
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
	if err != nil {
		sts.Remove(id)
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
	sts.Set(user)
	return nil
}

func (sts *StaticUserStore) Set(item *User) error {
	sts.Lock()
	user, ok := sts.items[item.ID]
	if ok {
		sts.Unlock()
		*user = *item
	} else if sts.length >= sts.capacity {
		sts.Unlock()
		return ErrStoreCapacityOverflow
	} else {
		sts.items[item.ID] = item
		sts.Unlock()
		sts.length++
	}
	return nil
}

func (sts *StaticUserStore) Add(item *User) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.Lock()
	sts.items[item.ID] = item
	sts.Unlock()
	sts.length++
	return nil
}

func (sts *StaticUserStore) AddUnsafe(item *User) error {
	if sts.length >= sts.capacity {
		return ErrStoreCapacityOverflow
	}
	sts.items[item.ID] = item
	sts.length++
	return nil
}

func (sts *StaticUserStore) Remove(id int) error {
	sts.Lock()
	delete(sts.items,id)
	sts.Unlock()
	sts.length--
	return nil
}

func (sts *StaticUserStore) RemoveUnsafe(id int) error {
	delete(sts.items,id)
	sts.length--
	return nil
}

func (sts *StaticUserStore) GetLength() int {
	return sts.length
}

func (sts *StaticUserStore) SetCapacity(capacity int) {
	sts.capacity = capacity
}

func (sts *StaticUserStore) GetCapacity() int {
	return sts.capacity
}

//type DynamicUserStore struct {
//	items_expiries list.List
//	items map[int]*User
//}

type SqlUserStore struct {
}

func NewSqlUserStore() *SqlUserStore {
	return &SqlUserStore{}
}

func (sus *SqlUserStore) Get(id int) (*User, error) {
	user := User{ID:id,Loggedin:true}
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)

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
	err := get_full_user_stmt.QueryRow(id).Scan(&user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
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
