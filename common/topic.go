/*
*
*	Gosora Topic File
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"html"
	"html/template"

	"strconv"
	"strings"
	"time"

	//"log"

	p "github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

// This is also in reply.go
//var ErrAlreadyLiked = errors.New("This item was already liked by this user")

// ? - Add a TopicMeta struct for *Forums?

type Topic struct {
	ID          int
	Link        string
	Title       string
	Content     string
	CreatedBy   int
	IsClosed    bool
	Sticky      bool
	CreatedAt   time.Time
	LastReplyAt time.Time
	LastReplyBy int
	LastReplyID int
	ParentID    int
	Status      string // Deprecated. Marked for removal. -Is there anything we could use it for?
	IP          string
	ViewCount   int64
	PostCount   int
	LikeCount   int
	AttachCount int
	WeekViews   int
	ClassName   string // CSS Class Name
	Poll        int
	Data        string // Used for report metadata

	Rids []int
}

type TopicUser struct {
	ID          int
	Link        string
	Title       string
	Content     string // TODO: Avoid converting this to bytes in templates, particularly if it's long
	CreatedBy   int
	IsClosed    bool
	Sticky      bool
	CreatedAt   time.Time
	LastReplyAt time.Time
	LastReplyBy int
	LastReplyID int
	ParentID    int
	Status      string // Deprecated. Marked for removal.
	IP          string
	ViewCount   int64
	PostCount   int
	LikeCount   int
	AttachCount int
	ClassName   string
	Poll        int
	Data        string // Used for report metadata

	UserLink      string
	CreatedByName string
	Group         int
	Avatar        string
	MicroAvatar   string
	ContentLines  int
	ContentHTML   string // TODO: Avoid converting this to bytes in templates, particularly if it's long
	Tag           string
	URL           string
	//URLPrefix     string
	//URLName       string
	Level int
	Liked bool

	Attachments []*MiniAttachment
	Rids        []int
	Deletable   bool
}

type TopicsRowMut struct {
	*TopicsRow
	CanMod bool
}

// TODO: Embed TopicUser to simplify this structure and it's related logic?
type TopicsRow struct {
	Topic
	LastPage int

	Creator      *User
	CSS          template.CSS
	ContentLines int
	LastUser     *User

	ForumName string //TopicsRow
	ForumLink string
}

type WsTopicsRow struct {
	ID                  int
	Link                string
	Title               string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           time.Time
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	LastReplyBy         int
	LastReplyID         int
	ParentID            int
	ViewCount           int64
	PostCount           int
	LikeCount           int
	AttachCount         int
	ClassName           string
	Creator             *WsJSONUser
	LastUser            *WsJSONUser
	ForumName           string
	ForumLink           string
	CanMod              bool
}

// TODO: Can we get the client side to render the relative times instead?
func (r *TopicsRow) WebSockets() *WsTopicsRow {
	return &WsTopicsRow{r.ID, r.Link, r.Title, r.CreatedBy, r.IsClosed, r.Sticky, r.CreatedAt, r.LastReplyAt, RelativeTime(r.LastReplyAt), r.LastReplyBy, r.LastReplyID, r.ParentID, r.ViewCount, r.PostCount, r.LikeCount, r.AttachCount, r.ClassName, r.Creator.WebSockets(), r.LastUser.WebSockets(), r.ForumName, r.ForumLink, false}
}

// TODO: Can we get the client side to render the relative times instead?
func (r *TopicsRow) WebSockets2(canMod bool) *WsTopicsRow {
	return &WsTopicsRow{r.ID, r.Link, r.Title, r.CreatedBy, r.IsClosed, r.Sticky, r.CreatedAt, r.LastReplyAt, RelativeTime(r.LastReplyAt), r.LastReplyBy, r.LastReplyID, r.ParentID, r.ViewCount, r.PostCount, r.LikeCount, r.AttachCount, r.ClassName, r.Creator.WebSockets(), r.LastUser.WebSockets(), r.ForumName, r.ForumLink, canMod}
}

// TODO: Stop relying on so many struct types?
// ! Not quite safe as Topic doesn't contain all the data needed to constructs a TopicsRow
func (t *Topic) TopicsRow() *TopicsRow {
	lastPage := 1
	var creator *User = nil
	contentLines := 1
	var lastUser *User = nil
	forumName := ""
	forumLink := ""

	//return &TopicsRow{t.ID, t.Link, t.Title, t.Content, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, t.LastReplyBy, t.LastReplyID, t.ParentID, t.Status, t.IP, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, lastPage, t.ClassName, t.Poll, t.Data, creator, "", contentLines, lastUser, forumName, forumLink, t.Rids}
	return &TopicsRow{*t, lastPage, creator, "", contentLines, lastUser, forumName, forumLink}
}

// ! Some data may be lost in the conversion
/*func (t *TopicsRow) Topic() *Topic {
	//return &Topic{t.ID, t.Link, t.Title, t.Content, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, t.LastReplyBy, t.LastReplyID, t.ParentID, t.Status, t.IP, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, t.ClassName, t.Poll, t.Data, t.Rids}
	return &t.Topic
}*/

// ! Not quite safe as Topic doesn't contain all the data needed to constructs a WsTopicsRow
/*func (t *Topic) WsTopicsRows() *WsTopicsRow {
	var creator *User = nil
	var lastUser *User = nil
	forumName := ""
	forumLink := ""
	return &WsTopicsRow{t.ID, t.Link, t.Title, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, RelativeTime(t.LastReplyAt), t.LastReplyBy, t.LastReplyID, t.ParentID, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, t.ClassName, creator, lastUser, forumName, forumLink}
}*/

