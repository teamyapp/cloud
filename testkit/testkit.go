package testkit

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/daotest"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics/metricstest"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/storage/storagetest"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web/webtest"
)

const appName = "cloud"
const serviceName = "backend"

var serviceLabels = []string{appName, serviceName}
var fullServiceName = strings.Join(serviceLabels, "-")

type TestKit struct {
	ServiceInstanceRunner         runner.ServiceRunner
	InMemoryDB                    *dbtest.InMemoryDB
	UniqueNumberGeneratorRegistry *service.UniqueNumberGenRegistry
	IdentityService               service.Identity
	AuthorizationService          service.Authorization
	FileService                   service.File
}

func New(cfg Config, network network.Network) (TestKit, *errs.Error) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	runnerConfig := runner.ServiceRunnerConfig{
		WebServerPort:        cfg.WebServerPort,
		GRPCServerPort:       cfg.GRPCServerPort,
		MonitoringServerPort: 10000,
		RequestTimeout:       10 * time.Second,
		EnableTracing:        false,
		IdentityAPIEndpoint:  fmt.Sprintf("%v/identity", cfg.WebAPIBaseURL),
	}

	httpClient := webtest.InsecureHTTPClient(network)
	githubOAuth := oauth.NewGitHub(
		httpClient,
		cfg.WebAPIBaseURL,
		cfg.GithubClientID,
		cfg.GithubClientSecret)

	jwtAuthority := security.NewJWTAuthority(logger, cfg.JWTSigningKey)
	googleOAuth := oauth.NewGoogle(
		httpClient,
		jwtAuthority,
		cfg.WebAPIBaseURL,
		cfg.GoogleClientID,
		cfg.GoogleClientSecret)
	slackOAuth := oauth.NewSlack(
		httpClient,
		jwtAuthority,
		cfg.WebAPIBaseURL,
		cfg.SlackClientID,
		cfg.SlackClientSecret)

	inMemoryDB := dbtest.NewInMemoryDB()
	inMemoryDB.CreateTable(daotest.AllocatedRangeTableName)
	inMemoryDB.CreateTable(daotest.ChunkMetadataTableName)
	inMemoryDB.CreateTable(daotest.FileMetadataTableName)
	inMemoryDB.CreateTable(daotest.OperationTableName)
	inMemoryDB.CreateTable(daotest.OperationRelationTableName)
	inMemoryDB.CreateTable(daotest.PermissionTableName)
	inMemoryDB.CreateTable(daotest.ResourceTableName)
	inMemoryDB.CreateTable(daotest.ResourceRelationTableName)
	inMemoryDB.CreateTable(daotest.ResourceTypeTableName)
	inMemoryDB.CreateTable(daotest.ServiceAccountTableName)
	inMemoryDB.CreateTable(daotest.SignInSessionTableName)
	inMemoryDB.CreateTable(daotest.UploadSessionTableName)
	inMemoryDB.CreateTable(daotest.UserGroupTableName)
	inMemoryDB.CreateTable(daotest.PermissionTableName)
	inMemoryDB.CreateTable(daotest.UserGroupMemberTableName)
	inMemoryDB.CreateTable(daotest.UserLinkTableName)

	allocatedRangeDao := daotest.NewAllocatedRange(inMemoryDB)
	uniqueNumberGeneratorRegistry := service.NewUniqueNumberGenRegistry(
		logger,
		allocatedRangeDao,
		cfg.GenUniqueNumberRangeSize)
	signInSessionDao := daotest.NewSignInSession(inMemoryDB)
	userLinkDao := daotest.NewUserLink(inMemoryDB)
	serviceAccountDao := daotest.NewServiceAccount(inMemoryDB)
	oauthProviders := []oauth.Provider{
		githubOAuth,
		googleOAuth,
		slackOAuth,
	}
	identityService, err := service.NewIdentity(
		logger,
		signInSessionDao,
		userLinkDao,
		serviceAccountDao,
		uniqueNumberGeneratorRegistry,
		jwtAuthority,
		oauthProviders,
		cfg.AccessTokenTTL)
	if err != nil {
		return TestKit{}, errs.NewError(errs.Unknown, err.Error())
	}

	identityAPI := api.NewIdentity(logger, identityService)
	generatorAPI := api.NewGenerator(logger, uniqueNumberGeneratorRegistry)

	resourceRelationDao := daotest.NewResourceRelation(inMemoryDB)
	userGroupMemberDao := daotest.NewUserGroupMember(inMemoryDB)
	permissionDao := daotest.NewPermission(inMemoryDB)
	operationRelationDao := daotest.NewOperationRelation(inMemoryDB)
	operationDao := daotest.NewOperation(inMemoryDB)
	resourceTypeDao := daotest.NewResourceType(inMemoryDB)
	resourceDao := daotest.NewResource(inMemoryDB)
	userGroupDao := daotest.NewUserGroup(inMemoryDB)
	authorizationService, err := service.NewAuthorization(
		logger,
		resourceRelationDao,
		userGroupMemberDao,
		permissionDao,
		operationRelationDao,
		operationDao,
		resourceTypeDao,
		resourceDao,
		userGroupDao,
		uniqueNumberGeneratorRegistry,
	)
	if err != nil {
		return TestKit{}, errs.NewError(errs.Unknown, err.Error())
	}

	authorizationAPI := api.NewAuthorization(logger, authorizationService)

	uploadSessionDao := daotest.NewUploadSession(inMemoryDB)
	fileMetadataDao := daotest.NewFileMetadata(inMemoryDB)
	chunkMetadataDao := daotest.NewChunkMetadata(inMemoryDB)
	inMemoryMapBackend := storagetest.NewInMemoryMap()
	fileService, err := service.NewFile(
		logger,
		inMemoryMapBackend,
		uniqueNumberGeneratorRegistry,
		uploadSessionDao,
		fileMetadataDao,
		chunkMetadataDao)
	if err != nil {
		return TestKit{}, errs.NewError(errs.Unknown, err.Error())
	}

	fileAPI := api.NewFile(logger, fileService)
	telemetryAPI := api.NewTelemetry(logger)

	serviceRunner := runner.NewServiceRunnerBuilder(
		logger,
		network,
		metricstest.NewNoopMetrics(),
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
	return TestKit{
		ServiceInstanceRunner:         serviceRunner,
		InMemoryDB:                    inMemoryDB,
		UniqueNumberGeneratorRegistry: uniqueNumberGeneratorRegistry,
		IdentityService:               identityService,
		AuthorizationService:          authorizationService,
		FileService:                   fileService,
	}, nil
}

func StartServiceInstance(
	config Config,
	virtualNetwork *networktest.VirtualNetwork,
	serviceRunner runner.ServiceRunner,
) {
	waitBootstrapCh := make(chan struct{})
	cloudBackendProxyRoutes := proxyRoutes(
		config.WebServerPort,
		config.GRPCServerPort)
	go func() *errs.Error {
		internalErr := serviceRunner.Start(func(listeners []net.Listener) *errs.Error {
			for _, proxyRoute := range cloudBackendProxyRoutes {
				for _, listener := range listeners {
					if proxyRoute.MatchTarget(listener.Addr()) {
						bindErr := virtualNetwork.BindProxyEndpoint(proxyRoute.Endpoint, listener)
						if bindErr != nil {
							return bindErr
						}
					}
				}
			}

			waitBootstrapCh <- struct{}{}
			return nil
		})
		panic(internalErr)
	}()
	<-waitBootstrapCh
}
