//go:build wireinject

package dep

import (
	"database/sql"
	"time"

	"github.com/teamyapp/cloud/app/dao/dao_test"
	"github.com/teamyapp/cloud/app/entity"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/security"
)

type OAuthProviders []oauth.Provider
type AccessTokenTTL time.Duration
type GenRangeSize uint64
type JWTSigningKey string
type WebAPIBaseURL string

type ClientID string
type ClientSecret string

func InitGoogleOAuthProvider(
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
	wire.Bind(new(dao.OperationRelation), new(dao_test.OperationRelation)),
	wire.Bind(new(dao.ResourceRelation), new(dao_test.ResourceRelation)),
	wire.Bind(new(dao.UserGroupMember), new(dao_test.UserGroupMember)),
	wire.Bind(new(dao.Permission), new(dao_test.Permission)),
	sqldb.NewAllocatedRange,
	sqldb.NewUserLink,
	sqldb.NewSignInSession,
	sqldb.NewServiceAccount,
	newOperationRelation,
	newResourceRelation,
	newUserGroupMember,
	newPermission,
)

func InitIdentityAPI(
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
	sqlDB *sql.DB,
) (api.Authorization, error) {
	wire.Build(
		daoSet,
		api.NewAuthorization,
		newAuthorizationService,
	)
	return api.Authorization{}, nil
}

func newGoogleOAuthProvider(
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.Google {
	return oauth.NewGoogle(jwtAuthority, string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func InitGitHubOAuthProvider(
	webAPIBaseURL WebAPIBaseURL,
	clientID ClientID,
	clientSecret ClientSecret,
) oauth.GitHub {
	return oauth.NewGitHub(string(webAPIBaseURL), string(clientID), string(clientSecret))
}

func newJWTAuthority(signingKey JWTSigningKey) security.JWTAuthority {
	return security.NewJWTAuthority(string(signingKey))
}

func newUniqueNumberGenFactory(allocatedRangeDao dao.AllocatedRange, genRangeSize GenRangeSize) gen.UniqueNumberFactory {
	return gen.NewUniqueNumberFactory(allocatedRangeDao, uint64(genRangeSize))
}

func newIdentityService(
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	serviceAccountDao dao.ServiceAccount,
	uniqueNumberFactory gen.UniqueNumberFactory,
	jwtAuthority security.JWTAuthority,
	oauthProviders OAuthProviders,
	accessTokenTLL AccessTokenTTL,
) (service.Identity, error) {
	return service.NewIdentity(
		signInSessionDao,
		userLinkDao,
		serviceAccountDao,
		uniqueNumberFactory,
		jwtAuthority,
		oauthProviders,
		time.Duration(accessTokenTLL))
}

func newAuthorizationService(
	resourceRelationDao dao.ResourceRelation,
	userGroupMemberDao dao.UserGroupMember,
	permissionDao dao.Permission,
	operationRelationDao dao.OperationRelation,
) (service.Authorization, error) {
	return service.NewAuthorization(
		resourceRelationDao,
		userGroupMemberDao,
		permissionDao,
		operationRelationDao,
	), nil
}

func newOperationRelation() dao_test.OperationRelation {
	return dao_test.NewOperationRelation(fakeOperationRelationDao)
}

func newResourceRelation() dao_test.ResourceRelation {
	return dao_test.NewResourceRelation(fakeResourceRelationDao)
}

func newUserGroupMember() dao_test.UserGroupMember {
	return dao_test.NewUserGroupMember(fakeUserGroupMemberDao)
}

func newPermission() dao_test.Permission {
	return dao_test.NewPermission(fakePermissionDao)
}

// TODO: replace with sqldb
var fakeOperationRelationDao = []entity.OperationRelation{}

var fakeResourceRelationDao = []entity.ResourceRelation{}

var fakeUserGroupMemberDao = []entity.UserGroupMember{}

var fakePermissionDao = []entity.Permission{}