type TopicStmts struct {
	getRids             *sql.Stmt
	getReplies          *sql.Stmt
	getReplies2         *sql.Stmt
	getReplies3         *sql.Stmt
	addReplies          *sql.Stmt
	updateLastReply     *sql.Stmt
	lock                *sql.Stmt
	unlock              *sql.Stmt
	moveTo              *sql.Stmt
	stick               *sql.Stmt
	unstick             *sql.Stmt
	hasLikedTopic       *sql.Stmt
	createLike          *sql.Stmt
	addLikesToTopic     *sql.Stmt
	delete              *sql.Stmt
	deleteReplies       *sql.Stmt
	deleteLikesForTopic *sql.Stmt
	deleteActivity      *sql.Stmt
	edit                *sql.Stmt
	setPoll             *sql.Stmt
	testSetCreatedAt    *sql.Stmt
	createAction        *sql.Stmt

	getTopicUser *sql.Stmt // TODO: Can we get rid of this?
	getByReplyID *sql.Stmt
}

var topicStmts TopicStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		t, w := "topics", "tid=?"
		set := func(s string) *sql.Stmt {
			return acc.Update(t).Set(s).Where(w).Prepare()
		}
		topicStmts = TopicStmts{
			getRids:             acc.Select("replies").Columns("rid").Where(w).Orderby("rid ASC").Limit("?,?").Prepare(),
			getReplies:          acc.SimpleLeftJoin("replies AS r", "users AS u", "r.rid, r.content, r.createdBy, r.createdAt, r.lastEdit, r.lastEditBy, u.avatar, u.name, u.group, u.level, r.ip, r.likeCount, r.attachCount, r.actionType", "r.createdBy=u.uid", "r.tid=?", "r.rid ASC", "?,?"),
			getReplies2:         acc.SimpleLeftJoin("replies AS r", "users AS u", "r.rid, r.content, r.createdBy, r.createdAt, r.lastEdit, r.lastEditBy, u.avatar, u.name, u.group, u.level, r.likeCount, r.attachCount, r.actionType", "r.createdBy=u.uid", "r.tid=?", "r.rid ASC", "?,?"),
			getReplies3:         acc.Select("replies").Columns("rid,content,createdBy,createdAt,lastEdit,lastEditBy,likeCount,attachCount,actionType").Where(w).Orderby("rid ASC").Limit("?,?").Prepare(),
			addReplies:          set("postCount=postCount+?,lastReplyBy=?,lastReplyAt=UTC_TIMESTAMP()"),
			updateLastReply:     acc.Update(t).Set("lastReplyID=?").Where("lastReplyID<? AND tid=?").Prepare(),
			lock:                set("is_closed=1"),
			unlock:              set("is_closed=0"),
			moveTo:              set("parentID=?"),
			stick:               set("sticky=1"),
			unstick:             set("sticky=0"),
			hasLikedTopic:       acc.Select("likes").Columns("targetItem").Where("sentBy=? and targetItem=? and targetType='topics'").Prepare(),
			createLike:          acc.Insert("likes").Columns("weight,targetItem,targetType,sentBy,createdAt").Fields("?,?,?,?,UTC_TIMESTAMP()").Prepare(),
			addLikesToTopic:     set("likeCount=likeCount+?"),
			delete:              acc.Delete(t).Where(w).Prepare(),
			deleteReplies:       acc.Delete("replies").Where(w).Prepare(),
			deleteLikesForTopic: acc.Delete("likes").Where("targetItem=? AND targetType='topics'").Prepare(),
			deleteActivity:      acc.Delete("activity_stream").Where("elementID=? AND elementType='topic'").Prepare(),
			edit:                set("title=?,content=?,parsed_content=?"), // TODO: Only run the content update bits on non-polls, does this matter?
			setPoll:             acc.Update(t).Set("poll=?").Where("tid=? AND poll=0").Prepare(),
			testSetCreatedAt:    set("createdAt=?"),
			createAction:        acc.Insert("replies").Columns("tid,actionType,ip,createdBy,createdAt,lastUpdated,content,parsed_content").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),'',''").Prepare(),

			getTopicUser: acc.SimpleLeftJoin("topics AS t", "users AS u", "t.title, t.content, t.createdBy, t.createdAt, t.lastReplyAt, t.lastReplyBy, t.lastReplyID, t.is_closed, t.sticky, t.parentID, t.ip, t.views, t.postCount, t.likeCount, t.attachCount,t.poll, u.name, u.avatar, u.group, u.level", "t.createdBy=u.uid", w, "", ""),
			getByReplyID: acc.SimpleLeftJoin("replies AS r", "topics AS t", "t.tid, t.title, t.content, t.createdBy, t.createdAt, t.is_closed, t.sticky, t.parentID, t.ip, t.views, t.postCount, t.likeCount, t.poll, t.data", "r.tid=t.tid", "rid=?", "", ""),
		}
		return acc.FirstError()
	})
}

// Flush the topic out of the cache
// ? - We do a CacheRemove() here instead of mutating the pointer to avoid creating a race condition
func (t *Topic) cacheRemove() {
	if tc := Topics.GetCache(); tc != nil {
		tc.Remove(t.ID)
	}
	TopicListThaw.Thaw()
}

