//go:build wireinject

package dep

import (
	"net/http"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/cloud/app/api/web"
	"github.com/teamyapp/cloud/app/api/web/identity"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/idgen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/pubsub"
	"github.com/teamyapp/cloud/app/repo"
	"github.com/teamyapp/cloud/app/security"
	"github.com/teamyapp/cloud/app/service"
	"google.golang.org/grpc"
)

var repoSet = wire.NewSet(
	wire.Bind(new(repo.UserLinking), new(*repo.InMemoryUserLinking)),
	wire.Bind(new(repo.IDRange), new(*repo.InMemoryIDRange)),

	repo.NewInMemoryUserLinking,
	repo.NewInMemoryIDRange,
)

func InitWebAPIServer(cfg config.Config) (*http.ServeMux, error) {
	wire.Build(
		repoSet,
		wire.Bind(new(idgen.Factory), new(idgen.InMemoryIDGeneratorFactory)),
		wire.Bind(new(pubsub.PubSub), new(*pubsub.InMemory)),

		newJWTAuthority,
		newInMemoryIDRangeFactory,
		pubsub.NewInMemory,
		newIdentityService,
		web.NewAPIServer,
	)
	return &http.ServeMux{}, nil
}

func InitGRPCAPIServer(cfg config.Config) *grpc.Server {
	wire.Build(rpc.NewAPIServer)
	return nil
}

func newIdentityService(
	jwtAuthority security.JWTAuthority,
	pubSub pubsub.PubSub,
	idGeneratorFactory idgen.Factory,
	userLinkingRepo repo.UserLinking,
	cfg config.Config,
) (service.Identity, error) {
	urlString, err := identity.GetRootURL(cfg.WebAPIBaseURL)
	if err != nil {
		return service.Identity{}, err
	}
	var oauthProviders = []oauth.OAuth{
		oauth.NewGoogle(jwtAuthority, cfg.GoogleClientID, cfg.GoogleClientSecret, urlString),
	}
	return service.NewIdentity(
		jwtAuthority,
		pubSub,
		idGeneratorFactory,
		userLinkingRepo,
		oauthProviders,
		cfg.AccessTokenTTL,
		cfg.SignInTimeOut,
	)
}

func newJWTAuthority(cfg config.Config) security.JWTAuthority {
	return security.NewJWTAuthority(cfg.JWTSigningKey)
}

func newInMemoryIDRangeFactory(idRangeRepo repo.IDRange, cfg config.Config) idgen.InMemoryIDGeneratorFactory {
	return idgen.NewInMemoryIDGeneratorFactory(idRangeRepo, cfg.IDRangeLength)
}
