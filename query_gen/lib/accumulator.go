/* WIP: A version of the builder which accumulates errors, we'll see if we can't unify the implementations at some point */
package qgen

import (
	"database/sql"
	"log"
)

var LogPrepares = true

type Accumulator struct {
	conn     *sql.DB
	adapter  Adapter
	firstErr error
}

func (build *Accumulator) SetConn(conn *sql.DB) {
	build.conn = conn
}

func (build *Accumulator) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	build.adapter = adap
	return nil
}

func (build *Accumulator) GetAdapter() Adapter {
	return build.adapter
}

func (build *Accumulator) FirstError() error {
	return build.firstErr
}

func (build *Accumulator) recordError(err error) {
	if err == nil {
		return
	}
	if build.firstErr == nil {
		build.firstErr = err
	}
}

func (build *Accumulator) prepare(res string, err error) *sql.Stmt {
	if LogPrepares {
		log.Print("res: ", res)
	}
	if err != nil {
		build.recordError(err)
		return nil
	}
	stmt, err := build.conn.Prepare(res)
	build.recordError(err)
	return stmt
}

func (build *Accumulator) Tx(handler func(*TransactionBuilder) error) {
	tx, err := build.conn.Begin()
	if err != nil {
		build.recordError(err)
		return
	}
	err = handler(&TransactionBuilder{tx, build.adapter, nil})
	if err != nil {
		tx.Rollback()
		build.recordError(err)
		return
	}
	build.recordError(tx.Commit())
}

func (build *Accumulator) SimpleSelect(table string, columns string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit))
}

func (build *Accumulator) SimpleCount(table string, where string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleCount("_builder", table, where, limit))
}

func (build *Accumulator) SimpleLeftJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *Accumulator) SimpleInnerJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *Accumulator) CreateTable(table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) *sql.Stmt {
	return build.prepare(build.adapter.CreateTable("_builder", table, charset, collation, columns, keys))
}

func (build *Accumulator) SimpleInsert(table string, columns string, fields string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsert("_builder", table, columns, fields))
}

func (build *Accumulator) SimpleInsertSelect(ins DBInsert, sel DBSelect) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertSelect("_builder", ins, sel))
}

func (build *Accumulator) SimpleInsertLeftJoin(ins DBInsert, sel DBJoin) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertLeftJoin("_builder", ins, sel))
}

func (build *Accumulator) SimpleInsertInnerJoin(ins DBInsert, sel DBJoin) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertInnerJoin("_builder", ins, sel))
}

func (build *Accumulator) SimpleUpdate(table string, set string, where string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleUpdate("_builder", table, set, where))
}

func (build *Accumulator) SimpleDelete(table string, where string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleDelete("_builder", table, where))
}

// I don't know why you need this, but here it is x.x
func (build *Accumulator) Purge(table string) *sql.Stmt {
	return build.prepare(build.adapter.Purge("_builder", table))
}

func (build *Accumulator) prepareTx(tx *sql.Tx, res string, err error) (stmt *sql.Stmt) {
	if err != nil {
		build.recordError(err)
		return nil
	}
	stmt, err = tx.Prepare(res)
	build.recordError(err)
	return stmt
}

// These ones support transactions
func (build *Accumulator) SimpleSelectTx(tx *sql.Tx, table string, columns string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleCountTx(tx *sql.Tx, table string, where string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleCount("_builder", table, where, limit)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleLeftJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInnerJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) CreateTableTx(tx *sql.Tx, table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt) {
	res, err := build.adapter.CreateTable("_builder", table, charset, collation, columns, keys)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertTx(tx *sql.Tx, table string, columns string, fields string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertSelectTx(tx *sql.Tx, ins DBInsert, sel DBSelect) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertSelect("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertLeftJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertInnerJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleUpdateTx(tx *sql.Tx, table string, set string, where string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleUpdate("_builder", table, set, where)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleDeleteTx(tx *sql.Tx, table string, where string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	return build.prepareTx(tx, res, err)
}

// I don't know why you need this, but here it is x.x
func (build *Accumulator) PurgeTx(tx *sql.Tx, table string) (stmt *sql.Stmt) {
	res, err := build.adapter.Purge("_builder", table)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) Delete(table string) *accDeleteBuilder {
	return &accDeleteBuilder{table, "", build}
}

func (build *Accumulator) Update(table string) *accUpdateBuilder {
	return &accUpdateBuilder{table, "", "", build}
}

func (build *Accumulator) Select(table string) *AccSelectBuilder {
	return &AccSelectBuilder{table, "", "", "", "", nil, nil, "", build}
}

func (build *Accumulator) Insert(table string) *accInsertBuilder {
	return &accInsertBuilder{table, "", "", build}
}

func (build *Accumulator) Count(table string) *accCountBuilder {
	return &accCountBuilder{table, "", "", build}
}
