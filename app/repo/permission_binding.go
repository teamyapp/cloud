package repo

import "github.com/teamyapp/cloud/app/entity"

type PermissionBinding interface {
	HasPermissionBinding(query entity.PermissionBinding) bool
	AddPermissionBinding(permission entity.PermissionBinding)
}
