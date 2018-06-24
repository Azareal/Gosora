/*
*
*	Gosora User File
*	Copyright Azareal 2017 - 2019
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
)

// TODO: Replace any literals with this
var BanGroup = 4

// TODO: Use something else as the guest avatar, maybe a question mark of some sort?
// GuestUser is an instance of user which holds guest data to avoid having to initialise a guest every time
var GuestUser = User{ID: 0, Name: "Guest", Link: "#", Group: 6, Perms: GuestPerms} // BuildAvatar is done in site.go to make sure it's done after init
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
	Liked     int
	LastIP    string // ! This part of the UserCache data might fall out of date
	TempGroup int
}

func (user *User) WebSockets() *WsJSONUser {
	var groupID = user.Group
	if user.TempGroup != 0 {
		groupID = user.TempGroup
	}
	// TODO: Do we want to leak the user's permissions? Users will probably be able to see their status from the group tags, but still
	return &WsJSONUser{user.ID, user.Link, user.Name, groupID, user.IsMod, user.Avatar, user.Level, user.Score, user.Liked}
}

// Use struct tags to avoid having to define this? It really depends on the circumstances, sometimes we want the whole thing, sometimes... not.
type WsJSONUser struct {
	ID     int
	Link   string
	Name   string
	Group  int // Be sure to mask with TempGroup
	IsMod  bool
	Avatar string
	Level  int
	Score  int
	Liked  int
}

type UserStmts struct {
	activate        *sql.Stmt
	changeGroup     *sql.Stmt
	delete          *sql.Stmt
	setAvatar       *sql.Stmt
	setUsername     *sql.Stmt
	incrementTopics *sql.Stmt
	updateLevel     *sql.Stmt
	update          *sql.Stmt

	// TODO: Split these into a sub-struct
	incrementScore     *sql.Stmt
	incrementPosts     *sql.Stmt
	incrementBigposts  *sql.Stmt
	incrementMegaposts *sql.Stmt
	incrementLiked     *sql.Stmt

	decrementLiked *sql.Stmt
	updateLastIP   *sql.Stmt

	setPassword *sql.Stmt
}

var userStmts UserStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		var where = "uid = ?"
		userStmts = UserStmts{
			activate:        acc.SimpleUpdate("users", "active = 1", where),
			changeGroup:     acc.SimpleUpdate("users", "group = ?", where), // TODO: Implement user_count for users_groups here
			delete:          acc.SimpleDelete("users", where),
			setAvatar:       acc.Update("users").Set("avatar = ?").Where(where).Prepare(),
			setUsername:     acc.Update("users").Set("name = ?").Where(where).Prepare(),
			incrementTopics: acc.SimpleUpdate("users", "topics =  topics + ?", where),
			updateLevel:     acc.SimpleUpdate("users", "level = ?", where),
			update:          acc.Update("users").Set("name = ?, email = ?, group = ?").Where("uid = ?").Prepare(), // TODO: Implement user_count for users_groups on things which use this

			incrementScore:     acc.SimpleUpdate("users", "score = score + ?", where),
			incrementPosts:     acc.SimpleUpdate("users", "posts = posts + ?", where),
			incrementBigposts:  acc.SimpleUpdate("users", "posts = posts + ?, bigposts = bigposts + ?", where),
			incrementMegaposts: acc.SimpleUpdate("users", "posts = posts + ?, bigposts = bigposts + ?, megaposts = megaposts + ?", where),
			incrementLiked:     acc.SimpleUpdate("users", "liked = liked + ?, lastLiked = UTC_TIMESTAMP()", where),
			decrementLiked:     acc.SimpleUpdate("users", "liked = liked - ?", where),
			//recalcLastLiked: acc...
			updateLastIP: acc.SimpleUpdate("users", "last_ip = ?", where),

			setPassword: acc.Update("users").Set("password = ?, salt = ?").Where(where).Prepare(),
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

func (user *User) bindStmt(stmt *sql.Stmt, params ...interface{}) (err error) {
	params = append(params, user.ID)
	_, err = stmt.Exec(params...)
	user.CacheRemove()
	return err
}

func (user *User) ChangeName(username string) (err error) {
	return user.bindStmt(userStmts.setUsername, username)
}

func (user *User) ChangeAvatar(avatar string) (err error) {
	return user.bindStmt(userStmts.setAvatar, avatar)
}

func (user *User) ChangeGroup(group int) (err error) {
	return user.bindStmt(userStmts.changeGroup, group)
}

// ! Only updates the database not the *User for safety reasons
func (user *User) UpdateIP(host string) error {
	_, err := userStmts.updateLastIP.Exec(host, user.ID)
	return err
}

func (user *User) Update(newname string, newemail string, newgroup int) (err error) {
	return user.bindStmt(userStmts.update, newname, newemail, newgroup)
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
	return new(User)
}

// TODO: Write unit tests for this
func BuildProfileURL(slug string, uid int) string {
	if slug == "" || !Config.BuildSlugs {
		return "/user/" + strconv.Itoa(uid)
	}
	return "/user/" + slug + "." + strconv.Itoa(uid)
}
