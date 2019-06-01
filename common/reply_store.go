package common

//import "log"
import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error)
	Count() (count int)

	SetCache(cache ReplyCache)
	GetCache() ReplyCache
}

type SQLReplyStore struct {
	cache ReplyCache

	get    *sql.Stmt
	create *sql.Stmt
	count *sql.Stmt
}

func NewSQLReplyStore(acc *qgen.Accumulator, cache ReplyCache) (*SQLReplyStore, error) {
	if cache == nil {
		cache = NewNullReplyCache()
	}
	return &SQLReplyStore{
		cache:  cache,
		get:    acc.Select("replies").Columns("tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount, attachCount, actionType").Where("rid = ?").Prepare(),
		create: acc.Insert("replies").Columns("tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy").Fields("?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?").Prepare(),
		count: acc.Count("replies").Prepare(),
	}, acc.FirstError()
}

func (s *SQLReplyStore) Get(id int) (*Reply, error) {
	r, err := s.cache.Get(id)
	if err == nil {
		return r, nil
	}

	r = &Reply{ID: id}
	err = s.get.QueryRow(id).Scan(&r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IPAddress, &r.LikeCount, &r.AttachCount, &r.ActionType)
	if err == nil {
		_ = s.cache.Set(r)
	}
	return r, err
}

// TODO: Write a test for this
func (s *SQLReplyStore) Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error) {
	wcount := WordCount(content)
	res, err := s.create.Exec(topic.ID, content, ParseMessage(content, topic.ParentID, "forums"), ipaddress, wcount, uid)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), topic.AddReply(int(lastID), uid)
}

// TODO: Write a test for this
// Count returns the total number of topic replies on these forums
func (s *SQLReplyStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *SQLReplyStore) SetCache(cache ReplyCache) {
	s.cache = cache
}

func (s *SQLReplyStore) GetCache() ReplyCache {
	return s.cache
}
