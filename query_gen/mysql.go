/* WIP Under Construction */
package qgen

import (
	"database/sql"
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var ErrNoCollation = errors.New("You didn't provide a collation")

func init() {
	Registry = append(Registry,
		&MysqlAdapter{Name: "mysql", Buffer: make(map[string]DBStmt)},
	)
}

type MysqlAdapter struct {
	Name        string // ? - Do we really need this? Can't we hard-code this?
	Buffer      map[string]DBStmt
	BufferOrder []string // Map iteration order is random, so we need this to track the order, so we don't get huge diffs every commit
}

// GetName gives you the name of the database adapter. In this case, it's mysql
func (adapter *MysqlAdapter) GetName() string {
	return adapter.Name
}

func (adapter *MysqlAdapter) GetStmt(name string) DBStmt {
	return adapter.Buffer[name]
}

func (adapter *MysqlAdapter) GetStmts() map[string]DBStmt {
	return adapter.Buffer
}

// TODO: Add an option to disable unix pipes
func (adapter *MysqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
	dbCollation, ok := config["collation"]
	if !ok {
		return nil, ErrNoCollation
	}
	var dbpassword string
	if config["password"] != "" {
		dbpassword = ":" + config["password"]
	}

	// First try opening a pipe as those are faster
	if runtime.GOOS == "linux" {
		var dbsocket = "/tmp/mysql.sock"
		if config["socket"] != "" {
			dbsocket = config["socket"]
		}

		// The MySQL adapter refuses to open any other connections, if the unix socket doesn't exist, so check for it first
		_, err := os.Stat(dbsocket)
		if err == nil {
			db, err := sql.Open("mysql", config["username"]+dbpassword+"@unix("+dbsocket+")/"+config["name"]+"?collation="+dbCollation+"&parseTime=true")
			if err == nil {
				// Make sure that the connection is alive
				return db, db.Ping()
			}
		}
	}

	// Open the database connection
	db, err := sql.Open("mysql", config["username"]+dbpassword+"@tcp("+config["host"]+":"+config["port"]+")/"+config["name"]+"?collation="+dbCollation+"&parseTime=true")
	if err != nil {
		return db, err
	}

	// Make sure that the connection is alive
	return db, db.Ping()
}

func (adapter *MysqlAdapter) DbVersion() string {
	return "SELECT VERSION()"
}

func (adapter *MysqlAdapter) DropTable(name string, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	querystr := "DROP TABLE IF EXISTS `" + table + "`;"
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "drop-table", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	var querystr = "CREATE TABLE `" + table + "` ("
	for _, column := range columns {
		column, size, end := adapter.parseColumn(column)
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

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "create-table", querystr+";")
	return querystr + ";", nil
}

func (adapter *MysqlAdapter) parseColumn(column DBTableColumn) (col DBTableColumn, size string, end string) {
	// Make it easier to support Cassandra in the future
	if column.Type == "createdAt" {
		column.Type = "datetime"
		// MySQL doesn't support this x.x
		/*if column.Default == "" {
			column.Default = "UTC_TIMESTAMP()"
		}*/
	} else if column.Type == "json" {
		column.Type = "text"
	}
	if column.Size > 0 {
		size = "(" + strconv.Itoa(column.Size) + ")"
	}

	// TODO: Exclude the other variants of text like mediumtext and longtext too
	if column.Default != "" && column.Type != "text" {
		end = " DEFAULT "
		/*if column.Type == "datetime" && column.Default[len(column.Default)-1] == ')' {
			end += column.Default
		} else */if adapter.stringyType(column.Type) && column.Default != "''" {
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
	if column.AutoIncrement {
		end += " AUTO_INCREMENT"
	}
	return column, size, end
}

// TODO: Support AFTER column
// TODO: Test to make sure everything works here
func (adapter *MysqlAdapter) AddColumn(name string, table string, column DBTableColumn, key *DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	column, size, end := adapter.parseColumn(column)
	querystr := "ALTER TABLE `" + table + "` ADD COLUMN " + "`" + column.Name + "` " + column.Type + size + end

	if key != nil {
		querystr += " " + key.Type
		if key.Type != "unique" {
			querystr += " key"
		} else if key.Type == "primary" {
			querystr += " first"
		}
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "add-column", querystr)
	return querystr, nil
}

// TODO: Test to make sure everything works here
func (adapter *MysqlAdapter) AddIndex(name string, table string, iname string, colname string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if iname == "" {
		return "", errors.New("You need a name for the index")
	}
	if colname == "" {
		return "", errors.New("You need a name for the column")
	}

	querystr := "ALTER TABLE `" + table + "` ADD INDEX " + "`" + iname + "` (`" + colname + "`);"
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "add-index", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleInsert(name string, table string, columns string, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var querystr = "INSERT INTO `" + table + "`("
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

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) buildColumns(columns string) (querystr string) {
	if columns == "" {
		return ""
	}
	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range processColumns(columns) {
		if column.Type == "function" {
			querystr += column.Left + ","
		} else {
			querystr += "`" + column.Left + "`,"
		}
	}
	return querystr[0 : len(querystr)-1]
}

// ! DEPRECATED
func (adapter *MysqlAdapter) SimpleReplace(name string, table string, columns string, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleInsert")
	}
	if len(fields) == 0 {
		return "", errors.New("No input data found for SimpleInsert")
	}

	var querystr = "REPLACE INTO `" + table + "`(" + adapter.buildColumns(columns) + ") VALUES ("
	for _, field := range processFields(fields) {
		querystr += field.Name + ","
	}
	querystr = querystr[0 : len(querystr)-1]

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "replace", querystr+")")
	return querystr + ")", nil
}

func (adapter *MysqlAdapter) SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error) {
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

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "upsert", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	var querystr = "UPDATE `" + up.table + "` SET "
	for _, item := range processSet(up.set) {
		querystr += "`" + item.Column + "` ="
		for _, token := range item.Expr {
			switch token.Type {
			case "function", "operator", "number", "substitute", "or":
				querystr += " " + token.Contents
			case "column":
				querystr += " `" + token.Contents + "`"
			case "string":
				querystr += " '" + token.Contents + "'"
			}
		}
		querystr += ","
	}
	querystr = querystr[0 : len(querystr)-1]

	whereStr, err := adapter.buildWhere(up.where)
	if err != nil {
		return querystr, err
	}
	querystr += whereStr

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(up.name, "update", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleDelete(name string, table string, where string) (string, error) {
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
			case "function", "operator", "number", "substitute", "or":
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
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	adapter.pushStatement(name, "delete", querystr)
	return querystr, nil
}

// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (adapter *MysqlAdapter) Purge(name string, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	adapter.pushStatement(name, "purge", "DELETE FROM `"+table+"`")
	return "DELETE FROM `" + table + "`", nil
}

func (adapter *MysqlAdapter) buildWhere(where string) (querystr string, err error) {
	if len(where) == 0 {
		return "", nil
	}
	querystr = " WHERE"
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case "function", "operator", "number", "substitute", "or":
				querystr += " " + token.Contents
			case "column":
				querystr += " `" + token.Contents + "`"
			case "string":
				querystr += " '" + token.Contents + "'"
			default:
				return querystr, errors.New("This token doesn't exist o_o")
			}
		}
		querystr += " AND"
	}
	return querystr[0 : len(querystr)-4], nil
}

