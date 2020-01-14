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
func (a *MysqlAdapter) GetName() string {
	return a.Name
}

func (a *MysqlAdapter) GetStmt(name string) DBStmt {
	return a.Buffer[name]
}

func (a *MysqlAdapter) GetStmts() map[string]DBStmt {
	return a.Buffer
}

// TODO: Add an option to disable unix pipes
func (a *MysqlAdapter) BuildConn(config map[string]string) (*sql.DB, error) {
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
		dbsocket := "/tmp/mysql.sock"
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

func (a *MysqlAdapter) DbVersion() string {
	return "SELECT VERSION()"
}

func (a *MysqlAdapter) DropTable(name, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "DROP TABLE IF EXISTS `" + table + "`;"
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "drop-table", q)
	return q, nil
}

func (a *MysqlAdapter) CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	q := "CREATE TABLE `" + table + "` ("
	for _, column := range columns {
		column, size, end := a.parseColumn(column)
		q += "\n\t`" + column.Name + "` " + column.Type + size + end + ","
	}

	if len(keys) > 0 {
		for _, key := range keys {
			q += "\n\t" + key.Type
			if key.Type != "unique" {
				q += " key"
			}
			if key.Type == "foreign" {
				cols := strings.Split(key.Columns, ",")
				q += "(`" + cols[0] + "`) REFERENCES `" + key.FTable + "`(`" + cols[1] + "`)"
				if key.Cascade {
					q += " ON DELETE CASCADE"
				}
				q += ","
			} else {
				q += "("
				for _, column := range strings.Split(key.Columns, ",") {
					q += "`" + column + "`,"
				}
				q = q[0:len(q)-1] + "),"
			}
		}
	}

	q = q[0:len(q)-1] + "\n)"
	if charset != "" {
		q += " CHARSET=" + charset
	}
	if collation != "" {
		q += " COLLATE " + collation
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q += ";"
	a.pushStatement(name, "create-table", q)
	return q, nil
}

