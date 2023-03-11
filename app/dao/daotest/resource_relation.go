package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceRelation struct {
	db *dbtest.InMemoryDB
}

var _ dao.ResourceRelation = (*ResourceRelation)(nil)

func (r ResourceRelation) FindResourceRelation(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) (entity.ResourceRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceRelationTableName)
	if err != nil {
		return entity.ResourceRelation{}, err
	}

	for _, rawRow := range table.Rows {
		resourceRelation := rawRow.(entity.ResourceRelation)
		if resourceRelation.ChildResourceType == childResourceType &&
			resourceRelation.ChildResourceID == childResourceID &&
			resourceRelation.ParentResourceType == parentResourceType &&
			resourceRelation.ParentResourceID == parentResourceID {
			return resourceRelation, nil
		}
	}

	return entity.ResourceRelation{}, &errs.Error{
		Code: errs.NotFound,
		Message: fmt.Sprintf("row not found: childResourceType=%v, childResourceID=%v, parentResourceType=%v, parentResourceID=%v",
			childResourceType,
			childResourceID,
			parentResourceType,
			parentResourceID),
	}
}

func (r ResourceRelation) FindResourceRelations(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
) ([]entity.ResourceRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceRelationTableName)
	if err != nil {
		return nil, err
	}

	resourceRelations := make([]entity.ResourceRelation, 0)
	for _, rawRow := range table.Rows {
		resourceRelation := rawRow.(entity.ResourceRelation)
		if resourceRelation.ChildResourceType == childResourceType &&
			resourceRelation.ChildResourceID == childResourceID {
			resourceRelations = append(resourceRelations, resourceRelation)
		}
	}

	return resourceRelations, nil
}

func (r ResourceRelation) FindAllResourceRelations(ct context.Context) ([]entity.ResourceRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceRelationTableName)
	if err != nil {
		return nil, err
	}

	resourceRelations := make([]entity.ResourceRelation, 0)
	for _, rawRow := range table.Rows {
		resourceRelation := rawRow.(entity.ResourceRelation)
		resourceRelations = append(resourceRelations, resourceRelation)
	}

	return resourceRelations, nil
}

func (r ResourceRelation) CreateResourceRelation(ct context.Context, resourceRelation entity.ResourceRelation) *errs.Error {
	_, err := r.FindResourceRelation(
		ct,
		resourceRelation.ChildResourceType,
		resourceRelation.ChildResourceID,
		resourceRelation.ParentResourceType,
		resourceRelation.ParentResourceID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: resourceRelation=%v", resourceRelation),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := r.db.GetTable(ResourceRelationTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, resourceRelation)
	return nil
}

func (r ResourceRelation) DeleteResourceRelation(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) *errs.Error {
	table, err := r.db.GetTable(ResourceRelationTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		resourceRelation := rawRow.(entity.ResourceRelation)
		if resourceRelation.ChildResourceType != childResourceType ||
			resourceRelation.ChildResourceID != childResourceID ||
			resourceRelation.ParentResourceType != parentResourceType ||
			resourceRelation.ParentResourceID != parentResourceID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewResourceRelation(db *dbtest.InMemoryDB) ResourceRelation {
	return ResourceRelation{
		db: db,
	}
}
