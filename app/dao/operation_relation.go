package dao

import "github.com/teamyapp/cloud/app/entity"

type OperationRelation interface {
	FindParentOperations(resourceOperation entity.OperationRelation) ([]entity.OperationRelation, error)
}
