package authorization

import (
	"os"

	"github.com/teamyapp/cloud/libs/errs"
	"gopkg.in/yaml.v3"
)

func ParseConfigFromFile(configPath string) (Config, *errs.Error) {
	yamlConfigContent, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, errs.NewError(errs.IO, err.Error())
	}

	return ParseConfig(string(yamlConfigContent))
}

func ParseConfig(configContent string) (Config, *errs.Error) {
	rawConfig := RawConfig{}
	err := yaml.Unmarshal([]byte(configContent), &rawConfig)
	if err != nil {
		return Config{}, errs.NewError(errs.Deserialization, err.Error())
	}

	return rawConfig.ToConfig()
}
