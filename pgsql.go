// +build pgsql

/* Copyright Azareal 2016 - 2018 */
/* Super experimental and incomplete. DON'T USE IT YET! */
package main

import "strings"
import "database/sql"
import _ "github.com/lib/pq"
import "./query_gen/lib"

// TO-DO: Add support for SSL for all database drivers, not just pgsql
var db_sslmode = "disable" // verify-full
var get_activity_feed_by_watcher_stmt *sql.Stmt
var get_activity_count_by_watcher_stmt *sql.Stmt
var todays_post_count_stmt *sql.Stmt
var todays_topic_count_stmt *sql.Stmt
var todays_report_count_stmt *sql.Stmt
var todays_newuser_count_stmt *sql.Stmt
// Note to self: PostgreSQL listens on a different port than MySQL does

func _init_database() (err error) {
	// TO-DO: Investigate connect_timeout to see what it does exactly and whether it's relevant to us
	var _dbpassword string
	if(dbpassword != ""){
		_dbpassword = " password='" + _escape_bit(dbpassword) + "'"
	}
	db, err = sql.Open("postgres", "host='" + _escape_bit(dbhost) + "' port='" + _escape_bit(dbport) + "' user='" + _escape_bit(dbuser) + "' dbname='" + _escape_bit(dbname) + "'" + _dbpassword + " sslmode='" + db_sslmode + "'")
	if err != nil {
		return err
	}

	// Fetch the database version
	db.QueryRow("SELECT VERSION()").Scan(&db_version)

	// Set the number of max open connections. How many do we need? Might need to do some tests.
	db.SetMaxOpenConns(64)

	err = _gen_pgsql()
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("pgsql")
	if err != nil {
		return err
	}

	// TO-DO Handle the queries which the query generator can't handle yet

	return nil
}

func _escape_bit(bit string) string {
	// TO-DO: Write a custom parser, so that backslashes work properly in the sql.Open string. Do something similar for the database driver, if possible?
	return strings.Replace(bit,"'","\\'",-1)
}
