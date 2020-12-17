/*
*
*	Gosora Topic Store
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	qgen "github.com/Azareal/Gosora/query_gen"
)

// TODO: Add the watchdog goroutine
// TODO: Add some sort of update method
// ? - Should we add stick, lock, unstick, and unlock methods? These might be better on the Topics not the TopicStore
var Topics TopicStore
var ErrNoTitle = errors.New("This message is missing a title")
var ErrLongTitle = errors.New("The title is too long")
var ErrNoBody = errors.New("This message is missing a body")

type TopicStore interface {
	DirtyGet(id int) *Topic
	Get(id int) (*Topic, error)
	BypassGet(id int) (*Topic, error)
	BulkGetMap(ids []int) (list map[int]*Topic, err error)
	Exists(id int) bool
	Create(fid int, name, content string, uid int, ip string) (tid int, err error)
	AddLastTopic(t *Topic, fid int) error // unimplemented
	Reload(id int) error                  // Too much SQL logic to move into TopicCache
	// TODO: Implement these two methods
	//Replies(tid int) ([]*Reply, error)
	//RepliesRange(tid, lower, higher int) ([]*Reply, error)
	Count() int
	CountUser(uid int) int
	CountMegaUser(uid int) int
	CountBigUser(uid int) int

	SetCache(cache TopicCache)
	GetCache() TopicCache
}

type DefaultTopicStore struct {
	cache TopicCache

	get           *sql.Stmt
	exists        *sql.Stmt
	count         *sql.Stmt
	countUser     *sql.Stmt
	countWordUser *sql.Stmt
	create        *sql.Stmt
}

// NewDefaultTopicStore gives you a new instance of DefaultTopicStore
func NewDefaultTopicStore(cache TopicCache) (*DefaultTopicStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullTopicCache()
	}
	t := "topics"
	return &DefaultTopicStore{
		cache:         cache,
		get:           acc.Select(t).Columns("title,content,createdBy,createdAt,lastReplyBy,lastReplyAt,lastReplyID,is_closed,sticky,parentID,ip,views,postCount,likeCount,attachCount,poll,data").Where("tid=?").Prepare(),
		exists:        acc.Exists(t, "tid").Prepare(),
		count:         acc.Count(t).Prepare(),
		countUser:     acc.Count(t).Where("createdBy=?").Prepare(),
		countWordUser: acc.Count(t).Where("createdBy=? AND words>=?").Prepare(),
		create:        acc.Insert(t).Columns("parentID, title, content, parsed_content, createdAt, lastReplyAt, lastReplyBy, ip, words, createdBy").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultTopicStore) DirtyGet(id int) *Topic {
	t, err := s.cache.Get(id)
	if err == nil {
		return t
	}
	t, err = s.BypassGet(id)
	if err == nil {
		_ = s.cache.Set(t)
		return t
	}
	return BlankTopic()
}

// TODO: Log weird cache errors?
func (s *DefaultTopicStore) Get(id int) (t *Topic, err error) {
	t, err = s.cache.Get(id)
	if err == nil {
		return t, nil
	}
	t, err = s.BypassGet(id)
	if err == nil {
		_ = s.cache.Set(t)
	}
	return t, err
}

// BypassGet will always bypass the cache and pull the topic directly from the database
func (s *DefaultTopicStore) BypassGet(id int) (*Topic, error) {
	t := &Topic{ID: id}
	err := s.get.QueryRow(id).Scan(&t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.LastReplyBy, &t.LastReplyAt, &t.LastReplyID, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
	if err == nil {
		t.Link = BuildTopicURL(NameToSlug(t.Title), id)
	}
	return t, err
}

/*func (s *DefaultTopicStore) GetByUser(uid int) (list map[int]*Topic, err error) {
	t := &Topic{ID: id}
	err := s.get.QueryRow(id).Scan(&t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.LastReplyBy, &t.LastReplyAt, &t.LastReplyID, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
	if err == nil {
		t.Link = BuildTopicURL(NameToSlug(t.Title), id)
	}
	return t, err
}*/

