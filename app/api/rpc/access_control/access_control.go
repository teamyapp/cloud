package access_control

import (
	"github.com/teamyapp/cloud/app/api/rpc/proto"
)

var _ proto.AccessControlServer = (*RPCServer)(nil)

type RPCServer struct {
	proto.UnimplementedAccessControlServer
}

func NewRPCServer() RPCServer {
	return RPCServer{}
}
