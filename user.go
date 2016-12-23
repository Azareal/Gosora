package main
import "strings"
import "strconv"
import "net/http"
import "golang.org/x/crypto/bcrypt"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type User struct
{
	ID int
	Name string
	Email string
	Group int
	Active bool
	Is_Mod bool
	Is_Super_Mod bool
	Is_Admin bool
	Is_Super_Admin bool
	Is_Banned bool
	Perms Perms
	Session string
	Loggedin bool
	Avatar string
	Message string
	URLPrefix string
	URLName string
	Tag string
}

func SetPassword(uid int, password string) (error) {
	salt, err := GenerateSafeString(saltLength)
	if err != nil {
		return err
	}
	
	password = password + salt
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	_, err = set_password_stmt.Exec(string(hashed_password), salt, uid)
	if err != nil {
		return err
	}
	return nil
}

func SessionCheck(w http.ResponseWriter, r *http.Request) (user User, noticeList map[int]string, success bool) {
	noticeList = make(map[int]string)
	
	// Are there any session cookies..?
	// Assign it to user.name to avoid having to create a temporary variable for the type conversion
	cookie, err := r.Cookie("uid")
	if err != nil {
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	user.Name = cookie.Value
	user.ID, err = strconv.Atoi(user.Name)
	if err != nil {
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	user.Session = cookie.Value
	
	// Is this session valid..?
	err = get_session_stmt.QueryRow(user.ID,user.Session).Scan(&user.ID, &user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName)
	if err == sql.ErrNoRows {
		user.ID = 0
		user.Session = ""
		user.Perms = GuestPerms
		return user, noticeList, true
	} else if err != nil {
		InternalError(err,w,r,user)
		return user, noticeList, false
	}
	
	user.Is_Admin = user.Is_Super_Admin || groups[user.Group].Is_Admin
	user.Is_Super_Mod = groups[user.Group].Is_Mod || user.Is_Admin
	user.Is_Mod = user.Is_Super_Mod
	user.Is_Banned = groups[user.Group].Is_Banned
	user.Loggedin = !user.Is_Banned || user.Is_Super_Mod
	if user.Is_Banned && user.Is_Super_Mod {
		user.Is_Banned = false
	}
	
	if user.Is_Super_Admin {
		user.Perms = AllPerms
	} else {
		user.Perms = groups[user.Group].Perms
	}
	
	if user.Is_Banned {
		noticeList[0] = "Your account has been suspended. Some of your permissions may have been revoked."
	}
	
	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	return user, noticeList, true
}

func SimpleSessionCheck(w http.ResponseWriter, r *http.Request) (user User, success bool) {
	// Are there any session cookies..?
	// Assign it to user.name to avoid having to create a temporary variable for the type conversion
	cookie, err := r.Cookie("uid")
	if err != nil {
		user.Perms = GuestPerms
		return user, true
	}
	user.Name = cookie.Value
	user.ID, err = strconv.Atoi(user.Name)
	if err != nil {
		user.Perms = GuestPerms
		return user, true
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		user.Perms = GuestPerms
		return user, true
	}
	user.Session = cookie.Value
	
	// Is this session valid..?
	err = get_session_stmt.QueryRow(user.ID,user.Session).Scan(&user.ID, &user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName)
	if err == sql.ErrNoRows {
		user.ID = 0
		user.Session = ""
		user.Perms = GuestPerms
		return user, true
	} else if err != nil {
		InternalError(err,w,r,user)
		return user, false
	}
	
	user.Is_Admin = user.Is_Super_Admin || groups[user.Group].Is_Admin
	user.Is_Super_Mod = groups[user.Group].Is_Mod || user.Is_Admin
	user.Is_Mod = user.Is_Super_Mod
	user.Is_Banned = groups[user.Group].Is_Banned
	user.Loggedin = !user.Is_Banned || user.Is_Super_Mod
	if user.Is_Banned && user.Is_Super_Mod {
		user.Is_Banned = false
	}
	
	if user.Is_Super_Admin {
		user.Perms = AllPerms
	} else {
		user.Perms = groups[user.Group].Perms
	}
	
	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	return user, true
}