package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Resource struct {
	resources []entity.Resource
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindParentResources(resource entity.Resource) ([]entity.Resource, error) {
	parentResources := make([]entity.Resource, 0)
	for _, resourceEntry := range r.resources {
		if resourceEntry.ID == resource.ID &&
			resourceEntry.ResourceType == resource.ResourceType {
			parentResources = append(parentResources, entity.Resource{
				ID:           resourceEntry.ParentResourceID,
				ResourceType: resourceEntry.ParentResourceType,
			})
		}
	}

	return parentResources, nil
}

func NewResource(resources []entity.Resource) Resource {
	return Resource{
		resources: resources,
	}
}
