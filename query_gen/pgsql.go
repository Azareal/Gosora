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
func (a *PgsqlAdapter) GetName() string {
	return a.Name
}

func (a *PgsqlAdapter) GetStmt(name string) DBStmt {
	return a.Buffer[name]
}

func (a *PgsqlAdapter) GetStmts() map[string]DBStmt {
	return a.Buffer
}

// TODO: Implement this
func (a *PgsqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
	return nil, nil
}

func (a *PgsqlAdapter) DbVersion() string {
	return "SELECT version()"
}

func (a *PgsqlAdapter) DropTable(name, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "DROP TABLE IF EXISTS \"" + table + "\";"
	a.pushStatement(name, "drop-table", q)
	return q, nil
}

// TODO: Implement this
// We may need to change the CreateTable API to better suit PGSQL and the other database drivers which are coming up
func (a *PgsqlAdapter) CreateTable(name, table, charset, collation string, cols []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(cols) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	q := "CREATE TABLE \"" + table + "\" ("
	for _, col := range cols {
		if col.AutoIncrement {
			col.Type = "serial"
		} else if col.Type == "createdAt" {
			col.Type = "timestamp"
		} else if col.Type == "datetime" {
			col.Type = "timestamp"
		}

		var size string
		if col.Size > 0 {
			size = " (" + strconv.Itoa(col.Size) + ")"
		}

		var end string
		if col.Default != "" {
			end = " DEFAULT "
			if a.stringyType(col.Type) && col.Default != "''" {
				end += "'" + col.Default + "'"
			} else {
				end += col.Default
			}
		}
		if !col.Null {
			end += " not null"
		}

		q += "\n\t`" + col.Name + "` " + col.Type + size + end + ","
	}

	if len(keys) > 0 {
		for _, key := range keys {
			q += "\n\t" + key.Type
			if key.Type != "unique" {
				q += " key"
			}
			q += "("
			for _, column := range strings.Split(key.Columns, ",") {
				q += "`" + column + "`,"
			}
			q = q[0:len(q)-1] + "),"
		}
	}

	q = q[0:len(q)-1] + "\n);"
	a.pushStatement(name, "create-table", q)
	return q, nil
}

// TODO: Implement this
func (a *PgsqlAdapter) AddColumn(name, table string, column DBTableColumn, key *DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) DropColumn(name, table, colName string) (string, error) {
	return "", errors.New("not implemented")
}

// TODO: Implement this
func (a *PgsqlAdapter) RenameColumn(name, table, oldName, newName string) (string, error) {
	return "", errors.New("not implemented")
}

// TODO: Implement this
func (a *PgsqlAdapter) ChangeColumn(name, table, colName string, col DBTableColumn) (string, error) {
	return "", errors.New("not implemented")
}

// TODO: Implement this
func (a *PgsqlAdapter) SetDefaultColumn(name, table, colName, colType, defaultStr string) (string, error) {
	if colType == "text" {
		return "", errors.New("text fields cannot have default values")
	}
	return "", errors.New("not implemented")
}

// TODO: Implement this
// TODO: Test to make sure everything works here
func (a *PgsqlAdapter) AddIndex(name, table, iname, colname string) (string, error) {
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

// TODO: Implement this
// TODO: Test to make sure everything works here
func (a *PgsqlAdapter) AddKey(name, table, column string, key DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if column == "" {
		return "", errors.New("You need a name for the column")
	}
	return "", errors.New("not implemented")
}

// TODO: Implement this
// TODO: Test to make sure everything works here
func (a *PgsqlAdapter) RemoveIndex(name, table, iname string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if iname == "" {
		return "", errors.New("You need a name for the index")
	}
	return "", errors.New("not implemented")
}

// TODO: Implement this
// TODO: Test to make sure everything works here
func (a *PgsqlAdapter) AddForeignKey(name, table, column, ftable, fcolumn string, cascade bool) (out string, e error) {
	var c = func(str string, val bool) {
		if e != nil || !val {
			return
		}
		e = errors.New("You need a " + str + " for this table")
	}
	c("name", table == "")
	c("column", column == "")
	c("ftable", ftable == "")
	c("fcolumn", fcolumn == "")
	if e != nil {
		return "", e
	}
	return "", errors.New("not implemented")
}

// TODO: Test this
// ! We need to get the last ID out of this somehow, maybe add returning to every query? Might require some sort of wrapper over the sql statements
func (a *PgsqlAdapter) SimpleInsert(name, table, columns, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	q := "INSERT INTO \"" + table + "\"("
	if columns != "" {
		q += a.buildColumns(columns) + ") VALUES ("
		for _, field := range processFields(fields) {
			nameLen := len(field.Name)
			if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
				field.Name = "'" + field.Name[1:nameLen-1] + "'"
			}
			if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
				field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
			}
			q += field.Name + ","
		}
		q = q[0 : len(q)-1]
	} else {
		q += ") VALUES ("
	}
	q += ")"

	a.pushStatement(name, "insert", q)
	return q, nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleBulkInsert(name, table, columns string, fieldSet []string) (string, error) {
	return "", nil
}

