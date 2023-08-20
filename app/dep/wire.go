//go:build wireinject

package dep

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

type OAuthProviders []oauth.Provider
type AccessTokenTTL time.Duration
type GenRangeSize uint64
type JWTSigningKey string
type WebAPIBaseURL string

type ClientID string
type ClientSecret string

type S3Endpoint string
type S3AccessKeyID string
type S3AccessKey string
type S3BucketName string

var daoSet = wire.NewSet(
	wire.Bind(new(dao.UserLink), new(sqldb.UserLink)),
	wire.Bind(new(dao.AllocatedRange), new(sqldb.AllocatedRange)),
	wire.Bind(new(dao.SignInSession), new(sqldb.SignInSession)),
	wire.Bind(new(dao.ServiceAccount), new(sqldb.ServiceAccount)),
	wire.Bind(new(dao.OperationRelation), new(sqldb.OperationRelation)),
	wire.Bind(new(dao.Operation), new(sqldb.Operation)),
	wire.Bind(new(dao.UserGroup), new(sqldb.UserGroup)),
	wire.Bind(new(dao.UserGroupMember), new(sqldb.UserGroupMember)),
	wire.Bind(new(dao.Permission), new(sqldb.Permission)),
	wire.Bind(new(dao.ResourceType), new(sqldb.ResourceType)),
	wire.Bind(new(dao.Resource), new(sqldb.Resource)),
	wire.Bind(new(dao.ResourceRelation), new(sqldb.ResourceRelation)),
	wire.Bind(new(dao.UploadSession), new(sqldb.UploadSession)),
	wire.Bind(new(dao.FileMetadata), new(sqldb.FileMetadata)),
	wire.Bind(new(dao.ChunkMetadata), new(sqldb.ChunkMetadata)),
	sqldb.NewAllocatedRange,
	sqldb.NewUserLink,
	sqldb.NewSignInSession,
	sqldb.NewServiceAccount,
	sqldb.NewOperationRelation,
	sqldb.NewOperation,
	sqldb.NewUserGroup,
	sqldb.NewUserGroupMember,
	sqldb.NewPermission,
	sqldb.NewResourceType,
	sqldb.NewResource,
	sqldb.NewResourceRelation,
	sqldb.NewUploadSession,
	sqldb.NewFileMetadata,
	sqldb.NewChunkMetadata,
)

var storageSet = wire.NewSet(
	wire.Bind(new(storage.MapClient), new(storage.S3Bucket)),
	wire.Bind(new(storage.MapRequestHandlers), new(storage.S3Bucket)),
	newS3Bucket,
)

func InitTelemetryAPI(logger telemetry.Logger) *api.Telemetry {
	wire.Build(
		api.NewTelemetry,
	)
	return nil
}

func InitIdentityAPI(
	logger telemetry.Logger,
	sqlDB *sql.DB,
	oauthProviders OAuthProviders,
	accessTokenTTL AccessTokenTTL,
	jwtSigningKey JWTSigningKey,
	genRangeSize GenRangeSize,
) (api.Identity, error) {
	wire.Build(
		daoSet,
		newUniqueNumberGenFactory,
		newJWTAuthority,
		api.NewIdentity,
		newIdentityService,
	)
	return api.Identity{}, nil
}

func InitGeneratorAPI(
	logger telemetry.Logger,
	sqlDB *sql.DB,
	genRangeSize GenRangeSize,
) (*api.Generator, error) {
	wire.Build(
		daoSet,
		newUniqueNumberGenFactory,
		api.NewGenerator,
	)
	return nil, nil
}

func InitAuthorizationAPI(
	logger telemetry.Logger,
	sqlDB *sql.DB,
	genRangeSize GenRangeSize,
) (api.Authorization, error) {
	wire.Build(
		daoSet,
		newUniqueNumberGenFactory,
		api.NewAuthorization,
		service.NewAuthorization,
	)
	return api.Authorization{}, nil
}