// TODO: Write a test for this
func (t *Topic) AddReply(rid, uid int) (e error) {
	_, e = topicStmts.addReplies.Exec(1, uid, t.ID)
	if e != nil {
		return e
	}
	_, e = topicStmts.updateLastReply.Exec(rid, rid, t.ID)
	t.cacheRemove()
	return e
}

func (t *Topic) Lock() (e error) {
	_, e = topicStmts.lock.Exec(t.ID)
	t.cacheRemove()
	return e
}

func (t *Topic) Unlock() (e error) {
	_, e = topicStmts.unlock.Exec(t.ID)
	t.cacheRemove()
	return e
}

func (t *Topic) MoveTo(destForum int) (e error) {
	_, e = topicStmts.moveTo.Exec(destForum, t.ID)
	t.cacheRemove()
	if e != nil {
		return e
	}
	e = Attachments.MoveTo(destForum, t.ID, "topics")
	if e != nil {
		return e
	}
	return Attachments.MoveToByExtra(destForum, "replies", strconv.Itoa(t.ID))
}

func (t *Topic) TestSetCreatedAt(s time.Time) (e error) {
	_, e = topicStmts.testSetCreatedAt.Exec(s, t.ID)
	t.cacheRemove()
	return e
}

// TODO: We might want more consistent terminology rather than using stick in some places and pin in others. If you don't understand the difference, there is none, they are one and the same.
func (t *Topic) Stick() (e error) {
	_, e = topicStmts.stick.Exec(t.ID)
	t.cacheRemove()
	return e
}

func (t *Topic) Unstick() (e error) {
	_, e = topicStmts.unstick.Exec(t.ID)
	t.cacheRemove()
	return e
}

// TODO: Test this
// TODO: Use a transaction for this
func (t *Topic) Like(score, uid int) (err error) {
	var disp int // Unused
	err = topicStmts.hasLikedTopic.QueryRow(uid, t.ID).Scan(&disp)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}
	_, err = topicStmts.createLike.Exec(score, t.ID, "topics", uid)
	if err != nil {
		return err
	}
	_, err = topicStmts.addLikesToTopic.Exec(1, t.ID)
	if err != nil {
		return err
	}
	_, err = userStmts.incLiked.Exec(1, uid)
	t.cacheRemove()
	return err
}

// TODO: Use a transaction
func (t *Topic) Unlike(uid int) error {
	e := Likes.Delete(t.ID, "topics")
	if e != nil {
		return e
	}
	_, e = topicStmts.addLikesToTopic.Exec(-1, t.ID)
	if e != nil {
		return e
	}
	_, e = userStmts.decLiked.Exec(1, uid)
	t.cacheRemove()
	return e
}

func handleLikedTopicReplies(tid int) error {
	rows, e := userStmts.getLikedRepliesOfTopic.Query(tid)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var rid int
		if e := rows.Scan(&rid); e != nil {
			return e
		}
		_, e = replyStmts.deleteLikesForReply.Exec(rid)
		if e != nil {
			return e
		}
		e = Activity.DeleteByParams("like", rid, "post")
		if e != nil {
			return e
		}
	}
	return rows.Err()
}

func handleTopicAttachments(tid int) error {
	e := handleAttachments(userStmts.getAttachmentsOfTopic, tid)
	if e != nil {
		return e
	}
	return handleAttachments(userStmts.getAttachmentsOfTopic2, tid)
}

func handleReplyAttachments(rid int) error {
	return handleAttachments(replyStmts.getAidsOfReply, rid)
}

func handleAttachments(stmt *sql.Stmt, id int) error {
	rows, e := stmt.Query(id)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var aid int
		if e := rows.Scan(&aid); e != nil {
			return e
		}
		a, e := Attachments.FGet(aid)
		if e != nil {
			return e
		}
		e = deleteAttachment(a)
		if e != nil && e != sql.ErrNoRows {
			return e
		}
	}
	return rows.Err()
}

// TODO: Only load a row per createdBy, maybe with group by?
func handleTopicReplies(umap map[int]struct{}, uid, tid int) error {
	rows, e := userStmts.getRepliesOfTopic.Query(uid, tid)
	if e != nil {
		return e
	}
	defer rows.Close()
	var createdBy int
	for rows.Next() {
		if e := rows.Scan(&createdBy); e != nil {
			return e
		}
		umap[createdBy] = struct{}{}
	}
	return rows.Err()
}

// TODO: Use a transaction here
func (t *Topic) Delete() error {
	/*creator, e := Users.Get(t.CreatedBy)
	if e == nil {
		e = creator.DecreasePostStats(WordCount(t.Content), true)
		if e != nil {
			return e
		}
	} else if e != ErrNoRows {
		return e
	}*/

	// TODO: Clear reply cache too
	_, e := topicStmts.delete.Exec(t.ID)
	t.cacheRemove()
	if e != nil {
		return e
	}
	e = Forums.RemoveTopic(t.ParentID)
	if e != nil && e != ErrNoRows {
		return e
	}
	_, e = topicStmts.deleteLikesForTopic.Exec(t.ID)
	if e != nil {
		return e
	}

	if t.PostCount > 1 {
		if e = handleLikedTopicReplies(t.ID); e != nil {
			return e
		}
		umap := make(map[int]struct{})
		e = handleTopicReplies(umap, t.CreatedBy, t.ID)
		if e != nil {
			return e
		}
		_, e = topicStmts.deleteReplies.Exec(t.ID)
		if e != nil {
			return e
		}
		for uid := range umap {
			e = (&User{ID: uid}).RecalcPostStats()
			if e != nil {
				//log.Printf("e: %+v\n", e)
				return e
			}
		}
	}
	e = (&User{ID: t.CreatedBy}).RecalcPostStats()
	if e != nil {
		return e
	}
	e = handleTopicAttachments(t.ID)
	if e != nil {
		return e
	}
	e = Subscriptions.DeleteResource(t.ID, "topic")
	if e != nil {
		return e
	}
	_, e = topicStmts.deleteActivity.Exec(t.ID)
	if e != nil {
		return e
	}
	if t.Poll > 0 {
		e = (&Poll{ID: t.Poll}).Delete()
		if e != nil {
			return e
		}
	}
	return nil
}

