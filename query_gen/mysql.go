/* WIP Under Construction */
package qgen

import (
	"database/sql"
	"errors"

	//"fmt"
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

func (a *MysqlAdapter) CreateTable(name, table, charset, collation string, cols []DBTableColumn, keys []DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(cols) == 0 {
		return "", errors.New("You can't have a table with no columns")
	}

	var qsb strings.Builder
	//q := "CREATE TABLE `" + table + "`("
	w := func(s string) {
		qsb.WriteString(s)
	}
	w("CREATE TABLE `")
	w(table)
	w("`(")
	for i, col := range cols {
		if i != 0 {
			w(",\n\t`")
		} else {
			w("\n\t`")
		}
		col, size, end := a.parseColumn(col)
		//q += "\n\t`" + col.Name + "` " + col.Type + size + end + ","
		w(col.Name)
		w("` ")
		w(col.Type)
		w(size)
		w(end)
	}

	if len(keys) > 0 {
		/*if len(cols) > 0 {
			w(",")
		}*/
		for _, k := range keys {
			/*if ii != 0 {
				w(",\n\t")
			} else {
				w("\n\t")
			}*/
			w(",\n\t")
			//q += "\n\t" + key.Type
			w(k.Type)
			if k.Type != "unique" {
				//q += " key"
				w(" key")
			}
			if k.Type == "foreign" {
				cols := strings.Split(k.Columns, ",")
				//q += "(`" + cols[0] + "`) REFERENCES `" + k.FTable + "`(`" + cols[1] + "`)"
				w("(`")
				w(cols[0])
				w("`) REFERENCES `")
				w(k.FTable)
				w("`(`")
				w(cols[1])
				w("`)")
				if k.Cascade {
					//q += " ON DELETE CASCADE"
					w(" ON DELETE CASCADE")
				}
				//q += ","
			} else {
				//q += "("
				w("(")
				for i, col := range strings.Split(k.Columns, ",") {
					/*if i != 0 {
						q += ",`" + col + "`"
					} else {
						q += "`" + col + "`"
					}*/
					if i != 0 {
						w(",`")
					} else {
						w("`")
					}
					w(col)
					w("`")
				}
				//q += "),"
				w(")")
			}
		}
	}

	//q = q[0:len(q)-1] + "\n)"
	w("\n)")
	if charset != "" {
		//q += " CHARSET=" + charset
		w(" CHARSET=")
		w(charset)
	}
	if collation != "" {
		//q += " COLLATE " + collation
		w(" COLLATE ")
		w(collation)
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	//q += ";"
	w(";")
	q := qsb.String()
	a.pushStatement(name, "create-table", q)
	return q, nil
}

func (a *MysqlAdapter) DropColumn(name, table, colName string) (string, error) {
	q := "ALTER TABLE `" + table + "` DROP COLUMN `" + colName + "`;"
	a.pushStatement(name, "drop-column", q)
	return q, nil
}

// ! Currently broken in MariaDB. Planned.
func (a *MysqlAdapter) RenameColumn(name, table, oldName, newName string) (string, error) {
	q := "ALTER TABLE `" + table + "` RENAME COLUMN `" + oldName + "` TO `" + newName + "`;"
	a.pushStatement(name, "rename-column", q)
	return q, nil
}

func (a *MysqlAdapter) ChangeColumn(name, table, colName string, col DBTableColumn) (string, error) {
	col.Default = ""
	col, size, end := a.parseColumn(col)
	q := "ALTER TABLE `" + table + "` CHANGE COLUMN `" + colName + "` `" + col.Name + "` " + col.Type + size + end
	a.pushStatement(name, "change-column", q)
	return q, nil
}

func (a *MysqlAdapter) SetDefaultColumn(name, table, colName, colType, defaultStr string) (string, error) {
	if defaultStr == "" {
		defaultStr = "''"
	}
	// TODO: Exclude the other variants of text like mediumtext and longtext too
	expr := ""
	/*if colType == "datetime" && defaultStr[len(defaultStr)-1] == ')' {
		end += defaultStr
	} else */if a.stringyType(colType) && defaultStr != "''" {
		expr += "'" + defaultStr + "'"
	} else {
		expr += defaultStr
	}
	q := "ALTER TABLE `" + table + "` ALTER COLUMN `" + colName + "` SET DEFAULT " + expr + ";"
	a.pushStatement(name, "set-default-column", q)
	return q, nil
}

func (a *MysqlAdapter) parseColumn(col DBTableColumn) (ocol DBTableColumn, size, end string) {
	// Make it easier to support Cassandra in the future
	if col.Type == "createdAt" {
		col.Type = "datetime"
		// MySQL doesn't support this x.x
		/*if col.Default == "" {
			col.Default = "UTC_TIMESTAMP()"
		}*/
	} else if col.Type == "json" {
		col.Type = "text"
	}
	if col.Size > 0 {
		size = "(" + strconv.Itoa(col.Size) + ")"
	}

	// TODO: Exclude the other variants of text like mediumtext and longtext too
	if col.Default != "" && col.Type != "text" {
		end = " DEFAULT "
		/*if col.Type == "datetime" && col.Default[len(col.Default)-1] == ')' {
			end += column.Default
		} else */if a.stringyType(col.Type) && col.Default != "''" {
			end += "'" + col.Default + "'"
		} else {
			end += col.Default
		}
	}

	if col.Null {
		end += " null"
	} else {
		end += " not null"
	}
	if col.AutoIncrement {
		end += " AUTO_INCREMENT"
	}
	return col, size, end
}

// TODO: Support AFTER column
// TODO: Test to make sure everything works here
func (a *MysqlAdapter) AddColumn(name, table string, col DBTableColumn, key *DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	col, size, end := a.parseColumn(col)
	q := "ALTER TABLE `" + table + "` ADD COLUMN " + "`" + col.Name + "` " + col.Type + size + end

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
func (a *MysqlAdapter) AddKey(name, table, cols string, key DBTableKey) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if cols == "" {
		return "", errors.New("You need to specify columns")
	}

	var colstr string
	for _, col := range strings.Split(cols, ",") {
		colstr += "`" + col + "`,"
	}
	if len(colstr) > 1 {
		colstr = colstr[:len(colstr)-1]
	}

	var q string
	if key.Type == "fulltext" {
		q = "ALTER TABLE `" + table + "` ADD FULLTEXT(" + colstr + ")"
	} else {
		return "", errors.New("Only fulltext is supported by AddKey right now")
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-key", q)
	return q, nil
}

func (a *MysqlAdapter) RemoveIndex(name, table, iname string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if iname == "" {
		return "", errors.New("You need a name for the index")
	}
	q := "ALTER TABLE `" + table + "` DROP INDEX `" + iname + "`"

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "remove-index", q)
	return q, nil
}

func (a *MysqlAdapter) AddForeignKey(name, table, col, ftable, fcolumn string, cascade bool) (out string, e error) {
	c := func(str string, val bool) {
		if e != nil || !val {
			return
		}
		e = errors.New("You need a " + str + " for this table")
	}
	c("name", table == "")
	c("col", col == "")
	c("ftable", ftable == "")
	c("fcolumn", fcolumn == "")
	if e != nil {
		return "", e
	}

	q := "ALTER TABLE `" + table + "` ADD CONSTRAINT `fk_" + col + "` FOREIGN KEY(`" + col + "`) REFERENCES `" + ftable + "`(`" + fcolumn + "`)"
	if cascade {
		q += " ON DELETE CASCADE"
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "add-foreign-key", q)
	return q, nil
}

const silen1 = len("INSERT INTO``()VALUES() ")

func (a *MysqlAdapter) SimpleInsert(name, table, cols, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}

	sb.Grow(silen1 + len(table))
	sb.WriteString("INSERT INTO`")
	sb.WriteString(table)
	if cols != "" {
		sb.WriteString("`(")
		sb.WriteString(a.buildColumns(cols))
		sb.WriteString(")VALUES(")
		fs := processFields(fields)
		sb.Grow(len(fs) * 3)
		for i, field := range fs {
			if i != 0 {
				sb.WriteString(",")
			}
			nameLen := len(field.Name)
			if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
				sb.WriteRune('\'')
				sb.WriteString(field.Name[1 : nameLen-1])
				sb.WriteRune('\'')
			} else if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
				sb.WriteRune('\'')
				sb.WriteString(strings.Replace(field.Name[1:nameLen-1], "'", "''", -1))
				sb.WriteRune('\'')
			} else {
				sb.WriteString(field.Name)
			}
		}
		sb.WriteString(")")
	} else {
		sb.WriteString("`()VALUES()")
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q := sb.String()
	queryStrPool.Put(sb)
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleBulkInsert(name, table, cols string, fieldSet []string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}

	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	sb.Grow(silen1 + len(table))
	sb.WriteString("INSERT INTO`")
	sb.WriteString(table)
	if cols != "" {
		sb.WriteString("`(")
		sb.WriteString(a.buildColumns(cols))
		sb.WriteString(")VALUES(")
		for oi, fields := range fieldSet {
			if oi != 0 {
				sb.WriteString(",(")
			}
			fs := processFields(fields)
			sb.Grow(len(fs) * 3)
			for i, field := range fs {
				if i != 0 {
					sb.WriteString(",")
				}
				nameLen := len(field.Name)
				if field.Name[0] == '"' && field.Name[nameLen-1] == '"' && nameLen >= 3 {
					//field.Name = "'" + field.Name[1:nameLen-1] + "'"
					sb.WriteString("'")
					sb.WriteString(field.Name[1 : nameLen-1])
					sb.WriteString("'")
					continue
				}
				if field.Name[0] == '\'' && field.Name[nameLen-1] == '\'' && nameLen >= 3 {
					//field.Name = "'" + strings.Replace(field.Name[1:nameLen-1], "'", "''", -1) + "'"
					sb.WriteString("'")
					sb.WriteString(strings.Replace(field.Name[1:nameLen-1], "'", "''", -1))
					sb.WriteString("'")
					continue
				}
				sb.WriteString(field.Name)
			}
			sb.WriteString(")")
		}
	} else {
		sb.WriteString("`()VALUES()")
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q := sb.String()
	queryStrPool.Put(sb)
	a.pushStatement(name, "bulk-insert", q)
	return q, nil
}

func (a *MysqlAdapter) buildColumns(cols string) string {
	if cols == "" {
		return ""
	}

	// Escape the column names, just in case we've used a reserved keyword
	var cb strings.Builder
	pcols := processColumns(cols)
	var n int
	for i, col := range pcols {
		if i != 0 {
			n += 1
		}
		if col.Type == TokenFunc {
			n += len(col.Left)
		} else {
			n += len(col.Left) + 2
		}
	}
	cb.Grow(n)

	for i, col := range pcols {
		if col.Type == TokenFunc {
			if i != 0 {
				cb.WriteString(",")
			}
			//q += col.Left + ","
			cb.WriteString(col.Left)
		} else {
			//q += "`" + col.Left + "`,"
			if i != 0 {
				cb.WriteString(",`")
			} else {
				cb.WriteString("`")
			}
			cb.WriteString(col.Left)
			cb.WriteString("`")
		}
	}

	return cb.String()
}

// ! DEPRECATED
func (a *MysqlAdapter) SimpleReplace(name, table, cols, fields string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(cols) == 0 {
		return "", errors.New("No columns found for SimpleInsert")
	}
	if len(fields) == 0 {
		return "", errors.New("No input data found for SimpleInsert")
	}

	q := "REPLACE INTO `" + table + "`(" + a.buildColumns(cols) + ") VALUES ("
	for _, field := range processFields(fields) {
		q += field.Name + ","
	}
	q = q[0:len(q)-1] + ")"

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

	var insertColumns, insertValues string
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

const sulen1 = len("UPDATE `` SET ")

func (a *MysqlAdapter) SimpleUpdate(up *updatePrebuilder) (string, error) {
	if up.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if up.set == "" {
		return "", errors.New("You need to set data in this update statement")
	}

	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
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

	e := a.buildFlexiWhereSb(sb, up.where, up.dateCutoff)
	if e != nil {
		return sb.String(), e
	}

	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	q := sb.String()
	queryStrPool.Put(sb)
	a.pushStatement(up.name, "update", q)
	return q, nil
}

const sdlen1 = len("DELETE FROM `` WHERE")

func (a *MysqlAdapter) SimpleDelete(name, table, where string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if where == "" {
		return "", errors.New("You need to specify what data you want to delete")
	}
	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	sb.Grow(sdlen1 + len(table))
	sb.WriteString("DELETE FROM `")
	sb.WriteString(table)
	sb.WriteString("` WHERE")

	// Add support for BETWEEN x.x
	for i, loc := range processWhere(where) {
		if i != 0 {
			sb.WriteString(" AND")
		}
		for _, token := range loc.Expr {
			switch token.Type {
			case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot:
				sb.WriteRune(' ')
				sb.WriteString(token.Contents)
			case TokenColumn:
				sb.WriteString(" `")
				sb.WriteString(token.Contents)
				sb.WriteRune('`')
			case TokenString:
				sb.WriteString(" '")
				sb.WriteString(token.Contents)
				sb.WriteRune('\'')
			default:
				panic("This token doesn't exist o_o")
			}
		}
	}

	q := strings.TrimSpace(sb.String())
	queryStrPool.Put(sb)
	// TODO: Shunt the table name logic and associated stmt list up to the a higher layer to reduce the amount of unnecessary overhead in the builder / accumulator
	a.pushStatement(name, "delete", q)
	return q, nil
}

const cdlen1 = len("DELETE FROM ``")

func (a *MysqlAdapter) ComplexDelete(b *deletePrebuilder) (string, error) {
	if b.table == "" {
		return "", errors.New("You need a name for this table")
	}
	if b.where == "" && b.dateCutoff == nil {
		return "", errors.New("You need to specify what data you want to delete")
	}
	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	sb.Grow(cdlen1 + len(b.table))
	sb.WriteString("DELETE FROM `")
	sb.WriteString(b.table)
	sb.WriteRune('`')

	e := a.buildFlexiWhereSb(sb, b.where, b.dateCutoff)
	if e != nil {
		return sb.String(), e
	}
	q := sb.String()
	queryStrPool.Put(sb)

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
const FlexiHint1 = len(` <UTC_TIMESTAMP()-interval ?  `)

func (a *MysqlAdapter) buildFlexiWhere(where string, dateCutoff *dateCutoff) (q string, err error) {
	if len(where) == 0 && dateCutoff == nil {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString(" WHERE")
	if dateCutoff != nil {
		sb.Grow(6 + FlexiHint1)
		sb.WriteRune(' ')
		sb.WriteString(dateCutoff.Column)
		switch dateCutoff.Type {
		case 0:
			sb.WriteString(" BETWEEN (UTC_TIMESTAMP()-interval ")
			sb.WriteString(strconv.Itoa(dateCutoff.Quantity))
			sb.WriteString(" ")
			sb.WriteString(dateCutoff.Unit)
			sb.WriteString(") AND UTC_TIMESTAMP()")
		case 11:
			sb.WriteString("<UTC_TIMESTAMP()-interval ? ")
			sb.WriteString(dateCutoff.Unit)
		default:
			sb.WriteString("<UTC_TIMESTAMP()-interval ")
			sb.WriteString(strconv.Itoa(dateCutoff.Quantity))
			sb.WriteRune(' ')
			sb.WriteString(dateCutoff.Unit)
		}
	}
	if dateCutoff != nil && len(where) != 0 {
		sb.WriteString(" AND")
	}

	if len(where) != 0 {
		wh := processWhere(where)
		sb.Grow((len(wh) * 8) - 5)
		for i, loc := range wh {
			if i != 0 {
				sb.WriteString(" AND ")
			}
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
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
				default:
					return sb.String(), errors.New("This token doesn't exist o_o")
				}
			}
		}
	}
	return sb.String(), nil
}

func (a *MysqlAdapter) buildFlexiWhereSb(sb *strings.Builder, where string, dateCutoff *dateCutoff) (err error) {
	if len(where) == 0 && dateCutoff == nil {
		return nil
	}

	sb.WriteString(" WHERE")
	if dateCutoff != nil {
		sb.Grow(6 + FlexiHint1)
		sb.WriteRune(' ')
		sb.WriteString(dateCutoff.Column)
		switch dateCutoff.Type {
		case 0:
			sb.WriteString(" BETWEEN (UTC_TIMESTAMP()-interval ")
			sb.WriteString(strconv.Itoa(dateCutoff.Quantity))
			sb.WriteString(" ")
			sb.WriteString(dateCutoff.Unit)
			sb.WriteString(") AND UTC_TIMESTAMP()")
		case 11:
			sb.WriteString("<UTC_TIMESTAMP()-interval ? ")
			sb.WriteString(dateCutoff.Unit)
		default:
			sb.WriteString("<UTC_TIMESTAMP()-interval ")
			sb.WriteString(strconv.Itoa(dateCutoff.Quantity))
			sb.WriteRune(' ')
			sb.WriteString(dateCutoff.Unit)
		}
	}
	if dateCutoff != nil && len(where) != 0 {
		sb.WriteString(" AND")
	}

	if len(where) != 0 {
		wh := processWhere(where)
		sb.Grow((len(wh) * 8) - 5)
		for i, loc := range wh {
			if i != 0 {
				sb.WriteString(" AND ")
			}
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
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
				default:
					return errors.New("This token doesn't exist o_o")
				}
			}
		}
	}
	return nil
}

