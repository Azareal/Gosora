package common

import (
	"database/sql"
	"html"
	"time"

	"github.com/Azareal/Gosora/query_gen"
)

var profileReplyStmts ProfileReplyStmts

type ProfileReply struct {
	ID           int
	ParentID     int
	Content      string
	CreatedBy    int
	Group        int
	CreatedAt    time.Time
	LastEdit     int
	LastEditBy   int
	ContentLines int
	IP    string
}

type ProfileReplyStmts struct {
	edit   *sql.Stmt
	delete *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		profileReplyStmts = ProfileReplyStmts{
			edit:   acc.Update("users_replies").Set("content = ?, parsed_content = ?").Where("rid = ?").Prepare(),
			delete: acc.Delete("users_replies").Where("rid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// Mostly for tests, so we don't wind up with out-of-date profile reply initialisation logic there
func BlankProfileReply(id int) *ProfileReply {
	return &ProfileReply{ID: id}
}

// TODO: Write tests for this
func (r *ProfileReply) Delete() error {
	_, err := profileReplyStmts.delete.Exec(r.ID)
	return err
}

func (r *ProfileReply) SetBody(content string) error {
	content = PreparseMessage(html.UnescapeString(content))
	_, err := profileReplyStmts.edit.Exec(content, ParseMessage(content, 0, ""), r.ID)
	return err
}

// TODO: We can get this from the topic store instead of a query which will always miss the cache...
func (r *ProfileReply) Creator() (*User, error) {
	return Users.Get(r.CreatedBy)
}
