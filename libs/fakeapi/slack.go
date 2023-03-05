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

func SlackProxyRoutes(webAPIServerPort int) []networktest.ProxyRoute {
	return []networktest.ProxyRoute{
		{
			Endpoint: "slack.com:80",
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webAPIServerPort))
			},
		},
	}
}

type Slack struct {
}

var _ runner.Service = (*Slack)(nil)

func (s Slack) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join("/openid", "connect", "authorize"),
			HandlerFunc: s.webOAuthAuthorize,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join("/api", "openid.connect.token"),
			HandlerFunc: s.webOAuthGetAccessToken,
		},
	})
	return nil
}

func (s Slack) webOAuthAuthorize(writer http.ResponseWriter, request *http.Request) {
	panic("implement me")
}

func (s Slack) webOAuthGetAccessToken(writer http.ResponseWriter, request *http.Request) {
	panic("implement me")
}
