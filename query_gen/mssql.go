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
func (a *MssqlAdapter) GetName() string {
	return a.Name
}

func (a *MssqlAdapter) GetStmt(name string) DBStmt {
	return a.Buffer[name]
}

func (a *MssqlAdapter) GetStmts() map[string]DBStmt {
	return a.Buffer
}

// TODO: Implement this
func (a *MssqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
	return nil, nil
}

func (a *MssqlAdapter) DbVersion() string {
	return "SELECT CONCAT(SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition'))"
}

func (a *MssqlAdapter) DropTable(name, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "DROP TABLE IF EXISTS [" + table + "];"
	a.pushStatement(name, "drop-table", q)
	return q, nil
}

// TODO: Add support for foreign keys?
// TODO: Convert any remaining stringy types to nvarchar
// We may need to change the CreateTable API to better suit Mssql and the other database drivers which are coming up
func (a *MssqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	q := "CREATE TABLE [" + table + "] ("
	for _, column := range columns {
		column, size, end := a.parseColumn(column)
		q += "\n\t[" + column.Name + "] " + column.Type + size + end + ","
	}

	if len(keys) > 0 {
		for _, key := range keys {
			q += "\n\t" + key.Type
			if key.Type != "unique" {
				q += " key"
			}
			q += "("
			for _, column := range strings.Split(key.Columns, ",") {
				q += "[" + column + "],"
			}
			q = q[0:len(q)-1] + "),"
		}
	}

	q = q[0:len(q)-1] + "\n);"
	a.pushStatement(name, "create-table", q)
	return q, nil
}

