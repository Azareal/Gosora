/* Under heavy development */
/* Copyright Azareal 2017 - 2018 */
package main

import "fmt"
import "strings"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

// We don't need SSL to run an installer... Do we?
var db_sslmode = "disable"

func _set_pgsql_adapter() {
	db_port = "3306"
	init_database = _init_pgsql
}

func _init_pgsql() (err error) {
	_db_password := db_password
	if _db_password != "" {
		_db_password = " password=" + _pg_escape_bit(_db_password)
	}
	db, err = sql.Open("postgres", "host='" + _pg_escape_bit(db_host) + "' port='" + _pg_escape_bit(db_port) + "' user='" + _pg_escape_bit(db_username) + "' dbname='" + _pg_escape_bit(db_name) + "'" + _db_password + " sslmode='" + db_sslmode + "'")
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to the database")
	
	// TO-DO: Create the database, if it doesn't exist
	
	return nil
}

func _pg_escape_bit(bit string) string {
	// TO-DO: Write a custom parser, so that backslashes work properly in the sql.Open string. Do something similar for the database driver, if possible?
	return strings.Replace(bit,"'","\\'",-1)
}