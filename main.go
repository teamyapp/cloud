package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/dep"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/runner"
)

func main() {
	logVisibleSeverity := obs.Severity(getEnv("LOG_VISIBLE_SEVERITY", "INFO"))
	dataCollector := dep.InitDataCollector(logVisibleSeverity)
	cfg, err := config.AppFromEnv(dataCollector)
	if err != nil {
		dataCollector.Logger.Log(obs.Fatal, obs.Props{obs.CauseProp: err})
		panic(err)
	}

	gitCommitLink := fmt.Sprintf("https://github.com/%s/%s/commit/%s",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: map[string]interface{}{
			"gitCommitLink": gitCommitLink,
		},
	})

	err = sqldb.Use(dataCollector, cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(dataCollector, sqlDB, sqldb.DefaultMigrationRoot, sqldb.MigrateAll)
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		runnerConfig, err := runner.ServiceRunnerConfigFromEnv(dataCollector)
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		}
		identityAPI, err := dep.InitIdentityAPI(
			dataCollector,
			sqlDB,
			oauthProviders,
			dep.AccessTokenTTL(cfg.AccessTokenTTL),
			dep.JWTSigningKey(cfg.JWTSigningKey),
			dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		generatorAPI, err := dep.InitGeneratorAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		authorizationAPI, err := dep.InitAuthorizationAPI(dataCollector, sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		rn := runner.NewServiceRunner(dataCollector, runnerConfig, []runner.Service{
			identityAPI,
			generatorAPI,
			authorizationAPI,
			fileAPI,
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