func (a *PgsqlAdapter) buildColumns(cols string) (q string) {
	if cols == "" {
		return ""
	}
	// Escape the column names, just in case we've used a reserved keyword
	for _, col := range processColumns(cols) {
		if col.Type == TokenFunc {
			q += col.Left + ","
		} else {
			q += "\"" + col.Left + "\","
		}
	}
	return q[0 : len(q)-1]
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleReplace(name, table, columns, fields string) (string, error) {
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
func (a *PgsqlAdapter) SimpleUpsert(name, table, columns, fields, where string) (string, error) {
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
func (a *PgsqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	q := "UPDATE \"" + up.table + "\" SET "
	for _, item := range processSet(up.set) {
		q += "`" + item.Column + "`="
		for _, token := range item.Expr {
			switch token.Type {
			case TokenFunc:
				// TODO: Write a more sophisticated function parser on the utils side.
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "LOCALTIMESTAMP()"
				}
				q += " " + token.Contents
			case TokenOp, TokenNumber, TokenSub, TokenOr:
				q += " " + token.Contents
			case TokenColumn:
				q += " `" + token.Contents + "`"
			case TokenString:
				q += " '" + token.Contents + "'"
			}
		}
		q += ","
	}
	q = q[0 : len(q)-1]

	// Add support for BETWEEN x.x
	if len(up.where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(up.where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc:
					// TODO: Write a more sophisticated function parser on the utils side. What's the situation in regards to case sensitivity?
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "LOCALTIMESTAMP()"
					}
					q += " " + token.Contents
				case TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					q += " " + token.Contents
				case TokenColumn:
					q += " `" + token.Contents + "`"
				case TokenString:
					q += " '" + token.Contents + "'"
				default:
					panic("This token doesn't exist o_o")
				}
			}
			q += " AND"
		}
		q = q[0 : len(q)-4]
	}

	a.pushStatement(up.name, "update", q)
	return q, nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleUpdateSelect(up *updatePrebuilder) (string, error) {
	return "", errors.New("not implemented")
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleDelete(name, table, where string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) ComplexDelete(b *deletePrebuilder) (string, error) {
	if b.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if b.where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	return "", nil
}

// TODO: Implement this
// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (a *PgsqlAdapter) Purge(name, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleSelect(name, table, columns, where, orderby, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) ComplexSelect(prebuilder *selectPrebuilder) (string, error) {
	if prebuilder.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(prebuilder.columns) == 0 {
		return "", errors.New("No columns found for ComplexSelect")
	}
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleLeftJoin(name, table1, table2, columns, joiners, where, orderby, limit string) (string, error) {
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
func (a *PgsqlAdapter) SimpleInnerJoin(name, table1, table2, columns, joiners, where, orderby, limit string) (string, error) {
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
func (a *PgsqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return "", nil
}

// TODO: Implement this
func (a *PgsqlAdapter) SimpleCount(name, table, where, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	return "", nil
}

func (a *PgsqlAdapter) Builder() *prebuilder {
	return &prebuilder{a}
}

func (a *PgsqlAdapter) Write() error {
	var stmts, body string
	for _, name := range a.BufferOrder {
		if name[0] == '_' {
			continue
		}
		stmt := a.Buffer[name]
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
func (a *PgsqlAdapter) pushStatement(name, stype, q string) {
	if name == "" {
		return
	}
	a.Buffer[name] = DBStmt{q, stype}
	a.BufferOrder = append(a.BufferOrder, name)
}

func (a *PgsqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "char" || ctype == "varchar" || ctype == "timestamp" || ctype == "text"
}
