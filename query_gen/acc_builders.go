package qgen

import (
	"database/sql"
	//"fmt"
	"strconv"
)

type accDeleteBuilder struct {
	table      string
	where      string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way

	build *Accumulator
}

func (b *accDeleteBuilder) Where(w string) *accDeleteBuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += w
	return b
}

func (b *accDeleteBuilder) DateCutoff(col string, quantity int, unit string) *accDeleteBuilder {
	b.dateCutoff = &dateCutoff{col, quantity, unit, 0}
	return b
}

func (b *accDeleteBuilder) DateOlderThan(col string, quantity int, unit string) *accDeleteBuilder {
	b.dateCutoff = &dateCutoff{col, quantity, unit, 1}
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

func (b *accUpdateBuilder) DateCutoff(col string, quantity int, unit string) *accUpdateBuilder {
	b.up.dateCutoff = &dateCutoff{col, quantity, unit, 0}
	return b
}

func (b *accUpdateBuilder) DateOlderThan(col string, quantity int, unit string) *accUpdateBuilder {
	b.up.dateCutoff = &dateCutoff{col, quantity, unit, 1}
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
	//fmt.Println("query:", query)
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

func (b *AccSelectBuilder) Columns(cols string) *AccSelectBuilder {
	b.columns = cols
	return b
}

func (b *AccSelectBuilder) Cols(cols string) *AccSelectBuilder {
	b.columns = cols
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
func (b *AccSelectBuilder) In(col string, inList []int) *AccSelectBuilder {
	if len(inList) == 0 {
		return b
	}

	// TODO: Optimise this
	where := col + " IN("
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
func (b *AccSelectBuilder) InPQuery(col string, inList []int) (*sql.Rows, error) {
	if len(inList) == 0 {
		return nil, sql.ErrNoRows
	}
	// TODO: Optimise this
	where := col + " IN("

	idList := make([]interface{}, len(inList))
	for i, id := range inList {
		idList[i] = strconv.Itoa(id)
		where += "?,"
	}
	where = where[0:len(where)-1] + ")"

	if b.where != "" {
		where += " AND " + b.where
	}

	b.where = where
	return b.Query(idList...)
}

func (b *AccSelectBuilder) InQ(col string, sb *AccSelectBuilder) *AccSelectBuilder {
	b.inChain = sb
	b.inColumn = col
	return b
}

func (b *AccSelectBuilder) DateCutoff(col string, quantity int, unit string) *AccSelectBuilder {
	b.dateCutoff = &dateCutoff{col, quantity, unit, 0}
	return b
}

func (b *AccSelectBuilder) DateOlderThanQ(col, unit string) *AccSelectBuilder {
	b.dateCutoff = &dateCutoff{col, 0, unit, 11}
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

func (b *AccSelectBuilder) ComplexPrepare() *sql.Stmt {
	selectBuilder := b.build.GetAdapter().Builder().Select().FromAcc(b)
	return b.build.prepare(b.build.GetAdapter().ComplexSelect(selectBuilder))
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

func (w *AccRowWrap) Scan(dest ...interface{}) error {
	if w.err != nil {
		return w.err
	}
	return w.row.Scan(dest...)
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
func (b *AccSelectBuilder) Each(h func(*sql.Rows) error) error {
	query, e := b.query()
	if e != nil {
		return e
	}
	rows, e := b.build.query(query)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		if e = h(rows); e != nil {
			return e
		}
	}
	return rows.Err()
}
func (b *AccSelectBuilder) EachP(h func(*sql.Rows) error, p ...interface{}) error {
	query, e := b.query()
	if e != nil {
		return e
	}
	rows, e := b.build.query(query, p)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		if e = h(rows); e != nil {
			return e
		}
	}
	return rows.Err()
}
func (b *AccSelectBuilder) EachInt(h func(int) error) error {
	query, e := b.query()
	if e != nil {
		return e
	}
	rows, e := b.build.query(query)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var theInt int
		if e = rows.Scan(&theInt); e != nil {
			return e
		}
		if e = h(theInt); e != nil {
			return e
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

func (b *accInsertBuilder) Columns(cols string) *accInsertBuilder {
	b.columns = cols
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

type accBulkInsertBuilder struct {
	table    string
	columns  string
	fieldSet []string

	build *Accumulator
}

func (b *accBulkInsertBuilder) Columns(cols string) *accBulkInsertBuilder {
	b.columns = cols
	return b
}

func (b *accBulkInsertBuilder) Fields(fieldSet ...string) *accBulkInsertBuilder {
	b.fieldSet = fieldSet
	return b
}

func (b *accBulkInsertBuilder) Prepare() *sql.Stmt {
	return b.build.SimpleBulkInsert(b.table, b.columns, b.fieldSet)
}

func (b *accBulkInsertBuilder) Exec(args ...interface{}) (res sql.Result, err error) {
	query, err := b.build.adapter.SimpleBulkInsert("", b.table, b.columns, b.fieldSet)
	if err != nil {
		return res, err
	}
	return b.build.exec(query, args...)
}

func (b *accBulkInsertBuilder) Run(args ...interface{}) (int, error) {
	query, err := b.build.adapter.SimpleBulkInsert("", b.table, b.columns, b.fieldSet)
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

func (b *accCountBuilder) Where(w string) *accCountBuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += w
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

func (b *accCountBuilder) TotalP(params ...interface{}) (total int, err error) {
	stmt := b.Prepare()
	if stmt == nil {
		return 0, b.build.FirstError()
	}
	err = stmt.QueryRow(params).Scan(&total)
	return total, err
}

// TODO: Add a Sum builder for summing viewchunks up into one number for the dashboard?
