package dao

import "github.com/teamyapp/cloud/app/entity"

type Permission interface {
	FindPermission(permissionQuery entity.PermissionQuery) (bool, error)
}
