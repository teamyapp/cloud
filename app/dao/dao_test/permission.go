package dao_test

import (
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type Permission struct {
	permissions []entity.Permission
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(permissionQuery entity.PermissionQuery) (entity.Permission, error) {
	permissions := collect.Filter(p.permissions, func(permission entity.Permission) bool {
		return matchPermission(permissionQuery, permission)
	})

	if len(permissions) == 0 {
		return entity.Permission{}, dao.ErrNotFound(fmt.Sprintf(
			"permission not found: id=%v",
			permissionQuery))
	}

	return permissions[0], nil
}

func NewPermission(permissions []entity.Permission) Permission {
	return Permission{
		permissions: permissions,
	}
}

func matchPermission(permissionQuery entity.PermissionQuery, permission entity.Permission) bool {
	return permissionQuery.ResourceID == permission.ResourceID &&
		permissionQuery.ResourceType == permission.ResourceType &&
		permissionQuery.Operation == permission.Operation &&
		permissionQuery.GroupID == permission.GroupID
}
