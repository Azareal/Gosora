package common

import (
	"io"
	"time"
	"database/sql"
	"encoding/hex"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	qgen "github.com/Azareal/Gosora/query_gen"
)

/*
conversations
conversations_posts
*/

var ConvoPostProcess ConvoPostProcessor = NewDefaultConvoPostProcessor()

type ConvoPostProcessor interface {
	OnLoad(co *ConversationPost) (*ConversationPost, error)
	OnSave(co *ConversationPost) (*ConversationPost, error)
}

type DefaultConvoPostProcessor struct {
}

func NewDefaultConvoPostProcessor() *DefaultConvoPostProcessor {
	return &DefaultConvoPostProcessor{}
}

func (pr *DefaultConvoPostProcessor) OnLoad(co *ConversationPost) (*ConversationPost, error) {
	return co, nil
}

func (pr *DefaultConvoPostProcessor) OnSave(co *ConversationPost) (*ConversationPost, error) {
	return co, nil
}

type AesConvoPostProcessor struct {
}

func NewAesConvoPostProcessor() *AesConvoPostProcessor {
	return &AesConvoPostProcessor{}
}

func (pr *AesConvoPostProcessor) OnLoad(co *ConversationPost) (*ConversationPost, error) {
	if co.Post != "aes" {
		return co, nil
	}
	key, _ := hex.DecodeString(Config.ConvoKey)

	ciphertext, err := hex.DecodeString(co.Body)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	lco := *co
	lco.Body = string(plaintext)
	return &lco, nil
}

func (pr *AesConvoPostProcessor) OnSave(co *ConversationPost) (*ConversationPost, error) {
	key, _ := hex.DecodeString(Config.ConvoKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(co.Body), nil)

	lco := *co
	lco.Body = hex.EncodeToString(ciphertext)
	lco.Post = "aes"
	return &lco, nil
}

var convoStmts ConvoStmts

type ConvoStmts struct {
	getPosts *sql.Stmt
	edit   *sql.Stmt
	create *sql.Stmt

	editPost *sql.Stmt
	createPost *sql.Stmt
}

/*func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		convoStmts = ConvoStmts{
			getPosts: acc.Select("conversations_posts").Columns("pid, body, post").Where("cid = ?").Prepare(),
			edit: acc.Update("conversations").Set("participants = ?, lastReplyAt = ?").Where("cid = ?").Prepare(),
			create: acc.Insert("conversations").Columns("participants, createdAt, lastReplyAt").Fields("?,UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),

			editPost: acc.Update("conversations_posts").Set("body = ?").Where("cid = ?").Prepare(),
			createPost: acc.Insert("conversations_posts").Columns("body").Fields("?").Prepare(),
		}
		return acc.FirstError()
	})
}*/

type Conversation struct {
	ID           int
	Participants string
	CreatedAt time.Time
	LastReplyAt time.Time
}

func (co *Conversation) Posts(offset int) (posts []*ConversationPost, err error) {
	rows, err := convoStmts.getPosts.Query(co.ID, offset, Config.ItemsPerPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		convo := &ConversationPost{CID: co.ID}
		err := rows.Scan(&convo.ID, &convo.Body, &convo.Post)
		if err != nil {
			return nil, err
		}
		convo, err = ConvoPostProcess.OnLoad(convo)
		if err != nil {
			return nil, err
		}
		posts = append(posts, convo)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	
	return posts, err
}

func (co *Conversation) Update() error {
	_, err := convoStmts.edit.Exec(co.Participants, co.CreatedAt, co.LastReplyAt, co.ID)
	return err
}

func (co *Conversation) Create() (int, error) {
	res, err := convoStmts.create.Exec(co.Participants)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

type ConversationPost struct {
	ID   int
	CID int
	Body string
	Post string // aes, ''
}

func (co *ConversationPost) Update() error {
	lco, err := ConvoPostProcess.OnSave(co)
	if err != nil {
		return err
	}
	//GetHookTable().VhookNoRet("convo_post_update", lco)
	_, err = convoStmts.editPost.Exec(lco.Body, lco.ID)
	return err
}

func (co *ConversationPost) Create() (int, error) {
	lco, err := ConvoPostProcess.OnSave(co)
	if err != nil {
		return 0, err
	}
	//GetHookTable().VhookNoRet("convo_post_create", lco)
	res, err := convoStmts.createPost.Exec(lco.Body)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

type ConversationStore interface {
	Get(id int) (*Conversation, error)
	Delete(id int) error
	Count() (count int)
}

type DefaultConversationStore struct {
	get    *sql.Stmt
	delete *sql.Stmt
	count  *sql.Stmt
}

func NewDefaultConversationStore(acc *qgen.Accumulator) (*DefaultConversationStore, error) {
	return &DefaultConversationStore{
		get:    acc.Select("conversations").Columns("participants, createdAt, lastReplyAt").Where("cid = ?").Prepare(),
		delete: acc.Delete("conversations").Where("cid = ?").Prepare(),
		count:  acc.Count("conversations").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultConversationStore) Get(id int) (*Conversation, error) {
	convo := &Conversation{ID: id}
	err := s.get.QueryRow(id).Scan(&convo.Participants, &convo.CreatedAt, &convo.LastReplyAt)
	return nil, err
}

func (s *DefaultConversationStore) Delete(id int) error {
	_, err := s.delete.Exec(id)
	return err
}

// Count returns the total number of topics on these forums
func (s *DefaultConversationStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}