/*
*
*	Gosora User File
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	//"log"

	qgen "github.com/Azareal/Gosora/query_gen"
	"github.com/go-sql-driver/mysql"
)

// TODO: Replace any literals with this
var BanGroup = 4

// TODO: Use something else as the guest avatar, maybe a question mark of some sort?
// GuestUser is an instance of user which holds guest data to avoid having to initialise a guest every time
var GuestUser = User{ID: 0, Name: "Guest", Link: "#", Group: 6, Perms: GuestPerms, CreatedAt: StartTime} // BuildAvatar is done in site.go to make sure it's done after init
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
	Loggedin    bool
	RawAvatar   string
	Avatar      string
	MicroAvatar string
	Message     string
	// TODO: Implement something like this for profiles?
	//URLPrefix   string // Move this to another table? Create a user lite?
	//URLName     string
	Tag       string
	Level     int
	Score     int
	Posts     int
	Liked     int
	CreatedAt time.Time
	LastIP    string // ! This part of the UserCache data might fall out of date
	LastAgent int    // ! Temporary hack for http push, don't use
	TempGroup int

	ParseSettings *ParseSettings
	Privacy       UserPrivacy
}

type UserPrivacy struct {
	ShowComments int // 0 = default, 1 = public, 2 = registered, 3 = friends, 4 = self, 5 = disabled / unused
	AllowMessage int // 0 = default, 1 = registered, 2 = friends, 3 = mods, 4 = disabled / unused
}

func (u *User) WebSockets() *WsJSONUser {
	groupID := u.Group
	if u.TempGroup != 0 {
		groupID = u.TempGroup
	}
	// TODO: Do we want to leak the user's permissions? Users will probably be able to see their status from the group tags, but still
	return &WsJSONUser{u.ID, u.Link, u.Name, groupID, u.IsMod, u.Avatar, u.MicroAvatar, u.Level, u.Score, u.Liked}
}

// Use struct tags to avoid having to define this? It really depends on the circumstances, sometimes we want the whole thing, sometimes... not.
type WsJSONUser struct {
	ID          int
	Link        string
	Name        string
	Group       int // Be sure to mask with TempGroup
	IsMod       bool
	Avatar      string
	MicroAvatar string
	Level       int
	Score       int
	Liked       int
}

func (u *User) Me() *MeUser {
	groupID := u.Group
	if u.TempGroup != 0 {
		groupID = u.TempGroup
	}
	return &MeUser{u.ID, u.Link, u.Name, groupID, u.Active, u.IsMod, u.IsSuperMod, u.IsAdmin, u.IsBanned, u.Session, u.Avatar, u.MicroAvatar, u.Tag, u.Level, u.Score, u.Liked}
}

// For when users need to see their own data, I've omitted some redundancies and less useful items, so we don't wind up sending them on every request
type MeUser struct {
	ID         int
	Link       string
	Name       string
	Group      int
	Active     bool
	IsMod      bool
	IsSuperMod bool
	IsAdmin    bool
	IsBanned   bool

	// TODO: Implement these as copies (might already be the case for Perms, but we'll want to look at it's definition anyway)
	//Perms       Perms
	//PluginPerms map[string]bool

	S           string // Session
	Avatar      string
	MicroAvatar string
	Tag         string
	Level       int
	Score       int
	Liked       int
}

type UserStmts struct {
	activate    *sql.Stmt
	changeGroup *sql.Stmt
	delete      *sql.Stmt
	setAvatar   *sql.Stmt
	setName     *sql.Stmt
	update      *sql.Stmt

	// TODO: Split these into a sub-struct
	incScore         *sql.Stmt
	incPosts         *sql.Stmt
	incBigposts      *sql.Stmt
	incMegaposts     *sql.Stmt
	incPostStats     *sql.Stmt
	incBigpostStats  *sql.Stmt
	incMegapostStats *sql.Stmt
	incLiked         *sql.Stmt
	incTopics        *sql.Stmt
	updateLevel      *sql.Stmt
	resetStats       *sql.Stmt
	setStats         *sql.Stmt

	decLiked      *sql.Stmt
	updateLastIP  *sql.Stmt
	updatePrivacy *sql.Stmt

	setPassword *sql.Stmt

	scheduleAvatarResize *sql.Stmt

	deletePosts            *sql.Stmt
	deleteProfilePosts     *sql.Stmt
	deleteReplyPosts       *sql.Stmt
	getLikedRepliesOfTopic *sql.Stmt
	getAttachmentsOfTopic  *sql.Stmt
	getAttachmentsOfTopic2 *sql.Stmt
	getRepliesOfTopic      *sql.Stmt
}

var userStmts UserStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		u := "users"
		w := "uid=?"
		userStmts = UserStmts{
			activate:    acc.SimpleUpdate(u, "active=1", w),
			changeGroup: acc.SimpleUpdate(u, "group=?", w), // TODO: Implement user_count for users_groups here
			delete:      acc.Delete(u).Where(w).Prepare(),
			setAvatar:   acc.Update(u).Set("avatar=?").Where(w).Prepare(),
			setName:     acc.Update(u).Set("name=?").Where(w).Prepare(),
			update:      acc.Update(u).Set("name=?,email=?,group=?").Where(w).Prepare(), // TODO: Implement user_count for users_groups on things which use this

			// Stat Statements
			// TODO: Do +0 to avoid having as many statements?
			incScore:         acc.Update(u).Set("score=score+?").Where(w).Prepare(),
			incPosts:         acc.Update(u).Set("posts=posts+?").Where(w).Prepare(),
			incBigposts:      acc.Update(u).Set("posts=posts+?,bigposts=bigposts+?").Where(w).Prepare(),
			incMegaposts:     acc.Update(u).Set("posts=posts+?,bigposts=bigposts+?,megaposts=megaposts+?").Where(w).Prepare(),
			incPostStats:     acc.Update(u).Set("posts=posts+?,score=score+?,level=?").Where(w).Prepare(),
			incBigpostStats:  acc.Update(u).Set("posts=posts+?,bigposts=bigposts+?,score=score+?,level=?").Where(w).Prepare(),
			incMegapostStats: acc.Update(u).Set("posts=posts+?,bigposts=bigposts+?,megaposts=megaposts+?,score=score+?,level=?").Where(w).Prepare(),
			incTopics:        acc.SimpleUpdate(u, "topics=topics+?", w),
			updateLevel:      acc.SimpleUpdate(u, "level=?", w),
			resetStats:       acc.Update(u).Set("score=0,posts=0,bigposts=0,megaposts=0,topics=0,level=0").Where(w).Prepare(),
			setStats:         acc.Update(u).Set("score=?,posts=?,bigposts=?,megaposts=?,topics=?,level=?").Where(w).Prepare(),

			incLiked: acc.Update(u).Set("liked=liked+?,lastLiked=UTC_TIMESTAMP()").Where(w).Prepare(),
			decLiked: acc.Update(u).Set("liked=liked-?").Where(w).Prepare(),
			//recalcLastLiked: acc...
			updateLastIP:  acc.SimpleUpdate(u, "last_ip=?", w),
			updatePrivacy: acc.Update(u).Set("profile_comments=?,enable_embeds=?").Where(w).Prepare(),

			setPassword: acc.Update(u).Set("password=?,salt=?").Where(w).Prepare(),

			scheduleAvatarResize: acc.Insert("users_avatar_queue").Columns("uid").Fields("?").Prepare(),

			// Delete All Posts Statements
			deletePosts:            acc.Select("topics").Columns("tid,parentID,postCount,poll").Where("createdBy=?").Prepare(),
			deleteProfilePosts:     acc.Select("users_replies").Columns("rid,uid").Where("createdBy=?").Prepare(),
			deleteReplyPosts:       acc.Select("replies").Columns("rid,tid").Where("createdBy=?").Prepare(),
			getLikedRepliesOfTopic: acc.Select("replies").Columns("rid").Where("tid=? AND likeCount>0").Prepare(),
			getAttachmentsOfTopic:  acc.Select("attachments").Columns("attachID").Where("originID=? AND originTable='topics'").Prepare(),
			getAttachmentsOfTopic2: acc.Select("attachments").Columns("attachID").Where("extra=? AND originTable='replies'").Prepare(),
			getRepliesOfTopic:      acc.Select("replies").Columns("words").Where("createdBy!=? AND tid=?").Prepare(),
		}
		return acc.FirstError()
	})
}

func (u *User) Init() {
	// TODO: Let admins configure the minimum default?
	if u.Privacy.ShowComments < 1 {
		u.Privacy.ShowComments = 1
	}
	u.Avatar, u.MicroAvatar = BuildAvatar(u.ID, u.RawAvatar)
	u.Link = BuildProfileURL(NameToSlug(u.Name), u.ID)
	u.Tag = Groups.DirtyGet(u.Group).Tag
	u.InitPerms()
}

// TODO: Refactor this idiom into something shorter, maybe with a NullUserCache when one isn't set?
func (u *User) CacheRemove() {
	if uc := Users.GetCache(); uc != nil {
		uc.Remove(u.ID)
	}
	TopicListThaw.Thaw()
}

func (u *User) Ban(dur time.Duration, issuedBy int) error {
	return u.ScheduleGroupUpdate(BanGroup, issuedBy, dur)
}

func (u *User) Unban() error {
	return u.RevertGroupUpdate()
}

func (u *User) deleteScheduleGroupTx(tx *sql.Tx) error {
	deleteScheduleGroupStmt, e := qgen.Builder.SimpleDeleteTx(tx, "users_groups_scheduler", "uid=?")
	if e != nil {
		return e
	}
	_, e = deleteScheduleGroupStmt.Exec(u.ID)
	return e
}

func (u *User) setTempGroupTx(tx *sql.Tx, tempGroup int) error {
	setTempGroupStmt, e := qgen.Builder.SimpleUpdateTx(tx, "users", "temp_group=?", "uid=?")
	if e != nil {
		return e
	}
	_, e = setTempGroupStmt.Exec(tempGroup, u.ID)
	return e
}

// Make this more stateless?
func (u *User) ScheduleGroupUpdate(gid, issuedBy int, dur time.Duration) error {
	var temp bool
	if dur.Nanoseconds() != 0 {
		temp = true
	}
	revertAt := time.Now().Add(dur)

	tx, err := qgen.Builder.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = u.deleteScheduleGroupTx(tx)
	if err != nil {
		return err
	}

	createScheduleGroupTx, err := qgen.Builder.SimpleInsertTx(tx, "users_groups_scheduler", "uid,set_group,issued_by,issued_at,revert_at,temporary", "?,?,?,UTC_TIMESTAMP(),?,?")
	if err != nil {
		return err
	}
	_, err = createScheduleGroupTx.Exec(u.ID, gid, issuedBy, revertAt, temp)
	if err != nil {
		return err
	}

	err = u.setTempGroupTx(tx, gid)
	if err != nil {
		return err
	}
	err = tx.Commit()

	u.CacheRemove()
	return err
}

func (u *User) RevertGroupUpdate() error {
	tx, e := qgen.Builder.Begin()
	if e != nil {
		return e
	}
	defer tx.Rollback()

	e = u.deleteScheduleGroupTx(tx)
	if e != nil {
		return e
	}

	e = u.setTempGroupTx(tx, 0)
	if e != nil {
		return e
	}
	e = tx.Commit()

	u.CacheRemove()
	return e
}

// TODO: Use a transaction here
// ? - Add a Deactivate method? Not really needed, if someone's been bad you could do a ban, I guess it might be useful, if someone says that email x isn't actually owned by the user in question?
func (u *User) Activate() (e error) {
	_, e = userStmts.activate.Exec(u.ID)
	if e != nil {
		return e
	}
	_, e = userStmts.changeGroup.Exec(Config.DefaultGroup, u.ID)
	u.CacheRemove()
	return e
}

// TODO: Write tests for this
// TODO: Delete this user's content too?
// TODO: Expose this to the admin?
func (u *User) Delete() error {
	_, e := userStmts.delete.Exec(u.ID)
	if e != nil {
		return e
	}
	u.CacheRemove()
	return nil
}

// TODO: dismiss-event
func (u *User) DeletePosts() error {
	rows, err := userStmts.deletePosts.Query(u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	defer TopicListThaw.Thaw()
	defer u.CacheRemove()

	updatedForums := make(map[int]int) // forum[count]
	tc := Topics.GetCache()
	umap := make(map[int]struct{})
	for rows.Next() {
		var tid, parentID, postCount, poll int
		err := rows.Scan(&tid, &parentID, &postCount, &poll)
		if err != nil {
			return err
		}
		// TODO: Clear reply cache too
		_, err = topicStmts.delete.Exec(tid)
		if tc != nil {
			tc.Remove(tid)
		}
		if err != nil {
			return err
		}
		updatedForums[parentID] = updatedForums[parentID] + 1

		_, err = topicStmts.deleteLikesForTopic.Exec(tid)
		if err != nil {
			return err
		}
		err = handleTopicAttachments(tid)
		if err != nil {
			return err
		}
		if postCount > 1 {
			err = handleLikedTopicReplies(tid)
			if err != nil {
				return err
			}
			err = handleTopicReplies(umap, u.ID, tid)
			if err != nil {
				return err
			}
			_, err = topicStmts.deleteReplies.Exec(tid)
			if err != nil {
				return err
			}
		}
		err = Subscriptions.DeleteResource(tid, "topic")
		if err != nil {
			return err
		}
		_, err = topicStmts.deleteActivity.Exec(tid)
		if err != nil {
			return err
		}
		if poll > 0 {
			err = (&Poll{ID: poll}).Delete()
			if err != nil {
				return err
			}
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}
	err = u.ResetPostStats()
	if err != nil {
		return err
	}
	for uid, _ := range umap {
		err = (&User{ID: uid}).RecalcPostStats()
		if err != nil {
			return err
		}
	}
	for fid, count := range updatedForums {
		err := Forums.RemoveTopics(fid, count)
		if err != nil && err != ErrNoRows {
			return err
		}
	}

	rows, err = userStmts.deleteProfilePosts.Query(u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rid, uid int
		err := rows.Scan(&rid, &uid)
		if err != nil {
			return err
		}
		_, err = profileReplyStmts.delete.Exec(rid)
		if err != nil {
			return err
		}
		// TODO: Optimise this
		// TODO: dismiss-event
		err = Activity.DeleteByParamsExtra("reply", uid, "user", strconv.Itoa(rid))
		if err != nil {
			return err
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}

	rows, err = userStmts.deleteReplyPosts.Query(u.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	rc := Rstore.GetCache()
	for rows.Next() {
		var rid, tid int
		err := rows.Scan(&rid, &tid)
		if err != nil {
			return err
		}
		_, err = replyStmts.delete.Exec(rid)
		if err != nil {
			return err
		}
		// TODO: Move this bit to *Topic
		_, err = replyStmts.removeRepliesFromTopic.Exec(1, tid)
		if err != nil {
			return err
		}
		_, err = replyStmts.updateTopicReplies.Exec(tid)
		if err != nil {
			return err
		}
		_, err = replyStmts.updateTopicReplies2.Exec(tid)
		if tc != nil {
			tc.Remove(tid)
		}
		_ = rc.Remove(rid)
		if err != nil {
			return err
		}

		_, err = replyStmts.deleteLikesForReply.Exec(rid)
		if err != nil {
			return err
		}
		err = Activity.DeleteByParamsExtra("reply", tid, "topic", strconv.Itoa(rid))
		if err != nil {
			return err
		}
		_, err = replyStmts.deleteActivitySubs.Exec(rid)
		if err != nil {
			return err
		}
		_, err = replyStmts.deleteActivity.Exec(rid)
		if err != nil {
			return err
		}
		// TODO: Restructure alerts so we can delete the "x replied to topic" ones too.
	}
	return rows.Err()
}

func (u *User) bindStmt(stmt *sql.Stmt, params ...interface{}) (e error) {
	params = append(params, u.ID)
	_, e = stmt.Exec(params...)
	u.CacheRemove()
	return e
}

func (u *User) ChangeName(name string) error {
	return u.bindStmt(userStmts.setName, name)
}

func (u *User) ChangeAvatar(avatar string) error {
	return u.bindStmt(userStmts.setAvatar, avatar)
}

// TODO: Abstract this with an interface so we can scale this with an actual dedicated queue in a real cluster
func (u *User) ScheduleAvatarResize() (err error) {
	_, err = userStmts.scheduleAvatarResize.Exec(u.ID)
	if err != nil {
		// TODO: Do a more generic check so that we're not as tied to MySQL
		me, ok := err.(*mysql.MySQLError)
		if !ok {
			return err
		}
		// If it's just telling us that the item already exists in the database, then we can ignore it, as it doesn't matter if it's this call or another which schedules the item in the queue
		if me.Number != 1062 {
			return err
		}
	}
	return nil
}

func (u *User) ChangeGroup(group int) error {
	return u.bindStmt(userStmts.changeGroup, group)
}

func (u *User) GetIP() string {
	spl := strings.Split(u.LastIP, "-")
	return spl[len(spl)-1]
}

// ! Only updates the database not the *User for safety reasons
func (u *User) UpdateIP(ip string) error {
	_, err := userStmts.updateLastIP.Exec(ip, u.ID)
	if uc := Users.GetCache(); uc != nil {
		uc.Remove(u.ID)
	}
	return err
}

//var ErrMalformedInteger = errors.New("malformed integer")
var ErrProfileCommentsOutOfBounds = errors.New("profile_comments must be an integer between -1 and 4")
var ErrEnableEmbedsOutOfBounds = errors.New("enable_embeds must be -1, 0 or 1")

/*func (u *User) UpdatePrivacyS(sProfileComments, sEnableEmbeds string) error {
	return u.UpdatePrivacy(profileComments, enableEmbeds)
}*/

