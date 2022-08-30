package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/obs"
)

type AllocatedRange struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.AllocatedRange = (*AllocatedRange)(nil)

func (a AllocatedRange) FindAllocatedRangeByKey(key string) (entity.AllocatedRange, error) {
	row := a.db.QueryRow(`
	SELECT key, range_end
	FROM allocated_range
	WHERE key = $1;
	`,
		key)

	allocatedRange := entity.AllocatedRange{}
	err := row.Scan(&allocatedRange.Key, &allocatedRange.RangeEnd)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.AllocatedRange{}, dao.ErrNotFound(fmt.Sprintf(
			"allocated range not found: key=%v",
			key))
	}

	if err != nil {
		a.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return allocatedRange, err
}

func (a AllocatedRange) CreateAllocatedRange(allocatedRange entity.AllocatedRange) error {
	_, err := a.db.Exec(`
	INSERT INTO allocated_range (key, range_end)
	VALUES ($1, $2);
	`,
		allocatedRange.Key,
		allocatedRange.RangeEnd)
	if err != nil {
		a.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (a AllocatedRange) UpdateAllocatedRange(allocatedRange entity.AllocatedRange) error {
	_, err := a.db.Exec(`
	UPDATE allocated_range
	SET range_end = $1
	WHERE key = $2;
	`,
		allocatedRange.RangeEnd,
		allocatedRange.Key)
	if err != nil {
		a.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewAllocatedRange(dataCollector obs.DataCollector, sqlDB *sql.DB) AllocatedRange {
	return AllocatedRange{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
