/*
*
* Gosora PostgreSQL Interface
* Under heavy development
* Copyright Azareal 2017 - 2018
*
 */
package main

import "fmt"
import "strings"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

// We don't need SSL to run an installer... Do we?
var dbSslmode = "disable"

func _setPgsqlAdapter() {
	dbPort = "5432"
	initDatabase = _initPgsql
}

func _initPgsql() (err error) {
	_dbPassword := dbPassword
	if _dbPassword != "" {
		_dbPassword = " password=" + _pgEscapeBit(_dbPassword)
	}
	db, err = sql.Open("postgres", "host='"+_pgEscapeBit(dbHost)+"' port='"+_pgEscapeBit(dbPort)+"' user='"+_pgEscapeBit(dbUsername)+"' dbname='"+_pgEscapeBit(dbName)+"'"+_dbPassword+" sslmode='"+dbSslmode+"'")
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to the database")

	// TODO: Create the database, if it doesn't exist

	return nil
}

func _pgEscapeBit(bit string) string {
	// TODO: Write a custom parser, so that backslashes work properly in the sql.Open string. Do something similar for the database driver, if possible?
	return strings.Replace(bit, "'", "\\'", -1)
}
