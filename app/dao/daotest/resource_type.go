package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceType struct {
	db *dbtest.InMemoryDB
}

var _ dao.ResourceType = (*ResourceType)(nil)

func (r ResourceType) FindResourceType(ct context.Context, resourceTypeName string) (entity.ResourceType, *errs.Error) {
	table, err := r.db.GetTable(ResourceTypeTableName)
	if err != nil {
		return entity.ResourceType{}, err
	}

	for _, rawRow := range table.Rows {
		resourceType := rawRow.(entity.ResourceType)
		if resourceType.ResourceTypeName == resourceTypeName {
			return resourceType, nil
		}
	}

	return entity.ResourceType{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: resourceTypeName=%v", resourceTypeName),
	}
}

func (r ResourceType) FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, *errs.Error) {
	table, err := r.db.GetTable(ResourceTypeTableName)
	if err != nil {
		return nil, err
	}

	resourceTypes := make([]entity.ResourceType, 0)
	for _, rawRow := range table.Rows {
		resourceType := rawRow.(entity.ResourceType)
		resourceTypes = append(resourceTypes, resourceType)
	}

	return resourceTypes, nil
}

func (r ResourceType) CreateResourceType(ct context.Context, resourceType entity.ResourceType) *errs.Error {
	_, err := r.FindResourceType(ct, resourceType.ResourceTypeName)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: resourceType=%v", resourceType),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := r.db.GetTable(ResourceTypeTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, resourceType)
	return nil
}

func (r ResourceType) DeleteResourceType(ct context.Context, resourceTypeName string) *errs.Error {
	table, err := r.db.GetTable(ResourceTypeTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		resourceType := rawRow.(entity.ResourceType)
		if resourceType.ResourceTypeName != resourceTypeName {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewResourceType(db *dbtest.InMemoryDB) ResourceType {
	return ResourceType{
		db: db,
	}
}
