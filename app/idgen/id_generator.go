package idgen

import (
	"fmt"
	"log"
	"math"

	"github.com/teamyapp/cloud/app/repo"
	"github.com/teamyapp/one/entity"
)

const maxID = math.MaxInt32

type idRange struct {
	rangeEnd     entity.ID
	nextUniqueID entity.ID
}

func (i idRange) String() string {
	return fmt.Sprintf("[idRange rangeEnd=%v nextUniqueID=%v]", i.rangeEnd, i.nextUniqueID)
}

type IDGenerator struct {
	idRangeRepo    repo.IDRange
	rangeLength    int
	resourceType   string
	allocatedRange *idRange
}

func (i *IDGenerator) GenerateUniqueID() (entity.ID, error) {
	if i.allocatedRange.nextUniqueID > i.allocatedRange.rangeEnd {
		newRange, err := i.allocateIDRange()
		if err != nil {
			return entity.ID(-1), err
		}
		i.allocatedRange = &newRange
	}
	id := i.allocatedRange.nextUniqueID
	i.allocatedRange.nextUniqueID++
	return id, nil
}

func (i *IDGenerator) init() error {
	allocatedRange, err := i.allocateIDRange()
	if err != nil {
		log.Printf("cannot allocate ID range")
		return err
	}
	i.allocatedRange = &allocatedRange
	return nil
}

func (i IDGenerator) allocateIDRange() (idRange, error) {
	// TODO: partition based on resource type for distributed systems
	rangeEnd, err := i.idRangeRepo.GetAllocationEnd(i.resourceType)
	if err != nil {
		return idRange{}, err
	}
	if rangeEnd == maxID {
		return idRange{}, fmt.Errorf("out of ID to allocate")
	}
	newRangeStart := rangeEnd + 1
	newRangeEnd := min(int(rangeEnd)+i.rangeLength, maxID)
	err = i.idRangeRepo.SetAllocationEnd(i.resourceType, entity.ID(newRangeEnd))
	newRange := idRange{
		rangeEnd:     entity.ID(newRangeEnd),
		nextUniqueID: newRangeStart,
	}
	if err == nil {
		log.Printf("allocated ID range: %v", newRange)
	}
	return newRange, err
}

func min(num1 int, num2 int) int {
	if num1 <= num2 {
		return num1
	} else {
		return num2
	}
}

func newIDGenerator(idRangeRepo repo.IDRange, rangeLength int, resourceType string) (*IDGenerator, error) {
	idGen := IDGenerator{
		idRangeRepo:  idRangeRepo,
		rangeLength:  rangeLength,
		resourceType: resourceType,
	}
	err := idGen.init()
	if err != nil {
		return nil, err
	}
	return &idGen, nil
}
