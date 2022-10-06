package middleware

import (
	"net/http"
)

type HTTPServerMiddleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

func HTTPServerWithMiddlewares(handlerFunc http.HandlerFunc, middlewares []HTTPServerMiddleware) http.HandlerFunc {
	if len(middlewares) == 0 {
		return handlerFunc
	}

	return middlewares[0](HTTPServerWithMiddlewares(handlerFunc, middlewares[1:]))
}
