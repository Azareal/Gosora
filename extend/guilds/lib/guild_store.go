package guilds

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Gstore GuildStore

type GuildStore interface {
	Get(id int) (g *Guild, err error)
	Create(name, desc string, active bool, privacy, uid, fid int) (int, error)
}

type SQLGuildStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLGuildStore() (*SQLGuildStore, error) {
	acc := qgen.NewAcc()
	return &SQLGuildStore{
		get:    acc.Select("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Where("guildID=?").Prepare(),
		create: acc.Insert("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Fields("?,?,?,?,1,?,1,?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),
	}, acc.FirstError()
}

func (s *SQLGuildStore) Close() {
	_ = s.get.Close()
	_ = s.create.Close()
}

func (s *SQLGuildStore) Get(id int) (g *Guild, err error) {
	g = &Guild{ID: id}
	err = s.get.QueryRow(id).Scan(&g.Name, &g.Desc, &g.Active, &g.Privacy, &g.Joinable, &g.Owner, &g.MemberCount, &g.MainForumID, &g.Backdrop, &g.CreatedAt, &g.LastUpdateTime)
	return g, err
}

func (s *SQLGuildStore) Create(name, desc string, active bool, privacy, uid, fid int) (int, error) {
	res, err := s.create.Exec(name, desc, active, privacy, uid, fid)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}
