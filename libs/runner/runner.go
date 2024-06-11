package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
	ProfilingServerPort    int           `envconfig:"SERVICE_RUNNER_PROFILING_SERVER_PORT" default:"10001"`
	FileServerPort         int           `envconfig:"SERVICE_RUNNER_FILE_SERVER_PORT" default:"10002"`
	IdentityAPIEndpoint    string        `envconfig:"SERVICE_RUNNER_IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	RequestTimeout         time.Duration `envconfig:"SERVICE_RUNNER_REQUEST_TIMEOUT" default:"10s"`
	EnableTracing          bool          `envconfig:"SERVICE_RUNNER_ENABLE_TRACING" default:"false"`
	TraceCollectorEndpoint string        `envconfig:"SERVICE_RUNNER_TRACE_COLLECTOR_ENDPOINT" default:"localhost:4317"`
}

func ServiceRunnerConfigFromEnv() (ServiceRunnerConfig, *errs.Error) {
	cfg := ServiceRunnerConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		return ServiceRunnerConfig{}, err
	}

	return cfg, nil
}

type DirRoute struct {
	WebPath        string
	FileSystemPath string
}

type ServiceRunner struct {
	logger                 telemetry.Logger
	network                network.Network
	config                 ServiceRunnerConfig
	serviceName            string
	httpClient             web.HTTPClient
	webRouter              chi.Router
	dirRoutes              []DirRoute
	gRPCServer             *grpc.Server
	services               []Service
	includeIdentityWebFunc middleware.IncludeIdentityWebFunc
}

func (s *ServiceRunner) Start(afterServicesStarted func(listeners []net.Listener) *errs.Error) *errs.Error {
	var shutdown func(ct context.Context) error
	if s.config.EnableTracing {
		var internalErr *errs.Error
		shutdown, internalErr = telemetry.InitTracerProvider(s.logger, s.config.TraceCollectorEndpoint, s.serviceName)
		if internalErr != nil {
			s.logger.Error(internalErr)
		}
	}

	for _, service := range s.services {
		err := service.Start(s)
		if err != nil {
			s.logger.Error(err)
			return err
		}
	}

	wg := sync.WaitGroup{}
	listeners := make([]net.Listener, 0)
	lis, err := s.startWebServer(&wg)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	listeners = append(listeners, lis)
	lis, err = s.startFileServer(&wg)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	listeners = append(listeners, lis)
	lis, err = s.startGRPCServer(&wg)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	listeners = append(listeners, lis)
	lis, err = s.startMonitoringServer(&wg)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	listeners = append(listeners, lis)
	lis, err = s.startProfilingServer(&wg)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	listeners = append(listeners, lis)
	if afterServicesStarted != nil {
		err = afterServicesStarted(listeners)
		if err != nil {
			s.logger.Error(err)
			return err
		}
	}

	wg.Wait()
	if s.config.EnableTracing {
		_ = shutdown(context.Background())
	}

	return nil
}

func (s *ServiceRunner) startFileServer(wg *sync.WaitGroup) (net.Listener, *errs.Error) {
	s.logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner File server started at %v", s.config.FileServerPort),
	})
	addressAndPort := fmt.Sprintf(":%d", s.config.FileServerPort)
	lis, err := s.network.Listen("tcp", addressAndPort)
	if err != nil {
		return lis, errs.NewError(errs.Unknown, err.Error())
	}

	mux := http.NewServeMux()
	for _, route := range s.dirRoutes {
		mux.Handle(route.WebPath, http.StripPrefix(route.WebPath, http.FileServer(http.Dir(route.FileSystemPath))))
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = http.Serve(lis, mux)
		if err != nil {
			s.logger.Fatal(errs.NewError(errs.Unknown, err.Error()))
		}
	}()
	return lis, nil
}

func (s *ServiceRunner) startWebServer(wg *sync.WaitGroup) (net.Listener, *errs.Error) {
	s.logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Web server started at %v", s.config.WebServerPort),
	})
	addressAndPort := fmt.Sprintf(":%d", s.config.WebServerPort)
	lis, err := s.network.Listen("tcp", addressAndPort)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = http.Serve(lis, s.webRouter); err != nil {
			s.logger.Fatal(errs.NewError(errs.Unknown, err.Error()))
		}
	}()
	return lis, nil
}

func (s *ServiceRunner) startGRPCServer(wg *sync.WaitGroup) (net.Listener, *errs.Error) {
	hostAndPort := fmt.Sprintf(":%d", s.config.GRPCServerPort)
	lis, err := s.network.Listen("tcp", hostAndPort)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	s.logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner gRPC server started at %v", s.config.GRPCServerPort),
	})

	grpcWebServer := grpcweb.WrapServer(s.gRPCServer, grpcweb.WithOriginFunc(func(origin string) bool {
		return true
	}))
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGrpcWebRequest(grpcWebServer, r) {
			grpcWebServer.ServeHTTP(w, r)
		} else if isGrpcRequest(r) {
			s.gRPCServer.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	httpServer := &http.Server{
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = httpServer.Serve(lis)
		if err != nil {
			s.logger.Fatal(errs.NewError(errs.Unknown, err.Error()))
		}
	}()
	return lis, nil
}

func (s *ServiceRunner) startMonitoringServer(wg *sync.WaitGroup) (net.Listener, *errs.Error) {
	s.logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Monitoring server started at %v", s.config.MonitoringServerPort),
	})
	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	hostAndPort := fmt.Sprintf(":%d", s.config.MonitoringServerPort)
	lis, err := s.network.Listen("tcp", hostAndPort)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = http.Serve(lis, router)
		if err != nil {
			s.logger.Fatal(errs.NewError(errs.Unknown, err.Error()))
		}
	}()
	return lis, nil
}

func (s *ServiceRunner) startProfilingServer(wg *sync.WaitGroup) (net.Listener, *errs.Error) {
	s.logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("service runner Profiling server started at %v", s.config.ProfilingServerPort),
	})
	router := chi.NewRouter()
	router.HandleFunc("/", pprof.Index)
	router.HandleFunc("/cmdline", pprof.Cmdline)
	router.HandleFunc("/profile", pprof.Profile)
	router.HandleFunc("/symbol", pprof.Symbol)
	router.HandleFunc("/trace", pprof.Trace)

	router.HandleFunc("/goroutine", pprof.Handler("goroutine").ServeHTTP)
	router.HandleFunc("/heap", pprof.Handler("heap").ServeHTTP)
	router.HandleFunc("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	router.HandleFunc("/block", pprof.Handler("block").ServeHTTP)
	router.HandleFunc("/allocs", pprof.Handler("allocs").ServeHTTP)
	router.HandleFunc("/mutex", pprof.Handler("mutex").ServeHTTP)

	hostAndPort := fmt.Sprintf(":%d", s.config.ProfilingServerPort)
	lis, err := s.network.Listen("tcp", hostAndPort)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = http.Serve(lis, router)
		if err != nil {
			s.logger.Fatal(errs.NewError(errs.Unknown, err.Error()))
		}
	}()
	return lis, nil
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
	logger                          telemetry.Logger
	network                         network.Network
	metrics                         metrics.Metrics
	config                          ServiceRunnerConfig
	dirRoutes                       []DirRoute
	serviceName                     string
	services                        []Service
	includeIdentityWebFunc          middleware.IncludeIdentityWebFunc
	includeIdentityGRPCFunc         middleware.IncludeIdentityGRPCFunc
	getClientHTTPRequestPatternFunc middleware.GetPatternFunc
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

func (s *ServiceRunnerBuilder) ServeDirs(dirRoutes []DirRoute) *ServiceRunnerBuilder {
	s.dirRoutes = append(s.dirRoutes, dirRoutes...)
	return s
}

func (s *ServiceRunnerBuilder) Build() ServiceRunner {
	rawHttpClient := web.NewHTTPClient(s.network)
	httpClientMiddlewares := []middleware.Middleware[web.HTTPClient]{
		middleware.ClientHTTPWithMetrics(s.metrics, s.getClientHTTPRequestPatternFunc),
		middleware.ClientHTTPWithOpenTelemetry(s.getClientHTTPRequestPatternFunc),
		middleware.ClientHTTPWithRequestID(s.logger),
	}
	httpClient := middleware.WithMiddlewares[web.HTTPClient](rawHttpClient, httpClientMiddlewares)
	webRouter := chi.NewRouter()
	httpServerMiddlewares := []middleware.Middleware[http.HandlerFunc]{
		middleware.ServerHTTPWithMetrics(s.metrics, getClientHTTPRequestPatternFunc),
		middleware.ServerHTTPWithOpenTelemetry(s.logger, getClientHTTPRequestPatternFunc),
		middleware.ServerHTTPEnableCORS,
		middleware.ServerHTTPWithRequestID(s.logger),
		middleware.ServerHTTPWithTimeout(s.config.RequestTimeout),
		middleware.ServerHTTPLogRequest(s.logger),
		middleware.ServerHTTPWithIdentity(
			s.logger,
			httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
		middleware.ServerWebSocketWithIdentity(
			s.logger,
			httpClient,
			s.config.IdentityAPIEndpoint,
			s.includeIdentityWebFunc),
	}
	webRouter.Use(func(handler http.Handler) http.Handler {
		return middleware.WithMiddlewares[http.HandlerFunc](handler.ServeHTTP, httpServerMiddlewares)
	})
	return ServiceRunner{
		logger:      s.logger,
		network:     s.network,
		config:      s.config,
		serviceName: s.serviceName,
		httpClient:  httpClient,
		webRouter:   webRouter,
		dirRoutes:   s.dirRoutes,
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.ServerGRPCWithMetrics(s.metrics),
				middleware.ServerGRPCUnaryWithOpenTelemetry(),
				middleware.ServerGRPCWithTimeout(s.config.RequestTimeout),
				middleware.ServerGRPCWithRequestID(s.logger),
				middleware.ServerGRPCLogRequest(s.logger),
				middleware.ServerGRPCWithIdentity(
					s.logger,
					httpClient,
					s.config.IdentityAPIEndpoint,
					s.includeIdentityGRPCFunc),
			)),
		services:               s.services,
		includeIdentityWebFunc: s.includeIdentityWebFunc,
	}
}

func NewServiceRunnerBuilder(
	logger telemetry.Logger,
	network network.Network,
	metrics metrics.Metrics,
	config ServiceRunnerConfig,
	serviceName string,
	services []Service,
) *ServiceRunnerBuilder {
	return &ServiceRunnerBuilder{
		logger:      logger,
		network:     network,
		metrics:     metrics,
		config:      config,
		serviceName: serviceName,
		services:    services,
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

func isGrpcWebRequest(grpcWebServer *grpcweb.WrappedGrpcServer, request *http.Request) bool {
	return grpcWebServer.IsGrpcWebRequest(request) || grpcWebServer.IsGrpcWebSocketRequest(request) || grpcWebServer.IsAcceptableGrpcCorsRequest(request)
}

func isGrpcRequest(request *http.Request) bool {
	return request.ProtoMajor == 2 && request.Header.Get("Content-Type") == "application/grpc"
}
