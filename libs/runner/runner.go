package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

const EnvConfigErr errs.ErrorCode = "EnvConfig"

type WebRoute struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type ServiceRunnerConfig struct {
	WebServerPort        int           `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9011"`
	GRPCServerPort       int           `envconfig:"SERVICE_RUNNER_GRPC_SERVER_PORT" default:"9012"`
	MonitoringServerPort int           `envconfig:"SERVICE_RUNNER_MONITORING_SERVER_PORT" default:"10000"`
	IdentityAPIEndpoint  string        `envconfig:"SERVICE_RUNNER_IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	RequestTimeout       time.Duration `envconfig:"REQUEST_TIMEOUT" default:"10s"`
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
	prometheus             metrics.Prometheus
	config                 ServiceRunnerConfig
	httpClient             web.HTTPClient
	webRouter              chi.Router
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.startMonitoringServer()
	}()

	wg.Wait()
}

func (s *ServiceRunner) startWebServer() {
	s.dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Web server started at %v", s.config.WebServerPort),
	})
	serveMux := http.NewServeMux()
	middlewares := []middleware.Middleware[http.HandlerFunc]{
		middleware.ServerHTTPWithMetrics(s.prometheus, func(request *http.Request) string {
			return chi.RouteContext(request.Context()).RoutePattern()
		}),
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

func (s *ServiceRunner) startMonitoringServer() {
	s.dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Monitoring server started at %v", s.config.MonitoringServerPort),
	})
	serveMux := http.NewServeMux()
	serveMux.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.MonitoringServerPort), serveMux); err != nil {
		panic(err)
	}
}

func (s *ServiceRunner) RegisterWebRoutes(routes []WebRoute) {
	for _, route := range routes {
		s.webRouter.MethodFunc(route.Method, route.Pattern, route.HandlerFunc)
	}
}

func (s *ServiceRunner) WithGRPCServer(withGRPCServer func(server *grpc.Server)) {
	withGRPCServer(s.gRPCServer)
}

type ServiceRunnerBuilder struct {
	dataCollector                   telemetry.DataCollector
	prometheus                      metrics.Prometheus
	config                          ServiceRunnerConfig
	services                        []Service
	includeIdentityWebFunc          middleware.IncludeIdentityWebFunc
	includeIdentityGRPCFunc         middleware.IncludeIdentityGRPCFunc
	getClientHTTPRequestPatternFunc func(request *http.Request) string
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

func (s *ServiceRunnerBuilder) GetClientHTTPRequestPatternFunc(
	getClientHTTPRequestPatternFunc func(request *http.Request) string,
) *ServiceRunnerBuilder {
	s.getClientHTTPRequestPatternFunc = getClientHTTPRequestPatternFunc
	return s
}

func (s *ServiceRunnerBuilder) Build() ServiceRunner {
	middlewares := []middleware.Middleware[web.HTTPClient]{
		middleware.ClientHTTPWithMetrics(s.prometheus, s.getClientHTTPRequestPatternFunc),
		middleware.ClientHTTPWithRequestID(s.dataCollector),
	}
	httpClient := middleware.WithMiddlewares[web.HTTPClient](
		func(ct context.Context, req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		}, middlewares)
	return ServiceRunner{
		dataCollector: s.dataCollector,
		prometheus:    s.prometheus,
		config:        s.config,
		httpClient:    httpClient,
		webRouter:     chi.NewRouter(),
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.ServerGRPCWithMetrics(s.prometheus),
				middleware.ServerGRPCWithTimeout(s.config.RequestTimeout),
				middleware.ServerGRPCWithRequestID(s.dataCollector),
				middleware.ServerGRPCLogRequest(s.dataCollector),
				middleware.ServerGRPCWithIdentity(
					s.dataCollector,
					httpClient,
					s.config.IdentityAPIEndpoint,
					s.includeIdentityGRPCFunc),
			)),
		services:               s.services,
		includeIdentityWebFunc: s.includeIdentityWebFunc,
	}
}

func NewServiceRunnerBuilder(
	dataCollector telemetry.DataCollector,
	prometheus metrics.Prometheus,
	config ServiceRunnerConfig,
	services []Service,
) *ServiceRunnerBuilder {
	return &ServiceRunnerBuilder{
		dataCollector: dataCollector,
		prometheus:    prometheus,
		config:        config,
		services:      services,
		includeIdentityWebFunc: func(request *http.Request) bool {
			return true
		},
		includeIdentityGRPCFunc: func(info *grpc.UnaryServerInfo) bool {
			return true
		},
		getClientHTTPRequestPatternFunc: func(request *http.Request) string {
			return request.URL.Path
		}}
}

func Param(paramName string) string {
	return fmt.Sprintf(`"{%s}"`, paramName)
}
