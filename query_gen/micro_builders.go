package qgen

type dateCutoff struct {
	Column   string
	Quantity int
	Unit     string
	Type     int
}

type prebuilder struct {
	adapter Adapter
}

func (b *prebuilder) Select(nlist ...string) *selectPrebuilder {
	name := optString(nlist, "")
	return &selectPrebuilder{name, "", "", "", "", "", nil, nil, "", b.adapter}
}

func (b *prebuilder) Count(nlist ...string) *selectPrebuilder {
	name := optString(nlist, "")
	return &selectPrebuilder{name, "", "COUNT(*)", "", "", "", nil, nil, "", b.adapter}
}

func (b *prebuilder) Insert(nlist ...string) *insertPrebuilder {
	name := optString(nlist, "")
	return &insertPrebuilder{name, "", "", "", b.adapter}
}

func (b *prebuilder) Update(nlist ...string) *updatePrebuilder {
	name := optString(nlist, "")
	return &updatePrebuilder{name, "", "", "", nil, nil, b.adapter}
}

func (b *prebuilder) Delete(nlist ...string) *deletePrebuilder {
	name := optString(nlist, "")
	return &deletePrebuilder{name, "", "", nil, b.adapter}
}

type deletePrebuilder struct {
	name       string
	table      string
	where      string
	dateCutoff *dateCutoff

	build Adapter
}

func (b *deletePrebuilder) Table(table string) *deletePrebuilder {
	b.table = table
	return b
}

func (b *deletePrebuilder) Where(where string) *deletePrebuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

// TODO: We probably want to avoid the double allocation of two builders somehow
func (b *deletePrebuilder) FromAcc(acc *accDeleteBuilder) *deletePrebuilder {
	b.table = acc.table
	b.where = acc.where
	b.dateCutoff = acc.dateCutoff
	return b
}

func (b *deletePrebuilder) Text() (string, error) {
	return b.build.SimpleDelete(b.name, b.table, b.where)
}

func (b *deletePrebuilder) Parse() {
	b.build.SimpleDelete(b.name, b.table, b.where)
}

type updatePrebuilder struct {
	name          string
	table         string
	set           string
	where         string
	dateCutoff    *dateCutoff // We might want to do this in a slightly less hacky way
	whereSubQuery *selectPrebuilder

	build Adapter
}

func qUpdate(table string, set string, where string) *updatePrebuilder {
	return &updatePrebuilder{table: table, set: set, where: where}
}

func (b *updatePrebuilder) Table(table string) *updatePrebuilder {
	b.table = table
	return b
}

func (b *updatePrebuilder) Set(set string) *updatePrebuilder {
	b.set = set
	return b
}

func (b *updatePrebuilder) Where(where string) *updatePrebuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

func (b *updatePrebuilder) WhereQ(sel *selectPrebuilder) *updatePrebuilder {
	b.whereSubQuery = sel
	return b
}

func (b *updatePrebuilder) Text() (string, error) {
	return b.build.SimpleUpdate(b)
}

func (b *updatePrebuilder) Parse() {
	b.build.SimpleUpdate(b)
}

type selectPrebuilder struct {
	name       string
	table      string
	columns    string
	where      string
	orderby    string
	limit      string
	dateCutoff *dateCutoff
	inChain    *selectPrebuilder
	inColumn   string // for inChain

	build Adapter
}

func (b *selectPrebuilder) Table(table string) *selectPrebuilder {
	b.table = table
	return b
}

func (b *selectPrebuilder) Columns(columns string) *selectPrebuilder {
	b.columns = columns
	return b
}

func (b *selectPrebuilder) Where(where string) *selectPrebuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

func (b *selectPrebuilder) InQ(subBuilder *selectPrebuilder) *selectPrebuilder {
	b.inChain = subBuilder
	return b
}

func (b *selectPrebuilder) Orderby(orderby string) *selectPrebuilder {
	b.orderby = orderby
	return b
}

func (b *selectPrebuilder) Limit(limit string) *selectPrebuilder {
	b.limit = limit
	return b
}

// TODO: We probably want to avoid the double allocation of two builders somehow
func (b *selectPrebuilder) FromAcc(acc *AccSelectBuilder) *selectPrebuilder {
	b.table = acc.table
	if acc.columns != "" {
		b.columns = acc.columns
	}
	b.where = acc.where
	b.orderby = acc.orderby
	b.limit = acc.limit

	b.dateCutoff = acc.dateCutoff
	if acc.inChain != nil {
		b.inChain = &selectPrebuilder{"", acc.inChain.table, acc.inChain.columns, acc.inChain.where, acc.inChain.orderby, acc.inChain.limit, acc.inChain.dateCutoff, nil, "", b.build}
		b.inColumn = acc.inColumn
	}
	return b
}

func (b *selectPrebuilder) FromCountAcc(acc *accCountBuilder) *selectPrebuilder {
	b.table = acc.table
	b.where = acc.where
	b.limit = acc.limit

	b.dateCutoff = acc.dateCutoff
	if acc.inChain != nil {
		b.inChain = &selectPrebuilder{"", acc.inChain.table, acc.inChain.columns, acc.inChain.where, acc.inChain.orderby, acc.inChain.limit, acc.inChain.dateCutoff, nil, "", b.build}
		b.inColumn = acc.inColumn
	}
	return b
}

// TODO: Add support for dateCutoff
func (b *selectPrebuilder) Text() (string, error) {
	return b.build.SimpleSelect(b.name, b.table, b.columns, b.where, b.orderby, b.limit)
}

// TODO: Add support for dateCutoff
func (b *selectPrebuilder) Parse() {
	b.build.SimpleSelect(b.name, b.table, b.columns, b.where, b.orderby, b.limit)
}

type insertPrebuilder struct {
	name    string
	table   string
	columns string
	fields  string

	build Adapter
}

func (b *insertPrebuilder) Table(table string) *insertPrebuilder {
	b.table = table
	return b
}

func (b *insertPrebuilder) Columns(columns string) *insertPrebuilder {
	b.columns = columns
	return b
}

func (b *insertPrebuilder) Fields(fields string) *insertPrebuilder {
	b.fields = fields
	return b
}

func (b *insertPrebuilder) Text() (string, error) {
	return b.build.SimpleInsert(b.name, b.table, b.columns, b.fields)
}

func (b *insertPrebuilder) Parse() {
	b.build.SimpleInsert(b.name, b.table, b.columns, b.fields)
}

/*type countPrebuilder struct {
	name  string
	table string
	where string
	limit string

	build Adapter
}

func (b *countPrebuilder) Table(table string) *countPrebuilder {
	b.table = table
	return b
}

func b *countPrebuilder) Where(where string) *countPrebuilder {
	if b.where != "" {
		b.where += " AND "
	}
	b.where += where
	return b
}

func (b *countPrebuilder) Limit(limit string) *countPrebuilder {
	b.limit = limit
	return b
}

func (b *countPrebuilder) Text() (string, error) {
	return b.build.SimpleCount(b.name, b.table, b.where, b.limit)
}

func (b *countPrebuilder) Parse() {
	b.build.SimpleCount(b.name, b.table, b.where, b.limit)
}*/

func optString(nlist []string, defaultStr string) string {
	if len(nlist) == 0 {
		return defaultStr
	}
	return nlist[0]
}
