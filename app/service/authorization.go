package service

import (
	"context"
	"errors"
	"time"

	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/libs/obs"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/ctx"
)

type Authorization struct {
	dataCollector        obs.DataCollector
	resourceRelationDao  dao.ResourceRelation
	userGroupMemberDao   dao.UserGroupMember
	permissionDao        dao.Permission
	operationRelationDao dao.OperationRelation
	operationDao         dao.Operation
	resourceTypeDao      dao.ResourceType
	resourceDao          dao.Resource
	userGroupDao         dao.UserGroup
	userGroupIDGenerator *gen.UniqueNumber
}

func (a Authorization) HasPermission(ct context.Context, resourceType string, resourceID uint64, operation string, userID uint64) (bool, error) {
	// No nested group allowed
	groupIDs, err := a.userGroupMemberDao.FindGroupIDsByUserID(ct, userID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
			a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			// Continue check permission in other groups if current group fails to grant permission
			continue
		}

		if hasPermission {
			return hasPermission, nil
		}
	}

	return false, nil
}

func (a Authorization) ListResourceTypes(ct context.Context, resourceTypeQuery ResourceTypeQuery) ([]entity.ResourceType, error) {
	allResourceTypeEntities, err := a.resourceTypeDao.FindAllResourceTypes(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return queryResourceTypes(allResourceTypeEntities, resourceTypeQuery), nil
}

func (a Authorization) RegisterResourceType(ct context.Context, resourceTypeName string) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	resourceTypeEntity := entity.ResourceType{
		ResourceTypeName: resourceTypeName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}

	return a.resourceTypeDao.CreateResourceType(ct, resourceTypeEntity)
}

func (a Authorization) UnregisterResourceType(ct context.Context, resourceTypeName string) error {
	return a.resourceTypeDao.DeleteResourceType(ct, resourceTypeName)
}

