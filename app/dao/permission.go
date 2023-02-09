package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Permission interface {
	FindPermission(ct context.Context, query entity.PermissionQuery) (entity.Permission, *errs.Error)
	FindAllPermissions(ct context.Context) ([]entity.Permission, *errs.Error)
	CreatePermission(ct context.Context, permission entity.Permission) *errs.Error
	DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error
}
