package authorization

import (
	"io/ioutil"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"gopkg.in/yaml.v3"
)

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

func ParseConfig(configPath string, dataCollector telemetry.DataCollector) (*Config, error) {
	yamlConfigContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		fileReadErr := errs.NewError(errs.IO, err.Error())
		dataCollector.Logger.Error(fileReadErr)
		return nil, err
	}

	authorizationConfig := Config{}
	err = yaml.Unmarshal(yamlConfigContent, &authorizationConfig)
	if err != nil {
		parseErr := errs.NewError(errs.Deserialization, err.Error())
		dataCollector.Logger.Error(parseErr)
		return nil, err
	}

	return &authorizationConfig, err
}
