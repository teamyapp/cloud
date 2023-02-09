package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceRelation struct {
	resourceRelations []entity.ResourceRelation
}

var _ dao.ResourceRelation = (*ResourceRelation)(nil)

func (r ResourceRelation) FindResourceRelation(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) (entity.ResourceRelation, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) FindResourceRelations(ct context.Context, childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, *errs.Error) {
	resourceRelations := collect.Filter(r.resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		return childResourceID == resourceRelation.ChildResourceID && childResourceType == resourceRelation.ChildResourceType
	})

	return resourceRelations, nil
}

func (r ResourceRelation) FindAllResourceRelations(ct context.Context) ([]entity.ResourceRelation, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) CreateResourceRelation(ct context.Context, resourceRelation entity.ResourceRelation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (r ResourceRelation) DeleteResourceRelation(ct context.Context, childResourceType string, childResourceID uint64, parentResourceType string, parentResourceID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewResourceRelation(resourceRelations []entity.ResourceRelation) ResourceRelation {
	copiedResourceRelations := make([]entity.ResourceRelation, len(resourceRelations))
	copy(copiedResourceRelations, resourceRelations)
	return ResourceRelation{
		resourceRelations: copiedResourceRelations,
	}
}
