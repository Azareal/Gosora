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
	//"log"
	"strconv"
	"strings"
	"time"

	p "github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
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
	Status      string // Deprecated. Marked for removal.
	IPAddress   string
	ViewCount   int64
	PostCount   int
	LikeCount   int
	AttachCount int
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
	IPAddress   string
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
	URLPrefix     string
	URLName       string
	Level         int
	Liked         bool

	Attachments []*MiniAttachment
	Rids        []int
}

// TODO: Embed TopicUser to simplify this structure and it's related logic?
type TopicsRow struct {
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
	IPAddress   string
	ViewCount   int64
	PostCount   int
	LikeCount   int
	AttachCount int
	LastPage    int
	ClassName   string
	Poll        int
	Data        string // Used for report metadata

	Creator      *User
	CSS          template.CSS
	ContentLines int
	LastUser     *User

	ForumName string //TopicsRow
	ForumLink string
	Rids      []int
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
}

// TODO: Can we get the client side to render the relative times instead?
func (row *TopicsRow) WebSockets() *WsTopicsRow {
	return &WsTopicsRow{row.ID, row.Link, row.Title, row.CreatedBy, row.IsClosed, row.Sticky, row.CreatedAt, row.LastReplyAt, RelativeTime(row.LastReplyAt), row.LastReplyBy, row.LastReplyID, row.ParentID, row.ViewCount, row.PostCount, row.LikeCount, row.AttachCount, row.ClassName, row.Creator.WebSockets(), row.LastUser.WebSockets(), row.ForumName, row.ForumLink}
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

	return &TopicsRow{t.ID, t.Link, t.Title, t.Content, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, t.LastReplyBy, t.LastReplyID, t.ParentID, t.Status, t.IPAddress, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, lastPage, t.ClassName, t.Poll, t.Data, creator, "", contentLines, lastUser, forumName, forumLink, t.Rids}
}

// ! Some data may be lost in the conversion
func (t *TopicsRow) Topic() *Topic {
	return &Topic{t.ID, t.Link, t.Title, t.Content, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, t.LastReplyBy, t.LastReplyID, t.ParentID, t.Status, t.IPAddress, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, t.ClassName, t.Poll, t.Data, t.Rids}
}

// ! Not quite safe as Topic doesn't contain all the data needed to constructs a WsTopicsRow
/*func (t *Topic) WsTopicsRows() *WsTopicsRow {
	var creator *User = nil
	var lastUser *User = nil
	forumName := ""
	forumLink := ""
	return &WsTopicsRow{t.ID, t.Link, t.Title, t.CreatedBy, t.IsClosed, t.Sticky, t.CreatedAt, t.LastReplyAt, RelativeTime(t.LastReplyAt), t.LastReplyBy, t.LastReplyID, t.ParentID, t.ViewCount, t.PostCount, t.LikeCount, t.AttachCount, t.ClassName, creator, lastUser, forumName, forumLink}
}*/

type TopicStmts struct {
	getRids            *sql.Stmt
	getReplies         *sql.Stmt
	addReplies         *sql.Stmt
	updateLastReply    *sql.Stmt
	lock               *sql.Stmt
	unlock             *sql.Stmt
	moveTo             *sql.Stmt
	stick              *sql.Stmt
	unstick            *sql.Stmt
	hasLikedTopic      *sql.Stmt
	createLike         *sql.Stmt
	addLikesToTopic    *sql.Stmt
	delete             *sql.Stmt
	deleteActivity     *sql.Stmt
	deleteActivitySubs *sql.Stmt
	edit               *sql.Stmt
	setPoll            *sql.Stmt
	createAction       *sql.Stmt

	getTopicUser *sql.Stmt // TODO: Can we get rid of this?
	getByReplyID *sql.Stmt
}

