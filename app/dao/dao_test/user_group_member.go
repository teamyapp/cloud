package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type UserGroupMember struct {
	userGroupMembers []entity.UserGroupMember
}

var _ dao.UserGroupMember = (*UserGroupMember)(nil)

func (u UserGroupMember) FindGroupIDsByUserID(ct context.Context, userID uint64) ([]uint64, error) {
	groupMembers := collect.Filter(u.userGroupMembers, func(userGroupMember entity.UserGroupMember) bool {
		return userGroupMember.UserID == userID
	})

	return collect.Map(groupMembers, func(groupMember entity.UserGroupMember, _ int) uint64 {
		return groupMember.GroupID
	}), nil
}

func (u UserGroupMember) FindUserGroupMembersByUserID(ct context.Context, userID uint64) ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindUserGroupMembersByGroupID(ct context.Context, groupID uint64) ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindUserGroupMember(ct context.Context, groupID uint64, userID uint64) (entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindAllUserGroupMembers(ct context.Context) ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) CreateUserGroupMember(ct context.Context, userGroupMember entity.UserGroupMember) error {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) DeleteUserGroupMember(ct context.Context, groupID uint64, userID uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewUserGroupMember(userGroupMembers []entity.UserGroupMember) UserGroupMember {
	copiedUserGroupMembers := make([]entity.UserGroupMember, len(userGroupMembers))
	copy(copiedUserGroupMembers, userGroupMembers)
	return UserGroupMember{
		userGroupMembers: copiedUserGroupMembers,
	}
}
