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

func (ins *installer) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	ins.SetAdapterInstance(adap)
	return nil
}

func (ins *installer) SetAdapterInstance(adapter Adapter) {
	ins.adapter = adapter
	ins.instructions = []DBInstallInstruction{}
}

func (ins *installer) AddPlugins(plugins ...QueryPlugin) {
	ins.plugins = append(ins.plugins, plugins...)
}

func (ins *installer) CreateTable(table string, charset string, collation string, columns []DBTableColumn, keys []DBTableKey) error {
	tableStruct := &DBInstallTable{table, charset, collation, columns, keys}
	err := ins.RunHook("CreateTableStart", tableStruct)
	if err != nil {
		return err
	}
	res, err := ins.adapter.CreateTable("", table, charset, collation, columns, keys)
	if err != nil {
		return err
	}
	err = ins.RunHook("CreateTableAfter", tableStruct)
	if err != nil {
		return err
	}
	ins.instructions = append(ins.instructions, DBInstallInstruction{table, res, "create-table"})
	ins.tables = append(ins.tables, tableStruct)
	return nil
}

// TODO: Let plugins manipulate the parameters like in CreateTable
func (ins *installer) AddIndex(table string, iname string, colname string) error {
	err := ins.RunHook("AddIndexStart", table, iname, colname)
	if err != nil {
		return err
	}
	res, err := ins.adapter.AddIndex("", table, iname, colname)
	if err != nil {
		return err
	}
	err = ins.RunHook("AddIndexAfter", table, iname, colname)
	if err != nil {
		return err
	}
	ins.instructions = append(ins.instructions, DBInstallInstruction{table, res, "index"})
	return nil
}

// TODO: Let plugins manipulate the parameters like in CreateTable
func (ins *installer) SimpleInsert(table string, columns string, fields string) error {
	err := ins.RunHook("SimpleInsertStart", table, columns, fields)
	if err != nil {
		return err
	}
	res, err := ins.adapter.SimpleInsert("", table, columns, fields)
	if err != nil {
		return err
	}
	err = ins.RunHook("SimpleInsertAfter", table, columns, fields, res)
	if err != nil {
		return err
	}
	ins.instructions = append(ins.instructions, DBInstallInstruction{table, res, "insert"})
	return nil
}

func (ins *installer) RunHook(name string, args ...interface{}) error {
	for _, plugin := range ins.plugins {
		err := plugin.Hook(name, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ins *installer) Write() error {
	var inserts string
	// We can't escape backticks, so we have to dump it out a file at a time
	for _, instr := range ins.instructions {
		if instr.Type == "create-table" {
			err := writeFile("./schema/"+ins.adapter.GetName()+"/query_"+instr.Table+".sql", instr.Contents)
			if err != nil {
				return err
			}
		} else {
			inserts += instr.Contents + ";\n"
		}
	}

	err := writeFile("./schema/"+ins.adapter.GetName()+"/inserts.sql", inserts)
	if err != nil {
		return err
	}

	for _, plugin := range ins.plugins {
		err := plugin.Write()
		if err != nil {
			return err
		}
	}

	return nil
}
