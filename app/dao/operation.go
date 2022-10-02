package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type Operation interface {
	FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, error)
	FindAllOperations(ct context.Context) ([]entity.Operation, error)
	CreateOperation(ct context.Context, operation entity.Operation) error
	DeleteOperation(ct context.Context, resourceTypeName string, operationName string) error
}
