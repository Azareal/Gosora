/* WIP: A version of the builder which accumulates errors, we'll see if we can't unify the implementations at some point */
package qgen

import (
	"database/sql"
	"log"
	"strings"
)

var LogPrepares = true

// So we don't have to do the qgen.Builder.Accumulator() boilerplate all the time
func NewAcc() *Accumulator {
	return Builder.Accumulator()
}

type Accumulator struct {
	conn     *sql.DB
	adapter  Adapter
	firstErr error
}

func (acc *Accumulator) SetConn(conn *sql.DB) {
	acc.conn = conn
}

func (acc *Accumulator) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	acc.adapter = adap
	return nil
}

func (acc *Accumulator) GetAdapter() Adapter {
	return acc.adapter
}

func (acc *Accumulator) FirstError() error {
	return acc.firstErr
}

func (acc *Accumulator) RecordError(err error) {
	if err == nil {
		return
	}
	if acc.firstErr == nil {
		acc.firstErr = err
	}
}

func (acc *Accumulator) prepare(res string, err error) *sql.Stmt {
	// TODO: Can we make this less noisy on debug mode?
	if LogPrepares {
		log.Print("res: ", res)
	}
	if err != nil {
		acc.RecordError(err)
		return nil
	}
	stmt, err := acc.conn.Prepare(res)
	acc.RecordError(err)
	return stmt
}

func (acc *Accumulator) RawPrepare(res string) *sql.Stmt {
	return acc.prepare(res, nil)
}

func (acc *Accumulator) query(q string, args ...interface{}) (rows *sql.Rows, err error) {
	err = acc.FirstError()
	if err != nil {
		return rows, err
	}
	return acc.conn.Query(q, args...)
}

func (acc *Accumulator) exec(q string, args ...interface{}) (res sql.Result, err error) {
	err = acc.FirstError()
	if err != nil {
		return res, err
	}
	return acc.conn.Exec(q, args...)
}

func (acc *Accumulator) Tx(handler func(*TransactionBuilder) error) {
	tx, err := acc.conn.Begin()
	if err != nil {
		acc.RecordError(err)
		return
	}
	err = handler(&TransactionBuilder{tx, acc.adapter, nil})
	if err != nil {
		tx.Rollback()
		acc.RecordError(err)
		return
	}
	acc.RecordError(tx.Commit())
}

func (acc *Accumulator) SimpleSelect(table, columns, where, orderby, limit string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleSelect("", table, columns, where, orderby, limit))
}

func (acc *Accumulator) SimpleCount(table, where, limit string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleCount("", table, where, limit))
}

func (acc *Accumulator) SimpleLeftJoin(table1, table2, columns, joiners, where, orderby, limit string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleLeftJoin("", table1, table2, columns, joiners, where, orderby, limit))
}

func (acc *Accumulator) SimpleInnerJoin(table1, table2, columns, joiners, where, orderby, limit string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleInnerJoin("", table1, table2, columns, joiners, where, orderby, limit))
}

func (acc *Accumulator) CreateTable(table, charset, collation string, columns []DBTableColumn, keys []DBTableKey) *sql.Stmt {
	return acc.prepare(acc.adapter.CreateTable("", table, charset, collation, columns, keys))
}

func (acc *Accumulator) SimpleInsert(table, columns, fields string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleInsert("", table, columns, fields))
}

func (acc *Accumulator) SimpleBulkInsert(table, cols string, fieldSet []string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleBulkInsert("", table, cols, fieldSet))
}

func (acc *Accumulator) SimpleInsertSelect(ins DBInsert, sel DBSelect) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleInsertSelect("", ins, sel))
}

func (acc *Accumulator) SimpleInsertLeftJoin(ins DBInsert, sel DBJoin) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleInsertLeftJoin("", ins, sel))
}

func (acc *Accumulator) SimpleInsertInnerJoin(ins DBInsert, sel DBJoin) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleInsertInnerJoin("", ins, sel))
}

func (acc *Accumulator) SimpleUpdate(table, set, where string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleUpdate(qUpdate(table, set, where)))
}

func (acc *Accumulator) SimpleUpdateSelect(table, set, table2, cols, where, orderby, limit string) *sql.Stmt {
	pre := qUpdate(table, set, "").WhereQ(acc.GetAdapter().Builder().Select().Table(table2).Columns(cols).Where(where).Orderby(orderby).Limit(limit))
	return acc.prepare(acc.adapter.SimpleUpdateSelect(pre))
}

func (acc *Accumulator) SimpleDelete(table, where string) *sql.Stmt {
	return acc.prepare(acc.adapter.SimpleDelete("", table, where))
}

