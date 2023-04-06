package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const dbType = "postgres"

const DefaultMigrationRoot = "app/dao/sqldb/migrations"
const DefaultSeedFile = "app/dao/sqldb/seed.sql"
const MigrateAll = 0

const lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
const upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"
const specialChars = "!@$%^&*()-+=?[]"
const dbPasswordLen = 20

type Config struct {
	DBHost                  string        `envconfig:"DB_HOST" default:"localhost"`
	DBPort                  int           `envconfig:"DB_PORT" default:"5432"`
	DBUser                  string        `envconfig:"DB_USER"`
	DBName                  string        `envconfig:"DB_NAME" default:"teamy"`
	DBPassword              string        `envconfig:"DB_PASSWORD"`
	DBSSLMode               string        `envconfig:"DB_SSL_MODE" default:"require"`
	DBMaxOpenConnections    int           `envconfig:"DB_MAX_OPEN_CONNECTIONS" default:"20"`
	DBMaxIdleConnections    int           `envconfig:"DB_MAX_IDLE_CONNECTIONS" default:"2"`
	DBConnectionMaxLifeTime time.Duration `envconfig:"DB_CONNECTION_MAX_LIFE_TIME" default:"0"`
	DBConnectionMaxIdleTime time.Duration `envconfig:"DB_CONNECTION_MAX_IDLE_TIME" default:"2m"`
}

func Use(logger telemetry.Logger, cfg Config, action func(sqlDB *sql.DB) *errs.Error) *errs.Error {
	sqlDB, err := connect(cfg)
	if err != nil {
		return err
	}

	defer sqlDB.Close()

	waitUntilReady(logger, sqlDB)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.DBConnectionMaxLifeTime)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnectionMaxIdleTime)
	return action(sqlDB)
}

func MigrateUp(logger telemetry.Logger, sqlDB *sql.DB, migrationRoot string, steps int) *errs.Error {
	return migrateDB(logger, sqlDB, migrationRoot, migrate.Up, steps)
}

func MigrateDown(logger telemetry.Logger, sqlDB *sql.DB, migrationRoot string, steps int) *errs.Error {
	return migrateDB(logger, sqlDB, migrationRoot, migrate.Down, steps)
}

func NewMigration(migrationDir string, fileName string) (string, *errs.Error) {
	now := time.Now()
	prefix := fmt.Sprintf(
		"%04d%02d%02d%02d%02d%02d_%s",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		fileName)

	err := os.MkdirAll(migrationDir, os.ModePerm)
	if err != nil {
		return "", errs.NewError(errs.OS, err.Error())
	}

	fileName = fmt.Sprintf("%s.sql", prefix)
	fullFilePath := filepath.Join(migrationDir, fileName)
	err = io.CreateFileWithLog(fullFilePath)
	if err != nil {
		return "", errs.NewError(errs.OS, err.Error())
	}

	return fullFilePath, nil
}

func New(logger telemetry.Logger, dbName string) {
	alphabet := concatenate([]string{
		lowerCaseLetters,
		upperCaseLetters,
		digits,
		specialChars,
	})
	dbNamePostfixAlphabet := concatenate([]string{lowerCaseLetters, upperCaseLetters, digits})
	dbNamePostfix := randgen.String(dbNamePostfixAlphabet, 5)
	fullDBName := fmt.Sprintf("%s-%s", dbName, dbNamePostfix)
	password := randgen.String(alphabet, dbPasswordLen)
	logger.
		Info(strings.TrimSpace(fmt.Sprintf(`
user: %s
password: %s
dbName: %s
SQL:
================================================================================
CREATE DATABASE "%s";
CREATE USER "%s" WITH PASSWORD '%s';
GRANT ALL PRIVILEGES ON DATABASE "%s" TO "%s";
================================================================================
`,
			fullDBName,
			password,
			fullDBName,
			fullDBName,
			fullDBName,
			password,
			fullDBName,
			fullDBName,
		)))
}

func ExecSQL(logger telemetry.Logger, sqlDB *sql.DB, sqlFileName string) *errs.Error {
	buf, err := os.ReadFile(sqlFileName)
	if err != nil {
		return errs.NewError(errs.OS, err.Error())
	}

	tx, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	_, err = tx.Exec(string(buf))
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	err = tx.Commit()
	if err == nil {
		logger.
			Info("successfully seeded DB")
	}

	return errs.NewError(errs.Unknown, err.Error())
}

func waitUntilReady(logger telemetry.Logger, sqlDB *sql.DB) {
	backOff := backoff.
		NewUniformBuilder().
		Delay(5 * time.Second).
		Build()
	runTime := runtime.NewBuiltInRuntime()
	rt := retry.NewInfinite(logger, runTime, backOff, backOff, nil)
	ct := context.Background()
	rt.WithRetry(ct, func() *errs.Error {
		err := sqlDB.Ping()
		if err == nil {
			logger.
				Info("successfully connected to the DB")
			return nil
		}

		return errs.NewError(errs.Unknown, fmt.Sprintf("failed to connect to the DB: %s", err.Error()))
	})
}

func connect(cfg Config) (*sql.DB, *errs.Error) {
	dbSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode)
	db, err := sql.Open(dbType, dbSource)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return db, nil
}

func migrateDB(
	logger telemetry.Logger,
	db *sql.DB,
	migrationRoot string,
	migrateDirection migrate.MigrationDirection,
	steps int,
) *errs.Error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationRoot,
	}
	_, err := migrate.ExecMax(db, dbType, migrations, migrateDirection, steps)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	logger.
		Info("migration finished")
	return nil
}

func concatenate(src []string) []rune {
	return []rune(strings.Join(src, ""))
}
