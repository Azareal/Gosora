package main
//import "fmt"
import "strings"
import "strconv"
import "net"
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
	Level int
	Score int
	Last_IP string
}

type Email struct
{
	UserID int
	Email string
	Validated bool
	Primary bool
	Token string
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

func SendValidationEmail(username string, email string, token string) bool {
	var schema string
	if enable_ssl {
		schema = "s"
	}
	
	subject := "Validate Your Email @ " + site_name
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\nClick on the following link to do so. http" + schema + "://" + site_url + "/user/edit/token/" + token + "\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	
	return SendEmail(email, subject, msg)
}

func SessionCheck(w http.ResponseWriter, r *http.Request) (user User, noticeList []string, success bool) {
	// Are there any session cookies..?
	cookie, err := r.Cookie("uid")
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	user.ID, err = strconv.Atoi(cookie.Value)
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, noticeList, true
	}
	
	// Is this session valid..?
	err = get_session_stmt.QueryRow(user.ID,cookie.Value).Scan(&user.ID, &user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
	if err == sql.ErrNoRows {
		user.ID = 0
		user.Session = ""
		user.Group = 6
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
		noticeList = append(noticeList, "Your account has been suspended. Some of your permissions may have been revoked.")
	}
	
	if user.Avatar != "" {
		if user.Avatar[0] == '.' {
			user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + user.Avatar
		}
	} else {
		user.Avatar = strings.Replace(noavatar,"{id}",strconv.Itoa(user.ID),1)
	}
	
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return user, noticeList, false
	}
	if host != user.Last_IP {
		go update_last_ip_stmt.Exec(host, user.ID)
	}
	return user, noticeList, true
}

func SimpleSessionCheck(w http.ResponseWriter, r *http.Request) (user User, success bool) {
	// Are there any session cookies..?
	cookie, err := r.Cookie("uid")
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, true
	}
	user.ID, err = strconv.Atoi(cookie.Value)
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, true
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		user.Group = 6
		user.Perms = GuestPerms
		return user, true
	}
	
	// Is this session valid..?
	err = get_session_stmt.QueryRow(user.ID,cookie.Value).Scan(&user.ID, &user.Name, &user.Group, &user.Is_Super_Admin, &user.Session, &user.Email, &user.Avatar, &user.Message, &user.URLPrefix, &user.URLName, &user.Level, &user.Score, &user.Last_IP)
	if err == sql.ErrNoRows {
		user.ID = 0
		user.Session = ""
		user.Group = 6
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
	
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		LocalError("Bad IP",w,r,user)
		return user, false
	}
	if host != user.Last_IP {
		//fmt.Println("Update")
		_, err = update_last_ip_stmt.Exec(host, user.ID)
		if err != nil {
			InternalError(err,w,r,user)
			return user, false
		}
	}
	return user, true
}

func increase_post_user_stats(wcount int, uid int, topic bool, user User) error {
	var mod int
	base_score := 1
	if topic {
		_, err := increment_user_topics_stmt.Exec(1, uid)
		if err != nil {
			return err
		}
		base_score = 2
	}
	
	if wcount > settings["megapost_min_chars"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(1,1,1,uid)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount > settings["bigpost_min_chars"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(1,1,uid)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(1,uid)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(base_score + mod, uid)
	if err != nil {
		return err
	}
	//fmt.Println(user.Score + base_score + mod)
	//fmt.Println(getLevel(user.Score + base_score + mod))
	_, err = update_user_level_stmt.Exec(getLevel(user.Score + base_score + mod), uid)
	return err
}

func decrease_post_user_stats(wcount int, uid int, topic bool, user User) error {
	var mod int
	base_score := -1
	if topic {
		_, err := increment_user_topics_stmt.Exec(-1, uid)
		if err != nil {
			return err
		}
		base_score = -2
	}
	
	if wcount > settings["megapost_min_chars"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(-1,-1,-1,uid)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount > settings["bigpost_min_chars"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(-1,-1,uid)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(-1,uid)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(base_score - mod, uid)
	if err != nil {
		return err
	}
	_, err = update_user_level_stmt.Exec(getLevel(user.Score - base_score - mod), uid)
	return err
}
