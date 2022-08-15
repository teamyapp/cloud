package dao

import "github.com/teamyapp/cloud/app/entity"

type UserGroupMember interface {
	FindGroupIDsByUserID(userID uint64) ([]uint64, error)
	FindUserGroupMembersByUserID(userID uint64) ([]entity.UserGroupMember, error)
	FindUserGroupMembersByGroupID(groupID uint64) ([]entity.UserGroupMember, error)
	FindUserGroupMember(groupID uint64, userID uint64) (entity.UserGroupMember, error)
	FindAllUserGroupMembers() ([]entity.UserGroupMember, error)
	CreateUserGroupMember(userGroupMember entity.UserGroupMember) error
	DeleteUserGroupMember(groupID uint64, userID uint64) error
}
