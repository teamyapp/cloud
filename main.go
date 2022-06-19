package main

import (
	"database/sql"
	"log"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/dep"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/runner"
)

func main() {
	cfg, err := config.AppFromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Printf(
		"Git Commit: https://github.com/%s/%s/commit/%s\n",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)

	err = sqldb.Use(cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(sqlDB, sqldb.DefaultMigrationRoot)
		if err != nil {
			log.Println(err)
			return err
		}

		runnerConfig, err := runner.ServiceRunnerConfigFromEnv()
		if err != nil {
			log.Println(err)
			return err
		}

		webAPIBaseURL := dep.WebAPIBaseURL(cfg.WebAPIBaseURL)
		oauthProviders := []oauth.Provider{
			dep.InitGoogleOAuthProvider(
				webAPIBaseURL,
				dep.JWTSigningKey(cfg.JWTSigningKey),
				dep.ClientID(cfg.GoogleClientID),
				dep.ClientSecret(cfg.GoogleClientSecret)),
			dep.InitGitHubOAuthProvider(
				webAPIBaseURL,
				dep.ClientID(cfg.GitHubClientID),
				dep.ClientSecret(cfg.GitHubClientSecret)),
		}
		identityAPI, err := dep.InitIdentityAPI(
			sqlDB,
			oauthProviders,
			dep.AccessTokenTTL(cfg.AccessTokenTTL),
			dep.JWTSigningKey(cfg.JWTSigningKey),
			dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			log.Println(err)
			return err
		}

		generatorAPI, err := dep.InitGeneratorAPI(sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
		if err != nil {
			log.Println(err)
			return err
		}

		rn := runner.NewServiceRunner(runnerConfig, []runner.Service{
			identityAPI,
			generatorAPI,
		})

		rn.Start()
		return nil
	})
}
