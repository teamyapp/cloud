package identity

import (
	"net/http"

	"github.com/teamyapp/cloud/app/api/web/route"
	"github.com/teamyapp/cloud/app/channel"
	"github.com/teamyapp/cloud/app/service"
)

func GetRoutes(
	webSocketOriginChecker channel.WebSocketOriginChecker,
	identityService service.Identity,
) []route.Route {
	return []route.Route{
		{
			Path:       "/identity/sign-in/oauth/{oauth-provider}",
			Method:     http.MethodGet,
			HandleFunc: newOAuthSignInHandler(identityService),
		},
		{
			Path:       "/identity/sign-in/oauth/{oauth-provider}/callback",
			Method:     http.MethodGet,
			HandleFunc: newOAuthSignInFinishHandler(identityService),
		},
		{
			Path:       "/identity/sign-in/session/new",
			Method:     http.MethodGet,
			HandleFunc: newGetNewSessionIDHandler(identityService),
		},
		{
			Path:       "/identity/sign-in/session/{session-id}/subscribe",
			Method:     http.MethodGet,
			HandleFunc: newSubscribeSessionHandler(webSocketOriginChecker, identityService),
		},
		{
			Path:       "/identity/verify-token",
			Method:     http.MethodPost,
			HandleFunc: newVerifyAccessTokenHandler(identityService),
		},
	}
}

func GetRootURL(webBaseURL string) (string, error) {
	return route.WithChildPath(webBaseURL, "identity")
}
