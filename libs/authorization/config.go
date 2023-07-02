package authorization

import (
	"path"

	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceTypeOperationsRow struct {
	ResourceType string   `yaml:"resourceType"`
	Operations   []string `yaml:"operations"`
}

type ResourceTypeOperations struct {
	ResourceType string
	Operations   map[string]bool
}

type ParentOperation struct {
	ParentResourceType string `yaml:"resourceType"`
	ParentOperation    string `yaml:"operation"`
}

type OperationRelationsRow struct {
	ResourceType     string            `yaml:"resourceType"`
	Operation        string            `yaml:"operation"`
	ParentOperations []ParentOperation `yaml:"parents"`
}

type OperationRelations struct {
	ResourceType     string
	Operation        string
	ParentOperations map[string]ParentOperation
}

type Config struct {
	ResourceTypeOperations map[string]ResourceTypeOperations
	OperationRelations     map[string]OperationRelations
}

type RawConfig struct {
	ResourceTypeOperations []ResourceTypeOperationsRow `yaml:"resourceTypeOperations"`
	OperationRelations     []OperationRelationsRow     `yaml:"operationRelations"`
}

func (r RawConfig) ToConfig() (Config, *errs.Error) {
	resourceTypeOperations := make(map[string]ResourceTypeOperations)
	for _, row := range r.ResourceTypeOperations {
		if _, ok := resourceTypeOperations[row.ResourceType]; ok {
			return Config{}, errs.NewError(errs.Unknown, "duplicate resource type")
		}

		operations, err := toMap(row.Operations)
		if err != nil {
			return Config{}, err
		}

		resourceTypeOperations[row.ResourceType] = ResourceTypeOperations{
			ResourceType: row.ResourceType,
			Operations:   operations,
		}
	}

	operationRelations := make(map[string]OperationRelations)
	for _, row := range r.OperationRelations {
		key := path.Join(row.ResourceType, row.Operation)
		if _, ok := operationRelations[key]; ok {
			return Config{}, errs.NewError(errs.Unknown, "duplicate resourceType operation")
		}

		parentOperations, err := toParentOperations(row.ParentOperations)
		if err != nil {
			return Config{}, err
		}

		operationRelations[key] = OperationRelations{
			ResourceType:     row.ResourceType,
			Operation:        row.Operation,
			ParentOperations: parentOperations,
		}
	}

	return Config{
		ResourceTypeOperations: resourceTypeOperations,
		OperationRelations:     operationRelations,
	}, nil
}

func toParentOperations(input []ParentOperation) (map[string]ParentOperation, *errs.Error) {
	output := make(map[string]ParentOperation)
	for _, parentOperation := range input {
		key := path.Join(parentOperation.ParentResourceType, parentOperation.ParentOperation)

		if _, ok := output[key]; ok {
			return nil, errs.NewError(errs.InvalidArgument, "duplicate parent operation")
		}

		output[key] = parentOperation
	}

	return output, nil
}

func toMap(input []string) (map[string]bool, *errs.Error) {
	output := make(map[string]bool)
	for _, item := range input {
		if _, ok := output[item]; ok {
			return nil, errs.NewError(errs.InvalidArgument, "duplicate item")
		}

		output[item] = true
	}

	return output, nil
}
