package metricstest

import (
	"time"

	"github.com/teamyapp/cloud/libs/middleware"
)

type Metrics interface {
	middleware.ServerHTTPMetrics
	middleware.ClientHTTPMetrics
	middleware.ServerGRPCMetrics
	middleware.ClientGRPCMetrics
}

type NoopMetrics struct{}

var _ Metrics = (*NoopMetrics)(nil)

func (n NoopMetrics) ReportHTTPIncomingRequest(method string, pattern string) {
}

func (n NoopMetrics) ReportHTTPIncomingRequestResponseTime(method string, pattern string, duration time.Duration) {
}

func (n NoopMetrics) ReportHTTPOutgoingRequest(target string, method string, pattern string) {
}

func (n NoopMetrics) ReportHTTPOutgoingRequestResponseTime(target string, method string, pattern string, duration time.Duration) {
}

func (n NoopMetrics) ReportGRPCIncomingRequest(method string) {
}

func (n NoopMetrics) ReportGRPCIncomingRequestResponseTime(method string, duration time.Duration) {
}

func (n NoopMetrics) ReportGRPCOutgoingRequest(target string, method string) {
}

func (n NoopMetrics) ReportGRPCOutgoingRequestResponseTime(target string, method string, duration time.Duration) {
}

func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}
