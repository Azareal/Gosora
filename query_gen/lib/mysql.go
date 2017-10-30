/* WIP Under Construction */
package qgen

//import "fmt"
import "strings"
import "strconv"
import "errors"

func init() {
	DB_Registry = append(DB_Registry,
		&MysqlAdapter{Name: "mysql", Buffer: make(map[string]DB_Stmt)},
	)
}

type MysqlAdapter struct {
	Name        string // ? - Do we really need this? Can't we hard-code this?
	Buffer      map[string]DB_Stmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
}

// GetName gives you the name of the database adapter. In this case, it's mysql
func (adapter *MysqlAdapter) GetName() string {
	return adapter.Name
}

func (adapter *MysqlAdapter) GetStmt(name string) DB_Stmt {
	return adapter.Buffer[name]
}

func (adapter *MysqlAdapter) GetStmts() map[string]DB_Stmt {
	return adapter.Buffer
}

func (adapter *MysqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (string, error) {
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

func (adapter *MysqlAdapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
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
	for _, column := range processColumns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}

	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += ") VALUES ("
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

	adapter.pushStatement(name, "insert", querystr+")")
	return querystr + ")", nil
}

// ! DEPRECATED
func (adapter *MysqlAdapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
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
	for _, column := range processColumns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += ") VALUES ("
	for _, field := range processFields(fields) {
		querystr += field.Name + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	adapter.pushStatement(name, "replace", querystr+")")
	return querystr + ")", nil
}

func (adapter *MysqlAdapter) SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error) {
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
	if where == "" {
		return "", errors.New("You need a where for this upsert")
	}

	var querystr = "INSERT INTO `" + table + "`("
	var parsedFields = processFields(fields)

	var insertColumns string
	var insertValues string
	var setBit = ") ON DUPLICATE KEY UPDATE "

	for columnID, column := range processColumns(columns) {
		field := parsedFields[columnID]
		if column.Type == "function" {
			insertColumns += column.Left + ","
			insertValues += field.Name + ","
			setBit += column.Left + " = " + field.Name + " AND "
		} else {
			insertColumns += "`" + column.Left + "`,"
			insertValues += field.Name + ","
			setBit += "`" + column.Left + "` = " + field.Name + " AND "
		}
	}
	insertColumns = insertColumns[0 : len(insertColumns)-1]
	insertValues = insertValues[0 : len(insertValues)-1]
	insertColumns += ") VALUES (" + insertValues
	setBit = setBit[0 : len(setBit)-5]

	querystr += insertColumns + setBit

	adapter.pushStatement(name, "upsert", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleUpdate(name string, table string, set string, where string) (string, error) {
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
	for _, item := range processSet(set) {
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
		for _, loc := range processWhere(where) {
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

func (adapter *MysqlAdapter) SimpleDelete(name string, table string, where string) (string, error) {
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
	for _, loc := range processWhere(where) {
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

// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (adapter *MysqlAdapter) Purge(name string, table string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	adapter.pushStatement(name, "purge", "DELETE FROM `"+table+"`")
	return "DELETE FROM `" + table + "`", nil
}

func (adapter *MysqlAdapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
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
		for _, loc := range processWhere(where) {
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
		for _, column := range processOrderby(orderby) {
			// TODO: We might want to escape this column
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

func (adapter *MysqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	for _, column := range processColumns(columns) {
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
	for _, joiner := range processJoiner(joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
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
		for _, column := range processOrderby(orderby) {
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

func (adapter *MysqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	for _, column := range processColumns(columns) {
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
	for _, joiner := range processJoiner(joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
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
		for _, column := range processOrderby(orderby) {
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

func (adapter *MysqlAdapter) SimpleInsertSelect(name string, ins DB_Insert, sel DB_Select) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range processColumns(sel.Columns) {
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
		for _, loc := range processWhere(sel.Where) {
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
		for _, column := range processOrderby(sel.Orderby) {
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

func (adapter *MysqlAdapter) SimpleInsertLeftJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range processColumns(sel.Columns) {
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
	for _, joiner := range processJoiner(sel.Joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(sel.Where) {
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
		for _, column := range processOrderby(sel.Orderby) {
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

func (adapter *MysqlAdapter) SimpleInsertInnerJoin(name string, ins DB_Insert, sel DB_Join) (string, error) {
	/* Insert Portion */

	var querystr = "INSERT INTO `" + ins.Table + "`("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT"

	/* Select Portion */

	for _, column := range processColumns(sel.Columns) {
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
	for _, joiner := range processJoiner(sel.Joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(sel.Where) {
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
		for _, column := range processOrderby(sel.Orderby) {
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

func (adapter *MysqlAdapter) SimpleCount(name string, table string, where string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "SELECT COUNT(*) AS `count` FROM `" + table + "`"

	// TODO: Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		//log.Print("SimpleCount: ", name)
		//log.Print("where: ", where)
		//log.Print("processWhere: ", processWhere(where))
		for _, loc := range processWhere(where) {
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

func (adapter *MysqlAdapter) Write() error {
	var stmts, body string
	for _, name := range adapter.BufferOrder {
		if name[0] == '_' {
			continue
		}
		stmt := adapter.Buffer[name]
		// ? - Table creation might be a little complex for Go to do outside a SQL file :(
		if stmt.Type == "upsert" {
			stmts += "var " + name + "Stmt *qgen.MySQLUpsertCallback\n"
			body += `	
	log.Print("Preparing ` + name + ` statement.")
	` + name + `Stmt, err = qgen.PrepareMySQLUpsertCallback(db, "` + stmt.Contents + `")
	if err != nil {
		return err
	}
	`
		} else if stmt.Type != "create-table" {
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

	out := `// +build !pgsql, !sqlite, !mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import "log"
import "database/sql"
//import "./query_gen/lib"

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
func (adapter *MysqlAdapter) pushStatement(name string, stype string, querystr string) {
	adapter.Buffer[name] = DB_Stmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *MysqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "varchar" || ctype == "tinytext" || ctype == "text" || ctype == "mediumtext" || ctype == "longtext" || ctype == "char" || ctype == "datetime" || ctype == "timestamp" || ctype == "time" || ctype == "date"
}