func (a *MssqlAdapter) parseColumn(column DBTableColumn) (col DBTableColumn, size string, end string) {
	var max, createdAt bool
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

	if column.Size > 0 {
		size = " (" + strconv.Itoa(column.Size) + ")"
	}
	if max {
		size = " (MAX)"
	}

	if column.Default != "" {
		end = " DEFAULT "
		if createdAt {
			end += "GETUTCDATE()" // TODO: Use GETUTCDATE() in updates instead of the neutral format
		} else if a.stringyType(column.Type) && column.Default != "''" {
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
	return column, size, end
}

// TODO: Test this, not sure if some things work
// TODO: Add support for keys
func (a *MssqlAdapter) AddColumn(name string, table string, column DBTableColumn, key *DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	column, size, end := a.parseColumn(column)
	q := "ALTER TABLE [" + table + "] ADD [" + column.Name + "] " + column.Type + size + end + ";"
	a.pushStatement(name, "add-column", q)
	return q, nil
}

// TODO: Implement this
// TODO: Test to make sure everything works here
func (a *MssqlAdapter) AddIndex(name, table, iname, colname string) (string, error) {
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
func (a *MssqlAdapter) AddKey(name string, table string, column string, key DBTableKey) (string, error) {
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
func (a *MssqlAdapter) AddForeignKey(name string, table string, column string, ftable string, fcolumn string, cascade bool) (out string, e error) {
	c := func(str string, val bool) {
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

func (a *MssqlAdapter) SimpleInsert(name, table, cols, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	q := "INSERT INTO [" + table + "] ("
	if cols == "" {
		q += ") VALUES ()"
		a.pushStatement(name, "insert", q)
		return q, nil
	}

	// Escape the column names, just in case we've used a reserved keyword
	for _, col := range processColumns(cols) {
		if col.Type == TokenFunc {
			q += col.Left + ","
		} else {
			q += "[" + col.Left + "],"
		}
	}
	q = q[0 : len(q)-1]

	q += ") VALUES ("
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
		q += field.Name + ","
	}
	q = q[0:len(q)-1] + ")"

	a.pushStatement(name, "insert", q)
	return q, nil
}

// ! DEPRECATED
func (a *MssqlAdapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
	log.Print("In SimpleReplace")
	key, ok := a.keys[table]
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
	return a.SimpleUpsert(name, table, columns, fields, "key = "+keyValue)
}

func (a *MssqlAdapter) SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error) {
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
	q := "MERGE [" + table + "] WITH(HOLDLOCK) as t1 USING (VALUES("
	parsedFields := processFields(fields)
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
	q += fieldOutput + ")) AS updates ("

	// nolint The linter wants this to be less readable
	for fieldID, _ := range parsedFields {
		q += "f" + strconv.Itoa(fieldID) + ","
	}
	q = q[0:len(q)-1] + ") ON "

	//querystr += "t1.[" + key + "] = "
	// Add support for BETWEEN x.x
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case TokenSub:
				q += " ?"
			case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				q += " " + token.Contents
			case TokenColumn:
				q += " [" + token.Contents + "]"
			case TokenString:
				q += " '" + token.Contents + "'"
			default:
				panic("This token doesn't exist o_o")
			}
		}
	}

	matched := " WHEN MATCHED THEN UPDATE SET "
	notMatched := "WHEN NOT MATCHED THEN INSERT("
	var fieldList string

	// Escape the column names, just in case we've used a reserved keyword
	for columnID, col := range processColumns(columns) {
		fieldList += "f" + strconv.Itoa(columnID) + ","
		if col.Type == TokenFunc {
			matched += col.Left + " = f" + strconv.Itoa(columnID) + ","
			notMatched += col.Left + ","
		} else {
			matched += "[" + col.Left + "] = f" + strconv.Itoa(columnID) + ","
			notMatched += "[" + col.Left + "],"
		}
	}

	matched = matched[0 : len(matched)-1]
	notMatched = notMatched[0 : len(notMatched)-1]
	fieldList = fieldList[0 : len(fieldList)-1]

	notMatched += ") VALUES (" + fieldList + ");"
	q += matched + " " + notMatched

	// TODO: Run this on debug mode?
	if name[0] == '_' {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "upsert", q)
	return q, nil
}

func (a *MssqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	q := "UPDATE [" + up.table + "] SET "
	for _, item := range processSet(up.set) {
		q += "[" + item.Column + "]="
		for _, token := range item.Expr {
			switch token.Type {
			case TokenSub:
				q += " ?"
			case TokenFunc, TokenOp, TokenNumber, TokenOr:
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				q += " " + token.Contents
			case TokenColumn:
				q += " [" + token.Contents + "]"
			case TokenString:
				q += " '" + token.Contents + "'"
			default:
				panic("This token doesn't exist o_o")
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
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					q += " [" + token.Contents + "]"
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

func (a *MssqlAdapter) SimpleUpdateSelect(b *updatePrebuilder) (string, error) {
	return "", errors.New("not implemented")
}

func (a *MssqlAdapter) SimpleDelete(name string, table string, where string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	q := "DELETE FROM [" + table + "] WHERE"

	// Add support for BETWEEN x.x
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case TokenSub:
				q += " ?"
			case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
				// TODO: Split the function case off to speed things up
				if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
					token.Contents = "GETUTCDATE()"
				}
				q += " " + token.Contents
			case TokenColumn:
				q += " [" + token.Contents + "]"
			case TokenString:
				q += " '" + token.Contents + "'"
			default:
				panic("This token doesn't exist o_o")
			}
		}
		q += " AND"
	}

	q = strings.TrimSpace(q[0 : len(q)-4])
	a.pushStatement(name, "delete", q)
	return q, nil
}

func (a *MssqlAdapter) ComplexDelete(b *deletePrebuilder) (string, error) {
	return "", errors.New("not implemented")
}

// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (a *MssqlAdapter) Purge(name string, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "DELETE FROM [" + table + "]"
	a.pushStatement(name, "purge", q)
	return q, nil
}

