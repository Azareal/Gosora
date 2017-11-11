package main

import (
	"database/sql"
	"log"

	"./common"
)

var stmts *Stmts

var db *sql.DB
var dbVersion string
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

	err = common.DbInits.Run()
	if err != nil {
		return err
	}

	log.Print("Loading the usergroups.")
	common.Gstore, err = common.NewMemoryGroupStore()
	if err != nil {
		return err
	}
	err2 := common.Gstore.LoadGroups()
	if err2 != nil {
		return err2
	}

	// We have to put this here, otherwise LoadForums() won't be able to get the last poster data when building it's forums
	log.Print("Initialising the user and topic stores")
	if common.Config.CacheTopicUser == common.CACHE_STATIC {
		common.Users, err = common.NewMemoryUserStore(common.Config.UserCacheCapacity)
		common.Topics, err2 = common.NewMemoryTopicStore(common.Config.TopicCacheCapacity)
	} else {
		common.Users, err = common.NewSQLUserStore()
		common.Topics, err2 = common.NewSQLTopicStore()
	}
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	log.Print("Loading the forums.")
	common.Fstore, err = common.NewMemoryForumStore()
	if err != nil {
		return err
	}
	err = common.Fstore.LoadForums()
	if err != nil {
		return err
	}

	log.Print("Loading the forum permissions.")
	common.Fpstore, err = common.NewMemoryForumPermsStore()
	if err != nil {
		return err
	}
	err = common.Fpstore.Init()
	if err != nil {
		return err
	}

	log.Print("Loading the settings.")
	err = common.LoadSettings()
	if err != nil {
		return err
	}

	log.Print("Loading the plugins.")
	err = common.InitExtend()
	if err != nil {
		return err
	}

	log.Print("Loading the themes.")
	return common.Themes.LoadActiveStatus()
}
