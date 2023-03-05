package authorization

import (
	"io/ioutil"

	"github.com/teamyapp/cloud/libs/errs"
	"gopkg.in/yaml.v3"
)

<<<<<<< HEAD
type ResourceTypeOperationsRow struct {
	ResourceType string   `yaml:"resourceType"`
	Operations   []string `yaml:"operations"`
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

type Config struct {
	ResourceTypeOperations []ResourceTypeOperationsRow `yaml:"resourceTypeOperations"`
	OperationRelations     []OperationRelationsRow     `yaml:"operationRelations"`
}

func ParseConfig(configPath string) (*Config, *errs.Error) {
	yamlConfigContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errs.NewError(errs.IO, err.Error())
	}

	authorizationConfig := Config{}
	err = yaml.Unmarshal(yamlConfigContent, &authorizationConfig)
	if err != nil {
		return nil, errs.NewError(errs.Deserialization, err.Error())
	}

	return &authorizationConfig, nil
}