func (u *User) UpdatePrivacy(profileComments, enableEmbeds int) error {
	if profileComments < -1 || profileComments > 4 {
		return ErrProfileCommentsOutOfBounds
	}
	if enableEmbeds < -1 || enableEmbeds > 1 {
		return ErrEnableEmbedsOutOfBounds
	}
	_, e := userStmts.updatePrivacy.Exec(profileComments, enableEmbeds, u.ID)
	if uc := Users.GetCache(); uc != nil {
		uc.Remove(u.ID)
	}
	return e
}

func (u *User) Update(name, email string, group int) (err error) {
	return u.bindStmt(userStmts.update, name, email, group)
}

func (u *User) IncreasePostStats(wcount int, topic bool) (err error) {
	baseScore := 1
	if topic {
		_, err = userStmts.incTopics.Exec(1, u.ID)
		if err != nil {
			return err
		}
		baseScore = 2
	}

	settings := SettingBox.Load().(SettingMap)
	var mod, level int
	if wcount >= settings["megapost_min_words"].(int) {
		mod = 4
		level = GetLevel(u.Score + baseScore + mod)
		_, err = userStmts.incMegapostStats.Exec(1, 1, 1, baseScore+mod, level, u.ID)
	} else if wcount >= settings["bigpost_min_words"].(int) {
		mod = 1
		level = GetLevel(u.Score + baseScore + mod)
		_, err = userStmts.incBigpostStats.Exec(1, 1, baseScore+mod, level, u.ID)
	} else {
		level = GetLevel(u.Score + baseScore + mod)
		_, err = userStmts.incPostStats.Exec(1, baseScore+mod, level, u.ID)
	}
	if err != nil {
		return err
	}
	err = GroupPromotions.PromoteIfEligible(u, level, u.Posts+1, u.CreatedAt)
	u.CacheRemove()
	return err
}

