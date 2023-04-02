package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type AllocatedRange struct {
	db *sql.DB
}

var _ dao.AllocatedRange = (*AllocatedRange)(nil)

func (a AllocatedRange) FindAllocatedRangeByKey(ct context.Context, key string) (entity.AllocatedRange, *errs.Error) {
	row := a.db.QueryRow(`
	SELECT key, range_end
	FROM allocated_range
	WHERE key = $1;
	`,
		key)

	allocatedRange := entity.AllocatedRange{}
	err := row.Scan(&allocatedRange.Key, &allocatedRange.RangeEnd)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.AllocatedRange{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("allocated range not found: key=%v", key))
	}

	if err != nil {
		return entity.AllocatedRange{}, errs.NewError(errs.Unknown, err.Error())
	}

	return allocatedRange, nil
}

func (a AllocatedRange) CreateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error {
	_, err := a.db.Exec(`
	INSERT INTO allocated_range (key, range_end)
	VALUES ($1, $2);
	`,
		allocatedRange.Key,
		allocatedRange.RangeEnd)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AllocatedRange) UpdateAllocatedRange(ct context.Context, allocatedRange entity.AllocatedRange) *errs.Error {
	_, err := a.db.Exec(`
	UPDATE allocated_range
	SET range_end = $1
	WHERE key = $2;
	`,
		allocatedRange.RangeEnd,
		allocatedRange.Key)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAllocatedRange(sqlDB *sql.DB) AllocatedRange {
	return AllocatedRange{
		db: sqlDB,
	}
}
