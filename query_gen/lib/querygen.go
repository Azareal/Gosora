/* WIP Under Construction */
package qgen

import (
	"database/sql"
	"errors"
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

type Adapter interface {
	GetName() string
	CreateTable(name string, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (string, error)
	SimpleInsert(name string, table string, columns string, fields string) (string, error)

	// ! DEPRECATED
	//SimpleReplace(name string, table string, columns string, fields string) (string, error)
	// ! NOTE: MySQL doesn't support upserts properly, so I'm removing this from the interface until we find a way to patch it in
	//SimpleUpsert(name string, table string, columns string, fields string, where string) (string, error)
	SimpleUpdate(name string, table string, set string, where string) (string, error)
	SimpleDelete(name string, table string, where string) (string, error)
	Purge(name string, table string) (string, error)
	SimpleSelect(name string, table string, columns string, where string, orderby string, limit string) (string, error)
	SimpleLeftJoin(name string, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (string, error)
	SimpleInnerJoin(string, string, string, string, string, string, string, string) (string, error)
	SimpleInsertSelect(string, DBInsert, DBSelect) (string, error)
	SimpleInsertLeftJoin(string, DBInsert, DBJoin) (string, error)
	SimpleInsertInnerJoin(string, DBInsert, DBJoin) (string, error)
	SimpleCount(string, string, string, string) (string, error)

	Select(name ...string) *selectPrebuilder
	Write() error

	// TODO: Add a simple query builder
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
