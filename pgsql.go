// +build pgsql

/* Copyright Azareal 2016 - 2019 */
/* Super experimental and incomplete. DON'T USE IT YET! */
package main

import (
	"database/sql"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
	_ "github.com/lib/pq"
)

// TODO: Add support for SSL for all database drivers, not just pgsql
var dbSslmode = "disable" // verify-full

func init() {
	dbAdapter = "pgsql"
	_initDatabase = initPgsql
}

func initPgsql() (err error) {
	// TODO: Investigate connect_timeout to see what it does exactly and whether it's relevant to us
	var _dbpassword string
	if common.DbConfig.Password != "" {
		_dbpassword = " password='" + _escape_bit(common.DbConfig.Password) + "'"
	}
	// TODO: Move this bit to the query gen lib
	db, err = sql.Open("postgres", "host='"+_escape_bit(common.DbConfig.Host)+"' port='"+_escape_bit(common.DbConfig.Port)+"' user='"+_escape_bit(common.DbConfig.Username)+"' dbname='"+_escape_bit(common.DbConfig.Dbname)+"'"+_dbpassword+" sslmode='"+dbSslmode+"'")
	if err != nil {
		return err
	}

	// Make sure that the connection is alive
	err = db.Ping()
	if err != nil {
		return err
	}

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
