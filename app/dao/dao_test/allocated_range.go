package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type AllocatedRange struct {
	allocatedRanges []entity.AllocatedRange
}

var _ dao.AllocatedRange = (*AllocatedRange)(nil)

func (a AllocatedRange) FindAllocatedRangeByKey(key string) (entity.AllocatedRange, error) {
	for _, allocatedRange := range a.allocatedRanges {
		if allocatedRange.Key == key {
			return entity.AllocatedRange{
				Key:        key,
				RangeEnd:   allocatedRange.RangeEnd,
				NextNumber: allocatedRange.NextNumber,
			}, nil
		}
	}

	return entity.AllocatedRange{}, nil
}

func (a AllocatedRange) CreateAllocatedRange(allocatedRange entity.AllocatedRange) error {
	a.allocatedRanges = append(a.allocatedRanges, entity.AllocatedRange{
		Key:        allocatedRange.Key,
		RangeEnd:   allocatedRange.RangeEnd,
		NextNumber: allocatedRange.NextNumber,
	})

	return nil
}

func (a AllocatedRange) UpdateAllocatedRange(allocatedRange entity.AllocatedRange) error {
	for _, currAllocatedRange := range a.allocatedRanges {
		if currAllocatedRange.Key == allocatedRange.Key {
			currAllocatedRange.RangeEnd = allocatedRange.RangeEnd
			currAllocatedRange.NextNumber = allocatedRange.NextNumber
		}
		return nil
	}

	return nil
}

func NewAllocatedRange(ranges []entity.AllocatedRange) AllocatedRange {
	copiedRanges := make([]entity.AllocatedRange, len(ranges))
	copy(copiedRanges, ranges)
	return AllocatedRange{
		allocatedRanges: copiedRanges,
	}
}
