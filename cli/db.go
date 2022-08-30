package main

import (
	"database/sql"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/obs"
)

var dbName string
var migrationFileName string
var migrationSteps int
var seedFilePath string
var dataCollector = obs.NewDataCollector(obs.NewRawLogger(obs.Info))

const migrationTemplate = `
-- +migrate Up

-- +migrate Down

`

var dbCmd = &cobra.Command{
	Use: "db",
}

var newDBCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate SQL to create new database",
	Run: func(cmd *cobra.Command, args []string) {
		sqldb.New(dataCollector, dbName)
	},
}

var migrateCmd = &cobra.Command{
	Use: "migrate",
}

var migrateUpCmd = &cobra.Command{
	Use: "up",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(dataCollector, func(sqlDB *sql.DB) error {
			return sqldb.MigrateUp(dataCollector, sqlDB, cliConfig.DBMigrationsDir, migrationSteps)
		})
	},
}

var migrateDownCmd = &cobra.Command{
	Use: "down",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(dataCollector, func(sqlDB *sql.DB) error {
			return sqldb.MigrateDown(dataCollector, sqlDB, cliConfig.DBMigrationsDir, migrationSteps)
		})
	},
}

var newMigrationCmd = &cobra.Command{
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {
		fullFilePath, err := sqldb.NewMigration(dataCollector, cliConfig.DBMigrationsDir, migrationFileName)
		if err != nil {
			return err
		}

		return os.WriteFile(fullFilePath, []byte(strings.TrimPrefix(migrationTemplate, "\n")), 0644)
	},
}

var seedCmd = &cobra.Command{
	Use: "seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(dataCollector, func(sqlDB *sql.DB) error {
			return sqldb.ExecSQL(dataCollector, sqlDB, seedFilePath)
		})
	},
}

var newSeedCmd = &cobra.Command{
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {
		return io.CreateFileWithLog(seedFilePath)
	},
}

func addDBCmd() {
	newMigrationCmd.Flags().StringVarP(
		&migrationFileName,
		"fileName",
		"f",
		"",
		"name of data migration file")
	newMigrationCmd.MarkFlagRequired("fileName")
	migrateCmd.AddCommand(newMigrationCmd)
	migrateCmd.PersistentFlags().IntVarP(
		&migrationSteps,
		"steps",
		"s",
		0,
		"number of migrations to perform")
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	dbCmd.AddCommand(migrateCmd)

	seedCmd.PersistentFlags().StringVarP(
		&seedFilePath,
		"file",
		"f",
		sqldb.DefaultSeedFile,
		"location of DB seed SQL")
	seedCmd.AddCommand(newSeedCmd)
	dbCmd.AddCommand(seedCmd)

	newDBCmd.Flags().StringVarP(
		&dbName,
		"name",
		"n",
		"",
		"name of new DB")
	newDBCmd.MarkFlagRequired("name")
	dbCmd.AddCommand(newDBCmd)

	rootCmd.AddCommand(dbCmd)
}

func useSQLDB(dataCollector obs.DataCollector, action func(sqlDB *sql.DB) error) error {
	cfg, err := config.AppFromEnv(dataCollector)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return sqldb.Use(dataCollector, cfg.Config, action)
}
