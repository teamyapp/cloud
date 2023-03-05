package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/dao/daotest"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var dataCollector = telemetry.NewDataCollector(
	telemetry.NewLogger(
		telemetry.NewOrderedColumnLineFormatter([]string{}),
		os.Stdout,
		telemetry.Info,
		[]telemetry.LogInterceptor{},
	))

func TestAuthorization_HasPermission(t *testing.T) {
	resourceRelations := []entity.ResourceRelation{
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

	userGroupMembers := []entity.UserGroupMember{
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

	permissions := []entity.Permission{
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

	operationRelations := []entity.OperationRelation{
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

	allocatedRanges := []entity.AllocatedRange{
		{
			Key:        "userGroupID",
			RangeEnd:   0,
			NextNumber: 1,
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
			name:                  "current permission found",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "current permission not found but parent permission found",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                3,
			expectedHasPermission: false,
		},
		{
			name:                  "operation and resource type of parent permission change and permission found",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "operation and resource type of parent permission change but no permission found",
			resourceType:          "task",
			resourceID:            6,
			operation:             "update",
			userID:                1,
			expectedHasPermission: false,
		},
		{
			name:                  "operation and resource type is same and one resource is parent of other resource and permission found",
			resourceType:          "task",
			resourceID:            5,
			operation:             "read",
			userID:                5,
			expectedHasPermission: true,
		},
		{
			name:                  "one resource type has multiple same level parent resource types and permission found",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "one resource type has multiple same level parent resource types but no permission found",
			resourceType:          "task",
			resourceID:            1,
			operation:             "update",
			userID:                2,
			expectedHasPermission: false,
		},
		{
			name:                  "permission found for user who is in a owner group",
			resourceType:          "team",
			resourceID:            1,
			operation:             "read",
			userID:                2,
			expectedHasPermission: true,
		},
		{
			name:                  "no permission for user who is outside of a owner group",
			resourceType:          "task",
			resourceID:            1,
			operation:             "read",
			userID:                4,
			expectedHasPermission: false,
		},
	}
	ct := context.Background()
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			inMemoryDB := daotest.NewInMemoryDB()
			inMemoryDB.CreateTable(daotest.OperationTableName)
			inMemoryDB.CreateTable(daotest.ResourceTypeTableName)
			inMemoryDB.CreateTable(daotest.ResourceTableName)
			inMemoryDB.CreateTable(daotest.UserGroupTableName)
			inMemoryDB.InitTable(
				daotest.AllocatedRangeTableName,
				collect.Map(allocatedRanges, func(allocatedRange entity.AllocatedRange, index int) interface{} {
					return allocatedRange
				}))
			inMemoryDB.InitTable(
				daotest.ResourceRelationTableName,
				collect.Map(resourceRelations, func(resourceRelation entity.ResourceRelation, index int) interface{} {
					return resourceRelation
				}))
			inMemoryDB.InitTable(
				daotest.UserGroupMemberTableName,
				collect.Map(userGroupMembers, func(userGroupMember entity.UserGroupMember, index int) interface{} {
					return userGroupMember
				}))
			inMemoryDB.InitTable(
				daotest.PermissionTableName,
				collect.Map(permissions, func(permission entity.Permission, index int) interface{} {
					return permission
				}))
			inMemoryDB.InitTable(
				daotest.OperationRelationTableName,
				collect.Map(operationRelations, func(operationRelation entity.OperationRelation, index int) interface{} {
					return operationRelation
				}))

			mockAllocatedRange := daotest.NewAllocatedRange(inMemoryDB)
			mockAuthorization, err := NewAuthorization(
				dataCollector,
				daotest.NewResourceRelation(inMemoryDB),
				daotest.NewUserGroupMember(inMemoryDB),
				daotest.NewPermission(inMemoryDB),
				daotest.NewOperationRelation(inMemoryDB),
				daotest.NewOperation(inMemoryDB),
				daotest.NewResourceType(inMemoryDB),
				daotest.NewResource(inMemoryDB),
				daotest.NewUserGroup(inMemoryDB),
				gen.NewUniqueNumberFactory(dataCollector, mockAllocatedRange, 0),
			)
			assert.Nil(t, err)

			hasPermission, internalErr := mockAuthorization.HasPermission(ct, testCase.resourceType, testCase.resourceID, testCase.operation, testCase.userID)
			assert.Nil(t, internalErr)
			assert.Equal(t, hasPermission, testCase.expectedHasPermission)
		})
	}
}
