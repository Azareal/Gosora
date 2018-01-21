package qgen

type dateCutoff struct {
	Column   string
	Quantity int
	Unit     string
}

type prebuilder struct {
	adapter Adapter
}

func (build *prebuilder) Select(nlist ...string) *selectPrebuilder {
	name := optString(nlist, "_builder")
	return &selectPrebuilder{name, "", "", "", "", "", nil, nil, "", build.adapter}
}

func (build *prebuilder) Insert(nlist ...string) *insertPrebuilder {
	name := optString(nlist, "_builder")
	return &insertPrebuilder{name, "", "", "", build.adapter}
}

func (build *prebuilder) Update(nlist ...string) *updatePrebuilder {
	name := optString(nlist, "_builder")
	return &updatePrebuilder{name, "", "", "", build.adapter}
}

func (build *prebuilder) Delete(nlist ...string) *deletePrebuilder {
	name := optString(nlist, "_builder")
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
	delete.where = where
	return delete
}

func (delete *deletePrebuilder) Text() (string, error) {
	return delete.build.SimpleDelete(delete.name, delete.table, delete.where)
}

func (delete *deletePrebuilder) Parse() {
	delete.build.SimpleDelete(delete.name, delete.table, delete.where)
}

type updatePrebuilder struct {
	name  string
	table string
	set   string
	where string

	build Adapter
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
	update.where = where
	return update
}

func (update *updatePrebuilder) Text() (string, error) {
	return update.build.SimpleUpdate(update.name, update.table, update.set, update.where)
}

func (update *updatePrebuilder) Parse() {
	update.build.SimpleUpdate(update.name, update.table, update.set, update.where)
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
	selectItem.where = where
	return selectItem
}

func (selectItem *selectPrebuilder) InQ(subBuilder *selectPrebuilder) *selectPrebuilder {
	selectItem.inChain = subBuilder
	return selectItem
}

func (selectItem *selectPrebuilder) Orderby(orderby string) *selectPrebuilder {
	selectItem.orderby = orderby
	return selectItem
}

func (selectItem *selectPrebuilder) Limit(limit string) *selectPrebuilder {
	selectItem.limit = limit
	return selectItem
}

// TODO: We probably want to avoid the double allocation of two builders somehow
func (selectItem *selectPrebuilder) FromAcc(accBuilder *accSelectBuilder) *selectPrebuilder {
	selectItem.table = accBuilder.table
	selectItem.columns = accBuilder.columns
	selectItem.where = accBuilder.where
	selectItem.orderby = accBuilder.orderby
	selectItem.limit = accBuilder.limit

	selectItem.dateCutoff = accBuilder.dateCutoff
	if accBuilder.inChain != nil {
		selectItem.inChain = &selectPrebuilder{"__builder", accBuilder.inChain.table, accBuilder.inChain.columns, accBuilder.inChain.where, accBuilder.inChain.orderby, accBuilder.inChain.limit, accBuilder.inChain.dateCutoff, nil, "", selectItem.build}
		selectItem.inColumn = accBuilder.inColumn
	}
	return selectItem
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

type countPrebuilder struct {
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
	count.where = where
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
}

func optString(nlist []string, defaultStr string) string {
	if len(nlist) == 0 {
		return defaultStr
	}
	return nlist[0]
}
