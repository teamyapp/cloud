package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type OperationRelation interface {
	FindOperationRelation(
		ct context.Context,
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) (entity.OperationRelation, *errs.Error)
	FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, *errs.Error)
	FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, *errs.Error)
	CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) *errs.Error
	DeleteOperationRelation(
		ct context.Context,
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) *errs.Error
}
