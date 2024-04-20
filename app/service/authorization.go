package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/delta"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Authorization struct {
	logger               telemetry.Logger
	resourceRelationDao  dao.ResourceRelation
	userGroupMemberDao   dao.UserGroupMember
	permissionDao        dao.Permission
	operationRelationDao dao.OperationRelation
	operationDao         dao.Operation
	resourceTypeDao      dao.ResourceType
	resourceDao          dao.Resource
	userGroupDao         dao.UserGroup
	userGroupIDGenerator *UniqueNumberGen
}

func (a Authorization) HasPermission(ct context.Context, resourceType string, resourceID uint64, operation string, userID uint64) (bool, *errs.Error) {
	// No nested group allowed
	groupIDs, err := a.userGroupMemberDao.FindGroupIDsByUserID(ct, userID)
	if err != nil {
		return false, err
	}

	for _, groupID := range groupIDs {
		hasPermission, err := a.groupHasPermission(ct, entity.PermissionQuery{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			Operation:    operation,
			GroupID:      groupID,
		})
		if err != nil {
			return false, err
		}

		if hasPermission {
			return hasPermission, nil
		}
	}

	return false, nil
}

func (a Authorization) ListResourceTypes(ct context.Context, resourceTypeQuery ResourceTypeQuery) ([]entity.ResourceType, *errs.Error) {
	allResourceTypeEntities, err := a.resourceTypeDao.FindAllResourceTypes(ct)
	if err != nil {
		return nil, err
	}

	return queryResourceTypes(allResourceTypeEntities, resourceTypeQuery), nil
}

func (a Authorization) RegisterResourceType(ct context.Context, resourceTypeName string) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	resourceTypeEntity := entity.ResourceType{
		ResourceTypeName: resourceTypeName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}

	return a.resourceTypeDao.CreateResourceType(ct, resourceTypeEntity)
}

func (a Authorization) UnregisterResourceType(ct context.Context, resourceTypeName string) *errs.Error {
	return a.resourceTypeDao.DeleteResourceType(ct, resourceTypeName)
}

func (a Authorization) ListResources(ct context.Context, resourceQuery ResourceQuery) ([]entity.Resource, *errs.Error) {
	allResources, err := a.resourceDao.FindAllResources(ct)
	if err != nil {
		return nil, err
	}

	return queryResources(allResources, resourceQuery), nil
}

func (a Authorization) RegisterResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	resource := entity.Resource{
		ResourceTypeName: resourceTypeName,
		ResourceID:       resourceID,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.resourceDao.CreateResource(ct, resource)
}

func (a Authorization) UnregisterResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error {
	return a.resourceDao.DeleteResource(ct, resourceTypeName, resourceID)
}

func (a Authorization) ListResourceRelations(ct context.Context, resourceRelationQuery ResourceRelationQuery) ([]entity.ResourceRelation, *errs.Error) {
	allResourceRelationEntities, err := a.resourceRelationDao.FindAllResourceRelations(ct)
	if err != nil {
		return nil, err
	}

	return queryResourceRelations(allResourceRelationEntities, resourceRelationQuery), nil
}

func (a Authorization) AssignParentResource(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	resourceRelation := entity.ResourceRelation{
		ChildResourceType:  childResourceType,
		ChildResourceID:    childResourceID,
		ParentResourceType: parentResourceType,
		ParentResourceID:   parentResourceID,
		CreatedAt:          time.Now().UTC(),
		CreatorUserID:      userID,
	}
	return a.resourceRelationDao.CreateResourceRelation(ct, resourceRelation)
}

func (a Authorization) UnassignParentResource(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) *errs.Error {
	return a.resourceRelationDao.DeleteResourceRelation(
		ct,
		childResourceType,
		childResourceID,
		parentResourceType,
		parentResourceID,
	)
}

func (a Authorization) ListOperations(ct context.Context, operationQuery OperationQuery) ([]entity.Operation, *errs.Error) {
	allOperations, err := a.operationDao.FindAllOperations(ct)
	if err != nil {
		return nil, err
	}

	return queryOperations(allOperations, operationQuery), nil
}

func (a Authorization) RegisterOperation(ct context.Context, resourceTypeName string, operationName string) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	operation := entity.Operation{
		ResourceTypeName: resourceTypeName,
		OperationName:    operationName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.operationDao.CreateOperation(ct, operation)
}

