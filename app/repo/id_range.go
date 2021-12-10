package repo

import (
	"github.com/teamyapp/one/entity"
)

type IDRange interface {
	GetAllocationEnd(resourceType string) (entity.ID, error)
	SetAllocationEnd(resourceType string, allocationEnd entity.ID) error
	ListAllResourceTypes() ([]string, error)
}

type InMemoryIDRange struct {
	allocationEnds map[string]entity.ID
}

var _ IDRange = (*InMemoryIDRange)(nil)

func (i InMemoryIDRange) GetAllocationEnd(resourceType string) (entity.ID, error) {
	allocationEnd, ok := i.allocationEnds[resourceType]
	if ok {
		return allocationEnd, nil
	}
	i.allocationEnds[resourceType] = -1
	return 0, nil
}

func (i InMemoryIDRange) SetAllocationEnd(resourceType string, allocationEnd entity.ID) error {
	i.allocationEnds[resourceType] = allocationEnd
	return nil
}

func (i InMemoryIDRange) ListAllResourceTypes() ([]string, error) {
	resourceTypes := make([]string, 0)
	for resourceType := range i.allocationEnds {
		resourceTypes = append(resourceTypes, resourceType)
	}
	return resourceTypes, nil
}

func NewInMemoryIDRange() *InMemoryIDRange {
	return &InMemoryIDRange{allocationEnds: make(map[string]entity.ID)}
}
