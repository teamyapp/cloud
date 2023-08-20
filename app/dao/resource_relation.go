package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceRelation interface {
	FindResourceRelation(
		ct context.Context,
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) (entity.ResourceRelation, *errs.Error)
	FindResourceRelations(ct context.Context, childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, *errs.Error)
	FindAllResourceRelations(ct context.Context) ([]entity.ResourceRelation, *errs.Error)
	CreateResourceRelation(ct context.Context, resourceRelation entity.ResourceRelation) *errs.Error
	DeleteResourceRelation(
		ct context.Context,
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) *errs.Error
}
