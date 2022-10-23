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
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/obs"
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
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER"`
	DBName     string `envconfig:"DB_NAME" default:"teamy"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`
}

func Use(dataCollector obs.DataCollector, cfg Config, action func(sqlDB *sql.DB) error) error {
	sqlDB, err := connect(cfg)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	waitUntilReady(dataCollector, sqlDB)
	return action(sqlDB)
}

func MigrateUp(dataCollector obs.DataCollector, sqlDB *sql.DB, migrationRoot string, steps int) error {
	return migrateDB(dataCollector, sqlDB, migrationRoot, migrate.Up, steps)
}

func MigrateDown(dataCollector obs.DataCollector, sqlDB *sql.DB, migrationRoot string, steps int) error {
	return migrateDB(dataCollector, sqlDB, migrationRoot, migrate.Down, steps)
}

func NewMigration(dataCollector obs.DataCollector, migrationDir string, fileName string) (string, error) {
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
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	fileName = fmt.Sprintf("%s.sql", prefix)
	fullFilePath := filepath.Join(migrationDir, fileName)
	err = io.CreateFileWithLog(fullFilePath)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	return fullFilePath, nil
}

func New(dbName string) {
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
	fmt.Println(strings.TrimSpace(fmt.Sprintf(`
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

func ExecSQL(dataCollector obs.DataCollector, sqlDB *sql.DB, sqlFileName string) error {
	buf, err := os.ReadFile(sqlFileName)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	_, err = tx.Exec(string(buf))
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = tx.Commit()
	if err == nil {
		dataCollector.Logger.Log(obs.Info, obs.Props{
			obs.MessageProp: "successfully seeded DB",
		})
	}

	return err
}

func waitUntilReady(dataCollector obs.DataCollector, sqlDB *sql.DB) {
	for {
		err := sqlDB.Ping()
		if err == nil {
			dataCollector.Logger.Log(obs.Info, obs.Props{
				obs.MessageProp: "successfully connected to the DB",
			})
			break
		}

		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		dataCollector.Logger.Log(obs.Warning, obs.Props{
			obs.MessageProp: "fail to connect to the DB",
		})
		dataCollector.Logger.Log(obs.Info, obs.Props{
			obs.MessageProp: "retry after 5 seconds",
		})
		time.Sleep(5 * time.Second)
	}
}

func connect(cfg Config) (*sql.DB, error) {
	dbSource := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode)
	return sql.Open(dbType, dbSource)
}

func migrateDB(
	dataCollector obs.DataCollector,
	db *sql.DB,
	migrationRoot string,
	migrateDirection migrate.MigrationDirection,
	steps int,
) error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationRoot,
	}
	_, err := migrate.ExecMax(db, dbType, migrations, migrateDirection, steps)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: "migration finished",
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
