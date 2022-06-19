package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/dao/dao_test"
	"github.com/teamyapp/cloud/app/entity"
)

func TestAuthorization_HasPermission(t *testing.T) {
	fakeResourceDao := []entity.Resource{
		{
			ID:                 1,
			ResourceType:       "task",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ID:                 2,
			ResourceType:       "task",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ID:                 1,
			ResourceType:       "invitation",
			ParentResourceID:   1,
			ParentResourceType: "team",
		},
		{
			ID:                 3,
			ResourceType:       "task",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ID:                 4,
			ResourceType:       "task",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ID:                 2,
			ResourceType:       "invitation",
			ParentResourceID:   2,
			ParentResourceType: "team",
		},
		{
			ID:           1,
			ResourceType: "team",
		},
		{
			ID:           2,
			ResourceType: "team",
		},
	}

	fakeSecurityGroupUserDao := []entity.SecurityGroupUser{
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

	fakeResourceOperationDao := []entity.ResourceOperation{
		{
			ResourceType:       "invitation",
			Operation:          "read",
			ParentResourceType: "invitation",
			ParentOperation:    "send",
		},
		{
			ResourceType:       "invitation",
			Operation:          "send",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ResourceType:       "task",
			Operation:          "read",
			ParentResourceType: "task",
			ParentOperation:    "update",
		},
		{
			ResourceType:       "task",
			Operation:          "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
		{
			ResourceType:       "task",
			Operation:          "update",
			ParentResourceType: "team",
			ParentOperation:    "update",
		},
		{
			ResourceType:       "invitation",
			Operation:          "read",
			ParentResourceType: "team",
			ParentOperation:    "read",
		},
		{
			ResourceType:       "team",
			Operation:          "read",
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
				securityGroupUserDao: dao_test.NewSecurityGroupUser(fakeSecurityGroupUserDao),
				resourceOperationDao: dao_test.NewResourceOperation(fakeResourceOperationDao),
				resourceDao:          dao_test.NewResource(fakeResourceDao),
			}

			hasPermission, err := mockAuthorization.HasPermission(testCase.resourceType, testCase.resourceID, testCase.operation, testCase.userID)
			assert.Nil(t, err)
			assert.Equal(t, hasPermission, testCase.expectedHasPermission)
		})
	}
}
