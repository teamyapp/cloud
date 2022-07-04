package rpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConnectionConfig struct {
	Host          string
	Port          int
	ShouldEncrypt bool
}

func NewClientConnection(cfg ClientConnectionConfig) (*grpc.ClientConn, error) {
	var cred credentials.TransportCredentials
	if cfg.ShouldEncrypt {
		cred = credentials.NewTLS(nil)
	} else {
		cred = insecure.NewCredentials()
	}

	opts := grpc.WithTransportCredentials(cred)
	return grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), opts)
}
