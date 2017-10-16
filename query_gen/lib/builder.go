/* WIP Under Construction */
package qgen

//import "log"
import "database/sql"

var Builder *builder

func init() {
	Builder = &builder{conn: nil}
}

// A set of wrappers around the generator methods, so that we can use this inline in Gosora
type builder struct {
	conn    *sql.DB
	adapter DB_Adapter
}

func (build *builder) SetConn(conn *sql.DB) {
	build.conn = conn
}

func (build *builder) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	build.adapter = adap
	return nil
}

func (build *builder) GetAdapter() DB_Adapter {
	return build.adapter
}

func (build *builder) SimpleSelect(table string, columns string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleCount(table string, where string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleCount("_builder", table, where, limit)
	if err != nil {
		return stmt, err
	}
	//log.Print("res",res)
	return build.conn.Prepare(res)
}

func (build *builder) SimpleLeftJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	if err != nil {
		return stmt, err
	}
	//log.Print("res",res)
	return build.conn.Prepare(res)
}

func (build *builder) SimpleInnerJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	if err != nil {
		return stmt, err
	}
	//log.Print("res",res)
	return build.conn.Prepare(res)
}

func (build *builder) CreateTable(table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.CreateTable("_builder", table, charset, collation, columns, keys)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleInsert(table string, columns string, fields string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleInsertSelect(ins DB_Insert, sel DB_Select) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertSelect("_builder", ins, sel)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleInsertLeftJoin(ins DB_Insert, sel DB_Join) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertLeftJoin("_builder", ins, sel)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleInsertInnerJoin(ins DB_Insert, sel DB_Join) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertInnerJoin("_builder", ins, sel)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleUpdate(table string, set string, where string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleUpdate("_builder", table, set, where)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleDelete(table string, where string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}

// I don't know why you need this, but here it is x.x
func (build *builder) Purge(table string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.Purge("_builder", table)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}
