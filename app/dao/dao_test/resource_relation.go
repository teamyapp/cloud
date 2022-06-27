package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type ResourceRelation struct {
	resourceRelations []entity.ResourceRelation
}

var _ dao.ResourceRelation = (*ResourceRelation)(nil)

func (r ResourceRelation) FindParentResources(resourceQuery entity.ResourceRelation) ([]entity.ResourceRelation, error) {
	resourceRelations := collect.Filter(r.resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		return resourceQuery.ID == resourceRelation.ID && resourceQuery.ResourceType == resourceRelation.ResourceType
	})

	parentResources := collect.Map(resourceRelations, func(resourceRelation entity.ResourceRelation, _ int) entity.ResourceRelation {
		return entity.ResourceRelation{
			ID:           resourceRelation.ParentResourceID,
			ResourceType: resourceRelation.ParentResourceType,
		}
	})

	return parentResources, nil
}

func NewResourceRelation(resourceRelations []entity.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		resourceRelations: resourceRelations,
	}
}
