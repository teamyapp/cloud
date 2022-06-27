package dao

import "github.com/teamyapp/cloud/app/entity"

type ResourceRelation interface {
	FindParentResources(resource entity.ResourceRelation) ([]entity.ResourceRelation, error)
}
