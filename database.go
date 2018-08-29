package main

import (
	"database/sql"
	"log"

	"./common"
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

	log.Print("Running the db handlers.")
	err = common.DbInits.Run()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the usergroups.")
	common.Groups, err = common.NewMemoryGroupStore()
	if err != nil {
		return errors.WithStack(err)
	}
	err2 := common.Groups.LoadGroups()
	if err2 != nil {
		return errors.WithStack(err2)
	}

	// We have to put this here, otherwise LoadForums() won't be able to get the last poster data when building it's forums
	log.Print("Initialising the user and topic stores")

	var ucache common.UserCache
	if common.Config.UserCache == "static" {
		ucache = common.NewMemoryUserCache(common.Config.UserCacheCapacity)
	}

	var tcache common.TopicCache
	if common.Config.TopicCache == "static" {
		tcache = common.NewMemoryTopicCache(common.Config.TopicCacheCapacity)
	}

	common.Users, err = common.NewDefaultUserStore(ucache)
	if err != nil {
		return errors.WithStack(err)
	}
	common.Topics, err = common.NewDefaultTopicStore(tcache)
	if err != nil {
		return errors.WithStack(err2)
	}

	log.Print("Loading the forums.")
	common.Forums, err = common.NewMemoryForumStore()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.Forums.LoadForums()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the forum permissions.")
	common.FPStore, err = common.NewMemoryForumPermsStore()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.FPStore.Init()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the settings.")
	err = common.LoadSettings()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the plugins.")
	err = common.InitExtend()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the themes.")
	if err != nil {
		return errors.WithStack(common.Themes.LoadActiveStatus())
	}
	return nil
}
