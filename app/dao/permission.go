package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type Permission interface {
	FindPermission(ct context.Context, query entity.PermissionQuery) (entity.Permission, error)
	FindAllPermissions(ct context.Context) ([]entity.Permission, error)
	CreatePermission(ct context.Context, permission entity.Permission) error
	DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) error
}
