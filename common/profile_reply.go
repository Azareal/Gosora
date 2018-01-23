package common

import (
	"database/sql"
	"html"
	"time"

	"../query_gen/lib"
)

var profileReplyStmts ProfileReplyStmts

type ProfileReply struct {
	ID                int
	ParentID          int
	Content           string
	CreatedBy         int
	Group             int
	CreatedAt         time.Time
	RelativeCreatedAt string
	LastEdit          int
	LastEditBy        int
	ContentLines      int
	IPAddress         string
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

// TODO: Write tests for this
func (reply *ProfileReply) Delete() error {
	_, err := profileReplyStmts.delete.Exec(reply.ID)
	return err
}

func (reply *ProfileReply) SetBody(content string) error {
	content = PreparseMessage(html.UnescapeString(content))
	parsedContent := ParseMessage(content, 0, "")
	_, err := profileReplyStmts.edit.Exec(content, parsedContent, reply.ID)
	return err
}

// TODO: We can get this from the topic store instead of a query which will always miss the cache...
func (reply *ProfileReply) Creator() (*User, error) {
	return Users.Get(reply.CreatedBy)
}