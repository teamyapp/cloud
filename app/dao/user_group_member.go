package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type UserGroupMember interface {
	FindGroupIDsByUserID(ct context.Context, userID uint64) ([]uint64, error)
	FindUserGroupMembersByUserID(ct context.Context, userID uint64) ([]entity.UserGroupMember, error)
	FindUserGroupMembersByGroupID(ct context.Context, groupID uint64) ([]entity.UserGroupMember, error)
	FindUserGroupMember(ct context.Context, groupID uint64, userID uint64) (entity.UserGroupMember, error)
	FindAllUserGroupMembers(ct context.Context) ([]entity.UserGroupMember, error)
	CreateUserGroupMember(ct context.Context, userGroupMember entity.UserGroupMember) error
	DeleteUserGroupMember(ct context.Context, groupID uint64, userID uint64) error
}
