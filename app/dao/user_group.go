package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserGroup interface {
	FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, *errs.Error)
	FindAllGroups(ct context.Context) ([]entity.UserGroup, *errs.Error)
	CreateGroup(ct context.Context, group entity.UserGroup) (entity.UserGroup, *errs.Error)
	UpdateGroup(ct context.Context, group entity.UserGroup) *errs.Error
	DeleteGroup(ct context.Context, groupID uint64) *errs.Error
}
