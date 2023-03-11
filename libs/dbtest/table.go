package dbtest

type Table struct {
	Rows []interface{}
}

func newTable() *Table {
	return &Table{Rows: []interface{}{}}
}
