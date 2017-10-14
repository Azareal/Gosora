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

func (install *installer) CreateTable(table string, charset string, collation string, columns []DB_Table_Column, keys []DB_Table_Key) error {
	res, err := install.adapter.CreateTable("_installer", table, charset, collation, columns, keys)
	if err != nil {
		return err
	}
	install.instructions = append(install.instructions, DB_Install_Instruction{table, res, "create-table"})
	return nil
}

func (install *installer) SimpleInsert(table string, columns string, fields string) error {
	res, err := install.adapter.SimpleInsert("_installer", table, columns, fields)
	if err != nil {
		return err
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
	return writeFile("./schema/"+install.adapter.GetName()+"/inserts.sql", inserts)
}
