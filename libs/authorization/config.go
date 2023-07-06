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

type Operation struct {
	ResourceType string `yaml:"resourceType"`
	Operation    string `yaml:"operation"`
}

type OperationRelationsRow struct {
	ResourceType     string      `yaml:"resourceType"`
	Operation        string      `yaml:"operation"`
	ParentOperations []Operation `yaml:"parents"`
}

type OperationRelations struct {
	ResourceType     string
	Operation        string
	ParentOperations map[string]Operation
}

type Config struct {
	ResourceTypeOperations map[string]ResourceTypeOperations
	OperationRelations     map[string]OperationRelations
}

func NewConfig(
	resourceTypeOperations map[string]ResourceTypeOperations,
	operationRelations map[string]OperationRelations,
) Config {
	return Config{
		ResourceTypeOperations: resourceTypeOperations,
		OperationRelations:     operationRelations,
	}
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

func GetOperationKey(resourceType string, operation string) string {
	return path.Join(resourceType, operation)
}

func toParentOperations(input []Operation) (map[string]Operation, *errs.Error) {
	output := make(map[string]Operation)
	for _, parentOperation := range input {
		key := path.Join(parentOperation.ResourceType, parentOperation.Operation)
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
