package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/runner"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (a Authorization) ListResourceTypes(ct context.Context, query *proto.ListResourceTypesQuery) (*proto.ListResourceTypesResponse, error) {
	resourceTypeQuery := service.ResourceTypeQuery{
		ResourceType:  query.ResourceType,
		CreatorUserID: query.CreatorUserId,
		Limit:         query.Limit,
	}

	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceTypeQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceTypeQuery.EndCreationTime = &endCreationTime
	}

	resourceTypeEntities, err := a.authorizationService.ListResourceTypes(ct, resourceTypeQuery)
	if err != nil {
		return nil, err
	}

	var resourceTypes []*proto.ResourceType
	resourceTypes = collect.Map(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType, _ int) *proto.ResourceType {
		return &proto.ResourceType{
			ResourceType:  resourceTypeEntity.ResourceType,
			CreatedAt:     timestamppb.New(resourceTypeEntity.CreatedAt),
			CreatorUserId: resourceTypeEntity.CreatorUserID,
		}
	})
	return &proto.ListResourceTypesResponse{ResourceTypes: resourceTypes}, nil
}

func (a Authorization) RegisterResourceType(ct context.Context, request *proto.RegisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResourceType(ct, request.ResourceType)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnregisterResourceType(ct context.Context, request *proto.UnregisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResourceType(ct, request.ResourceType)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListResources(ct context.Context, query *proto.ListResourcesQuery) (*proto.ListResourcesResponse, error) {
	resourceQuery := service.ResourceQuery{
		ResourceType:  query.ResourceType,
		ResourceID:    query.ResourceId,
		CreatorUserID: query.CreatorUserId,
		Limit:         query.Limit,
	}

	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceQuery.EndCreationTime = &endCreationTime
	}

	resourceEntities, err := a.authorizationService.ListResources(ct, resourceQuery)
	if err != nil {
		return nil, err
	}

	var resources []*proto.Resource
	resources = collect.Map(resourceEntities, func(resource entity.Resource, _ int) *proto.Resource {
		return &proto.Resource{
			ResourceType:  resource.ResourceType,
			ResourceId:    resource.ResourceID,
			CreatedAt:     timestamppb.New(resource.CreatedAt),
			CreatorUserId: resource.CreatorUserID,
		}
	})

	return &proto.ListResourcesResponse{Resources: resources}, nil
}

func (a Authorization) RegisterResource(ct context.Context, request *proto.RegisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResource(ct, request.ResourceType, request.ResourceId)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnregisterResource(ct context.Context, request *proto.UnregisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResource(ct, request.ResourceType, request.ResourceId)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListResourceRelations(ct context.Context, query *proto.ListResourceRelationsQuery) (*proto.ListResourceRelationsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AssignParentResource(ct context.Context, request *proto.AssignParentResourceRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UnassignParentResource(ct context.Context, request *proto.UnassignParentResourceRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListOperations(ct context.Context, query *proto.ListOperationsQuery) (*proto.ListOperationsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RegisterOperation(ct context.Context, request *proto.RegisterOperationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UnregisterOperation(ct context.Context, request *proto.UnregisterOperationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListOperationRelations(ct context.Context, query *proto.ListOperationRelationsQuery) (*proto.ListOperationRelationsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AssignParentOperation(ctx context.Context, request *proto.AssignParentOperationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UnassignParentOperation(ctx context.Context, request *proto.UnassignParentOperationRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListUserGroups(ctx context.Context, query *proto.ListUserGroupsQuery) (*proto.ListUserGroupsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) CreateUserGroup(ctx context.Context, request *proto.CreateUserGroupRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) UpdateUserGroup(ctx context.Context, request *proto.UpdateUserGroupRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) DeleteUserGroup(ctx context.Context, request *proto.DeleteUserGroupRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListUserGroupMembers(ctx context.Context, query *proto.ListUserGroupMembersQuery) (*proto.ListUserGroupMembersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AddUserGroupMember(ctx context.Context, request *proto.AddUserGroupMemberRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RemoveUserGroupMember(ctx context.Context, request *proto.RemoveUserGroupMemberRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListPermissions(ctx context.Context, query *proto.ListPermissionsQuery) (*proto.ListPermissionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AddPermission(ctx context.Context, request *proto.AddPermissionRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RemovePermission(ctx context.Context, request *proto.RemovePermissionRequest) (*emptypb.Empty, error) {
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
