package idgen

import (
	"github.com/teamyapp/cloud/app/repo"
)

type Factory interface {
	NewIDGenerator(resourceType string) (*IDGenerator, error)
}

type InMemoryIDGeneratorFactory struct {
	idRangeRepo repo.IDRange
	rangeLength int
}

var _ Factory = (*InMemoryIDGeneratorFactory)(nil)

func (i InMemoryIDGeneratorFactory) NewIDGenerator(resourceType string) (*IDGenerator, error) {
	return newIDGenerator(i.idRangeRepo, i.rangeLength, resourceType)
}

func NewInMemoryIDGeneratorFactory(idRangeRepo repo.IDRange, rangeLength int) InMemoryIDGeneratorFactory {
	return InMemoryIDGeneratorFactory{
		idRangeRepo: idRangeRepo,
		rangeLength: rangeLength,
	}
}
