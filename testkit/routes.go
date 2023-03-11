package testkit

import (
	"fmt"
	"net"
	"strings"

	"github.com/teamyapp/cloud/libs/network/networktest"
)

const WebServerHost = "web.backend.cloud"
const WebServerPort = 80
const GRPCServerHost = "rpc.backend.cloud"
const GRPCServerPort = 80

func proxyRoutes(webListenerPort int, gRPCListenerPort int) []networktest.ProxyRoute {
	return []networktest.ProxyRoute{
		{
			Endpoint: fmt.Sprintf("%s:%d", WebServerHost, WebServerPort),
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webListenerPort))
			},
		},
		{
			Endpoint: fmt.Sprintf("%s:%d", GRPCServerHost, GRPCServerPort),
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", gRPCListenerPort))
			},
		},
	}
}
