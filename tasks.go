/*
*
*	Gosora Task System
*	Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"log"
	"time"
)

var lastSync time.Time

func init() {
	lastSync = time.Now()
}

func handleExpiredScheduledGroups() error {
	rows, err := getExpiredScheduledGroupsStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var uid int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}

		// Sneaky way of initialising a *User, please use the methods on the UserStore instead
		user := getDummyUser()
		user.ID = uid
		err = user.RevertGroupUpdate()
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

func handleServerSync() error {
	var lastUpdate time.Time
	err := getSyncStmt.QueryRow().Scan(&lastUpdate)
	if err != nil {
		return err
	}

	if lastUpdate.After(lastSync) {
		// TODO: A more granular sync
		err = fstore.LoadForums()
		if err != nil {
			log.Print("Unable to reload the forums")
			return err
		}
		// TODO: Resync the groups
		// TODO: Resync the permissions
		err = LoadSettings()
		if err != nil {
			log.Print("Unable to reload the settings")
			return err
		}
		err = LoadWordFilters()
		if err != nil {
			log.Print("Unable to reload the word filters")
			return err
		}
	}
	return nil
}
