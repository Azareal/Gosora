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

	log.Print("Running the db handlers.")
	err = common.DbInits.Run()
	if err != nil {
		return err
	}

	log.Print("Loading the usergroups.")
	common.Groups, err = common.NewMemoryGroupStore()
	if err != nil {
		return err
	}
	err2 := common.Groups.LoadGroups()
	if err2 != nil {
		return err2
	}

	// We have to put this here, otherwise LoadForums() won't be able to get the last poster data when building it's forums
	log.Print("Initialising the user and topic stores")

	var ucache common.UserCache
	var tcache common.TopicCache
	if common.Config.CacheTopicUser == common.CACHE_STATIC {
		ucache = common.NewMemoryUserCache(common.Config.UserCacheCapacity)
		tcache = common.NewMemoryTopicCache(common.Config.TopicCacheCapacity)
	}

	common.Users, err = common.NewDefaultUserStore(ucache)
	if err != nil {
		return err
	}
	common.Topics, err = common.NewDefaultTopicStore(tcache)
	if err != nil {
		return err2
	}

	log.Print("Loading the forums.")
	common.Forums, err = common.NewMemoryForumStore()
	if err != nil {
		return err
	}
	err = common.Forums.LoadForums()
	if err != nil {
		return err
	}

	log.Print("Loading the forum permissions.")
	common.FPStore, err = common.NewMemoryForumPermsStore()
	if err != nil {
		return err
	}
	err = common.FPStore.Init()
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
