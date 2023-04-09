package main

import (
	"database/sql"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var dbName string
var migrationFileName string
var migrationSteps int
var seedFilePath string
var logger telemetry.Logger

func init() {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{
		telemetry.HappenAtProp,
		telemetry.SeverityProp,
		telemetry.FileNameProp,
		telemetry.LineNumberProp,
		telemetry.RequestIDProp,
		telemetry.ClientIDProp,
		telemetry.CauseProp,
		telemetry.MessageProp,
	})
	logger = telemetry.NewLogger(
		lineFormatter,
		os.Stdout,
		telemetry.Info,
		[]telemetry.LogInterceptor{
			telemetry.RequestLogInterceptor,
			telemetry.ClientLogInterceptor,
		},
	)
}

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
		sqldb.New(logger, dbName)
	},
}

var migrateCmd = &cobra.Command{
	Use: "migrate",
}

var migrateUpCmd = &cobra.Command{
	Use: "up",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := useSQLDB(logger, func(sqlDB *sql.DB) *errs.Error {
			return sqldb.MigrateUp(logger, sqlDB, cliConfig.DBMigrationsDir, migrationSteps)
		})
		if err != nil {
			return err.ToError()
		}

		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use: "down",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := useSQLDB(logger, func(sqlDB *sql.DB) *errs.Error {
			return sqldb.MigrateDown(logger, sqlDB, cliConfig.DBMigrationsDir, migrationSteps)
		})
		if err != nil {
			return err.ToError()
		}

		return nil
	},
}

var newMigrationCmd = &cobra.Command{
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {
		fullFilePath, err := sqldb.NewMigration(cliConfig.DBMigrationsDir, migrationFileName)
		if err != nil {
			return err.ToError()
		}

		return os.WriteFile(fullFilePath, []byte(strings.TrimPrefix(migrationTemplate, "\n")), 0644)
	},
}

var seedCmd = &cobra.Command{
	Use: "seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := useSQLDB(logger, func(sqlDB *sql.DB) *errs.Error {
			return sqldb.ExecSQL(logger, sqlDB, seedFilePath)
		})
		if err != nil {
			return err.ToError()
		}

		return nil
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

func useSQLDB(logger telemetry.Logger, action func(sqlDB *sql.DB) *errs.Error) *errs.Error {
	cfg, err := config.AppFromEnv()
	if err != nil {
		return err
	}

	return sqldb.Use(logger, cfg.Config, action)
}
