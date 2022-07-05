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

func (r ResourceRelation) FindResourceRelations(childResourceID uint64, childResourceType string) ([]entity.ResourceRelation, error) {
	resourceRelations := collect.Filter(r.resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		return childResourceID == resourceRelation.ChileResourceID && childResourceType == resourceRelation.ChildResourceType
	})

	return resourceRelations, nil
}

func NewResourceRelation(resourceRelations []entity.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		resourceRelations: resourceRelations,
	}
}