// TODO: Write tests for this
func (t *Topic) Update(name, content string) error {
	name = SanitiseSingleLine(html.UnescapeString(name))
	if name == "" {
		return ErrNoTitle
	}
	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(name) > Config.MaxTopicTitleLength {
		return ErrLongTitle
	}

	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, t.ParentID, "forums", nil, nil)
	_, err := topicStmts.edit.Exec(name, content, parsedContent, t.ID)
	t.cacheRemove()
	return err
}

func (t *Topic) SetPoll(pollID int) error {
	_, e := topicStmts.setPoll.Exec(pollID, t.ID) // TODO: Sniff if this changed anything to see if we hit an existing poll
	t.cacheRemove()
	return e
}

// TODO: Have this go through the ReplyStore?
// TODO: Return the rid?
func (t *Topic) CreateActionReply(action, ip string, uid int) (err error) {
	if Config.DisablePostIP {
		ip = ""
	}
	res, err := topicStmts.createAction.Exec(t.ID, action, ip, uid)
	if err != nil {
		return err
	}
	_, err = topicStmts.addReplies.Exec(1, uid, t.ID)
	if err != nil {
		return err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rid := int(lid)
	_, err = topicStmts.updateLastReply.Exec(rid, rid, t.ID)
	t.cacheRemove()
	// ? - Update the last topic cache for the parent forum?
	return err
}

func GetRidsForTopic(tid, offset int) (rids []int, e error) {
	rows, e := topicStmts.getRids.Query(tid, offset, Config.ItemsPerPage)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var rid int
	for rows.Next() {
		if e := rows.Scan(&rid); e != nil {
			return nil, e
		}
		rids = append(rids, rid)
	}
	return rids, rows.Err()
}

var aipost = ";&#xFE0E"
var lockai = "&#x1F512" + aipost
var unlockai = "&#x1F513"
var stickai = "&#x1F4CC"
var unstickai = "&#x1F4CC" + aipost

func (ru *ReplyUser) Init(u *User) (group *Group, err error) {
	ru.ContentLines = strings.Count(ru.Content, "\n")

	postGroup, err := Groups.Get(ru.Group)
	if err != nil {
		return nil, err
	}
	if postGroup.IsMod {
		ru.ClassName = Config.StaffCSS
	}
	ru.Tag = postGroup.Tag

	if u.ID != ru.CreatedBy {
		ru.UserLink = BuildProfileURL(NameToSlug(ru.CreatedByName), ru.CreatedBy)
		// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the c.UserStore initialise this?
		ru.Avatar, ru.MicroAvatar = BuildAvatar(ru.CreatedBy, ru.Avatar)
	} else {
		ru.UserLink = u.Link
		ru.Avatar, ru.MicroAvatar = u.Avatar, u.MicroAvatar
	}

	// We really shouldn't have inline HTML, we should do something about this...
	if ru.ActionType != "" {
		aarr := strings.Split(ru.ActionType, "-")
		action := aarr[0]
		switch action {
		case "lock":
			ru.ActionIcon = lockai
		case "unlock":
			ru.ActionIcon = unlockai
		case "stick":
			ru.ActionIcon = stickai
		case "unstick":
			ru.ActionIcon = unstickai
		case "move":
			if len(aarr) == 2 {
				fid, _ := strconv.Atoi(aarr[1])
				forum, err := Forums.Get(fid)
				if err == nil {
					ru.ActionType = p.GetTmplPhrasef("topic.action_topic_move_dest", forum.Link, forum.Name, ru.UserLink, ru.CreatedByName)
					return postGroup, nil
				}
			}
		default:
			// TODO: Only fire this off if a corresponding phrase for the ActionType doesn't exist? Or maybe have some sort of action registry?
			ru.ActionType = p.GetTmplPhrasef("topic.action_topic_default", ru.ActionType)
			return postGroup, nil
		}
		ru.ActionType = p.GetTmplPhrasef("topic.action_topic_"+action, ru.UserLink, ru.CreatedByName)
	}

	return postGroup, nil
}

func (ru *ReplyUser) Init2() (group *Group, err error) {
	//ru.UserLink = BuildProfileURL(NameToSlug(ru.CreatedByName), ru.CreatedBy)
	ru.ContentLines = strings.Count(ru.Content, "\n")

	postGroup, err := Groups.Get(ru.Group)
	if err != nil {
		return postGroup, err
	}
	if postGroup.IsMod {
		ru.ClassName = Config.StaffCSS
	}
	ru.Tag = postGroup.Tag

	// We really shouldn't have inline HTML, we should do something about this...
	if ru.ActionType != "" {
		aarr := strings.Split(ru.ActionType, "-")
		action := aarr[0]
		switch action {
		case "lock":
			ru.ActionIcon = lockai
		case "unlock":
			ru.ActionIcon = unlockai
		case "stick":
			ru.ActionIcon = stickai
		case "unstick":
			ru.ActionIcon = unstickai
		case "move":
			if len(aarr) == 2 {
				fid, _ := strconv.Atoi(aarr[1])
				forum, err := Forums.Get(fid)
				if err == nil {
					ru.ActionType = p.GetTmplPhrasef("topic.action_topic_move_dest", forum.Link, forum.Name, ru.UserLink, ru.CreatedByName)
					return postGroup, nil
				}
			}
		default:
			// TODO: Only fire this off if a corresponding phrase for the ActionType doesn't exist? Or maybe have some sort of action registry?
			ru.ActionType = p.GetTmplPhrasef("topic.action_topic_default", ru.ActionType)
			return postGroup, nil
		}
		ru.ActionType = p.GetTmplPhrasef("topic.action_topic_"+action, ru.UserLink, ru.CreatedByName)
	}

	return postGroup, nil
}

func (ru *ReplyUser) Init3(u *User, tu *TopicUser) (group *Group, err error) {
	ru.ContentLines = strings.Count(ru.Content, "\n")

	postGroup, err := Groups.Get(ru.Group)
	if err != nil {
		return postGroup, err
	}
	if postGroup.IsMod {
		ru.ClassName = Config.StaffCSS
	}
	ru.Tag = postGroup.Tag

	if u.ID == ru.CreatedBy {
		ru.UserLink = u.Link
		ru.Avatar, ru.MicroAvatar = u.Avatar, u.MicroAvatar
	} else if tu.CreatedBy == ru.CreatedBy {
		ru.UserLink = tu.UserLink
		ru.Avatar, ru.MicroAvatar = tu.Avatar, tu.MicroAvatar
	} else {
		ru.UserLink = BuildProfileURL(NameToSlug(ru.CreatedByName), ru.CreatedBy)
		// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the c.UserStore initialise this?
		ru.Avatar, ru.MicroAvatar = BuildAvatar(ru.CreatedBy, ru.Avatar)
	}

	// We really shouldn't have inline HTML, we should do something about this...
	if ru.ActionType != "" {
		aarr := strings.Split(ru.ActionType, "-")
		action := aarr[0]
		switch action {
		case "lock":
			ru.ActionIcon = lockai
			action = "topic.action_topic_lock"
		case "unlock":
			ru.ActionIcon = unlockai
			action = "topic.action_topic_unlock"
		case "stick":
			ru.ActionIcon = stickai
			action = "topic.action_topic_stick"
		case "unstick":
			ru.ActionIcon = unstickai
			action = "topic.action_topic_unstick"
		case "move":
			if len(aarr) == 2 {
				fid, _ := strconv.Atoi(aarr[1])
				forum, err := Forums.Get(fid)
				if err == nil {
					ru.ActionType = p.GetTmplPhrasef("topic.action_topic_move_dest", forum.Link, forum.Name, ru.UserLink, ru.CreatedByName)
					return postGroup, nil
				}
			}
		default:
			// TODO: Only fire this off if a corresponding phrase for the ActionType doesn't exist? Or maybe have some sort of action registry?
			ru.ActionType = p.GetTmplPhrasef("topic.action_topic_default", ru.ActionType)
			return postGroup, nil
		}
		ru.ActionType = p.GetTmplPhrasef(action, ru.UserLink, ru.CreatedByName)
	}

	return postGroup, nil
}

// TODO: Factor TopicUser into a *Topic and *User, as this starting to become overly complicated x.x
func (t *TopicUser) Replies(offset int /*pFrag int, */, user *User) (rlist []*ReplyUser /*, ogdesc string*/, externalHead bool, err error) {
	var likedMap, attachMap map[int]int
	var likedQueryList, attachQueryList []int

	var rid int
	if len(t.Rids) > 0 {
		//log.Print("have rid")
		rid = t.Rids[0]
	}
	re, err := Rstore.GetCache().Get(rid)
	ucache := Users.GetCache()
	var ruser *User
	if ucache != nil {
		//log.Print("ucache step")
		if err == nil {
			ruser = ucache.Getn(re.CreatedBy)
		} else if t.PostCount == 2 {
			ruser = ucache.Getn(t.LastReplyBy)
		}
	}

	hTbl := GetHookTable()
	rf := func(r *ReplyUser) (err error) {
		//log.Printf("before r: %+v\n", r)
		postGroup, err := r.Init3(user, t)
		if err != nil {
			return err
		}
		//log.Printf("after r: %+v\n", r)

		var parseSettings *ParseSettings
		if (Config.NoEmbed || !postGroup.Perms.AutoEmbed) && (user.ParseSettings == nil || !user.ParseSettings.NoEmbed) {
			parseSettings = DefaultParseSettings.CopyPtr()
			parseSettings.NoEmbed = true
		} else {
			parseSettings = user.ParseSettings
		}
		/*if user.ParseSettings == nil {
			parseSettings = DefaultParseSettings.CopyPtr()
			parseSettings.NoEmbed = Config.NoEmbed || !postGroup.Perms.AutoEmbed
			parseSettings.NoLink = !postGroup.Perms.AutoLink
		} else {
			parseSettings = user.ParseSettings
		}*/

		var eh bool
		r.ContentHtml, eh = ParseMessage2(r.Content, t.ParentID, "forums", parseSettings, user)
		if eh {
			externalHead = true
		}
		// TODO: Do this more efficiently by avoiding the allocations entirely in ParseMessage, if there's nothing to do.
		if r.ContentHtml == r.Content {
			r.ContentHtml = r.Content
		}
		r.Deletable = user.Perms.DeleteReply || r.CreatedBy == user.ID

		// TODO: This doesn't work properly so pick the first one instead?
		/*if r.ID == pFrag {
			ogdesc = r.Content
			if len(ogdesc) > 200 {
				ogdesc = ogdesc[:197] + "..."
			}
		}*/

		return nil
	}

	rf3 := func(r *ReplyUser) error {
		//log.Printf("before r: %+v\n", r)
		postGroup, err := r.Init2()
		if err != nil {
			return err
		}

		var parseSettings *ParseSettings
		if (Config.NoEmbed || !postGroup.Perms.AutoEmbed) && (user.ParseSettings == nil || !user.ParseSettings.NoEmbed) {
			parseSettings = DefaultParseSettings.CopyPtr()
			parseSettings.NoEmbed = true
		} else {
			parseSettings = user.ParseSettings
		}

		var eh bool
		r.ContentHtml, eh = ParseMessage2(r.Content, t.ParentID, "forums", parseSettings, user)
		if eh {
			externalHead = true
		}
		// TODO: Do this more efficiently by avoiding the allocations entirely in ParseMessage, if there's nothing to do.
		if r.ContentHtml == r.Content {
			r.ContentHtml = r.Content
		}
		r.Deletable = user.Perms.DeleteReply || r.CreatedBy == user.ID

		return nil
	}

	// TODO: Factor the user fields out and embed a user struct instead
	if err == nil && ruser != nil {
		//log.Print("reply cached serve")
		r := &ReplyUser{ /*ClassName: "", */ Reply: *re, CreatedByName: ruser.Name, UserLink: ruser.Link, Avatar: ruser.Avatar, MicroAvatar: ruser.MicroAvatar /*URLPrefix: ruser.URLPrefix, URLName: ruser.URLName, */, Group: ruser.Group, Level: ruser.Level, Tag: ruser.Tag}
		if err = rf3(r); err != nil {
			return nil, externalHead, err
		}

		if r.LikeCount > 0 && user.Liked > 0 {
			likedMap = map[int]int{r.ID: 0}
			likedQueryList = []int{r.ID}
		}
		if user.Perms.EditReply && r.AttachCount > 0 {
			if likedMap == nil {
				attachMap = map[int]int{r.ID: 0}
				attachQueryList = []int{r.ID}
			} else {
				attachMap = likedMap
				attachQueryList = likedQueryList
			}
		}

		H_topic_reply_row_assign_hook(hTbl, r)
		// TODO: Use a pointer instead to make it easier to abstract this loop? What impact would this have on escape analysis?
		rlist = []*ReplyUser{r}
		//log.Printf("r: %d-%d", r.ID, len(rlist)-1)
	} else {
		//log.Print("reply query serve")
		ap1 := func(r *ReplyUser) {
			H_topic_reply_row_assign_hook(hTbl, r)
			// TODO: Use a pointer instead to make it easier to abstract this loop? What impact would this have on escape analysis?
			rlist = append(rlist, r)
			//log.Printf("r: %d-%d", r.ID, len(rlist)-1)
		}
		rf2 := func(r *ReplyUser) {
			if r.LikeCount > 0 && user.Liked > 0 {
				if likedMap == nil {
					likedMap = map[int]int{r.ID: len(rlist)}
					likedQueryList = []int{r.ID}
				} else {
					likedMap[r.ID] = len(rlist)
					likedQueryList = append(likedQueryList, r.ID)
				}
			}
			if user.Perms.EditReply && r.AttachCount > 0 {
				if attachMap == nil {
					attachMap = map[int]int{r.ID: len(rlist)}
					attachQueryList = []int{r.ID}
				} else {
					attachMap[r.ID] = len(rlist)
					attachQueryList = append(attachQueryList, r.ID)
				}
			}
		}
		if !user.Perms.ViewIPs && ruser != nil {
			rows, e := topicStmts.getReplies3.Query(t.ID, offset, Config.ItemsPerPage)
			if err != nil {
				return nil, externalHead, e
			}
			defer rows.Close()
			for rows.Next() {
				r := &ReplyUser{Avatar: ruser.Avatar, MicroAvatar: ruser.MicroAvatar, UserLink: ruser.Link, CreatedByName: ruser.Name, Group: ruser.Group, Level: ruser.Level}
				e := rows.Scan(&r.ID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.LikeCount, &r.AttachCount, &r.ActionType)
				if e != nil {
					return nil, externalHead, e
				}
				if e = rf3(r); e != nil {
					return nil, externalHead, e
				}
				rf2(r)
				ap1(r)
			}
			if e = rows.Err(); e != nil {
				return nil, externalHead, e
			}
		} else if user.Perms.ViewIPs {
			rows, err := topicStmts.getReplies.Query(t.ID, offset, Config.ItemsPerPage)
			if err != nil {
				return nil, externalHead, err
			}
			defer rows.Close()
			for rows.Next() {
				r := &ReplyUser{}
				err := rows.Scan(&r.ID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.Avatar, &r.CreatedByName, &r.Group /*&r.URLPrefix, &r.URLName,*/, &r.Level, &r.IP, &r.LikeCount, &r.AttachCount, &r.ActionType)
				if err != nil {
					return nil, externalHead, err
				}
				if err = rf(r); err != nil {
					return nil, externalHead, err
				}
				rf2(r)
				ap1(r)
			}
			if err = rows.Err(); err != nil {
				return nil, externalHead, err
			}
		} else if t.PostCount >= 20 {
			//log.Print("t.PostCount >= 20")
			rows, err := topicStmts.getReplies3.Query(t.ID, offset, Config.ItemsPerPage)
			if err != nil {
				return nil, externalHead, err
			}
			defer rows.Close()
			reqUserList := make(map[int]bool)
			for rows.Next() {
				r := &ReplyUser{}
				err := rows.Scan(&r.ID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy /*&r.URLPrefix, &r.URLName,*/, &r.LikeCount, &r.AttachCount, &r.ActionType)
				if err != nil {
					return nil, externalHead, err
				}
				if r.CreatedBy != t.CreatedBy && r.CreatedBy != user.ID {
					reqUserList[r.CreatedBy] = true
				}
				ap1(r)
			}
			if err = rows.Err(); err != nil {
				return nil, externalHead, err
			}

			if len(reqUserList) == 1 {
				//log.Print("len(reqUserList) == 1: ", len(reqUserList) == 1)
				var uitem *User
				for uid, _ := range reqUserList {
					uitem, err = Users.Get(uid)
					if err != nil {
						return nil, externalHead, nil // TODO: Implement this!
					}
				}
				for _, r := range rlist {
					if r.CreatedBy == t.CreatedBy {
						r.CreatedByName = t.CreatedByName
						r.Avatar = t.Avatar
						r.MicroAvatar = t.MicroAvatar
						r.Group = t.Group
						r.Level = t.Level
					} else {
						var u *User
						if r.CreatedBy == user.ID {
							u = user
						} else {
							u = uitem
						}
						r.CreatedByName = u.Name
						r.Avatar = u.Avatar
						r.MicroAvatar = u.MicroAvatar
						r.Group = u.Group
						r.Level = u.Level
					}
					if err = rf(r); err != nil {
						return nil, externalHead, err
					}
					rf2(r)
				}
			} else {
				//log.Print("len(reqUserList) != 1: ", len(reqUserList) != 1)
				var userList map[int]*User
				if len(reqUserList) > 0 {
					//log.Print("len(reqUserList) > 0: ", len(reqUserList) > 0)
					// Convert the user ID map to a slice, then bulk load the users
					idSlice := make([]int, len(reqUserList))
					var i int
					for userID := range reqUserList {
						idSlice[i] = userID
						i++
					}
					userList, err = Users.BulkGetMap(idSlice)
					if err != nil {
						return nil, externalHead, nil // TODO: Implement this!
					}
				}

				for _, r := range rlist {
					if r.CreatedBy == t.CreatedBy {
						r.CreatedByName = t.CreatedByName
						r.Avatar = t.Avatar
						r.MicroAvatar = t.MicroAvatar
						r.Group = t.Group
						r.Level = t.Level
					} else {
						var u *User
						if r.CreatedBy == user.ID {
							u = user
						} else {
							u = userList[r.CreatedBy]
						}
						r.CreatedByName = u.Name
						r.Avatar = u.Avatar
						r.MicroAvatar = u.MicroAvatar
						r.Group = u.Group
						r.Level = u.Level
					}
					if err = rf(r); err != nil {
						return nil, externalHead, err
					}
					rf2(r)
				}
			}
		} else {
			//log.Print("reply fallback")
			rows, err := topicStmts.getReplies2.Query(t.ID, offset, Config.ItemsPerPage)
			if err != nil {
				return nil, externalHead, err
			}
			defer rows.Close()
			for rows.Next() {
				r := &ReplyUser{}
				err := rows.Scan(&r.ID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.Avatar, &r.CreatedByName, &r.Group /*&r.URLPrefix, &r.URLName,*/, &r.Level, &r.LikeCount, &r.AttachCount, &r.ActionType)
				if err != nil {
					return nil, externalHead, err
				}
				if err = rf(r); err != nil {
					return nil, externalHead, err
				}
				rf2(r)
				ap1(r)
			}
			if err = rows.Err(); err != nil {
				return nil, externalHead, err
			}
		}
	}

	// TODO: Add a config setting to disable the liked query for a burst of extra speed
	if user.Liked > 0 && len(likedQueryList) > 0 /*&& user.LastLiked <= time.Now()*/ {
		eids, err := Likes.BulkExists(likedQueryList, user.ID, "replies")
		if err != nil {
			return nil, externalHead, err
		}
		for _, eid := range eids {
			rlist[likedMap[eid]].Liked = true
		}
	}

	if user.Perms.EditReply && len(attachQueryList) > 0 {
		//log.Printf("attachQueryList: %+v\n", attachQueryList)
		amap, err := Attachments.BulkMiniGetList("replies", attachQueryList)
		if err != nil && err != sql.ErrNoRows {
			return nil, externalHead, err
		}
		//log.Printf("amap: %+v\n", amap)
		//log.Printf("attachMap: %+v\n", attachMap)
		for id, attach := range amap {
			//log.Print("id:", id)
			rlist[attachMap[id]].Attachments = attach
			/*for _, a := range attach {
				log.Printf("a: %+v\n", a)
			}*/
		}
	}

	//hTbl.VhookNoRet("topic_reply_end", &rlist)

	return rlist, externalHead, nil
}

// TODO: Test this
func (t *Topic) Author() (*User, error) {
	return Users.Get(t.CreatedBy)
}

func (t *Topic) GetID() int {
	return t.ID
}
func (t *Topic) GetTable() string {
	return "topics"
}

// Copy gives you a non-pointer concurrency safe copy of the topic
func (t *Topic) Copy() Topic {
	return *t
}

// TODO: Load LastReplyAt and LastReplyID?
func TopicByReplyID(rid int) (*Topic, error) {
	t := Topic{ID: 0}
	err := topicStmts.getByReplyID.QueryRow(rid).Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.Poll, &t.Data)
	t.Link = BuildTopicURL(NameToSlug(t.Title), t.ID)
	return &t, err
}

