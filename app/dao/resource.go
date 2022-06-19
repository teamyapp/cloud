package dao

import "github.com/teamyapp/cloud/app/entity"

type Resource interface {
	FindParentResources(resource entity.Resource) ([]entity.Resource, error)
}
