package qgen

import (
	"database/sql"
	"strconv"
)

type accDeleteBuilder struct {
	table string
	where string

	build *Accumulator
}

func (delete *accDeleteBuilder) Where(where string) *accDeleteBuilder {
	if delete.where != "" {
		delete.where += " AND "
	}
	delete.where += where
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
	if update.where != "" {
		update.where += " AND "
	}
	update.where += where
	return update
}

func (update *accUpdateBuilder) Prepare() *sql.Stmt {
	return update.build.SimpleUpdate(update.table, update.set, update.where)
}

type accSelectBuilder struct {
	table      string
	columns    string
	where      string
	orderby    string
	limit      string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way
	inChain    *accSelectBuilder
	inColumn   string

	build *Accumulator
}

func (selectItem *accSelectBuilder) Columns(columns string) *accSelectBuilder {
	selectItem.columns = columns
	return selectItem
}

func (selectItem *accSelectBuilder) Where(where string) *accSelectBuilder {
	if selectItem.where != "" {
		selectItem.where += " AND "
	}
	selectItem.where += where
	return selectItem
}

// TODO: Don't implement the SQL at the accumulator level but the adapter level
func (selectItem *accSelectBuilder) In(column string, inList []int) *accSelectBuilder {
	if len(inList) == 0 {
		return selectItem
	}

	var where = column + " IN("
	for _, item := range inList {
		where += strconv.Itoa(item) + ","
	}
	where = where[:len(where)-1] + ")"
	if selectItem.where != "" {
		where += " AND " + selectItem.where
	}

	selectItem.where = where
	return selectItem
}

func (selectItem *accSelectBuilder) InQ(column string, subBuilder *accSelectBuilder) *accSelectBuilder {
	selectItem.inChain = subBuilder
	selectItem.inColumn = column
	return selectItem
}

func (selectItem *accSelectBuilder) DateCutoff(column string, quantity int, unit string) *accSelectBuilder {
	selectItem.dateCutoff = &dateCutoff{column, quantity, unit}
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
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if selectItem.dateCutoff != nil || selectItem.inChain != nil {
		selectBuilder := selectItem.build.GetAdapter().Builder().Select().FromAcc(selectItem)
		return selectItem.build.prepare(selectItem.build.GetAdapter().ComplexSelect(selectBuilder))
	}
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

func (insert *accInsertBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	stmt := insert.Prepare()
	if stmt != nil {
		return stmt.Exec(args...)
	}
	return res, insert.build.FirstError()
}

type accCountBuilder struct {
	table string
	where string
	limit string

	build *Accumulator
}

func (count *accCountBuilder) Where(where string) *accCountBuilder {
	if count.where != "" {
		count.where += " AND "
	}
	count.where += where
	return count
}

func (count *accCountBuilder) Limit(limit string) *accCountBuilder {
	count.limit = limit
	return count
}

// TODO: Add QueryRow for this and use it in statistics.go
func (count *accCountBuilder) Prepare() *sql.Stmt {
	return count.build.SimpleCount(count.table, count.where, count.limit)
}

// TODO: Add a Sum builder for summing viewchunks up into one number for the dashboard?
