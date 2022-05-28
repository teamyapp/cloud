package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/cloud/app/api/web"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/dep"
	"github.com/teamyapp/cloud/app/oauth"
)

func main() {
	cfg, err := config.AppConfigFromEnv()
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
			panic(err)
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			StartWebServer(cfg, sqlDB)
		}()

		go func() {
			defer wg.Done()
			StartGrpcServer(cfg, sqlDB)
		}()
		wg.Wait()
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func StartWebServer(cfg config.Config, sqlDB *sql.DB) {
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
	identityAPI, err := dep.InitIdentityWebService(
		sqlDB,
		oauthProviders,
		dep.AccessTokenTTL(cfg.AccessTokenTTL),
		dep.JWTSigningKey(cfg.JWTSigningKey),
		dep.GenRangeSize(cfg.GenRangeSize))
	if err != nil {
		panic(err)
	}

	webServer := web.NewServer([]web.Service{identityAPI})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Web server started at port %d\n", cfg.WebAPIPort)
	if err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebAPIPort), webServer); err != nil {
		panic(err)
	}
}

func StartGrpcServer(cfg config.Config, sqlDB *sql.DB) {
	lis, err := net.Listen("tcp", fmt.Sprintf(fmt.Sprintf(":%d", cfg.GRPCAPIPort)))
	if err != nil {
		panic(err)
	}

	genRPCService, err := dep.InitGeneratorRPCService(sqlDB, dep.GenRangeSize(cfg.GenRangeSize))
	if err != nil {
		panic(err)
	}

	s := rpc.NewServer([]rpc.Service{genRPCService})
	fmt.Printf("GRPC server started at port %d\n", cfg.GRPCAPIPort)
	if err := s.Serve(lis); err != nil {
		panic(err)
	}
}
