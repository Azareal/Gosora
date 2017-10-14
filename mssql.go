// +build mssql

/*
*
*	Gosora MSSQL Interface
*	Copyright Azareal 2016 - 2018
*
 */
package main

//import "time"
import (
	"database/sql"
	"net/url"

	"./query_gen/lib"
	_ "github.com/denisenkom/go-mssqldb"
)

var dbInstance string = ""

var getActivityFeedByWatcherStmt *sql.Stmt
var getActivityCountByWatcherStmt *sql.Stmt
var todaysPostCountStmt *sql.Stmt
var todaysTopicCountStmt *sql.Stmt
var todaysReportCountStmt *sql.Stmt
var todaysNewUserCountStmt *sql.Stmt
var findUsersByIPUsersStmt *sql.Stmt
var findUsersByIPTopicsStmt *sql.Stmt
var findUsersByIPRepliesStmt *sql.Stmt

func init() {
	dbAdapter = "mssql"
	_initDatabase = initMSSQL
}

func initMSSQL() (err error) {
	// TODO: Move this bit to the query gen lib
	query := url.Values{}
	query.Add("database", dbConfig.Dbname)
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(dbConfig.Username, dbConfig.Password),
		Host:     dbConfig.Host + ":" + dbConfig.Port,
		Path:     dbInstance,
		RawQuery: query.Encode(),
	}
	db, err = sql.Open("mssql", u.String())
	if err != nil {
		return err
	}

	// Make sure that the connection is alive
	err = db.Ping()
	if err != nil {
		return err
	}

	// TODO: Fetch the database version

	// Set the number of max open connections
	db.SetMaxOpenConns(64)
	db.SetMaxIdleConns(32)

	// Only hold connections open for five seconds to avoid accumulating a large number of stale connections
	//db.SetConnMaxLifetime(5 * time.Second)

	// Build the generated prepared statements, we are going to slowly move the queries over to the query generator rather than writing them all by hand, this'll make it easier for us to implement database adapters for other databases like PostgreSQL, MSSQL, SQlite, etc.
	err = _gen_mssql()
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("mssql")
	if err != nil {
		return err
	}

	// TODO: Add the custom queries

	return nil
}
