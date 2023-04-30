package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceUserGroupRelation struct {
	db *dbtest.InMemoryDB
}

var _ dao.ResourceUserGroupRelation = (*ResourceUserGroupRelation)(nil)

func (r ResourceUserGroupRelation) FindResourceUserGroupRelationByUserGroup(ct context.Context, userGroupID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceUserGroupRelationTableName)
	if err != nil {
		return nil, err
	}

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for _, rawRow := range table.Rows {
		resourceUserGroupRelation := rawRow.(entity.ResourceUserGroupRelation)
		if resourceUserGroupRelation.GroupID == userGroupID {
			resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
		}
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) FindResourceUserGroupRelationByResource(ct context.Context, resourceType string, resourceID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceUserGroupRelationTableName)
	if err != nil {
		return nil, err
	}

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for _, rawRow := range table.Rows {
		resourceUserGroupRelation := rawRow.(entity.ResourceUserGroupRelation)
		if resourceUserGroupRelation.ResourceType == resourceType && resourceUserGroupRelation.ResourceID == resourceID {
			resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
		}
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) FindAllResourceUserGroupRelations(ct context.Context) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	table, err := r.db.GetTable(ResourceUserGroupRelationTableName)
	if err != nil {
		return nil, err
	}

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for _, rawRow := range table.Rows {
		resourceUserGroupRelation := rawRow.(entity.ResourceUserGroupRelation)
		resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) CreateResourceUserGroupRelation(ct context.Context, relation entity.ResourceUserGroupRelation) *errs.Error {
	resourceUserGroupRelations, err := r.FindResourceUserGroupRelationByUserGroup(ct, relation.GroupID)
	for _, resourceUserGroupRelation := range resourceUserGroupRelations {
		if resourceUserGroupRelation.ResourceType == relation.ResourceType && resourceUserGroupRelation.ResourceID == relation.ResourceID {
			return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: relation=%v", relation))
		}
	}

	table, err := r.db.GetTable(ResourceUserGroupRelationTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, relation)
	return nil
}

func (r ResourceUserGroupRelation) DeleteResourceUserGroupRelation(ct context.Context, resourceType string, resourceID uint64, userGroupID uint64) *errs.Error {
	table, err := r.db.GetTable(ResourceUserGroupRelationTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		resourceUserGroupRelation := rawRow.(entity.ResourceUserGroupRelation)
		if resourceUserGroupRelation.ResourceType != resourceType ||
			resourceUserGroupRelation.ResourceID != resourceID ||
			resourceUserGroupRelation.GroupID != userGroupID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewResourceUserGroupRelation(db *dbtest.InMemoryDB) ResourceUserGroupRelation {
	return ResourceUserGroupRelation{
		db: db,
	}
}
