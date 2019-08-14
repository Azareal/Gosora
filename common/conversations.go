package common

import (
	"errors"
	"time"

	//"strconv"
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

/*
conversations
conversations_posts
*/
var Convos ConversationStore
var convoStmts ConvoStmts

type ConvoStmts struct {
	fetchPost  *sql.Stmt
	getPosts   *sql.Stmt
	countPosts *sql.Stmt
	edit       *sql.Stmt
	create     *sql.Stmt
	delete     *sql.Stmt

	editPost   *sql.Stmt
	createPost *sql.Stmt
	deletePost *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		convoStmts = ConvoStmts{
			fetchPost:  acc.Select("conversations_posts").Columns("cid, body, post, createdBy").Where("pid = ?").Prepare(),
			getPosts:   acc.Select("conversations_posts").Columns("pid, body, post, createdBy").Where("cid = ?").Limit("?,?").Prepare(),
			countPosts: acc.Count("conversations_posts").Where("cid = ?").Prepare(),
			edit:       acc.Update("conversations").Set("lastReplyBy = ?, lastReplyAt = ?").Where("cid = ?").Prepare(),
			create:     acc.Insert("conversations").Columns("createdAt, lastReplyAt").Fields("UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),

			editPost:   acc.Update("conversations_posts").Set("body = ?, post = ?").Where("cid = ?").Prepare(),
			createPost: acc.Insert("conversations_posts").Columns("cid, body, post, createdBy").Fields("?,?,?,?").Prepare(),
			deletePost: acc.Delete("conversations_posts").Where("pid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

type Conversation struct {
	ID          int
	CreatedBy   int
	CreatedAt   time.Time
	LastReplyBy int
	LastReplyAt time.Time
}

func (co *Conversation) Posts(offset, itemsPerPage int) (posts []*ConversationPost, err error) {
	rows, err := convoStmts.getPosts.Query(co.ID, offset, itemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &ConversationPost{CID: co.ID}
		err := rows.Scan(&p.ID, &p.Body, &p.Post, &p.CreatedBy)
		if err != nil {
			return nil, err
		}
		p, err = ConvoPostProcess.OnLoad(p)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (co *Conversation) PostsCount() (count int) {
	err := convoStmts.countPosts.QueryRow(co.ID).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (co *Conversation) Update() error {
	_, err := convoStmts.edit.Exec(co.CreatedAt, co.LastReplyBy, co.LastReplyAt, co.ID)
	return err
}

func (co *Conversation) Create() (int, error) {
	res, err := convoStmts.create.Exec()
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

type ConversationStore interface {
	Get(id int) (*Conversation, error)
	GetUser(uid int, offset int) (cos []*Conversation, err error)
	GetUserCount(uid int) (count int)
	Delete(id int) error
	Count() (count int)
	Create(content string, createdBy int, participants []int) (int, error)
}

type DefaultConversationStore struct {
	get            *sql.Stmt
	getUser        *sql.Stmt
	getUserCount   *sql.Stmt
	delete         *sql.Stmt
	deletePosts    *sql.Stmt
	deleteParticipants *sql.Stmt
	create         *sql.Stmt
	addParticipant *sql.Stmt
	count          *sql.Stmt
}

func NewDefaultConversationStore(acc *qgen.Accumulator) (*DefaultConversationStore, error) {
	return &DefaultConversationStore{
		get:            acc.Select("conversations").Columns("createdBy, createdAt, lastReplyBy, lastReplyAt").Where("cid = ?").Prepare(),
		getUser:        acc.SimpleInnerJoin("conversations_participants AS cp", "conversations AS c", "cp.cid, c.createdBy, c.createdAt, c.lastReplyBy, c.lastReplyAt", "cp.cid = c.cid", "cp.uid = ?", "c.lastReplyAt DESC, c.createdAt DESC, c.cid DESC", "?,?"),
		getUserCount:   acc.Count("conversations_participants").Where("uid = ?").Prepare(),
		delete:         acc.Delete("conversations").Where("cid = ?").Prepare(),
		deletePosts:    acc.Delete("conversations_posts").Where("cid = ?").Prepare(),
		deleteParticipants:    acc.Delete("conversations_participants").Where("cid = ?").Prepare(),
		create:         acc.Insert("conversations").Columns("createdBy, createdAt, lastReplyAt").Fields("?,UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),
		addParticipant: acc.Insert("conversations_participants").Columns("uid, cid").Fields("?,?").Prepare(),
		count:          acc.Count("conversations").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultConversationStore) Get(id int) (*Conversation, error) {
	convo := &Conversation{ID: id}
	err := s.get.QueryRow(id).Scan(&convo.CreatedBy, &convo.CreatedAt, &convo.LastReplyBy, &convo.LastReplyAt)
	return convo, err
}

func (s *DefaultConversationStore) GetUser(uid int, offset int) (cos []*Conversation, err error) {
	rows, err := s.getUser.Query(uid, offset, Config.ItemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		co := &Conversation{}
		err := rows.Scan(&co.ID, &co.CreatedBy, &co.CreatedAt, &co.LastReplyBy, &co.LastReplyAt)
		if err != nil {
			return nil, err
		}
		cos = append(cos, co)
	}

	return cos, rows.Err()
}

func (s *DefaultConversationStore) GetUserCount(uid int) (count int) {
	err := s.getUserCount.QueryRow(uid).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

// TODO: Use a foreign key or transaction
func (s *DefaultConversationStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	if err != nil {
		return err
	}
	_, err = s.deletePosts.Exec(id)
	if err != nil {
		return err
	}
	_, err = s.deleteParticipants.Exec(id)
	return err
}

func (s *DefaultConversationStore) Create(content string, createdBy int, participants []int) (int, error) {
	if len(participants) == 0 {
		return 0, errors.New("no participants set")
	}
	res, err := s.create.Exec(createdBy)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	post := &ConversationPost{}
	post.CID = int(lastID)
	post.Body = content
	post.CreatedBy = createdBy
	_, err = post.Create()
	if err != nil {
		return 0, err
	}

	for _, p := range participants {
		_, err := s.addParticipant.Exec(p, lastID)
		if err != nil {
			return 0, err
		}
	}
	_, err = s.addParticipant.Exec(createdBy, lastID)
	if err != nil {
		return 0, err
	}

	return int(lastID), err
}

// Count returns the total number of topics on these forums
func (s *DefaultConversationStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}
