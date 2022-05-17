package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

const gatewayPathPrefix = "/gateway"

type GatewayAPI struct {
}

var _ Service = (*GatewayAPI)(nil)

func (g GatewayAPI) getRoutes() []Route {
	return []Route{
		{
			Path:        path.Join(gatewayPathPrefix, "inbound", "github"),
			Method:      http.MethodPost,
			HandlerFunc: g.receiveGithubInboundEvents,
		},
	}
}

func (g GatewayAPI) receiveGithubInboundEvents(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Printf("Github inbound event: %v\n", string(buf))
	w.WriteHeader(http.StatusNoContent)
}

func NewGatewayAPI() GatewayAPI {
	return GatewayAPI{}
}
