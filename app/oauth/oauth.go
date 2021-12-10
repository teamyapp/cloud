package oauth

import (
	"net/http"

	"github.com/teamyapp/cloud/app/entity"
	oneEntity "github.com/teamyapp/one/entity"
)

type OAuth interface {
	GetName() string
	GetSignInURL(stateID oneEntity.ID) (string, error)
	GetUser(authorizationCode string) (entity.ExternalUser, error)
	GetStateID(request *http.Request) string
	GetAuthorizationCode(request *http.Request) string
}
