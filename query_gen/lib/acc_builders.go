package qgen

import "database/sql"

type accDeleteBuilder struct {
	table string
	where string

	build *Accumulator
}

func (delete *accDeleteBuilder) Where(where string) *accDeleteBuilder {
	delete.where = where
	return delete
}

func (delete *accDeleteBuilder) Prepare() *sql.Stmt {
	return delete.build.SimpleDelete(delete.table, delete.where)
}

type accUpdateBuilder struct {
	table string
	set   string
	where string

	build *Accumulator
}

func (update *accUpdateBuilder) Set(set string) *accUpdateBuilder {
	update.set = set
	return update
}

func (update *accUpdateBuilder) Where(where string) *accUpdateBuilder {
	update.where = where
	return update
}

func (update *accUpdateBuilder) Prepare() *sql.Stmt {
	return update.build.SimpleUpdate(update.table, update.set, update.where)
}

type accSelectBuilder struct {
	table   string
	columns string
	where   string
	orderby string
	limit   string

	build *Accumulator
}

func (selectItem *accSelectBuilder) Columns(columns string) *accSelectBuilder {
	selectItem.columns = columns
	return selectItem
}

func (selectItem *accSelectBuilder) Where(where string) *accSelectBuilder {
	selectItem.where = where
	return selectItem
}

func (selectItem *accSelectBuilder) Orderby(orderby string) *accSelectBuilder {
	selectItem.orderby = orderby
	return selectItem
}

func (selectItem *accSelectBuilder) Limit(limit string) *accSelectBuilder {
	selectItem.limit = limit
	return selectItem
}

func (selectItem *accSelectBuilder) Prepare() *sql.Stmt {
	return selectItem.build.SimpleSelect(selectItem.table, selectItem.columns, selectItem.where, selectItem.orderby, selectItem.limit)
}

func (selectItem *accSelectBuilder) Query(args ...interface{}) (*sql.Rows, error) {
	stmt := selectItem.Prepare()
	if stmt != nil {
		return stmt.Query(args...)
	}
	return nil, selectItem.build.FirstError()
}

type accInsertBuilder struct {
	table   string
	columns string
	fields  string

	build *Accumulator
}

func (insert *accInsertBuilder) Columns(columns string) *accInsertBuilder {
	insert.columns = columns
	return insert
}

func (insert *accInsertBuilder) Fields(fields string) *accInsertBuilder {
	insert.fields = fields
	return insert
}

func (insert *accInsertBuilder) Prepare() *sql.Stmt {
	return insert.build.SimpleInsert(insert.table, insert.columns, insert.fields)
}

type accCountBuilder struct {
	table string
	where string
	limit string

	build *Accumulator
}

func (count *accCountBuilder) Where(where string) *accCountBuilder {
	count.where = where
	return count
}

func (count *accCountBuilder) Limit(limit string) *accCountBuilder {
	count.limit = limit
	return count
}

func (count *accCountBuilder) Prepare() *sql.Stmt {
	return count.build.SimpleCount(count.table, count.where, count.limit)
}
