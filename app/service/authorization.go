package service

import (
	"log"

	"github.com/teamyapp/cloud/app/dao"

	"github.com/teamyapp/cloud/app/entity"
)

type Authorization struct {
	permissionDao        dao.Permission
	securityGroupUserDao dao.SecurityGroupUser
	resourceOperationDao dao.ResourceOperation
	resourceDao          dao.Resource
}

func (a Authorization) HasPermission(resourceType string, resourceID uint64, operation string, userID uint64) (bool, error) {
	// No nested group
	groupIDs, err := a.securityGroupUserDao.FindGroupIDsByUserID(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	for _, groupID := range groupIDs {
		canFind, err := a.searchPermission(entity.PermissionQuery{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			Operation:    operation,
			GroupID:      groupID,
		})
		if err != nil {
			log.Println(err)
		}

		if canFind {
			return canFind, nil
		}
	}

	return false, err
}

func (a Authorization) searchPermission(permissionQuery entity.PermissionQuery) (bool, error) {
	groupId := permissionQuery.GroupID

	s := make([]entity.PermissionQuery, 0)
	s = append(s, permissionQuery)
	visited := make(map[entity.PermissionQuery]int)
	visited[permissionQuery] = 1

	for len(s) > 0 {
		currPermissionQuery := s[0]
		s = s[1:]

		canFind, err := a.permissionDao.FindPermission(currPermissionQuery)
		if err != nil {
			log.Println(err)
			return false, err
		}

		if canFind {
			return canFind, nil
		}

		parentResourceOperations, err := a.resourceOperationDao.GetAllParentResourceOperations(entity.ResourceOperation{
			ResourceType: currPermissionQuery.ResourceType,
			Operation:    currPermissionQuery.Operation,
		})
		if err != nil {
			log.Println(err)
			return false, err
		}

		parentResources, err := a.resourceDao.FindParentResources(entity.Resource{
			ID:           currPermissionQuery.ResourceID,
			ResourceType: currPermissionQuery.ResourceType,
		})
		if err != nil {
			log.Println(err)
			return false, err
		}

		for _, parentResourceOperation := range parentResourceOperations {
			for _, parentResource := range parentResources {
				// e.g. read task 1 -> update task 1
				if parentResourceOperation.ResourceType == currPermissionQuery.ResourceType {
					newPermissionQuery := entity.PermissionQuery{
						ResourceID:   currPermissionQuery.ResourceID,
						ResourceType: currPermissionQuery.ResourceType,
						Operation:    parentResourceOperation.Operation,
						GroupID:      groupId,
					}

					_, ok := visited[newPermissionQuery]
					if ok {
						continue
					}

					s = append(s, newPermissionQuery)
					visited[newPermissionQuery] = 1
				}

				// e.g. read task 1 -> read team 1
				if parentResourceOperation.Operation == currPermissionQuery.Operation {
					newPermissionQuery := entity.PermissionQuery{
						ResourceID:   parentResource.ID,
						ResourceType: parentResourceOperation.ResourceType,
						Operation:    parentResourceOperation.Operation,
						GroupID:      groupId,
					}

					_, ok := visited[newPermissionQuery]
					if ok {
						continue
					}

					s = append(s, newPermissionQuery)
					visited[newPermissionQuery] = 1
				}
			}
		}
	}

	return false, nil
}
