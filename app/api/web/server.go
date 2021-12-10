package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/api/web/identity"
	"github.com/teamyapp/cloud/app/service"
)

func NewAPIServer(identityService service.Identity) *http.ServeMux {
	serveMux := http.NewServeMux()
	router := mux.NewRouter()

	webSocketOriginChecker := func(r *http.Request) bool {
		return true
	}
	routes := identity.GetRoutes(
		webSocketOriginChecker,
		identityService)
	for _, r := range routes {
		router.HandleFunc(r.Path, r.HandleFunc).Methods(r.Method)
	}

	serveMux.HandleFunc("/", enableCORS(router.ServeHTTP))
	return serveMux
}

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if request.Method == http.MethodOptions {
			return
		}

		handlerFunc(writer, request)
	}
}
