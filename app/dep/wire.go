//go:build wireinject

package dep

import (
	"database/sql"
	"time"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/app/storage"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/security"
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

type ServiceName string
type Commit string

func InitDataCollector(serviceName ServiceName, commit Commit, visibleLevel obs.LogLevel) obs.DataCollector {
	wire.Build(
		newLogger,
		obs.NewDataCollector,
	)
	return obs.DataCollector{}
}

func InitGoogleOAuthProvider(
	dataCollector obs.DataCollector,
	webAPIBaseURL WebAPIBaseURL,
	jwtSigningKey JWTSigningKey,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Google {
	wire.Build(
		newJWTAuthority,
		newGoogleOAuthProvider,
	)
	return oauth.Google{}
}

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
	wire.Bind(new(storage.MapBackend), new(storage.S3Bucket)),
	newS3Bucket,
)

func InitTelemetryAPI(dataCollector obs.DataCollector) *api.Telemetry {
	wire.Build(
		api.NewTelemetry,
	)
	return nil
}

func InitIdentityAPI(
	dataCollector obs.DataCollector,
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
	dataCollector obs.DataCollector,
	sqlDB *sql.DB,
	genRangeSize GenRangeSize,
) (api.Generator, error) {
	wire.Build(
		daoSet,
		newUniqueNumberGenFactory,
		api.NewGenerator,
	)
	return api.Generator{}, nil
}

func InitAuthorizationAPI(
	dataCollector obs.DataCollector,
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
	dataCollector obs.DataCollector,
	env config.Environment,
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

func newS3Bucket(
	dataCollector obs.DataCollector,
	s3Endpoint S3Endpoint,
	s3AccessKeyID S3AccessKeyID,
	s3AccessKey S3AccessKey,
	s3BucketName S3BucketName,
	env config.Environment,
) (storage.S3Bucket, error) {
	return storage.NewS3Bucket(
		dataCollector,
		string(s3Endpoint),
		string(s3AccessKeyID),
		string(s3AccessKey),
		env,
		string(s3BucketName))
}

func newGoogleOAuthProvider(
	dataCollector obs.DataCollector,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Google {
	return oauth.NewGoogle(dataCollector, jwtAuthority, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func InitGitHubOAuthProvider(
	dataCollector obs.DataCollector,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.GitHub {
	return oauth.NewGitHub(dataCollector, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func newJWTAuthority(dataCollector obs.DataCollector, signingKey JWTSigningKey) security.JWTAuthority {
	return security.NewJWTAuthority(dataCollector, string(signingKey))
}

func newUniqueNumberGenFactory(
	dataCollector obs.DataCollector,
	allocatedRangeDao dao.AllocatedRange,
	genRangeSize GenRangeSize) gen.UniqueNumberFactory {
	return gen.NewUniqueNumberFactory(dataCollector, allocatedRangeDao, uint64(genRangeSize))
}

func newIdentityService(
	dataCollector obs.DataCollector,
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	serviceAccountDao dao.ServiceAccount,
	uniqueNumberFactory gen.UniqueNumberFactory,
	jwtAuthority security.JWTAuthority,
	oauthProviders OAuthProviders,
	accessTokenTLL AccessTokenTTL,
) (service.Identity, error) {
	return service.NewIdentity(
		dataCollector,
		signInSessionDao,
		userLinkDao,
		serviceAccountDao,
		uniqueNumberFactory,
		jwtAuthority,
		oauthProviders,
		time.Duration(accessTokenTLL))
}

func newLogger(serviceName ServiceName, commit Commit, visibleLevel obs.LogLevel) obs.Logger {
	return obs.NewServiceLogger(string(serviceName),
		obs.NewCommitLogger(string(commit),
			obs.NewRequestLogger(
				obs.NewClientLogger(
					obs.NewRawLogger(visibleLevel)))))
}
