package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type Permission struct {
	db *dbtest.InMemoryDB
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(ct context.Context, query entity.PermissionQuery) (entity.Permission, *errs.Error) {
	table, err := p.db.GetTable(PermissionTableName)
	if err != nil {
		return entity.Permission{}, err
	}

	for _, rawRow := range table.Rows {
		permission := rawRow.(entity.Permission)
		if permission.ResourceType == query.ResourceType &&
			permission.ResourceID == query.ResourceID &&
			permission.Operation == query.Operation &&
			permission.GroupID == query.GroupID {
			return permission, nil
		}
	}

	return entity.Permission{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: query=%v", query),
	}
}

func (p Permission) FindAllPermissions(ct context.Context) ([]entity.Permission, *errs.Error) {
	table, err := p.db.GetTable(PermissionTableName)
	if err != nil {
		return nil, err
	}

	permissions := make([]entity.Permission, 0)
	for _, rawRow := range table.Rows {
		permission := rawRow.(entity.Permission)
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (p Permission) CreatePermission(ct context.Context, permission entity.Permission) *errs.Error {
	query := entity.PermissionQuery{
		ResourceType: permission.ResourceType,
		ResourceID:   permission.ResourceID,
		Operation:    permission.Operation,
		GroupID:      permission.GroupID,
	}
	_, err := p.FindPermission(ct, query)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: query=%v", query),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := p.db.GetTable(PermissionTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, permission)
	return nil
}

func (p Permission) DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error {
	table, err := p.db.GetTable(PermissionTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		permission := rawRow.(entity.Permission)
		if permission.ResourceType != resourceType ||
			permission.ResourceID != resourceID ||
			permission.Operation != operation ||
			permission.GroupID != groupID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewPermission(db *dbtest.InMemoryDB) Permission {
	return Permission{
		db: db,
	}
}
