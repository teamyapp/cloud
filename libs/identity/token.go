package identity

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/teamyapp/cloud/libs/errs"
)

func GetBearerToken(request *http.Request) (string, *errs.Error) {
	value := request.Header.Get("Authorization")
	if len(value) == 0 {
		internalErr := errs.NewError(errs.NotFound, "authorization header not found")
		return "", internalErr
	}

	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		internalErr := errs.NewError(errs.InvalidFormat, fmt.Sprintf("authotization header must have 2 parts: header=%v", value))
		return "", internalErr
	}

	if parts[0] != "Bearer" {
		internalErr := errs.NewError(errs.InvalidFormat, fmt.Sprintf("missing beginning Bearer: header=%v", value))
		return "", internalErr
	}

	return parts[1], nil
}
