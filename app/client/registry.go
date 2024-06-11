package client

import (
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"google.golang.org/grpc"
)

type Registry struct {
	conn                *grpc.ClientConn
	generatorClient     pbcloud.GeneratorClient
	identityClient      pbcloud.IdentityClient
	authorizationClient pbcloud.AuthorizationClient
	fileClient          pbcloud.FileClient
}

func (r *Registry) GeneratorClient() pbcloud.GeneratorClient {
	if r.generatorClient == nil {
		r.generatorClient = pbcloud.NewGeneratorClient(r.conn)
	}

	return r.generatorClient
}

func (r *Registry) IdentityClient() pbcloud.IdentityClient {
	if r.identityClient == nil {
		r.identityClient = pbcloud.NewIdentityClient(r.conn)
	}

	return r.identityClient
}

func (r *Registry) AuthorizationClient() pbcloud.AuthorizationClient {
	if r.authorizationClient == nil {
		r.authorizationClient = pbcloud.NewAuthorizationClient(r.conn)
	}

	return r.authorizationClient
}

func (r *Registry) FileClient() pbcloud.FileClient {
	if r.fileClient == nil {
		r.fileClient = pbcloud.NewFileClient(r.conn)
	}

	return r.fileClient
}

func NewRegistry(
	logger telemetry.Logger,
	network network.Network,
	clientGRPCMetrics middleware.ClientGRPCMetrics,
	connCfg rpc.ConnectionConfig,
	makeRetry func() retry.Retry,
) (*Registry, error) {
	conn, err := rpc.NewClientConnection(logger, network, clientGRPCMetrics, connCfg, makeRetry)
	if err != nil {
		return nil, err
	}

	return &Registry{
		conn: conn,
	}, nil
}