var topicStmts TopicStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		topicStmts = TopicStmts{
			getRids:            acc.Select("replies").Columns("rid").Where("tid = ?").Orderby("rid ASC").Limit("?,?").Prepare(),
			getReplies:         acc.SimpleLeftJoin("replies", "users", "replies.rid, replies.content, replies.createdBy, replies.createdAt, replies.lastEdit, replies.lastEditBy, users.avatar, users.name, users.group, users.url_prefix, users.url_name, users.level, replies.ipaddress, replies.likeCount, replies.attachCount, replies.actionType", "replies.createdBy = users.uid", "replies.tid = ?", "replies.rid ASC", "?,?"),
			addReplies:         acc.Update("topics").Set("postCount = postCount + ?, lastReplyBy = ?, lastReplyAt = UTC_TIMESTAMP()").Where("tid = ?").Prepare(),
			updateLastReply:    acc.Update("topics").Set("lastReplyID = ?").Where("lastReplyID > ? AND tid = ?").Prepare(),
			lock:               acc.Update("topics").Set("is_closed = 1").Where("tid = ?").Prepare(),
			unlock:             acc.Update("topics").Set("is_closed = 0").Where("tid = ?").Prepare(),
			moveTo:             acc.Update("topics").Set("parentID = ?").Where("tid = ?").Prepare(),
			stick:              acc.Update("topics").Set("sticky = 1").Where("tid = ?").Prepare(),
			unstick:            acc.Update("topics").Set("sticky = 0").Where("tid = ?").Prepare(),
			hasLikedTopic:      acc.Select("likes").Columns("targetItem").Where("sentBy = ? and targetItem = ? and targetType = 'topics'").Prepare(),
			createLike:         acc.Insert("likes").Columns("weight, targetItem, targetType, sentBy, createdAt").Fields("?,?,?,?,UTC_TIMESTAMP()").Prepare(),
			addLikesToTopic:    acc.Update("topics").Set("likeCount = likeCount + ?").Where("tid = ?").Prepare(),
			delete:             acc.Delete("topics").Where("tid = ?").Prepare(),
			deleteActivity:     acc.Delete("activity_stream").Where("elementID = ? AND elementType = 'topic'").Prepare(),
			deleteActivitySubs: acc.Delete("activity_subscriptions").Where("targetID = ? AND targetType = 'topic'").Prepare(),
			edit:               acc.Update("topics").Set("title = ?, content = ?, parsed_content = ?").Where("tid = ?").Prepare(), // TODO: Only run the content update bits on non-polls, does this matter?
			setPoll:            acc.Update("topics").Set("poll = ?").Where("tid = ? AND poll = 0").Prepare(),
			createAction:       acc.Insert("replies").Columns("tid, actionType, ipaddress, createdBy, createdAt, lastUpdated, content, parsed_content").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),'',''").Prepare(),

			getTopicUser: acc.SimpleLeftJoin("topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.lastReplyAt, topics.lastReplyBy, topics.lastReplyID, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.views, topics.postCount, topics.likeCount, topics.attachCount,topics.poll, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level", "topics.createdBy = users.uid", "tid = ?", "", ""),
			getByReplyID: acc.SimpleLeftJoin("replies", "topics", "topics.tid, topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.views, topics.postCount, topics.likeCount, topics.poll, topics.data", "replies.tid = topics.tid", "rid = ?", "", ""),
		}
		return acc.FirstError()
	})
}

// Flush the topic out of the cache
// ? - We do a CacheRemove() here instead of mutating the pointer to avoid creating a race condition
func (topic *Topic) cacheRemove() {
	tcache := Topics.GetCache()
	if tcache != nil {
		tcache.Remove(topic.ID)
	}
	TopicListThaw.Thaw()
}

