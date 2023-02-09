package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Resource interface {
	FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, *errs.Error)
	FindAllResources(ct context.Context) ([]entity.Resource, *errs.Error)
	CreateResource(ct context.Context, resource entity.Resource) *errs.Error
	DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error
}
