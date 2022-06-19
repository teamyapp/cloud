package dao

import "github.com/teamyapp/cloud/app/entity"

type ResourceOperation interface {
	GetAllParentResourceOperations(resourceOperation entity.ResourceOperation) ([]entity.ResourceOperation, error)
}
