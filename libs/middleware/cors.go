package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
)

var ServerHTTPEnableCORS HTTPServerMiddleware = func(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers",
			fmt.Sprintf("Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, %v", strings.Join(ctx.GetSupportedCustomHeaders(), ", ")))
		if request.Method == http.MethodOptions {
			return
		}

		handlerFunc(writer, request)
	}
}