// TODO: Refactor the caller to take a Topic and a User rather than a combined TopicUser
// TODO: Load LastReplyAt everywhere in here?
func GetTopicUser(user *User, tid int) (tu TopicUser, err error) {
	tcache := Topics.GetCache()
	ucache := Users.GetCache()
	if tcache != nil && ucache != nil {
		topic, err := tcache.Get(tid)
		if err == nil {
			if topic.CreatedBy != user.ID {
				user, err = Users.Get(topic.CreatedBy)
				if err != nil {
					return TopicUser{ID: tid}, err
				}
			}
			// We might be better off just passing separate topic and user structs to the caller?
			return copyTopicToTopicUser(topic, user), nil
		} else if ucache.Length() < ucache.GetCapacity() {
			topic, err = Topics.Get(tid)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			if topic.CreatedBy != user.ID {
				user, err = Users.Get(topic.CreatedBy)
				if err != nil {
					return TopicUser{ID: tid}, err
				}
			}
			return copyTopicToTopicUser(topic, user), nil
		}
	}

	tu = TopicUser{ID: tid}
	// TODO: This misses some important bits...
	err = topicStmts.getTopicUser.QueryRow(tid).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.LastReplyAt, &tu.LastReplyBy, &tu.LastReplyID, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IP, &tu.ViewCount, &tu.PostCount, &tu.LikeCount, &tu.AttachCount, &tu.Poll, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.Level)
	tu.Avatar, tu.MicroAvatar = BuildAvatar(tu.CreatedBy, tu.Avatar)
	tu.Link = BuildTopicURL(NameToSlug(tu.Title), tu.ID)
	tu.UserLink = BuildProfileURL(NameToSlug(tu.CreatedByName), tu.CreatedBy)
	tu.Tag = Groups.DirtyGet(tu.Group).Tag

	if tcache != nil {
		// TODO: weekly views
		theTopic := Topic{ID: tu.ID, Link: tu.Link, Title: tu.Title, Content: tu.Content, CreatedBy: tu.CreatedBy, IsClosed: tu.IsClosed, Sticky: tu.Sticky, CreatedAt: tu.CreatedAt, LastReplyAt: tu.LastReplyAt, LastReplyID: tu.LastReplyID, ParentID: tu.ParentID, IP: tu.IP, ViewCount: tu.ViewCount, PostCount: tu.PostCount, LikeCount: tu.LikeCount, AttachCount: tu.AttachCount, Poll: tu.Poll}
		//log.Printf("theTopic: %+v\n", theTopic)
		_ = tcache.Set(&theTopic)
	}
	return tu, err
}

