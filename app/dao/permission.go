package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type Permission interface {
	FindPermission(query entity.PermissionQuery) (entity.Permission, error)
	FindAllPermissions() ([]entity.Permission, error)
	CreatePermission(permission entity.Permission) error
	DeletePermission(resourceType string, resourceID uint64, operation string, groupID uint64) error
}
