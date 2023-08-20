package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserGroupMember interface {
	FindGroupIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error)
	FindUserGroupMembersByUserID(ct context.Context, userID uint64) ([]entity.UserGroupMember, *errs.Error)
	FindUserGroupMembersByGroupID(ct context.Context, groupID uint64) ([]entity.UserGroupMember, *errs.Error)
	FindUserGroupMember(ct context.Context, groupID uint64, userID uint64) (entity.UserGroupMember, *errs.Error)
	FindAllUserGroupMembers(ct context.Context) ([]entity.UserGroupMember, *errs.Error)
	CreateUserGroupMember(ct context.Context, userGroupMember entity.UserGroupMember) *errs.Error
	DeleteUserGroupMember(ct context.Context, groupID uint64, userID uint64) *errs.Error
}
