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

func (builder *accDeleteBuilder) Where(where string) *accDeleteBuilder {
	if builder.where != "" {
		builder.where += " AND "
	}
	builder.where += where
	return builder
}

func (builder *accDeleteBuilder) Prepare() *sql.Stmt {
	return builder.build.SimpleDelete(builder.table, builder.where)
}

func (builder *accDeleteBuilder) Run(args ...interface{}) (int, error) {
	stmt := builder.Prepare()
	if stmt == nil {
		return 0, builder.build.FirstError()
	}

	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
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

type AccSelectBuilder struct {
	table      string
	columns    string
	where      string
	orderby    string
	limit      string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way
	inChain    *AccSelectBuilder
	inColumn   string

	build *Accumulator
}

func (selectItem *AccSelectBuilder) Columns(columns string) *AccSelectBuilder {
	selectItem.columns = columns
	return selectItem
}

func (selectItem *AccSelectBuilder) Where(where string) *AccSelectBuilder {
	if selectItem.where != "" {
		selectItem.where += " AND "
	}
	selectItem.where += where
	return selectItem
}

// TODO: Don't implement the SQL at the accumulator level but the adapter level
func (selectItem *AccSelectBuilder) In(column string, inList []int) *AccSelectBuilder {
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

func (selectItem *AccSelectBuilder) InQ(column string, subBuilder *AccSelectBuilder) *AccSelectBuilder {
	selectItem.inChain = subBuilder
	selectItem.inColumn = column
	return selectItem
}

func (selectItem *AccSelectBuilder) DateCutoff(column string, quantity int, unit string) *AccSelectBuilder {
	selectItem.dateCutoff = &dateCutoff{column, quantity, unit}
	return selectItem
}

func (selectItem *AccSelectBuilder) Orderby(orderby string) *AccSelectBuilder {
	selectItem.orderby = orderby
	return selectItem
}

func (selectItem *AccSelectBuilder) Limit(limit string) *AccSelectBuilder {
	selectItem.limit = limit
	return selectItem
}

func (selectItem *AccSelectBuilder) Prepare() *sql.Stmt {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if selectItem.dateCutoff != nil || selectItem.inChain != nil {
		selectBuilder := selectItem.build.GetAdapter().Builder().Select().FromAcc(selectItem)
		return selectItem.build.prepare(selectItem.build.GetAdapter().ComplexSelect(selectBuilder))
	}
	return selectItem.build.SimpleSelect(selectItem.table, selectItem.columns, selectItem.where, selectItem.orderby, selectItem.limit)
}

func (selectItem *AccSelectBuilder) Query(args ...interface{}) (*sql.Rows, error) {
	stmt := selectItem.Prepare()
	if stmt != nil {
		return stmt.Query(args...)
	}
	return nil, selectItem.build.FirstError()
}

type AccRowWrap struct {
	row *sql.Row
	err error
}

func (wrap *AccRowWrap) Scan(dest ...interface{}) error {
	if wrap.err != nil {
		return wrap.err
	}
	return wrap.row.Scan(dest...)
}

// TODO: Test to make sure the errors are passed up properly
func (selectItem *AccSelectBuilder) QueryRow(args ...interface{}) *AccRowWrap {
	stmt := selectItem.Prepare()
	if stmt != nil {
		return &AccRowWrap{stmt.QueryRow(args...), nil}
	}
	return &AccRowWrap{nil, selectItem.build.FirstError()}
}

// Experimental, reduces lines
func (selectItem *AccSelectBuilder) Each(handle func(*sql.Rows) error) error {
	rows, err := selectItem.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = handle(rows)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}
func (selectItem *AccSelectBuilder) EachInt(handle func(int) error) error {
	rows, err := selectItem.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var theInt int
		err = rows.Scan(&theInt)
		if err != nil {
			return err
		}
		err = handle(theInt)
		if err != nil {
			return err
		}
	}
	return rows.Err()
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

func (builder *accInsertBuilder) Run(args ...interface{}) (int, error) {
	stmt := builder.Prepare()
	if stmt == nil {
		return 0, builder.build.FirstError()
	}

	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
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
