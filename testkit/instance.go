package testkit

import (
	"os"
	"strings"
	"time"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/daotest"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/app/storage/storagetest"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web/webtest"
)

const appName = "cloud"
const serviceName = "backend"

var serviceLabels = []string{appName, serviceName}
var fullServiceName = strings.Join(serviceLabels, "-")

type Config struct {
	GenUniqueNumberRangeSize uint64
	JWTSigningKey            string
	AccessTokenTTL           time.Duration
	WebAPIBaseURL            string
	GithubClientID           string
	GithubClientSecret       string
	GoogleClientID           string
	GoogleClientSecret       string
	SlackClientID            string
	SlackClientSecret        string
	WebServerPort            int
	GRPCServerPort           int
}

type TestKit struct {
	InstanceRunner runner.ServiceRunner
	Refs           Refs
}

type Refs struct {
	InMemoryDB                   *dbtest.InMemoryDB
	UniqueNumberGeneratorFactory gen.UniqueNumberFactory
	IdentityService              service.Identity
	AuthorizationService         service.Authorization
	FileService                  service.File
}

func New(cfg Config, network network.Network) (TestKit, *errs.Error) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	dataCollector := telemetry.NewDataCollector(logger)
	virtualNetwork := networktest.NewVirtualNetwork()
	runnerConfig := runner.ServiceRunnerConfig{
		WebServerPort:        cfg.WebServerPort,
		GRPCServerPort:       cfg.GRPCServerPort,
		MonitoringServerPort: 10000,
		RequestTimeout:       10 * time.Second,
		EnableTracing:        false,
	}

	httpClient := webtest.InsecureHTTPClient(network)
	githubOAuth := oauth.NewGitHub(
		dataCollector,
		httpClient,
		cfg.WebAPIBaseURL,
		cfg.GithubClientID,
		cfg.GithubClientSecret)

	jwtAuthority := security.NewJWTAuthority(dataCollector, cfg.JWTSigningKey)
	googleOAuth := oauth.NewGoogle(
		dataCollector,
		httpClient,
		jwtAuthority,
		cfg.WebAPIBaseURL,
		cfg.GoogleClientID,
		cfg.GoogleClientSecret)
	slackOAuth := oauth.NewSlack(
		dataCollector,
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
	uniqueNumberGeneratorFactory := gen.NewUniqueNumberFactory(
		dataCollector,
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
		dataCollector,
		signInSessionDao,
		userLinkDao,
		serviceAccountDao,
		uniqueNumberGeneratorFactory,
		jwtAuthority,
		oauthProviders,
		cfg.AccessTokenTTL)
	if err != nil {
		return TestKit{}, &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
	}

	identityAPI := api.NewIdentity(dataCollector, identityService)
	generatorAPI := api.NewGenerator(dataCollector, uniqueNumberGeneratorFactory)

	resourceRelationDao := daotest.NewResourceRelation(inMemoryDB)
	userGroupMemberDao := daotest.NewUserGroupMember(inMemoryDB)
	permissionDao := daotest.NewPermission(inMemoryDB)
	operationRelationDao := daotest.NewOperationRelation(inMemoryDB)
	operationDao := daotest.NewOperation(inMemoryDB)
	resourceTypeDao := daotest.NewResourceType(inMemoryDB)
	resourceDao := daotest.NewResource(inMemoryDB)
	userGroupDao := daotest.NewUserGroup(inMemoryDB)
	authorizationService, err := service.NewAuthorization(
		dataCollector,
		resourceRelationDao,
		userGroupMemberDao,
		permissionDao,
		operationRelationDao,
		operationDao,
		resourceTypeDao,
		resourceDao,
		userGroupDao,
		uniqueNumberGeneratorFactory,
	)
	if err != nil {
		return TestKit{}, &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
	}

	authorizationAPI := api.NewAuthorization(dataCollector, authorizationService)

	uploadSessionDao := daotest.NewUploadSession(inMemoryDB)
	fileMetadataDao := daotest.NewFileMetadata(inMemoryDB)
	chunkMetadataDao := daotest.NewChunkMetadata(inMemoryDB)
	inMemoryMapBackend := storagetest.NewInMemoryMap()
	fileService, err := service.NewFile(
		dataCollector,
		inMemoryMapBackend,
		uniqueNumberGeneratorFactory,
		uploadSessionDao,
		fileMetadataDao,
		chunkMetadataDao)
	if err != nil {
		return TestKit{}, &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
	}

	fileAPI := api.NewFile(dataCollector, fileService)
	telemetryAPI := api.NewTelemetry(dataCollector)

	serviceRunner := runner.NewServiceRunnerBuilder(
		dataCollector,
		virtualNetwork,
		metrics.NewPrometheus(appName, serviceName, env.DevelopmentEnv),
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
		InstanceRunner: serviceRunner,
		Refs: Refs{
			InMemoryDB:                   inMemoryDB,
			UniqueNumberGeneratorFactory: uniqueNumberGeneratorFactory,
			IdentityService:              identityService,
			AuthorizationService:         authorizationService,
			FileService:                  fileService,
		},
	}, nil
}