func (a *MssqlAdapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
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
	subCount := 0
	q := ""

	// Escape the column names, just in case we've used a reserved keyword
	colslice := strings.Split(strings.TrimSpace(columns), ",")
	for _, column := range colslice {
		q += "[" + strings.TrimSpace(column) + "],"
	}
	q = q[0:len(q)-1] + " FROM [" + table + "]"

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenSub:
					subCount++
					q += " ?" + strconv.Itoa(subCount)
				case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					// MSSQL seems to convert the formats? so we'll compare it with a regular date. Do this with the other methods too?
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					q += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		q += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			// TODO: We might want to escape this column
			q += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	}

	if limit != "" {
		limiter := processLimit(limit)
		log.Printf("limiter: %+v\n", limiter)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				subCount++
				q += " OFFSET ?" + strconv.Itoa(subCount) + " ROWS"
			} else {
				q += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				subCount++
				limiter.MaxCount = "?" + strconv.Itoa(subCount)
			}
			q += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	q = strings.TrimSpace("SELECT " + q)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "select", q)
	return q, nil
}

// TODO: ComplexSelect
func (a *MssqlAdapter) ComplexSelect(preBuilder *selectPrebuilder) (string, error) {
	return "", nil
}

func (a *MssqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
	subCount := 0
	q := ""

	for _, col := range processColumns(columns) {
		var source, alias string
		// Escape the column names, just in case we've used a reserved keyword
		if col.Table != "" {
			source = "[" + col.Table + "].[" + col.Left + "]"
		} else if col.Type == TokenFunc {
			source = col.Left
		} else {
			source = "[" + col.Left + "]"
		}

		if col.Alias != "" {
			alias = " AS '" + col.Alias + "'"
		}
		q += source + alias + ","
	}
	// Remove the trailing comma
	q = q[0 : len(q)-1]

	q += " FROM [" + table1 + "] LEFT JOIN [" + table2 + "] ON "
	for _, j := range processJoiner(joiners) {
		q += "[" + j.LeftTable + "].[" + j.LeftColumn + "]" + j.Operator + "[" + j.RightTable + "].[" + j.RightColumn + "] AND "
	}
	// Remove the trailing AND
	q = q[0 : len(q)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenSub:
					subCount++
					q += " ?" + strconv.Itoa(subCount)
				case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						q += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						q += " [" + token.Contents + "]"
					}
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		q += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			q += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	} else if limit != "" {
		key, ok := a.keys[table1]
		if ok {
			q += " ORDER BY [" + table1 + "].[" + key + "]"
		}
	}

	if limit != "" {
		limiter := processLimit(limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				subCount++
				q += " OFFSET ?" + strconv.Itoa(subCount) + " ROWS"
			} else {
				q += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				subCount++
				limiter.MaxCount = "?" + strconv.Itoa(subCount)
			}
			q += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	q = strings.TrimSpace("SELECT " + q)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MssqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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
	subCount := 0
	q := ""

	for _, col := range processColumns(columns) {
		var source, alias string
		// Escape the column names, just in case we've used a reserved keyword
		if col.Table != "" {
			source = "[" + col.Table + "].[" + col.Left + "]"
		} else if col.Type == TokenFunc {
			source = col.Left
		} else {
			source = "[" + col.Left + "]"
		}

		if col.Alias != "" {
			alias = " AS '" + col.Alias + "'"
		}
		q += source + alias + ","
	}
	// Remove the trailing comma
	q = q[0 : len(q)-1]

	q += " FROM [" + table1 + "] INNER JOIN [" + table2 + "] ON "
	for _, j := range processJoiner(joiners) {
		q += "[" + j.LeftTable + "].[" + j.LeftColumn + "]" + j.Operator + "[" + j.RightTable + "].[" + j.RightColumn + "] AND "
	}
	// Remove the trailing AND
	q = q[0 : len(q)-4]

	// Add support for BETWEEN x.x
	if len(where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenSub:
					subCount++
					q += " ?" + strconv.Itoa(subCount)
				case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						q += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						q += " [" + token.Contents + "]"
					}
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(orderby) != 0 {
		q += " ORDER BY "
		for _, column := range processOrderby(orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			q += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	} else if limit != "" {
		key, ok := a.keys[table1]
		if ok {
			log.Print("key: ", key)
			q += " ORDER BY [" + table1 + "].[" + key + "]"
		}
	}

	if limit != "" {
		limiter := processLimit(limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				subCount++
				q += " OFFSET ?" + strconv.Itoa(subCount) + " ROWS"
			} else {
				q += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				subCount++
				limiter.MaxCount = "?" + strconv.Itoa(subCount)
			}
			q += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	q = strings.TrimSpace("SELECT " + q)
	// TODO: Run this on debug mode?
	if name[0] == '_' && limit == "" {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MssqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	// TODO: More errors.
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(sel.Orderby) == 0 && sel.Limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	/* Insert */
	q := "INSERT INTO [" + ins.Table + "] ("

	// Escape the column names, just in case we've used a reserved keyword
	for _, col := range processColumns(ins.Columns) {
		if col.Type == TokenFunc {
			q += col.Left + ","
		} else {
			q += "[" + col.Left + "],"
		}
	}
	q = q[0:len(q)-1] + ") SELECT "

	/* Select */
	subCount := 0

	for _, col := range processColumns(sel.Columns) {
		var source, alias string
		// Escape the column names, just in case we've used a reserved keyword
		if col.Type == TokenFunc || col.Type == TokenSub {
			source = col.Left
		} else {
			source = "[" + col.Left + "]"
		}
		if col.Alias != "" {
			alias = " AS [" + col.Alias + "]"
		}
		q += " " + source + alias + ","
	}
	q = q[0:len(q)-1] + " FROM [" + sel.Table + "] "

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenSub:
					subCount++
					q += " ?" + strconv.Itoa(subCount)
				case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					q += " [" + token.Contents + "]"
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(sel.Orderby) != 0 {
		q += " ORDER BY "
		for _, column := range processOrderby(sel.Orderby) {
			// TODO: We might want to escape this column
			q += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	} else if sel.Limit != "" {
		key, ok := a.keys[sel.Table]
		if ok {
			q += " ORDER BY [" + sel.Table + "].[" + key + "]"
		}
	}

	if sel.Limit != "" {
		limiter := processLimit(sel.Limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				subCount++
				q += " OFFSET ?" + strconv.Itoa(subCount) + " ROWS"
			} else {
				q += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				subCount++
				limiter.MaxCount = "?" + strconv.Itoa(subCount)
			}
			q += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	q = strings.TrimSpace(q)
	// TODO: Run this on debug mode?
	if name[0] == '_' && sel.Limit == "" {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MssqlAdapter) simpleJoin(name string, ins DBInsert, sel DBJoin, joinType string) (string, error) {
	// TODO: More errors.
	// TODO: Add this to the MySQL adapter in order to make this problem more discoverable?
	if len(sel.Orderby) == 0 && sel.Limit != "" {
		return "", errors.New("Orderby needs to be set to use limit on Mssql")
	}

	/* Insert */
	q := "INSERT INTO [" + ins.Table + "] ("

	// Escape the column names, just in case we've used a reserved keyword
	for _, col := range processColumns(ins.Columns) {
		if col.Type == TokenFunc {
			q += col.Left + ","
		} else {
			q += "[" + col.Left + "],"
		}
	}
	q = q[0:len(q)-1] + ") SELECT "

	/* Select */
	subCount := 0

	for _, col := range processColumns(sel.Columns) {
		var source, alias string
		// Escape the column names, just in case we've used a reserved keyword
		if col.Table != "" {
			source = "[" + col.Table + "].[" + col.Left + "]"
		} else if col.Type == TokenFunc {
			source = col.Left
		} else {
			source = "[" + col.Left + "]"
		}
		if col.Alias != "" {
			alias = " AS '" + col.Alias + "'"
		}
		q += source + alias + ","
	}
	// Remove the trailing comma
	q = q[0 : len(q)-1]

	q += " FROM [" + sel.Table1 + "] " + joinType + " JOIN [" + sel.Table2 + "] ON "
	for _, j := range processJoiner(sel.Joiners) {
		q += "[" + j.LeftTable + "].[" + j.LeftColumn + "] " + j.Operator + " [" + j.RightTable + "].[" + j.RightColumn + "] AND "
	}
	// Remove the trailing AND
	q = q[0 : len(q)-4]

	// Add support for BETWEEN x.x
	if len(sel.Where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(sel.Where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenSub:
					subCount++
					q += " ?" + strconv.Itoa(subCount)
				case TokenFunc, TokenOp, TokenNumber, TokenOr, TokenNot, TokenLike:
					// TODO: Split the function case off to speed things up
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						q += " [" + halves[0] + "].[" + halves[1] + "]"
					} else {
						q += " [" + token.Contents + "]"
					}
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

	// TODO: MSSQL requires ORDER BY for LIMIT
	if len(sel.Orderby) != 0 {
		q += " ORDER BY "
		for _, column := range processOrderby(sel.Orderby) {
			log.Print("column: ", column)
			// TODO: We might want to escape this column
			q += column.Column + " " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	} else if sel.Limit != "" {
		key, ok := a.keys[sel.Table1]
		if ok {
			q += " ORDER BY [" + sel.Table1 + "].[" + key + "]"
		}
	}

	if sel.Limit != "" {
		limiter := processLimit(sel.Limit)
		if limiter.Offset != "" {
			if limiter.Offset == "?" {
				subCount++
				q += " OFFSET ?" + strconv.Itoa(subCount) + " ROWS"
			} else {
				q += " OFFSET " + limiter.Offset + " ROWS"
			}
		}

		// ! Does this work without an offset?
		if limiter.MaxCount != "" {
			if limiter.MaxCount == "?" {
				subCount++
				limiter.MaxCount = "?" + strconv.Itoa(subCount)
			}
			q += " FETCH NEXT " + limiter.MaxCount + " ROWS ONLY "
		}
	}

	q = strings.TrimSpace(q)
	// TODO: Run this on debug mode?
	if name[0] == '_' && sel.Limit == "" {
		log.Print(name+" query: ", q)
	}
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MssqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return a.simpleJoin(name, ins, sel, "LEFT")
}

func (a *MssqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	return a.simpleJoin(name, ins, sel, "INNER")
}

func (a *MssqlAdapter) SimpleCount(name, table, where, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "SELECT COUNT(*) FROM [" + table + "]"

	// TODO: Add support for BETWEEN x.x
	if len(where) != 0 {
		q += " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					if strings.ToUpper(token.Contents) == "UTC_TIMESTAMP()" {
						token.Contents = "GETUTCDATE()"
					}
					q += " " + token.Contents
				case TokenColumn:
					q += " [" + token.Contents + "]"
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
	if limit != "" {
		q += " LIMIT " + limit
	}

	q = strings.TrimSpace(q)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MssqlAdapter) Builder() *prebuilder {
	return &prebuilder{a}
}

func (a *MssqlAdapter) Write() error {
	var stmts, body string
	for _, name := range a.BufferOrder {
		if name == "" {
			continue
		}
		stmt := a.Buffer[name]
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
import "github.com/Azareal/Gosora/common"

// nolint
type Stmts struct {
` + stmts + `
	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt

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
func (a *MssqlAdapter) pushStatement(name string, stype string, querystr string) {
	if name == "" {
		return
	}
	a.Buffer[name] = DBStmt{querystr, stype}
	a.BufferOrder = append(a.BufferOrder, name)
}

func (a *MssqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "char" || ctype == "varchar" || ctype == "datetime" || ctype == "text" || ctype == "nvarchar"
}

type SetPrimaryKeys interface {
	SetPrimaryKeys(keys map[string]string)
}

func (a *MssqlAdapter) SetPrimaryKeys(keys map[string]string) {
	a.keys = keys
}