func InitFileAPI(
	logger telemetry.Logger,
	env env.Environment,
	sqlDB *sql.DB,
	genRangeSize GenRangeSize,
	s3Endpoint S3Endpoint,
	s3AccessKeyID S3AccessKeyID,
	s3AccessKey S3AccessKey,
	s3BucketName S3BucketName,
) (api.File, error) {
	wire.Build(
		daoSet,
		storageSet,
		newUniqueNumberGenFactory,
		service.NewFile,
		api.NewFile,
	)
	return api.File{}, nil
}

func InitStreamAPI(
	logger telemetry.Logger,
	env env.Environment,
	sqlDB *sql.DB,
	s3Endpoint S3Endpoint,
	s3AccessKeyID S3AccessKeyID,
	s3AccessKey S3AccessKey,
	s3BucketName S3BucketName,
) (api.Stream, error) {
	wire.Build(
		daoSet,
		storageSet,
		service.NewStream,
		api.NewStream,
	)
	return api.Stream{}, nil
}

func newS3Bucket(
	logger telemetry.Logger,
	s3Endpoint S3Endpoint,
	s3AccessKeyID S3AccessKeyID,
	s3AccessKey S3AccessKey,
	s3BucketName S3BucketName,
	env env.Environment,
) (storage.S3Bucket, error) {
	return storage.NewS3Bucket(
		logger,
		string(s3Endpoint),
		string(s3AccessKeyID),
		string(s3AccessKey),
		env,
		string(s3BucketName))
}

func InitGitHubOAuthProvider(
	logger telemetry.Logger,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.GitHub {
	wire.Build(
		wire.Bind(new(network.Network), new(network.Socket)),
		wire.Bind(new(web.HTTPClient), new(*http.Client)),

		network.NewSocket,
		web.NewHTTPClient,
		newGithubOAuthProvider,
	)
	return oauth.GitHub{}
}

func InitGoogleOAuthProvider(
	logger telemetry.Logger,
	webAPIBaseURL WebAPIBaseURL,
	jwtSigningKey JWTSigningKey,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Google {
	wire.Build(
		wire.Bind(new(network.Network), new(network.Socket)),
		wire.Bind(new(web.HTTPClient), new(*http.Client)),

		network.NewSocket,
		web.NewHTTPClient,
		newJWTAuthority,
		newGoogleOAuthProvider,
	)
	return oauth.Google{}
}

func InitSlackOAuthProvider(
	logger telemetry.Logger,
	webAPIBaseURL WebAPIBaseURL,
	jwtSigningKey JWTSigningKey,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Slack {
	wire.Build(
		wire.Bind(new(network.Network), new(network.Socket)),
		wire.Bind(new(web.HTTPClient), new(*http.Client)),

		network.NewSocket,
		web.NewHTTPClient,
		newJWTAuthority,
		newSlackOAuthProvider,
	)
	return oauth.Slack{}
}

func newGithubOAuthProvider(
	httpClient web.HTTPClient,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.GitHub {
	return oauth.NewGitHub(httpClient, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func newGoogleOAuthProvider(
	httpClient web.HTTPClient,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Google {
	return oauth.NewGoogle(httpClient, jwtAuthority, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func newSlackOAuthProvider(
	httpClient web.HTTPClient,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Slack {
	return oauth.NewSlack(httpClient, jwtAuthority, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func newJWTAuthority(logger telemetry.Logger, signingKey JWTSigningKey) security.JWTAuthority {
	return security.NewJWTAuthority(logger, string(signingKey))
}

func newUniqueNumberGenFactory(
	logger telemetry.Logger,
	allocatedRangeDao dao.AllocatedRange,
	genRangeSize GenRangeSize,
) *service.UniqueNumberGenRegistry {
	return service.NewUniqueNumberGenRegistry(logger, allocatedRangeDao, uint64(genRangeSize))
}

func newIdentityService(
	logger telemetry.Logger,
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	serviceAccountDao dao.ServiceAccount,
	uniqueNumberRegistry *service.UniqueNumberGenRegistry,
	jwtAuthority security.JWTAuthority,
	oauthProviders OAuthProviders,
	accessTokenTLL AccessTokenTTL,
) (service.Identity, error) {
	return service.NewIdentity(
		logger,
		signInSessionDao,
		userLinkDao,
		serviceAccountDao,
		uniqueNumberRegistry,
		jwtAuthority,
		oauthProviders,
		time.Duration(accessTokenTLL))
}
