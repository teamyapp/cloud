package fakeapi

import (
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/runner"
)

func GoogleProxyRoutes(webAPIServerPort int) []networktest.ProxyRoute {
	return []networktest.ProxyRoute{
		{
			Endpoint: "accounts.google.com:80",
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webAPIServerPort))
			},
		},
		{
			Endpoint: "oauth2.googleapis.com:80",
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webAPIServerPort))
			},
		},
	}
}

type Google struct {
}

var _ runner.Service = (*Google)(nil)

func (g Google) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join("/o", "oauth2", "v2", "auth"),
			HandlerFunc: g.webOAuthAuthorize,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join("/token"),
			HandlerFunc: g.webOAuthGetAccessToken,
		},
	})
	return nil
}

func (g Google) webOAuthAuthorize(writer http.ResponseWriter, request *http.Request) {
	panic("implement me")
}

func (g Google) webOAuthGetAccessToken(writer http.ResponseWriter, request *http.Request) {
	panic("implement me")
}
