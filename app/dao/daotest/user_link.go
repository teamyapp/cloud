package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserLink struct {
	db *InMemoryDB
}

var _ dao.UserLink = (*UserLink)(nil)

func (u UserLink) FindUserLinkByExternalUserID(ct context.Context, authProvider string, externalUserID string) (entity.UserLink, *errs.Error) {
	table, err := u.db.GetTable(UserLinkTableName)
	if err != nil {
		return entity.UserLink{}, err
	}

	for _, rawRow := range table.rows {
		userLink := rawRow.(entity.UserLink)
		if userLink.AuthProvider == authProvider &&
			userLink.ExternalUserID == externalUserID {
			return userLink, nil
		}
	}

	return entity.UserLink{}, &errs.Error{
		Code: errs.NotFound,
		Message: fmt.Sprintf("row not found: authProvider=%v, externalUserID=%v",
			authProvider,
			externalUserID),
	}
}

func (u UserLink) FindUserLinksByInternalUserID(ct context.Context, internalUserID uint64) ([]entity.UserLink, *errs.Error) {
	table, err := u.db.GetTable(UserLinkTableName)
	if err != nil {
		return nil, err
	}

	userLinks := make([]entity.UserLink, 0)
	for _, rawRow := range table.rows {
		userLink := rawRow.(entity.UserLink)
		if userLink.InternalUserID == internalUserID {
			userLinks = append(userLinks, userLink)
		}
	}

	return userLinks, nil
}

func (u UserLink) CreateUserLink(ct context.Context, userLink entity.UserLink) *errs.Error {
	_, err := u.FindUserLinkByExternalUserID(ct, userLink.AuthProvider, userLink.ExternalUserID)
	if err == nil {
		return &errs.Error{
			Code:    errs.AlreadyExists,
			Message: fmt.Sprintf("row already exist: userLink=%v", userLink),
		}
	}

	if err.Code != errs.NotFound {
		return err
	}

	table, err := u.db.GetTable(UserLinkTableName)
	if err != nil {
		return err
	}

	table.rows = append(table.rows, userLink)
	return nil
}

func (u UserLink) DeleteUserLink(ct context.Context, authProvider string, internalUserID uint64) *errs.Error {
	table, err := u.db.GetTable(UserLinkTableName)
	if err != nil {
		return err
	}

	rows := make([]interface{}, 0)
	for _, rawRow := range table.rows {
		userLink := rawRow.(entity.UserLink)
		if userLink.AuthProvider != authProvider ||
			userLink.InternalUserID != internalUserID {
			rows = append(rows, rawRow)
		}
	}

	table.rows = rows
	return nil
}

func NewUserLink(db *InMemoryDB) UserLink {
	return UserLink{
		db: db,
	}
}
