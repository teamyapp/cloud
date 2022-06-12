package cmd

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/io"
)

var dbName string
var migrationDir string
var migrationFileName string
var seedFilePath string

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
		sqldb.New(dbName)
	},
}

var migrateCmd = &cobra.Command{
	Use: "migrate",
}

var migrateUpCmd = &cobra.Command{
	Use: "up",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(func(sqlDB *sql.DB) error {
			return sqldb.MigrateUp(sqlDB, migrationDir)
		})
	},
}

var migrateDownCmd = &cobra.Command{
	Use: "down",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(func(sqlDB *sql.DB) error {
			return sqldb.MigrateDown(sqlDB, migrationDir)
		})
	},
}

var newMigrationCmd = &cobra.Command{
	Use: "new",
	RunE: func(cmd *cobra.Command, args []string) error {
		fullFilePath, err := sqldb.NewMigration(migrationDir, migrationFileName)
		if err != nil {
			return err
		}

		return os.WriteFile(fullFilePath, []byte(strings.TrimPrefix(migrationTemplate, "\n")), 0644)
	},
}

var seedCmd = &cobra.Command{
	Use: "seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		return useSQLDB(func(sqlDB *sql.DB) error {
			return sqldb.ExecSQL(sqlDB, seedFilePath)
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

	migrateCmd.PersistentFlags().StringVarP(
		&migrationDir,
		"migrationDir",
		"d",
		sqldb.DefaultMigrationRoot,
		"location of DB migration files")
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

func useSQLDB(action func(sqlDB *sql.DB) error) error {
	cfg, err := config.AppFromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return sqldb.Use(cfg.Config, action)
}
