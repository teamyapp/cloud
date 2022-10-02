package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type Resource interface {
	FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, error)
	FindAllResources(ct context.Context) ([]entity.Resource, error)
	CreateResource(ct context.Context, resource entity.Resource) error
	DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) error
}
