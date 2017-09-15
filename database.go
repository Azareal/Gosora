package main

import "log"

import "database/sql"

var db *sql.DB
var dbVersion string
var dbAdapter string

// ErrNoRows is an alias of sql.ErrNoRows, just in case we end up with non-database/sql datastores
var ErrNoRows = sql.ErrNoRows

func initDatabase() (err error) {
	// Engine specific code
	err = _initDatabase()
	if err != nil {
		return err
	}

	log.Print("Loading the usergroups.")
	gstore = NewMemoryGroupStore()
	err = gstore.LoadGroups()
	if err != nil {
		return err
	}

	log.Print("Loading the forums.")
	fstore = NewMemoryForumStore()
	err = fstore.LoadForums()
	if err != nil {
		return err
	}

	log.Print("Loading the forum permissions.")
	err = buildForumPermissions()
	if err != nil {
		return err
	}

	log.Print("Loading the settings.")
	err = LoadSettings()
	if err != nil {
		return err
	}

	log.Print("Loading the plugins.")
	err = LoadPlugins()
	if err != nil {
		return err
	}

	log.Print("Loading the themes.")
	return LoadThemes()
}
