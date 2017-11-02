package main

import "log"

import "database/sql"

var db *sql.DB
var dbVersion string
var dbAdapter string

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

var _initDatabase func() error

func initDatabase() (err error) {
	// Engine specific code
	err = _initDatabase()
	if err != nil {
		return err
	}

	log.Print("Loading the usergroups.")
	gstore, err = NewMemoryGroupStore()
	if err != nil {
		return err
	}
	err = gstore.LoadGroups()
	if err != nil {
		return err
	}

	// We have to put this here, otherwise LoadForums() won't be able to get the last poster data when building it's forums
	log.Print("Initialising the user and topic stores")
	if config.CacheTopicUser == CACHE_STATIC {
		users, err = NewMemoryUserStore(config.UserCacheCapacity)
		if err != nil {
			return err
		}
		topics, err = NewMemoryTopicStore(config.TopicCacheCapacity)
		if err != nil {
			return err
		}
	} else {
		users, err = NewSQLUserStore()
		if err != nil {
			return err
		}
		topics, err = NewSQLTopicStore()
		if err != nil {
			return err
		}
	}

	log.Print("Loading the forums.")
	fstore, err = NewMemoryForumStore()
	if err != nil {
		return err
	}
	err = fstore.LoadForums()
	if err != nil {
		return err
	}

	log.Print("Loading the forum permissions.")
	err = buildForumPermissions()
	if err != nil {
		return err
	}
	fpstore = NewForumPermsStore()

	log.Print("Loading the settings.")
	err = LoadSettings()
	if err != nil {
		return err
	}

	log.Print("Loading the plugins.")
	err = initExtend()
	if err != nil {
		return err
	}

	log.Print("Loading the themes.")
	return LoadThemes()
}
