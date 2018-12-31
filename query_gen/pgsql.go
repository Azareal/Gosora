/* WIP Under Really Heavy Construction */
package qgen

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

func init() {
	Registry = append(Registry,
		&PgsqlAdapter{Name: "pgsql", Buffer: make(map[string]DBStmt)},
	)
}

type PgsqlAdapter struct {
	Name        string // ? - Do we really need this? Can't we hard-code this?
	Buffer      map[string]DBStmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
}

// GetName gives you the name of the database adapter. In this case, it's pgsql
func (adapter *PgsqlAdapter) GetName() string {
	return adapter.Name
}

func (adapter *PgsqlAdapter) GetStmt(name string) DBStmt {
	return adapter.Buffer[name]
}

func (adapter *PgsqlAdapter) GetStmts() map[string]DBStmt {
	return adapter.Buffer
}

// TODO: Implement this
func (adapter *PgsqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
	return nil, nil
}

func (adapter *PgsqlAdapter) DbVersion() string {
	return "SELECT version()"
}

func (adapter *PgsqlAdapter) DropTable(name string, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	querystr := "DROP TABLE IF EXISTS \"" + table + "\";"
	adapter.pushStatement(name, "drop-table", querystr)
	return querystr, nil
}

// TODO: Implement this
// We may need to change the CreateTable API to better suit PGSQL and the other database drivers which are coming up
func (adapter *PgsqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	var querystr = "CREATE TABLE \"" + table + "\" ("
	for _, column := range columns {
		if column.AutoIncrement {
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
func (adapter *PgsqlAdapter) AddColumn(name string, table string, column DBTableColumn) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

// TODO: Implement this
// TODO: Test to make sure everything works here
func (adapter *PgsqlAdapter) AddIndex(name string, table string, iname string, colname string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if iname == "" {
		return "", errors.New("You need a name for the index")
	}
	if colname == "" {
		return "", errors.New("You need a name for the column")
	}
	return "", errors.New("not implemented")
}

// TODO: Test this
// ! We need to get the last ID out of this somehow, maybe add returning to every query? Might require some sort of wrapper over the sql statements
func (adapter *PgsqlAdapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "INSERT INTO \"" + table + "\"("
	if columns != "" {
		querystr += adapter.buildColumns(columns) + ") VALUES ("
		for _, field := range processFields(fields) {
			nameLen := len(field.Name)
			if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
				field.Name = "'" + field.Name[1:nameLen-1] + "'"
			}
			if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
				field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
			}
			querystr += field.Name + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	} else {
		querystr += ") VALUES ("
	}
	querystr += ")"

	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *PgsqlAdapter) buildColumns(columns string) (querystr string) {
	if columns == "" {
		return ""
	}
	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "\"" + column.Left + "\","
		}
	}
	return querystr[0 : len(querystr)-1]
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
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
func (adapter *PgsqlAdapter) SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error) {
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
func (adapter *PgsqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	var querystr = "UPDATE \"" + up.table + "\" SET "
	for _, item := range processSet(up.set) {
		querystr += "`" + item.Column + "` ="
		for _, token := range item.Expr {
			switch token.Type {
			case "function":
				// TODO: Write a more sophisticated function parser on the utils side.
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "LOCALTIMESTAMP()"
				}
				querystr += " " + token.Contents
			case "operator", "number", "substitute", "or":
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
	if len(up.where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(up.where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function":
					// TODO: Write a more sophisticated function parser on the utils side. What's the situation in regards to case sensitivity?
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "LOCALTIMESTAMP()"
					}
					querystr += " " + token.Contents
				case "operator", "number", "substitute", "or":
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

	adapter.pushStatement(up.name, "update", querystr)
	return querystr, nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleUpdateSelect(up *updatePrebuilder) (string, error) {
	return "", errors.New("not implemented")
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleDelete(name string, table string, where string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	return "", nil
}

// TODO: Implement this
// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (adapter *PgsqlAdapter) Purge(name string, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) ComplexSelect(prebuilder *selectPrebuilder) (string, error) {
	if prebuilder.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(prebuilder.columns) == 0 {
		return "", errors.New("No columns found for ComplexSelect")
	}
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
func (adapter *PgsqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
func (adapter *PgsqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return "", nil
}

// TODO: Implement this
func (adapter *PgsqlAdapter) SimpleCount(name string, table string, where string, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

func (adapter *PgsqlAdapter) Builder() *prebuilder {
	return &prebuilder{adapter}
}

func (adapter *PgsqlAdapter) Write() error {
	var stmts, body string
	for _, name := range adapter.BufferOrder {
		if name[0] == '_' {
			continue
		}
		stmt := adapter.Buffer[name]
		// TODO: Add support for create-table? Table creation might be a little complex for Go to do outside a SQL file :(
		if stmt.Type != "create-table" {
			stmts += "\t" + name + " *sql.Stmt\n"
			body += `	
	common.DebugLog("Preparing ` + name + ` statement.")
	stmts.` + name + `, err = db.Prepare("` + strings.Replace(stmt.Contents, "\"", "\\\"", -1) + `")
	if err != nil {
		log.Print("Error in ` + name + ` statement.")
		return err
	}
	`
		}
	}

	// TODO: Move these custom queries out of this file
	out := `// +build pgsql

// This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time.
package main

import "log"
import "database/sql"
import "github.com/Azareal/Gosora/common"

// nolint
type Stmts struct {
` + stmts + `
	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt

	Mocks bool
}

// nolint
func _gen_pgsql() (err error) {
	common.DebugLog("Building the generated statements")
` + body + `
	return nil
}
`
	return writeFile("./gen_pgsql.go", out)
}

// Internal methods, not exposed in the interface
func (adapter *PgsqlAdapter) pushStatement(name string, stype string, querystr string) {
	if name == "" {
		return
	}
	adapter.Buffer[name] = DBStmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *PgsqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "char" || ctype == "varchar" || ctype == "timestamp" || ctype == "text"
}
