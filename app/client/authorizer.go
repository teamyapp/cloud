package client

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Authorizer struct {
	logger   telemetry.Logger
	registry *Registry
}

func (a Authorizer) HasPermission(ct context.Context, query authorization.Query) (bool, *errs.Error) {
	hasPermissionReq := &proto.HasPermissionRequest{
		ResourceType: query.ResourceType,
		ResourceId:   query.ResourceID,
		Operation:    query.Operation,
		UserId:       query.UserID,
	}
	hasPermissionRes, err := a.registry.AuthorizationClient().HasPermission(ct, hasPermissionReq)
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
	_, err := a.registry.AuthorizationClient().RegisterResource(ct, registerResourceReq)
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
	_, err := a.registry.AuthorizationClient().AssignParentResource(ct, assignParentResourceReq)
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
	_, err := a.registry.AuthorizationClient().AddUserGroupMember(ct, addUserGroupMemberReq)
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

	createUserGroupRes, err := a.registry.AuthorizationClient().CreateUserGroup(ct, createUserGroupReq)
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
	resourceOperation authorization.ResourceOperation,
	userGroupID uint64,
) *errs.Error {
	addPermissionReq := &proto.AddPermissionRequest{
		ResourceType: resourceOperation.ResourceType,
		ResourceId:   resourceOperation.ResourceID,
		Operation:    resourceOperation.Operation,
		GroupId:      userGroupID,
	}
	_, err := a.registry.AuthorizationClient().AddPermission(ct, addPermissionReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		return internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("Permission %s is successfully assigned", addPermissionReq))
	return nil
}

func (a Authorizer) assignUserGroupPermissions(
	ct context.Context,
	resourceOperations []authorization.ResourceOperation,
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
	resourceOperations []authorization.ResourceOperation,
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
	registry *Registry,
) Authorizer {
	return Authorizer{
		logger:   logger,
		registry: registry,
	}
}

func FilterAuthorizedItems[Item any](
	ct context.Context,
	authorizer Authorizer,
	items []Item,
	getAuthorizationQuery func(item Item) authorization.Query,
) ([]Item, *errs.Error) {
	authorizedItems := make([]Item, 0)
	for _, item := range items {
		query := getAuthorizationQuery(item)
		hasPermission, err := authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			continue
		}

		authorizedItems = append(authorizedItems, item)
	}

	return authorizedItems, nil
}
