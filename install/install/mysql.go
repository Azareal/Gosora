/*
*
* Gosora MySQL Interface
* Copyright Azareal 2017 - 2018
*
 */
package install

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"../../query_gen/lib"
	_ "github.com/go-sql-driver/mysql"
)

//var dbCollation string = "utf8mb4_general_ci"

func init() {
	adapters["mysql"] = &MysqlInstaller{dbHost: ""}
}

type MysqlInstaller struct {
	db         *sql.DB
	dbHost     string
	dbUsername string
	dbPassword string
	dbName     string
	dbPort     string
}

func (ins *MysqlInstaller) SetConfig(dbHost string, dbUsername string, dbPassword string, dbName string, dbPort string) {
	ins.dbHost = dbHost
	ins.dbUsername = dbUsername
	ins.dbPassword = dbPassword
	ins.dbName = dbName
	ins.dbPort = dbPort
}

func (ins *MysqlInstaller) Name() string {
	return "mysql"
}

func (ins *MysqlInstaller) DefaultPort() string {
	return "3306"
}

func (ins *MysqlInstaller) InitDatabase() (err error) {
	_dbPassword := ins.dbPassword
	if _dbPassword != "" {
		_dbPassword = ":" + _dbPassword
	}
	db, err := sql.Open("mysql", ins.dbUsername+_dbPassword+"@tcp("+ins.dbHost+":"+ins.dbPort+")/")
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
	err = db.QueryRow("SHOW DATABASES LIKE '" + ins.dbName + "'").Scan(&waste)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		fmt.Println("Unable to find the database. Attempting to create it")
		_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + ins.dbName)
		if err != nil {
			return err
		}
		fmt.Println("The database was successfully created")
	}

	fmt.Println("Switching to database " + ins.dbName)
	_, err = db.Exec("USE " + ins.dbName)
	if err != nil {
		return err
	}

	// Ready the query builder
	qgen.Builder.SetConn(db)
	err = qgen.Builder.SetAdapter("mysql")
	if err != nil {
		return err
	}

	ins.db = db
	return nil
}

func (ins *MysqlInstaller) TableDefs() error {
	//fmt.Println("Creating the tables")
	files, _ := ioutil.ReadDir("./schema/mysql/")
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "query_") {
			continue
		}

		var table, ext string
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

		_, err = ins.db.Exec(string(data))
		if err != nil {
			fmt.Println("Failed query:", string(data))
			return err
		}
	}
	//fmt.Println("Finished creating the tables")
	return nil
}

// ? - Moved this here since it was breaking the installer, we need to add this at some point
/* TODO: Implement the html-attribute setting type before deploying this */
/*INSERT INTO settings(`name`,`content`,`type`) VALUES ('meta_desc','','html-attribute');*/

func (ins *MysqlInstaller) InitialData() error {
	//fmt.Println("Seeding the tables")
	data, err := ioutil.ReadFile("./schema/mysql/inserts.sql")
	if err != nil {
		return err
	}
	data = bytes.TrimSpace(data)

	statements := bytes.Split(data, []byte(";"))
	for key, statement := range statements {
		if len(statement) == 0 {
			continue
		}

		fmt.Println("Executing query #" + strconv.Itoa(key) + " " + string(statement))
		_, err = ins.db.Exec(string(statement))
		if err != nil {
			return err
		}
	}

	//fmt.Println("Finished inserting the database data")
	return nil
}

func (ins *MysqlInstaller) CreateAdmin() error {
	return createAdmin()
}

func (ins *MysqlInstaller) DBHost() string {
	return ins.dbHost
}

func (ins *MysqlInstaller) DBUsername() string {
	return ins.dbUsername
}

func (ins *MysqlInstaller) DBPassword() string {
	return ins.dbPassword
}

func (ins *MysqlInstaller) DBName() string {
	return ins.dbName
}

func (ins *MysqlInstaller) DBPort() string {
	return ins.dbPort
}
