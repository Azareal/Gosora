package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"

	"../query_gen/lib"
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

	err := patcher(scanner)
	if err != nil {
		fmt.Println(err)
	}
}

func pressAnyKey(scanner *bufio.Scanner) {
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}

func patcher(scanner *bufio.Scanner) error {
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

func eachUser(handle func(int)) error {
	stmt, err := qgen.Builder.Select("users").Prepare()
	if err != nil {
		return err
	}

	rows, err := stmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uid int
		err := rows.Scan(&uid)
		if err != nil {
			return err
		}
		err = handle(uid)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}
