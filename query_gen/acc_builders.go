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
	up    *updatePrebuilder
	build *Accumulator
}

func (update *accUpdateBuilder) Set(set string) *accUpdateBuilder {
	update.up.set = set
	return update
}

func (update *accUpdateBuilder) Where(where string) *accUpdateBuilder {
	if update.up.where != "" {
		update.up.where += " AND "
	}
	update.up.where += where
	return update
}

func (update *accUpdateBuilder) WhereQ(sel *selectPrebuilder) *accUpdateBuilder {
	update.up.whereSubQuery = sel
	return update
}

func (builder *accUpdateBuilder) Prepare() *sql.Stmt {
	if builder.up.whereSubQuery != nil {
		return builder.build.prepare(builder.build.adapter.SimpleUpdateSelect(builder.up))
	}
	return builder.build.prepare(builder.build.adapter.SimpleUpdate(builder.up))
}

func (builder *accUpdateBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	query, err := builder.build.adapter.SimpleUpdate(builder.up)
	if err != nil {
		return res, err
	}
	return builder.build.exec(query, args...)
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

func (builder *AccSelectBuilder) Columns(columns string) *AccSelectBuilder {
	builder.columns = columns
	return builder
}

func (builder *AccSelectBuilder) Cols(columns string) *AccSelectBuilder {
	builder.columns = columns
	return builder
}

func (builder *AccSelectBuilder) Where(where string) *AccSelectBuilder {
	if builder.where != "" {
		builder.where += " AND "
	}
	builder.where += where
	return builder
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

func (builder *AccSelectBuilder) DateCutoff(column string, quantity int, unit string) *AccSelectBuilder {
	builder.dateCutoff = &dateCutoff{column, quantity, unit}
	return builder
}

func (builder *AccSelectBuilder) Orderby(orderby string) *AccSelectBuilder {
	builder.orderby = orderby
	return builder
}

func (builder *AccSelectBuilder) Limit(limit string) *AccSelectBuilder {
	builder.limit = limit
	return builder
}

func (builder *AccSelectBuilder) Prepare() *sql.Stmt {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if builder.dateCutoff != nil || builder.inChain != nil {
		selectBuilder := builder.build.GetAdapter().Builder().Select().FromAcc(builder)
		return builder.build.prepare(builder.build.GetAdapter().ComplexSelect(selectBuilder))
	}
	return builder.build.SimpleSelect(builder.table, builder.columns, builder.where, builder.orderby, builder.limit)
}

func (builder *AccSelectBuilder) query() (string, error) {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if builder.dateCutoff != nil || builder.inChain != nil {
		selectBuilder := builder.build.GetAdapter().Builder().Select().FromAcc(builder)
		return builder.build.GetAdapter().ComplexSelect(selectBuilder)
	}
	return builder.build.adapter.SimpleSelect("", builder.table, builder.columns, builder.where, builder.orderby, builder.limit)
}

func (builder *AccSelectBuilder) Query(args ...interface{}) (*sql.Rows, error) {
	stmt := builder.Prepare()
	if stmt != nil {
		return stmt.Query(args...)
	}
	return nil, builder.build.FirstError()
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
func (builder *AccSelectBuilder) QueryRow(args ...interface{}) *AccRowWrap {
	stmt := builder.Prepare()
	if stmt != nil {
		return &AccRowWrap{stmt.QueryRow(args...), nil}
	}
	return &AccRowWrap{nil, builder.build.FirstError()}
}

// Experimental, reduces lines
func (builder *AccSelectBuilder) Each(handle func(*sql.Rows) error) error {
	query, err := builder.query()
	if err != nil {
		return err
	}
	rows, err := builder.build.query(query)
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
func (builder *AccSelectBuilder) EachInt(handle func(int) error) error {
	query, err := builder.query()
	if err != nil {
		return err
	}
	rows, err := builder.build.query(query)
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

func (builder *accInsertBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	query, err := builder.build.adapter.SimpleInsert("", builder.table, builder.columns, builder.fields)
	if err != nil {
		return res, err
	}
	return builder.build.exec(query, args...)
}

func (builder *accInsertBuilder) Run(args ...interface{}) (int, error) {
	query, err := builder.build.adapter.SimpleInsert("", builder.table, builder.columns, builder.fields)
	if err != nil {
		return 0, err
	}
	res, err := builder.build.exec(query, args...)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	return int(lastID), err
}

type accCountBuilder struct {
	table      string
	where      string
	limit      string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way
	inChain    *AccSelectBuilder
	inColumn   string

	build *Accumulator
}

func (b *accCountBuilder) Where(where string) *accCountBuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

func (b *accCountBuilder) Limit(limit string) *accCountBuilder {
	b.limit = limit
	return b
}

func (b *accCountBuilder) DateCutoff(column string, quantity int, unit string) *accCountBuilder {
	b.dateCutoff = &dateCutoff{column, quantity, unit}
	return b
}

// TODO: Fix this nasty hack
func (b *accCountBuilder) Prepare() *sql.Stmt {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if b.dateCutoff != nil || b.inChain != nil {
		selBuilder := b.build.GetAdapter().Builder().Count().FromCountAcc(b)
		selBuilder.columns = "COUNT(*)"
		return b.build.prepare(b.build.GetAdapter().ComplexSelect(selBuilder))
	}
	return b.build.SimpleCount(b.table, b.where, b.limit)
}

func (b *accCountBuilder) Total() (total int, err error) {
	stmt := b.Prepare()
	if stmt == nil {
		return 0, b.build.FirstError()
	}
	err = stmt.QueryRow().Scan(&total)
	return total, err
}

// TODO: Add a Sum builder for summing viewchunks up into one number for the dashboard?
