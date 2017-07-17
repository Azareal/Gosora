// +build !pgsql !sqlite !mssql

/* Copyright Azareal 2016 - 2018 */
package main

import "log"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import "./query_gen/lib"

var db_collation string = "utf8mb4_general_ci"
var get_activity_feed_by_watcher_stmt *sql.Stmt
var get_activity_count_by_watcher_stmt *sql.Stmt
var todays_post_count_stmt *sql.Stmt
var todays_topic_count_stmt *sql.Stmt
var todays_report_count_stmt *sql.Stmt
var todays_newuser_count_stmt *sql.Stmt

func _init_database() (err error) {
	var _dbpassword string
	if(db_config.Password != ""){
		_dbpassword = ":" + db_config.Password
	}

	// Open the database connection
	db, err = sql.Open("mysql", db_config.Username + _dbpassword + "@tcp(" + db_config.Host + ":" + db_config.Port + ")/" + db_config.Dbname + "?collation=" + db_collation)
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

	// Set the number of max open connections
	db.SetMaxOpenConns(64)

	// Build the generated prepared statements, we are going to slowly move the queries over to the query generator rather than writing them all by hand, this'll make it easier for us to implement database adapters for other databases like PostgreSQL, MSSQL, SQlite, etc.
	err = _gen_mysql()
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("mysql")
	if err != nil {
		return err
	}

	log.Print("Preparing get_activity_feed_by_watcher statement.")
	get_activity_feed_by_watcher_stmt, err = db.Prepare("SELECT activity_stream_matches.asid, activity_stream.actor, activity_stream.targetUser, activity_stream.event, activity_stream.elementType, activity_stream.elementID FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ? ORDER BY activity_stream.asid ASC LIMIT 8")
	if err != nil {
		return err
	}

	log.Print("Preparing get_activity_count_by_watcher statement.")
	get_activity_count_by_watcher_stmt, err = db.Prepare("SELECT count(*) FROM `activity_stream_matches` INNER JOIN `activity_stream` ON activity_stream_matches.asid = activity_stream.asid AND activity_stream_matches.watcher != activity_stream.actor WHERE `watcher` = ?")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_post_count statement.")
	todays_post_count_stmt, err = db.Prepare("select count(*) from replies where createdAt BETWEEN (utc_timestamp() - interval 1 day) and utc_timestamp()")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_topic_count statement.")
	todays_topic_count_stmt, err = db.Prepare("select count(*) from topics where createdAt BETWEEN (utc_timestamp() - interval 1 day) and utc_timestamp()")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_report_count statement.")
	todays_report_count_stmt, err = db.Prepare("select count(*) from topics where createdAt BETWEEN (utc_timestamp() - interval 1 day) and utc_timestamp() and parentID = 1")
	if err != nil {
		return err
	}

	log.Print("Preparing todays_newuser_count statement.")
	todays_newuser_count_stmt, err = db.Prepare("select count(*) from users where createdAt BETWEEN (utc_timestamp() - interval 1 day) and utc_timestamp()")
	if err != nil {
		return err
	}

	return nil
}
