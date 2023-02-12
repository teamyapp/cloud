package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

const EnvConfigErr errs.ErrorCode = "EnvConfig"

type WebRoute struct {
	Path        string
	Method      string
	HandlerFunc http.HandlerFunc
}

type ServiceRunnerConfig struct {
	WebServerPort       int           `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9011"`
	GRPCServerPort      int           `envconfig:"SERVICE_RUNNER_GRPC_SERVER_PORT" default:"9012"`
	IdentityAPIEndpoint string        `envconfig:"SERVICE_RUNNER_IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	RequestTimeout      time.Duration `envconfig:"REQUEST_TIMEOUT" default:"10s"`
}

func ServiceRunnerConfigFromEnv(dataCollector telemetry.DataCollector) (ServiceRunnerConfig, *errs.Error) {
	cfg := ServiceRunnerConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		internalErr := &errs.Error{
			Code:     EnvConfigErr,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return ServiceRunnerConfig{}, internalErr
	}

	return cfg, nil
}

type ServiceRunner struct {
	dataCollector          telemetry.DataCollector
	config                 ServiceRunnerConfig
	httpClient             web.HTTPClient
	webRouter              *mux.Router
	gRPCServer             *grpc.Server
	services               []Service
	includeIdentityWebFunc middleware.IncludeIdentityWebFunc
}

func (s *ServiceRunner) Start() {
	for _, service := range s.services {
		err := service.Start(s)
		if err != nil {
			s.dataCollector.Logger.Log(telemetry.Fatal, telemetry.Props{telemetry.CauseProp: err})
			panic(err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.startWebServer()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.startGRPCServer()
	}()
	wg.Wait()
}

func (s *ServiceRunner) startWebServer() {
	s.dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Web server started at %v", s.config.WebServerPort),
	})
	serveMux := http.NewServeMux()
	middlewares := []middleware.Middleware[http.HandlerFunc]{
		middleware.ServerHTTPEnableCORS,
		middleware.ServerHTTPWithRequestID(s.dataCollector),
		middleware.ServerHTTPWithTimeout(s.config.RequestTimeout),
		middleware.ServerHTTPLogRequest(s.dataCollector),
		middleware.ServerHTTPWithIdentity(
			s.dataCollector,
			s.httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
		middleware.ServerWebSocketWithIdentity(
			s.dataCollector,
			s.httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
	}
	handlerFunc := middleware.WithMiddlewares[http.HandlerFunc](s.webRouter.ServeHTTP, middlewares)
	serveMux.HandleFunc("/", handlerFunc)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebServerPort), serveMux); err != nil {
		panic(err)
	}
}

func (s *ServiceRunner) startGRPCServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCServerPort))
	if err != nil {
		panic(err)
	}

	s.dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner gRPC server started at %v", s.config.GRPCServerPort),
	})
	err = s.gRPCServer.Serve(lis)
	if err != nil {
		s.dataCollector.Logger.Log(telemetry.Fatal, telemetry.Props{telemetry.CauseProp: err})
		panic(err)
	}
}

func (s *ServiceRunner) RegisterWebRoutes(routes []WebRoute) {
	for _, route := range routes {
		s.webRouter.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
	}
}

func (s *ServiceRunner) WithGRPCServer(withGRPCServer func(server *grpc.Server)) {
	withGRPCServer(s.gRPCServer)
}

type ServiceRunnerBuilder struct {
	dataCollector           telemetry.DataCollector
	config                  ServiceRunnerConfig
	httpClient              web.HTTPClient
	services                []Service
	includeIdentityWebFunc  middleware.IncludeIdentityWebFunc
	includeIdentityGRPCFunc middleware.IncludeIdentityGRPCFunc
}

func (s *ServiceRunnerBuilder) IncludeIdentityWebFunc(
	includeIdentityWebFunc middleware.IncludeIdentityWebFunc,
) *ServiceRunnerBuilder {
	s.includeIdentityWebFunc = includeIdentityWebFunc
	return s
}

func (s *ServiceRunnerBuilder) IncludeIdentityGRPCFunc(
	includeIdentityGRPCFunc middleware.IncludeIdentityGRPCFunc,
) *ServiceRunnerBuilder {
	s.includeIdentityGRPCFunc = includeIdentityGRPCFunc
	return s
}

func (s *ServiceRunnerBuilder) Build() ServiceRunner {
	return ServiceRunner{
		dataCollector: s.dataCollector,
		config:        s.config,
		httpClient:    s.httpClient,
		webRouter:     mux.NewRouter(),
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.ServerGRPCWithTimeout(s.config.RequestTimeout),
				middleware.ServerGRPCWithRequestID(s.dataCollector),
				middleware.ServerGRPCLogRequest(s.dataCollector),
				middleware.ServerGRPCWithIdentity(
					s.dataCollector,
					s.httpClient,
					s.config.IdentityAPIEndpoint,
					s.includeIdentityGRPCFunc),
			)),
		services:               s.services,
		includeIdentityWebFunc: s.includeIdentityWebFunc,
	}
}

func NewServiceRunnerBuilder(
	dataCollector telemetry.DataCollector,
	config ServiceRunnerConfig,
	services []Service,
) *ServiceRunnerBuilder {
	middlewares := []middleware.Middleware[web.HTTPClient]{
		middleware.ClientHTTPWithRequestID(dataCollector),
	}
	httpClient := middleware.WithMiddlewares[web.HTTPClient](
		func(ct context.Context, req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		}, middlewares)
	return &ServiceRunnerBuilder{
		dataCollector: dataCollector,
		config:        config,
		httpClient:    httpClient,
		services:      services,
		includeIdentityWebFunc: func(request *http.Request) bool {
			return true
		},
		includeIdentityGRPCFunc: func(info *grpc.UnaryServerInfo) bool {
			return true
		}}
}
