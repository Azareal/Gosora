/* WIP Under Construction */
package main

import "fmt"
import "strings"
import "errors"

func init() {
	db_registry = append(db_registry,&Mysql_Adapter{Name:"mysql"})
}

type Mysql_Adapter struct
{
	Name string
	Stmts string
	Body string
}

func (adapter *Mysql_Adapter) get_name() string {
	return adapter.Name
}

func (adapter *Mysql_Adapter) simple_insert(name string, table string, columns string, fields []string, quoteWhat []bool) error {
	if name == "" {
		return errors.New("You need a name for this statement")
	}
	if table == "" {
		return errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return errors.New("No columns found for simple_insert")
	}
	if len(fields) == 0 {
		return errors.New("No input data found for simple_insert")
	}
	
	// Slice up the user friendly strings into something easier to process
	var colslice []string = strings.Split(strings.TrimSpace(columns),",")
	
	var noquotes bool = (len(quoteWhat) == 0)
	var querystr string = "INSERT INTO " + table + "("
	
	// Escape the column names, just in case we've used a reserved keyword
	// TO-DO: Subparse the columns to allow functions and AS keywords
	for _, column := range colslice {
		querystr += "`" + column + "`,"
	}
	
	// Remove the trailing comma
	querystr = querystr[0:len(querystr) - 1]
	
	querystr += ") VALUES ("
	for fid, field := range fields {
		if !noquotes && quoteWhat[fid] {
			querystr += "'" + field + "',"
		} else {
			querystr += field + ","
		}
	}
	querystr = querystr[0:len(querystr) - 1]
	
	adapter.write_statement(name,querystr + ")")
	return nil
}

func (adapter *Mysql_Adapter) simple_update() error {
	return nil
}

func (adapter *Mysql_Adapter) simple_select(name string, table string, columns string, where string, orderby []DB_Order/*, offset int, maxCount int*/) error {
	if name == "" {
		return errors.New("You need a name for this statement")
	}
	if table == "" {
		return errors.New("You need a name for this table")
	}
	if len(columns) == 0 {
		return errors.New("No columns found for simple_select")
	}
	
	// Slice up the user friendly strings into something easier to process
	var colslice []string = strings.Split(strings.TrimSpace(columns),",")
	
	var querystr string = "SELECT "
	
	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range colslice {
		querystr += "`" + strings.TrimSpace(column) + "`,"
	}
	
	// Remove the trailing comma
	querystr = querystr[0:len(querystr) - 1]
	
	querystr += " FROM " + table
	if len(where) != 0 {
		querystr += " WHERE"
		fmt.Println("where",where)
		fmt.Println("_process_where(where)",_process_where(where))
		for _, loc := range _process_where(where) {
			var lquote, rquote string
			if loc.LeftType == "column" {
				lquote = "`"
			}
			if loc.RightType == "column" {
				rquote = "`"
			}
			querystr += " " + lquote + loc.Left + lquote + loc.Operator + " " + rquote + loc.Right + rquote + " AND "
		}
	}
	// Remove the trailing AND
	querystr = querystr[0:len(querystr) - 4]
	
	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range orderby {
			querystr += column.Column + " " + column.Order + ","
		}
	}
	querystr = querystr[0:len(querystr) - 1]
	
	adapter.write_statement(name,querystr)
	return nil
}

func (adapter *Mysql_Adapter) simple_left_join(name string, table1 string, table2 string, columns string, joiners []DB_Joiner, where string, orderby []DB_Order/*, offset int, maxCount int*/) error {
	if name == "" {
		return errors.New("You need a name for this statement")
	}
	if table1 == "" {
		return errors.New("You need a name for the left table")
	}
	if table2 == "" {
		return errors.New("You need a name for the right table")
	}
	if len(columns) == 0 {
		return errors.New("No columns found for simple_left_join")
	}
	if len(joiners) == 0 {
		return errors.New("No joiners found for simple_left_join")
	}
	
	// Slice up the user friendly strings into something easier to process
	var colslice []string = strings.Split(strings.TrimSpace(columns),",")
	
	var querystr string = "SELECT "
	
	// Escape the column names, just in case we've used a reserved keyword
	for _, column := range colslice {
		querystr += "`" + strings.TrimSpace(column) + "`,"
	}
	
	// Remove the trailing comma
	querystr = querystr[0:len(querystr) - 1]
	
	querystr += " FROM " + table1 + " LEFT JOIN " + table2 + " ON "
	for _, joiner := range joiners {
		querystr += "`" + joiner.Left + "`=`" + joiner.Right + "` AND "
	}
	// Remove the trailing AND
	querystr = querystr[0:len(querystr) - 4]
	
	if len(where) != 0 {
		querystr += " " + "WHERE"
		for _, loc := range _process_where(where) {
			var lquote, rquote string
			if loc.LeftType == "column" {
				lquote = "`"
			}
			if loc.RightType == "column" {
				rquote = "`"
			}
			querystr += " " + lquote + loc.Left + lquote + loc.Operator + " " + rquote + loc.Right + rquote + " AND "
		}
	}
	querystr = querystr[0:len(querystr) - 3]
	
	if len(orderby) != 0 {
		querystr += " ORDER BY "
		for _, column := range orderby {
			querystr += column.Column + " " + column.Order + ","
		}
	}
	querystr = querystr[0:len(querystr) - 1]
	
	adapter.write_statement(name,querystr)
	return nil
}

func (adapter *Mysql_Adapter) write() error {
	out := `/* This file was generated by Gosora's Query Generator */
package main

import "log"
import "database/sql"

` + adapter.Stmts + `
func gen_mysql() (err error) {
	if debug {
		log.Print("Building the generated statements")
	}
` + adapter.Body + `
	return nil
}
`
	return write_file("./gen_mysql.go", out)
}

// Internal method, not exposed in the interface
func (adapter *Mysql_Adapter) write_statement(name string, querystr string ) {
	adapter.Stmts += "var " + name + "_stmt *sql.Stmt\n"
	
	adapter.Body += `	
	log.Print("Preparing ` + name + ` statement.")
	` + name + `_stmt, err = db.Prepare("` + querystr + `")
	if err != nil {
		return err
	}
	`
}
