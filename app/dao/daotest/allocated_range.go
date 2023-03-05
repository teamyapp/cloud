package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type AllocatedRange struct {
	db *InMemoryDB
}

var _ dao.AllocatedRange = (*AllocatedRange)(nil)

func (a AllocatedRange) FindAllocatedRangeByKey(ct context.Context, key string) (entity.AllocatedRange, *errs.Error) {
	table, err := a.db.GetTable(AllocatedRangeTableName)
	if err != nil {
		return entity.AllocatedRange{}, err
	}

	for _, rawRow := range table.rows {
		allocatedRange := rawRow.(entity.AllocatedRange)
		if allocatedRange.Key == key {
			return allocatedRange, nil
		}
	}

	return entity.AllocatedRange{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: key=%v", key),
	}
}

func (a AllocatedRange) CreateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error {
	_, err := a.FindAllocatedRangeByKey(ct, allocatedRange.Key)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: key=%v", allocatedRange.Key),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := a.db.GetTable(AllocatedRangeTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, allocatedRange)
	return nil
}

func (a AllocatedRange) UpdateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error {
	table, err := a.db.GetTable(AllocatedRangeTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		currAllocatedRange := rawRow.(entity.AllocatedRange)
		if currAllocatedRange.Key == allocatedRange.Key {
			rows = append(rows, allocatedRange)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: key=%v", allocatedRange.Key),
	}
}

func NewAllocatedRange(db *InMemoryDB) AllocatedRange {
	return AllocatedRange{
		db: db,
	}
}
