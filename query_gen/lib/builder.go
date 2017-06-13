/* WIP Under Construction */
package qgen
import "database/sql"

var Builder *builder

func init() {
	Builder = &builder{conn:nil}
}

// A set of wrappers around the generator methods, so that we can use this inline in Gosora
type builder struct
{
	conn *sql.DB
	adapter DB_Adapter
}

func (build *builder) SetConn(conn *sql.DB) {
	build.conn = conn
}

func (build *builder) SetAdapter(name string) error {
	adap, err := GetAdapter(name)
	if err != nil {
		return err
	}
	build.adapter = adap
	return nil
}

func (build *builder) SimpleSelect(table string, columns string, where string, orderby string/*, offset int, maxCount int*/) (stmt *sql.Stmt, err error) {
	res, err := build.adapter.SimpleSelect("_builder", table, columns, where, orderby /*, offset, maxCount*/)
	if err != nil {
		return stmt, err
	}
	return build.conn.Prepare(res)
}
