package gen

import (
	"context"
	"math"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type UniqueNumber struct {
	dataCollector     telemetry.DataCollector
	allocatedRangeDao dao.AllocatedRange
	name              string
	rangeSize         uint64
	allocatedRange    entity.AllocatedRange
}

func (u *UniqueNumber) GenerateUniqueNumber(ct context.Context) (uint64, *errs.Error) {
	if u.allocatedRange.NextNumber > u.allocatedRange.RangeEnd {
		err := u.allocateNewRange(ct)
		if err != nil {
			u.dataCollector.Logger.ErrorWithContext(ct, err)
			return uint64(0), err
		}
	}

	num := u.allocatedRange.NextNumber
	u.allocatedRange.NextNumber++
	return num, nil
}

func (u *UniqueNumber) allocateNewRange(ct context.Context) *errs.Error {
	if u.allocatedRange.RangeEnd == math.MaxInt64 {
		internalErr := &errs.Error{
			Code:    errs.Unknown,
			Message: "out of number to allocate",
		}
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	newRangeStart := u.allocatedRange.RangeEnd + 1
	newRangeEnd := min(u.allocatedRange.RangeEnd+u.rangeSize, math.MaxUint64)
	newRange := entity.AllocatedRange{
		Key:        u.name,
		RangeEnd:   newRangeEnd,
		NextNumber: newRangeStart,
	}
	err := u.allocatedRangeDao.UpdateAllocatedRange(ct, newRange)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	u.allocatedRange = newRange
	u.dataCollector.Logger.InfoWithContext(ct, newRange)
	return nil
}

func min[Number int | uint64](num1 Number, num2 Number) Number {
	if num1 <= num2 {
		return num1
	} else {
		return num2
	}
}

func newUniqueNumber(
	dataCollector telemetry.DataCollector,
	allocatedRangeDao dao.AllocatedRange,
	name string,
	rangeSize uint64,
) (*UniqueNumber, *errs.Error) {
	ct := context.Background()
	allocatedRange, err := allocatedRangeDao.FindAllocatedRangeByKey(ct, name)
	if err != nil {
		if err.Code != errs.NotFound {
			return nil, err
		}

		allocatedRange = entity.AllocatedRange{
			Key:        name,
			RangeEnd:   0,
			NextNumber: 0,
		}

		err = allocatedRangeDao.CreateAllocatedRange(ct, allocatedRange)
		if err != nil {
			dataCollector.Logger.WarningWithContext(ct, err)
			return nil, err
		}
	}

	uniqueNumber := &UniqueNumber{
		dataCollector:     dataCollector,
		name:              name,
		rangeSize:         rangeSize,
		allocatedRange:    allocatedRange,
		allocatedRangeDao: allocatedRangeDao,
	}
	err = uniqueNumber.allocateNewRange(ct)
	if err != nil {
		dataCollector.Logger.ErrorWithContext(ct, err)
	}

	return uniqueNumber, err
}

type UniqueNumberFactory struct {
	dataCollector     telemetry.DataCollector
	allocatedRangeDao dao.AllocatedRange
	rangeSize         uint64
}

func (u UniqueNumberFactory) MakeUniqueNumber(name string) (*UniqueNumber, *errs.Error) {
	return newUniqueNumber(u.dataCollector, u.allocatedRangeDao, name, u.rangeSize)
}

func NewUniqueNumberFactory(
	dataCollector telemetry.DataCollector,
	allocatedRangeDao dao.AllocatedRange,
	rangeSize uint64,
) UniqueNumberFactory {
	return UniqueNumberFactory{
		dataCollector:     dataCollector,
		allocatedRangeDao: allocatedRangeDao,
		rangeSize:         rangeSize,
	}
}
