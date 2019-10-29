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
	adapter    Adapter
	textToStmt map[string]*transactionStmt
}

func (b *TransactionBuilder) SimpleDelete(table string, where string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleDelete("", table, where)
	if err != nil {
		return stmt, err
	}
	return b.tx.Prepare(res)
}

// Quick* versions refer to it being quick to type not the performance. For performance critical transactions, you might want to use the Simple* methods or the *Tx methods on the main builder. Alternate suggestions for names are welcome :)
func (b *TransactionBuilder) QuickDelete(table string, where string) *transactionStmt {
	res, err := b.adapter.SimpleDelete("", table, where)
	if err != nil {
		return newTransactionStmt(nil, err)
	}

	stmt, ok := b.textToStmt[res]
	if ok {
		return stmt
	}
	stmt = newTransactionStmt(b.tx.Prepare(res))
	b.textToStmt[res] = stmt
	return stmt
}

func (b *TransactionBuilder) SimpleInsert(table string, columns string, fields string) (stmt *sql.Stmt, err error) {
	res, err := b.adapter.SimpleInsert("", table, columns, fields)
	if err != nil {
		return stmt, err
	}
	return b.tx.Prepare(res)
}

func (b *TransactionBuilder) QuickInsert(table string, where string) *transactionStmt {
	res, err := b.adapter.SimpleDelete("", table, where)
	if err != nil {
		return newTransactionStmt(nil, err)
	}

	stmt, ok := b.textToStmt[res]
	if ok {
		return stmt
	}
	stmt = newTransactionStmt(b.tx.Prepare(res))
	b.textToStmt[res] = stmt
	return stmt
}
