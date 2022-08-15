package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type UserGroupMember struct {
	userGroupMembers []entity.UserGroupMember
}

var _ dao.UserGroupMember = (*UserGroupMember)(nil)

func (u UserGroupMember) FindGroupIDsByUserID(userID uint64) ([]uint64, error) {
	groupMembers := collect.Filter(u.userGroupMembers, func(userGroupMember entity.UserGroupMember) bool {
		return userGroupMember.UserID == userID
	})

	return collect.Map(groupMembers, func(groupMember entity.UserGroupMember, _ int) uint64 {
		return groupMember.GroupID
	}), nil
}

func (u UserGroupMember) FindUserGroupMembersByUserID(userID uint64) ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindUserGroupMembersByGroupID(groupID uint64) ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindUserGroupMember(groupID uint64, userID uint64) (entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) FindAllUserGroupMembers() ([]entity.UserGroupMember, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) CreateUserGroupMember(userGroupMember entity.UserGroupMember) error {
	//TODO implement me
	panic("implement me")
}

func (u UserGroupMember) DeleteUserGroupMember(groupID uint64, userID uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewUserGroupMember(userGroupMembers []entity.UserGroupMember) UserGroupMember {
	return UserGroupMember{
		userGroupMembers: userGroupMembers,
	}
}
