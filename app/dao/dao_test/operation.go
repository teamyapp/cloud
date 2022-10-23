package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Operation struct {
	operations []entity.Operation
}

var _ dao.Operation = (*Operation)(nil)

func (o Operation) FindOperation(resourceTypeName string, operationName string) (entity.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (o Operation) FindAllOperations() ([]entity.Operation, error) {
	//TODO implement me
	panic("implement me")
}

func (o Operation) CreateOperation(operation entity.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (o Operation) DeleteOperation(resourceTypeName string, operationName string) error {
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
