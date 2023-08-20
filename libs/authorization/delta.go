package authorization

import (
	"github.com/teamyapp/cloud/libs/delta"
)

type ConfigDelta struct {
	ResourceTypeOperationsDelta delta.Delta[map[string]delta.KeyValueDelta[ResourceTypeOperationsDelta]]
	OperationRelationsDelta     delta.Delta[map[string]delta.KeyValueDelta[OperationRelationsDelta]]
}

type ResourceTypeOperationsDelta struct {
	ResourceType    string
	OperationsDelta delta.Delta[map[string]delta.KeyValueDelta[bool]]
}

type OperationRelationsDelta struct {
	ChildResourceType     string
	ChildOperation        string
	ParentOperationsDelta delta.Delta[map[string]delta.KeyValueDelta[Operation]]
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
	if resourceTypeOperationsDelta.Status != delta.UnchangedStatus ||
		operationRelationsDelta.Status != delta.UnchangedStatus {
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
			ResourceType:    newResourceTypeOperations.ResourceType,
			OperationsDelta: operationsDeltaMap,
		},
	}
}

func toResourceTypeOperationsDelta(
	status delta.Status,
	resourceTypeOperations ResourceTypeOperations,
) ResourceTypeOperationsDelta {
	return ResourceTypeOperationsDelta{
		ResourceType: resourceTypeOperations.ResourceType,
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
		delta.ToValueDelta[Operation])

	status := delta.UnchangedStatus
	if resourceTypeStatus != delta.UnchangedStatus ||
		operationStatus != delta.UnchangedStatus ||
		parentOperationsDeltaMap.Status != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[OperationRelationsDelta]{
		Status: status,
		Value: OperationRelationsDelta{
			ChildResourceType:     newRelations.ResourceType,
			ChildOperation:        newRelations.Operation,
			ParentOperationsDelta: parentOperationsDeltaMap,
		},
	}
}

func toOperationRelationsDelta(status delta.Status, value OperationRelations) OperationRelationsDelta {
	return OperationRelationsDelta{
		ChildResourceType: value.ResourceType,
		ChildOperation:    value.Operation,
		ParentOperationsDelta: delta.Delta[map[string]delta.KeyValueDelta[Operation]]{
			Status: status,
			Value:  delta.ToMapDelta(status, value.ParentOperations, delta.ToValueDelta[Operation]),
		},
	}
}

func detectParentOperationDelta(oldParentOperation Operation, newParentOperation Operation) delta.Delta[Operation] {
	resourceTypeStatus := delta.UnchangedStatus
	if oldParentOperation.ResourceType != newParentOperation.ResourceType {
		resourceTypeStatus = delta.UpdatedStatus
	}

	operationStatus := delta.UnchangedStatus
	if oldParentOperation.Operation != newParentOperation.Operation {
		operationStatus = delta.UpdatedStatus
	}

	status := delta.UnchangedStatus
	if resourceTypeStatus != delta.UnchangedStatus ||
		operationStatus != delta.UnchangedStatus {
		status = delta.UpdatedStatus
	}

	return delta.Delta[Operation]{
		Status: status,
		Value: Operation{
			ResourceType: newParentOperation.ResourceType,
			Operation:    newParentOperation.Operation,
		},
	}
}
