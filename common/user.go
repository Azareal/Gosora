/*
*
*	Gosora User File
*	Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"../query_gen/lib"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Replace any literals with this
var BanGroup = 4

// GuestUser is an instance of user which holds guest data to avoid having to initialise a guest every time
var GuestUser = User{ID: 0, Link: "#", Group: 6, Perms: GuestPerms}

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
	//AuthToken    string
	Loggedin  bool
	Avatar    string
	Message   string
	URLPrefix string // Move this to another table? Create a user lite?
	URLName   string
	Tag       string
	Level     int
	Score     int
	LastIP    string // ! This part of the UserCache data might fall out of date
	TempGroup int
}

type UserStmts struct {
	activate           *sql.Stmt
	changeGroup        *sql.Stmt
	delete             *sql.Stmt
	setAvatar          *sql.Stmt
	setUsername        *sql.Stmt
	updateGroup        *sql.Stmt
	incrementTopics    *sql.Stmt
	updateLevel        *sql.Stmt
	incrementScore     *sql.Stmt
	incrementPosts     *sql.Stmt
	incrementBigposts  *sql.Stmt
	incrementMegaposts *sql.Stmt
	updateLastIP       *sql.Stmt

	setPassword *sql.Stmt
}

var userStmts UserStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		userStmts = UserStmts{
			activate:           acc.SimpleUpdate("users", "active = 1", "uid = ?"),
			changeGroup:        acc.SimpleUpdate("users", "group = ?", "uid = ?"),
			delete:             acc.SimpleDelete("users", "uid = ?"),
			setAvatar:          acc.SimpleUpdate("users", "avatar = ?", "uid = ?"),
			setUsername:        acc.SimpleUpdate("users", "name = ?", "uid = ?"),
			updateGroup:        acc.SimpleUpdate("users", "group = ?", "uid = ?"),
			incrementTopics:    acc.SimpleUpdate("users", "topics =  topics + ?", "uid = ?"),
			updateLevel:        acc.SimpleUpdate("users", "level = ?", "uid = ?"),
			incrementScore:     acc.SimpleUpdate("users", "score = score + ?", "uid = ?"),
			incrementPosts:     acc.SimpleUpdate("users", "posts = posts + ?", "uid = ?"),
			incrementBigposts:  acc.SimpleUpdate("users", "posts = posts + ?, bigposts = bigposts + ?", "uid = ?"),
			incrementMegaposts: acc.SimpleUpdate("users", "posts = posts + ?, bigposts = bigposts + ?, megaposts = megaposts + ?", "uid = ?"),
			updateLastIP:       acc.SimpleUpdate("users", "last_ip = ?", "uid = ?"),

			setPassword: acc.SimpleUpdate("users", "password = ?, salt = ?", "uid = ?"),
		}
		return acc.FirstError()
	})
}

func (user *User) Init() {
	user.Avatar = BuildAvatar(user.ID, user.Avatar)
	user.Link = BuildProfileURL(NameToSlug(user.Name), user.ID)
	user.Tag = Groups.DirtyGet(user.Group).Tag
	user.InitPerms()
}

// TODO: Refactor this idiom into something shorter, maybe with a NullUserCache when one isn't set?
func (user *User) CacheRemove() {
	ucache := Users.GetCache()
	if ucache != nil {
		ucache.Remove(user.ID)
	}
}

func (user *User) Ban(duration time.Duration, issuedBy int) error {
	return user.ScheduleGroupUpdate(BanGroup, issuedBy, duration)
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

	tx, err := qgen.Builder.Begin()
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

	user.CacheRemove()
	return err
}

func (user *User) RevertGroupUpdate() error {
	tx, err := qgen.Builder.Begin()
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

	user.CacheRemove()
	return err
}

// TODO: Use a transaction here
// ? - Add a Deactivate method? Not really needed, if someone's been bad you could do a ban, I guess it might be useful, if someone says that email x isn't actually owned by the user in question?
func (user *User) Activate() (err error) {
	_, err = userStmts.activate.Exec(user.ID)
	if err != nil {
		return err
	}
	_, err = userStmts.changeGroup.Exec(Config.DefaultGroup, user.ID)
	user.CacheRemove()
	return err
}

// TODO: Write tests for this
// TODO: Delete this user's content too?
// TODO: Expose this to the admin?
func (user *User) Delete() error {
	_, err := userStmts.delete.Exec(user.ID)
	if err != nil {
		return err
	}
	user.CacheRemove()
	return err
}

func (user *User) ChangeName(username string) (err error) {
	_, err = userStmts.setUsername.Exec(username, user.ID)
	user.CacheRemove()
	return err
}

func (user *User) ChangeAvatar(avatar string) (err error) {
	_, err = userStmts.setAvatar.Exec(avatar, user.ID)
	user.CacheRemove()
	return err
}

func (user *User) ChangeGroup(group int) (err error) {
	_, err = userStmts.updateGroup.Exec(group, user.ID)
	user.CacheRemove()
	return err
}

// ! Only updates the database not the *User for safety reasons
func (user *User) UpdateIP(host string) error {
	_, err := userStmts.updateLastIP.Exec(host, user.ID)
	return err
}

func (user *User) IncreasePostStats(wcount int, topic bool) (err error) {
	var mod int
	baseScore := 1
	if topic {
		_, err = userStmts.incrementTopics.Exec(1, user.ID)
		if err != nil {
			return err
		}
		baseScore = 2
	}

	settings := SettingBox.Load().(SettingMap)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err = userStmts.incrementMegaposts.Exec(1, 1, 1, user.ID)
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err = userStmts.incrementBigposts.Exec(1, 1, user.ID)
		mod = 1
	} else {
		_, err = userStmts.incrementPosts.Exec(1, user.ID)
	}
	if err != nil {
		return err
	}

	_, err = userStmts.incrementScore.Exec(baseScore+mod, user.ID)
	if err != nil {
		return err
	}
	//log.Print(user.Score + base_score + mod)
	//log.Print(getLevel(user.Score + base_score + mod))
	// TODO: Use a transaction to prevent level desyncs?
	_, err = userStmts.updateLevel.Exec(GetLevel(user.Score+baseScore+mod), user.ID)
	return err
}

func (user *User) DecreasePostStats(wcount int, topic bool) (err error) {
	var mod int
	baseScore := -1
	if topic {
		_, err = userStmts.incrementTopics.Exec(-1, user.ID)
		if err != nil {
			return err
		}
		baseScore = -2
	}

	settings := SettingBox.Load().(SettingMap)
	if wcount >= settings["megapost_min_words"].(int) {
		_, err = userStmts.incrementMegaposts.Exec(-1, -1, -1, user.ID)
		mod = 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		_, err = userStmts.incrementBigposts.Exec(-1, -1, user.ID)
		mod = 1
	} else {
		_, err = userStmts.incrementPosts.Exec(-1, user.ID)
	}
	if err != nil {
		return err
	}

	_, err = userStmts.incrementScore.Exec(baseScore-mod, user.ID)
	if err != nil {
		return err
	}
	// TODO: Use a transaction to prevent level desyncs?
	_, err = userStmts.updateLevel.Exec(GetLevel(user.Score-baseScore-mod), user.ID)
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the user
func (user *User) Copy() User {
	return *user
}

// TODO: Write unit tests for this
func (user *User) InitPerms() {
	if user.TempGroup != 0 {
		user.Group = user.TempGroup
	}

	group := Groups.DirtyGet(user.Group)
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

// ? Make this part of *User?
func BuildAvatar(uid int, avatar string) string {
	if avatar != "" {
		if avatar[0] == '.' {
			return "/uploads/avatar_" + strconv.Itoa(uid) + avatar
		}
		return avatar
	}
	return strings.Replace(Config.Noavatar, "{id}", strconv.Itoa(uid), 1)
}

func BcryptCheckPassword(realPassword string, password string, salt string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(realPassword), []byte(password+salt))
}

// Investigate. Do we need the extra salt?
func BcryptGeneratePassword(password string) (hashedPassword string, salt string, err error) {
	salt, err = GenerateSafeString(SaltLength)
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

// TODO: Move this to *User
func SetPassword(uid int, password string) error {
	hashedPassword, salt, err := GeneratePassword(password)
	if err != nil {
		return err
	}
	_, err = userStmts.setPassword.Exec(hashedPassword, salt, uid)
	return err
}

// TODO: Write units tests for this
func wordsToScore(wcount int, topic bool) (score int) {
	if topic {
		score = 2
	} else {
		score = 1
	}

	settings := SettingBox.Load().(SettingMap)
	if wcount >= settings["megapost_min_words"].(int) {
		score += 4
	} else if wcount >= settings["bigpost_min_words"].(int) {
		score++
	}
	return score
}

// For use in tests and to help generate dummy users for forums which don't have last posters
func BlankUser() *User {
	return &User{ID: 0, Name: ""}
}

// TODO: Write unit tests for this
func BuildProfileURL(slug string, uid int) string {
	if slug == "" {
		return "/user/" + strconv.Itoa(uid)
	}
	return "/user/" + slug + "." + strconv.Itoa(uid)
}
