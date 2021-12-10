package repo

import "github.com/teamyapp/cloud/app/entity"

type PermissionGraph interface {
	GetNeighbours(permission entity.Permission) []entity.Permission
	AddNeighbour(node entity.Permission, neighbour entity.Permission)
}
