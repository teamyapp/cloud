package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Permission struct {
	permissions []entity.Permission
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(permissionQuery entity.PermissionQuery) (bool, error) {
	for _, permissionEntry := range p.permissions {
		if permissionEntry.ResourceType == permissionQuery.ResourceType &&
			permissionEntry.ResourceID == permissionQuery.ResourceID &&
			permissionEntry.Operation == permissionQuery.Operation &&
			permissionEntry.GroupID == permissionQuery.GroupID {
			return true, nil
		}
	}
	return false, nil
}

func NewPermission(permissions []entity.Permission) Permission {
	return Permission{
		permissions: permissions,
	}
}
