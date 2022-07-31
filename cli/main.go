package main

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

const configFilePath = ".cli.toml"

type Config struct {
	DBMigrationsDir string `toml:"dbMigrationsDir"`
}

var cliConfig = Config{
	DBMigrationsDir: "app/dao/sqldb/migrations",
}

func main() {
	if _, err := os.Stat(configFilePath); err == nil {
		data, err := ioutil.ReadFile(configFilePath)
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
