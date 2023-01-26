package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type ResourceRelation interface {
	FindResourceRelation(
		ct context.Context,
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) (entity.ResourceRelation, error)
	FindResourceRelations(ct context.Context, childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, error)
	FindAllResourceRelations(ct context.Context) ([]entity.ResourceRelation, error)
	CreateResourceRelation(ct context.Context, resourceRelation entity.ResourceRelation) error
	DeleteResourceRelation(
		ct context.Context,
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) error
}