func (a Authorization) ListResources(ct context.Context, resourceQuery ResourceQuery) ([]entity.Resource, error) {
	allResources, err := a.resourceDao.FindAllResources(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return queryResources(allResources, resourceQuery), nil
}

func (a Authorization) RegisterResource(ct context.Context, resourceTypeName string, resourceID uint64) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	resource := entity.Resource{
		ResourceTypeName: resourceTypeName,
		ResourceID:       resourceID,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.resourceDao.CreateResource(ct, resource)
}

func (a Authorization) UnregisterResource(ct context.Context, resourceTypeName string, resourceID uint64) error {
	return a.resourceDao.DeleteResource(ct, resourceTypeName, resourceID)
}

func (a Authorization) ListResourceRelations(ct context.Context, resourceRelationQuery ResourceRelationQuery) ([]entity.ResourceRelation, error) {
	allResourceRelationEntities, err := a.resourceRelationDao.FindAllResourceRelations(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
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
) error {
	return a.resourceRelationDao.DeleteResourceRelation(
		ct,
		childResourceType,
		childResourceID,
		parentResourceType,
		parentResourceID,
	)
}

func (a Authorization) ListOperations(ct context.Context, operationQuery OperationQuery) ([]entity.Operation, error) {
	allOperations, err := a.operationDao.FindAllOperations(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return queryOperations(allOperations, operationQuery), nil
}

func (a Authorization) RegisterOperation(ct context.Context, resourceTypeName string, operationName string) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	operation := entity.Operation{
		ResourceTypeName: resourceTypeName,
		OperationName:    operationName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.operationDao.CreateOperation(ct, operation)
}

func (a Authorization) UnregisterOperation(ct context.Context, resourceTypeName string, operationName string) error {
	return a.operationDao.DeleteOperation(ct, resourceTypeName, operationName)
}

func (a Authorization) ListOperationRelations(ct context.Context, operationRelationQuery OperationRelationQuery) ([]entity.OperationRelation, error) {
	allOperationRelations, err := a.operationRelationDao.FindAllOperationRelations(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
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
) error {
	return a.operationRelationDao.DeleteOperationRelation(
		ct,
		childResourceType,
		childOperation,
		parentResourceType,
		parentOperation,
	)
}

func (a Authorization) ListUserGroups(ct context.Context, query UserGroupQuery) ([]entity.UserGroup, error) {
	allGroups, err := a.userGroupDao.FindAllGroups(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return queryUserGroups(allGroups, query), nil
}

func (a Authorization) CreateUserGroup(ct context.Context, name string, description *string) (entity.UserGroup, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.UserGroup{}, err
	}

	groupID, err := a.userGroupIDGenerator.GenerateUniqueNumber(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.UserGroup{}, err
	}

	userGroup := entity.UserGroup{
		ID:            groupID,
		Name:          name,
		Description:   description,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: userID,
	}

	createdUserGroup, err := a.userGroupDao.CreateGroup(ct, userGroup)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.UserGroup{}, err
	}

	return createdUserGroup, nil
}

func (a Authorization) UpdateUserGroup(ct context.Context, groupID uint64, name *string, description *string) error {
	userGroup, err := a.userGroupDao.FindGroupByID(ct, groupID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func (a Authorization) DeleteUserGroup(ct context.Context, groupID uint64) error {
	return a.userGroupDao.DeleteGroup(ct, groupID)
}

func (a Authorization) ListUserGroupMembers(ct context.Context, query UserGroupMemberQuery) ([]entity.UserGroupMember, error) {
	allUserGroupMembers, err := a.userGroupMemberDao.FindAllUserGroupMembers(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	userGroupMembers := queryUserGroupMembers(allUserGroupMembers, query)
	return userGroupMembers, nil
}

func (a Authorization) AddUserGroupMember(ct context.Context, groupID uint64, userID uint64) error {
	creatorUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	userGroupMember := entity.UserGroupMember{
		GroupID:       groupID,
		UserID:        userID,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: creatorUserID,
	}
	return a.userGroupMemberDao.CreateUserGroupMember(ct, userGroupMember)
}

func (a Authorization) RemoveUserGroupMember(ct context.Context, groupID uint64, userID uint64) error {
	return a.userGroupMemberDao.DeleteUserGroupMember(ct, groupID, userID)
}

func (a Authorization) ListPermissions(ct context.Context, query PermissionQuery) ([]entity.Permission, error) {
	allPermissions, err := a.permissionDao.FindAllPermissions(ct)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	permissions := queryPermission(allPermissions, query)
	return permissions, nil
}

func (a Authorization) AddPermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) error {
	creatorUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
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

func (a Authorization) RemovePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) error {
	return a.permissionDao.DeletePermission(ct, resourceType, resourceID, operation, groupID)
}

func (a Authorization) groupHasPermission(ct context.Context, permissionQuery entity.PermissionQuery) (bool, error) {
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

		var errNotFound dao.ErrNotFound
		if !errors.As(err, &errNotFound) {
			a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		parentPermissionQueries, err := a.getParentPermissionQueries(ct, currQuery, visited)
		if err != nil {
			return false, err
		}

		queries = append(queries, parentPermissionQueries...)
	}

	return false, nil
}

func (a Authorization) getParentPermissionQueries(ct context.Context, currQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool) ([]entity.PermissionQuery, error) {
	var parentPermissionQueries []entity.PermissionQuery
	operationRelations, err := a.operationRelationDao.FindOperationRelations(ct, currQuery.ResourceType, currQuery.Operation)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	resourceRelations, err := a.resourceRelationDao.FindResourceRelations(ct, currQuery.ResourceType, currQuery.ResourceID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func tryAddPermissionQueryToQueue(permissionQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool, queries []entity.PermissionQuery) []entity.PermissionQuery {
	_, ok := visited[permissionQuery]
	if ok {
		return queries
	}

	visited[permissionQuery] = true
	return append(queries, permissionQuery)
}

func NewAuthorization(
	dataCollector obs.DataCollector,
	resourceRelationDao dao.ResourceRelation,
	userGroupMemberDao dao.UserGroupMember,
	permissionDao dao.Permission,
	operationRelationDao dao.OperationRelation,
	operationDao dao.Operation,
	resourceTypeDao dao.ResourceType,
	resourceDao dao.Resource,
	userGroupDao dao.UserGroup,
	uniqueNumberFactory gen.UniqueNumberFactory,
) (Authorization, error) {
	userGroupIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("userGroupID")
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Authorization{}, err
	}

	return Authorization{
		dataCollector:        dataCollector,
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