// I don't know why you need this, but here it is x.x
func (acc *Accumulator) Purge(table string) *sql.Stmt {
	return acc.prepare(acc.adapter.Purge("", table))
}

func (acc *Accumulator) prepareTx(tx *sql.Tx, res string, err error) (stmt *sql.Stmt) {
	if err != nil {
		acc.RecordError(err)
		return nil
	}
	stmt, err = tx.Prepare(res)
	acc.RecordError(err)
	return stmt
}

// These ones support transactions
func (acc *Accumulator) SimpleSelectTx(tx *sql.Tx, table, columns, where, orderby, limit string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleSelect("", table, columns, where, orderby, limit)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleCountTx(tx *sql.Tx, table, where, limit string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleCount("", table, where, limit)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleLeftJoinTx(tx *sql.Tx, table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleLeftJoin("", table1, table2, columns, joiners, where, orderby, limit)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleInnerJoinTx(tx *sql.Tx, table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleInnerJoin("", table1, table2, columns, joiners, where, orderby, limit)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) CreateTableTx(tx *sql.Tx, table, charset, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt) {
	res, err := acc.adapter.CreateTable("", table, charset, collation, columns, keys)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleInsertTx(tx *sql.Tx, table, columns, fields string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleInsert("", table, columns, fields)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleInsertSelectTx(tx *sql.Tx, ins DBInsert, sel DBSelect) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleInsertSelect("", ins, sel)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleInsertLeftJoin("", ins, sel)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleInsertInnerJoin("", ins, sel)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleUpdateTx(tx *sql.Tx, table, set, where string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleUpdate(qUpdate(table, set, where))
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) SimpleDeleteTx(tx *sql.Tx, table, where string) (stmt *sql.Stmt) {
	res, err := acc.adapter.SimpleDelete("", table, where)
	return acc.prepareTx(tx, res, err)
}

// I don't know why you need this, but here it is x.x
func (acc *Accumulator) PurgeTx(tx *sql.Tx, table string) (stmt *sql.Stmt) {
	res, err := acc.adapter.Purge("", table)
	return acc.prepareTx(tx, res, err)
}

func (acc *Accumulator) Delete(table string) *accDeleteBuilder {
	return &accDeleteBuilder{table, "", nil, acc}
}

func (acc *Accumulator) Update(table string) *accUpdateBuilder {
	return &accUpdateBuilder{qUpdate(table, "", ""), acc}
}

func (acc *Accumulator) Select(table string) *AccSelectBuilder {
	return &AccSelectBuilder{table, "", "", "", "", nil, nil, "", acc}
}

func (acc *Accumulator) Exists(tbl, col string) *AccSelectBuilder {
	return acc.Select(tbl).Columns(col).Where(col + "=?")
}

func (acc *Accumulator) Insert(table string) *accInsertBuilder {
	return &accInsertBuilder{table, "", "", acc}
}

func (acc *Accumulator) BulkInsert(table string) *accBulkInsertBuilder {
	return &accBulkInsertBuilder{table, "", nil, acc}
}

func (acc *Accumulator) Count(table string) *accCountBuilder {
	return &accCountBuilder{table, "", "", nil, nil, "", acc}
}

type SimpleModel struct {
	delete *sql.Stmt
	create *sql.Stmt
	update *sql.Stmt
}

func (acc *Accumulator) SimpleModel(tbl, colstr, primary string) SimpleModel {
	var qlist, uplist string
	for _, col := range strings.Split(colstr, ",") {
		qlist += "?,"
		uplist += col + "=?,"
	}
	if len(qlist) > 0 {
		qlist = qlist[0 : len(qlist)-1]
		uplist = uplist[0 : len(uplist)-1]
	}

	where := primary + "=?"
	return SimpleModel{
		delete: acc.Delete(tbl).Where(where).Prepare(),
		create: acc.Insert(tbl).Columns(colstr).Fields(qlist).Prepare(),
		update: acc.Update(tbl).Set(uplist).Where(where).Prepare(),
	}
}

func (m SimpleModel) Delete(keyVal interface{}) error {
	_, err := m.delete.Exec(keyVal)
	return err
}

func (m SimpleModel) Update(args ...interface{}) error {
	_, err := m.update.Exec(args...)
	return err
}

func (m SimpleModel) Create(args ...interface{}) error {
	_, err := m.create.Exec(args...)
	return err
}

func (m SimpleModel) CreateID(args ...interface{}) (int, error) {
	res, err := m.create.Exec(args...)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}

func (acc *Accumulator) Model(table string) *accModelBuilder {
	return &accModelBuilder{table, "", acc}
}

type accModelBuilder struct {
	table   string
	primary string

	build *Accumulator
}

func (b *accModelBuilder) Primary(col string) *accModelBuilder {
	b.primary = col
	return b
}
