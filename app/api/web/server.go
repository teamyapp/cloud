package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Service interface {
	getRoutes() []Route
}

func NewServer(services []Service) *http.ServeMux {
	serveMux := http.NewServeMux()
	router := mux.NewRouter()

	for _, service := range services {
		routes := service.getRoutes()
		for _, r := range routes {
			router.HandleFunc(r.Path, r.HandlerFunc).Methods(r.Method)
		}
	}

	serveMux.HandleFunc("/", EnableCORS(router.ServeHTTP))
	return serveMux
}

func EnableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
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
