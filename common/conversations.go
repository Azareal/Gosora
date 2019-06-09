package common

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

/*
conversations
conversations_posts
*/

var convoStmts ConvoStmts

type ConvoStmts struct {
	edit *sql.Stmt
	create *sql.Stmt
}

func init() {
	/*DbInits.Add(func(acc *qgen.Accumulator) error {
		convoStmts = ConvoStmts{
			edit: acc.Update("conversations").Set("participants = ?").Where("cid = ?").Prepare(),
			create: acc.Insert("conversations").Columns("participants").Fields("?").Prepare(),
		}
		return acc.FirstError()
	})*/
}

type Conversation struct {
	ID int
	Participants string
}

func (co *Conversation) Update() error {
	_, err := convoStmts.edit.Exec(co.Participants, co.ID)
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