func (u *User) countf(stmt *sql.Stmt) (count int) {
	e := stmt.QueryRow().Scan(&count)
	if e != nil {
		LogError(e)
	}
	return count
}

func (u *User) RecalcPostStats() error {
	var score int
	tcount := Topics.CountUser(u.ID)
	rcount := Rstore.CountUser(u.ID)
	//log.Print("tcount:", tcount)
	//log.Print("rcount:", rcount)
	score += tcount * 2
	score += rcount

	var tmega, tbig, rmega, rbig int
	if tcount > 0 {
		tmega = Topics.CountMegaUser(u.ID)
		score += tmega * 3
		tbig := Topics.CountBigUser(u.ID)
		score += tbig
	}
	if rcount > 0 {
		rmega = Rstore.CountMegaUser(u.ID)
		score += rmega * 3
		rbig = Rstore.CountBigUser(u.ID)
		score += rbig
	}

	_, err := userStmts.setStats.Exec(score, tcount+rcount, tbig+rbig, tmega+rmega, tcount, GetLevel(score), u.ID)
	u.CacheRemove()
	return err
}

func (u *User) DecreasePostStats(wcount int, topic bool) (err error) {
	baseScore := -1
	if topic {
		_, err = userStmts.incTopics.Exec(-1, u.ID)
		if err != nil {
			return err
		}
		baseScore = -2
	}

	// TODO: Use a transaction to prevent level desyncs?
	var mod int
	settings := SettingBox.Load().(SettingMap)
	if wcount >= settings["megapost_min_words"].(int) {
		mod = 4
		_, err = userStmts.incMegapostStats.Exec(-1, -1, -1, baseScore-mod, GetLevel(u.Score-baseScore-mod), u.ID)
	} else if wcount >= settings["bigpost_min_words"].(int) {
		mod = 1
		_, err = userStmts.incBigpostStats.Exec(-1, -1, baseScore-mod, GetLevel(u.Score-baseScore-mod), u.ID)
	} else {
		_, err = userStmts.incPostStats.Exec(-1, baseScore-mod, GetLevel(u.Score-baseScore-mod), u.ID)
	}
	u.CacheRemove()
	return err
}