// The new version of buildWhere() currently only used in ComplexSelect for complex OO builder queries
func (adapter *MysqlAdapter) buildFlexiWhere(where string, dateCutoff *dateCutoff) (querystr string, err error) {
	if len(where) == 0 && dateCutoff == nil {
		return "", nil
	}
	querystr = " WHERE"
	if dateCutoff != nil {
		querystr += " " + dateCutoff.Column + " BETWEEN (UTC_TIMESTAMP() - interval " + strconv.Itoa(dateCutoff.Quantity) + " " + dateCutoff.Unit + ") AND UTC_TIMESTAMP() AND"
	}
	if len(where) != 0 {
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute", "or":
					querystr += " " + token.Contents
				case "column":
					querystr += " `" + token.Contents + "`"
				case "string":
					querystr += " '" + token.Contents + "'"
				default:
					return querystr, errors.New("This token doesn't exist o_o")
				}
			}
			querystr += " AND"
		}
	}
	return querystr[0 : len(querystr)-4], nil
}

func (adapter *MysqlAdapter) buildOrderby(orderby string) (querystr string) {
	if len(orderby) != 0 {
		querystr = " ORDER BY "
		for _, column := range processOrderby(orderby) {
			// TODO: We might want to escape this column
			querystr += "`" + strings.Replace(column.Column, ".", "`.`", -1) + "` " + strings.ToUpper(column.Order) + ","
		}
		querystr = querystr[0 : len(querystr)-1]
	}
	return querystr
}

