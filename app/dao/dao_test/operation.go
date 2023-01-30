package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Operation struct {
	operations []entity.Operation
}

var _ dao.Operation = (*Operation)(nil)

func (o Operation) FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (o Operation) FindAllOperations(ct context.Context) ([]entity.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (o Operation) CreateOperation(ct context.Context, operation entity.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (o Operation) DeleteOperation(ct context.Context, resourceTypeName string, operationName string) error {
	//TODO implement me
	panic("implement me")
}

func NewOperation(operations []entity.Operation) Operation {
	copiedOperations := make([]entity.Operation, len(operations))
	copy(copiedOperations, operations)
	return Operation{
		operations: copiedOperations,
	}
}
