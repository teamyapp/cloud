package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type OperationRelation struct {
	db *InMemoryDB
}

var _ dao.OperationRelation = (*OperationRelation)(nil)

func (o OperationRelation) FindOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) (entity.OperationRelation, *errs.Error) {
	table, err := o.db.GetTable(OperationRelationTableName)
	if err != nil {
		return entity.OperationRelation{}, err
	}

	for _, rawRow := range table.rows {
		operationRelation := rawRow.(entity.OperationRelation)
		if operationRelation.ChildResourceType == childResourceType &&
			operationRelation.ChildOperation == childOperation &&
			operationRelation.ParentResourceType == parentResourceType &&
			operationRelation.ParentOperation == parentOperation {
			return operationRelation, nil
		}
	}

	return entity.OperationRelation{}, &errs.Error{
		Code: errs.NotFound,
		Message: fmt.Sprintf("row not found: childResourceType=%v, childOperation=%v, parentResourceType=%v, parentOperation=%v",
			childResourceType,
			childOperation,
			parentResourceType,
			parentOperation),
	}
}

func (o OperationRelation) FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, *errs.Error) {
	table, err := o.db.GetTable(OperationRelationTableName)
	if err != nil {
		return nil, err
	}

	operationRelations := make([]entity.OperationRelation, 0)
	for _, rawRow := range table.rows {
		operationRelation := rawRow.(entity.OperationRelation)
		if operationRelation.ChildResourceType == childResourceType &&
			operationRelation.ChildOperation == childOperation {
			operationRelations = append(operationRelations, operationRelation)
		}
	}

	return operationRelations, nil
}

func (o OperationRelation) FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, *errs.Error) {
	table, err := o.db.GetTable(OperationRelationTableName)
	if err != nil {
		return nil, err
	}

	operationRelations := make([]entity.OperationRelation, 0)
	for _, rawRow := range table.rows {
		operationRelation := rawRow.(entity.OperationRelation)
		operationRelations = append(operationRelations, operationRelation)
	}

	return operationRelations, nil
}

func (o OperationRelation) CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) *errs.Error {
	_, err := o.FindOperationRelation(
		ct,
		operationRelation.ChildResourceType,
		operationRelation.ChildOperation,
		operationRelation.ParentResourceType,
		operationRelation.ParentOperation)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: operationRelation=%v", operationRelation),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := o.db.GetTable(OperationRelationTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, operationRelation)
	return nil
}

func (o OperationRelation) DeleteOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) *errs.Error {
	table, err := o.db.GetTable(OperationRelationTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		operationRelation := rawRow.(entity.OperationRelation)
		if operationRelation.ChildResourceType != childResourceType ||
			operationRelation.ChildOperation != childOperation ||
			operationRelation.ParentResourceType != parentResourceType ||
			operationRelation.ParentOperation != parentOperation {
			rows = append(rows, rawRow)
		}
	}

	table.rows = rows
	return nil
}

func NewOperationRelation(db *InMemoryDB) OperationRelation {
	return OperationRelation{
		db: db,
	}
}
