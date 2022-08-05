package dao

import "github.com/teamyapp/cloud/app/entity"

type Resource interface {
	FindResource(resourceType string, resourceID uint64) (entity.Resource, error)
	FindAllResources() ([]entity.Resource, error)
	CreateResource(resource entity.Resource) error
	DeleteResource(resourceType string, resourceID uint64) error
}
