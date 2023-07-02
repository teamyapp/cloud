package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.AddedStatus,
										Value:  "Task",
									},
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "Task",
									},
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Task",
									},
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
							"Task/Read": {
								KeyStatus:   delta.AddedStatus,
								ValueStatus: delta.AddedStatus,
								Value: OperationRelationsDelta{
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.AddedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.AddedStatus,
										Value:  "Read",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.AddedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Task/Update": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Update",
													},
												},
											},
											"Task/Delete": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Delete",
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "Read",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Task/Update": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Update",
													},
												},
											},
											"Task/Delete": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Delete",
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
							"Task/Read": {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: OperationRelationsDelta{
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Read",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Task/Update": {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.UnchangedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.UnchangedStatus,
														Value:  "Update",
													},
												},
											},
											"Task/Operate": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Operate",
													},
												},
											},
											"Task/Delete": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Delete",
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Task",
									},
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
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "TaskLink",
									},
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
							"Task/Read": {
								KeyStatus:   delta.UnchangedStatus,
								ValueStatus: delta.UpdatedStatus,
								Value: OperationRelationsDelta{
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.UnchangedStatus,
										Value:  "Read",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.UpdatedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Task/Update": {
												KeyStatus:   delta.UnchangedStatus,
												ValueStatus: delta.UnchangedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.UnchangedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.UnchangedStatus,
														Value:  "Update",
													},
												},
											},
											"Task/Operate": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Operate",
													},
												},
											},
											"Task/Delete": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Task",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Delete",
													},
												},
											},
										},
									},
								},
							},
							"Task/Update": {
								KeyStatus:   delta.RemovedStatus,
								ValueStatus: delta.RemovedStatus,
								Value: OperationRelationsDelta{
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.RemovedStatus,
										Value:  "Update",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.RemovedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Team/UpdateTask": {
												KeyStatus:   delta.RemovedStatus,
												ValueStatus: delta.RemovedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "Team",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.RemovedStatus,
														Value:  "UpdateTask",
													},
												},
											},
										},
									},
								},
							},
							"Task/Manage": {
								KeyStatus:   delta.AddedStatus,
								ValueStatus: delta.AddedStatus,
								Value: OperationRelationsDelta{
									ResourceTypeDelta: delta.Delta[string]{
										Status: delta.AddedStatus,
										Value:  "Task",
									},
									OperationDelta: delta.Delta[string]{
										Status: delta.AddedStatus,
										Value:  "Manage",
									},
									ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
										Status: delta.AddedStatus,
										Value: map[string]delta.KeyValueDelta[ParentOperationDelta]{
											"Team/ManageTask": {
												KeyStatus:   delta.AddedStatus,
												ValueStatus: delta.AddedStatus,
												Value: ParentOperationDelta{
													ParentResourceTypeDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "Team",
													},
													ParentOperationDelta: delta.Delta[string]{
														Status: delta.AddedStatus,
														Value:  "ManageTask",
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
			assert.Equal(t, testCase.expectedDelta, actualDelta)
		})
	}
}
