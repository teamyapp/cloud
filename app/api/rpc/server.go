package rpc

import (
	"github.com/teamyapp/cloud/app/api/rpc/access_control"
	"github.com/teamyapp/cloud/app/api/rpc/proto"
	"google.golang.org/grpc"
)

func NewAPIServer() *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterAccessControlServer(server, access_control.NewRPCServer())
	return server
}
