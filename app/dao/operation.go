package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Operation interface {
	FindAllOperations(ct context.Context) ([]entity.Operation, *errs.Error)
	FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, *errs.Error)
	FindOperationsByResourceType(ct context.Context, resourceTypeName string) ([]entity.Operation, *errs.Error)
	CreateOperation(ct context.Context, operation entity.Operation) *errs.Error
	DeleteOperation(ct context.Context, resourceTypeName string, operationName string) *errs.Error
}
