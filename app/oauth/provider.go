package oauth

import (
	"context"
	"net/http"

	"github.com/teamyapp/cloud/app/entity"
)

type Provider interface {
	GetName() string
	GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, error)
	GetStateID(request *http.Request) (uint64, error)
	GetAuthorizationCode(request *http.Request) string
	GetSignInURL(ct context.Context, stateID uint64) (string, error)
}
