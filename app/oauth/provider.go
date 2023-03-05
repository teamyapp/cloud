package oauth

import (
	"context"
	"net/url"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Provider interface {
	GetName() string
	GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, *errs.Error)
	GetStateID(ct context.Context, fullURL *url.URL) (uint64, *errs.Error)
	GetAuthorizationCode(ct context.Context, fullURL *url.URL) string
	GetSignInURL(ct context.Context, stateID uint64) (string, *errs.Error)
}
