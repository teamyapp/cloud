package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/runner"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Authorization struct {
	authorizationService service.Authorization
	proto.UnimplementedAuthorizationServer
}

var _ runner.Service = (*Authorization)(nil)
var _ proto.AuthorizationServer = (*Authorization)(nil)

func (a Authorization) HasPermission(ctx context.Context, req *proto.HasPermissionRequest) (*proto.HasPermissionResponse, error) {
	hasPermission, err := a.authorizationService.HasPermission(req.ResourceType, req.ResourceId, req.Operation, req.UserId)
	if err != nil {
		return nil, err
	}

	return &proto.HasPermissionResponse{HasPermission: hasPermission}, nil
}

func (a Authorization) RegisterResourceRelation(ctx context.Context, request *proto.RegisterResourceRelationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UnregisterResourceRelation(ctx context.Context, request *proto.UnregisterResourceRelationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RegisterUserGroup(ctx context.Context, request *proto.RegisterUserGroupRequest) (*proto.RegisterUserGroupResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UnregisterUserGroup(ctx context.Context, request *proto.UnregisterUserGroupRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) Start(rn *runner.ServiceRunner) error {
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterAuthorizationServer(server, a)
	})
	return nil
}

func NewAuthorization(authorizationService service.Authorization) Authorization {
	return Authorization{
		authorizationService: authorizationService,
	}
}
