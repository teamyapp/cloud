package api

import (
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/rpc"
	"google.golang.org/grpc"
)

type ClientRegistry struct {
	conn                *grpc.ClientConn
	generatorClient     proto.GeneratorClient
	identityClient      proto.IdentityClient
	authorizationClient proto.AuthorizationClient
	fileClient          proto.FileClient
}

func (c *ClientRegistry) GeneratorClient() proto.GeneratorClient {
	if c.generatorClient == nil {
		c.generatorClient = proto.NewGeneratorClient(c.conn)
	}

	return c.generatorClient
}

func (c *ClientRegistry) IdentityClient() proto.IdentityClient {
	if c.identityClient == nil {
		c.identityClient = proto.NewIdentityClient(c.conn)
	}

	return c.identityClient
}

func (c *ClientRegistry) AuthorizationClient() proto.AuthorizationClient {
	if c.authorizationClient == nil {
		c.authorizationClient = proto.NewAuthorizationClient(c.conn)
	}

	return c.authorizationClient
}

func (c *ClientRegistry) FileClient() proto.FileClient {
	if c.fileClient == nil {
		c.fileClient = proto.NewFileClient(c.conn)
	}

	return c.fileClient
}

func NewClientRegistry(dataCollector obs.DataCollector, connCfg rpc.ConnectionConfig) (*ClientRegistry, error) {
	conn, err := rpc.NewClientConnection(dataCollector, connCfg)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return &ClientRegistry{
		conn: conn,
	}, nil
}
