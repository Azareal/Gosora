/* WIP Under Construction */
package qgen // import "github.com/Azareal/Gosora/query_gen"

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

var Registry []Adapter
var ErrNoAdapter = errors.New("This adapter doesn't exist")

type DBTableColumn struct {
	Name          string
	Type          string
	Size          int
	Null          bool
	AutoIncrement bool
	Default       string
}

type DBTableKey struct {
	Columns string
	Type    string
}

type DBSelect struct {
	Table   string
	Columns string
	Where   string
	Orderby string
	Limit   string
}

type DBJoin struct {
	Table1  string
	Table2  string
	Columns string
	Joiners string
	Where   string
	Orderby string
	Limit   string
}

type DBInsert struct {
	Table   string
	Columns string
	Fields  string
}

type DBColumn struct {
	Table string
	Left  string // Could be a function or a column, so I'm naming this Left
	Alias string // aka AS Blah, if it's present
	Type  string // function or column
}

type DBField struct {
	Name string
	Type string
}

type DBWhere struct {
	Expr []DBToken // Simple expressions, the innards of functions are opaque for now.
}

type DBJoiner struct {
	LeftTable   string
	LeftColumn  string
	RightTable  string
	RightColumn string
	Operator    string
}

type DBOrder struct {
	Column string
	Order  string
}

type DBToken struct {
	Contents string
	Type     string // function, operator, column, number, string, substitute
}

type DBSetter struct {
	Column string
	Expr   []DBToken // Simple expressions, the innards of functions are opaque for now.
}

type DBLimit struct {
	Offset   string // ? or int
	MaxCount string // ? or int
}

type DBStmt struct {
	Contents string
	Type     string // create-table, insert, update, delete
}

// TODO: Add the DropTable, TableExists, AddColumn, ColumnExists, and RemoveColumn methods
type Adapter interface {
	GetName() string
	BuildConn(config map[string]string) (*sql.DB, error)
	DbVersion() string

	DropTable(name string, table string) (string, error)
	CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error)
	// TODO: Some way to add indices and keys
	// TODO: Test this
	AddColumn(name string, table string, column DBTableColumn, key *DBTableKey) (string, error)
	AddIndex(name string, table string, iname string, colname string) (string, error)
	SimpleInsert(name string, table string, columns string, fields string) (string, error)
	SimpleUpdate(up *updatePrebuilder) (string, error)
	SimpleUpdateSelect(up *updatePrebuilder) (string, error) // ! Experimental
	SimpleDelete(name string, table string, where string) (string, error)
	Purge(name string, table string) (string, error)
	SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error)
	SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error)
	SimpleInnerJoin(string, string, string, string, string, string, string, string) (string, error)
	SimpleInsertSelect(string, DBInsert, DBSelect) (string, error)
	SimpleInsertLeftJoin(string, DBInsert, DBJoin) (string, error)
	SimpleInsertInnerJoin(string, DBInsert, DBJoin) (string, error)
	SimpleCount(string, string, string, string) (string, error)

	ComplexSelect(*selectPrebuilder) (string, error)

	Builder() *prebuilder
	Write() error
}

func GetAdapter(name string) (adap Adapter, err error) {
	for _, adapter := range Registry {
		if adapter.GetName() == name {
			return adapter, nil
		}
	}
	return adap, ErrNoAdapter
}

type QueryPlugin interface {
	Hook(name string, args ...interface{}) error
	Write() error
}

type MySQLUpsertCallback struct {
	stmt *sql.Stmt
}

func (double *MySQLUpsertCallback) Exec(args ...interface{}) (res sql.Result, err error) {
	if len(args) < 2 {
		return res, errors.New("Need two or more arguments")
	}
	args = args[:len(args)-1]
	return double.stmt.Exec(append(args, args...)...)
}

func PrepareMySQLUpsertCallback(db *sql.DB, query string) (*MySQLUpsertCallback, error) {
	stmt, err := db.Prepare(query)
	return &MySQLUpsertCallback{stmt}, err
}

type LitStr string

// TODO: Test this
func InterfaceMapToInsertStrings(data map[string]interface{}, order string) (cols string, values string) {
	var done = make(map[string]bool)
	var addValue = func(value interface{}) {
		switch value := value.(type) {
		case string:
			values += "'" + strings.Replace(value, "'", "\\'", -1) + "',"
		case int:
			values += strconv.Itoa(value) + ","
		case LitStr:
			values += string(value) + ","
		case bool:
			if value {
				values += "1,"
			} else {
				values += "0,"
			}
		}
	}

	// Add the ordered items
	for _, col := range strings.Split(order, ",") {
		col = strings.TrimSpace(col)
		value, ok := data[col]
		if ok {
			cols += col + ","
			addValue(value)
			done[col] = true
		}
	}

	// Go over any unordered items and add them at the end
	if len(data) > len(done) {
		for col, value := range data {
			_, ok := done[col]
			if ok {
				continue
			}
			cols += col + ","
			addValue(value)
		}
	}

	if cols != "" {
		cols = cols[:len(cols)-1]
		values = values[:len(values)-1]
	}
	return cols, values
}
