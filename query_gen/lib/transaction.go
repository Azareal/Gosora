package qgen

import "database/sql"

type transactionStmt struct {
	stmt     *sql.Stmt
	firstErr error // This'll let us chain the methods to reduce boilerplate
}

func newTransactionStmt(stmt *sql.Stmt, err error) *transactionStmt {
	return &transactionStmt{stmt, err}
}

func (stmt *transactionStmt) Exec(args ...interface{}) (*sql.Result, error) {
	if stmt.firstErr != nil {
		return nil, stmt.firstErr
	}
	return stmt.Exec(args...)
}

type TransactionBuilder struct {
	tx         *sql.Tx
	adapter    DB_Adapter
	textToStmt map[string]*transactionStmt
}

func (build *TransactionBuilder) SimpleDelete(table string, where string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	if err != nil {
		return stmt, err
	}
	return build.tx.Prepare(res)
}

// Quick* versions refer to it being quick to type not the performance. For performance critical transactions, you might want to use the Simple* methods or the *Tx methods on the main builder. Alternate suggestions for names are welcome :)
func (build *TransactionBuilder) QuickDelete(table string, where string) *transactionStmt {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	if err != nil {
		return newTransactionStmt(nil, err)
	}

	stmt, ok := build.textToStmt[res]
	if ok {
		return stmt
	}
	stmt = newTransactionStmt(build.tx.Prepare(res))
	build.textToStmt[res] = stmt
	return stmt
}

func (build *TransactionBuilder) SimpleInsert(table string, columns string, fields string) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleInsert("_builder", table, columns, fields)
	if err != nil {
		return stmt, err
	}
	return build.tx.Prepare(res)
}

func (build *TransactionBuilder) QuickInsert(table string, where string) *transactionStmt {
	res, err := build.adapter.SimpleDelete("_builder", table, where)
	if err != nil {
		return newTransactionStmt(nil, err)
	}

	stmt, ok := build.textToStmt[res]
	if ok {
		return stmt
	}
	stmt = newTransactionStmt(build.tx.Prepare(res))
	build.textToStmt[res] = stmt
	return stmt
}
