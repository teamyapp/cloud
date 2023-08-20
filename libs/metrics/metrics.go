package metrics

import "github.com/teamyapp/cloud/libs/middleware"

const EnvironmentLabel string = "environment"

var ResponseTimeBuckets = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

type Metrics interface {
	middleware.ServerHTTPMetrics
	middleware.ClientHTTPMetrics
	middleware.ServerGRPCMetrics
	middleware.ClientGRPCMetrics
}
