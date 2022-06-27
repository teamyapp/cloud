package service

import (
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Authorization struct {
	permissionDao        dao.Permission
	userGroupMemberDao   dao.UserGroupMember
	operationRelationDao dao.OperationRelation
	resourceRelationDao  dao.ResourceRelation
}

func (a Authorization) HasPermission(resourceType string, resourceID uint64, operation string, userID uint64) (bool, error) {
	// No nested group
	groupIDs, err := a.userGroupMemberDao.FindGroupIDsByUserID(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	for _, groupID := range groupIDs {
		hasPermission, err := a.groupHasPermission(entity.PermissionQuery{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			Operation:    operation,
			GroupID:      groupID,
		})
		if err != nil {
			log.Println(err)
			// user should continue to find permission in other groups if current group fails to grant permission
			continue
		}

		if hasPermission {
			return hasPermission, nil
		}
	}

	return false, err
}

func (a Authorization) groupHasPermission(permissionQuery entity.PermissionQuery) (bool, error) {
	groupId := permissionQuery.GroupID
	queries := []entity.PermissionQuery{permissionQuery}
	visited := make(map[entity.PermissionQuery]bool)
	visited[permissionQuery] = true
	for len(queries) > 0 {
		currPermissionQuery := queries[0]
		queries = queries[1:]

		_, err := a.permissionDao.FindPermission(currPermissionQuery)
		if err == nil {
			return true, nil
		}

		log.Println(err)
		parentOperations, err := a.operationRelationDao.FindParentOperations(entity.OperationRelation{
			ResourceType: currPermissionQuery.ResourceType,
			Operation:    currPermissionQuery.Operation,
		})
		if err != nil {
			log.Println(err)
			return false, err
		}

		parentResources, err := a.resourceRelationDao.FindParentResources(entity.ResourceRelation{
			ID:           currPermissionQuery.ResourceID,
			ResourceType: currPermissionQuery.ResourceType,
		})
		if err != nil {
			log.Println(err)
			return false, err
		}

		for _, parentOperation := range parentOperations {
			if parentOperation.ResourceType == currPermissionQuery.ResourceType {
				newPermissionQuery := entity.PermissionQuery{
					ResourceID:   currPermissionQuery.ResourceID,
					ResourceType: parentOperation.ResourceType,
					Operation:    parentOperation.Operation,
					GroupID:      groupId,
				}

				_, ok := visited[newPermissionQuery]
				if ok {
					continue
				}

				visited[newPermissionQuery] = true
				queries = append(queries, newPermissionQuery)
				continue
			}

			for _, parentResource := range parentResources {
				if parentOperation.ResourceType != parentResource.ResourceType {
					continue
				}

				newPermissionQuery := entity.PermissionQuery{
					ResourceID:   parentResource.ID,
					ResourceType: parentOperation.ResourceType,
					Operation:    parentOperation.Operation,
					GroupID:      groupId,
				}
				_, ok := visited[newPermissionQuery]
				if ok {
					continue
				}

				visited[newPermissionQuery] = true
				queries = append(queries, newPermissionQuery)
			}
		}
	}

	return false, nil
}
