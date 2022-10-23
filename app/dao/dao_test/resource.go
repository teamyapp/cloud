package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Resource struct {
	resources []entity.Resource
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(resourceTypeName string, resourceID uint64) (entity.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (r Resource) FindAllResources() ([]entity.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (r Resource) CreateResource(resource entity.Resource) error {
	//TODO implement me
	panic("implement me")
}

func (r Resource) DeleteResource(resourceTypeName string, resourceID uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewResource(resources []entity.Resource) Resource {
	copiedResources := make([]entity.Resource, len(resources))
	copy(copiedResources, resources)
	return Resource{
		resources: copiedResources,
	}
}
