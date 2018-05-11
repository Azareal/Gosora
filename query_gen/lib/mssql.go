/* WIP Under Really Heavy Construction */
package qgen

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
)

func init() {
	Registry = append(Registry,
		&MssqlAdapter{Name: "mssql", Buffer: make(map[string]DBStmt)},
	)
}

type MssqlAdapter struct {
	Name        string // ? - Do we really need this? Can't we hard-code this?
	Buffer      map[string]DBStmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
	keys        map[string]string
}

// GetName gives you the name of the database adapter. In this case, it's Mssql
func (adapter *MssqlAdapter) GetName() string {
	return adapter.Name
}

func (adapter *MssqlAdapter) GetStmt(name string) DBStmt {
	return adapter.Buffer[name]
}

func (adapter *MssqlAdapter) GetStmts() map[string]DBStmt {
	return adapter.Buffer
}

// TODO: Implement this
func (adapter *MssqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
	return nil, nil
}

func (adapter *MssqlAdapter) DbVersion() string {
	return "SELECT CONCAT(SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition'))"
}

func (adapter *MssqlAdapter) DropTable(name string, table string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	querystr := "DROP TABLE IF EXISTS [" + table + "];"
	adapter.pushStatement(name, "drop-table", querystr)
	return querystr, nil
}

// TODO: Convert any remaining stringy types to nvarchar
// We may need to change the CreateTable API to better suit Mssql and the other database drivers which are coming up
func (adapter *MssqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	var querystr = "CREATE TABLE [" + table + "] ("
	for _, column := range columns {
		var max bool
		var createdAt bool
		switch column.Type {
		case "createdAt":
			column.Type = "datetime"
			createdAt = true
		case "varchar":
			column.Type = "nvarchar"
		case "text":
			column.Type = "nvarchar"
			max = true
		case "json":
			column.Type = "nvarchar"
			max = true
		case "boolean":
			column.Type = "bit"
		}

		var size string
		if column.Size > 0 {
			size = " (" + strconv.Itoa(column.Size) + ")"
		}
		if max {
			size = " (MAX)"
		}

		var end string
		if column.Default != "" {
			end = " DEFAULT "
			if createdAt {
				end += "GETUTCDATE()" // TODO: Use GETUTCDATE() in updates instead of the neutral format
			} else if adapter.stringyType(column.Type) && column.Default != "''" {
				end += "'" + column.Default + "'"
			} else {
				end += column.Default
			}
		}

		if !column.Null {
			end += " not null"
		}

		// ! Not exactly the meaning of auto increment...
		if column.AutoIncrement {
			end += " IDENTITY"
		}

		querystr += "\n\t[" + column.Name + "] " + column.Type + size + end + ","
	}

	if len(keys) > 0 {
		for _, key := range keys {
			querystr += "\n\t" + key.Type
			if key.Type != "unique" {
				querystr += " key"
			}
			querystr += "("
			for _, column := range strings.Split(key.Columns, ",") {
				querystr += "[" + column + "],"
			}
			querystr = querystr[0:len(querystr)-1] + "),"
		}
	}

	querystr = querystr[0:len(querystr)-1] + "\n);"
	adapter.pushStatement(name, "create-table", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "INSERT INTO [" + table + "] ("
	if columns == "" {
		querystr += ") VALUES ()"
		adapter.pushStatement(name, "insert", querystr)
		return querystr, nil
	}

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "[" + column.Left + "],"
		}
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += ") VALUES ("
	for _, field := range processFields(fields) {
		field.Name = strings.Replace(field.Name, "UTC_TIMESTAMP()", "GETUTCDATE()", -1)
		//log.Print("field.Name ", field.Name)
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
func (adapter *MssqlAdapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
	log.Print("In SimpleReplace")
	key, ok := adapter.keys[table]
	if !ok {
		return "", errors.New("Unable to elide key from table '" + table + "', please use SimpleUpsert (coming soon!) instead")
	}
	log.Print("After the key check")

	// Escape the column names, just in case we've used a reserved keyword
	var keyPosition int
	for _, column := range processColumns(columns) {
		if column.Left == key {
			continue
		}
		keyPosition++
	}

	var keyValue string
	for fieldID, field := range processFields(fields) {
		field.Name = strings.Replace(field.Name, "UTC_TIMESTAMP()", "GETUTCDATE()", -1)
		nameLen := len(field.Name)
		if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
			field.Name = "'" + field.Name[1:nameLen-1] + "'"
		}
		if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
			field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
		}
		if keyPosition == fieldID {
			keyValue = field.Name
			continue
		}
	}
	return adapter.SimpleUpsert(name, table, columns, fields, "key = "+keyValue)
}

