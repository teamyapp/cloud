package authorization

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type ResourceOperation struct {
	ResourceType string
	Operation    string
	ResourceID   uint64
}

type ResourceTypeOperation struct {
	ResourceType string
	Operation    string
}

type Query struct {
	ResourceType string
	ResourceID   uint64
	Operation    string
	UserID       uint64
}

func (q Query) String() string {
	return fmt.Sprintf("[Query UserID=%v Operation=%v ResourceType=%v ResourceID=%v]", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}

type Authorizer struct {
	logger              telemetry.Logger
	cloudClientRegistry *api.ClientRegistry
}

func (a Authorizer) HasPermission(ct context.Context, query Query) (bool, *errs.Error) {
	hasPermissionReq := &proto.HasPermissionRequest{
		ResourceType: query.ResourceType,
		ResourceId:   query.ResourceID,
		Operation:    query.Operation,
		UserId:       query.UserID,
	}
	hasPermissionRes, err := a.cloudClientRegistry.AuthorizationClient().HasPermission(ct, hasPermissionReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return false, internalErr
	}

	return hasPermissionRes.HasPermission, nil
}

func (a Authorizer) RegisterResource(ct context.Context, resourceType string, resourceID uint64) *errs.Error {
	registerResourceReq := &proto.RegisterResourceRequest{
		ResourceType: resourceType,
		ResourceId:   resourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().RegisterResource(ct, registerResourceReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return internalErr
	}

	return nil
}

func (a Authorizer) AssignParentResource(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64) *errs.Error {
	assignParentResourceReq := &proto.AssignParentResourceRequest{
		ChildResourceType:  childResourceType,
		ChildResourceId:    childResourceID,
		ParentResourceType: parentResourceType,
		ParentResourceId:   parentResourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AssignParentResource(ct, assignParentResourceReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return internalErr
	}

	return nil
}

func (a Authorizer) AddMemberToUserGroup(ct context.Context, userGroupID uint64, memberID uint64) *errs.Error {
	addUserGroupMemberReq := &proto.AddUserGroupMemberRequest{
		GroupId: userGroupID,
		UserId:  memberID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AddUserGroupMember(ct, addUserGroupMemberReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return internalErr
	}

	return nil
}

func (a Authorizer) CreateUserGroup(ct context.Context, creatorUserID uint64, userGroupName string, description *string) (uint64, *errs.Error) {
	createUserGroupReq := &proto.CreateUserGroupRequest{
		Name:        userGroupName,
		Description: description,
	}

	createUserGroupRes, err := a.cloudClientRegistry.AuthorizationClient().CreateUserGroup(ct, createUserGroupReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return 0, internalErr
	}

	// add the group creator to the newly created userGroup
	internalErr := a.AddMemberToUserGroup(ct, createUserGroupRes.UserGroup.GroupId, creatorUserID)
	if err != nil {
		return 0, internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("UserGroup %s is successfully created", userGroupName))
	return createUserGroupRes.UserGroup.GroupId, nil
}

func (a Authorizer) AssignPermission(
	ct context.Context,
	resourceOperation ResourceOperation,
	userGroupID uint64,
) *errs.Error {
	addPermissionReq := &proto.AddPermissionRequest{
		ResourceType: resourceOperation.ResourceType,
		ResourceId:   resourceOperation.ResourceID,
		Operation:    resourceOperation.Operation,
		GroupId:      userGroupID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AddPermission(ct, addPermissionReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("Permission %s is successfully assigned", addPermissionReq))
	return nil
}

func (a Authorizer) assignUserGroupPermissions(
	ct context.Context,
	resourceOperations []ResourceOperation,
	groupID uint64,
) *errs.Error {
	for _, resourceOperation := range resourceOperations {
		err := a.AssignPermission(ct, resourceOperation, groupID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a Authorizer) CreateUserGroupAndAssignPermissions(
	ct context.Context,
	creatorUserID uint64,
	userGroupName string,
	description *string,
	resourceOperations []ResourceOperation,
) (uint64, *errs.Error) {
	userGroupID, err := a.CreateUserGroup(ct, creatorUserID, userGroupName, description)
	if err != nil {
		return 0, err
	}

	err = a.assignUserGroupPermissions(ct, resourceOperations, userGroupID)
	if err != nil {
		return 0, err
	}

	return userGroupID, nil
}

func NewAuthorizer(
	logger telemetry.Logger,
	cloudClientRegistry *api.ClientRegistry,
) Authorizer {
	return Authorizer{
		logger:              logger,
		cloudClientRegistry: cloudClientRegistry,
	}
}
