package dao_test

import (
	"context"
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type ResourceType struct {
	resourceTypes []entity.ResourceType
}

var _ dao.ResourceType = (*ResourceType)(nil)

func (r ResourceType) FindResourceType(ct context.Context, resourceType string) (entity.ResourceType, error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceType) FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceType) CreateResourceType(ct context.Context, resourceTypeEntity entity.ResourceType) error {
	//TODO implement me
	panic("implement me")
}

func (r ResourceType) DeleteResourceType(ct context.Context, resourceType string) error {
	//TODO implement me
	panic("implement me")
}

func NewResourceType(resourceTypes []entity.ResourceType) ResourceType {
	copiedResourceTypes := make([]entity.ResourceType, len(resourceTypes))
	copy(copiedResourceTypes, resourceTypes)
	return ResourceType{
		resourceTypes: copiedResourceTypes,
	}
}
