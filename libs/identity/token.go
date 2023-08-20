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
		return "", errs.NewError(errs.NotFound, "authorization header not found")
	}

	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		return "", errs.NewError(
			errs.InvalidFormat,
			fmt.Sprintf("authotization header must have 2 parts: header=%v", value),
		)
	}

	if parts[0] != "Bearer" {
		return "", errs.NewError(
			errs.InvalidFormat,
			fmt.Sprintf("missing beginning Bearer: header=%v", value),
		)
	}

	return parts[1], nil
}
