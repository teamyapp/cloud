package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

const configFilePath = ".cli.toml"

type Config struct {
	DBMigrationsDir            string `toml:"dbMigrationsDir"`
	AuthorizationCoreSrcFile   string `toml:"authorizationCoreSrcFile"`
	AuthorizationCoreOutputDir string `toml:"authorizationCoreOutputFile"`
}

var cliConfig = Config{
	DBMigrationsDir:            "app/dao/sqldb/migrations",
	AuthorizationCoreSrcFile:   "core/authorization.yml",
	AuthorizationCoreOutputDir: "core/authorization/out",
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
