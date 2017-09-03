/*
*
* Gosora MySQL Interface
* Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"../query_gen/lib"
	_ "github.com/go-sql-driver/mysql"
)

//var dbCollation string = "utf8mb4_general_ci"

func _setMysqlAdapter() {
	dbPort = "3306"
	initDatabase = _initMysql
	tableDefs = _tableDefsMysql
	initialData = _initialDataMysql
}

func _initMysql() (err error) {
	_dbPassword := dbPassword
	if _dbPassword != "" {
		_dbPassword = ":" + _dbPassword
	}
	db, err = sql.Open("mysql", dbUsername+_dbPassword+"@tcp("+dbHost+":"+dbPort+")/")
	if err != nil {
		return err
	}

	// Make sure that the connection is alive..
	err = db.Ping()
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to the database")

	var waste string
	err = db.QueryRow("SHOW DATABASES LIKE '" + dbName + "'").Scan(&waste)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		fmt.Println("Unable to find the database. Attempting to create it")
		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + "")
		if err != nil {
			return err
		}
		fmt.Println("The database was successfully created")
	}

	fmt.Println("Switching to database " + dbName)
	_, err = db.Exec("USE " + dbName)
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("mysql")
	if err != nil {
		return err
	}

	return nil
}

func _tableDefsMysql() error {
	//fmt.Println("Creating the tables")
	files, _ := ioutil.ReadDir("./schema/mysql/")
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "query_") {
			continue
		}

		var table string
		var ext string
		table = strings.TrimPrefix(f.Name(), "query_")
		ext = filepath.Ext(table)
		if ext != ".sql" {
			continue
		}
		table = strings.TrimSuffix(table, ext)

		fmt.Println("Creating table '" + table + "'")
		data, err := ioutil.ReadFile("./schema/mysql/" + f.Name())
		if err != nil {
			return err
		}
		data = bytes.TrimSpace(data)

		_, err = db.Exec(string(data))
		if err != nil {
			fmt.Println("Failed query:", string(data))
			return err
		}
	}
	//fmt.Println("Finished creating the tables")
	return nil
}

func _initialDataMysql() error {
	return nil // Coming Soon

	/*fmt.Println("Seeding the tables")
	data, err := ioutil.ReadFile("./schema/mysql/inserts.sql")
	if err != nil {
		return err
	}
	data = bytes.TrimSpace(data)

	fmt.Println("Executing query",string(data))
	_, err = db.Exec(string(data))
	if err != nil {
		return err
	}

	//fmt.Println("Finished inserting the database data")
	return nil*/
}

func _mysqlSeedDatabase() error {
	fmt.Println("Opening the database seed file")
	sqlContents, err := ioutil.ReadFile("./mysql.sql")
	if err != nil {
		return err
	}

	fmt.Println("Preparing installation queries")
	sqlContents = bytes.TrimSpace(sqlContents)
	statements := bytes.Split(sqlContents, []byte(";"))
	for key, statement := range statements {
		if len(statement) == 0 {
			continue
		}

		fmt.Println("Executing query #" + strconv.Itoa(key) + " " + string(statement))
		_, err = db.Exec(string(statement))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished inserting the database data")
	return nil
}
