package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

const configFilePath = ".cli.toml"

type AuthorizationOption struct {
	ConfigFilePath string `toml:"configFilePath"`
	OutputDir      string `toml:"outputDir"`
}

type Config struct {
	DBMigrationsDir      string                `toml:"dbMigrationsDir"`
	AuthorizationOptions []AuthorizationOption `toml:"AuthorizationOptions"`
}

var cliConfig = Config{
	DBMigrationsDir: "app/dao/sqldb/migrations",
	AuthorizationOptions: []AuthorizationOption{
		{
			ConfigFilePath: "authorization.yml",
			OutputDir:      "app/authorizationv2",
		},
	},
}

func main() {
	if _, err := os.Stat(configFilePath); err == nil {
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			panic(err)
		}

		_, err = toml.Decode(string(data), &cliConfig)
		if err != nil {
			panic(err)
		}
	}

	Execute()
}
