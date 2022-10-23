package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UserGroup struct {
	userGroups []entity.UserGroup
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(groupID uint64) (entity.UserGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) FindAllGroups() ([]entity.UserGroup, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) CreateGroup(group entity.UserGroup) error {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) UpdateGroup(group entity.UserGroup) error {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) DeleteGroup(groupID uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewUserGroup(userGroups []entity.UserGroup) UserGroup {
	copiedUserGroups := make([]entity.UserGroup, len(userGroups))
	copy(copiedUserGroups, userGroups)
	return UserGroup{
		userGroups: copiedUserGroups,
	}
}
