package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type UserGroup interface {
	FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, error)
	FindAllGroups(ct context.Context) ([]entity.UserGroup, error)
	CreateGroup(ct context.Context, group entity.UserGroup) (entity.UserGroup, error)
	UpdateGroup(ct context.Context, group entity.UserGroup) error
	DeleteGroup(ct context.Context, groupID uint64) error
}