func (a *MysqlAdapter) buildOrderby(orderby string) (q string) {
	if len(orderby) != 0 {
		var sb strings.Builder
		ord := processOrderby(orderby)
		sb.Grow(10 + (len(ord) * 8) - 1)
		sb.WriteString(" ORDER BY ")
		for i, col := range ord {
			// TODO: We might want to escape this column
			if i != 0 {
				sb.WriteString(",`")
			} else {
				sb.WriteString("`")
			}
			sb.WriteString(strings.Replace(col.Column, ".", "`.`", -1))
			sb.WriteString("` ")
			sb.WriteString(strings.ToUpper(col.Order))
		}
		q = sb.String()
	}
	return q
}

func (a *MysqlAdapter) buildOrderbySb(sb *strings.Builder, orderby string) {
	if len(orderby) != 0 {
		ord := processOrderby(orderby)
		sb.Grow(10 + (len(ord) * 8) - 1)
		sb.WriteString(" ORDER BY ")
		for i, col := range ord {
			// TODO: We might want to escape this column
			if i != 0 {
				sb.WriteString(",`")
			} else {
				sb.WriteString("`")
			}
			sb.WriteString(strings.Replace(col.Column, ".", "`.`", -1))
			sb.WriteString("` ")
			sb.WriteString(strings.ToUpper(col.Order))
		}
	}
}