func (a *MysqlAdapter) parseColumn(column DBTableColumn) (col DBTableColumn, size string, end string) {
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
		} else */if a.stringyType(column.Type) && column.Default != "''" {
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
func (a *MysqlAdapter) AddColumn(name string, table string, column DBTableColumn, key *DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	column, size, end := a.parseColumn(column)
	q := "ALTER TABLE `" + table + "` ADD COLUMN " + "`" + column.Name + "` " + column.Type + size + end

	if key != nil {
		q += " " + key.Type
		if key.Type != "unique" {
			q += " key"
		} else if key.Type == "primary" {
			q += " first"
		}
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-column", q)
	return q, nil
}

// TODO: Test to make sure everything works here
func (a *MysqlAdapter) AddIndex(name, table, iname, colname string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if iname == "" {
		return "", errors.New("You need a name for the index")
	}
	if colname == "" {
		return "", errors.New("You need a name for the column")
	}

	q := "ALTER TABLE `" + table + "` ADD INDEX " + "`i_" + iname + "` (`" + colname + "`);"
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-index", q)
	return q, nil
}

// TODO: Test to make sure everything works here
// Only supports FULLTEXT right now
func (a *MysqlAdapter) AddKey(name string, table string, column string, key DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	var q string
	if key.Type == "fulltext" {
		q = "ALTER TABLE `" + table + "` ADD FULLTEXT(`" + column + "`)"
	} else {
		return "", errors.New("Only fulltext is supported by AddKey right now")
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-key", q)
	return q, nil
}

func (a *MysqlAdapter) AddForeignKey(name string, table string, column string, ftable string, fcolumn string, cascade bool) (out string, e error) {
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

	q := "ALTER TABLE `" + table + "` ADD CONSTRAINT `fk_" + column + "` FOREIGN KEY(`" + column + "`) REFERENCES `" + ftable + "`(`" + fcolumn + "`)"
	if cascade {
		q += " ON DELETE CASCADE"
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-foreign-key", q)
	return q, nil
}

var silen1 = len("INSERT INTO ``() VALUES () ")

func (a *MysqlAdapter) SimpleInsert(name, table, columns, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var sb strings.Builder
	sb.Grow(silen1 + len(table))
	sb.WriteString("INSERT INTO `")
	sb.WriteString(table)
	sb.WriteString("`(")
	if columns != "" {
		sb.WriteString(a.buildColumns(columns))
		sb.WriteString(") VALUES (")
		fs := processFields(fields)
		sb.Grow(len(fs) * 3)
		for i, field := range fs {
			if i != 0 {
				sb.WriteString(",")
			}
			nameLen := len(field.Name)
			if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
				field.Name = "'" + field.Name[1:nameLen-1] + "'"
			}
			if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
				field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
			}
			sb.WriteString(field.Name)
		}
		sb.WriteString(")")
	} else {
		sb.WriteString(") VALUES ()")
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q := sb.String()
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MysqlAdapter) buildColumns(columns string) (q string) {
	if columns == "" {
		return ""
	}
	// Escape the column names, just in case we've used a reserved keyword
	for _, col := range processColumns(columns) {
		if col.Type == TokenFunc {
			q += col.Left + ","
		} else {
			q += "`" + col.Left + "`,"
		}
	}
	return q[0 : len(q)-1]
}

// ! DEPRECATED
func (a *MysqlAdapter) SimpleReplace(name, table, columns, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleInsert")
	}
	if len(fields) == 0 {
		return "", errors.New("No input data found for SimpleInsert")
	}

	q := "REPLACE INTO `" + table + "`(" + a.buildColumns(columns) + ") VALUES ("
	for _, field := range processFields(fields) {
		q += field.Name + ","
	}
	q = q[0 : len(q)-1] + ")"

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "replace", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleUpsert(name, table, columns, fields, where string) (string, error) {
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

	q := "INSERT INTO `" + table + "`("
	parsedFields := processFields(fields)

	var insertColumns string
	var insertValues string
	setBit := ") ON DUPLICATE KEY UPDATE "

	for columnID, col := range processColumns(columns) {
		field := parsedFields[columnID]
		if col.Type == TokenFunc {
			insertColumns += col.Left + ","
			insertValues += field.Name + ","
			setBit += col.Left + " = " + field.Name + " AND "
		} else {
			insertColumns += "`" + col.Left + "`,"
			insertValues += field.Name + ","
			setBit += "`" + col.Left + "` = " + field.Name + " AND "
		}
	}
	insertColumns = insertColumns[0 : len(insertColumns)-1]
	insertValues = insertValues[0 : len(insertValues)-1]
	insertColumns += ") VALUES (" + insertValues
	setBit = setBit[0 : len(setBit)-5]

	q += insertColumns + setBit

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "upsert", q)
	return q, nil
}

var sulen1 = len("UPDATE `` SET ")

func (a *MysqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}
	var sb strings.Builder
	sb.Grow(sulen1 + len(up.table))
	sb.WriteString("UPDATE `")
	sb.WriteString(up.table)
	sb.WriteString("` SET ")

	set := processSet(up.set)
	sb.Grow(len(set) * 6)
	for i, item := range set {
		if i != 0 {
			sb.WriteString(",`")
		} else {
			sb.WriteString("`")
		}
		sb.WriteString(item.Column)
		sb.WriteString("`=")
		for _, token := range item.Expr {
			switch token.Type {
			case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr:
				sb.WriteString(" ")
				sb.WriteString(token.Contents)
			case TokenColumn:
				sb.WriteString(" `")
				sb.WriteString(token.Contents)
				sb.WriteString("`")
			case TokenString:
				sb.WriteString(" '")
				sb.WriteString(token.Contents)
				sb.WriteString("'")
			}
		}
	}

	whereStr, err := a.buildFlexiWhere(up.where, up.dateCutoff)
	sb.WriteString(whereStr)
	if err != nil {
		return sb.String(), err
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q := sb.String()
	a.pushStatement(up.name, "update", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleDelete(name, table, where string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	q := "DELETE FROM `" + table + "` WHERE"

	// Add support for BETWEEN x.x
	for _, loc := range processWhere(where) {
		for _, token := range loc.Expr {
			switch token.Type {
			case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot:
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

	q = strings.TrimSpace(q[0 : len(q)-4])
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "delete", q)
	return q, nil
}

func (a *MysqlAdapter) ComplexDelete(b *deletePrebuilder) (string, error) {
	if b.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if b.where == "" && b.dateCutoff == nil {
		return "", errors.New("You need to specify what data you want to delete")
	}
	q := "DELETE FROM `" + b.table + "`"

	whereStr, err := a.buildFlexiWhere(b.where, b.dateCutoff)
	if err != nil {
		return q, err
	}
	q += whereStr

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(b.name, "delete", q)
	return q, nil
}

// We don't want to accidentally wipe tables, so we'll have a separate method for purging tables instead
func (a *MysqlAdapter) Purge(name, table string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	q := "DELETE FROM `" + table + "`"
	a.pushStatement(name, "purge", q)
	return q, nil
}

func (a *MysqlAdapter) buildWhere(where string, sb *strings.Builder) error {
	if len(where) == 0 {
		return nil
	}
	spl := processWhere(where)
	sb.Grow(len(spl) * 8)
	for i, loc := range spl {
		if i != 0 {
			sb.WriteString(" AND ")
		} else {
			sb.WriteString(" WHERE ")
		}
		for _, token := range loc.Expr {
			switch token.Type {
			case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
				sb.WriteString(token.Contents)
				sb.WriteRune(' ')
			case TokenColumn:
				sb.WriteRune('`')
				sb.WriteString(token.Contents)
				sb.WriteRune('`')
			case TokenString:
				sb.WriteRune('\'')
				sb.WriteString(token.Contents)
				sb.WriteRune('\'')
			default:
				return errors.New("This token doesn't exist o_o")
			}
		}
	}
	return nil
}

// The new version of buildWhere() currently only used in ComplexSelect for complex OO builder queries
func (a *MysqlAdapter) buildFlexiWhere(where string, dateCutoff *dateCutoff) (q string, err error) {
	if len(where) == 0 && dateCutoff == nil {
		return "", nil
	}
	q = " WHERE"
	if dateCutoff != nil {
		if dateCutoff.Type == 0 {
			q += " " + dateCutoff.Column + " BETWEEN (UTC_TIMESTAMP() - interval " + strconv.Itoa(dateCutoff.Quantity) + " " + dateCutoff.Unit + ") AND UTC_TIMESTAMP() AND"
		} else {
			q += " " + dateCutoff.Column + " < UTC_TIMESTAMP() - interval " + strconv.Itoa(dateCutoff.Quantity) + " " + dateCutoff.Unit + " AND"
		}
	}

	if len(where) != 0 {
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					q += " " + token.Contents
				case TokenColumn:
					q += " `" + token.Contents + "`"
				case TokenString:
					q += " '" + token.Contents + "'"
				default:
					return q, errors.New("This token doesn't exist o_o")
				}
			}
			q += " AND"
		}
	}

	return q[0 : len(q)-4], nil
}

func (a *MysqlAdapter) buildOrderby(orderby string) (q string) {
	if len(orderby) != 0 {
		q = " ORDER BY "
		for _, column := range processOrderby(orderby) {
			// TODO: We might want to escape this column
			q += "`" + strings.Replace(column.Column, ".", "`.`", -1) + "` " + strings.ToUpper(column.Order) + ","
		}
		q = q[0 : len(q)-1]
	}
	return q
}

func (a *MysqlAdapter) SimpleSelect(name, table, columns, where, orderby, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	q := "SELECT "

	// Slice up the user friendly strings into something easier to process
	for _, column := range strings.Split(strings.TrimSpace(columns), ",") {
		q += "`" + strings.TrimSpace(column) + "`,"
	}
	q = q[0 : len(q)-1]

	sb := &strings.Builder{}
	err := a.buildWhere(where, sb)
	if err != nil {
		return "", err
	}
	q += " FROM `" + table + "`" + sb.String() + a.buildOrderby(orderby) + a.buildLimit(limit)

	q = strings.TrimSpace(q)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MysqlAdapter) ComplexSelect(preBuilder *selectPrebuilder) (out string, err error) {
	sb := &strings.Builder{}
	err = a.complexSelect(preBuilder, sb)
	out = sb.String()
	a.pushStatement(preBuilder.name, "select", out)
	return out, err
}

var cslen1 = len("SELECT  FROM ``")
var cslen2 = len(" WHERE `` IN(")

func (a *MysqlAdapter) complexSelect(preBuilder *selectPrebuilder, sb *strings.Builder) error {
	if preBuilder.table == "" {
		return errors.New("You need a name for this table")
	}
	if len(preBuilder.columns) == 0 {
		return errors.New("No columns found for ComplexSelect")
	}

	cols := a.buildJoinColumns(preBuilder.columns)
	sb.Grow(cslen1 + len(cols) + len(preBuilder.table))
	sb.WriteString("SELECT ")
	sb.WriteString(cols)
	sb.WriteString(" FROM `")
	sb.WriteString(preBuilder.table)
	sb.WriteRune('`')

	// TODO: Let callers have a Where() and a InQ()
	if preBuilder.inChain != nil {
		sb.Grow(cslen2 + len(preBuilder.inColumn))
		sb.WriteString(" WHERE `")
		sb.WriteString(preBuilder.inColumn)
		sb.WriteString("` IN(")
		err := a.complexSelect(preBuilder.inChain, sb)
		if err != nil {
			return err
		}
		sb.WriteRune(')')
	} else {
		whereStr, err := a.buildFlexiWhere(preBuilder.where, preBuilder.dateCutoff)
		if err != nil {
			return err
		}
		sb.WriteString(whereStr)
	}

	orderby := a.buildOrderby(preBuilder.orderby)
	limit := a.buildLimit(preBuilder.limit)
	sb.Grow(len(orderby) + len(limit))
	sb.WriteString(orderby)
	sb.WriteString(limit)
	return nil
}

func (a *MysqlAdapter) SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	whereStr, err := a.buildJoinWhere(where)
	if err != nil {
		return "", err
	}

	thalf1 := strings.Split(strings.Replace(table1, " as ", " AS ", -1), " AS ")
	var as1 string
	if len(thalf1) == 2 {
		as1 = " AS `" + thalf1[1] + "`"
	}
	thalf2 := strings.Split(strings.Replace(table2, " as ", " AS ", -1), " AS ")
	var as2 string
	if len(thalf2) == 2 {
		as2 = " AS `" + thalf2[1] + "`"
	}

	q := "SELECT" + a.buildJoinColumns(columns) + " FROM `" + thalf1[0] + "`" + as1 + " LEFT JOIN `" + thalf2[0] + "`" + as2 + " ON " + a.buildJoiners(joiners) + whereStr + a.buildOrderby(orderby) + a.buildLimit(limit)

	q = strings.TrimSpace(q)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleInnerJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error) {
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

	whereStr, err := a.buildJoinWhere(where)
	if err != nil {
		return "", err
	}

	thalf1 := strings.Split(strings.Replace(table1, " as ", " AS ", -1), " AS ")
	var as1 string
	if len(thalf1) == 2 {
		as1 = " AS `" + thalf1[1] + "`"
	}
	thalf2 := strings.Split(strings.Replace(table2, " as ", " AS ", -1), " AS ")
	var as2 string
	if len(thalf2) == 2 {
		as2 = " AS `" + thalf2[1] + "`"
	}

	q := "SELECT " + a.buildJoinColumns(columns) + " FROM `" + thalf1[0] + "`" + as1 + " INNER JOIN `" + thalf2[0] + "`" + as2 + " ON " + a.buildJoiners(joiners) + whereStr + a.buildOrderby(orderby) + a.buildLimit(limit)

	q = strings.TrimSpace(q)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleUpdateSelect(up *updatePrebuilder) (string, error) {
	sel := up.whereSubQuery
	sb := &strings.Builder{}
	err := a.buildWhere(sel.where, sb)
	if err != nil {
		return "", err
	}

	var setter string
	for _, item := range processSet(up.set) {
		setter += "`" + item.Column + "`="
		for _, token := range item.Expr {
			switch token.Type {
			case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr:
				setter += token.Contents
			case TokenColumn:
				setter += "`" + token.Contents + "`"
			case TokenString:
				setter += "'" + token.Contents + "'"
			}
		}
		setter += ","
	}
	setter = setter[0 : len(setter)-1]

	q := "UPDATE `" + up.table + "` SET " + setter + " WHERE (SELECT" + a.buildJoinColumns(sel.columns) + " FROM `" + sel.table + "`" + sb.String() + a.buildOrderby(sel.orderby) + a.buildLimit(sel.limit) + ")"
	q = strings.TrimSpace(q)
	a.pushStatement(up.name, "update", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleInsertSelect(name string, ins DBInsert, sel DBSelect) (string, error) {
	sb := &strings.Builder{}
	err := a.buildWhere(sel.Where, sb)
	if err != nil {
		return "", err
	}

	q := "INSERT INTO `" + ins.Table + "`(" + a.buildColumns(ins.Columns) + ") SELECT" + a.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table + "`" + sb.String() + a.buildOrderby(sel.Orderby) + a.buildLimit(sel.Limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	whereStr, err := a.buildJoinWhere(sel.Where)
	if err != nil {
		return "", err
	}

	q := "INSERT INTO `" + ins.Table + "`(" + a.buildColumns(ins.Columns) + ") SELECT" + a.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table1 + "` LEFT JOIN `" + sel.Table2 + "` ON " + a.buildJoiners(sel.Joiners) + whereStr + a.buildOrderby(sel.Orderby) + a.buildLimit(sel.Limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "insert", q)
	return q, nil
}

// TODO: Make this more consistent with the other build* methods?
func (a *MysqlAdapter) buildJoiners(joiners string) (q string) {
	for _, j := range processJoiner(joiners) {
		q += "`" + j.LeftTable + "`.`" + j.LeftColumn + "` " + j.Operator + " `" + j.RightTable + "`.`" + j.RightColumn + "` AND "
	}
	// Remove the trailing AND
	return q[0 : len(q)-4]
}

// Add support for BETWEEN x.x
func (a *MysqlAdapter) buildJoinWhere(where string) (q string, err error) {
	if len(where) != 0 {
		q = " WHERE"
		for _, loc := range processWhere(where) {
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					q += " " + token.Contents
				case TokenColumn:
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						q += " `" + halves[0] + "`.`" + halves[1] + "`"
					} else {
						q += " `" + token.Contents + "`"
					}
				case TokenString:
					q += " '" + token.Contents + "'"
				default:
					return q, errors.New("This token doesn't exist o_o")
				}
			}
			q += " AND"
		}
		q = q[0 : len(q)-4]
	}
	return q, nil
}

func (a *MysqlAdapter) buildLimit(limit string) (q string) {
	if limit != "" {
		q = " LIMIT " + limit
	}
	return q
}

func (a *MysqlAdapter) buildJoinColumns(cols string) (q string) {
	for _, col := range processColumns(cols) {
		// TODO: Move the stirng and number logic to processColumns?
		// TODO: Error if [0] doesn't exist
		firstChar := col.Left[0]
		if firstChar == '\'' {
			col.Type = TokenString
		} else {
			_, err := strconv.Atoi(string(firstChar))
			if err == nil {
				col.Type = TokenNumber
			}
		}

		// Escape the column names, just in case we've used a reserved keyword
		source := col.Left
		if col.Table != "" {
			source = "`" + col.Table + "`.`" + source + "`"
		} else if col.Type != TokenFunc && col.Type != TokenNumber && col.Type != TokenSub && col.Type != TokenString {
			source = "`" + source + "`"
		}

		var alias string
		if col.Alias != "" {
			alias = " AS `" + col.Alias + "`"
		}
		q += " " + source + alias + ","
	}
	return q[0 : len(q)-1]
}

func (a *MysqlAdapter) SimpleInsertInnerJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	whereStr, err := a.buildJoinWhere(sel.Where)
	if err != nil {
		return "", err
	}

	q := "INSERT INTO `" + ins.Table + "`(" + a.buildColumns(ins.Columns) + ") SELECT" + a.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table1 + "` INNER JOIN `" + sel.Table2 + "` ON " + a.buildJoiners(sel.Joiners) + whereStr + a.buildOrderby(sel.Orderby) + a.buildLimit(sel.Limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleCount(name, table, where, limit string) (q string, err error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	sb := &strings.Builder{}
	err = a.buildWhere(where, sb)
	if err != nil {
		return "", err
	}

	q = "SELECT COUNT(*) FROM `" + table + "`" + sb.String() + a.buildLimit(limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MysqlAdapter) Builder() *prebuilder {
	return &prebuilder{a}
}

func (a *MysqlAdapter) Write() error {
	var stmts, body string
	for _, name := range a.BufferOrder {
		if name[0] == '_' {
			continue
		}
		stmt := a.Buffer[name]
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
func (a *MysqlAdapter) pushStatement(name, stype, q string) {
	if name == "" {
		return
	}
	a.Buffer[name] = DBStmt{q, stype}
	a.BufferOrder = append(a.BufferOrder, name)
}

func (a *MysqlAdapter) stringyType(ct string) bool {
	ct = strings.ToLower(ct)
	return ct == "varchar" || ct == "tinytext" || ct == "text" || ct == "mediumtext" || ct == "longtext" || ct == "char" || ct == "datetime" || ct == "timestamp" || ct == "time" || ct == "date"
}
