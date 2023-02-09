package dao_test

import (
	"context"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserGroup struct {
	userGroups []entity.UserGroup
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) FindAllGroups(ct context.Context) ([]entity.UserGroup, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) CreateGroup(ct context.Context, group entity.UserGroup) (entity.UserGroup, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) UpdateGroup(ct context.Context, group entity.UserGroup) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u UserGroup) DeleteGroup(ct context.Context, groupID uint64) *errs.Error {
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
