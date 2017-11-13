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
	adapter Adapter
}

func (build *builder) Accumulator() *Accumulator {
	return &Accumulator{build.conn, build.adapter, nil}
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

func (build *builder) GetAdapter() Adapter {
	return build.adapter
}

func (build *builder) Begin() (*sql.Tx, error) {
	return build.conn.Begin()
}

func (build *builder) Tx(handler func(*TransactionBuilder) error) error {
	tx, err := build.conn.Begin()
	if err != nil {
		return err
	}
	err = handler(&TransactionBuilder{tx, build.adapter, nil})
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (build *builder) prepare(res string, err error) (*sql.Stmt, error) {
	if err != nil {
		return nil, err
	}
	return build.conn.Prepare(res)
}

func (build *builder) SimpleSelect(table string, columns string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit))
}

func (build *builder) SimpleCount(table string, where string, limit string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleCount("_builder", table, where, limit))
}

func (build *builder) SimpleLeftJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *builder) SimpleInnerJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *builder) CreateTable(table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.CreateTable("_builder", table, charset, collation, columns, keys))
}

func (build *builder) SimpleInsert(table string, columns string, fields string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleInsert("_builder", table, columns, fields))
}

func (build *builder) SimpleInsertSelect(ins DBInsert, sel DBSelect) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleInsertSelect("_builder", ins, sel))
}

func (build *builder) SimpleInsertLeftJoin(ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleInsertLeftJoin("_builder", ins, sel))
}

func (build *builder) SimpleInsertInnerJoin(ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleInsertInnerJoin("_builder", ins, sel))
}

func (build *builder) SimpleUpdate(table string, set string, where string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleUpdate("_builder", table, set, where))
}

func (build *builder) SimpleDelete(table string, where string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.SimpleDelete("_builder", table, where))
}

// I don't know why you need this, but here it is x.x
func (build *builder) Purge(table string) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.Purge("_builder", table))
}

func (build *builder) prepareTx(tx *sql.Tx, res string, err error) (*sql.Stmt, error) {
	if err != nil {
		return nil, err
	}
	return tx.Prepare(res)
}

// These ones support transactions
func (build *builder) SimpleSelectTx(tx *sql.Tx, table string, columns string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleCountTx(tx *sql.Tx, table string, where string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleCount("_builder", table, where, limit)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleLeftJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleInnerJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *builder) CreateTableTx(tx *sql.Tx, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.CreateTable("_builder", table, charset, collation, columns, keys)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleInsertTx(tx *sql.Tx, table string, columns string, fields string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleInsertSelectTx(tx *sql.Tx, ins DBInsert, sel DBSelect) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertSelect("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertLeftJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsertInnerJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleUpdateTx(tx *sql.Tx, table string, set string, where string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleUpdate("_builder", table, set, where)
	return build.prepareTx(tx, res, err)
}

func (build *builder) SimpleDeleteTx(tx *sql.Tx, table string, where string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	return build.prepareTx(tx, res, err)
}

// I don't know why you need this, but here it is x.x
func (build *builder) PurgeTx(tx *sql.Tx, table string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.Purge("_builder", table)
	return build.prepareTx(tx, res, err)
}
