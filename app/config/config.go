package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Repo struct {
	GitLongCommitHash string `envconfig:"GIT_LONG_COMMIT_HASH"`
	GitRepoOwner      string `envconfig:"GIT_REPO_OWNER"`
	GitRepoName       string `envconfig:"GIT_REPO_NAME"`
}

type Service struct {
	Environment     env.Environment    `json:"ENVIRONMENT" default:"development"`
	LogVisibleLevel telemetry.LogLevel `envconfig:"LOG_VISIBLE_LEVEL" default:"Info"`
}

type App struct {
	Repo
	sqldb.Config
	Service
	WebAPIBaseURL      string        `envconfig:"WEB_API_BASE_URL" default:""`
	AccessTokenTTL     time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:""`
	GoogleClientID     string        `envconfig:"GOOGLE_CLIENT_ID" default:""`
	GoogleClientSecret string        `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
	GitHubClientID     string        `envconfig:"GITHUB_CLIENT_ID" default:""`
	GitHubClientSecret string        `envconfig:"GITHUB_CLIENT_SECRET" default:""`
	SlackClientID      string        `envconfig:"SLACK_CLIENT_ID" default:""`
	SlackClientSecret  string        `envconfig:"SLACK_CLIENT_SECRET" default:""`
	JWTSigningKey      string        `envconfig:"JWT_SIGNING_KEY" default:""`
	GenRangeSize       int           `envconfig:"GEN_RANGE_SIZE" default:"100"`
	S3Endpoint         string        `envconfig:"S3_ENDPOINT" default:""`
	S3AccessKeyID      string        `envconfig:"S3_ACCESS_KEY_ID" default:""`
	S3AccessKey        string        `envconfig:"S3_ACCESS_KEY" default:""`
	S3BucketName       string        `envconfig:"S3_BUCKET_NAME" default:"teamyapp"`
}

func AppFromEnv() (App, *errs.Error) {
	cfg := App{}
	err := FromEnv(&cfg)
	if err != nil {
		return App{}, err
	}

	return cfg, nil
}

func FromEnv[Config any](config Config) *errs.Error {
	internalErr := autoLoadEnv(".env")
	if internalErr != nil {
		return internalErr
	}

	internalErr = autoLoadEnv(".repo.env")
	if internalErr != nil {
		return internalErr
	}

	err := envconfig.Process("", config)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func autoLoadEnv(fileName string) *errs.Error {
	_, err := os.Stat(fileName)
	if err == nil {
		err := godotenv.Load(fileName)
		if err != nil {
			return errs.NewError(errs.Unknown, err.Error())
		}

		return nil
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return errs.NewError(errs.Unknown, err.Error())
	}
}
