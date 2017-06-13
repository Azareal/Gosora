package main

import (
	//"fmt"
	"strconv"
	"net"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var guest_user User = User{ID:0,Group:6,Perms:GuestPerms}

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
	//WS_Conn interface{}
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
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. http" + schema + "://" + site_url + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

func SimpleForumSessionCheck(w http.ResponseWriter, r *http.Request, fid int) (user User, success bool) {
	if !forum_exists(fid) {
		PreError("The target forum doesn't exist.",w,r)
		return user, false
	}
	user, success = SimpleSessionCheck(w,r)
	fperms := groups[user.Group].Forums[fid]
	if fperms.Overrides && !user.Is_Super_Admin {
		user.Perms.ViewTopic = fperms.ViewTopic
		user.Perms.LikeItem = fperms.LikeItem
		user.Perms.CreateTopic = fperms.CreateTopic
		user.Perms.EditTopic = fperms.EditTopic
		user.Perms.DeleteTopic = fperms.DeleteTopic
		user.Perms.CreateReply = fperms.CreateReply
		user.Perms.EditReply = fperms.EditReply
		user.Perms.DeleteReply = fperms.DeleteReply
		user.Perms.PinTopic = fperms.PinTopic
		user.Perms.CloseTopic = fperms.CloseTopic
	}
	return user, success
}

func ForumSessionCheck(w http.ResponseWriter, r *http.Request, fid int) (user User, noticeList []string, success bool) {
	if !forum_exists(fid) {
		NotFound(w,r)
		return user, noticeList, false
	}
	user, success = SimpleSessionCheck(w,r)
	fperms := groups[user.Group].Forums[fid]
	//fmt.Printf("%+v\n", user.Perms)
	//fmt.Printf("%+v\n", fperms)
	if fperms.Overrides && !user.Is_Super_Admin {
		user.Perms.ViewTopic = fperms.ViewTopic
		user.Perms.LikeItem = fperms.LikeItem
		user.Perms.CreateTopic = fperms.CreateTopic
		user.Perms.EditTopic = fperms.EditTopic
		user.Perms.DeleteTopic = fperms.DeleteTopic
		user.Perms.CreateReply = fperms.CreateReply
		user.Perms.EditReply = fperms.EditReply
		user.Perms.DeleteReply = fperms.DeleteReply
		user.Perms.PinTopic = fperms.PinTopic
		user.Perms.CloseTopic = fperms.CloseTopic
	}
	if user.Is_Banned {
		noticeList = append(noticeList,"Your account has been suspended. Some of your permissions may have been revoked.")
	}
	return user, noticeList, success
}

func SessionCheck(w http.ResponseWriter, r *http.Request) (user User, noticeList []string, success bool) {
	user, success = SimpleSessionCheck(w,r)
	if user.Is_Banned {
		noticeList = append(noticeList,"Your account has been suspended. Some of your permissions may have been revoked.")
	}
	return user, noticeList, success
}

func SimpleSessionCheck(w http.ResponseWriter, r *http.Request) (User,bool) {
	// Are there any session cookies..?
	cookie, err := r.Cookie("uid")
	if err != nil {
		return guest_user, true
	}
	uid, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return guest_user, true
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		return guest_user, true
	}

	// Is this session valid..?
	user, err := users.CascadeGet(uid)
	if err == sql.ErrNoRows {
		return guest_user, true
	} else if err != nil {
		InternalError(err,w,r)
		return guest_user, false
	}

	if user.Session == "" || cookie.Value != user.Session {
		return guest_user, true
	}

	if user.Is_Super_Admin {
		user.Perms = AllPerms
	} else {
		user.Perms = groups[user.Group].Perms
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		PreError("Bad IP",w,r)
		return *user, false
	}
	if host != user.Last_IP {
		_, err = update_last_ip_stmt.Exec(host, user.ID)
		if err != nil {
			InternalError(err,w,r)
			return *user, false
		}
	}
	return *user, true
}

func words_to_score(wcount int, topic bool) (score int) {
	if topic {
		score = 2
	} else {
		score = 1
	}

	if wcount >= settings["megapost_min_words"].(int) {
		score += 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		score += 1
	}
	return score
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

	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(1,1,1,uid)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
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

	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(-1,-1,-1,uid)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
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

func init_user_perms(user *User) {
	user.Is_Admin = user.Is_Super_Admin || groups[user.Group].Is_Admin
	user.Is_Super_Mod = user.Is_Admin || groups[user.Group].Is_Mod
	user.Is_Mod = user.Is_Super_Mod
	user.Is_Banned = groups[user.Group].Is_Banned
	if user.Is_Banned && user.Is_Super_Mod {
		user.Is_Banned = false
	}
}

func build_profile_url(uid int) string {
	return "/user/" + strconv.Itoa(uid)
}
