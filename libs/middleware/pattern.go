package middleware

import (
	"net/http"
)

type GetPatternFunc func(request *http.Request) (string, bool)
