package dao

import (
	"context"

	"github.com/teamyapp/cloud/app/entity"
)

type UserLink interface {
	FindUserLinkByExternalUserID(ct context.Context, authProvider string, externalUserID string) (entity.UserLink, error)
	FindUserLinksByInternalUserID(ct context.Context, internalUserID uint64) ([]entity.UserLink, error)
	CreateUserLink(ct context.Context, userLink entity.UserLink) error
	DeleteUserLink(ct context.Context, authProvider string, internalUserID uint64) error
}
