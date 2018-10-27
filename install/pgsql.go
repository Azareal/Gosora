/*
*
* Gosora PostgreSQL Interface
* Under heavy development
* Copyright Azareal 2017 - 2018
*
 */
package install

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Azareal/Gosora/query_gen"
	_ "github.com/go-sql-driver/mysql"
)

// We don't need SSL to run an installer... Do we?
var dbSslmode = "disable"

func init() {
	adapters["pgsql"] = &PgsqlInstaller{dbHost: ""}
}

type PgsqlInstaller struct {
	db         *sql.DB
	dbHost     string
	dbUsername string
	dbPassword string
	dbName     string
	dbPort     string
}

func (ins *PgsqlInstaller) SetConfig(dbHost string, dbUsername string, dbPassword string, dbName string, dbPort string) {
	ins.dbHost = dbHost
	ins.dbUsername = dbUsername
	ins.dbPassword = dbPassword
	ins.dbName = dbName
	ins.dbPort = dbPort
}

func (ins *PgsqlInstaller) Name() string {
	return "pgsql"
}

func (ins *PgsqlInstaller) DefaultPort() string {
	return "5432"
}

func (ins *PgsqlInstaller) InitDatabase() (err error) {
	_dbPassword := ins.dbPassword
	if _dbPassword != "" {
		_dbPassword = " password=" + pgEscapeBit(_dbPassword)
	}
	db, err := sql.Open("postgres", "host='"+pgEscapeBit(ins.dbHost)+"' port='"+pgEscapeBit(ins.dbPort)+"' user='"+pgEscapeBit(ins.dbUsername)+"' dbname='"+pgEscapeBit(ins.dbName)+"'"+_dbPassword+" sslmode='"+dbSslmode+"'")
	if err != nil {
		return err
	}

	// Make sure that the connection is alive..
	err = db.Ping()
	if err != nil {
		return err
	}
	fmt.Println("Successfully connected to the database")

	// TODO: Create the database, if it doesn't exist

	// Ready the query builder
	ins.db = db
	qgen.Builder.SetConn(db)
	return qgen.Builder.SetAdapter("pgsql")
}

func (ins *PgsqlInstaller) TableDefs() (err error) {
	return errors.New("TableDefs() not implemented")
}

func (ins *PgsqlInstaller) InitialData() (err error) {
	return errors.New("InitialData() not implemented")
}

func (ins *PgsqlInstaller) CreateAdmin() error {
	return createAdmin()
}

func (ins *PgsqlInstaller) DBHost() string {
	return ins.dbHost
}

func (ins *PgsqlInstaller) DBUsername() string {
	return ins.dbUsername
}

func (ins *PgsqlInstaller) DBPassword() string {
	return ins.dbPassword
}

func (ins *PgsqlInstaller) DBName() string {
	return ins.dbName
}

func (ins *PgsqlInstaller) DBPort() string {
	return ins.dbPort
}

func pgEscapeBit(bit string) string {
	// TODO: Write a custom parser, so that backslashes work properly in the sql.Open string. Do something similar for the database driver, if possible?
	return strings.Replace(bit, "'", "\\'", -1)
}
