/* WIP Under Construction */
package qgen

var Install *installer

func init() {
	Install = &installer{instructions: []DB_Install_Instruction{}}
}

type DB_Install_Instruction struct {
	Table    string
	Contents string
	Type     string
}

// A set of wrappers around the generator methods, so we can use this in the installer
// TODO: Re-implement the query generation, query builder and installer adapters as layers on-top of a query text adapter
type installer struct {
	adapter      DB_Adapter
	instructions []DB_Install_Instruction
	plugins      []QueryPlugin
}

func (install *installer) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	install.adapter = adap
	install.instructions = []DB_Install_Instruction{}
	return nil
}

func (install *installer) SetAdapterInstance(adapter DB_Adapter) {
	install.adapter = adapter
	install.instructions = []DB_Install_Instruction{}
}

func (install *installer) RegisterPlugin(plugin QueryPlugin) {
	install.plugins = append(install.plugins, plugin)
}

func (install *installer) CreateTable(table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) error {
	for _, plugin := range install.plugins {
		err := plugin.Hook("CreateTableStart", table, charset, collation, columns, keys)
		if err != nil {
			return err
		}
	}
	res, err := install.adapter.CreateTable("_installer", table, charset, collation, columns, keys)
	if err != nil {
		return err
	}
	for _, plugin := range install.plugins {
		err := plugin.Hook("CreateTableAfter", table, charset, collation, columns, keys, res)
		if err != nil {
			return err
		}
	}
	install.instructions = append(install.instructions, DB_Install_Instruction{table, res, "create-table"})
	return nil
}

func (install *installer) SimpleInsert(table string, columns string, fields string) error {
	for _, plugin := range install.plugins {
		err := plugin.Hook("SimpleInsertStart", table, columns, fields)
		if err != nil {
			return err
		}
	}
	res, err := install.adapter.SimpleInsert("_installer", table, columns, fields)
	if err != nil {
		return err
	}
	for _, plugin := range install.plugins {
		err := plugin.Hook("SimpleInsertAfter", table, columns, fields, res)
		if err != nil {
			return err
		}
	}
	install.instructions = append(install.instructions, DB_Install_Instruction{table, res, "insert"})
	return nil
}

func (install *installer) Write() error {
	var inserts string
	// We can't escape backticks, so we have to dump it out a file at a time
	for _, instr := range install.instructions {
		if instr.Type == "create-table" {
			err := writeFile("./schema/"+install.adapter.GetName()+"/query_"+instr.Table+".sql", instr.Contents)
			if err != nil {
				return err
			}
		} else {
			inserts += instr.Contents + ";\n"
		}
	}

	err := writeFile("./schema/"+install.adapter.GetName()+"/inserts.sql", inserts)
	if err != nil {
		return err
	}

	for _, plugin := range install.plugins {
		err := plugin.Write()
		if err != nil {
			return err
		}
	}

	return nil
}
