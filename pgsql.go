// +build pgsql

/* Copyright Azareal 2016 - 2018 */
/* Super experimental and incomplete. DON'T USE IT YET! */
package main

import "strings"

//import "time"
import "database/sql"
import _ "github.com/lib/pq"
import "./query_gen/lib"

// TODO: Add support for SSL for all database drivers, not just pgsql
var db_sslmode = "disable" // verify-full
var get_activity_feed_by_watcher_stmt *sql.Stmt
var get_activity_count_by_watcher_stmt *sql.Stmt
var todays_post_count_stmt *sql.Stmt
var todays_topic_count_stmt *sql.Stmt
var todays_report_count_stmt *sql.Stmt
var todays_newuser_count_stmt *sql.Stmt

func init() {
	db_adapter = "pgsql"
	_initDatabase = initPgsql
}

func initPgsql() (err error) {
	// TODO: Investigate connect_timeout to see what it does exactly and whether it's relevant to us
	var _dbpassword string
	if dbpassword != "" {
		_dbpassword = " password='" + _escape_bit(db_config.Password) + "'"
	}
	// TODO: Move this bit to the query gen lib
	db, err = sql.Open("postgres", "host='"+_escape_bit(db_config.Host)+"' port='"+_escape_bit(db_config.Port)+"' user='"+_escape_bit(db_config.Username)+"' dbname='"+_escape_bit(config.Dbname)+"'"+_dbpassword+" sslmode='"+db_sslmode+"'")
	if err != nil {
		return err
	}

	// Make sure that the connection is alive
	err = db.Ping()
	if err != nil {
		return err
	}

	// Fetch the database version
	db.QueryRow("SELECT VERSION()").Scan(&db_version)

	// Set the number of max open connections. How many do we need? Might need to do some tests.
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(32)

	// Only hold connections open for five seconds to avoid accumulating a large number of stale connections
	//db.SetConnMaxLifetime(5 * time.Second)

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
	// TODO: Write a custom parser, so that backslashes work properly in the sql.Open string. Do something similar for the database driver, if possible?
	return strings.Replace(bit, "'", "\\'", -1)
}
