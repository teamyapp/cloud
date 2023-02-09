package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type AllocatedRange struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
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
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("allocated range not found: key=%v", key),
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: err,
		})
		return entity.AllocatedRange{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AllocatedRange{}, internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewAllocatedRange(dataCollector telemetry.DataCollector, sqlDB *sql.DB) AllocatedRange {
	return AllocatedRange{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
