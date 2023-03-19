package rpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ConnectionConfig struct {
	Host           string
	Port           int
	ShouldEncrypt  bool
	GetAccessToken func() string
	RequestTimeout time.Duration
}

func NewClientConnection(
	dataCollector telemetry.DataCollector,
	network network.Network,
	clientGRPCMetrics middleware.ClientGRPCMetrics,
	cfg ConnectionConfig,
	makeRetry func() retry.Retry,
) (*grpc.ClientConn, error) {
	var cred credentials.TransportCredentials
	if cfg.ShouldEncrypt {
		cred = credentials.NewTLS(nil)
	} else {
		cred = insecure.NewCredentials()
	}

	return grpc.Dial(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		grpc.WithTransportCredentials(cred),
		grpc.WithChainUnaryInterceptor(
			middleware.ClientGRPCWithMetrics(clientGRPCMetrics),
			middleware.ClientGRPCUnaryWithOpenTelemetry(),
			middleware.ClientGRPCWithRetry(makeRetry),
			middleware.ClientGRPCWithRequestID(dataCollector),
			middleware.ClientGRPCWithTimout(cfg.RequestTimeout),
			middleware.ClientGRPCWithIdentity(cfg.GetAccessToken),
		),
		grpc.WithContextDialer(func(ctx context.Context, hostAndPort string) (net.Conn, error) {
			return network.Dial("tcp", hostAndPort)
		}))
}
