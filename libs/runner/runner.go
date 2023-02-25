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
	WebServerPort          int           `envconfig:"SERVICE_RUNNER_WEB_SERVER_PORT" default:"9011"`
	GRPCServerPort         int           `envconfig:"SERVICE_RUNNER_GRPC_SERVER_PORT" default:"9012"`
	MonitoringServerPort   int           `envconfig:"SERVICE_RUNNER_MONITORING_SERVER_PORT" default:"10000"`
	IdentityAPIEndpoint    string        `envconfig:"SERVICE_RUNNER_IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	RequestTimeout         time.Duration `envconfig:"SERVICE_RUNNER_REQUEST_TIMEOUT" default:"10s"`
	EnableTracing          bool          `envconfig:"SERVICE_RUNNER_ENABLE_TRACING" default:"false"`
	TraceCollectorEndpoint string        `envconfig:"SERVICE_RUNNER_TRACE_COLLECTOR_ENDPOINT" default:"localhost:4317"`
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
	serviceName            string
	httpClient             web.HTTPClient
	webRouter              chi.Router
	gRPCServer             *grpc.Server
	services               []Service
	includeIdentityWebFunc middleware.IncludeIdentityWebFunc
}

func (s *ServiceRunner) Start() {
	var shutdown func(ct context.Context) error
	if s.config.EnableTracing {
		var internalErr *errs.Error
		shutdown, internalErr = telemetry.InitTracerProvider(s.dataCollector, s.config.TraceCollectorEndpoint, s.serviceName)
		if internalErr != nil {
			s.dataCollector.Logger.Error(internalErr)
		}
	}

	for _, service := range s.services {
		err := service.Start(s)
		if err != nil {
			s.dataCollector.Logger.Fatal(err)
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
	if s.config.EnableTracing {
		_ = shutdown(context.Background())
	}
}

func (s *ServiceRunner) startWebServer() {
	s.dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Web server started at %v", s.config.WebServerPort),
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.WebServerPort), s.webRouter); err != nil {
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
	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.MonitoringServerPort), router); err != nil {
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
	getClientHTTPRequestPatternFunc middleware.GetPatternFunc
	serviceName                     string
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
	getClientHTTPRequestPatternFunc middleware.GetPatternFunc,
) *ServiceRunnerBuilder {
	s.getClientHTTPRequestPatternFunc = getClientHTTPRequestPatternFunc
	return s
}

func (s *ServiceRunnerBuilder) Build() ServiceRunner {
	httpClientMiddlewares := []middleware.Middleware[web.HTTPClient]{
		middleware.ClientHTTPWithMetrics(s.prometheus, s.getClientHTTPRequestPatternFunc),
		middleware.ClientHTTPWithOpenTelemetry(s.getClientHTTPRequestPatternFunc),
		middleware.ClientHTTPWithRequestID(s.dataCollector),
	}
	httpClient := middleware.WithMiddlewares[web.HTTPClient](
		func(ct context.Context, req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		}, httpClientMiddlewares)
	webRouter := chi.NewRouter()
	httpServerMiddlewares := []middleware.Middleware[http.HandlerFunc]{
		middleware.ServerHTTPWithMetrics(s.prometheus, getClientHTTPRequestPatternFunc),
		middleware.ServerHTTPWithOpenTelemetry(s.dataCollector, getClientHTTPRequestPatternFunc),
		middleware.ServerHTTPEnableCORS,
		middleware.ServerHTTPWithRequestID(s.dataCollector),
		middleware.ServerHTTPWithTimeout(s.config.RequestTimeout),
		middleware.ServerHTTPLogRequest(s.dataCollector),
		middleware.ServerHTTPWithIdentity(
			s.dataCollector,
			httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
		middleware.ServerWebSocketWithIdentity(
			s.dataCollector,
			httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
	}
	webRouter.Use(func(handler http.Handler) http.Handler {
		return middleware.WithMiddlewares[http.HandlerFunc](handler.ServeHTTP, httpServerMiddlewares)
	})
	return ServiceRunner{
		dataCollector: s.dataCollector,
		prometheus:    s.prometheus,
		config:        s.config,
		serviceName:   s.serviceName,
		httpClient:    httpClient,
		webRouter:     webRouter,
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.ServerGRPCWithMetrics(s.prometheus),
				middleware.ServerGRPCUnaryWithOpenTelemetry(),
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
	serviceName string,
	services []Service,
) *ServiceRunnerBuilder {
	return &ServiceRunnerBuilder{
		dataCollector: dataCollector,
		prometheus:    prometheus,
		config:        config,
		serviceName:   serviceName,
		services:      services,
		includeIdentityWebFunc: func(request *http.Request) bool {
			return true
		},
		includeIdentityGRPCFunc: func(info *grpc.UnaryServerInfo) bool {
			return true
		},
		getClientHTTPRequestPatternFunc: func(request *http.Request) (string, bool) {
			return request.URL.Path, true
		}}
}

func Param(paramName string) string {
	return fmt.Sprintf(`{%s}`, paramName)
}

func getClientHTTPRequestPatternFunc(request *http.Request) (string, bool) {
	rctx := chi.RouteContext(request.Context())
	if pattern := rctx.RoutePattern(); len(pattern) > 0 {
		return pattern, true
	}

	tmpCtx := chi.NewRouteContext()
	if !rctx.Routes.Match(tmpCtx, request.Method, request.URL.Path) {
		return "", false
	}

	return tmpCtx.RoutePattern(), true
}