// TODO: Write a test for this
func (topic *Topic) AddReply(rid int, uid int) (err error) {
	_, err = topicStmts.addReplies.Exec(1, uid, topic.ID)
	if err != nil {
		return err
	}
	_, err = topicStmts.updateLastReply.Exec(rid, rid, topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Lock() (err error) {
	_, err = topicStmts.lock.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Unlock() (err error) {
	_, err = topicStmts.unlock.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) MoveTo(destForum int) (err error) {
	_, err = topicStmts.moveTo.Exec(destForum, topic.ID)
	topic.cacheRemove()
	if err != nil {
		return err
	}
	err = Attachments.MoveTo(destForum, topic.ID, "topics")
	if err != nil {
		return err
	}
	return Attachments.MoveToByExtra(destForum, "replies", strconv.Itoa(topic.ID))
}

// TODO: We might want more consistent terminology rather than using stick in some places and pin in others. If you don't understand the difference, there is none, they are one and the same.
func (topic *Topic) Stick() (err error) {
	_, err = topicStmts.stick.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Unstick() (err error) {
	_, err = topicStmts.unstick.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

// TODO: Test this
// TODO: Use a transaction for this
func (topic *Topic) Like(score int, uid int) (err error) {
	var disp int // Unused
	err = topicStmts.hasLikedTopic.QueryRow(uid, topic.ID).Scan(&disp)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	_, err = topicStmts.createLike.Exec(score, topic.ID, "topics", uid)
	if err != nil {
		return err
	}

	_, err = topicStmts.addLikesToTopic.Exec(1, topic.ID)
	if err != nil {
		return err
	}
	_, err = userStmts.incrementLiked.Exec(1, uid)
	topic.cacheRemove()
	return err
}

// TODO: Implement this
func (topic *Topic) Unlike(uid int) error {
	topic.cacheRemove()
	return nil
}

// TODO: Use a transaction here
func (topic *Topic) Delete() error {
	topicCreator, err := Users.Get(topic.CreatedBy)
	if err == nil {
		wcount := WordCount(topic.Content)
		err = topicCreator.DecreasePostStats(wcount, true)
		if err != nil {
			return err
		}
	} else if err != ErrNoRows {
		return err
	}

	err = Forums.RemoveTopic(topic.ParentID)
	if err != nil && err != ErrNoRows {
		return err
	}

	_, err = topicStmts.delete.Exec(topic.ID)
	topic.cacheRemove()
	if err != nil {
		return err
	}
	_, err = topicStmts.deleteActivitySubs.Exec(topic.ID)
	if err != nil {
		return err
	}
	_, err = topicStmts.deleteActivity.Exec(topic.ID)
	return err
}

// TODO: Write tests for this
func (topic *Topic) Update(name string, content string) error {
	name = SanitiseSingleLine(html.UnescapeString(name))
	if name == "" {
		return ErrNoTitle
	}
	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(name) > Config.MaxTopicTitleLength {
		return ErrLongTitle
	}

	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, topic.ParentID, "forums")
	_, err := topicStmts.edit.Exec(name, content, parsedContent, topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) SetPoll(pollID int) error {
	_, err := topicStmts.setPoll.Exec(pollID, topic.ID) // TODO: Sniff if this changed anything to see if we hit an existing poll
	topic.cacheRemove()
	return err
}

// TODO: Have this go through the ReplyStore?
func (topic *Topic) CreateActionReply(action string, ipaddress string, uid int) (err error) {
	res, err := topicStmts.createAction.Exec(topic.ID, action, ipaddress, uid)
	if err != nil {
		return err
	}
	_, err = topicStmts.addReplies.Exec(1, uid, topic.ID)
	if err != nil {
		return err
	}
	lid, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rid := int(lid)
	_, err = topicStmts.updateLastReply.Exec(rid, rid, topic.ID)
	topic.cacheRemove()
	// ? - Update the last topic cache for the parent forum?
	return err
}

func GetRidsForTopic(tid int, offset int) (rids []int, err error) {
	rows, err := topicStmts.getRids.Query(tid, offset, Config.ItemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rid int
	for rows.Next() {
		err := rows.Scan(&rid)
		if err != nil {
			return nil, err
		}
		rids = append(rids, rid)
	}

	return rids, rows.Err()
}

func (ru *ReplyUser) Init() error {
	ru.UserLink = BuildProfileURL(NameToSlug(ru.CreatedByName), ru.CreatedBy)
	ru.ContentLines = strings.Count(ru.Content, "\n")

	postGroup, err := Groups.Get(ru.Group)
	if err != nil {
		return err
	}
	if postGroup.IsMod {
		ru.ClassName = Config.StaffCSS
	}

	// TODO: Make a function for this? Build a more sophisticated noavatar handling system? Do bulk user loads and let the c.UserStore initialise this?
	ru.Avatar, ru.MicroAvatar = BuildAvatar(ru.CreatedBy, ru.Avatar)
	if ru.Tag == "" {
		ru.Tag = postGroup.Tag
	}

	// We really shouldn't have inline HTML, we should do something about this...
	if ru.ActionType != "" {
		var action string
		aarr := strings.Split(ru.ActionType, "-")
		switch aarr[0] {
		case "lock":
			action = aarr[0]
			ru.ActionIcon = "&#x1F512;&#xFE0E"
		case "unlock":
			action = aarr[0]
			ru.ActionIcon = "&#x1F513;&#xFE0E"
		case "stick":
			action = aarr[0]
			ru.ActionIcon = "&#x1F4CC;&#xFE0E"
		case "unstick":
			action = aarr[0]
			ru.ActionIcon = "&#x1F4CC;&#xFE0E"
		case "move":
			if len(aarr) == 2 {
				fid, _ := strconv.Atoi(aarr[1])
				forum, err := Forums.Get(fid)
				if err == nil {
					ru.ActionType = p.GetTmplPhrasef("topic.action_topic_move_dest", forum.Link, forum.Name, ru.UserLink, ru.CreatedByName)
				} else {
					action = aarr[0]
				}
			} else {
				action = aarr[0]
			}
		default:
			// TODO: Only fire this off if a corresponding phrase for the ActionType doesn't exist? Or maybe have some sort of action registry?
			ru.ActionType = p.GetTmplPhrasef("topic.action_topic_default", ru.ActionType)
		}
		if action != "" {
			ru.ActionType = p.GetTmplPhrasef("topic.action_topic_"+action, ru.UserLink, ru.CreatedByName)
		}
	}

	return nil
}

// TODO: Factor TopicUser into a *Topic and *User, as this starting to become overly complicated x.x
func (topic *TopicUser) Replies(offset int, pFrag int, user *User) (rlist []*ReplyUser, ogdesc string, err error) {
	var likedMap map[int]int
	if user.Liked > 0 {
		likedMap = make(map[int]int)
	}
	var likedQueryList = []int{user.ID}

	var attachMap map[int]int
	if user.Perms.EditReply {
		attachMap = make(map[int]int)
	}
	var attachQueryList = []int{}

	var rid int
	if len(topic.Rids) > 0 {
		//log.Print("have rid")
		rid = topic.Rids[0]
	}
	re, err := Rstore.GetCache().Get(rid)
	ucache := Users.GetCache()
	var ruser *User
	if err == nil && ucache != nil {
		//log.Print("ucache step")
		ruser, err = ucache.Get(re.CreatedBy)
	}

	// TODO: Factor the user fields out and embed a user struct instead
	var reply *ReplyUser
	hTbl := GetHookTable()
	if err == nil {
		//log.Print("reply cached serve")
		reply = &ReplyUser{ClassName: "", Reply: *re, CreatedByName: ruser.Name, Avatar: ruser.Avatar, URLPrefix: ruser.URLPrefix, URLName: ruser.URLName, Level: ruser.Level}

		err := reply.Init()
		if err != nil {
			return nil, "", err
		}
		reply.ContentHtml = ParseMessage(reply.Content, topic.ParentID, "forums")
		// TODO: Do this more efficiently by avoiding the allocations entirely in ParseMessage, if there's nothing to do.
		if reply.ContentHtml == reply.Content {
			reply.ContentHtml = reply.Content
		}

		if reply.ID == pFrag {
			ogdesc = reply.Content
			if len(ogdesc) > 200 {
				ogdesc = ogdesc[:197] + "..."
			}
		}

		if reply.LikeCount > 0 && user.Liked > 0 {
			likedMap[reply.ID] = len(rlist)
			likedQueryList = append(likedQueryList, reply.ID)
		}
		if user.Perms.EditReply && reply.AttachCount > 0 {
			attachMap[reply.ID] = len(rlist)
			attachQueryList = append(attachQueryList, reply.ID)
		}

		hTbl.VhookNoRet("topic_reply_row_assign", &rlist, &reply)
		rlist = append(rlist, reply)
	} else {
		rows, err := topicStmts.getReplies.Query(topic.ID, offset, Config.ItemsPerPage)
		if err != nil {
			return nil, "", err
		}
		defer rows.Close()

		for rows.Next() {
			reply = &ReplyUser{}
			err := rows.Scan(&reply.ID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.Avatar, &reply.CreatedByName, &reply.Group, &reply.URLPrefix, &reply.URLName, &reply.Level, &reply.IPAddress, &reply.LikeCount, &reply.AttachCount, &reply.ActionType)
			if err != nil {
				return nil, "", err
			}

			err = reply.Init()
			if err != nil {
				return nil, "", err
			}
			reply.ContentHtml = ParseMessage(reply.Content, topic.ParentID, "forums")

			if reply.ID == pFrag {
				ogdesc = reply.Content
				if len(ogdesc) > 200 {
					ogdesc = ogdesc[:197] + "..."
				}
			}

			if reply.LikeCount > 0 && user.Liked > 0 {
				likedMap[reply.ID] = len(rlist)
				likedQueryList = append(likedQueryList, reply.ID)
			}
			if user.Perms.EditReply && reply.AttachCount > 0 {
				attachMap[reply.ID] = len(rlist)
				attachQueryList = append(attachQueryList, reply.ID)
			}

			hTbl.VhookNoRet("topic_reply_row_assign", &rlist, &reply)
			// TODO: Use a pointer instead to make it easier to abstract this loop? What impact would this have on escape analysis?
			rlist = append(rlist, reply)
			//log.Printf("r: %d-%d", reply.ID, len(rlist)-1)
		}
		err = rows.Err()
		if err != nil {
			return nil, "", err
		}
	}

	// TODO: Add a config setting to disable the liked query for a burst of extra speed
	if user.Liked > 0 && len(likedQueryList) > 1 /*&& user.LastLiked <= time.Now()*/ {
		eids, err := Likes.BulkExists(likedQueryList[1:], user.ID, "replies")
		if err != nil {
			return nil, "", err
		}
		for _, eid := range eids {
			rlist[likedMap[eid]].Liked = true
		}
	}

	if user.Perms.EditReply && len(attachQueryList) > 0 {
		//log.Printf("attachQueryList: %+v\n", attachQueryList)
		amap, err := Attachments.BulkMiniGetList("replies", attachQueryList)
		if err != nil && err != sql.ErrNoRows {
			return nil, "", err
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

	return rlist, ogdesc, nil
}

// TODO: Test this
func (topic *Topic) Author() (*User, error) {
	return Users.Get(topic.CreatedBy)
}

func (topic *Topic) GetID() int {
	return topic.ID
}
func (topic *Topic) GetTable() string {
	return "topics"
}

// Copy gives you a non-pointer concurrency safe copy of the topic
func (topic *Topic) Copy() Topic {
	return *topic
}

// TODO: Load LastReplyAt and LastReplyID?
func TopicByReplyID(rid int) (*Topic, error) {
	topic := Topic{ID: 0}
	err := topicStmts.getByReplyID.QueryRow(rid).Scan(&topic.ID, &topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.ViewCount, &topic.PostCount, &topic.LikeCount, &topic.Poll, &topic.Data)
	topic.Link = BuildTopicURL(NameToSlug(topic.Title), topic.ID)
	return &topic, err
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
	err = topicStmts.getTopicUser.QueryRow(tid).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.LastReplyAt, &tu.LastReplyBy, &tu.LastReplyID, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.ViewCount, &tu.PostCount, &tu.LikeCount, &tu.AttachCount, &tu.Poll, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
	tu.Avatar, tu.MicroAvatar = BuildAvatar(tu.CreatedBy, tu.Avatar)
	tu.Link = BuildTopicURL(NameToSlug(tu.Title), tu.ID)
	tu.UserLink = BuildProfileURL(NameToSlug(tu.CreatedByName), tu.CreatedBy)
	tu.Tag = Groups.DirtyGet(tu.Group).Tag

	if tcache != nil {
		theTopic := Topic{ID: tu.ID, Link: tu.Link, Title: tu.Title, Content: tu.Content, CreatedBy: tu.CreatedBy, IsClosed: tu.IsClosed, Sticky: tu.Sticky, CreatedAt: tu.CreatedAt, LastReplyAt: tu.LastReplyAt, LastReplyID: tu.LastReplyID, ParentID: tu.ParentID, IPAddress: tu.IPAddress, ViewCount: tu.ViewCount, PostCount: tu.PostCount, LikeCount: tu.LikeCount, AttachCount: tu.AttachCount, Poll: tu.Poll}
		//log.Printf("theTopic: %+v\n", theTopic)
		_ = tcache.Set(&theTopic)
	}
	return tu, err
}

func copyTopicToTopicUser(topic *Topic, user *User) (tu TopicUser) {
	tu.UserLink = user.Link
	tu.CreatedByName = user.Name
	tu.Group = user.Group
	tu.Avatar = user.Avatar
	tu.MicroAvatar = user.MicroAvatar
	tu.URLPrefix = user.URLPrefix
	tu.URLName = user.URLName
	tu.Level = user.Level

	tu.ID = topic.ID
	tu.Link = topic.Link
	tu.Title = topic.Title
	tu.Content = topic.Content
	tu.CreatedBy = topic.CreatedBy
	tu.IsClosed = topic.IsClosed
	tu.Sticky = topic.Sticky
	tu.CreatedAt = topic.CreatedAt
	tu.LastReplyAt = topic.LastReplyAt
	tu.LastReplyBy = topic.LastReplyBy
	tu.ParentID = topic.ParentID
	tu.IPAddress = topic.IPAddress
	tu.ViewCount = topic.ViewCount
	tu.PostCount = topic.PostCount
	tu.LikeCount = topic.LikeCount
	tu.AttachCount = topic.AttachCount
	tu.Poll = topic.Poll
	tu.Data = topic.Data
	tu.Rids = topic.Rids

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

// I don't care if it isn't used,, it will likely be in the future. Nolint.
// nolint
func getTopicURLPrefix() string {
	return "/topic/"
}
