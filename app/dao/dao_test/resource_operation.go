package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type ResourceOperation struct {
	resourceOperations []entity.ResourceOperation
}

var _ dao.ResourceOperation = (*ResourceOperation)(nil)

func (r ResourceOperation) GetAllParentResourceOperations(resourceOperation entity.ResourceOperation) ([]entity.ResourceOperation, error) {
	parentResourceOperations := make([]entity.ResourceOperation, 0)
	for _, resourceOperationEntry := range r.resourceOperations {
		if resourceOperationEntry.ResourceType == resourceOperation.ResourceType &&
			resourceOperationEntry.Operation == resourceOperation.Operation {
			parentResourceOperations = append(parentResourceOperations, entity.ResourceOperation{
				ResourceType: resourceOperationEntry.ParentResourceType,
				Operation:    resourceOperationEntry.ParentOperation,
			})
		}
	}

	return parentResourceOperations, nil
}

func NewResourceOperation(resourceOperations []entity.ResourceOperation) ResourceOperation {
	return ResourceOperation{
		resourceOperations: resourceOperations,
	}
}