func (adapter *MssqlAdapter) SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error) {
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

	var fieldCount int
	var fieldOutput string
	var querystr = "MERGE [" + table + "] WITH(HOLDLOCK) as t1 USING (VALUES("

	var parsedFields = processFields(fields)
	for _, field := range parsedFields {
		fieldCount++
		field.Name = strings.Replace(field.Name, "UTC_TIMESTAMP()", "GETUTCDATE()", -1)
		//log.Print("field.Name ", field.Name)
		nameLen := len(field.Name)
		if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
			field.Name = "'" + field.Name[1:nameLen-1] + "'"
		}
		if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
			field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
		}
		fieldOutput += field.Name + ","
	}
	fieldOutput = fieldOutput[0 : len(fieldOutput)-1]
	querystr += fieldOutput + ")) AS updates ("

	// nolint The linter wants this to be less readable
	for fieldID, _ := range parsedFields {
		querystr += "f" + strconv.Itoa(fieldID) + ","
	}
	querystr = querystr[0 : len(querystr)-1]
	querystr += ") ON "

	//querystr += "t1.[" + key + "] = "
	// Add support for BETWEEN x.x
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case "substitute":
				querystr += " ?"
			case "function", "operator", "number", "or":
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				querystr += " " + token.Contents
			case "column":
				querystr += " [" + token.Contents + "]"
			case "string":
				querystr += " '" + token.Contents + "'"
			default:
				panic("This token doesn't exist o_o")
			}
		}
	}

	var matched = " WHEN MATCHED THEN UPDATE SET "
	var notMatched = "WHEN NOT MATCHED THEN INSERT("
	var fieldList string

	// Escape the column names, just in case we've used a reserved keyword
	for columnID, column := range processColumns(columns) {
		fieldList += "f" + strconv.Itoa(columnID) + ","
		if column.Type == "function" {
			matched += column.Left + " = f" + strconv.Itoa(columnID) + ","
			notMatched += column.Left + ","
		} else {
			matched += "[" + column.Left + "] = f" + strconv.Itoa(columnID) + ","
			notMatched += "[" + column.Left + "],"
		}
	}

	matched = matched[0 : len(matched)-1]
	notMatched = notMatched[0 : len(notMatched)-1]
	fieldList = fieldList[0 : len(fieldList)-1]

	notMatched += ") VALUES (" + fieldList + ");"
	querystr += matched + " " + notMatched

	// TODO: Run this on debug mode?
	if name[0] == '_' {
		log.Print(name+" query: ", querystr)
	}
	adapter.pushStatement(name, "upsert", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) SimpleUpdate(name string, table string, set string, where string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	var querystr = "UPDATE [" + table + "] SET "
	for _, item := range processSet(set) {
		querystr += "[" + item.Column + "] ="
		for _, token := range item.Expr {
			switch token.Type {
			case "substitute":
				querystr += " ?"
			case "function", "operator", "number", "or":
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				querystr += " " + token.Contents
			case "column":
				querystr += " [" + token.Contents + "]"
			case "string":
				querystr += " '" + token.Contents + "'"
			default:
				panic("This token doesn't exist o_o")
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
				case "function", "operator", "number", "substitute", "or":
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					querystr += " [" + token.Contents + "]"
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

func (adapter *MssqlAdapter) SimpleDelete(name string, table string, where string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}

	var querystr = "DELETE FROM [" + table + "] WHERE"

	// Add support for BETWEEN x.x
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case "substitute":
				querystr += " ?"
			case "function", "operator", "number", "or":
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				querystr += " " + token.Contents
			case "column":
				querystr += " [" + token.Contents + "]"
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
func (adapter *MssqlAdapter) Purge(name string, table string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	adapter.pushStatement(name, "purge", "DELETE FROM ["+table+"]")
	return "DELETE FROM [" + table + "]", nil
}

func (adapter *MssqlAdapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(orderby) == 0 && limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	var substituteCount = 0
	var querystr = ""

	// Escape the column names, just in case we've used a reserved keyword
	var colslice = strings.Split(strings.TrimSpace(columns), ",")
	for _, column := range colslice {
		querystr += "[" + strings.TrimSpace(column) + "],"
	}
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM [" + table + "]"

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "substitute":
					substituteCount++
					querystr += " ?" + strconv.Itoa(substituteCount)
				case "function", "operator", "number", "or":
					// TODO: Split the function case off to speed things up
					// MSSQL seems to convert the formats? so we'll compare it with a regular date. Do this with the other methods too?
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					querystr += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			// TODO: We might want to escape this column
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}

	if limit != "" {
		limiter := processLimit(limit)
		log.Printf("limiter: %+v\n", limiter)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				substituteCount++
				querystr += " OFFSET ?" + strconv.Itoa(substituteCount) + " ROWS"
			} else {
				querystr += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				substituteCount++
				limiter.MaxCount = "?" + strconv.Itoa(substituteCount)
			}
			querystr += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	querystr = strings.TrimSpace("SELECT " + querystr)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", querystr)
	}
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

// TODO: ComplexSelect
func (adapter *MssqlAdapter) ComplexSelect(preBuilder *selectPrebuilder) (string, error) {
	return "", nil
}

func (adapter *MssqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(orderby) == 0 && limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}
	var substituteCount = 0
	var querystr = ""

	for _, column := range processColumns(columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "[" + column.Table + "].[" + column.Left + "]"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "[" + column.Left + "]"
		}

		if column.Alias != "" {
			alias = " AS '" + column.Alias + "'"
		}
		querystr += source + alias + ","
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM [" + table1 + "] LEFT JOIN [" + table2 + "] ON "
	for _, joiner := range processJoiner(joiners) {
		querystr += "[" + joiner.LeftTable + "].[" + joiner.LeftColumn + "] " + joiner.Operator + " [" + joiner.RightTable + "].[" + joiner.RightColumn + "] AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "substitute":
					substituteCount++
					querystr += " ?" + strconv.Itoa(substituteCount)
				case "function", "operator", "number", "or":
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						querystr += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	} else if limit != "" {
		key, ok := adapter.keys[table1]
		if ok {
			querystr += " ORDER BY [" + table1 + "].[" + key + "]"
		}
	}

	if limit != "" {
		limiter := processLimit(limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				substituteCount++
				querystr += " OFFSET ?" + strconv.Itoa(substituteCount) + " ROWS"
			} else {
				querystr += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				substituteCount++
				limiter.MaxCount = "?" + strconv.Itoa(substituteCount)
			}
			querystr += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	querystr = strings.TrimSpace("SELECT " + querystr)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", querystr)
	}
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(orderby) == 0 && limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	var substituteCount = 0
	var querystr = ""

	for _, column := range processColumns(columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "[" + column.Table + "].[" + column.Left + "]"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "[" + column.Left + "]"
		}

		if column.Alias != "" {
			alias = " AS '" + column.Alias + "'"
		}
		querystr += source + alias + ","
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM [" + table1 + "] INNER JOIN [" + table2 + "] ON "
	for _, joiner := range processJoiner(joiners) {
		querystr += "[" + joiner.LeftTable + "].[" + joiner.LeftColumn + "] " + joiner.Operator + " [" + joiner.RightTable + "].[" + joiner.RightColumn + "] AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "substitute":
					substituteCount++
					querystr += " ?" + strconv.Itoa(substituteCount)
				case "function", "operator", "number", "or":
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						querystr += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	} else if limit != "" {
		key, ok := adapter.keys[table1]
		if ok {
			log.Print("key: ", key)
			querystr += " ORDER BY [" + table1 + "].[" + key + "]"
		}
	}

	if limit != "" {
		limiter := processLimit(limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				substituteCount++
				querystr += " OFFSET ?" + strconv.Itoa(substituteCount) + " ROWS"
			} else {
				querystr += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				substituteCount++
				limiter.MaxCount = "?" + strconv.Itoa(substituteCount)
			}
			querystr += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	querystr = strings.TrimSpace("SELECT " + querystr)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", querystr)
	}
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	// TODO: More errors.
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(sel.Orderby) == 0 && sel.Limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	/* Insert */
	var querystr = "INSERT INTO [" + ins.Table + "] ("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "[" + column.Left + "],"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT "

	/* Select */
	var substituteCount = 0

	for _, column := range processColumns(sel.Columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Type == "function" || column.Type == "substitute" {
			source = column.Left
		} else {
			source = "[" + column.Left + "]"
		}

		if column.Alias != "" {
			alias = " AS [" + column.Alias + "]"
		}
		querystr += " " + source + alias + ","
	}
	querystr = querystr[0 : len(querystr)-1]
	querystr += " FROM [" + sel.Table + "] "

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "substitute":
					substituteCount++
					querystr += " ?" + strconv.Itoa(substituteCount)
				case "function", "operator", "number", "or":
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					querystr += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(sel.Orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range processOrderby(sel.Orderby) {
			// TODO: We might want to escape this column
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	} else if sel.Limit != "" {
		key, ok := adapter.keys[sel.Table]
		if ok {
			querystr += " ORDER BY [" + sel.Table + "].[" + key + "]"
		}
	}

	if sel.Limit != "" {
		limiter := processLimit(sel.Limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				substituteCount++
				querystr += " OFFSET ?" + strconv.Itoa(substituteCount) + " ROWS"
			} else {
				querystr += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				substituteCount++
				limiter.MaxCount = "?" + strconv.Itoa(substituteCount)
			}
			querystr += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	querystr = strings.TrimSpace(querystr)
	// TODO: Run this on debug mode?
	if name[0] == '_' && sel.Limit == "" {
		log.Print(name+" query: ", querystr)
	}

	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) simpleJoin(name string, ins DBInsert, sel DBJoin, joinType string) (string, error) {
	// TODO: More errors.
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(sel.Orderby) == 0 && sel.Limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	/* Insert */
	var querystr = "INSERT INTO [" + ins.Table + "] ("

	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(ins.Columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "[" + column.Left + "],"
		}
	}
	querystr = querystr[0:len(querystr)-1] + ") SELECT "

	/* Select */
	var substituteCount = 0

	for _, column := range processColumns(sel.Columns) {
		var source, alias string

		// Escape the column names, just in case we've used a reserved keyword
		if column.Table != "" {
			source = "[" + column.Table + "].[" + column.Left + "]"
		} else if column.Type == "function" {
			source = column.Left
		} else {
			source = "[" + column.Left + "]"
		}

		if column.Alias != "" {
			alias = " AS '" + column.Alias + "'"
		}
		querystr += source + alias + ","
	}
	// Remove the trailing comma
	querystr = querystr[0 : len(querystr)-1]

	querystr += " FROM [" + sel.Table1 + "] " + joinType + " JOIN [" + sel.Table2 + "] ON "
	for _, joiner := range processJoiner(sel.Joiners) {
		querystr += "[" + joiner.LeftTable + "].[" + joiner.LeftColumn + "] " + joiner.Operator + " [" + joiner.RightTable + "].[" + joiner.RightColumn + "] AND "
	}
	// Remove the trailing AND
	querystr = querystr[0 : len(querystr)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "substitute":
					substituteCount++
					querystr += " ?" + strconv.Itoa(substituteCount)
				case "function", "operator", "number", "or":
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						querystr += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						querystr += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(sel.Orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range processOrderby(sel.Orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			querystr += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	} else if sel.Limit != "" {
		key, ok := adapter.keys[sel.Table1]
		if ok {
			querystr += " ORDER BY [" + sel.Table1 + "].[" + key + "]"
		}
	}

	if sel.Limit != "" {
		limiter := processLimit(sel.Limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				substituteCount++
				querystr += " OFFSET ?" + strconv.Itoa(substituteCount) + " ROWS"
			} else {
				querystr += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				substituteCount++
				limiter.MaxCount = "?" + strconv.Itoa(substituteCount)
			}
			querystr += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	querystr = strings.TrimSpace(querystr)
	// TODO: Run this on debug mode?
	if name[0] == '_' && sel.Limit == "" {
		log.Print(name+" query: ", querystr)
	}

	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *MssqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return adapter.simpleJoin(name, ins, sel, "LEFT")
}

func (adapter *MssqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return adapter.simpleJoin(name, ins, sel, "INNER")
}

func (adapter *MssqlAdapter) SimpleCount(name string, table string, where string, limit string) (string, error) {
	if name == "" {
		return "", errors.New("You need a name for this statement")
	}
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "SELECT COUNT(*) AS [count] FROM [" + table + "]"

	// TODO: Add support for BETWEEN x.x
	if len(where) != 0 {
		querystr += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute", "or":
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					querystr += " " + token.Contents
				case "column":
					querystr += " [" + token.Contents + "]"
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

func (adapter *MssqlAdapter) Builder() *prebuilder {
	return &prebuilder{adapter}
}

func (adapter *MssqlAdapter) Write() error {
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
	stmts.` + name + `, err = db.Prepare("` + stmt.Contents + `")
	if err != nil {
		log.Print("Error in ` + name + ` statement.")
		log.Print("Bad Query: ","` + stmt.Contents + `")
		return err
	}
	`
		}
	}

	// TODO: Move these custom queries out of this file
	out := `// +build mssql

// This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time.
package main

import "log"
import "database/sql"
import "./common"

// nolint
type Stmts struct {
` + stmts + `
	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt
	todaysPostCount *sql.Stmt
	todaysTopicCount *sql.Stmt
	todaysReportCount *sql.Stmt
	todaysNewUserCount *sql.Stmt

	Mocks bool
}

// nolint
func _gen_mssql() (err error) {
	common.DebugLog("Building the generated statements")
` + body + `
	return nil
}
`
	return writeFile("./gen_mssql.go", out)
}

// Internal methods, not exposed in the interface
func (adapter *MssqlAdapter) pushStatement(name string, stype string, querystr string) {
	if name[0] == '_' {
		return
	}
	adapter.Buffer[name] = DBStmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *MssqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "char" || ctype == "varchar" || ctype == "datetime" || ctype == "text" || ctype == "nvarchar"
}

type SetPrimaryKeys interface {
	SetPrimaryKeys(keys map[string]string)
}

func (adapter *MssqlAdapter) SetPrimaryKeys(keys map[string]string) {
	adapter.keys = keys
}
