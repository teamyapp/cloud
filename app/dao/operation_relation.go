package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type OperationRelation interface {
	FindOperationRelation(
		ct context.Context,
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) (entity.OperationRelation, error)
	FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, error)
	FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, error)
	CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) error
	DeleteOperationRelation(
		ct context.Context,
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) error
}
