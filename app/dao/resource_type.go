package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceType interface {
	FindResourceType(ct context.Context, resourceType string) (entity.ResourceType, *errs.Error)
	FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, *errs.Error)
	CreateResourceType(ct context.Context, resourceTypeEntity entity.ResourceType) *errs.Error
	DeleteResourceType(ct context.Context, resourceType string) *errs.Error
}
