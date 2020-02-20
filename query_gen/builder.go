/* WIP Under Construction */
package qgen

import (
	"database/sql"
	"log"
)

var Builder *builder

func init() {
	Builder = &builder{conn: nil}
}

// A set of wrappers around the generator methods, so that we can use this inline in Gosora
type builder struct {
	conn    *sql.DB
	adapter Adapter
}

func (b *builder) Accumulator() *Accumulator {
	return &Accumulator{b.conn, b.adapter, nil}
}

// TODO: Move this method out of builder?
func (b *builder) Init(adapter string, config map[string]string) error {
	err := b.SetAdapter(adapter)
	if err != nil {
		return err
	}
	conn, err := b.adapter.BuildConn(config)
	b.conn = conn
	log.Print("err:", err) // Is the problem here somehow?
	return err
}

func (b *builder) SetConn(conn *sql.DB) {
	b.conn = conn
}

func (b *builder) GetConn() *sql.DB {
	return b.conn
}

func (b *builder) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	b.adapter = adap
	return nil
}

func (b *builder) GetAdapter() Adapter {
	return b.adapter
}

func (b *builder) DbVersion() (dbVersion string) {
	b.conn.QueryRow(b.adapter.DbVersion()).Scan(&dbVersion)
	return dbVersion
}

func (b *builder) Begin() (*sql.Tx, error) {
	return b.conn.Begin()
}

func (b *builder) Tx(h func(*TransactionBuilder) error) error {
	tx, err := b.conn.Begin()
	if err != nil {
		return err
	}
	err = h(&TransactionBuilder{tx, b.adapter, nil})
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (b *builder) prepare(res string, err error) (*sql.Stmt, error) {
	if err != nil {
		return nil, err
	}
	return b.conn.Prepare(res)
}

func (b *builder) SimpleSelect(table, columns, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleSelect("", table, columns, where, orderby, limit))
}

func (b *builder) SimpleCount(table, where, limit string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleCount("", table, where, limit))
}

func (b *builder) SimpleLeftJoin(table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleLeftJoin("", table1, table2, columns, joiners, where, orderby, limit))
}

func (b *builder) SimpleInnerJoin(table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleInnerJoin("", table1, table2, columns, joiners, where, orderby, limit))
}

func (b *builder) DropTable(table string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.DropTable("", table))
}

func (build *builder) CreateTable(table, charset, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt, err error) {
	return build.prepare(build.adapter.CreateTable("", table, charset, collation, columns, keys))
}

func (b *builder) AddColumn(table string, column DBTableColumn, key *DBTableKey) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.AddColumn("", table, column, key))
}

func (b *builder) DropColumn(table, colName string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.DropColumn("", table, colName))
}

func (b *builder) RenameColumn(table, oldName, newName string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.RenameColumn("", table, oldName, newName))
}

func (b *builder) ChangeColumn(table, colName string, col DBTableColumn) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.ChangeColumn("", table, colName, col))
}

func (b *builder) SetDefaultColumn(table, colName, colType, defaultStr string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SetDefaultColumn("", table, colName, colType, defaultStr))
}

func (b *builder) AddIndex(table, iname, colname string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.AddIndex("", table, iname, colname))
}

func (b *builder) AddKey(table, column string, key DBTableKey) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.AddKey("", table, column, key))
}

func (b *builder) RemoveIndex(table, iname string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.RemoveIndex("", table, iname))
}

func (b *builder) AddForeignKey(table, column, ftable, fcolumn string, cascade bool) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.AddForeignKey("", table, column, ftable, fcolumn, cascade))
}

func (b *builder) SimpleInsert(table, columns, fields string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleInsert("", table, columns, fields))
}

func (b *builder) SimpleInsertSelect(ins DBInsert, sel DBSelect) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleInsertSelect("", ins, sel))
}

func (b *builder) SimpleInsertLeftJoin(ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleInsertLeftJoin("", ins, sel))
}

func (b *builder) SimpleInsertInnerJoin(ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleInsertInnerJoin("", ins, sel))
}

func (b *builder) SimpleUpdate(table, set, where string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleUpdate(qUpdate(table, set, where)))
}

func (b *builder) SimpleDelete(table, where string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.SimpleDelete("", table, where))
}

// I don't know why you need this, but here it is x.x
func (b *builder) Purge(table string) (stmt *sql.Stmt, err error) {
	return b.prepare(b.adapter.Purge("", table))
}

func (b *builder) prepareTx(tx *sql.Tx, res string, err error) (*sql.Stmt, error) {
	if err != nil {
		return nil, err
	}
	return tx.Prepare(res)
}

// These ones support transactions
func (b *builder) SimpleSelectTx(tx *sql.Tx, table, columns, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleSelect("", table, columns, where, orderby, limit)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleCountTx(tx *sql.Tx, table, where, limit string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleCount("", table, where, limit)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleLeftJoinTx(tx *sql.Tx, table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleLeftJoin("", table1, table2, columns, joiners, where, orderby, limit)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleInnerJoinTx(tx *sql.Tx, table1, table2, columns, joiners, where, orderby, limit string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInnerJoin("", table1, table2, columns, joiners, where, orderby, limit)
	return b.prepareTx(tx, res, err)
}

func (b *builder) CreateTableTx(tx *sql.Tx, table, charset, collation string, columns []DBTableColumn, keys []DBTableKey) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.CreateTable("", table, charset, collation, columns, keys)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleInsertTx(tx *sql.Tx, table, columns, fields string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInsert("", table, columns, fields)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleInsertSelectTx(tx *sql.Tx, ins DBInsert, sel DBSelect) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInsertSelect("", ins, sel)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleInsertLeftJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInsertLeftJoin("", ins, sel)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleInsertInnerJoinTx(tx *sql.Tx, ins DBInsert, sel DBJoin) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInsertInnerJoin("", ins, sel)
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleUpdateTx(tx *sql.Tx, table, set, where string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleUpdate(qUpdate(table, set, where))
	return b.prepareTx(tx, res, err)
}

func (b *builder) SimpleDeleteTx(tx *sql.Tx, table, where string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleDelete("", table, where)
	return b.prepareTx(tx, res, err)
}

// I don't know why you need this, but here it is x.x
func (b *builder) PurgeTx(tx *sql.Tx, table string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.Purge("", table)
	return b.prepareTx(tx, res, err)
}
