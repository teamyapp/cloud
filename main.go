package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/dep"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const appName = "cloud"
const serviceName = "backend"

var serviceLabels = []string{appName, serviceName}
var fullServiceName = strings.Join(serviceLabels, "-")

func main() {
	cfg, err := config.AppFromEnv()
	if err != nil {
		panic(err)
	}

	lineFormatter := newLineFormatter(cfg.Environment)

	logFileName := fmt.Sprintf("%v.log", fullServiceName)
	logFilePath := getEnv("LOG_OUTPUT_FILE", filepath.Join("..", "logs", logFileName))
	logOutput, err := telemetry.NewLogOutput(cfg.Environment, logFilePath)
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
			telemetry.TraceLogInterceptor,
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

	err = sqldb.Use(dataCollector, cfg.Config, func(sqlDB *sql.DB) *errs.Error {
		internalErr := sqldb.MigrateUp(dataCollector, sqlDB, sqldb.DefaultMigrationRoot, sqldb.MigrateAll)
		if internalErr != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		runnerConfig, internalErr := runner.ServiceRunnerConfigFromEnv()
		if internalErr != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		webAPIBaseURL := dep.WebAPIBaseURL(cfg.WebAPIBaseURL)
		oauthProviders := []oauth.Provider{
			dep.InitGitHubOAuthProvider(
				dataCollector,
				webAPIBaseURL,
				dep.ClientID(cfg.GitHubClientID),
				dep.ClientSecret(cfg.GitHubClientSecret)),
			dep.InitGoogleOAuthProvider(
				dataCollector,
				webAPIBaseURL,
				dep.JWTSigningKey(cfg.JWTSigningKey),
				dep.ClientID(cfg.GoogleClientID),
				dep.ClientSecret(cfg.GoogleClientSecret)),
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
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		generatorAPI, err := dep.InitGeneratorAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		authorizationAPI, err := dep.InitAuthorizationAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
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
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		telemetryAPI := dep.InitTelemetryAPI(dataCollector)
		rn := runner.NewServiceRunnerBuilder(
			dataCollector,
			network.NewSocket(),
			metrics.NewPrometheus(appName, serviceName, cfg.Environment),
			runnerConfig,
			fullServiceName,
			[]runner.Service{
				identityAPI,
				generatorAPI,
				authorizationAPI,
				fileAPI,
				telemetryAPI,
			}).
			IncludeIdentityWebFunc(api.IncludeIdentityWebFunc).
			Build()
		return rn.Start(nil)
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
			telemetry.TraceIDProp,
			telemetry.SpanIDProp,
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
			telemetry.StackTraceProp,
			telemetry.MessageProp,
		})
	}

	return telemetry.NewJSONLineFormatter()
}
