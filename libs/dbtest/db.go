package dbtest

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/teamyapp/cloud/libs/errs"
)

type InMemoryDB struct {
	tables map[string]*Table
}

func (i InMemoryDB) GetTable(name string) (*Table, *errs.Error) {
	table, ok := i.tables[name]
	if !ok {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("table not found: name=%v", name),
		}
	}

	return table, nil
}

func (i InMemoryDB) CreateTable(name string) {
	i.tables[name] = newTable()
}

func (i InMemoryDB) InitTable(name string, rows []interface{}) {
	table := newTable()
	table.Rows = rows
	i.tables[name] = table
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{tables: map[string]*Table{}}
}
