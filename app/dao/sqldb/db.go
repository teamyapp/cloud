package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
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

func Use(dataCollector telemetry.DataCollector, cfg Config, action func(sqlDB *sql.DB) *errs.Error) *errs.Error {
	sqlDB, err := connect(cfg)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	waitUntilReady(dataCollector, sqlDB)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.DBConnectionMaxLifeTime)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnectionMaxIdleTime)
	return action(sqlDB)
}

func MigrateUp(dataCollector telemetry.DataCollector, sqlDB *sql.DB, migrationRoot string, steps int) *errs.Error {
	return migrateDB(dataCollector, sqlDB, migrationRoot, migrate.Up, steps)
}

func MigrateDown(dataCollector telemetry.DataCollector, sqlDB *sql.DB, migrationRoot string, steps int) *errs.Error {
	return migrateDB(dataCollector, sqlDB, migrationRoot, migrate.Down, steps)
}

func NewMigration(dataCollector telemetry.DataCollector, migrationDir string, fileName string) (string, *errs.Error) {
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
		internalErr := &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return "", internalErr
	}

	fileName = fmt.Sprintf("%s.sql", prefix)
	fullFilePath := filepath.Join(migrationDir, fileName)
	err = io.CreateFileWithLog(fullFilePath)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return "", internalErr
	}

	return fullFilePath, nil
}

func New(dataCollector telemetry.DataCollector, dbName string) {
	alphabet := concatenate([]string{
		lowerCaseLetters,
		upperCaseLetters,
		digits,
		specialChars,
	})
	dbNamePostfixAlphabet := concatenate([]string{lowerCaseLetters, upperCaseLetters, digits})
	dbNamePostfix := randString(dbNamePostfixAlphabet, 5)
	fullDBName := fmt.Sprintf("%s-%s", dbName, dbNamePostfix)
	password := randString(alphabet, dbPasswordLen)
	dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: strings.TrimSpace(fmt.Sprintf(`
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
		)),
	})
}

func ExecSQL(dataCollector telemetry.DataCollector, sqlDB *sql.DB, sqlFileName string) *errs.Error {
	buf, err := os.ReadFile(sqlFileName)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.OS,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	tx, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	_, err = tx.Exec(string(buf))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	err = tx.Commit()
	if err == nil {
		dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
			telemetry.MessageProp: "successfully seeded DB",
		})
	}

	return &errs.Error{
		Code:     errs.Unknown,
		EmbedErr: err,
	}
}

func waitUntilReady(dataCollector telemetry.DataCollector, sqlDB *sql.DB) {
	for {
		err := sqlDB.Ping()
		if err == nil {
			dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
				telemetry.MessageProp: "successfully connected to the DB",
			})
			break
		}

		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		dataCollector.Logger.Log(telemetry.Warning, telemetry.Props{
			telemetry.MessageProp: "fail to connect to the DB",
		})
		dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
			telemetry.MessageProp: "retry after 5 seconds",
		})
		time.Sleep(5 * time.Second)
	}
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		return nil, internalErr
	}

	return db, nil
}

func migrateDB(
	dataCollector telemetry.DataCollector,
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: "migration finished",
	})
	return nil
}

func concatenate(src []string) []rune {
	return []rune(strings.Join(src, ""))
}

func randString(alphabet []rune, length int) string {
	alphabetEndIndex := len(alphabet) - 1
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		randomIndex := randInt(0, alphabetEndIndex)
		result[i] = alphabet[randomIndex]
	}
	return string(result)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}
