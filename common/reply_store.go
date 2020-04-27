package common

//import "log"
import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Rstore ReplyStore

type ReplyStore interface {
	Get(id int) (*Reply, error)
	Each(f func(*Reply) error) error
	Exists(id int) bool
	Create(t *Topic, content, ip string, uid int) (id int, err error)
	Count() (count int)
	CountUser(uid int) (count int)
	CountMegaUser(uid int) (count int)
	CountBigUser(uid int) (count int)

	SetCache(cache ReplyCache)
	GetCache() ReplyCache
}

type SQLReplyStore struct {
	cache ReplyCache

	get           *sql.Stmt
	getAll        *sql.Stmt
	exists        *sql.Stmt
	create        *sql.Stmt
	count         *sql.Stmt
	countUser     *sql.Stmt
	countWordUser *sql.Stmt
}

func NewSQLReplyStore(acc *qgen.Accumulator, cache ReplyCache) (*SQLReplyStore, error) {
	if cache == nil {
		cache = NewNullReplyCache()
	}
	re := "replies"
	return &SQLReplyStore{
		cache:         cache,
		get:           acc.Select(re).Columns("tid,content,createdBy,createdAt,lastEdit,lastEditBy,ip,likeCount,attachCount,actionType").Where("rid=?").Prepare(),
		getAll:        acc.Select(re).Columns("rid,tid,content,createdBy,createdAt,lastEdit,lastEditBy,ip,likeCount,attachCount,actionType").Prepare(),
		exists:        acc.Exists(re, "rid").Prepare(),
		create:        acc.Insert(re).Columns("tid,content,parsed_content,createdAt,lastUpdated,ip,words,createdBy").Fields("?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?").Prepare(),
		count:         acc.Count(re).Prepare(),
		countUser:     acc.Count(re).Where("createdBy=?").Prepare(),
		countWordUser: acc.Count(re).Where("createdBy=? AND words>=?").Prepare(),
	}, acc.FirstError()
}

func (s *SQLReplyStore) Get(id int) (*Reply, error) {
	r, err := s.cache.Get(id)
	if err == nil {
		return r, nil
	}

	r = &Reply{ID: id}
	err = s.get.QueryRow(id).Scan(&r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IP, &r.LikeCount, &r.AttachCount, &r.ActionType)
	if err == nil {
		_ = s.cache.Set(r)
	}
	return r, err
}

/*func (s *SQLReplyStore) eachr(f func(*sql.Rows) error) error {
	return eachall(s.getAll, f)
}*/

func (s *SQLReplyStore) Each(f func(*Reply) error) error {
	rows, err := s.getAll.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		r := new(Reply)
		if err := rows.Scan(&r.ID, &r.ParentID, &r.Content, &r.CreatedBy, &r.CreatedAt, &r.LastEdit, &r.LastEditBy, &r.IP, &r.LikeCount, &r.AttachCount, &r.ActionType); err != nil {
			return err
		}
		if err := f(r); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *SQLReplyStore) Exists(id int) bool {
	err := s.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

// TODO: Write a test for this
func (s *SQLReplyStore) Create(t *Topic, content, ip string, uid int) (id int, err error) {
	if Config.DisablePostIP {
		ip = ""
	}
	res, err := s.create.Exec(t.ID, content, ParseMessage(content, t.ParentID, "forums", nil, nil), ip, WordCount(content), uid)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	id = int(lastID)
	return id, t.AddReply(id, uid)
}

// TODO: Write a test for this
// Count returns the total number of topic replies on these forums
func (s *SQLReplyStore) Count() (count int) {
	return Countf(s.count)
}
func (s *SQLReplyStore) CountUser(uid int) (count int) {
	return Countf(s.countUser, uid)
}
func (s *SQLReplyStore) CountMegaUser(uid int) (count int) {
	return Countf(s.countWordUser, uid, SettingBox.Load().(SettingMap)["megapost_min_words"].(int))
}
func (s *SQLReplyStore) CountBigUser(uid int) (count int) {
	return Countf(s.countWordUser, uid, SettingBox.Load().(SettingMap)["bigpost_min_words"].(int))
}

func (s *SQLReplyStore) SetCache(cache ReplyCache) {
	s.cache = cache
}

func (s *SQLReplyStore) GetCache() ReplyCache {
	return s.cache
}
