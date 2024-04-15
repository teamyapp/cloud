package metrics

import "github.com/teamyapp/cloud/libs/middleware"

const EnvironmentLabel string = "environment"

type Metrics interface {
	middleware.ServerHTTPMetrics
	middleware.ClientHTTPMetrics
	middleware.ServerGRPCMetrics
	middleware.ClientGRPCMetrics
}
