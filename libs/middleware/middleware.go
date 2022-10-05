package middleware

import (
	"net/http"
)

type HTTPMiddleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

func HTTPWithMiddlewares(handlerFunc http.HandlerFunc, middlewares []HTTPMiddleware) http.HandlerFunc {
	if len(middlewares) == 0 {
		return handlerFunc
	}

	return middlewares[0](HTTPWithMiddlewares(handlerFunc, middlewares[1:]))
}
