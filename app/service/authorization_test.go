package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/dao/dao_test"
	"github.com/teamyapp/cloud/app/entity"
)

func TestAuthorization_HasPermission(t *testing.T) {
	fakeResourceDao := []entity.ResourceRelation{
		{
			ChileResourceID:    1,
			ChildResourceType:  "task",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:    2,
			ChildResourceType:  "task",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:    1,
			ChildResourceType:  "invitation",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:    3,
			ChildResourceType:  "task",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:    4,
			ChildResourceType:  "task",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:    2,
			ChildResourceType:  "invitation",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ChileResourceID:   1,
			ChildResourceType: "team",
		},
		{
			ChileResourceID:   2,
			ChildResourceType: "team",
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
	}

	fakePermissionDao := []entity.Permission{
		{
			ResourceType: "team",
			ResourceID:   1,
			Operation:    "update",
			GroupID:      2,
		},
		{
			ResourceType: "task",
			ResourceID:   1,
			Operation:    "update",
			GroupID:      1,
		},
		{
			ResourceType: "task",
			ResourceID:   2,
			Operation:    "update",
			GroupID:      1,
		},
		{
			ResourceType: "team",
			ResourceID:   2,
			Operation:    "update",
			GroupID:      4,
		},
		{
			ResourceType: "task",
			ResourceID:   3,
			Operation:    "update",
			GroupID:      3,
		},
		{
			ResourceType: "task",
			ResourceID:   4,
			Operation:    "update",
			GroupID:      3,
		},
	}

	fakeResourceOperationDao := []entity.OperationRelation{
		{
			ChildResourceType:  "invitation",
			ChildOperation:     "read",
			ParentResourceType: "invitation",
			ParentOperation:    "send",
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
			ParentResourceType: "task",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
		{
			ChildResourceType:  "task",
			ChildOperation:     "update",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ChildResourceType:  "invitation",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
		{
			ChildResourceType:  "team",
			ChildOperation:     "read",
			ParentResourceType: "team",
			ParentOperation:    "update",
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
			name:                  "has permission by traversing permission table",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "has no permission by traversing permission table",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                3,
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
				operationRelationDao: dao_test.NewOperationRelation(fakeResourceOperationDao),
				resourceRelationDao:  dao_test.NewResourceRelation(fakeResourceDao),
			}

			hasPermission, err := mockAuthorization.HasPermission(testCase.resourceType, testCase.resourceID, testCase.operation, testCase.userID)
			assert.Nil(t, err)
			assert.Equal(t, hasPermission, testCase.expectedHasPermission)
		})
	}
}
