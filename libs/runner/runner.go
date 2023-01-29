package runner

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
)

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

func ServiceRunnerConfigFromEnv(dataCollector telemetry.DataCollector) (ServiceRunnerConfig, error) {
	cfg := ServiceRunnerConfig{}
	err := config.FromEnv(dataCollector, &cfg)
	if err != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return ServiceRunnerConfig{}, err
	}

	return cfg, nil
}

type ServiceRunner struct {
	dataCollector telemetry.DataCollector
	config        ServiceRunnerConfig
	webRouter     *mux.Router
	gRPCServer    *grpc.Server
	services      []Service
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
		telemetry.MessageProp: telemetry.Props{
			"Summary": "service runner Web server started",
			"Port":    s.config.WebServerPort,
		},
	})
	serveMux := http.NewServeMux()
	middlewares := []middleware.Middleware[http.HandlerFunc]{
		middleware.ServerHTTPEnableCORS,
		middleware.ServerHTTPWithRequestID(s.dataCollector),
		middleware.ServerHTTPWithTimeout(s.config.RequestTimeout),
		middleware.ServerHTTPLogRequest(s.dataCollector),
		middleware.ServerHTTPWithIdentity(s.dataCollector, s.config.IdentityAPIEndpoint),
		middleware.ServerWebSocketWithIdentity(s.dataCollector, s.config.IdentityAPIEndpoint),
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
		telemetry.MessageProp: telemetry.Props{
			"Summary": "service runner gRPC server started",
			"Port":    s.config.GRPCServerPort,
		},
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

func NewServiceRunner(dataCollector telemetry.DataCollector, config ServiceRunnerConfig, services []Service) ServiceRunner {
	return ServiceRunner{
		dataCollector: dataCollector,
		config:        config,
		webRouter:     mux.NewRouter(),
		gRPCServer: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				middleware.ServerGRPCWithTimeout(config.RequestTimeout),
				middleware.ServerGRPCWithRequestID(dataCollector),
				middleware.ServerGRPCLogRequest(dataCollector),
				middleware.ServerGRPCWithIdentity(dataCollector, config.IdentityAPIEndpoint),
			)),
		services: services,
	}
}
