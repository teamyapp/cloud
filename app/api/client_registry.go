package api

import (
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/rpc"
	"google.golang.org/grpc"
)

type ClientRegistryConfig struct {
	Host          string `envconfig:"CLOUD_API_HOST" default:"localhost"`
	Port          int    `envconfig:"CLOUD_API_PORT" default:"9501"`
	ShouldEncrypt bool   `envconfig:"CLOUD_API_SHOULD_ENCRYPT" default:"false"`
}

type ClientRegistry struct {
	conn            *grpc.ClientConn
	generatorClient proto.GeneratorClient
}

func (c *ClientRegistry) GeneratorClient() proto.GeneratorClient {
	if c.generatorClient == nil {
		c.generatorClient = proto.NewGeneratorClient(c.conn)
	}

	return c.generatorClient
}

func NewClientRegistry(cfg ClientRegistryConfig) (*ClientRegistry, error) {
	conn, err := rpc.NewClientConnection(rpc.ClientConnectionConfig{
		Host:          cfg.Host,
		Port:          cfg.Port,
		ShouldEncrypt: cfg.ShouldEncrypt,
	})
	if err != nil {
		return nil, err
	}

	return &ClientRegistry{
		conn: conn,
	}, nil
}
