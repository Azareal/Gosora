package common

//import "log"
import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

var Rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Create(topic *Topic, content string, ipaddress string, uid int) (id int, err error)

	SetCache(cache ReplyCache)
	GetCache() ReplyCache
}

type SQLReplyStore struct {
	cache ReplyCache

	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLReplyStore(acc *qgen.Accumulator, cache ReplyCache) (*SQLReplyStore, error) {
	if cache == nil {
		cache = NewNullReplyCache()
	}
	return &SQLReplyStore{
		cache:  cache,
		get:    acc.Select("replies").Columns("tid, content, createdBy, createdAt, lastEdit, lastEditBy, ipaddress, likeCount, attachCount, actionType").Where("rid = ?").Prepare(),
		create: acc.Insert("replies").Columns("tid, content, parsed_content, createdAt, lastUpdated, ipaddress, words, createdBy").Fields("?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?").Prepare(),
	}, acc.FirstError()
}

func (s *SQLReplyStore) Get(id int) (*Reply, error) {
	//log.Print("SQLReplyStore.Get")
	reply, err := s.cache.Get(id)
	if err == nil {
		return reply, nil
	}

	reply = &Reply{ID: id}
	err = s.get.QueryRow(id).Scan(&reply.ParentID, &reply.Content, &reply.CreatedBy, &reply.CreatedAt, &reply.LastEdit, &reply.LastEditBy, &reply.IPAddress, &reply.LikeCount, &reply.AttachCount, &reply.ActionType)
	if err == nil {
		_ = s.cache.Set(reply)
	}
	return reply, err
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

func (s *SQLReplyStore) SetCache(cache ReplyCache) {
	s.cache = cache
}

func (s *SQLReplyStore) GetCache() ReplyCache {
	return s.cache
}
