package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type UserLink interface {
	FindUserLinkByExternalUserID(ct context.Context, authProvider string, externalUserID string) (entity.UserLink, *errs.Error)
	FindUserLinksByInternalUserID(ct context.Context, internalUserID uint64) ([]entity.UserLink, *errs.Error)
	CreateUserLink(ct context.Context, userLink entity.UserLink) *errs.Error
	DeleteUserLink(ct context.Context, authProvider string, internalUserID uint64) *errs.Error
}
