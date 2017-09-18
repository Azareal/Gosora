/* WIP Under Really Heavy Construction */
package qgen

import "strings"
import "strconv"
import "errors"

func init() {
	DB_Registry = append(DB_Registry,
		&Pgsql_Adapter{Name: "pgsql", Buffer: make(map[string]DB_Stmt)},
	)
}

type Pgsql_Adapter struct {
	Name        string
	Buffer      map[string]DB_Stmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
}

func (adapter *Pgsql_Adapter) GetName() string {
	return adapter.Name
}

func (adapter *Pgsql_Adapter) GetStmt(name string) DB_Stmt {
	return adapter.Buffer[name]
}

func (adapter *Pgsql_Adapter) GetStmts() map[string]DB_Stmt {
	return adapter.Buffer
}

// TODO: Implement this
// We may need to change the CreateTable API to better suit PGSQL and the other database drivers which are coming up
func (adapter *Pgsql_Adapter) CreateTable(name string, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	var querystr = "CREATE TABLE `" + table + "` ("
	for _, column := range columns {
		if column.Auto_Increment {
			column.Type = "serial"
		} else if column.Type == "createdAt" {
			column.Type = "timestamp"
		} else if column.Type == "datetime" {
			column.Type = "timestamp"
		}

		var size string
		if column.Size > 0 {
			size = " (" + strconv.Itoa(column.Size) + ")"
		}

		var end string
		if column.Default != "" {
			end = " DEFAULT "
			if adapter.stringyType(column.Type) && column.Default != "''" {
				end += "'" + column.Default + "'"
			} else {
				end += column.Default
			}
		}

		if !column.Null {
			end += " not null"
		}

		querystr += "\n\t`" + column.Name + "` " + column.Type + size + end + ","
	}

	if len(keys) > 0 {
		for _, key := range keys {
			querystr += "\n\t" + key.Type
			if key.Type != "unique" {
				querystr += " key"
			}
			querystr += "("
			for _, column := range strings.Split(key.Columns, ",") {
				querystr += "`" + column + "`,"
			}
			querystr = querystr[0:len(querystr)-1] + "),"
		}
	}

	querystr = querystr[0:len(querystr)-1] + "\n);"
	adapter.pushStatement(name, "create-table", querystr)
	return querystr, nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleInsert")
	}
	if len(fields) == 0 {
		return "", errors.New("No input data found for SimpleInsert")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleInsert")
	}
	if len(fields) == 0 {
		return "", errors.New("No input data found for SimpleInsert")
	}
	return "", nil
}

// TODO: Implemented, but we need CreateTable and a better installer to *test* it
func (adapter *Pgsql_Adapter) SimpleUpdate(name string, table string, set string, where string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if set == "" {
		return "", errors.New("You need to set data in this update statement")
	}
	var querystr = "UPDATE `" + table + "` SET "
	for _, item := range _process_set(set) {
		querystr += "`" + item.Column + "` ="
		for _, token := range item.Expr {
			switch token.Type {
			case "function":
				// TODO: Write a more sophisticated function parser on the utils side.
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "LOCALTIMESTAMP()"
				}
				querystr += " " + token.Contents
			case "operator", "number", "substitute":
				querystr += " " + token.Contents
			case "column":
				querystr += " `" + token.Contents + "`"
			case "string":
				querystr += " '" + token.Contents + "'"
			}
		}
		querystr += ","
	}

	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function":
					// TODO: Write a more sophisticated function parser on the utils side. What's the situation in regards to case sensitivity?
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "LOCALTIMESTAMP()"
					}
					querystr += " " + token.Contents
				case "operator", "number", "substitute":
					querystr += " " + token.Contents
				case "column":
					querystr += " `" + token.Contents + "`"
				case "string":
					querystr += " '" + token.Contents + "'"
				default:
					panic("This token doesn't exist o_o")
				}
			}
			querystr += " AND"
		}
		querystr = querystr[0 : len(querystr)-4]
	}

	adapter.pushStatement(name, "update", querystr)
	return querystr, nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleDelete(name string, table string, where string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	return "", nil
}

// TODO: Implement this
// We don't want to accidentally wipe tables, so we'll have a seperate method for purging tables instead
func (adapter *Pgsql_Adapter) Purge(name string, table string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table1 == "" {
		return "", errors.New("You need a name for the left table")
	}
	if table2 == "" {
		return "", errors.New("You need a name for the right table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleLeftJoin")
	}
	if len(joiners) == 0 {
		return "", errors.New("No joiners found for SimpleLeftJoin")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table1 == "" {
		return "", errors.New("You need a name for the left table")
	}
	if table2 == "" {
		return "", errors.New("You need a name for the right table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleInnerJoin")
	}
	if len(joiners) == 0 {
		return "", errors.New("No joiners found for SimpleInnerJoin")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleInsertSelect(name string, ins DB_Insert, sel DB_Select) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleInsertLeftJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleInsertInnerJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *Pgsql_Adapter) SimpleCount(name string, table string, where string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

func (adapter *Pgsql_Adapter) Write() error {
	var stmts, body string
	for _, name := range adapter.BufferOrder {
		stmt := adapter.Buffer[name]
		// TODO: Add support for create-table? Table creation might be a little complex for Go to do outside a SQL file :(
		if stmt.Type != "create-table" {
			stmts += "var " + name + "Stmt *sql.Stmt\n"
			body += `	
	log.Print("Preparing ` + name + ` statement.")
	` + name + `Stmt, err = db.Prepare("` + stmt.Contents + `")
	if err != nil {
		return err
	}
	`
		}
	}

	out := `// +build pgsql

// This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time.
package main

import "log"
import "database/sql"

// nolint
` + stmts + `
// nolint
func _gen_pgsql() (err error) {
	if dev.DebugMode {
		log.Print("Building the generated statements")
	}
` + body + `
	return nil
}
`
	return writeFile("./gen_pgsql.go", out)
}

// Internal methods, not exposed in the interface
func (adapter *Pgsql_Adapter) pushStatement(name string, stype string, querystr string) {
	adapter.Buffer[name] = DB_Stmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *Pgsql_Adapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "char" || ctype == "varchar" || ctype == "timestamp" || ctype == "text"
}
