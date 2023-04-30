package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceUserGroupRelation interface {
	FindResourceUserGroupRelationByUserGroup(ct context.Context,
		userGroupID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error)
	FindResourceUserGroupRelationByResource(ct context.Context, resourceType string,
		resourceID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error)
	FindAllResourceUserGroupRelations(ct context.Context) ([]entity.ResourceUserGroupRelation, *errs.Error)
	CreateResourceUserGroupRelation(ct context.Context, relation entity.ResourceUserGroupRelation) *errs.Error
	DeleteResourceUserGroupRelation(ct context.Context, resourceType string, resourceID uint64,
		userGroupID uint64) *errs.Error
}
