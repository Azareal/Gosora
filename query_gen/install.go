package qgen

var Install *installer

func init() {
	Install = &installer{instructions: []DBInstallInstruction{}}
}

type DBInstallInstruction struct {
	Table    string
	Contents string
	Type     string
}

// TODO: Add methods to this to construct it OO-like
type DBInstallTable struct {
	Name      string
	Charset   string
	Collation string
	Columns   []DBTableColumn
	Keys      []DBTableKey
}

// A set of wrappers around the generator methods, so we can use this in the installer
// TODO: Re-implement the query generation, query builder and installer adapters as layers on-top of a query text adapter
type installer struct {
	adapter      Adapter
	instructions []DBInstallInstruction
	tables       []*DBInstallTable // TODO: Use this in Record() in the next commit to allow us to auto-migrate settings rather than manually patching them in on upgrade
	plugins      []QueryPlugin
}

func (i *installer) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	i.SetAdapterInstance(adap)
	return nil
}

func (i *installer) SetAdapterInstance(adap Adapter) {
	i.adapter = adap
	i.instructions = []DBInstallInstruction{}
}

func (i *installer) AddPlugins(plugins ...QueryPlugin) {
	i.plugins = append(i.plugins, plugins...)
}

func (i *installer) CreateTable(table, charset, collation string, columns []DBTableColumn, keys []DBTableKey) error {
	tableStruct := &DBInstallTable{table, charset, collation, columns, keys}
	err := i.RunHook("CreateTableStart", tableStruct)
	if err != nil {
		return err
	}
	res, err := i.adapter.CreateTable("", table, charset, collation, columns, keys)
	if err != nil {
		return err
	}
	err = i.RunHook("CreateTableAfter", tableStruct)
	if err != nil {
		return err
	}
	i.instructions = append(i.instructions, DBInstallInstruction{table, res, "create-table"})
	i.tables = append(i.tables, tableStruct)
	return nil
}

// TODO: Let plugins manipulate the parameters like in CreateTable
func (i *installer) AddIndex(table, iname, colname string) error {
	err := i.RunHook("AddIndexStart", table, iname, colname)
	if err != nil {
		return err
	}
	res, err := i.adapter.AddIndex("", table, iname, colname)
	if err != nil {
		return err
	}
	err = i.RunHook("AddIndexAfter", table, iname, colname)
	if err != nil {
		return err
	}
	i.instructions = append(i.instructions, DBInstallInstruction{table, res, "index"})
	return nil
}

func (i *installer) AddKey(table, column string, key DBTableKey) error {
	err := i.RunHook("AddKeyStart", table, column, key)
	if err != nil {
		return err
	}
	res, err := i.adapter.AddKey("", table, column, key)
	if err != nil {
		return err
	}
	err = i.RunHook("AddKeyAfter", table, column, key)
	if err != nil {
		return err
	}
	i.instructions = append(i.instructions, DBInstallInstruction{table, res, "key"})
	return nil
}

// TODO: Let plugins manipulate the parameters like in CreateTable
func (i *installer) SimpleInsert(table, columns, fields string) error {
	err := i.RunHook("SimpleInsertStart", table, columns, fields)
	if err != nil {
		return err
	}
	res, err := i.adapter.SimpleInsert("", table, columns, fields)
	if err != nil {
		return err
	}
	err = i.RunHook("SimpleInsertAfter", table, columns, fields, res)
	if err != nil {
		return err
	}
	i.instructions = append(i.instructions, DBInstallInstruction{table, res, "insert"})
	return nil
}

func (i *installer) RunHook(name string, args ...interface{}) error {
	for _, plugin := range i.plugins {
		err := plugin.Hook(name, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *installer) Write() error {
	var inserts string
	// We can't escape backticks, so we have to dump it out a file at a time
	for _, instr := range i.instructions {
		if instr.Type == "create-table" {
			err := writeFile("./schema/"+i.adapter.GetName()+"/query_"+instr.Table+".sql", instr.Contents)
			if err != nil {
				return err
			}
		} else {
			inserts += instr.Contents + ";\n"
		}
	}

	err := writeFile("./schema/"+i.adapter.GetName()+"/inserts.sql", inserts)
	if err != nil {
		return err
	}

	for _, plugin := range i.plugins {
		err := plugin.Write()
		if err != nil {
			return err
		}
	}

	return nil
}