func (a Authorization) UnregisterOperation(ct context.Context, resourceTypeName string, operationName string) *errs.Error {
	return a.operationDao.DeleteOperation(ct, resourceTypeName, operationName)
}

func (a Authorization) ListOperationRelations(ct context.Context, operationRelationQuery OperationRelationQuery) ([]entity.OperationRelation, *errs.Error) {
	allOperationRelations, err := a.operationRelationDao.FindAllOperationRelations(ct)
	if err != nil {
		return nil, err
	}

	return queryOperationRelations(allOperationRelations, operationRelationQuery), nil
}

func (a Authorization) AssignParentOperation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	operationRelation := entity.OperationRelation{
		ChildResourceType:  childResourceType,
		ChildOperation:     childOperation,
		ParentResourceType: parentResourceType,
		ParentOperation:    parentOperation,
		CreatedAt:          time.Now().UTC(),
		CreatorUserID:      userID,
	}
	return a.operationRelationDao.CreateOperationRelation(ct, operationRelation)
}

func (a Authorization) UnassignParentOperation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) *errs.Error {
	return a.operationRelationDao.DeleteOperationRelation(
		ct,
		childResourceType,
		childOperation,
		parentResourceType,
		parentOperation,
	)
}

func (a Authorization) ListUserGroups(ct context.Context, query UserGroupQuery) ([]entity.UserGroup, *errs.Error) {
	allGroups, err := a.userGroupDao.FindAllGroups(ct)
	if err != nil {
		return nil, err
	}

	return queryUserGroups(allGroups, query), nil
}