// TODO: Avoid duplicating much of this logic from user_store.go
func (s *DefaultTopicStore) BulkGetMap(ids []int) (list map[int]*Topic, err error) {
	idCount := len(ids)
	list = make(map[int]*Topic)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := s.cache.BulkGet(ids)
	if len(sliceList) > 0 {
		for i, sliceItem := range sliceList {
			if sliceItem != nil {
				list[sliceItem.ID] = sliceItem
			} else {
				stillHere = append(stillHere, ids[i])
			}
		}
		ids = stillHere
	}

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	} else if len(ids) == 1 {
		topic, err := s.Get(ids[0])
		if err != nil {
			return list, err
		}
		list[topic.ID] = topic
		return list, nil
	}

	// TODO: Add a function for the qlist stuff
	var q string
	idList := make([]interface{}, len(ids))
	for i, id := range ids {
		idList[i] = strconv.Itoa(id)
		q += "?,"
	}
	q = q[0 : len(q)-1]

	rows, err := qgen.NewAcc().Select("topics").Columns("tid,title,content,createdBy,createdAt,lastReplyBy,lastReplyAt,lastReplyID,is_closed,sticky,parentID,ip,views,postCount,likeCount,attachCount,poll,data").Where("tid IN(" + q + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		t := &Topic{}
		err := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.LastReplyBy, &t.LastReplyAt, &t.LastReplyID, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
		if err != nil {
			return list, err
		}
		t.Link = BuildTopicURL(NameToSlug(t.Title), t.ID)
		s.cache.Set(t)
		list[t.ID] = t
	}
	if err = rows.Err(); err != nil {
		return list, err
	}

	// Did we miss any topics?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			_, ok := list[id]
			if !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}
		if sidList != "" {
			sidList = sidList[0 : len(sidList)-1]
			err = errors.New("Unable to find topics with the following IDs: " + sidList)
		}
	}

	return list, err
}

func (s *DefaultTopicStore) Reload(id int) error {
	topic, err := s.BypassGet(id)
	if err == nil {
		_ = s.cache.Set(topic)
	} else {
		_ = s.cache.Remove(id)
	}
	TopicListThaw.Thaw()
	return err
}

func (s *DefaultTopicStore) Exists(id int) bool {
	return s.exists.QueryRow(id).Scan(&id) == nil
}

func (s *DefaultTopicStore) Create(fid int, name, content string, uid int, ip string) (tid int, err error) {
	if name == "" {
		return 0, ErrNoTitle
	}
	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(name) > Config.MaxTopicTitleLength {
		return 0, ErrLongTitle
	}

	parsedContent := strings.TrimSpace(ParseMessage(content, fid, "forums", nil, nil))
	if parsedContent == "" {
		return 0, ErrNoBody
	}

	// TODO: Move this statement into the topic store
	if Config.DisablePostIP {
		ip = ""
	}
	res, err := s.create.Exec(fid, name, content, parsedContent, uid, ip, WordCount(content), uid)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	tid = int(lastID)
	TopicListThaw.Thaw()

	return tid, Forums.AddTopic(tid, uid, fid)
}

// ? - What is this? Do we need it? Should it be in the main store interface?
func (s *DefaultTopicStore) AddLastTopic(t *Topic, fid int) error {
	// Coming Soon...
	return nil
}

// Count returns the total number of topics on these forums
func (s *DefaultTopicStore) Count() (count int) {
	return Countf(s.count)
}
func (s *DefaultTopicStore) CountUser(uid int) (count int) {
	return Countf(s.countUser, uid)
}
func (s *DefaultTopicStore) CountMegaUser(uid int) (count int) {
	return Countf(s.countWordUser, uid, SettingBox.Load().(SettingMap)["megapost_min_words"].(int))
}
func (s *DefaultTopicStore) CountBigUser(uid int) (count int) {
	return Countf(s.countWordUser, uid, SettingBox.Load().(SettingMap)["bigpost_min_words"].(int))
}

func (s *DefaultTopicStore) SetCache(cache TopicCache) {
	s.cache = cache
}

// TODO: We're temporarily doing this so that you can do tcache != nil in getTopicUser. Refactor it.
func (s *DefaultTopicStore) GetCache() TopicCache {
	_, ok := s.cache.(*NullTopicCache)
	if ok {
		return nil
	}
	return s.cache
}
