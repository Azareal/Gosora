package guilds

import "database/sql"
import "../../../query_gen/lib"

var Gstore GuildStore

type GuildStore interface {
	Get(guildID int) (guild *Guild, err error)
	Create(name string, desc string, active bool, privacy int, uid int, fid int) (int, error)
}

type SQLGuildStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLGuildStore() (*SQLGuildStore, error) {
	acc := qgen.NewAcc()
	return &SQLGuildStore{
		get:    acc.Select("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Where("guildID = ?").Prepare(),
		create: acc.Insert("guilds").Columns("name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime").Fields("?,?,?,?,1,?,1,?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(),
	}, acc.FirstError()
}

func (store *SQLGuildStore) Close() {
	_ = store.get.Close()
	_ = store.create.Close()
}

func (store *SQLGuildStore) Get(guildID int) (guild *Guild, err error) {
	guild = &Guild{ID: guildID}
	err = store.get.QueryRow(guildID).Scan(&guild.Name, &guild.Desc, &guild.Active, &guild.Privacy, &guild.Joinable, &guild.Owner, &guild.MemberCount, &guild.MainForumID, &guild.Backdrop, &guild.CreatedAt, &guild.LastUpdateTime)
	return guild, err
}

func (store *SQLGuildStore) Create(name string, desc string, active bool, privacy int, uid int, fid int) (int, error) {
	res, err := store.create.Exec(name, desc, active, privacy, uid, fid)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}
