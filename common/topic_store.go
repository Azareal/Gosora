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

	ClearIPs() error
	LockMany(tids []int) error

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

	clearIPs *sql.Stmt
	lockTen *sql.Stmt
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
		get:           acc.Select(t).Columns("title,content,createdBy,createdAt,lastReplyBy,lastReplyAt,lastReplyID,is_closed,sticky,parentID,ip,views,postCount,likeCount,attachCount,poll,data").Where("tid=?").Stmt(),
		exists:        acc.Exists(t, "tid").Stmt(),
		count:         acc.Count(t).Stmt(),
		countUser:     acc.Count(t).Where("createdBy=?").Stmt(),
		countWordUser: acc.Count(t).Where("createdBy=? AND words>=?").Stmt(),
		create:        acc.Insert(t).Columns("parentID,title,content,parsed_content,createdAt,lastReplyAt,lastReplyBy,ip,words,createdBy").Fields("?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,?").Prepare(),

		clearIPs: acc.Update(t).Set("ip=''").Where("ip!=''").Stmt(),
		lockTen: acc.Update(t).Set("is_closed=1").Where("tid IN(" + inqbuild2(10) + ")").Stmt(),
	}, acc.FirstError()
}

func (s *DefaultTopicStore) DirtyGet(id int) *Topic {
	t, e := s.cache.Get(id)
	if e == nil {
		return t
	}
	t, e = s.BypassGet(id)
	if e == nil {
		_ = s.cache.Set(t)
		return t
	}
	return BlankTopic()
}

// TODO: Log weird cache errors?
func (s *DefaultTopicStore) Get(id int) (t *Topic, e error) {
	t, e = s.cache.Get(id)
	if e == nil {
		return t, nil
	}
	t, e = s.BypassGet(id)
	if e == nil {
		_ = s.cache.Set(t)
	}
	return t, e
}

// BypassGet will always bypass the cache and pull the topic directly from the database
func (s *DefaultTopicStore) BypassGet(id int) (*Topic, error) {
	t := &Topic{ID: id}
	e := s.get.QueryRow(id).Scan(&t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.LastReplyBy, &t.LastReplyAt, &t.LastReplyID, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
	if e == nil {
		t.Link = BuildTopicURL(NameToSlug(t.Title), id)
	}
	return t, e
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
func (s *DefaultTopicStore) BulkGetMap(ids []int) (list map[int]*Topic, e error) {
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
		t, e := s.Get(ids[0])
		if e != nil {
			return list, e
		}
		list[t.ID] = t
		return list, nil
	}

	idList, q := inqbuild(ids)
	rows, e := qgen.NewAcc().Select("topics").Columns("tid,title,content,createdBy,createdAt,lastReplyBy,lastReplyAt,lastReplyID,is_closed,sticky,parentID,ip,views,postCount,likeCount,attachCount,poll,data").Where("tid IN(" + q + ")").Query(idList...)
	if e != nil {
		return list, e
	}
	defer rows.Close()

	for rows.Next() {
		t := &Topic{}
		e := rows.Scan(&t.ID, &t.Title, &t.Content, &t.CreatedBy, &t.CreatedAt, &t.LastReplyBy, &t.LastReplyAt, &t.LastReplyID, &t.IsClosed, &t.Sticky, &t.ParentID, &t.IP, &t.ViewCount, &t.PostCount, &t.LikeCount, &t.AttachCount, &t.Poll, &t.Data)
		if e != nil {
			return list, e
		}
		t.Link = BuildTopicURL(NameToSlug(t.Title), t.ID)
		_ = s.cache.Set(t)
		list[t.ID] = t
	}
	if e = rows.Err(); e != nil {
		return list, e
	}

	// Did we miss any topics?
	if idCount > len(list) {
		var sidList string
		for i, id := range ids {
			if _, ok := list[id]; !ok {
				if i == 0 {
					sidList += strconv.Itoa(id)
				} else {
					sidList += ","+strconv.Itoa(id)
				}
			}
		}
		if sidList != "" {
			e = errors.New("Unable to find topics with the following IDs: " + sidList)
		}
	}

	return list, e
}

func (s *DefaultTopicStore) Reload(id int) error {
	t, e := s.BypassGet(id)
	if e == nil {
		_ = s.cache.Set(t)
	} else {
		_ = s.cache.Remove(id)
	}
	TopicListThaw.Thaw()
	return e
}

func (s *DefaultTopicStore) Exists(id int) bool {
	return s.exists.QueryRow(id).Scan(&id) == nil
}

func (s *DefaultTopicStore) ClearIPs() error {
	_, e := s.clearIPs.Exec()
	return e
}

func (s *DefaultTopicStore) LockMany(tids []int) (e error) {
	tc, i := Topics.GetCache(), 0
	singles := func() error {
		for ; i < len(tids); i++ {
			_, e := topicStmts.lock.Exec(tids[i])
			if e != nil {
				return e
			}
		}
		return nil
	}

	if len(tids) < 10 {
		if e = singles(); e != nil {
			return e
		}
		if tc != nil {
			_ = tc.RemoveMany(tids)
		}
		TopicListThaw.Thaw()
		return nil
	}

	for ; (i + 10) < len(tids); i += 10 {
		_, e := s.lockTen.Exec(tids[i], tids[i+1], tids[i+2], tids[i+3], tids[i+4], tids[i+5], tids[i+6], tids[i+7], tids[i+8], tids[i+9])
		if e != nil {
			return e
		}
	}

	if e = singles(); e != nil {
		return e
	}
	if tc != nil {
		_ = tc.RemoveMany(tids)
	}
	TopicListThaw.Thaw()
	return nil
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
	//TopicListThaw.Thaw() // redundant

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
