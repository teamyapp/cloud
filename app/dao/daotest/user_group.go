package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserGroup struct {
	db *dbtest.InMemoryDB
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, *errs.Error) {
	table, err := u.db.GetTable(UserGroupTableName)
	if err != nil {
		return entity.UserGroup{}, err
	}

	for _, rawRow := range table.Rows {
		userGroup := rawRow.(entity.UserGroup)
		if userGroup.ID == groupID {
			return userGroup, nil
		}
	}

	return entity.UserGroup{}, &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: groupID=%v", groupID),
	}
}

func (u UserGroup) FindAllGroups(ct context.Context) ([]entity.UserGroup, *errs.Error) {
	table, err := u.db.GetTable(UserGroupTableName)
	if err != nil {
		return nil, err
	}

	userGroups := make([]entity.UserGroup, 0)
	for _, rawRow := range table.Rows {
		userGroup := rawRow.(entity.UserGroup)
		userGroups = append(userGroups, userGroup)
	}

	return userGroups, nil
}

func (u UserGroup) CreateGroup(ct context.Context, group entity.UserGroup) *errs.Error {
	_, err := u.FindGroupByID(ct, group.ID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: group=%v", group),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := u.db.GetTable(UserGroupTableName)
	if err != nil {
		return err
	}

	table.Rows = append(table.Rows, group)
	return nil
}

func (u UserGroup) UpdateGroup(ct context.Context, group entity.UserGroup) *errs.Error {
	table, err := u.db.GetTable(UserGroupTableName)
	if err != nil {
		return err
	}

	var updated bool
	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		userGroup := rawRow.(entity.UserGroup)
		if userGroup.ID == group.ID {
			rows = append(rows, group)
			updated = true
		} else {
			rows = append(rows, rawRow)
		}
	}

	if updated {
		table.Rows = rows
		return nil
	}

	return &errs.Error{
		Code:    errs.NotFound,
		Message: fmt.Sprintf("row not found: id=%v", group.ID),
	}
}

func (u UserGroup) DeleteGroup(ct context.Context, groupID uint64) *errs.Error {
	table, err := u.db.GetTable(UserGroupTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.Rows {
		userGroup := rawRow.(entity.UserGroup)
		if userGroup.ID != groupID {
			rows = append(rows, rawRow)
		}
	}

	table.Rows = rows
	return nil
}

func NewUserGroup(db *dbtest.InMemoryDB) UserGroup {
	return UserGroup{
		db: db,
	}
}
