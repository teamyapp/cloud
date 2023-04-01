package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserGroupMember struct {
	db *dbtest.InMemoryDB
}

var _ dao.UserGroupMember = (*UserGroupMember)(nil)

func (u UserGroupMember) FindGroupIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return nil, err
	}

	userGroupIDs := make([]uint64, 0)
	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		if userGroupMember.UserID == userID {
			userGroupIDs = append(userGroupIDs, userGroupMember.GroupID)
		}
	}

	return userGroupIDs, nil
}

func (u UserGroupMember) FindUserGroupMembersByUserID(ct context.Context, userID uint64) ([]entity.UserGroupMember, *errs.Error) {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return nil, err
	}

	userGroupMembers := make([]entity.UserGroupMember, 0)
	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		if userGroupMember.UserID == userID {
			userGroupMembers = append(userGroupMembers, userGroupMember)
		}
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) FindUserGroupMembersByGroupID(ct context.Context, groupID uint64) ([]entity.UserGroupMember, *errs.Error) {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return nil, err
	}

	userGroupMembers := make([]entity.UserGroupMember, 0)
	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		if userGroupMember.GroupID == groupID {
			userGroupMembers = append(userGroupMembers, userGroupMember)
		}
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) FindUserGroupMember(ct context.Context, groupID uint64, userID uint64) (entity.UserGroupMember, *errs.Error) {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return entity.UserGroupMember{}, err
	}

	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		if userGroupMember.GroupID == groupID &&
			userGroupMember.UserID == userID {
			return userGroupMember, nil
		}
	}

	return entity.UserGroupMember{}, errs.NewError(
		errs.NotFound,
		fmt.Sprintf("row not found: groupID=%v, userID=%v", groupID, userID))
}

func (u UserGroupMember) FindAllUserGroupMembers(ct context.Context) ([]entity.UserGroupMember, *errs.Error) {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return nil, err
	}

	userGroupMembers := make([]entity.UserGroupMember, 0)
	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) CreateUserGroupMember(ct context.Context, userGroupMember entity.UserGroupMember) *errs.Error {
	_, err := u.FindUserGroupMember(ct, userGroupMember.GroupID, userGroupMember.UserID)
	if err == nil {
		return errs.NewError(
			errs.Unknown,
			fmt.Sprintf("row already exist: userGroupMember=%v", userGroupMember))
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, userGroupMember)
	return nil
}

func (u UserGroupMember) DeleteUserGroupMember(ct context.Context, groupID uint64, userID uint64) *errs.Error {
	table, err := u.db.GetTable(UserGroupMemberTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		userGroupMember := rawRow.(entity.UserGroupMember)
		if userGroupMember.GroupID != groupID ||
			userGroupMember.UserID != userID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewUserGroupMember(db *dbtest.InMemoryDB) UserGroupMember {
	return UserGroupMember{
		db: db,
	}
}
