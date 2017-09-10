/* WIP Under Construction */
package qgen

//import "fmt"
import "strings"
import "strconv"
import "errors"

func init() {
	DB_Registry = append(DB_Registry,
		&Mysql_Adapter{Name: "mysql", Buffer: make(map[string]DB_Stmt)},
	)
}

type Mysql_Adapter struct {
	Name        string
	Buffer      map[string]DB_Stmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
}

func (adapter *Mysql_Adapter) GetName() string {
	return adapter.Name
}

func (adapter *Mysql_Adapter) GetStmt(name string) DB_Stmt {
	return adapter.Buffer[name]
}

func (adapter *Mysql_Adapter) GetStmts() map[string]DB_Stmt {
	return adapter.Buffer
}

func (adapter *Mysql_Adapter) CreateTable(name string, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (string, error) {
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
		// Make it easier to support Cassandra in the future
		if column.Type == "createdAt" {
			column.Type = "datetime"
		}

		var size string
		if column.Size > 0 {
			size = "(" + strconv.Itoa(column.Size) + ")"
		}

		var end string
		// TODO: Exclude the other variants of text like mediumtext and longtext too
		if column.Default != "" && column.Type != "text" {
			end = " DEFAULT "
			if adapter.stringyType(column.Type) && column.Default != "''" {
				end += "'" + column.Default + "'"
			} else {
				end += column.Default
			}
		}

		if column.Null {
			end += " null"
		} else {
			end += " not null"
		}

		if column.Auto_Increment {
			end += " AUTO_INCREMENT"
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

	querystr = querystr[0:len(querystr)-1] + "\n)"
	if charset != "" {
		querystr += " CHARSET=" + charset
	}
	if collation != "" {
		querystr += " COLLATE " + collation
	}

	adapter.pushStatement(name, "create-table", querystr+";")
	return querystr + ";", nil
}

func (adapter *Mysql_Adapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
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

	var querystr = "INSERT INTO `" + table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range _process_columns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}

	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += ") VALUES ("
	for _, field := range _processFields(fields) {
		querystr += field.Name + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	adapter.pushStatement(name, "insert", querystr+")")
	return querystr + ")", nil
}

func (adapter *Mysql_Adapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
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

	var querystr = "REPLACE INTO `" + table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range _process_columns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += ") VALUES ("
	for _, field := range _processFields(fields) {
		querystr += field.Name + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	adapter.pushStatement(name, "replace", querystr+")")
	return querystr + ")", nil
}

func (adapter *Mysql_Adapter) SimpleUpdate(name string, table string, set string, where string) (string, error) {
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
			case "function", "operator", "number", "substitute":
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
				case "function", "operator", "number", "substitute":
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

func (adapter *Mysql_Adapter) SimpleDelete(name string, table string, where string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}

	var querystr = "DELETE FROM `" + table + "` WHERE"

	// Add support for BETWEEN x.x
	for _, loc := range _processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case "function", "operator", "number", "substitute":
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

	querystr = strings.TrimSpace(querystr[0 : len(querystr)-4])
	adapter.pushStatement(name, "delete", querystr)
	return querystr, nil
}

// We don't want to accidentally wipe tables, so we'll have a seperate method for purging tables instead
func (adapter *Mysql_Adapter) Purge(name string, table string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	adapter.pushStatement(name, "purge", "DELETE FROM `"+table+"`")
	return "DELETE FROM `" + table + "`", nil
}

func (adapter *Mysql_Adapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}

	// Slice up the user friendly strings into something easier to process
	var colslice = strings.Split(strings.TrimSpace(columns), ",")

	var querystr = "SELECT "

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range colslice {
		querystr += "`" + strings.TrimSpace(column) + "`,"
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + table + "`"

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
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

	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if limit != "" {
		querystr += " LIMIT " + limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	var querystr = "SELECT "

	for _, column := range _process_columns(columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "`" + column.Table + "`.`" + column.Left + "`"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "`" + column.Left + "`"
		}

		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += source + alias + ","
	}

	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + table1 + "` LEFT JOIN `" + table2 + "` ON "
	for _, joiner := range _processJoiner(joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " `" + halves[0] + "`.`" + halves[1] + "`"
					} else {
						querystr += " `" + token.Contents + "`"
					}
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

	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if limit != "" {
		querystr += " LIMIT " + limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	var querystr = "SELECT "

	for _, column := range _process_columns(columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "`" + column.Table + "`.`" + column.Left + "`"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "`" + column.Left + "`"
		}

		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += source + alias + ","
	}

	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + table1 + "` INNER JOIN `" + table2 + "` ON "
	for _, joiner := range _processJoiner(joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " `" + halves[0] + "`.`" + halves[1] + "`"
					} else {
						querystr += " `" + token.Contents + "`"
					}
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

	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if limit != "" {
		querystr += " LIMIT " + limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleInsertSelect(name string, ins DB_Insert, sel DB_Select) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range _process_columns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range _process_columns(sel.Columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Type == "function" || column.Type == "substitute" {
			source = column.Left
		} else {
			source = "`" + column.Left + "`"
		}

		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += " " + source + alias + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + sel.Table + "`"

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
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

	if len(sel.Orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(sel.Orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if sel.Limit != "" {
		querystr += " LIMIT " + sel.Limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleInsertLeftJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range _process_columns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range _process_columns(sel.Columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "`" + column.Table + "`.`" + column.Left + "`"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "`" + column.Left + "`"
		}

		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += " " + source + alias + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + sel.Table1 + "` LEFT JOIN `" + sel.Table2 + "` ON "
	for _, joiner := range _processJoiner(sel.Joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " `" + halves[0] + "`.`" + halves[1] + "`"
					} else {
						querystr += " `" + token.Contents + "`"
					}
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

	if len(sel.Orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(sel.Orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if sel.Limit != "" {
		querystr += " LIMIT " + sel.Limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleInsertInnerJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range _process_columns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range _process_columns(sel.Columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "`" + column.Table + "`.`" + column.Left + "`"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "`" + column.Left + "`"
		}

		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += " " + source + alias + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM `" + sel.Table1 + "` INNER JOIN `" + sel.Table2 + "` ON "
	for _, joiner := range _processJoiner(sel.Joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range _processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " `" + halves[0] + "`.`" + halves[1] + "`"
					} else {
						querystr += " `" + token.Contents + "`"
					}
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

	if len(sel.Orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range _process_orderby(sel.Orderby) {
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if sel.Limit != "" {
		querystr += " LIMIT " + sel.Limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) SimpleCount(name string, table string, where string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "SELECT COUNT(*) AS `count` FROM `" + table + "`"

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		//fmt.Println("SimpleCount:",name)
		//fmt.Println("where:",where)
		//fmt.Println("_process_where:",_process_where(where))
		for _, loc := range _processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute":
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

	if limit != "" {
		querystr += " LIMIT " + limit
	}

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *Mysql_Adapter) Write() error {
	var stmts, body string
	for _, name := range adapter.BufferOrder {
		stmt := adapter.Buffer[name]
		// TODO: Add support for create-table? Table creation might be a little complex for Go to do outside a SQL file :(
		if stmt.Type != "create-table" {
			stmts += "var " + name + "_stmt *sql.Stmt\n"
			body += `	
	log.Print("Preparing ` + name + ` statement.")
	` + name + `_stmt, err = db.Prepare("` + stmt.Contents + `")
	if err != nil {
		return err
	}
	`
		}
	}

	out := `// +build !pgsql !sqlite !mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import "log"
import "database/sql"

// nolint
` + stmts + `
// nolint
func _gen_mysql() (err error) {
	if dev.DebugMode {
		log.Print("Building the generated statements")
	}
` + body + `
	return nil
}
`
	return writeFile("./gen_mysql.go", out)
}

// Internal methods, not exposed in the interface
func (adapter *Mysql_Adapter) pushStatement(name string, stype string, querystr string) {
	adapter.Buffer[name] = DB_Stmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *Mysql_Adapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "varchar" || ctype == "tinytext" || ctype == "text" || ctype == "mediumtext" || ctype == "longtext" || ctype == "char" || ctype == "datetime" || ctype == "timestamp" || ctype == "time" || ctype == "date"
}
