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

func (r ResourceRelation) FindResourceRelation(
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) (entity.ResourceRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) FindResourceRelations(childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, error) {
	resourceRelations := collect.Filter(r.resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		return childResourceID == resourceRelation.ChildResourceID && childResourceType == resourceRelation.ChildResourceType
	})

	return resourceRelations, nil
}

func (r ResourceRelation) FindAllResourceRelations() ([]entity.ResourceRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) CreateResourceRelation(resourceRelation entity.ResourceRelation) error {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) DeleteResourceRelation(childResourceType string, childResourceID uint64, parentResourceType string, parentResourceID uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewResourceRelation(resourceRelations []entity.ResourceRelation) ResourceRelation {
	return ResourceRelation{
		resourceRelations: resourceRelations,
	}
}
