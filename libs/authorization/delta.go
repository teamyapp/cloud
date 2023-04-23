package authorization

import (
	"github.com/teamyapp/cloud/libs/delta"
)

type ConfigDelta struct {
	ResourceTypeOperationsDelta delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]
	OperationRelationsDelta     delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]
}

type ResourceTypeOperationsDelta struct {
	ResourceTypeDelta delta.Delta[string]
	OperationsDelta   delta.Delta[map[string]delta.KeyValueDelta[bool]]
}

type OperationRelationsDelta struct {
	ResourceTypeDelta delta.Delta[string]
	Operation         delta.Delta[string]
	ParentOperations  delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]
}

type ParentOperationDelta struct {
	ParentResourceType delta.Delta[string]
	ParentOperation    delta.Delta[string]
}

func DetectConfigDelta(
	oldConfig Config,
	newConfig Config,
) delta.Delta[ConfigDelta] {
	resourceTypeOperationsDelta := delta.DetectMapDelta(
		oldConfig.ResourceTypeOperations,
		newConfig.ResourceTypeOperations,
		detectResourceTypeOperationsDelta,
		toResourceTypeOperationsDelta)
	operationRelationsDelta := delta.DetectMapDelta(
		oldConfig.OperationRelations,
		newConfig.OperationRelations,
		detectOperationRelationsDelta,
		toOperationRelationsDelta)
	configDelta := ConfigDelta{
		ResourceTypeOperationsDelta: resourceTypeOperationsDelta,
		OperationRelationsDelta:     operationRelationsDelta,
	}

	status := delta.UnchangedStatus
	if configDelta.OperationRelationsDelta.Status != delta.UnchangedStatus ||
		configDelta.ResourceTypeOperationsDelta.Status != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[ConfigDelta]{
		Status: status,
		Value:  configDelta,
	}
}

func detectResourceTypeOperationsDelta(
	oldResourceTypeOperations ResourceTypeOperations,
	newResourceTypeOperations ResourceTypeOperations,
) delta.Delta[ResourceTypeOperationsDelta] {
	resourceTypeStatus := delta.UnchangedStatus
	if oldResourceTypeOperations.ResourceType != newResourceTypeOperations.ResourceType {
		resourceTypeStatus = delta.UpdatedStatus
	}

	operationsDeltaMap := delta.DetectMapDelta(
		oldResourceTypeOperations.Operations,
		newResourceTypeOperations.Operations,
		delta.DetectValueDelta[bool],
		delta.ToValueDelta[bool])

	status := delta.UnchangedStatus
	if resourceTypeStatus != delta.UnchangedStatus ||
		operationsDeltaMap.Status != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[ResourceTypeOperationsDelta]{
		Status: status,
		Value: ResourceTypeOperationsDelta{
			ResourceTypeDelta: delta.Delta[string]{
				Status: resourceTypeStatus,
				Value:  newResourceTypeOperations.ResourceType,
			},
			OperationsDelta: operationsDeltaMap,
		},
	}
}

func toResourceTypeOperationsDelta(
	status delta.Status,
	resourceTypeOperations ResourceTypeOperations,
) ResourceTypeOperationsDelta {
	return ResourceTypeOperationsDelta{
		ResourceTypeDelta: delta.Delta[string]{
			Status: status,
			Value:  resourceTypeOperations.ResourceType,
		},
		OperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[bool]]{
			Status: status,
			Value:  delta.ToMapDelta(status, resourceTypeOperations.Operations, delta.ToValueDelta[bool]),
		},
	}
}

func detectOperationRelationsDelta(oldRelations OperationRelations, newRelations OperationRelations) delta.Delta[OperationRelationsDelta] {
	resourceTypeStatus := delta.UnchangedStatus
	if oldRelations.ResourceType != newRelations.ResourceType {
		resourceTypeStatus = delta.UpdatedStatus
	}

	operationStatus := delta.UnchangedStatus
	if oldRelations.Operation != newRelations.Operation {
		operationStatus = delta.UpdatedStatus
	}

	parentOperationsDeltaMap := delta.DetectMapDelta(
		oldRelations.ParentOperations,
		newRelations.ParentOperations,
		detectParentOperationDelta,
		toParentOperationDelta)

	status := delta.UnchangedStatus
	if resourceTypeStatus != delta.UnchangedStatus ||
		operationStatus != delta.UnchangedStatus ||
		parentOperationsDeltaMap.Status != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[OperationRelationsDelta]{
		Status: status,
		Value: OperationRelationsDelta{
			ResourceTypeDelta: delta.Delta[string]{
				Status: resourceTypeStatus,
				Value:  newRelations.ResourceType,
			},
			Operation: delta.Delta[string]{
				Status: operationStatus,
				Value:  newRelations.Operation,
			},
			ParentOperations: parentOperationsDeltaMap,
		},
	}
}

func toOperationRelationsDelta(status delta.Status, value OperationRelations) OperationRelationsDelta {
	return OperationRelationsDelta{
		ResourceTypeDelta: delta.Delta[string]{
			Status: status,
			Value:  value.ResourceType,
		},
		Operation: delta.Delta[string]{
			Status: status,
			Value:  value.Operation,
		},
		ParentOperations: delta.Delta[map[string]delta.KeyValueDelta[ParentOperationDelta]]{
			Status: status,
			Value:  delta.ToMapDelta(status, value.ParentOperations, toParentOperationDelta),
		},
	}
}

func detectParentOperationDelta(oldParentOperation ParentOperation, newParentOperation ParentOperation) delta.Delta[ParentOperationDelta] {
	resourceTypeStatus := delta.UnchangedStatus
	if oldParentOperation.ParentResourceType != newParentOperation.ParentResourceType {
		resourceTypeStatus = delta.UpdatedStatus
	}

	operationStatus := delta.UnchangedStatus
	if oldParentOperation.ParentOperation != newParentOperation.ParentOperation {
		operationStatus = delta.UpdatedStatus
	}

	status := delta.UnchangedStatus
	if resourceTypeStatus != delta.UnchangedStatus ||
		operationStatus != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[ParentOperationDelta]{
		Status: status,
		Value: ParentOperationDelta{
			ParentResourceType: delta.Delta[string]{
				Status: resourceTypeStatus,
				Value:  newParentOperation.ParentResourceType,
			},
			ParentOperation: delta.Delta[string]{
				Status: operationStatus,
				Value:  newParentOperation.ParentOperation,
			},
		},
	}
}

func toParentOperationDelta(status delta.Status, value ParentOperation) ParentOperationDelta {
	return ParentOperationDelta{
		ParentResourceType: delta.Delta[string]{
			Status: status,
			Value:  value.ParentResourceType,
		},
		ParentOperation: delta.Delta[string]{
			Status: status,
			Value:  value.ParentOperation,
		},
	}
}
