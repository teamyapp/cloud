package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type ResourceType interface {
	FindResourceType(ct context.Context, resourceType string) (entity.ResourceType, error)
	FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, error)
	CreateResourceType(ct context.Context, resourceTypeEntity entity.ResourceType) error
	DeleteResourceType(ct context.Context, resourceType string) error
}
