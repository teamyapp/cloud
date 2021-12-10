package service

import (
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/errs"
	"github.com/teamyapp/cloud/app/repo"
)

type AccessControl struct {
	permissionBindingRepo repo.PermissionBinding
	permissionGraphRepo   repo.PermissionGraph
	resourceGraphRepo     repo.ResourceGraph
}

func (p *AccessControl) HasPermission(query entity.AuthorizationQuery) (bool, error) {
	// TODO: Discuss impact of data inconsistency
	// TODO: Discuss when to pause the bfs
	//          - timeout the query?
	//          - limit the depth?

	if !query.IsValid() {
		return false, errs.InvalidAuthorizationQuery(query)
	}

	queue := make([]entity.AuthorizationQuery, 0)
	visited := make(map[entity.AuthorizationQuery]bool)

	visited[query] = true
	queue = append(queue, query)

	for len(queue) != 0 {
		currQuery := queue[0]
		queue = queue[1:]

		if p.permissionBindingRepo.HasPermissionBinding(currQuery.PermissionBindingFromAuthQuery()) {
			return true, nil
		}

		neighbourPermissions := p.permissionGraphRepo.GetNeighbours(entity.Permission{
			PermissionType: currQuery.PermissionType,
			ResourceType:   currQuery.ResourceType,
		})

		for _, neighbourPermission := range neighbourPermissions {
			neighbourResources := p.resourceGraphRepo.FindNeighboursWithType(entity.Resource{
				Id:   currQuery.ResourceId,
				Type: currQuery.ResourceType,
			}, neighbourPermission.ResourceType)

			for _, neighbourResource := range neighbourResources {
				nextQuery := entity.AuthorizationQuery{
					PermissionType: neighbourPermission.PermissionType,
					ResourceId:     neighbourResource.Id,
					ResourceType:   neighbourResource.Type,
					UserOrGroupId:  query.UserOrGroupId,
				}

				if visited[nextQuery] {
					continue
				}
				visited[nextQuery] = true

				queue = append(queue, nextQuery)
			}
		}
	}

	return false, nil
}

func NewAccessControl(
	permissionBindingRepo repo.PermissionBinding,
	permissionGraphRepo repo.PermissionGraph,
	resourceGraphRepo repo.ResourceGraph,
) AccessControl {
	return AccessControl{
		permissionBindingRepo,
		permissionGraphRepo,
		resourceGraphRepo,
	}
}
