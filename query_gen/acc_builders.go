package qgen

import (
	"database/sql"
	"strconv"
)

type accDeleteBuilder struct {
	table string
	where string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way

	build *Accumulator
}

func (b *accDeleteBuilder) Where(where string) *accDeleteBuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

func (b *accDeleteBuilder) DateCutoff(column string, quantity int, unit string) *accDeleteBuilder {
	b.dateCutoff = &dateCutoff{column, quantity, unit, 0}
	return b
}

func (b *accDeleteBuilder) DateOlderThan(column string, quantity int, unit string) *accDeleteBuilder {
	b.dateCutoff = &dateCutoff{column, quantity, unit, 1}
	return b
}

/*func (b *accDeleteBuilder) Prepare() *sql.Stmt {
	return b.build.SimpleDelete(b.table, b.where)
}*/

// TODO: Fix this nasty hack
func (b *accDeleteBuilder) Prepare() *sql.Stmt {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if b.dateCutoff != nil {
		dBuilder := b.build.GetAdapter().Builder().Delete().FromAcc(b)
		return b.build.prepare(b.build.GetAdapter().ComplexDelete(dBuilder))
	}
	return b.build.SimpleDelete(b.table, b.where)
}

func (b *accDeleteBuilder) Run(args ...interface{}) (int, error) {
	stmt := b.Prepare()
	if stmt == nil {
		return 0, b.build.FirstError()
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

func (u *accUpdateBuilder) Set(set string) *accUpdateBuilder {
	u.up.set = set
	return u
}

func (u *accUpdateBuilder) Where(where string) *accUpdateBuilder {
	if u.up.where != "" {
		u.up.where += " AND "
	}
	u.up.where += where
	return u
}

func (b *accUpdateBuilder) DateCutoff(column string, quantity int, unit string) *accUpdateBuilder {
	b.up.dateCutoff = &dateCutoff{column, quantity, unit, 0}
	return b
}

func (b *accUpdateBuilder) DateOlderThan(column string, quantity int, unit string) *accUpdateBuilder {
	b.up.dateCutoff = &dateCutoff{column, quantity, unit, 1}
	return b
}

func (b *accUpdateBuilder) WhereQ(sel *selectPrebuilder) *accUpdateBuilder {
	b.up.whereSubQuery = sel
	return b
}

func (b *accUpdateBuilder) Prepare() *sql.Stmt {
	if b.up.whereSubQuery != nil {
		return b.build.prepare(b.build.adapter.SimpleUpdateSelect(b.up))
	}
	return b.build.prepare(b.build.adapter.SimpleUpdate(b.up))
}

func (b *accUpdateBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	query, err := b.build.adapter.SimpleUpdate(b.up)
	if err != nil {
		return res, err
	}
	return b.build.exec(query, args...)
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

func (b *AccSelectBuilder) Columns(columns string) *AccSelectBuilder {
	b.columns = columns
	return b
}

func (b *AccSelectBuilder) Cols(columns string) *AccSelectBuilder {
	b.columns = columns
	return b
}

func (b *AccSelectBuilder) Where(where string) *AccSelectBuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

// TODO: Don't implement the SQL at the accumulator level but the adapter level
func (b *AccSelectBuilder) In(column string, inList []int) *AccSelectBuilder {
	if len(inList) == 0 {
		return b
	}

	// TODO: Optimise this
	where := column + " IN("
	for _, item := range inList {
		where += strconv.Itoa(item) + ","
	}
	where = where[:len(where)-1] + ")"
	if b.where != "" {
		where += " AND " + b.where
	}

	b.where = where
	return b
}

// TODO: Don't implement the SQL at the accumulator level but the adapter level
func (b *AccSelectBuilder) InPQuery(column string, inList []int) (*sql.Rows, error) {
	if len(inList) == 0 {
		return nil, sql.ErrNoRows
	}
	// TODO: Optimise this
	where := column + " IN("

	idList := make([]interface{},len(inList))
	for i, id := range inList {
		idList[i] = strconv.Itoa(id)
		where += "?,"
	}
	where = where[0 : len(where)-1] + ")"

	if b.where != "" {
		where += " AND " + b.where
	}

	b.where = where
	return b.Query(idList...)
}

func (b *AccSelectBuilder) InQ(column string, subBuilder *AccSelectBuilder) *AccSelectBuilder {
	b.inChain = subBuilder
	b.inColumn = column
	return b
}

func (b *AccSelectBuilder) DateCutoff(column string, quantity int, unit string) *AccSelectBuilder {
	b.dateCutoff = &dateCutoff{column, quantity, unit, 0}
	return b
}

func (b *AccSelectBuilder) Orderby(orderby string) *AccSelectBuilder {
	b.orderby = orderby
	return b
}

func (b *AccSelectBuilder) Limit(limit string) *AccSelectBuilder {
	b.limit = limit
	return b
}

func (b *AccSelectBuilder) Prepare() *sql.Stmt {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if b.dateCutoff != nil || b.inChain != nil {
		selectBuilder := b.build.GetAdapter().Builder().Select().FromAcc(b)
		return b.build.prepare(b.build.GetAdapter().ComplexSelect(selectBuilder))
	}
	return b.build.SimpleSelect(b.table, b.columns, b.where, b.orderby, b.limit)
}

func (b *AccSelectBuilder) query() (string, error) {
	// TODO: Phase out the procedural API and use the adapter's OO API? The OO API might need a bit more work before we do that and it needs to be rolled out to MSSQL.
	if b.dateCutoff != nil || b.inChain != nil {
		selectBuilder := b.build.GetAdapter().Builder().Select().FromAcc(b)
		return b.build.GetAdapter().ComplexSelect(selectBuilder)
	}
	return b.build.adapter.SimpleSelect("", b.table, b.columns, b.where, b.orderby, b.limit)
}

func (b *AccSelectBuilder) Query(args ...interface{}) (*sql.Rows, error) {
	stmt := b.Prepare()
	if stmt != nil {
		return stmt.Query(args...)
	}
	return nil, b.build.FirstError()
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
func (b *AccSelectBuilder) QueryRow(args ...interface{}) *AccRowWrap {
	stmt := b.Prepare()
	if stmt != nil {
		return &AccRowWrap{stmt.QueryRow(args...), nil}
	}
	return &AccRowWrap{nil, b.build.FirstError()}
}

// Experimental, reduces lines
func (b *AccSelectBuilder) Each(handle func(*sql.Rows) error) error {
	query, err := b.query()
	if err != nil {
		return err
	}
	rows, err := b.build.query(query)
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
func (b *AccSelectBuilder) EachInt(handle func(int) error) error {
	query, err := b.query()
	if err != nil {
		return err
	}
	rows, err := b.build.query(query)
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

func (b *accInsertBuilder) Columns(columns string) *accInsertBuilder {
	b.columns = columns
	return b
}

func (b *accInsertBuilder) Fields(fields string) *accInsertBuilder {
	b.fields = fields
	return b
}

func (b *accInsertBuilder) Prepare() *sql.Stmt {
	return b.build.SimpleInsert(b.table, b.columns, b.fields)
}

func (b *accInsertBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	query, err := b.build.adapter.SimpleInsert("", b.table, b.columns, b.fields)
	if err != nil {
		return res, err
	}
	return b.build.exec(query, args...)
}

func (b *accInsertBuilder) Run(args ...interface{}) (int, error) {
	query, err := b.build.adapter.SimpleInsert("", b.table, b.columns, b.fields)
	if err != nil {
		return 0, err
	}
	res, err := b.build.exec(query, args...)
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
	b.dateCutoff = &dateCutoff{column, quantity, unit, 0}
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
