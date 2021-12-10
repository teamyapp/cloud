package errs

import (
	"github.com/teamyapp/cloud/app/entity"
)

type NotFound struct {
	Message string
}

var _ error = (*NotFound)(nil)

func (n NotFound) Error() string {
	return n.Message
}

type InvalidAuthorizationQuery entity.AuthorizationQuery

var _ error = (*InvalidAuthorizationQuery)(nil)

func (i InvalidAuthorizationQuery) Error() string {
	return "invalid authorization query"
}
