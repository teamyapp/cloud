package middleware

import (
	"net/http"
	"strings"

	"github.com/teamyapp/cloud/libs/ctx"
)

var ServerHTTPEnableCORS Middleware[http.HandlerFunc] = func(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		supportedHeaders := []string{"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization"}
		supportedCustomHeaders := ctx.GetSupportedCustomHeaders()
		supportedHeaders = append(supportedHeaders, supportedCustomHeaders...)
		writer.Header().Set("Access-Control-Allow-Headers", strings.Join(supportedHeaders, ", "))
		if request.Method == http.MethodOptions {
			return
		}

		handlerFunc(writer, request)
	}
}
