package authorization

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/delta"
)

func TestDetectConfigDelta(t *testing.T) {
	testCases := []struct {
		name          string
		oldRawConfig  string
		newRawConfig  string
		expectedDelta delta.Delta[ConfigDelta]
	}{
		{
			name:         "empty configs",
			oldRawConfig: "",
			newRawConfig: "",
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UnchangedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[OperationRelationsDelta]{},
					},
				},
			},
		},
		{
			name:         "add resource type operations",
			oldRawConfig: "",
			newRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Update
      - Delete
`,
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{
							"Task": {
								KeyStatus:   delta.AddedStatus,
								ValueStatus: delta.AddedStatus,
								Value: ResourceTypeOperationsDelta{
									ResourceType: "Task",
									OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
										Status: delta.AddedStatus,
										Value: map[string]delta.KeyValueDelta[bool]{
											"Read": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value:       true,
											},
											"Update": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value:       true,
											},
											"Delete": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value:       true,
											},
										},
									},
								},
							},
						},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[OperationRelationsDelta]{},
					},
				},
			},
		},
		{
			name: "remove resource type operations",
			oldRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Update
      - Delete
`,
			newRawConfig: "",
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{
							"Task": {
								KeyStatus:   delta.RemovedStatus,
								ValueStatus: delta.RemovedStatus,
								Value: ResourceTypeOperationsDelta{
									ResourceType: "Task",
									OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[bool]{
											"Read": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
											"Update": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
											"Delete": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
										},
									},
								},
							},
						},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[OperationRelationsDelta]{},
					},
				},
			},
		},
		{
			name: "update resource type operations",
			oldRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Delete
      - Operate
`,
			newRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Update
      - Delete
`,
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{
							"Task": {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: ResourceTypeOperationsDelta{
									ResourceType: "Task",
									OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[bool]{
											"Read": {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value:       true,
											},
											"Update": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value:       true,
											},
											"Delete": {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value:       true,
											},
											"Operate": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
										},
									},
								},
							},
						},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[OperationRelationsDelta]{},
					},
				},
			},
		},
		{
			name:         "add operation relations",
			oldRawConfig: "",
			newRawConfig: `
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Delete
`,
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[OperationRelationsDelta]{
							path.Join("Task", "Read"): {
								KeyStatus:   delta.AddedStatus,
								ValueStatus: delta.AddedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Read",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.AddedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Task", "Update"): {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Update",
												},
											},
											path.Join("Task", "Delete"): {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Delete",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "remove operation relations",
			oldRawConfig: `
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Delete
`,
			newRawConfig: "",
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[OperationRelationsDelta]{
							"Task/Read": {
								KeyStatus:   delta.RemovedStatus,
								ValueStatus: delta.RemovedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Read",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Task", "Update"): {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Update",
												},
											},
											path.Join("Task", "Delete"): {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Delete",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "update operation relations",
			oldRawConfig: `
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Delete
`,
			newRawConfig: `
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Operate
`,
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UnchangedStatus,
						Value:  map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[OperationRelationsDelta]{
							path.Join("Task", "Read"): {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Read",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Task", "Update"): {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Update",
												},
											},
											path.Join("Task", "Operate"): {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Operate",
												},
											},
											path.Join("Task", "Delete"): {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Delete",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "complex authorization config changes",
			oldRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Delete
  - resourceType: TaskLink
    operations:
      - Read
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Delete
  - resourceType: Task
    operation: Update
    parents:
      - resourceType: Team
        operation: UpdateTask
`,
			newRawConfig: `
resourceTypeOperations:
  - resourceType: Task
    operations:
      - Read
      - Operate
operationRelations:
  - resourceType: Task
    operation: Read
    parents:
      - resourceType: Task
        operation: Update
      - resourceType: Task
        operation: Operate
  - resourceType: Task
    operation: Manage
    parents:
      - resourceType: Team
        operation: ManageTask
`,
			expectedDelta: delta.Delta[ConfigDelta]{
				Status: delta.UpdatedStatus,
				Value: ConfigDelta{
					ResourceTypeOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]{
							"Task": {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: ResourceTypeOperationsDelta{
									ResourceType: "Task",
									OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[bool]{
											"Read": {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value:       true,
											},
											"Operate": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value:       true,
											},
											"Delete": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
										},
									},
								},
							},
							"TaskLink": {
								KeyStatus:   delta.RemovedStatus,
								ValueStatus: delta.RemovedStatus,
								Value: ResourceTypeOperationsDelta{
									ResourceType: "TaskLink",
									OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[bool]{
											"Read": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value:       true,
											},
										},
									},
								},
							},
						},
					},
					OperationRelationsDelta: delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]{
						Status: delta.UpdatedStatus,
						Value: map[string]delta.KeyValueDelta[OperationRelationsDelta]{
							path.Join("Task", "Read"): {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Read",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Task", "Update"): {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Update",
												},
											},
											path.Join("Task", "Operate"): {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Operate",
												},
											},
											path.Join("Task", "Delete"): {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: Operation{
													ResourceType: "Task",
													Operation:    "Delete",
												},
											},
										},
									},
								},
							},
							path.Join("Task", "Update"): {
								KeyStatus:   delta.RemovedStatus,
								ValueStatus: delta.RemovedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Update",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Team", "UpdateTask"): {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: Operation{
													ResourceType: "Team",
													Operation:    "UpdateTask",
												},
											},
										},
									},
								},
							},
							path.Join("Task", "Manage"): {
								KeyStatus:   delta.AddedStatus,
								ValueStatus: delta.AddedStatus,
								Value: OperationRelationsDelta{
									ChildResourceType: "Task",
									ChildOperation:    "Manage",
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
										Status: delta.AddedStatus,
										Value: map[string]delta.KeyValueDelta[Operation]{
											path.Join("Team", "ManageTask"): {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: Operation{
													ResourceType: "Team",
													Operation:    "ManageTask",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			oldConfig, err := ParseConfig(testCase.oldRawConfig)
			if err != nil {
				t.Fatal(err)
			}

			newConfig, err := ParseConfig(testCase.newRawConfig)
			if err != nil {
				t.Fatal(err)
			}

			actualDelta := DetectConfigDelta(oldConfig, newConfig)
			require.Equal(t, testCase.expectedDelta, actualDelta)
		})
	}
}