func (u *User) ResetPostStats() error {
	_, err := userStmts.resetStats.Exec(u.ID)
	u.CacheRemove()
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the user
func (u *User) Copy() User {
	return *u
}

// TODO: Write unit tests for this
func (u *User) InitPerms() {
	if u.TempGroup != 0 {
		u.Group = u.TempGroup
	}

	group := Groups.DirtyGet(u.Group)
	if u.IsSuperAdmin {
		u.Perms = AllPerms
		u.PluginPerms = AllPluginPerms
	} else {
		u.Perms = group.Perms
		u.PluginPerms = group.PluginPerms
	}
	/*if len(group.CanSee) == 0 {
		panic("should not be zero")
	}*/

	u.IsAdmin = u.IsSuperAdmin || group.IsAdmin
	u.IsSuperMod = u.IsAdmin || group.IsMod
	u.IsMod = u.IsSuperMod
	u.IsBanned = group.IsBanned
	if u.IsBanned && u.IsSuperMod {
		u.IsBanned = false
	}
}

// TODO: Write tests
// TODO: Implement and use this
// TODO: Implement friends
func PrivacyAllowMessage(pu, u *User) (canMsg bool) {
	switch pu.Privacy.AllowMessage {
	case 4: // Unused
		canMsg = false
	case 3: // mods
		canMsg = u.IsSuperMod
	//case 2: // friends
	case 1: // registered
		canMsg = true
	default: // 0
		canMsg = true
	}
	return canMsg
}

// TODO: Implement friend system
func PrivacyCommentsShow(pu, u *User) (showComments bool) {
	switch pu.Privacy.ShowComments {
	case 5: // Unused
		showComments = false
	case 4: // Self
		showComments = u.ID == pu.ID
	case 3: // friends
		showComments = u.ID == pu.ID
	case 2: // registered
		showComments = u.Loggedin
	case 1: // public
		showComments = true
	default: // 0
		showComments = true
	}
	return showComments
}

var guestAvatar GuestAvatar

type GuestAvatar struct {
	Normal string
	Micro  string
}

func buildNoavatar(uid, width int) string {
	if !Config.DisableNoavatarRange {
		// TODO: Find a faster algorithm
		if uid > 50000 {
			uid -= 50000
		}
		if uid > 5000 {
			uid -= 5000
		}
		if uid > 500 {
			uid -= 500
		}
		for uid > 50 {
			uid -= 50
		}
	}
	if !Config.DisableDefaultNoavatar && uid < 5 {
		if width == 200 {
			return noavatarCache200[uid]
		} else if width == 48 {
			return noavatarCache48[uid]
		}
		return StaticFiles.Prefix + "n" + strconv.Itoa(uid) + "-" + strconv.Itoa(width) + ".png?i=0"
	}
	// ? - Add a prefix setting to make this faster?
	return strings.Replace(strings.Replace(Config.Noavatar, "{id}", strconv.Itoa(uid), 1), "{width}", strconv.Itoa(width), 1)
}

// ? - Make this part of *User?
// TODO: Write tests for this
func BuildAvatar(uid int, avatar string) (normalAvatar, microAvatar string) {
	if avatar == "" {
		if uid == 0 {
			return guestAvatar.Normal, guestAvatar.Micro
		}
		return buildNoavatar(uid, 200), buildNoavatar(uid, 48)
	}
	if avatar[0] == '.' {
		if avatar[1] == '.' {
			normalAvatar = Config.AvatarResBase + "avatar_" + strconv.Itoa(uid) + "_tmp" + avatar[1:]
			return normalAvatar, normalAvatar
		}
		normalAvatar = Config.AvatarResBase + "avatar_" + strconv.Itoa(uid) + avatar
		return normalAvatar, normalAvatar
	}
	return avatar, avatar
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

func BuildProfileURLSb(sb *strings.Builder, slug string, uid int) {
	if slug == "" || !Config.BuildSlugs {
		sb.Grow(6 + 1)
		sb.WriteString("/user/")
		sb.WriteString(strconv.Itoa(uid))
		return
	}
	sb.Grow(7 + 1 + len(slug))
	sb.WriteString("/user/")
	sb.WriteString(slug)
	sb.WriteRune('.')
	sb.WriteString(strconv.Itoa(uid))
}
