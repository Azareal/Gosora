/* WIP Under Construction */
package qgen

import "errors"

var DB_Registry []DB_Adapter
var No_Adapter = errors.New("This adapter doesn't exist")

type DB_Table_Column struct
{
	Name string
	Type string
	Size int
	Null bool
	Auto_Increment bool
	Default string
}

type DB_Table_Key struct
{
	Columns string
	Type string
}

type DB_Select struct
{
	Table string
	Columns string
	Where string
	Orderby string
	Limit string
}

type DB_Join struct
{
	Table1 string
	Table2 string
	Columns string
	Joiners string
	Where string
	Orderby string
	Limit string
}

type DB_Insert struct
{
	Table string
	Columns string
	Fields string
}

type DB_Column struct
{
	Table string
	Left string // Could be a function or a column, so I'm naming this Left
	Alias string // aka AS Blah, if it's present
	Type string // function or column
}

type DB_Field struct
{
	Name string
	Type string
}

type DB_Where struct
{
	Expr []DB_Token // Simple expressions, the innards of functions are opaque for now.
}

type DB_Joiner struct
{
	LeftTable string
	LeftColumn string
	RightTable string
	RightColumn string
	Operator string
}

type DB_Order struct
{
	Column string
	Order string
}

type DB_Token struct {
	Contents string
	Type string // function, operator, column, number, string, substitute
}

type DB_Setter struct {
	Column string
	Expr []DB_Token // Simple expressions, the innards of functions are opaque for now.
}

type DB_Limit struct {
	Offset string // ? or int
	MaxCount string // ? or int
}

type DB_Adapter interface {
	GetName() string
	CreateTable(name string, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (string, error)
	SimpleInsert(name string, table string, columns string, fields string) (string, error)
	SimpleReplace(string,string,string,string) (string, error)
	SimpleUpdate(string,string,string,string) (string, error)
	SimpleDelete(string,string,string) (string, error)
	Purge(string,string) (string, error)
	SimpleSelect(string,string,string,string,string,string) (string, error)
	SimpleLeftJoin(string,string,string,string,string,string,string,string) (string, error)
	SimpleInnerJoin(string,string,string,string,string,string,string,string) (string, error)
	SimpleInsertSelect(string,DB_Insert,DB_Select) (string,error)
	SimpleInsertLeftJoin(string,DB_Insert,DB_Join) (string,error)
	SimpleInsertInnerJoin(string,DB_Insert,DB_Join) (string,error)
	SimpleCount(string,string,string,string) (string, error)
	Write() error
	
	// TO-DO: Add a simple query builder
}

func GetAdapter(name string) (adap DB_Adapter, err error) {
	for _, adapter := range DB_Registry {
		if adapter.GetName() == name {
			return adapter, nil
		}
	}
	return adap, No_Adapter
}