func (a *MysqlAdapter) SimpleSelect(name, table, cols, where, orderby, limit string) (string, error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	if len(cols) == 0 {
		return "", errors.New("No columns found for SimpleSelect")
	}
	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	sb.WriteString("SELECT ")

	// Slice up the user friendly strings into something easier to process
	for i, col := range strings.Split(strings.TrimSpace(cols), ",") {
		if i != 0 {
			sb.WriteString("`,`")
		} else {
			sb.WriteRune('`')
		}
		sb.WriteString(strings.TrimSpace(col))
	}

	sb.WriteString("`FROM`")
	sb.WriteString(table)
	sb.WriteRune('`')
	err := a.buildWhere(where, sb)
	if err != nil {
		return "", err
	}
	a.buildOrderbySb(sb, orderby)
	a.buildLimitSb(sb, limit)

	q := strings.TrimSpace(sb.String())
	queryStrPool.Put(sb)
	a.pushStatement(name, "select", q)
	return q, nil
}

func (a *MysqlAdapter) ComplexSelect(preBuilder *selectPrebuilder) (out string, e error) {
	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	e = a.complexSelect(preBuilder, sb)
	out = sb.String()
	queryStrPool.Put(sb)
	a.pushStatement(preBuilder.name, "select", out)
	return out, e
}

