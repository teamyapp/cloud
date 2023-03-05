package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Resource struct {
	db *InMemoryDB
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, *errs.Error) {
	table, err := r.db.GetTable(ResourceTableName)
	if err != nil {
		return entity.Resource{}, err
	}

	for _, rawRow := range table.rows {
		resource := rawRow.(entity.Resource)
		if resource.ResourceTypeName == resourceTypeName &&
			resource.ResourceID == resourceID {
			return resource, nil
		}
	}

	return entity.Resource{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: resourceTypeName=%v, resourceID=%v", resourceTypeName, resourceID),
	}
}

func (r Resource) FindAllResources(ct context.Context) ([]entity.Resource, *errs.Error) {
	table, err := r.db.GetTable(ResourceTableName)
	if err != nil {
		return nil, err
	}

	resources := make([]entity.Resource, 0)
	for _, rawRow := range table.rows {
		resource := rawRow.(entity.Resource)
		resources = append(resources, resource)
	}

	return resources, nil
}

func (r Resource) CreateResource(ct context.Context, resource entity.Resource) *errs.Error {
	_, err := r.FindResource(ct, resource.ResourceTypeName, resource.ResourceID)
	if err == nil {
		return &errs.Error{
			Code: errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: resourceTypeName=%v, resourceID=%v",
				resource.ResourceTypeName,
				resource.ResourceID),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := r.db.GetTable(ResourceTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, resource)
	return nil
}

func (r Resource) DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error {
	table, err := r.db.GetTable(ResourceTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		resource := rawRow.(entity.Resource)
		if resource.ResourceTypeName != resourceTypeName ||
			resource.ResourceID != resourceID {
			rows = append(rows, rawRow)
		}
	}

	table.rows = rows
	return nil
}

func NewResource(db *InMemoryDB) Resource {
	return Resource{
		db: db,
	}
}
