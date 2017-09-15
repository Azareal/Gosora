/*
*
*	Gosora User File
*	Copyright Azareal 2017 - 2018
*
 */
package main

import (
	//"log"
	//"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var guestUser = User{ID: 0, Link: "#", Group: 6, Perms: GuestPerms}

//func(real_password string, password string, salt string) (err error)
var CheckPassword = BcryptCheckPassword

//func(password string) (hashed_password string, salt string, err error)
var GeneratePassword = BcryptGeneratePassword

type User struct {
	ID           int
	Link         string
	Name         string
	Email        string
	Group        int
	Active       bool
	IsMod        bool
	IsSuperMod   bool
	IsAdmin      bool
	IsSuperAdmin bool
	IsBanned     bool
	Perms        Perms
	PluginPerms  map[string]bool
	Session      string
	Loggedin     bool
	Avatar       string
	Message      string
	URLPrefix    string // Move this to another table? Create a user lite?
	URLName      string
	Tag          string
	Level        int
	Score        int
	LastIP       string
	TempGroup    int
}

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

func (user *User) Ban(duration time.Duration, issuedBy int) error {
	return user.ScheduleGroupUpdate(4, issuedBy, duration)
}

func (user *User) Unban() error {
	err := user.RevertGroupUpdate()
	if err != nil {
		return err
	}
	return users.Reload(user.ID)
}

// TODO: Use a transaction to avoid race conditions
// Make this more stateless?
func (user *User) ScheduleGroupUpdate(gid int, issuedBy int, duration time.Duration) error {
	var temporary bool
	if duration.Nanoseconds() != 0 {
		temporary = true
	}

	revertAt := time.Now().Add(duration)
	_, err := replace_schedule_group_stmt.Exec(user.ID, gid, issuedBy, revertAt, temporary)
	if err != nil {
		return err
	}
	_, err = set_temp_group_stmt.Exec(gid, user.ID)
	if err != nil {
		return err
	}
	return users.Reload(user.ID)
}

// TODO: Use a transaction to avoid race conditions
func (user *User) RevertGroupUpdate() error {
	_, err := replace_schedule_group_stmt.Exec(user.ID, 0, 0, time.Now(), false)
	if err != nil {
		return err
	}
	_, err = set_temp_group_stmt.Exec(0, user.ID)
	if err != nil {
		return err
	}
	return users.Reload(user.ID)
}

func BcryptCheckPassword(realPassword string, password string, salt string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(realPassword), []byte(password+salt))
}

// Investigate. Do we need the extra salt?
func BcryptGeneratePassword(password string) (hashedPassword string, salt string, err error) {
	salt, err = GenerateSafeString(saltLength)
	if err != nil {
		return "", "", err
	}

	password = password + salt
	hashedPassword, err = BcryptGeneratePasswordNoSalt(password)
	if err != nil {
		return "", "", err
	}
	return hashedPassword, salt, nil
}

func BcryptGeneratePasswordNoSalt(password string) (hash string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func SetPassword(uid int, password string) error {
	hashedPassword, salt, err := GeneratePassword(password)
	if err != nil {
		return err
	}
	_, err = set_password_stmt.Exec(hashedPassword, salt, uid)
	return err
}

func SendValidationEmail(username string, email string, token string) bool {
	var schema = "http"
	if site.EnableSsl {
		schema += "s"
	}

	// TODO: Move these to the phrase system
	subject := "Validate Your Email @ " + site.Name
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. " + schema + "://" + site.URL + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

func wordsToScore(wcount int, topic bool) (score int) {
	if topic {
		score = 2
	} else {
		score = 1
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		score += 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		score++
	}
	return score
}

// TODO: Move this to where the other User methods are
func (user *User) increasePostStats(wcount int, topic bool) error {
	var mod int
	baseScore := 1
	if topic {
		_, err := increment_user_topics_stmt.Exec(1, user.ID)
		if err != nil {
			return err
		}
		baseScore = 2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(1, 1, 1, user.ID)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(1, 1, user.ID)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(1, user.ID)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(baseScore+mod, user.ID)
	if err != nil {
		return err
	}
	//log.Print(user.Score + base_score + mod)
	//log.Print(getLevel(user.Score + base_score + mod))
	// TODO: Use a transaction to prevent level desyncs?
	_, err = update_user_level_stmt.Exec(getLevel(user.Score+baseScore+mod), user.ID)
	return err
}

// TODO: Move this to where the other User methods are
func (user *User) decreasePostStats(wcount int, topic bool) error {
	var mod int
	baseScore := -1
	if topic {
		_, err := increment_user_topics_stmt.Exec(-1, user.ID)
		if err != nil {
			return err
		}
		baseScore = -2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err := increment_user_megaposts_stmt.Exec(-1, -1, -1, user.ID)
		if err != nil {
			return err
		}
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err := increment_user_bigposts_stmt.Exec(-1, -1, user.ID)
		if err != nil {
			return err
		}
		mod = 1
	} else {
		_, err := increment_user_posts_stmt.Exec(-1, user.ID)
		if err != nil {
			return err
		}
	}
	_, err := increment_user_score_stmt.Exec(baseScore-mod, user.ID)
	if err != nil {
		return err
	}
	// TODO: Use a transaction to prevent level desyncs?
	_, err = update_user_level_stmt.Exec(getLevel(user.Score-baseScore-mod), user.ID)
	return err
}

func initUserPerms(user *User) {
	if user.TempGroup != 0 {
		user.Group = user.TempGroup
	}

	group := gstore.DirtyGet(user.Group)
	if user.IsSuperAdmin {
		user.Perms = AllPerms
		user.PluginPerms = AllPluginPerms
	} else {
		user.Perms = group.Perms
		user.PluginPerms = group.PluginPerms
	}

	user.IsAdmin = user.IsSuperAdmin || group.IsAdmin
	user.IsSuperMod = user.IsAdmin || group.IsMod
	user.IsMod = user.IsSuperMod
	user.IsBanned = group.IsBanned
	if user.IsBanned && user.IsSuperMod {
		user.IsBanned = false
	}
}

func buildProfileURL(slug string, uid int) string {
	if slug == "" {
		return "/user/" + strconv.Itoa(uid)
	}
	return "/user/" + slug + "." + strconv.Itoa(uid)
}
