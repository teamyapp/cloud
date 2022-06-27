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

func (o OperationRelation) FindParentOperations(operationQuery entity.OperationRelation) ([]entity.OperationRelation, error) {
	operationRelations := collect.Filter(o.operationRelations, func(operationRelation entity.OperationRelation) bool {
		return operationQuery.ResourceType == operationRelation.ResourceType && operationQuery.Operation == operationRelation.Operation
	})

	parentOperations := collect.Map(operationRelations, func(operationRelation entity.OperationRelation, _ int) entity.OperationRelation {
		return entity.OperationRelation{
			ResourceType: operationRelation.ParentResourceType,
			Operation:    operationRelation.ParentOperation,
		}
	})

	return parentOperations, nil
}

func NewOperationRelation(operationRelations []entity.OperationRelation) OperationRelation {
	return OperationRelation{
		operationRelations: operationRelations,
	}
}
