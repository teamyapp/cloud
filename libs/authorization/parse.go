package authorization

import (
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	ResourceTypeOperations []struct {
		ResourceType string   `yaml:"resourceType"`
		Operations   []string `yaml:"operations"`
	} `yaml:"resourceTypeOperations"`
	OperationRelations []struct {
		ResourceType     string `yaml:"resourceType"`
		Operation        string `yaml:"operation"`
		ParentOperations []struct {
			ParentResourceType string `yaml:"resourceType"`
			ParentOperation    string `yaml:"operation"`
		} `yaml:"parents"`
	} `yaml:"operationRelations"`
}

func Parse(configPath string, dataCollector telemetry.DataCollector) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		fileReadErr := &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(fileReadErr)
		return nil, err
	}

	authorizationConfig := Config{}

	err = yaml.Unmarshal(yamlFile, &authorizationConfig)
	if err != nil {
		parseErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(parseErr)
		return nil, err
	}

	return &authorizationConfig, err
}
