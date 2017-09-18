/*
*
*	Gosora Topic File
*	Copyright Azareal 2017 - 2018
*
 */
package main

//import "fmt"
import "strconv"
import "html/template"

type Topic struct {
	ID          int
	Link        string
	Title       string
	Content     string
	CreatedBy   int
	IsClosed    bool
	Sticky      bool
	CreatedAt   string
	LastReplyAt string
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
	ID          int
	Link        string
	Title       string
	Content     string
	CreatedBy   int
	IsClosed    bool
	Sticky      bool
	CreatedAt   string
	LastReplyAt string
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
	Tag           string
	URL           string
	URLPrefix     string
	URLName       string
	Level         int
	Liked         bool
}

type TopicsRow struct {
	ID          int
	Link        string
	Title       string
	Content     string
	CreatedBy   int
	IsClosed    bool
	Sticky      bool
	CreatedAt   string
	LastReplyAt string
	LastReplyBy int
	ParentID    int
	Status      string // Deprecated. Marked for removal. -Is there anything we could use it for?
	IPAddress   string
	PostCount   int
	LikeCount   int
	ClassName   string
	Data        string // Used for report metadata

	Creator      *User
	CSS          template.CSS
	ContentLines int
	LastUser     *User

	ForumName string //TopicsRow
	ForumLink string
}

// TODO: Refactor the caller to take a Topic and a User rather than a combined TopicUser
func getTopicuser(tid int) (TopicUser, error) {
	tcache, tok := topics.(TopicCache)
	ucache, uok := users.(UserCache)
	if tok && uok {
		topic, err := tcache.CacheGet(tid)
		if err == nil {
			user, err := users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}

			// We might be better off just passing seperate topic and user structs to the caller?
			return copyTopicToTopicuser(topic, user), nil
		} else if ucache.GetLength() < ucache.GetCapacity() {
			topic, err = topics.Get(tid)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			user, err := users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			return copyTopicToTopicuser(topic, user), nil
		}
	}

	tu := TopicUser{ID: tid}
	err := getTopicUserStmt.QueryRow(tid).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
	tu.Link = buildTopicURL(nameToSlug(tu.Title), tu.ID)
	tu.UserLink = buildProfileURL(nameToSlug(tu.CreatedByName), tu.CreatedBy)
	tu.Tag = gstore.DirtyGet(tu.Group).Tag

	if tok {
		theTopic := Topic{ID: tu.ID, Link: tu.Link, Title: tu.Title, Content: tu.Content, CreatedBy: tu.CreatedBy, IsClosed: tu.IsClosed, Sticky: tu.Sticky, CreatedAt: tu.CreatedAt, LastReplyAt: tu.LastReplyAt, ParentID: tu.ParentID, IPAddress: tu.IPAddress, PostCount: tu.PostCount, LikeCount: tu.LikeCount}
		//log.Printf("the_topic: %+v\n", theTopic)
		_ = tcache.CacheAdd(&theTopic)
	}
	return tu, err
}

func copyTopicToTopicuser(topic *Topic, user *User) (tu TopicUser) {
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

func getTopicByReply(rid int) (*Topic, error) {
	topic := Topic{ID: 0}
	err := getTopicByReplyStmt.QueryRow(rid).Scan(&topic.ID, &topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), topic.ID)
	return &topic, err
}

func buildTopicURL(slug string, tid int) string {
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