func (a Authorization) CreateUserGroup(ct context.Context, name string, description *string) (entity.UserGroup, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.UserGroup{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	groupID, err := a.userGroupIDGenerator.GenerateUniqueNumber(ct)
	if err != nil {
		return entity.UserGroup{}, err
	}

	userGroup := entity.UserGroup{
		ID:            groupID,
		Name:          name,
		Description:   description,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: userID,
	}

	err = a.userGroupDao.CreateGroup(ct, userGroup)
	if err != nil {
		return entity.UserGroup{}, err
	}

	return userGroup, nil
}

func (a Authorization) UpdateUserGroup(ct context.Context, groupID uint64, name *string, description *string) *errs.Error {
	userGroup, err := a.userGroupDao.FindGroupByID(ct, groupID)
	if err != nil {
		return err
	}

	if name != nil {
		userGroup.Name = *name
	}

	if description != nil {
		userGroup.Description = description
	}

	nowTime := time.Now().UTC()
	userGroup.UpdatedAt = &nowTime
	return a.userGroupDao.UpdateGroup(ct, userGroup)
}

func (a Authorization) DeleteUserGroup(ct context.Context, groupID uint64) *errs.Error {
	return a.userGroupDao.DeleteGroup(ct, groupID)
}

func (a Authorization) ListUserGroupMembers(ct context.Context, query UserGroupMemberQuery) ([]entity.UserGroupMember, *errs.Error) {
	allUserGroupMembers, err := a.userGroupMemberDao.FindAllUserGroupMembers(ct)
	if err != nil {
		return nil, err
	}

	userGroupMembers := queryUserGroupMembers(allUserGroupMembers, query)
	return userGroupMembers, nil
}

func (a Authorization) AddUserGroupMember(ct context.Context, groupID uint64, userID uint64) *errs.Error {
	creatorUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	userGroupMember := entity.UserGroupMember{
		GroupID:       groupID,
		UserID:        userID,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: creatorUserID,
	}
	return a.userGroupMemberDao.CreateUserGroupMember(ct, userGroupMember)
}

func (a Authorization) RemoveUserGroupMember(ct context.Context, groupID uint64, userID uint64) *errs.Error {
	return a.userGroupMemberDao.DeleteUserGroupMember(ct, groupID, userID)
}

func (a Authorization) ListPermissions(ct context.Context, query PermissionQuery) ([]entity.Permission, *errs.Error) {
	allPermissions, err := a.permissionDao.FindAllPermissions(ct)
	if err != nil {
		return nil, err
	}

	permissions := queryPermission(allPermissions, query)
	return permissions, nil
}

func (a Authorization) AddPermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error {
	creatorUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	permission := entity.Permission{
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		Operation:     operation,
		GroupID:       groupID,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: creatorUserID,
	}
	return a.permissionDao.CreatePermission(ct, permission)
}

func (a Authorization) RemovePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error {
	return a.permissionDao.DeletePermission(ct, resourceType, resourceID, operation, groupID)
}

func (a Authorization) groupHasPermission(ct context.Context, permissionQuery entity.PermissionQuery) (bool, *errs.Error) {
	visited := make(map[entity.PermissionQuery]bool)
	visited[permissionQuery] = true
	queries := []entity.PermissionQuery{permissionQuery}
	for len(queries) > 0 {
		currQuery := queries[0]
		queries = queries[1:]

		_, err := a.permissionDao.FindPermission(ct, currQuery)
		if err == nil {
			return true, nil
		}

		if err.Code != errs.NotFound {
			return false, err
		}

		parentPermissionQueries, err := a.getParentPermissionQueries(ct, currQuery, visited)
		if err != nil {
			return false, err
		}

		queries = append(queries, parentPermissionQueries...)
	}

	return false, nil
}

func (a Authorization) getParentPermissionQueries(ct context.Context, currQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool) ([]entity.PermissionQuery, *errs.Error) {
	var parentPermissionQueries []entity.PermissionQuery
	operationRelations, err := a.operationRelationDao.FindOperationRelations(ct, currQuery.ResourceType, currQuery.Operation)
	if err != nil {
		return nil, err
	}

	resourceRelations, err := a.resourceRelationDao.FindResourceRelations(ct, currQuery.ResourceType, currQuery.ResourceID)
	if err != nil {
		return nil, err
	}

	// Handle sub resources of the same type
	// Eg. subtask of a task are both of task resource type
	for _, resourceRelation := range resourceRelations {
		if resourceRelation.ParentResourceType == currQuery.ResourceType {
			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   resourceRelation.ParentResourceID,
				ResourceType: currQuery.ResourceType,
				Operation:    currQuery.Operation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
		}
	}

	for _, operationRelation := range operationRelations {
		// Handle different operations for the same resource type
		// Eg. If a user can edit a task, the user can also read a task
		if operationRelation.ParentResourceType == currQuery.ResourceType {
			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   currQuery.ResourceID,
				ResourceType: operationRelation.ParentResourceType,
				Operation:    operationRelation.ParentOperation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
			continue
		}

		// Handle permission with different resource types
		for _, resourceRelation := range resourceRelations {
			if resourceRelation.ParentResourceType != operationRelation.ParentResourceType {
				continue
			}

			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   resourceRelation.ParentResourceID,
				ResourceType: operationRelation.ParentResourceType,
				Operation:    operationRelation.ParentOperation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
		}
	}

	return parentPermissionQueries, nil
}

func (a Authorization) ApplyAuthorizationConfig(ct context.Context, configContent string) *errs.Error {
	newConfig, err := authorization.ParseConfig(configContent)
	if err != nil {
		return err
	}

	oldConfig, err := a.configFromCurrentData(ct)
	if err != nil {
		return err
	}

	configDelta := authorization.DetectConfigDelta(oldConfig, newConfig)
	if configDelta.Status == delta.UnchangedStatus {
		return nil
	}

	return a.applyConfigDelta(ct, configDelta.Value)
}

func (a Authorization) configFromCurrentData(ct context.Context) (authorization.Config, *errs.Error) {
	resourceTypeOperationsMap := make(map[string]authorization.ResourceTypeOperations)
	resourceTypes, err := a.resourceTypeDao.FindAllResourceTypes(ct)
	if err != nil {
		return authorization.Config{}, err
	}

	for _, resourceType := range resourceTypes {
		resourceTypeOperations, err := a.getResourceTypeOperations(ct, resourceType.ResourceTypeName)
		if err != nil {
			return authorization.Config{}, nil
		}

		resourceTypeOperationsMap[resourceType.ResourceTypeName] = resourceTypeOperations
	}

	operationRelations, err := a.operationRelationDao.FindAllOperationRelations(ct)
	if err != nil {
		return authorization.Config{}, nil
	}

	operationRelationsMap := getOperationRelationsMap(operationRelations)
	return authorization.NewConfig(resourceTypeOperationsMap, operationRelationsMap), nil
}

func (a Authorization) applyConfigDelta(ct context.Context, configDelta authorization.ConfigDelta) *errs.Error {
	err := a.applyResourceTypeOperationsMapDelta(ct, configDelta.ResourceTypeOperationsDelta)
	if err != nil {
		return err
	}

	return a.applyOperationRelationsMapDelta(ct, configDelta.OperationRelationsDelta)
}

func (a Authorization) getResourceTypeOperations(ct context.Context, resourceTypeName string) (
	authorization.ResourceTypeOperations,
	*errs.Error,
) {
	operations, err := a.operationDao.FindOperationsByResourceType(ct, resourceTypeName)
	if err != nil {
		return authorization.ResourceTypeOperations{}, err
	}

	opsMap := make(map[string]bool)
	for _, operation := range operations {
		opsMap[operation.OperationName] = true
	}

	return authorization.ResourceTypeOperations{
		ResourceType: resourceTypeName,
		Operations:   opsMap,
	}, nil
}

func (a Authorization) applyResourceTypeOperationsMapDelta(
	ct context.Context,
	dt delta.Delta[map[string]delta.KeyValueDelta[authorization.ResourceTypeOperationsDelta]],
) *errs.Error {
	if dt.Status == delta.UnchangedStatus {
		return nil
	}

	for resourceTypeName, keyValueDelta := range dt.Value {
		switch keyValueDelta.KeyStatus {
		case delta.UnchangedStatus:
			resourceTypeOperations := keyValueDelta.Value
			err := a.applyResourceTypeOperationsDelta(ct, resourceTypeOperations)
			if err != nil {
				return err
			}
		case delta.AddedStatus:
			creatorUserID, ok := ctx.UserIDFromContext(ct)
			if !ok {
				return errs.NewError(errs.Unauthenticated, "user id not found")
			}

			resourceTypeOperations := keyValueDelta.Value
			err := a.resourceTypeDao.CreateResourceType(ct, entity.ResourceType{
				ResourceTypeName: resourceTypeName,
				CreatedAt:        time.Now().UTC(),
				CreatorUserID:    creatorUserID,
			})
			if err != nil {
				return err
			}

			err = a.applyResourceTypeOperationsDelta(ct, resourceTypeOperations)
			if err != nil {
				return err
			}
		case delta.RemovedStatus:
			resourceTypeOperations := keyValueDelta.Value
			err := a.applyResourceTypeOperationsDelta(ct, resourceTypeOperations)
			if err != nil {
				return err
			}

			err = a.resourceTypeDao.DeleteResourceType(
				ct,
				resourceTypeOperations.ResourceType)
			if err != nil {
				return err
			}
		case delta.UpdatedStatus:
			return errs.NewError(
				errs.InvalidArgument,
				fmt.Sprintf("resource type name cannot be updated: %v", resourceTypeName))
		}
	}

	return nil
}

func (a *Authorization) applyResourceTypeOperationsDelta(
	ct context.Context,
	dt authorization.ResourceTypeOperationsDelta,
) *errs.Error {
	for operation, operationDelta := range dt.OperationsDelta.Value {
		switch operationDelta.KeyStatus {
		case delta.AddedStatus:
			creatorUserID, ok := ctx.UserIDFromContext(ct)
			if !ok {
				return errs.NewError(errs.Unauthenticated, "user ID not found")
			}

			err := a.operationDao.CreateOperation(ct, entity.Operation{
				ResourceTypeName: dt.ResourceType,
				OperationName:    operation,
				CreatedAt:        time.Now().UTC(),
				CreatorUserID:    creatorUserID,
			})
			if err != nil {
				return err
			}
		case delta.RemovedStatus:
			err := a.operationDao.DeleteOperation(ct, dt.ResourceType, operation)
			if err != nil {
				return err
			}
		case delta.UpdatedStatus:
			return errs.NewError(
				errs.InvalidArgument,
				fmt.Sprintf("operation cannot be updated: %v", operation))
		}
	}

	return nil
}

func (a Authorization) applyOperationRelationsMapDelta(
	ct context.Context,
	dt delta.Delta[map[string]delta.KeyValueDelta[authorization.OperationRelationsDelta]],
) *errs.Error {
	if dt.Status == delta.UnchangedStatus {
		return nil
	}

	for key, keyValueDelta := range dt.Value {
		switch keyValueDelta.KeyStatus {
		case delta.UnchangedStatus, delta.AddedStatus, delta.RemovedStatus:
			operationRelations := keyValueDelta.Value
			err := a.applyOperationRelationsDelta(ct, operationRelations)
			if err != nil {
				return err
			}
		case delta.UpdatedStatus:
			return errs.NewError(
				errs.InvalidArgument,
				fmt.Sprintf("operation relations key cannot be updated: %v", key))
		}
	}

	return nil
}

func (a Authorization) applyOperationRelationsDelta(
	ct context.Context,
	dt authorization.OperationRelationsDelta,
) *errs.Error {
	switch dt.ParentOperationsDelta.Status {
	case delta.AddedStatus, delta.RemovedStatus:
		for _, parentOperationDelta := range dt.ParentOperationsDelta.Value {
			err := a.applyOperationRelationDelta(
				ct,
				dt.ChildResourceType,
				dt.ChildOperation,
				parentOperationDelta.ValueStatus,
				parentOperationDelta.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a Authorization) applyOperationRelationDelta(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentOperationStatus delta.Status,
	parentOperation authorization.Operation,
) *errs.Error {
	switch parentOperationStatus {
	case delta.AddedStatus:
		creatorUserID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		return a.operationRelationDao.CreateOperationRelation(ct, entity.OperationRelation{
			ChildResourceType:  childResourceType,
			ChildOperation:     childOperation,
			ParentResourceType: parentOperation.ResourceType,
			ParentOperation:    parentOperation.Operation,
			CreatedAt:          time.Now().UTC(),
			CreatorUserID:      creatorUserID,
		})
	case delta.RemovedStatus:
		return a.operationRelationDao.DeleteOperationRelation(
			ct, childResourceType,
			childOperation,
			parentOperation.ResourceType,
			parentOperation.Operation,
		)
	case delta.UpdatedStatus:
		return errs.NewError(
			errs.InvalidArgument,
			fmt.Sprintf("parent operation cannot be updated: %v", parentOperation))
	}

	return nil
}

func getOperationRelationsMap(
	operationRelations []entity.OperationRelation,
) map[string]authorization.OperationRelations {
	operationRelationsMap := make(map[string]authorization.OperationRelations)
	for _, operationRelation := range operationRelations {
		childOpKey := authorization.GetOperationKey(
			operationRelation.ChildResourceType,
			operationRelation.ChildOperation,
		)
		operationRelationsItem, ok := operationRelationsMap[childOpKey]
		if !ok {
			operationRelationsItem = authorization.OperationRelations{
				ResourceType:     operationRelation.ChildResourceType,
				Operation:        operationRelation.ChildOperation,
				ParentOperations: make(map[string]authorization.Operation),
			}
		}

		parentOpKey := authorization.GetOperationKey(
			operationRelation.ParentResourceType,
			operationRelation.ParentOperation,
		)
		operationRelationsItem.ParentOperations[parentOpKey] = authorization.Operation{
			ResourceType: operationRelation.ParentResourceType,
			Operation:    operationRelation.ParentOperation,
		}
		operationRelationsMap[childOpKey] = operationRelationsItem
	}

	return operationRelationsMap
}

func tryAddPermissionQueryToQueue(permissionQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool, queries []entity.PermissionQuery) []entity.PermissionQuery {
	_, ok := visited[permissionQuery]
	if ok {
		return queries
	}

	visited[permissionQuery] = true
	return append(queries, permissionQuery)
}

func NewAuthorization(
	logger telemetry.Logger,
	resourceRelationDao dao.ResourceRelation,
	userGroupMemberDao dao.UserGroupMember,
	permissionDao dao.Permission,
	operationRelationDao dao.OperationRelation,
	operationDao dao.Operation,
	resourceTypeDao dao.ResourceType,
	resourceDao dao.Resource,
	userGroupDao dao.UserGroup,
	uniqueNumberRegistry *UniqueNumberGenRegistry,
) (Authorization, error) {
	userGroupIDGenerator, err := uniqueNumberRegistry.GetUniqueNumberGen("userGroupID")
	if err != nil {
		return Authorization{}, err.ToError()
	}

	return Authorization{
		logger:               logger,
		resourceRelationDao:  resourceRelationDao,
		userGroupMemberDao:   userGroupMemberDao,
		permissionDao:        permissionDao,
		operationRelationDao: operationRelationDao,
		operationDao:         operationDao,
		resourceTypeDao:      resourceTypeDao,
		resourceDao:          resourceDao,
		userGroupDao:         userGroupDao,
		userGroupIDGenerator: userGroupIDGenerator,
	}, nil
}
