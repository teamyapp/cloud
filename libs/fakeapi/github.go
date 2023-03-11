package fakeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/identity"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

const githubAppName = "github"
const githubServiceName = "backend"

var githubServiceLabels = []string{githubAppName, githubServiceName}
var githubFullServiceName = strings.Join(githubServiceLabels, "-")

type thirdPartyClient struct {
	secret             string
	authorizationCodes map[string]uint64
}

type GithubUser struct {
	ID    uint64 `json:"id"`
	Login string `json:"login"`
}

type Github struct {
	dataCollector     telemetry.DataCollector
	githubUsers       map[uint64]*GithubUser
	thirdPartyClients map[string]*thirdPartyClient
	accessTokenToUser map[string]*GithubUser
}

var _ runner.Service = (*Github)(nil)

func (g *Github) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join("/login", "oauth", "authorize"),
			HandlerFunc: g.webOAuthAuthorize,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join("/login", "oauth", "select_user"),
			HandlerFunc: g.webOAuthSelectUser,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join("/login", "oauth", "access_token"),
			HandlerFunc: g.webOAuthGetAccessToken,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join("/user"),
			HandlerFunc: g.webOAuthGetUser,
		},
	})
	return nil
}

func (g *Github) webOAuthAuthorize(writer http.ResponseWriter, request *http.Request) {
	reqQuery := request.URL.Query()
	clientID := reqQuery.Get("client_id")
	_, ok := g.thirdPartyClients[clientID]
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("client id not found: clientID=%v", clientID)))
		return
	}

	rawRedirectURI := reqQuery.Get("redirect_uri")
	_, err := url.Parse(rawRedirectURI)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("invalid redirect uri: redirect_uri=%v", rawRedirectURI)))
		return
	}

	u, err := url.Parse("https://api.github.com/login/oauth/select_user")
	u.RawQuery = request.URL.RawQuery
	writer.Write([]byte(u.String()))
}

func (g *Github) webOAuthSelectUser(writer http.ResponseWriter, request *http.Request) {
	reqQuery := request.URL.Query()
	clientID := reqQuery.Get("client_id")
	_, ok := g.thirdPartyClients[clientID]
	if !ok {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("client id not found: clientID=%v", clientID)))
		return
	}

	rawRedirectURI := reqQuery.Get("redirect_uri")
	redirectURI, err := url.Parse(rawRedirectURI)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("invalid redirect uri: redirect_uri=%v", rawRedirectURI)))
		return
	}

	state := reqQuery.Get("state")
	resQuery := redirectURI.Query()
	resQuery.Set("state", state)

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	defer request.Body.Close()
	reqBody := struct {
		UserID uint64 `json:"user_id"`
	}{}
	err = json.Unmarshal(buf, &reqBody)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationCode := randgen.String(randgen.Base62, 20)
	client := g.thirdPartyClients[clientID]
	client.authorizationCodes[authorizationCode] = reqBody.UserID

	resQuery.Set("code", authorizationCode)
	redirectURI.RawQuery = resQuery.Encode()
	http.Redirect(writer, request, redirectURI.String(), http.StatusSeeOther)
}

func (g *Github) webOAuthGetAccessToken(writer http.ResponseWriter, request *http.Request) {
	buf, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	defer request.Body.Close()
	reqBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
	}{}
	err = json.Unmarshal(buf, &reqBody)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	client, ok := g.thirdPartyClients[reqBody.ClientID]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte(fmt.Sprintf("client not found: clientID=%v", reqBody.ClientID)))
		return
	}

	if reqBody.ClientSecret != client.secret {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte(fmt.Sprintf("invalid clientID or secret: clientID=%v", reqBody.ClientID)))
		return
	}

	userID, ok := client.authorizationCodes[reqBody.Code]
	if !ok {
		writer.WriteHeader(http.StatusForbidden)
		writer.Write([]byte(fmt.Sprintf("invalid authorization code=%v", reqBody.ClientID)))
		return
	}

	resBody := struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}{
		Scope:     "repo",
		TokenType: "bearer",
	}

	randomTokenPart := randgen.String(randgen.Base62, 20)
	accessToken := fmt.Sprintf("gho_%v", randomTokenPart)
	g.accessTokenToUser[accessToken] = g.githubUsers[userID]
	resBody.AccessToken = accessToken

	delete(client.authorizationCodes, reqBody.Code)
	ct := context.Background()
	web.WriteJSONToResponse(ct, g.dataCollector, writer, resBody)
}

func (g *Github) webOAuthGetUser(writer http.ResponseWriter, request *http.Request) {
	accessToken, internalErr := identity.GetBearerToken(request)
	if internalErr != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	githubUser, ok := g.accessTokenToUser[accessToken]
	if !ok {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	buf, err := json.Marshal(githubUser)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Write(buf)
}

func (g *Github) RegisterClient(clientID string, secret string) {
	g.thirdPartyClients[clientID] = &thirdPartyClient{
		secret:             secret,
		authorizationCodes: map[string]uint64{},
	}
}

func (g *Github) RegisterUser(user *GithubUser) {
	g.githubUsers[user.ID] = user
}

func NewGithub(dataCollector telemetry.DataCollector) *Github {
	return &Github{
		dataCollector:     dataCollector,
		githubUsers:       map[uint64]*GithubUser{},
		thirdPartyClients: map[string]*thirdPartyClient{},
		accessTokenToUser: map[string]*GithubUser{},
	}
}

type TestKit struct {
	ServiceInstanceRunner runner.ServiceRunner
	Refs                  Refs
}

type Refs struct {
	FakeGithubAPI *Github
}

type GithubTestKitConfig struct {
	WebServerPort  int
	GRPCServerPort int
}

func NewGithubTestKit(cfg GithubTestKitConfig, network network.Network) TestKit {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	dataCollector := telemetry.NewDataCollector(logger)
	runnerConfig := runner.ServiceRunnerConfig{
		WebServerPort:        cfg.WebServerPort,
		GRPCServerPort:       cfg.GRPCServerPort,
		MonitoringServerPort: 1000,
		RequestTimeout:       10 * time.Second,
		EnableTracing:        false,
	}
	fakeGithubAPI := NewGithub(dataCollector)
	serviceRunner := runner.NewServiceRunnerBuilder(
		dataCollector,
		network,
		metrics.NewPrometheus(githubAppName, githubServiceName, env.DevelopmentEnv),
		runnerConfig,
		githubFullServiceName,
		[]runner.Service{
			fakeGithubAPI,
		}).
		IncludeIdentityWebFunc(func(request *http.Request) bool {
			return false
		}).
		Build()
	return TestKit{
		ServiceInstanceRunner: serviceRunner,
		Refs: Refs{
			FakeGithubAPI: fakeGithubAPI,
		},
	}
}

func StartGithubServiceInstance(
	githubWebServerPort int,
	virtualNetwork *networktest.VirtualNetwork,
	serviceRunner runner.ServiceRunner,
) {
	waitBootstrapCh := make(chan struct{})
	cloudBackendProxyRoutes := githubProxyRoutes(githubWebServerPort)
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

func githubProxyRoutes(webServerPort int) []networktest.ProxyRoute {
	return []networktest.ProxyRoute{
		{
			Endpoint: "github.com:80",
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webServerPort))
			},
		},
		{
			Endpoint: "api.github.com:80",
			MatchTarget: func(addr net.Addr) bool {
				return addr.Network() == "tcp" &&
					strings.HasSuffix(addr.String(), fmt.Sprintf(":%d", webServerPort))
			},
		},
	}
}
