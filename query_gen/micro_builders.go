package qgen

type dateCutoff struct {
	Column   string
	Quantity int
	Unit     string
	Type int
}

type prebuilder struct {
	adapter Adapter
}

func (build *prebuilder) Select(nlist ...string) *selectPrebuilder {
	name := optString(nlist, "")
	return &selectPrebuilder{name, "", "", "", "", "", nil, nil, "", build.adapter}
}

func (build *prebuilder) Count(nlist ...string) *selectPrebuilder {
	name := optString(nlist, "")
	return &selectPrebuilder{name, "", "COUNT(*)", "", "", "", nil, nil, "", build.adapter}
}

func (build *prebuilder) Insert(nlist ...string) *insertPrebuilder {
	name := optString(nlist, "")
	return &insertPrebuilder{name, "", "", "", build.adapter}
}

func (build *prebuilder) Update(nlist ...string) *updatePrebuilder {
	name := optString(nlist, "")
	return &updatePrebuilder{name, "", "", "", nil, nil, build.adapter}
}

func (build *prebuilder) Delete(nlist ...string) *deletePrebuilder {
	name := optString(nlist, "")
	return &deletePrebuilder{name, "", "", build.adapter}
}

type deletePrebuilder struct {
	name  string
	table string
	where string

	build Adapter
}

func (delete *deletePrebuilder) Table(table string) *deletePrebuilder {
	delete.table = table
	return delete
}

func (delete *deletePrebuilder) Where(where string) *deletePrebuilder {
	if delete.where != "" {
		delete.where += " AND "
	}
	delete.where += where
	return delete
}

func (delete *deletePrebuilder) Text() (string, error) {
	return delete.build.SimpleDelete(delete.name, delete.table, delete.where)
}

func (delete *deletePrebuilder) Parse() {
	delete.build.SimpleDelete(delete.name, delete.table, delete.where)
}

type updatePrebuilder struct {
	name          string
	table         string
	set           string
	where         string
	dateCutoff *dateCutoff // We might want to do this in a slightly less hacky way
	whereSubQuery *selectPrebuilder

	build Adapter
}

func qUpdate(table string, set string, where string) *updatePrebuilder {
	return &updatePrebuilder{table: table, set: set, where: where}
}

func (update *updatePrebuilder) Table(table string) *updatePrebuilder {
	update.table = table
	return update
}

func (update *updatePrebuilder) Set(set string) *updatePrebuilder {
	update.set = set
	return update
}

func (update *updatePrebuilder) Where(where string) *updatePrebuilder {
	if update.where != "" {
		update.where += " AND "
	}
	update.where += where
	return update
}

func (update *updatePrebuilder) WhereQ(sel *selectPrebuilder) *updatePrebuilder {
	update.whereSubQuery = sel
	return update
}

func (update *updatePrebuilder) Text() (string, error) {
	return update.build.SimpleUpdate(update)
}

func (update *updatePrebuilder) Parse() {
	update.build.SimpleUpdate(update)
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

func (selectItem *selectPrebuilder) Table(table string) *selectPrebuilder {
	selectItem.table = table
	return selectItem
}

func (selectItem *selectPrebuilder) Columns(columns string) *selectPrebuilder {
	selectItem.columns = columns
	return selectItem
}

func (selectItem *selectPrebuilder) Where(where string) *selectPrebuilder {
	if selectItem.where != "" {
		selectItem.where += " AND "
	}
	selectItem.where += where
	return selectItem
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
func (selectItem *selectPrebuilder) Text() (string, error) {
	return selectItem.build.SimpleSelect(selectItem.name, selectItem.table, selectItem.columns, selectItem.where, selectItem.orderby, selectItem.limit)
}

// TODO: Add support for dateCutoff
func (selectItem *selectPrebuilder) Parse() {
	selectItem.build.SimpleSelect(selectItem.name, selectItem.table, selectItem.columns, selectItem.where, selectItem.orderby, selectItem.limit)
}

type insertPrebuilder struct {
	name    string
	table   string
	columns string
	fields  string

	build Adapter
}

func (insert *insertPrebuilder) Table(table string) *insertPrebuilder {
	insert.table = table
	return insert
}

func (insert *insertPrebuilder) Columns(columns string) *insertPrebuilder {
	insert.columns = columns
	return insert
}

func (insert *insertPrebuilder) Fields(fields string) *insertPrebuilder {
	insert.fields = fields
	return insert
}

func (insert *insertPrebuilder) Text() (string, error) {
	return insert.build.SimpleInsert(insert.name, insert.table, insert.columns, insert.fields)
}

func (insert *insertPrebuilder) Parse() {
	insert.build.SimpleInsert(insert.name, insert.table, insert.columns, insert.fields)
}

/*type countPrebuilder struct {
	name  string
	table string
	where string
	limit string

	build Adapter
}

func (count *countPrebuilder) Table(table string) *countPrebuilder {
	count.table = table
	return count
}

func (count *countPrebuilder) Where(where string) *countPrebuilder {
	if count.where != "" {
		count.where += " AND "
	}
	count.where += where
	return count
}

func (count *countPrebuilder) Limit(limit string) *countPrebuilder {
	count.limit = limit
	return count
}

func (count *countPrebuilder) Text() (string, error) {
	return count.build.SimpleCount(count.name, count.table, count.where, count.limit)
}

func (count *countPrebuilder) Parse() {
	count.build.SimpleCount(count.name, count.table, count.where, count.limit)
}*/

func optString(nlist []string, defaultStr string) string {
	if len(nlist) == 0 {
		return defaultStr
	}
	return nlist[0]
}
