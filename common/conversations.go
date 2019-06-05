package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

/*
conversations
conversations_posts
*/

type Conversation struct {
	ID int
	Participants string
}

func (co *Conversation) Create() (int, error) {
	return 0, sql.ErrNoRows
}

type ConversationPost struct {
}

type ConversationStore interface {
	Get(id int) (*Conversation, error)
	Delete(id int) error
	Count() (count int)
}

type DefaultConversationStore struct {
	get *sql.Stmt
	delete *sql.Stmt
	count *sql.Stmt
}

func NewDefaultConversationStore(acc *qgen.Accumulator) (*DefaultConversationStore, error) {
	return &DefaultConversationStore{
		get: acc.Select("conversations").Columns("participants").Where("cid = ?").Prepare(),
		delete: acc.Delete("conversations").Where("cid = ?").Prepare(),
		count: acc.Count("conversations").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultConversationStore) Get(id int) (*Conversation, error) {
	convo := &Conversation{ID:id}
	err := s.get.QueryRow(id).Scan(&convo.Participants)
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