package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/dep"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var serviceLabels = []string{"cloud", "backend"}

func main() {
	cfg, err := config.AppFromEnv()
	if err != nil {
		panic(err)
	}

	lineFormatter := newLineFormatter(cfg.Environment)
	logOutput, err := newLogOutput(cfg.Environment, strings.Join(serviceLabels, "-"))
	if err != nil {
		panic(err)
	}

	defer logOutput.Close()
	serviceLabelsWithEnv := append([]string{}, serviceLabels...)
	serviceLabelsWithEnv = append(serviceLabelsWithEnv, strings.ToLower(string(cfg.Environment)))
	logger := telemetry.NewLogger(
		lineFormatter,
		logOutput,
		cfg.LogVisibleLevel,
		[]telemetry.LogInterceptor{
			telemetry.NewCommitLogInterceptor(cfg.GitLongCommitHash),
			telemetry.NewServiceLogInterceptor(strings.Join(serviceLabelsWithEnv, "/")),
			telemetry.RequestLogInterceptor,
			telemetry.ClientLogInterceptor,
		},
	)

	dataCollector := telemetry.NewDataCollector(logger)
	gitCommitLink := fmt.Sprintf("https://github.com/%s/%s/commit/%s",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: gitCommitLink,
	})

	err = sqldb.Use(dataCollector, cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(dataCollector, sqlDB, sqldb.DefaultMigrationRoot, sqldb.MigrateAll)
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		runnerConfig, err := runner.ServiceRunnerConfigFromEnv(dataCollector)
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		webAPIBaseURL := dep.WebAPIBaseURL(cfg.WebAPIBaseURL)
		oauthProviders := []oauth.Provider{
			dep.InitGoogleOAuthProvider(
				dataCollector,
				webAPIBaseURL,
				dep.JWTSigningKey(cfg.JWTSigningKey),
				dep.ClientID(cfg.GoogleClientID),
				dep.ClientSecret(cfg.GoogleClientSecret)),
			dep.InitGitHubOAuthProvider(
				dataCollector,
				webAPIBaseURL,
				dep.ClientID(cfg.GitHubClientID),
				dep.ClientSecret(cfg.GitHubClientSecret)),
			dep.InitSlackOAuthProvider(
				dataCollector,
				webAPIBaseURL,
				dep.JWTSigningKey(cfg.JWTSigningKey),
				dep.ClientID(cfg.SlackClientID),
				dep.ClientSecret(cfg.SlackClientSecret)),
		}
		identityAPI, err := dep.InitIdentityAPI(
			dataCollector,
			sqlDB,
			oauthProviders,
			dep.AccessTokenTTL(cfg.AccessTokenTTL),
			dep.JWTSigningKey(cfg.JWTSigningKey),
			dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		generatorAPI, err := dep.InitGeneratorAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		authorizationAPI, err := dep.InitAuthorizationAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		fileAPI, err := dep.InitFileAPI(
			dataCollector,
			cfg.Environment,
			sqlDB,
			dep.GenRangeSize(cfg.GenRangeSize),
			dep.S3Endpoint(cfg.S3Endpoint),
			dep.S3AccessKeyID(cfg.S3AccessKeyID),
			dep.S3AccessKey(cfg.S3AccessKey),
			dep.S3BucketName(cfg.S3BucketName))
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

		telemetryAPI := dep.InitTelemetryAPI(dataCollector)
		rn := runner.NewServiceRunner(dataCollector, runnerConfig, []runner.Service{
			identityAPI,
			generatorAPI,
			authorizationAPI,
			fileAPI,
			telemetryAPI,
		})

		rn.Start()
		return nil
	})
}

func getEnv(name string, defaultVal string) string {
	value := os.Getenv(name)
	if len(value) > 0 {
		return value
	}

	return defaultVal
}

func newLineFormatter(environment env.Environment) telemetry.LineFormatter {
	if environment == env.DevelopmentEnv {
		return telemetry.NewOrderedColumnLineFormatter([]string{
			telemetry.HappenAtProp,
			telemetry.SeverityProp,
			telemetry.FileNameProp,
			telemetry.LineNumberProp,
			telemetry.RequestIDProp,
			telemetry.ClientIDProp,
			middleware.ProtocolProp,
			middleware.StageProp,
			middleware.HostProp,
			middleware.MethodProp,
			middleware.PathProp,
			middleware.HeadersProp,
			middleware.MetadataProp,
			middleware.BodySizeProp,
			middleware.BodyProp,
			telemetry.CauseProp,
			telemetry.MessageProp,
		})
	}

	return telemetry.NewJSONLineFormatter()
}

func newLogOutput(environment env.Environment, serviceName string) (io.WriteCloser, error) {
	if environment == env.DevelopmentEnv {
		logFileName := fmt.Sprintf("%v.log", serviceName)
		logFilePath := getEnv("LOG_OUTPUT_FILE", filepath.Join("..", "logs", logFileName))
		logDir := filepath.Dir(logFilePath)

		// MkdirAll requires at least 700 permission:
		// https://github.com/golang/go/issues/22323
		err := os.MkdirAll(logDir, 0744)
		if err != nil {
			return nil, err
		}

		return os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	}

	return os.Stdout, nil
}
