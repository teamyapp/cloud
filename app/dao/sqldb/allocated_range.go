package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type AllocatedRange struct {
	db *sql.DB
}

var _ dao.AllocatedRange = (*AllocatedRange)(nil)

func (a AllocatedRange) FindByKey(key string) (entity.AllocatedRange, error) {
	row := a.db.QueryRow(`
SELECT key, range_end, next_number
FROM identity_allocated_range
WHERE key = $1;
`,
		key)

	allocatedRange := entity.AllocatedRange{}
	err := row.Scan(&allocatedRange.Key, &allocatedRange.RangeEnd, &allocatedRange.NextNumber)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.AllocatedRange{}, dao.ErrNotFound(fmt.Sprintf(
			"allocated range not found: key=%v",
			key))
	}

	return allocatedRange, err
}

func (a AllocatedRange) Add(allocatedRange entity.AllocatedRange) error {
	_, err := a.db.Exec(`
INSERT INTO identity_allocated_range (key, range_end, next_number)
VALUES ($1, $2, $3);
`,
		allocatedRange.Key,
		allocatedRange.RangeEnd,
		allocatedRange.NextNumber)
	return err
}

func (a AllocatedRange) Update(allocatedRange entity.AllocatedRange) error {
	_, err := a.db.Exec(`
UPDATE identity_sign_in_session
SET range_end = $1, next_number = $2
WHERE key = $3;
`,
		allocatedRange.RangeEnd,
		allocatedRange.NextNumber,
		allocatedRange.Key)
	return err
}

func NewAllocatedRange(sqlDB *sql.DB) AllocatedRange {
	return AllocatedRange{
		db: sqlDB,
	}
}
