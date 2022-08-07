package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/dao/dao_test"
	"github.com/teamyapp/cloud/app/entity"
)

func TestAuthorization_HasPermission(t *testing.T) {
	fakeResourceRelationDao := []entity.ResourceRelation{
		{
			ChildResourceID:   1,
			ChildResourceType: "team",
		},
		{
			ChildResourceID:   2,
			ChildResourceType: "team",
		},
		{
			ChildResourceID:    11,
			ChildResourceType:  "project",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    12,
			ChildResourceType:  "project",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    21,
			ChildResourceType:  "sprint",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    1,
			ChildResourceType:  "invitation",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    23,
			ChildResourceType:  "project",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    4,
			ChildResourceType:  "task",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    2,
			ChildResourceType:  "invitation",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChildResourceID:    1,
			ChildResourceType:  "task",
			ParentResourceID:   11,
			ParentResourceType: "project",
		},
		{
			ChildResourceID:    1,
			ChildResourceType:  "task",
			ParentResourceID:   21,
			ParentResourceType: "sprint",
		},
		{
			ChildResourceID:    1,
			ChildResourceType:  "task",
			ParentResourceID:   12,
			ParentResourceType: "project",
		},
		{
			ChildResourceID:    2,
			ChildResourceType:  "task",
			ParentResourceID:   12,
			ParentResourceType: "project",
		},
		{
			ChildResourceID:    3,
			ChildResourceType:  "task",
			ParentResourceID:   23,
			ParentResourceType: "project",
		},
		{
			ChildResourceID:    6,
			ChildResourceType:  "task",
			ParentResourceID:   11,
			ParentResourceType: "project",
		},
		{
			ChildResourceID:    5,
			ChildResourceType:  "task",
			ParentResourceID:   2,
			ParentResourceType: "task",
		},
	}

	fakeUserGroupMemberDao := []entity.UserGroupMember{
		{
			GroupID: 1,
			UserID:  1,
		},
		{
			GroupID: 2,
			UserID:  2,
		},
		{
			GroupID: 3,
			UserID:  3,
		},
		{
			GroupID: 4,
			UserID:  4,
		},
		{
			GroupID: 1,
			UserID:  5,
		},
		{
			GroupID: 2,
			UserID:  5,
		},
		{
			GroupID: 5,
			UserID:  2,
		},
	}

	fakePermissionDao := []entity.Permission{
		{
			ResourceType: "project",
			ResourceID:   11,
			Operation:    "read",
			GroupID:      1,
		},
		{
			ResourceType: "sprint",
			ResourceID:   21,
			Operation:    "update",
			GroupID:      1,
		},
		{
			ResourceType: "sprint",
			ResourceID:   21,
			Operation:    "read",
			GroupID:      1,
		},
		{
			ResourceType: "project",
			ResourceID:   12,
			Operation:    "read",
			GroupID:      1,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "read",
			GroupID:      1,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "update",
			GroupID:      2,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "updateTask",
			GroupID:      2,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "readTask",
			GroupID:      2,
		},
		{
			ResourceType: "project",
			ResourceID:   23,
			Operation:    "read",
			GroupID:      3,
		},
		{
			ResourceType: "team",
			ResourceID:   2,
			Operation:    "read",
			GroupID:      3,
		},
		{
			ResourceType: "team",
			ResourceID:   2,
			Operation:    "update",
			GroupID:      4,
		},
		{
			ResourceType: "team",
			ResourceID:   2,
			Operation:    "updateTask",
			GroupID:      4,
		},
		{
			ResourceType: "team",
			ResourceID:   2,
			Operation:    "readTask",
			GroupID:      4,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "addUserTo",
			GroupID:      5,
		},
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "removeUserFrom",
			GroupID:      5,
		},
	}

	fakeOperationRelationDao := []entity.OperationRelation{
		{
			ChildResourceType: "team",
			ChildOperation:    "addUserTo",
		},
		{
			ChildResourceType: "team",
			ChildOperation:    "removeUserFrom",
		},
		{
			ChildResourceType:  "team",
			ChildOperation:     "update",
			ParentResourceType: "team",
			ParentOperation:    "addUserTo",
		},
		{
			ChildResourceType:  "team",
			ChildOperation:     "update",
			ParentResourceType: "team",
			ParentOperation:    "removeUserFrom",
		},
		{
			ChildResourceType: "team",
			ChildOperation:    "updateTask",
		},
		{
			ChildResourceType:  "team",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ChildResourceType: "sprint",
			ChildOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "update",
			ParentResourceType: "sprint",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "update",
			ParentResourceType: "team",
			ParentOperation:    "updateTask",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "update",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "project",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
		{
			ChildResourceType: "team",
			ChildOperation:    "readTask",
		},
		{
			ChildResourceType:  "invitation",
			ChildOperation:     "send",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "sprint",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "task",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "project",
			ParentOperation:    "read",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "readTask",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "task",
			ParentOperation:    "read",
		},
		{
			ChildResourceType:  "invitation",
			ChildOperation:     "read",
			ParentResourceType: "invitation",
			ParentOperation:    "send",
		},
		{
			ChildResourceType:  "invitation",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
	}

	testCases := []struct {
		name                  string
		resourceType          string
		resourceID            uint64
		operation             string
		userID                uint64
		expectedHasPermission bool
	}{
		{
			name:                  "Case 0: has permission by traversing permission table",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "Case 0: has no permission by traversing permission table",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                3,
			expectedHasPermission: false,
		},
		{
			name:                  "Case 1: both operation and resource type change happy case",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "Case 1: both operation and resource type change failed case",
			resourceType:          "task",
			resourceID:            6,
			operation:             "update",
			userID:                1,
			expectedHasPermission: false,
		},
		{
			name:                  "Case 2: operation and resource type do not change ",
			resourceType:          "task",
			resourceID:            2,
			operation:             "read",
			userID:                5,
			expectedHasPermission: true,
		},
		{
			name:                  "Case 3: one resource type has multiple same level parent resource types happy case",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "Case 3: one resource type has multiple same level parent resource types failed case",
			resourceType:          "task",
			resourceID:            1,
			operation:             "update",
			userID:                2,
			expectedHasPermission: false,
		},
		{
			name:                  "Case 4: owner group",
			resourceType:          "team",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "Case 5: cross team",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                4,
			expectedHasPermission: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockAuthorization := Authorization{
				permissionDao:        dao_test.NewPermission(fakePermissionDao),
				userGroupMemberDao:   dao_test.NewUserGroupMember(fakeUserGroupMemberDao),
				operationRelationDao: dao_test.NewOperationRelation(fakeOperationRelationDao),
				resourceRelationDao:  dao_test.NewResourceRelation(fakeResourceRelationDao),
			}

			hasPermission, err := mockAuthorization.HasPermission(testCase.resourceType, testCase.resourceID, testCase.operation, testCase.userID)
			assert.Nil(t, err)
			assert.Equal(t, hasPermission, testCase.expectedHasPermission)
		})
	}
}