const cslen1 = len("SELECT  FROM ``")
const cslen2 = len("WHERE``IN(")

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
		sb.WriteString("WHERE`")
		sb.WriteString(preBuilder.inColumn)
		sb.WriteString("`IN(")
		e := a.complexSelect(preBuilder.inChain, sb)
		if e != nil {
			return e
		}
		sb.WriteRune(')')
	} else {
		e := a.buildFlexiWhereSb(sb, preBuilder.where, preBuilder.dateCutoff)
		if e != nil {
			return e
		}
	}

	a.buildOrderbySb(sb, preBuilder.orderby)
	a.buildLimitSb(sb, preBuilder.limit)
	return nil
}

func (a *MysqlAdapter) SimpleLeftJoin(name, table1, table2, columns, joiners, where, orderby, limit string) (string, error) {
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

func (a *MysqlAdapter) SimpleInnerJoin(name, table1, table2, columns, joiners, where, orderby, limit string) (string, error) {
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
	e := a.buildWhere(sel.Where, sb)
	if e != nil {
		return "", e
	}

	q := "INSERT INTO `" + ins.Table + "`(" + a.buildColumns(ins.Columns) + ") SELECT" + a.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table + "`" + sb.String() + a.buildOrderby(sel.Orderby) + a.buildLimit(sel.Limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "insert", q)
	return q, nil
}

func (a *MysqlAdapter) SimpleInsertLeftJoin(name string, ins DBInsert, sel DBJoin) (string, error) {
	whereStr, e := a.buildJoinWhere(sel.Where)
	if e != nil {
		return "", e
	}

	q := "INSERT INTO `" + ins.Table + "`(" + a.buildColumns(ins.Columns) + ") SELECT" + a.buildJoinColumns(sel.Columns) + " FROM `" + sel.Table1 + "` LEFT JOIN `" + sel.Table2 + "` ON " + a.buildJoiners(sel.Joiners) + whereStr + a.buildOrderby(sel.Orderby) + a.buildLimit(sel.Limit)
	q = strings.TrimSpace(q)
	a.pushStatement(name, "insert", q)
	return q, nil
}

// TODO: Make this more consistent with the other build* methods?
func (a *MysqlAdapter) buildJoiners(joiners string) (q string) {
	var qb strings.Builder
	for i, j := range processJoiner(joiners) {
		//q += "`" + j.LeftTable + "`.`" + j.LeftColumn + "` " + j.Operator + " `" + j.RightTable + "`.`" + j.RightColumn + "` AND "
		if i != 0 {
			qb.WriteString("AND`")
		} else {
			qb.WriteString("`")
		}
		qb.WriteString(j.LeftTable)
		qb.WriteString("`.`")
		qb.WriteString(j.LeftColumn)
		qb.WriteString("` ")
		qb.WriteString(j.Operator)
		qb.WriteString(" `")
		qb.WriteString(j.RightTable)
		qb.WriteString("`.`")
		qb.WriteString(j.RightColumn)
		qb.WriteString("`")
	}
	return qb.String()
}

// Add support for BETWEEN x.x
func (a *MysqlAdapter) buildJoinWhere(where string) (q string, e error) {
	if len(where) != 0 {
		var qsb strings.Builder
		ws := processWhere(where)
		qsb.Grow(6 + (len(ws) * 5))
		qsb.WriteString(" WHERE")
		for i, loc := range ws {
			if i != 0 {
				qsb.WriteString(" AND")
			}
			for _, token := range loc.Expr {
				switch token.Type {
				case TokenFunc, TokenOp, TokenNumber, TokenSub, TokenOr, TokenNot, TokenLike:
					qsb.WriteRune(' ')
					qsb.WriteString(token.Contents)
				case TokenColumn:
					qsb.WriteString(" `")
					halves := strings.Split(token.Contents, ".")
					if len(halves) == 2 {
						qsb.WriteString(halves[0])
						qsb.WriteString("`.`")
						qsb.WriteString(halves[1])
					} else {
						qsb.WriteString(token.Contents)
					}
					qsb.WriteRune('`')
				case TokenString:
					qsb.WriteString(" '")
					qsb.WriteString(token.Contents)
					qsb.WriteRune('\'')
				default:
					return qsb.String(), errors.New("This token doesn't exist o_o")
				}
			}
		}
		return qsb.String(), nil
	}
	return q, nil
}

func (a *MysqlAdapter) buildLimit(limit string) (q string) {
	if limit != "" {
		q = " LIMIT " + limit
	}
	return q
}

func (a *MysqlAdapter) buildLimitSb(sb *strings.Builder, limit string) {
	if limit != "" {
		sb.WriteString(" LIMIT ")
		sb.WriteString(limit)
	}
}

func (a *MysqlAdapter) buildJoinColumns(cols string) (q string) {
	for _, col := range processColumns(cols) {
		// TODO: Move the stirng and number logic to processColumns?
		// TODO: Error if [0] doesn't exist
		firstChar := col.Left[0]
		if firstChar == '\'' {
			col.Type = TokenString
		} else {
			_, e := strconv.Atoi(string(firstChar))
			if e == nil {
				col.Type = TokenNumber
			}
		}

		// Escape the column names, just in case we've used a reserved keyword
		source := col.Left
		if col.Table != "" {
			source = "`" + col.Table + "`.`" + source + "`"
		} else if col.Type != TokenScope && col.Type != TokenFunc && col.Type != TokenNumber && col.Type != TokenSub && col.Type != TokenString {
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

const sclen1 = len("SELECT COUNT(*) FROM``")

func (a *MysqlAdapter) SimpleCount(name, table, where, limit string) (q string, e error) {
	if table == "" {
		return "", errors.New("You need a name for this table")
	}
	var sb *strings.Builder
	ii := queryStrPool.Get()
	if ii == nil {
		sb = &strings.Builder{}
	} else {
		sb = ii.(*strings.Builder)
		sb.Reset()
	}
	sb.Grow(sclen1 + len(table))
	sb.WriteString("SELECT COUNT(*) FROM`")
	sb.WriteString(table)
	sb.WriteRune('`')
	if e = a.buildWhere(where, sb); e != nil {
		return "", e
	}
	a.buildLimitSb(sb, limit)

	q = strings.TrimSpace(sb.String())
	queryStrPool.Put(sb)
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
	dl("Preparing ` + name + ` statement.")
	stmts.` + name + `, e = qgen.PrepareMySQLUpsertCallback(db,"` + stmt.Contents + `")
	if e != nil {
		l("Error in ` + name + ` statement.")
		return e
	}
	`
		} else if stmt.Type != "create-table" {
			stmts += "\t" + name + " *sql.Stmt\n"
			body += `	
	dl("Preparing ` + name + ` statement.")
	stmts.` + name + `, e = db.Prepare("` + stmt.Contents + `")
	if e != nil {
		l("Error in ` + name + ` statement.")
		return e
	}
	`
		}
	}

	// TODO: Move these custom queries out of this file
	out := `// +build !pgsql,!mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import(
	"log"
	"database/sql"
	c "github.com/Azareal/Gosora/common"
	//"github.com/Azareal/Gosora/query_gen"
)

// nolint
type Stmts struct {
` + stmts + `
	getActivityFeedByWatcher *sql.Stmt
	//getActivityFeedByWatcherAfter *sql.Stmt
	getActivityCountByWatcher *sql.Stmt

	Mocks bool
}

// nolint
func _gen_mysql() (e error) {
	dl := c.DebugLog
	dl("Building the generated statements")
	l := log.Print
	_ = l
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
