package main

import "database/sql"
import "github.com/Azareal/Gosora/query_gen"

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
	err := qgen.NewAcc().Select("users").Each(func(rows *sql.Rows) error {
		var uid int
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}
		return handle(uid)
	})
	return err
}
