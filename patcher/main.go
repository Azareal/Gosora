package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"../query_gen/lib"
	"./common"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

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
	err := qgen.Builder.Init("mysql", map[string]string{
		"host":      common.DbConfig.Host,
		"port":      common.DbConfig.Port,
		"name":      common.DbConfig.Dbname,
		"username":  common.DbConfig.Username,
		"password":  common.DbConfig.Password,
		"collation": "utf8mb4_general_ci",
	})
	if err != nil {
		return err
	}

	// Ready the query builder
	db = qgen.Builder.GetConn()
	return qgen.Builder.SetAdapter("mysql")
}

func execStmt(stmt *sql.Stmt, err error) error {
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func patcher(scanner *bufio.Scanner) error {
	err := execStmt(qgen.Builder.CreateTable("menus", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"mid", "int", 0, false, true, ""},
		},
		[]qgen.DBTableKey{
			qgen.DBTableKey{"mid", "primary"},
		},
	))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.CreateTable("menu_items", "", "",
		[]qgen.DBTableColumn{
			qgen.DBTableColumn{"mid", "int", 0, false, false, ""},
			qgen.DBTableColumn{"htmlID", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"cssClass", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"position", "varchar", 100, false, false, ""},
			qgen.DBTableColumn{"path", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"aria", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"tooltip", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"tmplName", "varchar", 200, false, false, "''"},
			qgen.DBTableColumn{"order", "int", 0, false, false, "0"},

			qgen.DBTableColumn{"guestOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"memberOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"staffOnly", "boolean", 0, false, false, "0"},
			qgen.DBTableColumn{"adminOnly", "boolean", 0, false, false, "0"},
		},
		[]qgen.DBTableKey{},
	))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleInsert("menus", "", ""))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleInsert("menu_items", "mid, htmlID, position, path, aria, tooltip, order", "1,'menu_forums','left','/forums/','{lang.menu_forums_aria}','{lang.menu_forums_tooltip}',0"))
	if err != nil {
		return err
	}

	err = execStmt(qgen.Builder.SimpleInsert("menu_items", "mid, htmlID, cssClass, position, path, aria, tooltip, order", "1,'menu_topics','menu_topics','left','/topics/','{lang.menu_topics_aria}','{lang.menu_topics_tooltip}',1"))
	if err != nil {
		return err
	}

	stmt, err = execStmt(qgen.Builder.SimpleInsert("menu_items", "mid, htmlID, cssClass, position, tmplName, order", "1,'general_alerts','menu_alerts','right','menu_alerts',2"))
	if err != nil {
		return err
	}

	return nil
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
