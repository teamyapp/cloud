package service

import (
	"context"
	"math"
	"sync"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type UniqueNumberGen struct {
	logger            telemetry.Logger
	allocatedRangeDao dao.AllocatedRange
	name              string
	rangeSize         uint64
	allocatedRange    entity.AllocatedRange
	mu                sync.Mutex
}

func (u *UniqueNumberGen) GenerateUniqueNumber(ct context.Context) (uint64, *errs.Error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.allocatedRange.NextNumber > u.allocatedRange.RangeEnd {
		err := u.allocateNewRange(ct)
		if err != nil {
			return uint64(0), err
		}
	}

	num := u.allocatedRange.NextNumber
	u.allocatedRange.NextNumber++
	return num, nil
}

func (u *UniqueNumberGen) allocateNewRange(ct context.Context) *errs.Error {
	if u.allocatedRange.RangeEnd == math.MaxInt64 {
		return errs.NewError(errs.ResourceExhausted, "out of number to allocate")
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
		return err
	}

	u.allocatedRange = newRange
	u.logger.InfoWithContext(ct, newRange)
	return nil
}

func min[Number int | uint64](num1 Number, num2 Number) Number {
	if num1 <= num2 {
		return num1
	} else {
		return num2
	}
}

func newUniqueNumberGen(
	logger telemetry.Logger,
	allocatedRangeDao dao.AllocatedRange,
	name string,
	rangeSize uint64,
) (*UniqueNumberGen, *errs.Error) {
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
			logger.WarningWithContext(ct, err)
			return nil, err
		}
	}

	uniqueNumberGen := &UniqueNumberGen{
		logger:            logger,
		name:              name,
		rangeSize:         rangeSize,
		allocatedRange:    allocatedRange,
		allocatedRangeDao: allocatedRangeDao,
	}
	err = uniqueNumberGen.allocateNewRange(ct)
	if err != nil {
		logger.ErrorWithContext(ct, err)
	}

	return uniqueNumberGen, err
}

type UniqueNumberGenFactory struct {
	logger            telemetry.Logger
	allocatedRangeDao dao.AllocatedRange
	rangeSize         uint64
}

func (u UniqueNumberGenFactory) MakeUniqueNumberGen(name string) (*UniqueNumberGen, *errs.Error) {
	return newUniqueNumberGen(u.logger, u.allocatedRangeDao, name, u.rangeSize)
}

func NewUniqueNumberFactory(
	logger telemetry.Logger,
	allocatedRangeDao dao.AllocatedRange,
	rangeSize uint64,
) UniqueNumberGenFactory {
	return UniqueNumberGenFactory{
		logger:            logger,
		allocatedRangeDao: allocatedRangeDao,
		rangeSize:         rangeSize,
	}
}
