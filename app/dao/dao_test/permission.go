package dao_test

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
)

type Permission struct {
	permissions []entity.Permission
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(ct context.Context, permissionQuery entity.PermissionQuery) (entity.Permission, *errs.Error) {
	permissions := collect.Filter(p.permissions, func(permission entity.Permission) bool {
		return matchPermission(permissionQuery, permission)
	})

	if len(permissions) == 0 {
		return entity.Permission{}, &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("permission not found: query=%v", permissionQuery),
		}
	}

	return permissions[0], nil
}

func (p Permission) FindAllPermissions(ct context.Context) ([]entity.Permission, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (p Permission) CreatePermission(ct context.Context, permission entity.Permission) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (p Permission) DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewPermission(permissions []entity.Permission) Permission {
	copiedPermissions := make([]entity.Permission, len(permissions))
	copy(copiedPermissions, permissions)
	return Permission{
		permissions: copiedPermissions,
	}
}

func matchPermission(permissionQuery entity.PermissionQuery, permission entity.Permission) bool {
	return permissionQuery.ResourceID == permission.ResourceID &&
		permissionQuery.ResourceType == permission.ResourceType &&
		permissionQuery.Operation == permission.Operation &&
		permissionQuery.GroupID == permission.GroupID
}
