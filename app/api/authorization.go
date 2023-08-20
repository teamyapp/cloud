package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Authorization struct {
	logger               telemetry.Logger
	authorizationService service.Authorization
	proto.UnimplementedAuthorizationServer
}

var _ runner.Service = (*Authorization)(nil)
var _ proto.AuthorizationServer = (*Authorization)(nil)

func (a Authorization) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterAuthorizationServer(server, a)
	})
	return nil
}

func (a Authorization) HasPermission(ct context.Context, req *proto.HasPermissionRequest) (*proto.HasPermissionResponse, error) {
	hasPermission, err := a.authorizationService.HasPermission(ct, req.ResourceType, req.ResourceId, req.Operation, req.UserId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.HasPermissionResponse{HasPermission: hasPermission}, nil
}

func (a Authorization) ListResourceTypes(ct context.Context, query *proto.ListResourceTypesQuery) (*proto.ListResourceTypesResponse, error) {
	resourceTypeQuery := service.ResourceTypeQuery{
		ResourceTypeName: query.ResourceType,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
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
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var resourceTypes []*proto.ResourceType
	resourceTypes = collect.Map(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType, _ int) *proto.ResourceType {
		return &proto.ResourceType{
			ResourceType:  resourceTypeEntity.ResourceTypeName,
			CreatedAt:     timestamppb.New(resourceTypeEntity.CreatedAt),
			CreatorUserId: resourceTypeEntity.CreatorUserID,
		}
	})
	return &proto.ListResourceTypesResponse{ResourceTypes: resourceTypes}, nil
}

func (a Authorization) RegisterResourceType(ct context.Context, request *proto.RegisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResourceType(ct, request.ResourceType)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) UnregisterResourceType(ct context.Context, request *proto.UnregisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResourceType(ct, request.ResourceType)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListResources(ct context.Context, query *proto.ListResourcesQuery) (*proto.ListResourcesResponse, error) {
	resourceQuery := service.ResourceQuery{
		ResourceTypeName: query.ResourceType,
		ResourceID:       query.ResourceId,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
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
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	resources := collect.Map(resourceEntities, func(resource entity.Resource, _ int) *proto.Resource {
		return &proto.Resource{
			ResourceType:  resource.ResourceTypeName,
			ResourceId:    resource.ResourceID,
			CreatedAt:     timestamppb.New(resource.CreatedAt),
			CreatorUserId: resource.CreatorUserID,
		}
	})

	return &proto.ListResourcesResponse{Resources: resources}, nil
}

func (a Authorization) RegisterResource(ct context.Context, request *proto.RegisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResource(ct, request.ResourceType, request.ResourceId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) UnregisterResource(ct context.Context, request *proto.UnregisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResource(ct, request.ResourceType, request.ResourceId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListResourceRelations(ct context.Context, query *proto.ListResourceRelationsQuery) (*proto.ListResourceRelationsResponse, error) {
	resourceRelationQuery := service.ResourceRelationQuery{
		ChildResourceType:  query.ChildResourceType,
		ChildResourceID:    query.ChildResourceId,
		ParentResourceType: query.ParentResourceType,
		ParentResourceID:   query.ParentResourceId,
		CreatorUserID:      query.CreatorUserId,
		Limit:              query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceRelationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceRelationQuery.EndCreationTime = &endCreationTime
	}

	resourceRelationEntities, err := a.authorizationService.ListResourceRelations(ct, resourceRelationQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	resourceRelations := collect.Map(resourceRelationEntities, func(resourceRelation entity.ResourceRelation, _ int) *proto.ResourceRelation {
		return &proto.ResourceRelation{
			ChildResourceType:  resourceRelation.ChildResourceType,
			ChildResourceId:    resourceRelation.ChildResourceID,
			ParentResourceType: resourceRelation.ParentResourceType,
			ParentResourceId:   resourceRelation.ParentResourceID,
			CreatedAt:          timestamppb.New(resourceRelation.CreatedAt),
			CreatorUserId:      resourceRelation.CreatorUserID,
		}
	})

	return &proto.ListResourceRelationsResponse{ResourceRelations: resourceRelations}, nil
}

func (a Authorization) AssignParentResource(ct context.Context, request *proto.AssignParentResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AssignParentResource(
		ct,
		request.ChildResourceType,
		request.ChildResourceId,
		request.ParentResourceType,
		request.ParentResourceId,
	)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return &emptypb.Empty{}, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) UnassignParentResource(ct context.Context, request *proto.UnassignParentResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnassignParentResource(
		ct,
		request.ChildResourceType,
		request.ChildResourceId,
		request.ParentResourceType,
		request.ParentResourceId,
	)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListOperations(ct context.Context, query *proto.ListOperationsQuery) (*proto.ListOperationsResponse, error) {
	operationQuery := service.OperationQuery{
		ResourceTypeName: query.ResourceType,
		OperationName:    query.Operation,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		operationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		operationQuery.EndCreationTime = &endCreationTime
	}

	operationEntities, err := a.authorizationService.ListOperations(ct, operationQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	operations := collect.Map(operationEntities, func(operation entity.Operation, _ int) *proto.Operation {
		return &proto.Operation{
			ResourceType:  operation.ResourceTypeName,
			Operation:     operation.OperationName,
			CreatedAt:     timestamppb.New(operation.CreatedAt),
			CreatorUserId: operation.CreatorUserID,
		}
	})

	return &proto.ListOperationsResponse{Operations: operations}, nil
}

func (a Authorization) RegisterOperation(ct context.Context, request *proto.RegisterOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterOperation(ct, request.ResourceType, request.Operation)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) UnregisterOperation(ct context.Context, request *proto.UnregisterOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterOperation(ct, request.ResourceType, request.Operation)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return &emptypb.Empty{}, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListOperationRelations(ct context.Context, query *proto.ListOperationRelationsQuery) (*proto.ListOperationRelationsResponse, error) {
	operationRelationQuery := service.OperationRelationQuery{
		ChildResourceType:  query.ChildResourceType,
		ChildOperation:     query.ChildOperation,
		ParentResourceType: query.ParentResourceType,
		ParentOperation:    query.ParentOperation,
		CreatorUserID:      query.CreatorUserId,
		Limit:              query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		operationRelationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		operationRelationQuery.EndCreationTime = &endCreationTime
	}

	operationRelationEntities, err := a.authorizationService.ListOperationRelations(ct, operationRelationQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	operationRelations := collect.Map(operationRelationEntities, func(operationRelation entity.OperationRelation, _ int) *proto.OperationRelation {
		return &proto.OperationRelation{
			ChildResourceType:  operationRelation.ChildResourceType,
			ChildOperation:     operationRelation.ChildOperation,
			ParentResourceType: operationRelation.ParentResourceType,
			ParentOperation:    operationRelation.ParentOperation,
			CreatedAt:          timestamppb.New(operationRelation.CreatedAt),
			CreatorUserId:      operationRelation.CreatorUserID,
		}
	})

	return &proto.ListOperationRelationsResponse{OperationRelations: operationRelations}, nil
}

func (a Authorization) AssignParentOperation(ct context.Context, request *proto.AssignParentOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AssignParentOperation(
		ct,
		request.ChildResourceType,
		request.ChildOperation,
		request.ParentResourceType,
		request.ParentOperation,
	)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) UnassignParentOperation(ct context.Context, request *proto.UnassignParentOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnassignParentOperation(
		ct,
		request.ChildResourceType,
		request.ChildOperation,
		request.ParentResourceType,
		request.ParentOperation,
	)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListUserGroups(ct context.Context, query *proto.ListUserGroupsQuery) (*proto.ListUserGroupsResponse, error) {
	userGroupQuery := service.UserGroupQuery{
		ID:                  query.Id,
		NameContains:        query.NameContains,
		DescriptionContains: query.DescriptionContains,
		CreatorUserID:       query.CreatorUserId,
		Limit:               query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		userGroupQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		userGroupQuery.EndCreationTime = &endCreationTime
	}

	userGroupEntities, err := a.authorizationService.ListUserGroups(ct, userGroupQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	userGroups := collect.Map(userGroupEntities, func(userGroup entity.UserGroup, _ int) *proto.UserGroup {
		return &proto.UserGroup{
			GroupId:       userGroup.ID,
			Name:          userGroup.Name,
			Description:   userGroup.Description,
			CreatedAt:     timestamppb.New(userGroup.CreatedAt),
			CreatorUserId: userGroup.CreatorUserID,
			UpdatedAt:     toProtoTimePtr(userGroup.UpdatedAt),
		}
	})
	return &proto.ListUserGroupsResponse{UserGroups: userGroups}, nil
}

func (a Authorization) CreateUserGroup(ct context.Context, request *proto.CreateUserGroupRequest) (*proto.CreateUserGroupResponse, error) {
	userGroup, err := a.authorizationService.CreateUserGroup(ct, request.Name, request.Description)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.CreateUserGroupResponse{UserGroup: &proto.UserGroup{
		GroupId:       userGroup.ID,
		Name:          userGroup.Name,
		Description:   userGroup.Description,
		CreatedAt:     timestamppb.New(userGroup.CreatedAt),
		CreatorUserId: userGroup.CreatorUserID,
		UpdatedAt:     toProtoTimePtr(userGroup.UpdatedAt),
	}}, nil
}

func (a Authorization) UpdateUserGroup(ct context.Context, request *proto.UpdateUserGroupRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UpdateUserGroup(ct, request.GroupId, request.Name, request.Description)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) DeleteUserGroup(ct context.Context, request *proto.DeleteUserGroupRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.DeleteUserGroup(ct, request.GroupId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListUserGroupMembers(ct context.Context, query *proto.ListUserGroupMembersQuery) (*proto.ListUserGroupMembersResponse, error) {
	userGroupMemberQuery := service.UserGroupMemberQuery{
		GroupID:       query.GroupId,
		UserID:        query.UserId,
		CreatorUserID: query.CreatorUserId,
		Limit:         query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		userGroupMemberQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		userGroupMemberQuery.EndCreationTime = &endCreationTime
	}

	userGroupMemberEntities, err := a.authorizationService.ListUserGroupMembers(ct, userGroupMemberQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	userGroupMembers := collect.Map(userGroupMemberEntities, func(userGroupMember entity.UserGroupMember, _ int) *proto.UserGroupMember {
		return &proto.UserGroupMember{
			GroupId:       userGroupMember.GroupID,
			UserId:        userGroupMember.UserID,
			CreatedAt:     timestamppb.New(userGroupMember.CreatedAt),
			CreatorUserId: userGroupMember.CreatorUserID,
		}
	})
	return &proto.ListUserGroupMembersResponse{UserGroupMembers: userGroupMembers}, nil
}

func (a Authorization) AddUserGroupMember(ct context.Context, request *proto.AddUserGroupMemberRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AddUserGroupMember(ct, request.GroupId, request.UserId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) RemoveUserGroupMember(ct context.Context, request *proto.RemoveUserGroupMemberRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RemoveUserGroupMember(ct, request.GroupId, request.UserId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ListPermissions(ct context.Context, query *proto.ListPermissionsQuery) (*proto.ListPermissionsResponse, error) {
	permissionQuery := service.PermissionQuery{
		ResourceType:  query.ResourceType,
		ResourceID:    query.ResourceId,
		Operation:     query.Operation,
		GroupID:       query.GroupId,
		CreatorUserID: query.CreatorUserId,
		Limit:         query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		permissionQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		permissionQuery.EndCreationTime = &endCreationTime
	}

	permissionEntities, err := a.authorizationService.ListPermissions(ct, permissionQuery)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	permissions := collect.Map(permissionEntities, func(permission entity.Permission, _ int) *proto.Permission {
		return &proto.Permission{
			ResourceType:  permission.ResourceType,
			ResourceId:    permission.ResourceID,
			Operation:     permission.Operation,
			GroupId:       permission.GroupID,
			CreatedAt:     timestamppb.New(permission.CreatedAt),
			CreatorUserId: permission.CreatorUserID,
		}
	})
	return &proto.ListPermissionsResponse{Permissions: permissions}, nil
}

func (a Authorization) AddPermission(ct context.Context, request *proto.AddPermissionRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AddPermission(ct, request.ResourceType, request.ResourceId, request.Operation, request.GroupId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) RemovePermission(ct context.Context, request *proto.RemovePermissionRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RemovePermission(ct, request.ResourceType, request.ResourceId, request.Operation, request.GroupId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func (a Authorization) ApplyAuthorizationConfig(ct context.Context, request *proto.ApplyAuthorizationConfigRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.ApplyAuthorizationConfig(ct, request.ConfigContent)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewAuthorization(
	logger telemetry.Logger,
	authorizationService service.Authorization,
) Authorization {
	return Authorization{
		logger:               logger,
		authorizationService: authorizationService,
	}
}
