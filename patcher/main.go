package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"

	"../query_gen/lib"
	"./common"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			pressAnyKey(scanner)
			return
		}
	}()

	if common.DbConfig != "mysql" && common.DbConfig != "" {
		log.Fatal("Only MySQL is supported for upgrades right now, please wait for a newer build of the patcher")
	}

	err := prepMySQL()
	if err != nil {
		log.Fatal(err)
	}

	err = patcher(scanner)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove("./patcher/config.go")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove("./patcher/common/site.go")
	if err != nil {
		log.Fatal(err)
	}
}

func pressAnyKey(scanner *bufio.Scanner) {
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}

func prepMySQL() error {
	return qgen.Builder.Init("mysql", map[string]string{
		"host":      common.DbConfig.Host,
		"port":      common.DbConfig.Port,
		"name":      common.DbConfig.Dbname,
		"username":  common.DbConfig.Username,
		"password":  common.DbConfig.Password,
		"collation": "utf8mb4_general_ci",
	})
}

type SchemaFile struct {
	DBVersion          int // Current version of the database schema
	DynamicFileVersion int
	MinGoVersion       string // TODO: Minimum version of Go needed to install this version
	MinVersion         string // TODO: Minimum version of Gosora to jump to this version, might be tricky as we don't store this in the schema file, maybe store it in the database
}

func patcher(scanner *bufio.Scanner) error {
	data, err := ioutil.ReadFile("./schema/lastSchema.json")
	if err != nil {
		return err
	}

	var schemaFile LanguagePack
	err = json.Unmarshal(data, &schemaFile)
	if err != nil {
		return err
	}
	_ = schemaFile
	return patch0(scanner)
}

func execStmt(stmt *sql.Stmt, err error) error {
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

/*func eachUserQuick(handle func(int)) error {
	stmt, err := qgen.Builder.Select("users").Orderby("uid desc").Limit(1).Prepare()
	if err != nil {
		return err
	}

	var topID int
	err := stmt.QueryRow(topID)
	if err != nil {
		return err
	}

	for i := 1; i <= topID; i++ {
		err = handle(i)
		if err != nil {
			return err
		}
	}
}*/

func eachUser(handle func(int) error) error {
	acc := qgen.Builder.Accumulator()
	err := acc.Select("users").Each(func(rows *sql.Rows) error {
		var uid int
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}
		return handle(uid)
	})
	return err
}
