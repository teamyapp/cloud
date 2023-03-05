package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Operation struct {
	db *InMemoryDB
}

var _ dao.Operation = (*Operation)(nil)

func (o Operation) FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, *errs.Error) {
	table, err := o.db.GetTable(OperationTableName)
	if err != nil {
		return entity.Operation{}, err
	}

	for _, rawRow := range table.rows {
		operation := rawRow.(entity.Operation)
		if operation.ResourceTypeName == resourceTypeName &&
			operation.OperationName == operationName {
			return operation, nil
		}
	}

	return entity.Operation{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: resourceTypeName=%v, operationName=%v", resourceTypeName, operationName),
	}
}

func (o Operation) FindAllOperations(ct context.Context) ([]entity.Operation, *errs.Error) {
	table, err := o.db.GetTable(OperationTableName)
	if err != nil {
		return nil, err
	}

	operations := make([]entity.Operation, 0)
	for _, rawRow := range table.rows {
		operation := rawRow.(entity.Operation)
		operations = append(operations, operation)
	}

	return operations, nil
}

func (o Operation) CreateOperation(ct context.Context, operation entity.Operation) *errs.Error {
	_, err := o.FindOperation(ct, operation.ResourceTypeName, operation.OperationName)
	if err == nil {
		return &errs.Error{
			Code: errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: resourceTypeName=%v, operation=%v",
				operation.ResourceTypeName,
				operation.OperationName),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := o.db.GetTable(OperationTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, operation)
	return nil
}

func (o Operation) DeleteOperation(ct context.Context, resourceTypeName string, operationName string) *errs.Error {
	table, err := o.db.GetTable(OperationTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		operation := rawRow.(entity.Operation)
		if operation.ResourceTypeName != resourceTypeName ||
			operation.OperationName != operationName {
			rows = append(rows, rawRow)
		}
	}

	table.rows = rows
	return nil
}

func NewOperation(db *InMemoryDB) Operation {
	return Operation{
		db: db,
	}
}
