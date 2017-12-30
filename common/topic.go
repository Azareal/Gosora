/*
*
*	Gosora Topic File
*	Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"database/sql"
	"html"
	"html/template"
	"strconv"
	"time"

	"../query_gen/lib"
)

// This is also in reply.go
//var ErrAlreadyLiked = errors.New("This item was already liked by this user")

// ? - Add a TopicMeta struct for *Forums?

type Topic struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           time.Time
	RelativeCreatedAt   string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	//LastReplyBy int
	ParentID  int
	Status    string // Deprecated. Marked for removal.
	IPAddress string
	PostCount int
	LikeCount int
	ClassName string // CSS Class Name
	Data      string // Used for report metadata
}

type TopicUser struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           time.Time
	RelativeCreatedAt   string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	//LastReplyBy int
	ParentID  int
	Status    string // Deprecated. Marked for removal.
	IPAddress string
	PostCount int
	LikeCount int
	ClassName string
	Data      string // Used for report metadata

	UserLink      string
	CreatedByName string
	Group         int
	Avatar        string
	ContentLines  int
	ContentHTML   string
	Tag           string
	URL           string
	URLPrefix     string
	URLName       string
	Level         int
	Liked         bool
}

type TopicsRow struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	LastReplyBy         int
	ParentID            int
	Status              string // Deprecated. Marked for removal. -Is there anything we could use it for?
	IPAddress           string
	PostCount           int
	LikeCount           int
	ClassName           string
	Data                string // Used for report metadata

	Creator      *User
	CSS          template.CSS
	ContentLines int
	LastUser     *User

	ForumName string //TopicsRow
	ForumLink string
}

type TopicStmts struct {
	addRepliesToTopic *sql.Stmt
	lock              *sql.Stmt
	unlock            *sql.Stmt
	stick             *sql.Stmt
	unstick           *sql.Stmt
	hasLikedTopic     *sql.Stmt
	createLike        *sql.Stmt
	addLikesToTopic   *sql.Stmt
	delete            *sql.Stmt
	edit              *sql.Stmt
	createActionReply *sql.Stmt

	getTopicUser *sql.Stmt // TODO: Can we get rid of this?
}

var topicStmts TopicStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		topicStmts = TopicStmts{
			addRepliesToTopic: acc.Update("topics").Set("postCount = postCount + ?, lastReplyBy = ?, lastReplyAt = UTC_TIMESTAMP()").Where("tid = ?").Prepare(),
			lock:              acc.Update("topics").Set("is_closed = 1").Where("tid = ?").Prepare(),
			unlock:            acc.Update("topics").Set("is_closed = 0").Where("tid = ?").Prepare(),
			stick:             acc.Update("topics").Set("sticky = 1").Where("tid = ?").Prepare(),
			unstick:           acc.Update("topics").Set("sticky = 0").Where("tid = ?").Prepare(),
			hasLikedTopic:     acc.Select("likes").Columns("targetItem").Where("sentBy = ? and targetItem = ? and targetType = 'topics'").Prepare(),
			createLike:        acc.Insert("likes").Columns("weight, targetItem, targetType, sentBy").Fields("?,?,?,?").Prepare(),
			addLikesToTopic:   acc.Update("topics").Set("likeCount = likeCount + ?").Where("tid = ?").Prepare(),
			delete:            acc.Delete("topics").Where("tid = ?").Prepare(),
			edit:              acc.Update("topics").Set("title = ?, content = ?, parsed_content = ?").Where("tid = ?").Prepare(),
			createActionReply: acc.Insert("replies").Columns("tid, actionType, ipaddress, createdBy, createdAt, lastUpdated, content, parsed_content").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),'',''").Prepare(),

			getTopicUser: acc.SimpleLeftJoin("topics", "users", "topics.title, topics.content, topics.createdBy, topics.createdAt, topics.is_closed, topics.sticky, topics.parentID, topics.ipaddress, topics.postCount, topics.likeCount, users.name, users.avatar, users.group, users.url_prefix, users.url_name, users.level", "topics.createdBy = users.uid", "tid = ?", "", ""),
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
}

// TODO: Write a test for this
func (topic *Topic) AddReply(uid int) (err error) {
	_, err = topicStmts.addRepliesToTopic.Exec(1, uid, topic.ID)
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
	var tid int // Unused
	err = topicStmts.hasLikedTopic.QueryRow(uid, topic.ID).Scan(&tid)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	_, err = topicStmts.createLike.Exec(score, tid, "topics", uid)
	if err != nil {
		return err
	}

	_, err = topicStmts.addLikesToTopic.Exec(1, tid)
	topic.cacheRemove()
	return err
}

// TODO: Implement this
func (topic *Topic) Unlike(uid int) error {
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
	return err
}

// TODO: Write tests for this
func (topic *Topic) Update(name string, content string) error {
	name = html.EscapeString(html.UnescapeString(name))
	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, topic.ParentID, "forums")
	_, err := topicStmts.edit.Exec(name, content, parsedContent, topic.ID)
	topic.cacheRemove()
	return err
}

// TODO: Have this go through the ReplyStore?
func (topic *Topic) CreateActionReply(action string, ipaddress string, user User) (err error) {
	_, err = topicStmts.createActionReply.Exec(topic.ID, action, ipaddress, user.ID)
	if err != nil {
		return err
	}
	_, err = topicStmts.addRepliesToTopic.Exec(1, user.ID, topic.ID)
	topic.cacheRemove()
	// ? - Update the last topic cache for the parent forum?
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the topic
func (topic *Topic) Copy() Topic {
	return *topic
}

// TODO: Refactor the caller to take a Topic and a User rather than a combined TopicUser
func GetTopicUser(tid int) (TopicUser, error) {
	tcache := Topics.GetCache()
	ucache := Users.GetCache()
	if tcache != nil && ucache != nil {
		topic, err := tcache.Get(tid)
		if err == nil {
			user, err := Users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}

			// We might be better off just passing separate topic and user structs to the caller?
			return copyTopicToTopicUser(topic, user), nil
		} else if ucache.Length() < ucache.GetCapacity() {
			topic, err = Topics.Get(tid)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			user, err := Users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			return copyTopicToTopicUser(topic, user), nil
		}
	}

	tu := TopicUser{ID: tid}
	err := topicStmts.getTopicUser.QueryRow(tid).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
	tu.Link = BuildTopicURL(NameToSlug(tu.Title), tu.ID)
	tu.UserLink = BuildProfileURL(NameToSlug(tu.CreatedByName), tu.CreatedBy)
	tu.Tag = Groups.DirtyGet(tu.Group).Tag

	if tcache != nil {
		theTopic := Topic{ID: tu.ID, Link: tu.Link, Title: tu.Title, Content: tu.Content, CreatedBy: tu.CreatedBy, IsClosed: tu.IsClosed, Sticky: tu.Sticky, CreatedAt: tu.CreatedAt, LastReplyAt: tu.LastReplyAt, ParentID: tu.ParentID, IPAddress: tu.IPAddress, PostCount: tu.PostCount, LikeCount: tu.LikeCount}
		//log.Printf("theTopic: %+v\n", theTopic)
		_ = tcache.Add(&theTopic)
	}
	return tu, err
}

func copyTopicToTopicUser(topic *Topic, user *User) (tu TopicUser) {
	tu.UserLink = user.Link
	tu.CreatedByName = user.Name
	tu.Group = user.Group
	tu.Avatar = user.Avatar
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
	tu.ParentID = topic.ParentID
	tu.IPAddress = topic.IPAddress
	tu.PostCount = topic.PostCount
	tu.LikeCount = topic.LikeCount
	tu.Data = topic.Data

	return tu
}

// For use in tests and for generating blank topics for forums which don't have a last poster
func BlankTopic() *Topic {
	return &Topic{ID: 0, Title: ""}
}

func BuildTopicURL(slug string, tid int) string {
	if slug == "" {
		return "/topic/" + strconv.Itoa(tid)
	}
	return "/topic/" + slug + "." + strconv.Itoa(tid)
}

// I don't care if it isn't used,, it will likely be in the future. Nolint.
// nolint
func getTopicURLPrefix() string {
	return "/topic/"
}
