//go:build wireinject

package dep

import (
	"database/sql"
	"time"

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
	sqldb.NewAllocatedRange,
	sqldb.NewUserLink,
	sqldb.NewSignInSession,
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
	uniqueNumberFactory gen.UniqueNumberFactory,
	jwtAuthority security.JWTAuthority,
	oauthProviders OAuthProviders,
	accessTokenTLL AccessTokenTTL,
) (service.Identity, error) {
	return service.NewIdentity(
		signInSessionDao,
		userLinkDao,
		uniqueNumberFactory,
		jwtAuthority,
		oauthProviders,
		time.Duration(accessTokenTLL))
}
