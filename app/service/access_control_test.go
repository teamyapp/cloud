package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/repo"
	"github.com/teamyapp/cloud/app/repo/repo_testing"
)

type permissionGraphEdge struct {
	from entity.Permission
	to   entity.Permission
}

type resourceGraphEdge struct {
	from entity.Resource
	to   entity.Resource
}

func TestAccessControl_HasPermission(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		permissionGraphEdges []permissionGraphEdge
		resourceGraphEdges   []resourceGraphEdge
		permissionBindings   []entity.PermissionBinding
		query                entity.AuthorizationQuery
		expHasPermission     bool
		expHasError          bool
	}{
		{
			name: "empty permission type in query",
			query: entity.AuthorizationQuery{
				PermissionType: "",
				ResourceId:     "resource-id",
				ResourceType:   "resource-type",
				UserOrGroupId:  "user-id",
			},
			expHasPermission: false,
			expHasError:      true,
		},
		{
			name: "empty resource id in query",
			query: entity.AuthorizationQuery{
				PermissionType: "permission-type",
				ResourceId:     "",
				ResourceType:   "resource-type",
				UserOrGroupId:  "user-id",
			},
			expHasPermission: false,
			expHasError:      true,
		},
		{
			name: "empty resource type in query",
			query: entity.AuthorizationQuery{
				PermissionType: "permission-type",
				ResourceId:     "resource-id",
				ResourceType:   "",
				UserOrGroupId:  "user-id",
			},
			expHasPermission: false,
			expHasError:      true,
		},
		{
			name: "empty user or group ID in query",
			query: entity.AuthorizationQuery{
				PermissionType: "permission-type",
				ResourceId:     "resource-id",
				ResourceType:   "resource-type",
				UserOrGroupId:  "",
			},
			expHasPermission: false,
			expHasError:      true,
		},
		{
			name: "explicit permission binding exists",
			permissionGraphEdges: []permissionGraphEdge{
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "read",
						ResourceType:   "team",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "team",
					},
				},
			},
			resourceGraphEdges: []resourceGraphEdge{
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "team-1",
						Type: "team",
					},
				},
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
				},
			},
			permissionBindings: []entity.PermissionBinding{
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "write",
					ResourceId:     "project-1",
					ResourceType:   "project",
					UserOrGroupId:  "user-2",
				},
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-3",
				},
			},
			query: entity.AuthorizationQuery{
				PermissionType: "read",
				ResourceId:     "team-1",
				ResourceType:   "team",
				UserOrGroupId:  "user-1",
			},
			expHasPermission: true,
			expHasError:      false,
		},
		{
			name: "no explicit or implicit permission binding",
			permissionGraphEdges: []permissionGraphEdge{
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "read",
						ResourceType:   "team",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "team",
					},
				},
			},
			resourceGraphEdges: []resourceGraphEdge{
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "team-1",
						Type: "team",
					},
				},
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
				},
			},
			permissionBindings: []entity.PermissionBinding{
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "write",
					ResourceId:     "project-1",
					ResourceType:   "project",
					UserOrGroupId:  "user-2",
				},
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-3",
				},
			},
			query: entity.AuthorizationQuery{
				PermissionType: "read",
				ResourceId:     "team-1",
				ResourceType:   "team",
				UserOrGroupId:  "user-2",
			},
			expHasPermission: false,
			expHasError:      false,
		},
		{
			name: "implicitly permission binding exists",
			permissionGraphEdges: []permissionGraphEdge{
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "read",
						ResourceType:   "team",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "team",
					},
				},
			},
			resourceGraphEdges: []resourceGraphEdge{
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "team-1",
						Type: "team",
					},
				},
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
				},
			},
			permissionBindings: []entity.PermissionBinding{
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "write",
					ResourceId:     "project-1",
					ResourceType:   "project",
					UserOrGroupId:  "user-2",
				},
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-3",
				},
			},
			query: entity.AuthorizationQuery{
				PermissionType: "read",
				ResourceId:     "project-1",
				ResourceType:   "project",
				UserOrGroupId:  "user-2",
			},
			expHasPermission: true,
			expHasError:      false,
		},
		{
			name: "implicit permission binding exists for different resource type",
			permissionGraphEdges: []permissionGraphEdge{
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "read",
						ResourceType:   "team",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "team",
					},
				},
			},
			resourceGraphEdges: []resourceGraphEdge{
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "team-1",
						Type: "team",
					},
				},
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
				},
			},
			permissionBindings: []entity.PermissionBinding{
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "project-1",
					ResourceType:   "project",
					UserOrGroupId:  "user-2",
				},
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-3",
				},
			},
			query: entity.AuthorizationQuery{
				PermissionType: "read",
				ResourceId:     "project-1",
				ResourceType:   "project",
				UserOrGroupId:  "user-1",
			},
			expHasPermission: true,
			expHasError:      false,
		},
		{
			name: "implicit permission binding exists for different resource type in second level",
			permissionGraphEdges: []permissionGraphEdge{
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "read",
						ResourceType:   "team",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "read",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
				},
				{
					from: entity.Permission{
						PermissionType: "write",
						ResourceType:   "project",
					},
					to: entity.Permission{
						PermissionType: "write",
						ResourceType:   "team",
					},
				},
			},
			resourceGraphEdges: []resourceGraphEdge{
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "team-1",
						Type: "team",
					},
				},
				{
					from: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
					to: entity.Resource{
						Id:   "project-1",
						Type: "project",
					},
				},
			},
			permissionBindings: []entity.PermissionBinding{
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-1",
				},
				{
					PermissionType: "read",
					ResourceId:     "project-1",
					ResourceType:   "project",
					UserOrGroupId:  "user-2",
				},
				{
					PermissionType: "write",
					ResourceId:     "team-1",
					ResourceType:   "team",
					UserOrGroupId:  "user-3",
				},
			},
			query: entity.AuthorizationQuery{
				PermissionType: "read",
				ResourceId:     "project-1",
				ResourceType:   "project",
				UserOrGroupId:  "user-3",
			},
			expHasPermission: true,
			expHasError:      false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			accessControlService := NewAccessControl(
				setUpPermissionBindingRepo(testCase.permissionBindings),
				setUpPermissionGraphRepo(testCase.permissionGraphEdges),
				setUpResourceGraphRepo(testCase.resourceGraphEdges),
			)

			actualHasPermission, err := accessControlService.HasPermission(testCase.query)
			if testCase.expHasError {
				assert.NotEqual(t, nil, err)
				return
			}
			assert.Equal(t, testCase.expHasPermission, actualHasPermission)
		})
	}
}

func setUpPermissionBindingRepo(permissionBindings []entity.PermissionBinding) repo.PermissionBinding {
	permissionBindingRepo := repo_testing.NewFakePermissionBinding()
	for _, permissionBinding := range permissionBindings {
		permissionBindingRepo.AddPermissionBinding(permissionBinding)
	}
	return &permissionBindingRepo
}

func setUpPermissionGraphRepo(edges []permissionGraphEdge) repo.PermissionGraph {
	permissionGraph := repo_testing.NewFakePermissionGraph()
	for _, edge := range edges {
		permissionGraph.AddNeighbour(edge.from, edge.to)
	}
	return &permissionGraph
}

func setUpResourceGraphRepo(edges []resourceGraphEdge) repo.ResourceGraph {
	resourceGraph := repo_testing.NewFakeResourceGraph()
	for _, edge := range edges {
		resourceGraph.AddNeighbour(edge.from, edge.to)
	}
	return &resourceGraph
}
