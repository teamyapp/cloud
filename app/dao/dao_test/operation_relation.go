package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type OperationRelation struct {
	operationRelations []entity.OperationRelation
}

var _ dao.OperationRelation = (*OperationRelation)(nil)

func (o OperationRelation) FindOperationRelations(childResourceType string, childOperation string) ([]entity.OperationRelation, error) {
	operationRelations := collect.Filter(o.operationRelations, func(operationRelation entity.OperationRelation) bool {
		return childResourceType == operationRelation.ChildResourceType && childOperation == operationRelation.ChildOperation
	})

	return operationRelations, nil
}

func NewOperationRelation(operationRelations []entity.OperationRelation) OperationRelation {
	return OperationRelation{
		operationRelations: operationRelations,
	}
}
