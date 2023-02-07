package rpc

import (
	"fmt"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const ConnectionErr errs.ErrorCode = "Connection"

type ConnectionConfig struct {
	Host           string
	Port           int
	ShouldEncrypt  bool
	GetAccessToken func() string
	RequestTimeout time.Duration
}

func NewClientConnection(dataCollector telemetry.DataCollector, cfg ConnectionConfig, retry retry.Retry) (*grpc.ClientConn, error) {
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
			middleware.ClientGRPCWithRetry(retry),
			middleware.ClientGRPCWithRequestID(dataCollector),
			middleware.ClientGRPCWithTimout(cfg.RequestTimeout),
			middleware.ClientGRPCWithIdentity(cfg.GetAccessToken),
		))
}
