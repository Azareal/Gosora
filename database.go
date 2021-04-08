package main

import (
	"database/sql"
	"log"

	c "github.com/Azareal/Gosora/common"
	"github.com/pkg/errors"
)

var stmts *Stmts

var db *sql.DB
var dbAdapter string

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

var _initDatabase func() error

func InitDatabase() (err error) {
	stmts = &Stmts{Mocks: false}

	// Engine specific code
	err = _initDatabase()
	if err != nil {
		return err
	}
	globs = &Globs{stmts}
	ws := errors.WithStack

	log.Print("Running the db handlers.")
	err = c.DbInits.Run()
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the usergroups.")
	c.Groups, err = c.NewMemoryGroupStore()
	if err != nil {
		return ws(err)
	}
	err2 := c.Groups.LoadGroups()
	if err2 != nil {
		return ws(err2)
	}

	// We have to put this here, otherwise LoadForums() won't be able to get the last poster data when building it's forums
	log.Print("Initialising the user and topic stores")

	var ucache c.UserCache
	if c.Config.UserCache == "static" {
		ucache = c.NewMemoryUserCache(c.Config.UserCacheCapacity)
	}
	var tcache c.TopicCache
	if c.Config.TopicCache == "static" {
		tcache = c.NewMemoryTopicCache(c.Config.TopicCacheCapacity)
	}

	c.Users, err = c.NewDefaultUserStore(ucache)
	if err != nil {
		return ws(err)
	}
	c.Topics, err = c.NewDefaultTopicStore(tcache)
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the forums.")
	c.Forums, err = c.NewMemoryForumStore()
	if err != nil {
		return ws(err)
	}
	err = c.Forums.LoadForums()
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the forum permissions.")
	c.FPStore, err = c.NewMemoryForumPermsStore()
	if err != nil {
		return ws(err)
	}
	err = c.FPStore.Init()
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the settings.")
	err = c.LoadSettings()
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the plugins.")
	err = c.InitExtend()
	if err != nil {
		return ws(err)
	}

	log.Print("Loading the themes.")
	err = c.Themes.LoadActiveStatus()
	if err != nil {
		return ws(err)
	}
	return nil
}
