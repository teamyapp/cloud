package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type OperationRelation struct {
	operationRelations []entity.OperationRelation
}

var _ dao.OperationRelation = (*OperationRelation)(nil)

func (o OperationRelation) FindOperationRelation(ct context.Context, childResourceType string, childOperation string, parentResourceType string, parentOperation string) (entity.OperationRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (o OperationRelation) FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, error) {
	operationRelations := collect.Filter(o.operationRelations, func(operationRelation entity.OperationRelation) bool {
		return childResourceType == operationRelation.ChildResourceType && childOperation == operationRelation.ChildOperation
	})

	return operationRelations, nil
}

func (o OperationRelation) FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (o OperationRelation) CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) error {
	//TODO implement me
	panic("implement me")
}

func (o OperationRelation) DeleteOperationRelation(ct context.Context, childResourceType string, childOperation string, parentResourceType string, parentOperation string) error {
	//TODO implement me
	panic("implement me")
}

func NewOperationRelation(operationRelations []entity.OperationRelation) OperationRelation {
	copiedOperationRelations := make([]entity.OperationRelation, len(operationRelations))
	copy(copiedOperationRelations, operationRelations)
	return OperationRelation{
		operationRelations: copiedOperationRelations,
	}
}