func copyTopicToTopicUser(t *Topic, u *User) (tu TopicUser) {
	tu.UserLink = u.Link
	tu.CreatedByName = u.Name
	tu.Group = u.Group
	tu.Avatar = u.Avatar
	tu.MicroAvatar = u.MicroAvatar
	//tu.URLPrefix = u.URLPrefix
	//tu.URLName = u.URLName
	tu.Level = u.Level

	tu.ID = t.ID
	tu.Link = t.Link
	tu.Title = t.Title
	tu.Content = t.Content
	tu.CreatedBy = t.CreatedBy
	tu.IsClosed = t.IsClosed
	tu.Sticky = t.Sticky
	tu.CreatedAt = t.CreatedAt
	tu.LastReplyAt = t.LastReplyAt
	tu.LastReplyBy = t.LastReplyBy
	tu.ParentID = t.ParentID
	tu.IP = t.IP
	tu.ViewCount = t.ViewCount
	tu.PostCount = t.PostCount
	tu.LikeCount = t.LikeCount
	tu.AttachCount = t.AttachCount
	tu.Poll = t.Poll
	tu.Data = t.Data
	tu.Rids = t.Rids

	return tu
}

// For use in tests and for generating blank topics for forums which don't have a last poster
func BlankTopic() *Topic {
	return new(Topic)
}

func BuildTopicURL(slug string, tid int) string {
	if slug == "" || !Config.BuildSlugs {
		return "/topic/" + strconv.Itoa(tid)
	}
	return "/topic/" + slug + "." + strconv.Itoa(tid)
}

func BuildTopicURLSb(sb *strings.Builder, slug string, tid int) {
	if slug == "" || !Config.BuildSlugs {
		sb.Grow(7 + 2)
		sb.WriteString("/topic/")
		sb.WriteString(strconv.Itoa(tid))
		return
	}
	sb.Grow(7 + 3 + len(slug))
	sb.WriteString("/topic/")
	sb.WriteString(slug)
	sb.WriteRune('.')
	sb.WriteString(strconv.Itoa(tid))
}

// I don't care if it isn't used,, it will likely be in the future. Nolint.
// nolint
func getTopicURLPrefix() string {
	return "/topic/"
}
