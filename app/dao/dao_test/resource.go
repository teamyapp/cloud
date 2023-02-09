package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Resource struct {
	resources []entity.Resource
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (r Resource) FindAllResources(ct context.Context) ([]entity.Resource, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (r Resource) CreateResource(ct context.Context, resource entity.Resource) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (r Resource) DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error {
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
