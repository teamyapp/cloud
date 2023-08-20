package client

import (
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
)

type Registry struct {
	conn                *grpc.ClientConn
	generatorClient     proto.GeneratorClient
	identityClient      proto.IdentityClient
	authorizationClient proto.AuthorizationClient
	fileClient          proto.FileClient
}

func (r *Registry) GeneratorClient() proto.GeneratorClient {
	if r.generatorClient == nil {
		r.generatorClient = proto.NewGeneratorClient(r.conn)
	}

	return r.generatorClient
}

func (r *Registry) IdentityClient() proto.IdentityClient {
	if r.identityClient == nil {
		r.identityClient = proto.NewIdentityClient(r.conn)
	}

	return r.identityClient
}

func (r *Registry) AuthorizationClient() proto.AuthorizationClient {
	if r.authorizationClient == nil {
		r.authorizationClient = proto.NewAuthorizationClient(r.conn)
	}

	return r.authorizationClient
}

func (r *Registry) FileClient() proto.FileClient {
	if r.fileClient == nil {
		r.fileClient = proto.NewFileClient(r.conn)
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
