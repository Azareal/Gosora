/*
*
* Gosora MSSQL Interface
* Copyright Azareal 2017 - 2018
*
 */
package install

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"../../query_gen/lib"
	_ "github.com/denisenkom/go-mssqldb"
)

func init() {
	adapters["mssql"] = &MssqlInstaller{dbHost: ""}
}

type MssqlInstaller struct {
	db         *sql.DB
	dbHost     string
	dbUsername string
	dbPassword string
	dbName     string
	dbInstance string
	dbPort     string
}

func (ins *MssqlInstaller) SetConfig(dbHost string, dbUsername string, dbPassword string, dbName string, dbPort string) {
	ins.dbHost = dbHost
	ins.dbUsername = dbUsername
	ins.dbPassword = dbPassword
	ins.dbName = dbName
	ins.dbInstance = "" // You can't set this from the installer right now, it allows you to connect to a named instance instead of a port
	ins.dbPort = dbPort
}

func (ins *MssqlInstaller) Name() string {
	return "mssql"
}

func (ins *MssqlInstaller) DefaultPort() string {
	return "1433"
}

func (ins *MssqlInstaller) InitDatabase() (err error) {
	query := url.Values{}
	query.Add("database", ins.dbName)
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(ins.dbUsername, ins.dbPassword),
		Host:     ins.dbHost + ":" + ins.dbPort,
		Path:     ins.dbInstance,
		RawQuery: query.Encode(),
	}
	log.Print("u.String() ", u.String())

	db, err := sql.Open("mssql", u.String())
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
	return qgen.Builder.SetAdapter("mssql")
}

func (ins *MssqlInstaller) TableDefs() (err error) {
	//fmt.Println("Creating the tables")
	files, _ := ioutil.ReadDir("./schema/mssql/")
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

		// ? - This is mainly here for tests, although it might allow the installer to overwrite a production database, so we might want to proceed with caution
		_, err = ins.db.Exec("DROP TABLE IF EXISTS [" + table + "];")
		if err != nil {
			fmt.Println("Failed query:", "DROP TABLE IF EXISTS ["+table+"]")
			return err
		}

		fmt.Println("Creating table '" + table + "'")
		data, err := ioutil.ReadFile("./schema/mssql/" + f.Name())
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
	return nil
}

func (ins *MssqlInstaller) InitialData() (err error) {
	//fmt.Println("Seeding the tables")
	data, err := ioutil.ReadFile("./schema/mssql/inserts.sql")
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
	return nil
}

func (ins *MssqlInstaller) CreateAdmin() error {
	return createAdmin()
}

func (ins *MssqlInstaller) DBHost() string {
	return ins.dbHost
}

func (ins *MssqlInstaller) DBUsername() string {
	return ins.dbUsername
}

func (ins *MssqlInstaller) DBPassword() string {
	return ins.dbPassword
}

func (ins *MssqlInstaller) DBName() string {
	return ins.dbName
}

func (ins *MssqlInstaller) DBPort() string {
	return ins.dbPort
}
