package gen

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/obs"
)

type UniqueNumber struct {
	dataCollector     obs.DataCollector
	allocatedRangeDao dao.AllocatedRange
	name              string
	rangeSize         uint64
	allocatedRange    entity.AllocatedRange
}

func (u *UniqueNumber) GenerateUniqueNumber(ct context.Context) (uint64, error) {
	if u.allocatedRange.NextNumber > u.allocatedRange.RangeEnd {
		err := u.allocateNewRange()
		if err != nil {
			u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})

			return uint64(0), err
		}
	}

	num := u.allocatedRange.NextNumber
	u.allocatedRange.NextNumber++
	return num, nil
}

func (u *UniqueNumber) allocateNewRange() error {
	if u.allocatedRange.RangeEnd == math.MaxInt64 {
		err := fmt.Errorf("out of number to allocate")
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	newRangeStart := u.allocatedRange.RangeEnd + 1
	newRangeEnd := min(u.allocatedRange.RangeEnd+u.rangeSize, math.MaxUint64)
	newRange := entity.AllocatedRange{
		Key:        u.name,
		RangeEnd:   newRangeEnd,
		NextNumber: newRangeStart,
	}
	err := u.allocatedRangeDao.UpdateAllocatedRange(newRange)
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	u.allocatedRange = newRange
	u.dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"AllocatedRange": newRange.String(),
		},
	})
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
	dataCollector obs.DataCollector,
	allocatedRangeDao dao.AllocatedRange,
	name string,
	rangeSize uint64,
) (*UniqueNumber, error) {
	allocatedRange, err := allocatedRangeDao.FindAllocatedRangeByKey(name)
	var errNotFound dao.ErrNotFound
	if err != nil {
		if !errors.As(err, &errNotFound) {
			return nil, err
		}

		allocatedRange = entity.AllocatedRange{
			Key:        name,
			RangeEnd:   0,
			NextNumber: 0,
		}

		err = allocatedRangeDao.CreateAllocatedRange(allocatedRange)
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	err = uniqueNumber.allocateNewRange()
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return uniqueNumber, err
}

type UniqueNumberFactory struct {
	dataCollector     obs.DataCollector
	allocatedRangeDao dao.AllocatedRange
	rangeSize         uint64
}

func (u UniqueNumberFactory) MakeUniqueNumber(name string) (*UniqueNumber, error) {
	return newUniqueNumber(u.dataCollector, u.allocatedRangeDao, name, u.rangeSize)
}

func NewUniqueNumberFactory(
	dataCollector obs.DataCollector,
	allocatedRangeDao dao.AllocatedRange,
	rangeSize uint64,
) UniqueNumberFactory {
	return UniqueNumberFactory{
		dataCollector:     dataCollector,
		allocatedRangeDao: allocatedRangeDao,
		rangeSize:         rangeSize,
	}
}
