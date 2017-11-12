/* WIP: A version of the builder which accumulates errors, we'll see if we can't unify the implementations at some point */
package qgen

import (
	"database/sql"
)

type Accumulator struct {
	conn     *sql.DB
	adapter  DB_Adapter
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

func (build *Accumulator) GetAdapter() DB_Adapter {
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

func (build *Accumulator) CreateTable(table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) *sql.Stmt {
	return build.prepare(build.adapter.CreateTable("_builder", table, charset, collation, columns, keys))
}

func (build *Accumulator) SimpleInsert(table string, columns string, fields string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsert("_builder", table, columns, fields))
}

func (build *Accumulator) SimpleInsertSelect(ins DB_Insert, sel DB_Select) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertSelect("_builder", ins, sel))
}

func (build *Accumulator) SimpleInsertLeftJoin(ins DB_Insert, sel DB_Join) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertLeftJoin("_builder", ins, sel))
}

func (build *Accumulator) SimpleInsertInnerJoin(ins DB_Insert, sel DB_Join) *sql.Stmt {
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

func (build *Accumulator) CreateTableTx(tx *sql.Tx, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (stmt *sql.Stmt) {
	res, err := build.adapter.CreateTable("_builder", table, charset, collation, columns, keys)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertTx(tx *sql.Tx, table string, columns string, fields string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertSelectTx(tx *sql.Tx, ins DB_Insert, sel DB_Select) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertSelect("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DB_Insert, sel DB_Join) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertLeftJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *Accumulator) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DB_Insert, sel DB_Join) (stmt *sql.Stmt) {
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

func (build *Accumulator) Delete(table string) *deleteBuilder {
	return &deleteBuilder{table, "", build}
}

type deleteBuilder struct {
	table string
	where string

	build *Accumulator
}

func (delete *deleteBuilder) Where(where string) *deleteBuilder {
	delete.where = where
	return delete
}

func (delete *deleteBuilder) Prepare() *sql.Stmt {
	return delete.build.SimpleDelete(delete.table, delete.where)
}

func (build *Accumulator) Update(table string) *updateBuilder {
	return &updateBuilder{table, "", "", build}
}

type updateBuilder struct {
	table string
	set   string
	where string

	build *Accumulator
}

func (update *updateBuilder) Set(set string) *updateBuilder {
	update.set = set
	return update
}

func (update *updateBuilder) Where(where string) *updateBuilder {
	update.where = where
	return update
}

func (update *updateBuilder) Prepare() *sql.Stmt {
	return update.build.SimpleUpdate(update.table, update.set, update.where)
}

func (build *Accumulator) Select(table string) *selectBuilder {
	return &selectBuilder{table, "", "", "", "", build}
}

type selectBuilder struct {
	table   string
	columns string
	where   string
	orderby string
	limit   string

	build *Accumulator
}

func (selectItem *selectBuilder) Columns(columns string) *selectBuilder {
	selectItem.columns = columns
	return selectItem
}

func (selectItem *selectBuilder) Where(where string) *selectBuilder {
	selectItem.where = where
	return selectItem
}

func (selectItem *selectBuilder) Orderby(orderby string) *selectBuilder {
	selectItem.orderby = orderby
	return selectItem
}

func (selectItem *selectBuilder) Limit(limit string) *selectBuilder {
	selectItem.limit = limit
	return selectItem
}

func (selectItem *selectBuilder) Prepare() *sql.Stmt {
	return selectItem.build.SimpleSelect(selectItem.table, selectItem.columns, selectItem.where, selectItem.orderby, selectItem.limit)
}

func (selectItem *selectBuilder) Query(args ...interface{}) (*sql.Rows, error) {
	stmt := selectItem.Prepare()
	if stmt != nil {
		return stmt.Query(args...)
	}
	return nil, selectItem.build.FirstError()
}

func (build *Accumulator) Insert(table string) *insertBuilder {
	return &insertBuilder{table, "", "", build}
}

type insertBuilder struct {
	table   string
	columns string
	fields  string

	build *Accumulator
}

func (insert *insertBuilder) Columns(columns string) *insertBuilder {
	insert.columns = columns
	return insert
}

func (insert *insertBuilder) Fields(fields string) *insertBuilder {
	insert.fields = fields
	return insert
}

func (insert *insertBuilder) Prepare() *sql.Stmt {
	return insert.build.SimpleInsert(insert.table, insert.columns, insert.fields)
}

func (build *Accumulator) Count(table string) *countBuilder {
	return &countBuilder{table, "", "", build}
}

type countBuilder struct {
	table string
	where string
	limit string

	build *Accumulator
}

func (count *countBuilder) Where(where string) *countBuilder {
	count.where = where
	return count
}

func (count *countBuilder) Limit(limit string) *countBuilder {
	count.limit = limit
	return count
}

func (count *countBuilder) Prepare() *sql.Stmt {
	return count.build.SimpleCount(count.table, count.where, count.limit)
}
