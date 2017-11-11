/* WIP: A version of the builder which accumulates errors, we'll see if we can't unify the implementations at some point */
package qgen

import (
	"database/sql"
)

type accBuilder struct {
	conn     *sql.DB
	adapter  DB_Adapter
	firstErr error
}

func (build *accBuilder) SetConn(conn *sql.DB) {
	build.conn = conn
}

func (build *accBuilder) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	build.adapter = adap
	return nil
}

func (build *accBuilder) GetAdapter() DB_Adapter {
	return build.adapter
}

func (build *accBuilder) FirstError() error {
	return build.firstErr
}

func (build *accBuilder) recordError(err error) {
	if err == nil {
		return
	}
	if build.firstErr == nil {
		build.firstErr = err
	}
}

func (build *accBuilder) prepare(res string, err error) *sql.Stmt {
	if err != nil {
		build.recordError(err)
		return nil
	}
	stmt, err := build.conn.Prepare(res)
	build.recordError(err)
	return stmt
}

func (build *accBuilder) Tx(handler func(*TransactionBuilder) error) {
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

func (build *accBuilder) SimpleSelect(table string, columns string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit))
}

func (build *accBuilder) SimpleCount(table string, where string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleCount("_builder", table, where, limit))
}

func (build *accBuilder) SimpleLeftJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *accBuilder) SimpleInnerJoin(table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit))
}

func (build *accBuilder) CreateTable(table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) *sql.Stmt {
	return build.prepare(build.adapter.CreateTable("_builder", table, charset, collation, columns, keys))
}

func (build *accBuilder) SimpleInsert(table string, columns string, fields string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsert("_builder", table, columns, fields))
}

func (build *accBuilder) SimpleInsertSelect(ins DB_Insert, sel DB_Select) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertSelect("_builder", ins, sel))
}

func (build *accBuilder) SimpleInsertLeftJoin(ins DB_Insert, sel DB_Join) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertLeftJoin("_builder", ins, sel))
}

func (build *accBuilder) SimpleInsertInnerJoin(ins DB_Insert, sel DB_Join) *sql.Stmt {
	return build.prepare(build.adapter.SimpleInsertInnerJoin("_builder", ins, sel))
}

func (build *accBuilder) SimpleUpdate(table string, set string, where string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleUpdate("_builder", table, set, where))
}

func (build *accBuilder) SimpleDelete(table string, where string) *sql.Stmt {
	return build.prepare(build.adapter.SimpleDelete("_builder", table, where))
}

// I don't know why you need this, but here it is x.x
func (build *accBuilder) Purge(table string) *sql.Stmt {
	return build.prepare(build.adapter.Purge("_builder", table))
}

func (build *accBuilder) prepareTx(tx *sql.Tx, res string, err error) (stmt *sql.Stmt) {
	if err != nil {
		build.recordError(err)
		return nil
	}
	stmt, err = tx.Prepare(res)
	build.recordError(err)
	return stmt
}

// These ones support transactions
func (build *accBuilder) SimpleSelectTx(tx *sql.Tx, table string, columns string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleSelect("_builder", table, columns, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleCountTx(tx *sql.Tx, table string, where string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleCount("_builder", table, where, limit)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleLeftJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleLeftJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleInnerJoinTx(tx *sql.Tx, table1 string, table2 string, columns string, joiners string, where string, orderby string, limit string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInnerJoin("_builder", table1, table2, columns, joiners, where, orderby, limit)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) CreateTableTx(tx *sql.Tx, table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) (stmt *sql.Stmt) {
	res, err := build.adapter.CreateTable("_builder", table, charset, collation, columns, keys)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleInsertTx(tx *sql.Tx, table string, columns string, fields string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleInsertSelectTx(tx *sql.Tx, ins DB_Insert, sel DB_Select) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertSelect("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DB_Insert, sel DB_Join) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertLeftJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DB_Insert, sel DB_Join) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleInsertInnerJoin("_builder", ins, sel)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleUpdateTx(tx *sql.Tx, table string, set string, where string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleUpdate("_builder", table, set, where)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) SimpleDeleteTx(tx *sql.Tx, table string, where string) (stmt *sql.Stmt) {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	return build.prepareTx(tx, res, err)
}

// I don't know why you need this, but here it is x.x
func (build *accBuilder) PurgeTx(tx *sql.Tx, table string) (stmt *sql.Stmt) {
	res, err := build.adapter.Purge("_builder", table)
	return build.prepareTx(tx, res, err)
}

func (build *accBuilder) Delete(table string) *deleteBuilder {
	return &deleteBuilder{table, "", build}
}

type deleteBuilder struct {
	table string
	where string

	build *accBuilder
}

func (delete *deleteBuilder) Where(where string) *deleteBuilder {
	delete.where = where
	return delete
}

func (delete *deleteBuilder) Prepare() *sql.Stmt {
	return delete.build.SimpleDelete(delete.table, delete.where)
}

func (build *accBuilder) Update(table string) *updateBuilder {
	return &updateBuilder{table, "", "", build}
}

type updateBuilder struct {
	table string
	set   string
	where string

	build *accBuilder
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

func (build *accBuilder) Select(table string) *selectBuilder {
	return &selectBuilder{table, "", "", "", "", build}
}

type selectBuilder struct {
	table   string
	columns string
	where   string
	orderby string
	limit   string

	build *accBuilder
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

func (build *accBuilder) Insert(table string) *insertBuilder {
	return &insertBuilder{table, "", "", build}
}

type insertBuilder struct {
	table   string
	columns string
	fields  string

	build *accBuilder
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