func (adapter *MysqlAdapter) SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}

	var querystr = "SELECT "

	// Slice up the user friendly strings into something easier to process
	for _, column := range strings.Split(strings.TrimSpace(columns), ",") {
		querystr += "`" + strings.TrimSpace(column) + "`,"
	}
	querystr = querystr[0 : len(querystr)-1]

	whereStr, err := adapter.buildWhere(where)
	if err != nil {
		return querystr, err
	}

	querystr += " FROM `" + table + "`" + whereStr + adapter.buildOrderby(orderby) + adapter.buildLimit(limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) ComplexSelect(preBuilder *selectPrebuilder) (out string, err error) {
	if preBuilder.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(preBuilder.columns) == 0 {
		return "", errors.New("No columns found for ComplexSelect")
	}

	var querystr = "SELECT "

	// Slice up the user friendly strings into something easier to process
	for _, column := range strings.Split(strings.TrimSpace(preBuilder.columns), ",") {
		querystr += "`" + strings.TrimSpace(column) + "`,"
	}
	querystr = querystr[0 : len(querystr)-1]

	var whereStr string
	// TODO: Let callers have a Where() and a InQ()
	if preBuilder.inChain != nil {
		whereStr, err = adapter.ComplexSelect(preBuilder.inChain)
		if err != nil {
			return querystr, err
		}
		whereStr = " WHERE `" + preBuilder.inColumn + "` IN(" + whereStr + ")"
	} else {
		whereStr, err = adapter.buildFlexiWhere(preBuilder.where, preBuilder.dateCutoff)
		if err != nil {
			return querystr, err
		}
	}

	querystr += " FROM `" + preBuilder.table + "`" + whereStr + adapter.buildOrderby(preBuilder.orderby) + adapter.buildLimit(preBuilder.limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(preBuilder.name, "select", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	whereStr, err := adapter.buildJoinWhere(where)
	if err != nil {
		return "", err
	}

	var querystr = "SELECT" + adapter.buildJoinColumns(columns) + " FROM `" + table1 + "` LEFT JOIN `" + table2 + "` ON " + adapter.buildJoiners(joiners) + whereStr + adapter.buildOrderby(orderby) + adapter.buildLimit(limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	whereStr, err := adapter.buildJoinWhere(where)
	if err != nil {
		return "", err
	}

	var querystr = "SELECT " + adapter.buildJoinColumns(columns) + " FROM `" + table1 + "` INNER JOIN `" + table2 + "` ON " + adapter.buildJoiners(joiners) + whereStr + adapter.buildOrderby(orderby) + adapter.buildLimit(limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleUpdateSelect(up *updatePrebuilder) (string, error) {
	sel := up.whereSubQuery
	whereStr, err := adapter.buildWhere(sel.where)
	if err != nil {
		return "", err
	}

	var setter string
	for _, item := range processSet(up.set) {
		setter += "`" + item.Column + "` ="
		for _, token := range item.Expr {
			switch token.Type {
			case "function", "operator", "number", "substitute", "or":
				setter += " " + token.Contents
			case "column":
				setter += " `" + token.Contents + "`"
			case "string":
				setter += " '" + token.Contents + "'"
			}
		}
		setter += ","
	}
	setter = setter[0 : len(setter)-1]

	var querystr = "UPDATE `" + up.table + "` SET " + setter + " WHERE (SELECT" + adapter.buildJoinColumns(sel.columns) + " FROM `" + sel.table + "`" + whereStr + adapter.buildOrderby(sel.orderby) + adapter.buildLimit(sel.limit) + ")"

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(up.name, "update", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	whereStr, err := adapter.buildWhere(sel.Where)
	if err != nil {
		return "", err
	}

	var querystr = "INSERT INTO `" + ins.Table + "`(" + adapter.buildColumns(ins.Columns) + ") SELECT" + adapter.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table + "`" + whereStr + adapter.buildOrderby(sel.Orderby) + adapter.buildLimit(sel.Limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	whereStr, err := adapter.buildJoinWhere(sel.Where)
	if err != nil {
		return "", err
	}

	var querystr = "INSERT INTO `" + ins.Table + "`(" + adapter.buildColumns(ins.Columns) + ") SELECT" + adapter.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table1 + "` LEFT JOIN `" + sel.Table2 + "` ON " + adapter.buildJoiners(sel.Joiners) + whereStr + adapter.buildOrderby(sel.Orderby) + adapter.buildLimit(sel.Limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

// TODO: Make this more consistent with the other build* methods?
func (adapter *MysqlAdapter) buildJoiners(joiners string) (querystr string) {
	for _, joiner := range processJoiner(joiners) {
		querystr += "`" + joiner.LeftTable + "`.`" + joiner.LeftColumn + "` " + joiner.Operator + " `" + joiner.RightTable + "`.`" + joiner.RightColumn + "` AND "
	}
	// Remove the trailing AND
	return querystr[0 : len(querystr)-4]
}

// Add support for BETWEEN x.x
func (adapter *MysqlAdapter) buildJoinWhere(where string) (querystr string, err error) {
	if len(where) != 0 {
		querystr = " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case "function", "operator", "number", "substitute", "or":
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
					return querystr, errors.New("This token doesn't exist o_o")
				}
			}
			querystr += " AND"
		}
		querystr = querystr[0 : len(querystr)-4]
	}
	return querystr, nil
}

func (adapter *MysqlAdapter) buildLimit(limit string) (querystr string) {
	if limit != "" {
		querystr = " LIMIT " + limit
	}
	return querystr
}

func (adapter *MysqlAdapter) buildJoinColumns(columns string) (querystr string) {
	for _, column := range processColumns(columns) {
		// Escape the column names, just in case we've used a reserved keyword
		var source = column.Left
		if column.Table != "" {
			source = "`" + column.Table + "`.`" + source + "`"
		} else if column.Type != "function" {
			source = "`" + source + "`"
		}

		var alias string
		if column.Alias != "" {
			alias = " AS `" + column.Alias + "`"
		}
		querystr += " " + source + alias + ","
	}
	return querystr[0 : len(querystr)-1]
}

func (adapter *MysqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	whereStr, err := adapter.buildJoinWhere(sel.Where)
	if err != nil {
		return "", err
	}

	var querystr = "INSERT INTO `" + ins.Table + "`(" + adapter.buildColumns(ins.Columns) + ") SELECT" + adapter.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table1 + "` INNER JOIN `" + sel.Table2 + "` ON " + adapter.buildJoiners(sel.Joiners) + whereStr + adapter.buildOrderby(sel.Orderby) + adapter.buildLimit(sel.Limit)

	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "insert", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) SimpleCount(name string, table string, where string, limit string) (querystr string, err error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	whereStr, err := adapter.buildWhere(where)
	if err != nil {
		return "", err
	}

	querystr = "SELECT COUNT(*) AS `count` FROM `" + table + "`" + whereStr + adapter.buildLimit(limit)
	querystr = strings.TrimSpace(querystr)
	adapter.pushStatement(name, "select", querystr)
	return querystr, nil
}

func (adapter *MysqlAdapter) Builder() *prebuilder {
	return &prebuilder{adapter}
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
			stmts += "\t" + name + " *qgen.MySQLUpsertCallback\n"
			body += `	
	common.DebugLog("Preparing ` + name + ` statement.")
	stmts.` + name + `, err = qgen.PrepareMySQLUpsertCallback(db, "` + stmt.Contents + `")
	if err != nil {
		log.Print("Error in ` + name + ` statement.")
		return err
	}
	`
		} else if stmt.Type != "create-table" {
			stmts += "\t" + name + " *sql.Stmt\n"
			body += `	
	common.DebugLog("Preparing ` + name + ` statement.")
	stmts.` + name + `, err = db.Prepare("` + stmt.Contents + `")
	if err != nil {
		log.Print("Error in ` + name + ` statement.")
		return err
	}
	`
		}
	}

	// TODO: Move these custom queries out of this file
	out := `// +build !pgsql,!mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import "log"
import "database/sql"
import "github.com/Azareal/Gosora/common"
//import "github.com/Azareal/Gosora/query_gen"

// nolint
type Stmts struct {
` + stmts + `
	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt

	Mocks bool
}

// nolint
func _gen_mysql() (err error) {
	common.DebugLog("Building the generated statements")
` + body + `
	return nil
}
`
	return writeFile("./gen_mysql.go", out)
}

// Internal methods, not exposed in the interface
func (adapter *MysqlAdapter) pushStatement(name string, stype string, querystr string) {
	if name == "" {
		return
	}
	adapter.Buffer[name] = DBStmt{querystr, stype}
	adapter.BufferOrder = append(adapter.BufferOrder, name)
}

func (adapter *MysqlAdapter) stringyType(ctype string) bool {
	ctype = strings.ToLower(ctype)
	return ctype == "varchar" || ctype == "tinytext" || ctype == "text" || ctype == "mediumtext" || ctype == "longtext" || ctype == "char" || ctype == "datetime" || ctype == "timestamp" || ctype == "time" || ctype == "date"
}
