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
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"./query_gen/lib"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Replace any literals with this
var banGroup = 4

var guestUser = User{ID: 0, Link: "#", Group: 6, Perms: GuestPerms}

//func(real_password string, password string, salt string) (err error)
var CheckPassword = BcryptCheckPassword

//func(password string) (hashed_password string, salt string, err error)
var GeneratePassword = BcryptGeneratePassword
var ErrNoTempGroup = errors.New("We couldn't find a temporary group for this user")

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
	LastIP       string // ! This part of the UserCache data might fall out of date
	TempGroup    int
}

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

func (user *User) Init() {
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
}

func (user *User) Ban(duration time.Duration, issuedBy int) error {
	return user.ScheduleGroupUpdate(banGroup, issuedBy, duration)
}

func (user *User) Unban() error {
	return user.RevertGroupUpdate()
}

func (user *User) deleteScheduleGroupTx(tx *sql.Tx) error {
	deleteScheduleGroupStmt, err := qgen.Builder.SimpleDeleteTx(tx, "users_groups_scheduler", "uid = ?")
	if err != nil {
		return err
	}
	_, err = deleteScheduleGroupStmt.Exec(user.ID)
	return err
}

func (user *User) setTempGroupTx(tx *sql.Tx, tempGroup int) error {
	setTempGroupStmt, err := qgen.Builder.SimpleUpdateTx(tx, "users", "temp_group = ?", "uid = ?")
	if err != nil {
		return err
	}
	_, err = setTempGroupStmt.Exec(tempGroup, user.ID)
	return err
}

// Make this more stateless?
func (user *User) ScheduleGroupUpdate(gid int, issuedBy int, duration time.Duration) error {
	var temporary bool
	if duration.Nanoseconds() != 0 {
		temporary = true
	}
	revertAt := time.Now().Add(duration)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = user.deleteScheduleGroupTx(tx)
	if err != nil {
		return err
	}

	createScheduleGroupTx, err := qgen.Builder.SimpleInsertTx(tx, "users_groups_scheduler", "uid, set_group, issued_by, issued_at, revert_at, temporary", "?,?,?,UTC_TIMESTAMP(),?,?")
	if err != nil {
		return err
	}
	_, err = createScheduleGroupTx.Exec(user.ID, gid, issuedBy, revertAt, temporary)
	if err != nil {
		return err
	}

	err = user.setTempGroupTx(tx, gid)
	if err != nil {
		return err
	}
	err = tx.Commit()

	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

func (user *User) RevertGroupUpdate() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = user.deleteScheduleGroupTx(tx)
	if err != nil {
		return err
	}

	err = user.setTempGroupTx(tx, 0)
	if err != nil {
		return err
	}
	err = tx.Commit()

	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

// TODO: Use a transaction here
// ? - Add a Deactivate method? Not really needed, if someone's been bad you could do a ban, I guess it might be useful, if someone says that email x isn't actually owned by the user in question?
func (user *User) Activate() (err error) {
	_, err = stmts.activateUser.Exec(user.ID)
	if err != nil {
		return err
	}
	_, err = stmts.changeGroup.Exec(config.DefaultGroup, user.ID)
	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

// TODO: Write tests for this
// TODO: Delete this user's content too?
// TODO: Expose this to the admin?
func (user *User) Delete() error {
	_, err := stmts.deleteUser.Exec(user.ID)
	if err != nil {
		return err
	}
	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

func (user *User) ChangeName(username string) (err error) {
	_, err = stmts.setUsername.Exec(username, user.ID)
	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

func (user *User) ChangeAvatar(avatar string) (err error) {
	_, err = stmts.setAvatar.Exec(avatar, user.ID)
	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

func (user *User) ChangeGroup(group int) (err error) {
	_, err = stmts.updateUserGroup.Exec(group, user.ID)
	ucache, ok := users.(UserCache)
	if ok {
		ucache.CacheRemove(user.ID)
	}
	return err
}

func (user *User) increasePostStats(wcount int, topic bool) (err error) {
	var mod int
	baseScore := 1
	if topic {
		_, err = stmts.incrementUserTopics.Exec(1, user.ID)
		if err != nil {
			return err
		}
		baseScore = 2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err = stmts.incrementUserMegaposts.Exec(1, 1, 1, user.ID)
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err = stmts.incrementUserBigposts.Exec(1, 1, user.ID)
		mod = 1
	} else {
		_, err = stmts.incrementUserPosts.Exec(1, user.ID)
	}
	if err != nil {
		return err
	}

	_, err = stmts.incrementUserScore.Exec(baseScore+mod, user.ID)
	if err != nil {
		return err
	}
	//log.Print(user.Score + base_score + mod)
	//log.Print(getLevel(user.Score + base_score + mod))
	// TODO: Use a transaction to prevent level desyncs?
	_, err = stmts.updateUserLevel.Exec(getLevel(user.Score+baseScore+mod), user.ID)
	return err
}

func (user *User) decreasePostStats(wcount int, topic bool) (err error) {
	var mod int
	baseScore := -1
	if topic {
		_, err = stmts.incrementUserTopics.Exec(-1, user.ID)
		if err != nil {
			return err
		}
		baseScore = -2
	}

	settings := settingBox.Load().(SettingBox)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err = stmts.incrementUserMegaposts.Exec(-1, -1, -1, user.ID)
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err = stmts.incrementUserBigposts.Exec(-1, -1, user.ID)
		mod = 1
	} else {
		_, err = stmts.incrementUserPosts.Exec(-1, user.ID)
	}
	if err != nil {
		return err
	}

	_, err = stmts.incrementUserScore.Exec(baseScore-mod, user.ID)
	if err != nil {
		return err
	}
	// TODO: Use a transaction to prevent level desyncs?
	_, err = stmts.updateUserLevel.Exec(getLevel(user.Score-baseScore-mod), user.ID)
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the user
func (user *User) Copy() User {
	return *user
}

// TODO: Write unit tests for this
func (user *User) initPerms() {
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
	_, err = stmts.setPassword.Exec(hashedPassword, salt, uid)
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

// TODO: Write units tests for this
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

// For use in tests and to help generate dummy users for forums which don't have last posters
func getDummyUser() *User {
	return &User{ID: 0, Name: ""}
}

// TODO: Write unit tests for this
func buildProfileURL(slug string, uid int) string {
	if slug == "" {
		return "/user/" + strconv.Itoa(uid)
	}
	return "/user/" + slug + "." + strconv.Itoa(uid)
}
